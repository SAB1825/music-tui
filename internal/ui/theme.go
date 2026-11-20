package ui

import (
	"image/color"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

type theme struct {
	background, foreground                                                                                              color.Color
	base, sidebar, panel, inset, elevated, featured, card, heading, muted, dim, accent, selected, rule, button, warning lipgloss.Style
	covers                                                                                                              [4][4]string
}

func newTheme() theme {
	bg, fg := lipgloss.Color("#15100E"), lipgloss.Color("#F4E6D2")
	base := lipgloss.NewStyle().Foreground(fg).Background(bg)
	return theme{
		background: bg, foreground: fg, base: base,
		sidebar:  base.Background(lipgloss.Color("#211813")),
		panel:    base.Background(lipgloss.Color("#2A1F19")),
		elevated: base.Background(lipgloss.Color("#38291F")),
		featured: base.Background(lipgloss.Color("#423516")),
		card:     base.Background(lipgloss.Color("#351D26")),
		inset:    base.Background(lipgloss.Color("#180F11")),
		heading:  base.Bold(true),
		muted:    base.Foreground(lipgloss.Color("#BCA991")),
		dim:      base.Foreground(lipgloss.Color("#A48D73")),
		accent:   base.Foreground(lipgloss.Color("#F0CB79")).Bold(true),
		selected: base.Background(lipgloss.Color("#59402D")).Bold(true),
		rule:     base.Foreground(lipgloss.Color("#49372B")),
		button:   base.Background(lipgloss.Color("#F0CB79")).Foreground(bg).Bold(true),
		warning:  base.Foreground(lipgloss.Color("#D47B73")),
		covers: [4][4]string{
			{"#FCE2A3", "#DC9C91", "#995670", "#41233B"},
			{"#BEE0D0", "#8AB9AC", "#4B777C", "#243448"},
			{"#FFD4AE", "#E69C68", "#AB544B", "#4B2637"},
			{"#E4D2F1", "#B3A0D0", "#78638E", "#372B50"},
		},
	}
}

func line(style lipgloss.Style, width int, value string) string {
	if width <= 0 {
		return ""
	}
	return style.Width(width).Render(ansi.Truncate(value, width, ""))
}

func label(style lipgloss.Style, width int, value string) string {
	return line(style, width, ansi.Truncate(value, width, "…"))
}
