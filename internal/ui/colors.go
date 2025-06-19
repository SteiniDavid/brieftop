package ui

import (
	"cpu-monitor/internal/monitor"

	"github.com/gdamore/tcell/v2"
)

type ColorScheme struct {
	Background   tcell.Color
	Text         tcell.Color
	Header       tcell.Color
	LowUsage     tcell.Color
	MediumUsage  tcell.Color
	HighUsage    tcell.Color
	Selected     tcell.Color
	Thread       tcell.Color
	ChildProcess tcell.Color
}

func NewColorScheme() *ColorScheme {
	return &ColorScheme{
		Background:   tcell.ColorBlack,
		Text:         tcell.ColorWhite,
		Header:       tcell.ColorYellow,
		LowUsage:     tcell.ColorGreen,
		MediumUsage:  tcell.ColorYellow,
		HighUsage:    tcell.ColorRed,
		Selected:     tcell.ColorBlue,
		Thread:       tcell.ColorGray,
		ChildProcess: tcell.ColorTeal,
	}
}

func (cs *ColorScheme) GetProcessColor(level monitor.ResourceLevel) tcell.Color {
	switch level {
	case monitor.Low:
		return cs.LowUsage
	case monitor.Medium:
		return cs.MediumUsage
	case monitor.High:
		return cs.HighUsage
	default:
		return cs.Text
	}
}

func (cs *ColorScheme) GetStyle(color tcell.Color, selected bool) tcell.Style {
	style := tcell.StyleDefault.Foreground(color).Background(cs.Background)
	if selected {
		style = style.Background(cs.Selected)
	}
	return style
}