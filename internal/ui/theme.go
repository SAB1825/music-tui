package ui

import (
	"image/color"

	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

type theme struct {
	background, foreground                                                                                      color.Color
	base, sidebar, panel, inset, elevated, card, heading, muted, dim, accent, selected, button, hover, disabled lipgloss.Style
	focusBorder                                                                                                 color.Color
	covers                                                                                                      [4][4]string
}

func newTheme() theme {
	bg, fg := lipgloss.Color("#0D0D0D"), lipgloss.Color("#F2F2F2")
	base := lipgloss.NewStyle().Foreground(fg).Background(bg)
	surface := base.Background(lipgloss.Color("#181818"))
	elevated := base.Background(lipgloss.Color("#202020"))
	muted := lipgloss.Color("#B0B0B0")
	return theme{
		background: bg, foreground: fg, base: base,
		sidebar: base.Background(lipgloss.Color("#121212")),
		panel:   surface, elevated: elevated, card: surface,
		inset:       base,
		heading:     base.Bold(true),
		muted:       base.Foreground(muted),
		dim:         base.Foreground(lipgloss.Color("#858585")),
		accent:      base.Foreground(lipgloss.Color("#FFFFFF")).Bold(true),
		selected:    base.Background(lipgloss.Color("#343434")).Bold(true),
		button:      base.Background(lipgloss.Color("#FFFFFF")).Foreground(lipgloss.Color("#111111")).Bold(true),
		hover:       base.Background(lipgloss.Color("#282828")),
		disabled:    base.Foreground(lipgloss.Color("#606060")),
		focusBorder: lipgloss.Color("#777777"),
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

func (t theme) inputStyles() textinput.Styles {
	s := textinput.DefaultDarkStyles()
	muted := lipgloss.NewStyle().Foreground(t.muted.GetForeground())
	s.Focused.Placeholder, s.Blurred.Placeholder = muted, muted
	s.Focused.Suggestion, s.Blurred.Suggestion = muted, muted
	s.Focused.Prompt, s.Blurred.Prompt = muted, muted
	s.Blurred.Text = muted
	s.Cursor.Color = t.foreground
	return s
}
