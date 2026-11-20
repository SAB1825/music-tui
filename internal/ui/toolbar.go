package ui

import (
	"strings"
)

func (m model) searchWidth() int {
	if m.width >= 72 && m.layout().toolbar == 3 {
		return m.width - 52
	}
	return m.width
}

func (m model) toolbar(l layout) string {
	t, w := m.theme, l.width
	search := "/  " + m.search.Placeholder
	style := t.elevated
	if m.focus == searchFocus || m.search.Value() != "" {
		search = "/  " + m.search.View()
	}
	if m.focus == searchFocus {
		style = t.selected
	}
	if l.toolbar < 3 || w < 72 {
		rows := []string{line(t.sidebar.Bold(true), w, "  ◆ AFTERHOURS    [ Back  ] Fwd   ·   , Settings"), label(style, w, search)}
		if l.toolbar == 3 {
			rows = append(rows, line(t.rule, w, strings.Repeat("─", w)))
		}
		return strings.Join(rows, "\n")
	}
	back, forward := t.dim, t.accent
	if m.historyIndex > 0 {
		back = t.accent
	}
	if m.historyIndex+1 == len(m.history) {
		forward = t.dim
	}
	searchW := m.searchWidth()
	toolbar := line(t.sidebar.Bold(true), 17, "  ◆ AFTERHOURS") +
		line(back.Background(t.sidebar.GetBackground()), 4, " [‹") +
		line(forward.Background(t.sidebar.GetBackground()), 4, " ]›") +
		label(style, searchW, search) +
		line(t.sidebar, 27, "  ◉ MOCK MUSIC   , Settings")
	return line(t.sidebar, w, "") + "\n" + toolbar + "\n" + line(t.rule, w, strings.Repeat("─", w))
}
