package ui

import (
	"cpu-monitor/internal/monitor"
	"fmt"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
)

type Display struct {
	screen        tcell.Screen
	monitor       *monitor.Monitor
	colorScheme   *ColorScheme
	inputHandler  *InputHandler
	config        ConfigInterface
	mu            sync.RWMutex
	processes     []*monitor.ProcessInfo
	selectedIndex int
	paused        bool
	forceRefresh  bool
	running       bool
}

type ConfigInterface interface {
	GetRefreshRate() time.Duration
	GetCPUThreshold() float64
	GetMemoryThreshold() uint64
}

func New(config ConfigInterface, mon *monitor.Monitor) *Display {
	d := &Display{
		monitor:       mon,
		colorScheme:   NewColorScheme(),
		config:        config,
		selectedIndex: 0,
		paused:        false,
		forceRefresh:  false,
		running:       true,
	}
	d.inputHandler = NewInputHandler(d)
	return d
}

func (d *Display) Run() error {
	var err error
	d.screen, err = tcell.NewScreen()
	if err != nil {
		return fmt.Errorf("failed to create screen: %w", err)
	}

	if err = d.screen.Init(); err != nil {
		return fmt.Errorf("failed to initialize screen: %w", err)
	}

	d.screen.SetStyle(tcell.StyleDefault.Background(d.colorScheme.Background).Foreground(d.colorScheme.Text))
	d.screen.Clear()

	go d.updateLoop()
	go d.inputLoop()

	for d.running {
		d.render()
		time.Sleep(50 * time.Millisecond)
	}

	d.screen.Fini()
	return nil
}

func (d *Display) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.running = false
	if d.screen != nil {
		d.screen.Fini()
	}
}

func (d *Display) updateLoop() {
	ticker := time.NewTicker(d.config.GetRefreshRate())
	defer ticker.Stop()

	for d.running {
		select {
		case <-ticker.C:
			if !d.paused || d.forceRefresh {
				d.updateProcesses()
				d.forceRefresh = false
			}
		}
	}
}

func (d *Display) inputLoop() {
	for d.running {
		ev := d.screen.PollEvent()
		if ev == nil {
			continue
		}

		switch ev := ev.(type) {
		case *tcell.EventKey:
			if !d.inputHandler.HandleInput(ev) {
				d.Stop()
				return
			}
		case *tcell.EventResize:
			d.screen.Sync()
		}
	}
}

func (d *Display) updateProcesses() {
	processes, err := d.monitor.GetFilteredProcesses()
	if err != nil {
		return
	}

	d.mu.Lock()
	d.processes = processes
	if d.selectedIndex >= len(d.processes) {
		d.selectedIndex = len(d.processes) - 1
	}
	if d.selectedIndex < 0 {
		d.selectedIndex = 0
	}
	d.mu.Unlock()
}

func (d *Display) render() {
	d.mu.RLock()
	defer d.mu.RUnlock()

	d.screen.Clear()
	width, height := d.screen.Size()

	d.renderHeader(width)
	d.renderProcesses(width, height)
	d.renderFooter(width, height)

	d.screen.Show()
}

func (d *Display) renderHeader(width int) {
	headerText := fmt.Sprintf("CPU Monitor - Processes using >%.1f%% CPU or >%dMB RAM", 
		d.config.GetCPUThreshold(), d.config.GetMemoryThreshold()/(1024*1024))
	
	if d.paused {
		headerText += " [PAUSED]"
	}

	d.drawText(0, 0, width, headerText, d.colorScheme.GetStyle(d.colorScheme.Header, false))
	
	// New column order: PID, CPU, MEMORY, CHILDREN, NAME (expanded to fill remaining space)
	columnHeaders := fmt.Sprintf("%-8s %8s %10s %8s %s", "PID", "CPU", "MEMORY", "CHILDREN", "NAME")
	d.drawText(0, 1, width, columnHeaders, d.colorScheme.GetStyle(d.colorScheme.Header, false))
}

func (d *Display) renderProcesses(width, height int) {
	startY := 3
	maxRows := height - 5
	currentY := startY

	for i, proc := range d.processes {
		if currentY >= startY + maxRows {
			break
		}

		isSelected := i == d.selectedIndex

		level := d.monitor.GetResourceLevel(proc.CPUPercent, proc.MemoryMB)
		color := d.colorScheme.GetProcessColor(level)
		style := d.colorScheme.GetStyle(color, isSelected)

		childCount := len(proc.Children)
		expandIndicator := ""
		if childCount > 0 {
			if proc.Expanded {
				expandIndicator = "▼"
			} else {
				expandIndicator = "▶"
			}
		}

		// Calculate available space for name (total width - fixed columns)
		// Format: [▼]PID(8) CPU(9) MEMORY(11) CHILDREN(9) + spaces = ~40 chars
		fixedColumnsWidth := 40
		availableNameWidth := width - fixedColumnsWidth
		if availableNameWidth < 20 {
			availableNameWidth = 20 // minimum name width
		}
		
		processLine := fmt.Sprintf("%s%-7d %7.1f%% %9.1fMB %7d %s", 
			expandIndicator, proc.PID, proc.CPUPercent, proc.MemoryMB, childCount,
			truncateString(proc.Name, availableNameWidth))

		d.drawText(0, currentY, width, processLine, style)
		currentY++

		if proc.Expanded && childCount > 0 {
			// First show the parent process itself as the first entry
			if currentY < startY + maxRows {
				parentPrefix := "  ├─ "  // Parent process indicator
				parentStyle := d.colorScheme.GetStyle(d.colorScheme.Text, false)
				
				// Calculate available space for parent name
				parentFixedWidth := 35
				parentNameWidth := width - parentFixedWidth
				if parentNameWidth < 15 {
					parentNameWidth = 15
				}
				
				parentLine := fmt.Sprintf("%s%-5d %7.1f%% %9.1fMB %s (parent)", 
					parentPrefix, proc.PID, proc.ParentCPU, float64(proc.ParentMemory)/(1024*1024),
					truncateString(proc.Name, parentNameWidth-9)) // -9 for " (parent)"
				
				d.drawText(0, currentY, width, parentLine, parentStyle)
				currentY++
			}
			
			// Then show all children
			for _, child := range proc.Children {
				if currentY >= startY + maxRows {
					break
				}
				
				// Different visual indicators for threads vs child processes
				var prefix string
				var childStyle tcell.Style
				if child.IsThread {
					prefix = "  ╠═ "  // Thread indicator
					childStyle = d.colorScheme.GetStyle(d.colorScheme.Thread, false)
				} else {
					prefix = "  ├─ "  // Child process indicator
					childStyle = d.colorScheme.GetStyle(d.colorScheme.ChildProcess, false)
				}
				
				// Calculate available space for child name
				childFixedWidth := 35 // prefix + PID + CPU + Memory + spaces
				childNameWidth := width - childFixedWidth
				if childNameWidth < 15 {
					childNameWidth = 15
				}
				
				childLine := fmt.Sprintf("%s%-5d %7.1f%% %9.1fMB %s", 
					prefix, child.PID, child.CPUPercent, float64(child.MemoryBytes)/(1024*1024),
					truncateString(child.Name, childNameWidth))
				
				d.drawText(0, currentY, width, childLine, childStyle)
				currentY++
			}
		}
	}
}

func (d *Display) renderFooter(width, height int) {
	footerY := height - 1
	footerText := "Controls: ↑/↓ Navigate | Enter Expand/Collapse | Space Pause | R Refresh | Q Quit"
	d.drawText(0, footerY, width, footerText, d.colorScheme.GetStyle(d.colorScheme.Header, false))
}

func (d *Display) drawText(x, y, maxWidth int, text string, style tcell.Style) {
	runes := []rune(text)
	for i, r := range runes {
		if x+i >= maxWidth {
			break
		}
		d.screen.SetContent(x+i, y, r, nil, style)
	}
}

func truncateString(s string, maxLen int) string {
	if maxLen < 4 {
		maxLen = 4 // Minimum to show "..."
	}
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}