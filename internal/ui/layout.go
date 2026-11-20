package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type layout struct {
	width, toolbar, body, player, sidebar int
}

func calculateLayout(width, height int) layout {
	l := layout{width: max(0, width)}
	switch {
	case height >= 15:
		l.toolbar, l.player = 3, 5
	case height >= 10:
		l.toolbar, l.player = 2, 4
	case height >= 5:
		l.toolbar, l.player = 2, 2
	case height >= 1:
		l.player = 1
	}
	l.body = max(0, height-l.toolbar-l.player)
	if width >= 98 && l.body >= 8 {
		l.sidebar = width * 21 / 100
	}
	return l
}

func (l layout) content() int { return l.width - l.sidebar }

func (m model) layout() layout    { return calculateLayout(m.width, m.height) }
func (m model) bodyHeight() int   { return m.layout().body }
func (m model) playerHeight() int { return m.layout().player }

func (m model) View() tea.View {
	view := tea.NewView(m.render())
	view.AltScreen = true
	view.WindowTitle = "Afterhours Music"
	view.BackgroundColor = m.theme.background
	view.ForegroundColor = m.theme.foreground
	return view
}

func (m model) render() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	l := m.layout()
	if l.body == 0 || l.toolbar == 0 {
		return strings.Repeat("\n", m.height-l.player) + m.playerBar()
	}
	body := m.mainPanel(l.content(), l.body)
	if l.sidebar > 0 {
		body = lipgloss.JoinHorizontal(lipgloss.Top, m.sidebarPanel(l.sidebar, l.body), body)
	}
	return strings.Join([]string{m.toolbar(l), body, m.playerBar()}, "\n")
}

func (m model) mainPanel(width, height int) string {
	if m.help {
		return m.helpPanel(width, height)
	}
	if m.layout().sidebar == 0 && m.focus == sidebarFocus {
		return m.sidebarPanel(width, height)
	}
	if m.discoveryVisible() {
		return m.homePanel(width, height)
	}
	return m.libraryPanel(width, height)
}

func (m model) discoveryVisible() bool {
	return m.page == home && m.search.Value() == "" && m.bodyHeight() >= 16 && m.layout().content() >= 46
}

func (m model) trackRows(height int) int {
	if m.discoveryVisible() {
		return discoveryColumns(m.layout().content()) * homeDimensions(m.layout().content(), height).discoveryRows
	}
	return max(0, height-3)
}

func (m model) libraryPanel(width, height int) string {
	t := m.theme
	title := m.page.String()
	switch {
	case m.search.Value() != "":
		title = "Search results"
	case m.page == settings:
		return m.settingsPanel(width, height)
	case m.page.unavailable():
		return strings.Join(padRows([]string{
			line(t.heading, width, "  "+title),
			line(t.muted, width, "  Unavailable with the local mock catalog."),
			line(t.dim, width, "  [ Back  ·  Tab Browse"),
		}, t.base, width, height), "\n")
	}
	matches := m.matches()
	summary := line(t.muted, width, fmt.Sprintf("  %d tracks", len(matches)))
	if m.page == queue && m.search.Value() == "" {
		summary = label(t.muted, width, fmt.Sprintf("  %d tracks · Catalog order", len(matches)))
	}
	if m.page == home && m.search.Value() == "" && m.focus == categoryFocus {
		summary = label(t.selected, width, "  "+categories()[m.category]+"  ←→ Change filter")
	}
	rows := []string{line(t.heading, width, "  "+title), summary}
	visible := m.trackRows(height)
	start := min(max(min(m.offset, m.selected), m.selected-visible+1), max(0, len(matches)-visible))
	if len(matches) == 0 {
		rows = append(rows, line(t.muted, width, "  No tracks to show"))
	}
	for i := start; i < min(len(matches), start+visible); i++ {
		rows = append(rows, m.trackLine(width, i, m.catalog[matches[i]]))
	}
	rows = padRows(rows, t.base, width, max(0, height-1))
	return strings.Join(append(rows, line(t.dim, width, "  ↑↓ Browse  ·  Enter Play  ·  Tab Menu  ·  ? Help")), "\n")
}

func (m model) helpPanel(width, height int) string {
	lines := []string{
		"  Keyboard shortcuts", "  ↑↓ / j k    Move selection", "  ←→ / h l    Move across cards or filters", "  Enter       Open or play",
		"  Space       Play / pause", "  /           Search local library", "  Tab         Switch menu / tracks",
		"  c           Focus discovery filters", "  [ / ]       Back / forward", "  n / p       Next / previous",
		"  + / -       Volume", "  f / s / r   Favorite / shuffle / repeat", "  u / ,       Queue / settings",
		"  Esc / ?     Close help", "  q / Ctrl+C  Quit",
	}
	var rows []string
	for _, value := range lines {
		rows = append(rows, line(m.theme.panel, width, value))
	}
	return strings.Join(padRows(rows, m.theme.panel, width, height), "\n")
}

func padRows(rows []string, style lipgloss.Style, width, height int) []string {
	for len(rows) < height {
		rows = append(rows, line(style, width, ""))
	}
	return rows[:height]
}
