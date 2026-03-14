package main

import "github.com/charmbracelet/lipgloss"

var (
	colorPrimary = lipgloss.Color("#7D56F4")
	colorBright  = lipgloss.Color("#FAFAFA")
	colorDim     = lipgloss.Color("#999999")
	colorMuted   = lipgloss.Color("#555555")
	colorGreen   = lipgloss.Color("#73D216")
	colorYellow  = lipgloss.Color("#FFC107")
	colorRed     = lipgloss.Color("#FF4444")
	colorCyan    = lipgloss.Color("#6C9EFF")

	titleBarStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorBright).
			Background(colorPrimary).
			Padding(0, 1)

	sectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			MarginTop(1)

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorBright).
			Background(colorPrimary)

	branchStyle = lipgloss.NewStyle().
			Foreground(colorCyan)

	dimStyle = lipgloss.NewStyle().
			Foreground(colorDim)

	mutedStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Italic(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorRed)

	labelStyle = lipgloss.NewStyle().
			Foreground(colorBright).
			Width(12)
)

func barColor(pct float64) lipgloss.Color {
	switch {
	case pct > 50:
		return colorGreen
	case pct > 20:
		return colorYellow
	default:
		return colorRed
	}
}
