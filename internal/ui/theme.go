package ui

import (
	"image/color"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

type theme struct {
	background, foreground                                        color.Color
	base, panel, featured, heading, muted, accent, selected, rule lipgloss.Style
}

func newTheme() theme {
	background := lipgloss.Color("#17110e")
	foreground := lipgloss.Color("#f7e9d4")
	surface := lipgloss.Color("#231913")
	cream := lipgloss.NewStyle().Foreground(foreground)
	return theme{
		background: background,
		foreground: foreground,
		base:       cream,
		panel:      cream.Background(surface),
		featured:   cream.Background(lipgloss.Color("#3e2d1b")),
		heading:    cream.Bold(true),
		muted:      cream.Foreground(lipgloss.Color("#c0a88d")),
		accent:     cream.Foreground(lipgloss.Color("#f6cd7a")).Bold(true),
		selected:   cream.Background(lipgloss.Color("#513826")).Bold(true),
		rule:       cream.Foreground(lipgloss.Color("#74563c")),
	}
}

func line(style lipgloss.Style, width int, value string) string {
	return style.Width(width).Render(ansi.Truncate(value, width, ""))
}
