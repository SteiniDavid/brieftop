package ui

import (
	"github.com/SteiniDavid/brieftop/internal/monitor"
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
	Border       tcell.Color
	Accent       tcell.Color
	Muted        tcell.Color
	Success      tcell.Color
	Warning      tcell.Color
	Error        tcell.Color
}

func NewColorScheme() *ColorScheme {
	return &ColorScheme{
		Background:   tcell.NewRGBColor(15, 15, 25),    // Dark navy background
		Text:         tcell.NewRGBColor(220, 225, 235), // Light gray text
		Header:       tcell.NewRGBColor(100, 200, 255), // Bright blue header
		LowUsage:     tcell.NewRGBColor(80, 200, 120),  // Vibrant green
		MediumUsage:  tcell.NewRGBColor(255, 180, 50),  // Warm orange
		HighUsage:    tcell.NewRGBColor(255, 85, 85),   // Bright red
		Selected:     tcell.NewRGBColor(70, 130, 255),  // Bright blue selection
		Thread:       tcell.NewRGBColor(150, 160, 180), // Muted gray for threads
		ChildProcess: tcell.NewRGBColor(120, 200, 200), // Cyan for child processes
		Border:       tcell.NewRGBColor(60, 70, 90),    // Subtle border color
		Accent:       tcell.NewRGBColor(200, 120, 255), // Purple accent
		Muted:        tcell.NewRGBColor(120, 130, 140), // Muted text
		Success:      tcell.NewRGBColor(50, 255, 120),  // Bright success green
		Warning:      tcell.NewRGBColor(255, 200, 50),  // Warning yellow
		Error:        tcell.NewRGBColor(255, 100, 100), // Error red
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

// GetProgressBarColor returns color based on percentage
func (cs *ColorScheme) GetProgressBarColor(percent float64) tcell.Color {
	if percent >= 75 {
		return cs.HighUsage
	} else if percent >= 50 {
		return cs.MediumUsage
	} else if percent >= 25 {
		return cs.Warning
	}
	return cs.LowUsage
}

// CreateProgressBar creates a visual progress bar string
func CreateProgressBar(percent float64, width int) string {
	if width < 2 {
		return ""
	}

	filledWidth := int((percent / 100.0) * float64(width))
	if filledWidth > width {
		filledWidth = width
	}

	bar := ""
	for i := 0; i < width; i++ {
		if i < filledWidth {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	return bar
}

// GetStatusIcon returns an appropriate icon for process status
func GetStatusIcon(cpuPercent float64, isExpanded bool, hasChildren bool) string {
	if hasChildren {
		if isExpanded {
			return "▼"
		} else {
			return "▶"
		}
	}

	if cpuPercent >= 50 {
		return "●" // High CPU
	} else if cpuPercent >= 20 {
		return "●" // Medium CPU
	} else if cpuPercent >= 5 {
		return "●" // Active
	}
	return "○" // Low activity
}
