package ui

import (
	"strings"
)

const toolbarPrefix = 25
const toolbarSuffix = 22

func (m model) searchWidth() int {
	if m.width >= 72 && m.layout().toolbar == 3 {
		return m.width - toolbarPrefix - toolbarSuffix
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
	back, forward := line(t.disabled, 4, " [‹"), line(t.disabled, 4, " ]›")
	y := 1
	wide := l.toolbar == 3 && w >= 72
	if !wide {
		y = 0
	}
	if w >= toolbarPrefix {
		if m.historyIndex > 0 {
			back = m.control(t.base, backAction, 17, y, 4, " [‹")
		}
		if m.historyIndex+1 < len(m.history) {
			forward = m.control(t.base, forwardAction, 21, y, 4, " ]›")
		}
	}
	brand := line(t.heading, 17, "  ◆ AFTERHOURS")
	if l.sidebar == 0 && w >= 12 {
		brand = m.control(t.heading, menuAction, 0, y, 17, "  ≡ AFTERHOURS")
	}
	if !wide {
		top := brand + back + forward
		if w >= 37 {
			m.region(navigateAction, int(settings), 25, 0, 12, 1)
			top += label(m.hoverStyle(t.base, navigateAction, int(settings)), 12, "  Settings")
		}
		m.region(searchAction, 0, 0, 1, w, 1)
		rows := []string{label(t.base, w, top), label(m.hoverStyle(style, searchAction, 0), w, search)}
		if l.toolbar == 3 {
			rows = append(rows, line(t.base, w, ""))
		}
		return strings.Join(rows, "\n")
	}
	sw := m.searchWidth()
	m.region(searchAction, 0, toolbarPrefix, 1, sw, 1)
	sx := toolbarPrefix + sw + 10
	m.region(navigateAction, int(settings), sx, 1, 12, 1)
	return strings.Join([]string{
		line(t.base, w, ""),
		brand + back + forward + label(m.hoverStyle(style, searchAction, 0), sw, search) +
			label(t.dim, 10, "  "+m.providerName()) + line(m.hoverStyle(t.base, navigateAction, int(settings)), 12, "  Settings"),
		line(t.base, w, ""),
	}, "\n")
}
