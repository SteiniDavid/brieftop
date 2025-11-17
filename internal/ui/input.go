package ui

import (
	"github.com/gdamore/tcell/v2"
)

type InputHandler struct {
	display *Display
}

func NewInputHandler(display *Display) *InputHandler {
	return &InputHandler{
		display: display,
	}
}

func (ih *InputHandler) HandleInput(ev *tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyEscape, tcell.KeyCtrlC:
		return false
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'q', 'Q':
			return false
		case ' ':
			ih.display.TogglePause()
		case 'r', 'R':
			ih.display.ForceRefresh()
		}
	case tcell.KeyUp:
		ih.display.MoveCursor(-1)
	case tcell.KeyDown:
		ih.display.MoveCursor(1)
	case tcell.KeyEnter:
		ih.display.ToggleExpanded()
	case tcell.KeyHome:
		ih.display.SetCursor(0)
	case tcell.KeyEnd:
		ih.display.SetCursor(-1)
	}
	return true
}

func (d *Display) TogglePause() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.paused = !d.paused
}

func (d *Display) ForceRefresh() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.forceRefresh = true
}

func (d *Display) MoveCursor(delta int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.processes) == 0 {
		return
	}

	newPos := d.selectedIndex + delta
	if newPos < 0 {
		newPos = len(d.processes) - 1
	} else if newPos >= len(d.processes) {
		newPos = 0
	}
	d.selectedIndex = newPos
	d.adjustScrollOffset()
}

func (d *Display) SetCursor(pos int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.processes) == 0 {
		return
	}

	if pos < 0 {
		d.selectedIndex = len(d.processes) - 1
	} else if pos >= len(d.processes) {
		d.selectedIndex = len(d.processes) - 1
	} else {
		d.selectedIndex = pos
	}
	d.adjustScrollOffset()
}

func (d *Display) ToggleExpanded() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.processes) == 0 || d.selectedIndex >= len(d.processes) {
		return
	}
	selectedProcess := d.processes[d.selectedIndex]
	d.monitor.ToggleExpanded(selectedProcess.PID)
}
