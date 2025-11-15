package ui

import (
	"brieftop/internal/monitor"
	"fmt"
	"strings"
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

	// Draw main border
	d.drawBorder(0, 0, width, height)
	
	d.renderHeader(width)
	d.renderProcesses(width, height)
	d.renderFooter(width, height)

	d.screen.Show()
}

func (d *Display) renderHeader(width int) {
	// Header with better formatting and icons
	status := "✓ RUNNING"
	statusColor := d.colorScheme.Success
	if d.paused {
		status = "⏸ PAUSED"
		statusColor = d.colorScheme.Warning
	}
	
	headerText := fmt.Sprintf("⚙️  brieftop - Processes >%.1f%% CPU or >%dMB RAM",
		d.config.GetCPUThreshold(), d.config.GetMemoryThreshold()/(1024*1024))
	
	// Main header
	d.drawText(2, 1, width-4, headerText, d.colorScheme.GetStyle(d.colorScheme.Header, false))
	
	// Status indicator
	statusX := width - len(status) - 3
	d.drawText(statusX, 1, len(status), status, d.colorScheme.GetStyle(statusColor, false))
	
	// Separator line
	d.drawHorizontalLine(2, 2, width-4, "─", d.colorScheme.Border)
	
	// Column headers with better spacing
	columnHeaders := fmt.Sprintf("  %-7s %7s %10s %6s %s", 
		"PID", "CPU", "MEMORY", "CHILD", "PROCESS NAME")
	d.drawText(2, 3, width-4, columnHeaders, d.colorScheme.GetStyle(d.colorScheme.Accent, false))
	
	// Header separator
	d.drawHorizontalLine(2, 4, width-4, "━", d.colorScheme.Border)
}

func (d *Display) renderProcesses(width, height int) {
	startY := 5  // Start after enhanced header
	maxRows := height - 8  // Leave space for footer
	currentY := startY

	for i, proc := range d.processes {
		if currentY >= startY + maxRows {
			break
		}

		isSelected := i == d.selectedIndex
		childCount := len(proc.Children)
		
		// Enhanced status icon
		statusIcon := GetStatusIcon(proc.CPUPercent, proc.Expanded, childCount > 0)
		
		// Color based on resource usage
		level := d.monitor.GetResourceLevel(proc.CPUPercent, proc.MemoryMB)
		color := d.colorScheme.GetProcessColor(level)
		style := d.colorScheme.GetStyle(color, isSelected)
		
		// Calculate available space for name
		availableNameWidth := width - 45
		if availableNameWidth < 20 {
			availableNameWidth = 20
		}
		
		// Main process line with proper formatting
		processLine := fmt.Sprintf("%s %-6d %6.1f%% %9.1fMB %5d %s", 
			statusIcon, proc.PID, proc.CPUPercent, proc.MemoryMB, childCount,
			truncateString(proc.Name, availableNameWidth))

		d.drawText(3, currentY, width-6, processLine, style)
		currentY++

		if proc.Expanded && childCount > 0 {
			// First show the parent process itself
			if currentY < startY + maxRows {
				parentPrefix := "    ├─●"  // Parent indicator
				parentStyle := d.colorScheme.GetStyle(d.colorScheme.Text, false)
				
				availableParentNameWidth := width - 50
				if availableParentNameWidth < 15 {
					availableParentNameWidth = 15
				}
				
				parentLine := fmt.Sprintf("%s %-5d %6.1f%% %9.1fMB      %s (parent)", 
					parentPrefix, proc.PID, proc.ParentCPU, float64(proc.ParentMemory)/(1024*1024),
					truncateString(proc.Name, availableParentNameWidth-9))
				
				d.drawText(3, currentY, width-6, parentLine, parentStyle)
				currentY++
			}
			
			// Then show all children
			for _, child := range proc.Children {
				if currentY >= startY + maxRows {
					break
				}
				
				// Visual indicators for different types
				var prefix string
				var childStyle tcell.Style
				var typeLabel string
				
				if child.IsThread {
					prefix = "    ╠═"  // Thread indicator
					childStyle = d.colorScheme.GetStyle(d.colorScheme.Thread, false)
					typeLabel = "thread"
				} else {
					prefix = "    ├─"  // Child process indicator
					childStyle = d.colorScheme.GetStyle(d.colorScheme.ChildProcess, false)
					typeLabel = "child"
				}
				
				availableChildNameWidth := width - 55
				if availableChildNameWidth < 15 {
					availableChildNameWidth = 15
				}
				
				childLine := fmt.Sprintf("%s %-5d %6.1f%% %9.1fMB      %s (%s)", 
					prefix, child.PID, child.CPUPercent, float64(child.MemoryBytes)/(1024*1024),
					truncateString(child.Name, availableChildNameWidth-len(typeLabel)-3), typeLabel)
				
				d.drawText(3, currentY, width-6, childLine, childStyle)
				currentY++
			}
		}
	}
}

func (d *Display) renderFooter(width, height int) {
	footerY := height - 3
	
	// Footer border
	d.drawHorizontalLine(2, footerY, width-4, "─", d.colorScheme.Border)
	
	// Enhanced controls with icons
	controls := []string{
		"↑↓ Navigate",
		"⏎ Expand",
		"⏸ Pause",
		"↻ Refresh",
		"✗ Quit",
	}
	
	footerText := "🎮 Controls: " + fmt.Sprintf("%s", strings.Join(controls, " │ "))
	d.drawText(3, footerY+1, width-6, footerText, d.colorScheme.GetStyle(d.colorScheme.Accent, false))
	
	// Process count and stats
	processCount := len(d.processes)
	statsText := fmt.Sprintf("📊 Showing %d processes", processCount)
	d.drawText(width-len(statsText)-3, footerY+1, len(statsText), statsText, 
		d.colorScheme.GetStyle(d.colorScheme.Muted, false))
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

// drawBorder draws a border around the specified area
func (d *Display) drawBorder(x, y, width, height int) {
	borderStyle := d.colorScheme.GetStyle(d.colorScheme.Border, false)
	
	// Corners
	d.screen.SetContent(x, y, '┌', nil, borderStyle)                    // Top-left
	d.screen.SetContent(x+width-1, y, '┐', nil, borderStyle)           // Top-right
	d.screen.SetContent(x, y+height-1, '└', nil, borderStyle)           // Bottom-left
	d.screen.SetContent(x+width-1, y+height-1, '┘', nil, borderStyle) // Bottom-right
	
	// Horizontal lines
	for i := x + 1; i < x+width-1; i++ {
		d.screen.SetContent(i, y, '─', nil, borderStyle)         // Top
		d.screen.SetContent(i, y+height-1, '─', nil, borderStyle) // Bottom
	}
	
	// Vertical lines
	for i := y + 1; i < y+height-1; i++ {
		d.screen.SetContent(x, i, '│', nil, borderStyle)         // Left
		d.screen.SetContent(x+width-1, i, '│', nil, borderStyle) // Right
	}
}

// drawHorizontalLine draws a horizontal line
func (d *Display) drawHorizontalLine(x, y, width int, char string, color tcell.Color) {
	style := d.colorScheme.GetStyle(color, false)
	runes := []rune(char)
	if len(runes) == 0 {
		return
	}
	lineChar := runes[0]
	
	for i := 0; i < width; i++ {
		d.screen.SetContent(x+i, y, lineChar, nil, style)
	}
}