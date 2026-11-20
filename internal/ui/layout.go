package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const (
	wideWidth           = 96
	sidebarWidth        = 66
	playerFullHeight    = 10
	playerCompactHeight = 5
)

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
	if m.bodyHeight() == 0 {
		return strings.Repeat("\n", m.height-m.playerHeight()) + m.playerBar()
	}
	t := m.theme
	header := line(t.panel.Bold(true), m.width, "  ◆  AFTERHOURS MUSIC")
	search := "  /  " + m.search.Placeholder
	if m.focus == searchFocus || m.search.Value() != "" {
		search = "  / " + m.search.View()
	}
	searchStyle := t.muted
	if m.focus == searchFocus {
		searchStyle = t.selected
	}
	width := m.width
	if m.hasSidebar() {
		width -= m.sidebarSize()
	}
	body := m.mainPanel(width, m.bodyHeight())
	if m.hasSidebar() {
		body = lipgloss.JoinHorizontal(lipgloss.Top, m.sidebarPanel(m.sidebarSize(), m.bodyHeight()), body)
	}
	return strings.Join([]string{header, line(searchStyle, m.width, search), body, m.playerBar()}, "\n")
}

func (m model) playerHeight() int {
	if m.height >= playerFullHeight {
		return 4
	}
	if m.height >= playerCompactHeight {
		return 2
	}
	return 1
}

func (m model) bodyHeight() int {
	return max(0, m.height-2-m.playerHeight())
}

func (m model) hasSidebar() bool {
	return m.width >= sidebarWidth && m.bodyHeight() >= 4
}

func (m model) sidebarSize() int {
	if m.width >= wideWidth {
		return 22
	}
	return 15
}

func (m model) sidebarPanel(width, height int) string {
	t := m.theme
	rows := []string{line(t.panel.Bold(true), width, "  BROWSE"), line(t.panel, width, "")}
	for _, p := range []page{home, library} {
		style, mark := t.panel, " "
		if p == m.page {
			style, mark = t.panel.Bold(true), "•"
		}
		if m.focus == sidebarFocus && p == m.sidebar {
			style, mark = t.selected, "›"
		}
		rows = append(rows, line(style, width, fmt.Sprintf("  %s %s", mark, p)))
	}
	return lipgloss.NewStyle().Height(height).Background(t.panel.GetBackground()).Render(strings.Join(rows, "\n"))
}

func (m model) hasFeature(height int) bool {
	return m.page == home && len(m.catalog) > 0 && m.search.Value() == "" && height >= 11
}

func (m model) trackRows(height int) int {
	rows := height - 2
	if m.hasFeature(height) {
		rows -= 5
	}
	if !m.hasSidebar() && m.focus == sidebarFocus && height >= 3 {
		rows--
	}
	return max(0, rows)
}

func (m model) mainPanel(width, height int) string {
	t := m.theme
	if m.help {
		return m.helpPanel(width, height)
	}
	title := m.page.String()
	if m.search.Value() != "" {
		title = "Search results"
	}
	matches := m.matches()
	rows := []string{line(t.heading, width, fmt.Sprintf("  %s  /  %d tracks", title, len(matches)))}
	if !m.hasSidebar() && m.focus == sidebarFocus {
		if height == 1 {
			return line(t.selected, width, "› "+m.sidebar.String()+"  ↑↓ ↵")
		}
		if height >= 3 {
			rows = append(rows, line(t.selected, width, "  › "+m.sidebar.String()+"  ↑↓ choose  ·  ↵ open"))
		}
	}
	if m.hasFeature(height) {
		track := m.catalog[0]
		rows = append(rows,
			line(t.base, width, ""),
			line(t.featured.Bold(true), width, "  FEATURED"),
			line(t.featured.Bold(true), width, "  "+track.Title),
			line(t.featured, width, "  "+track.Artist+"  /  "+track.Album),
			line(t.base, width, ""),
		)
	}
	visible := m.trackRows(height)
	start := max(0, min(m.offset, m.selected))
	if visible > 0 {
		start = min(max(start, m.selected-visible+1), max(0, len(matches)-visible))
	}
	if len(matches) == 0 && visible > 0 {
		rows = append(rows, line(t.muted, width, "  No matching tracks"))
	}
	for i := start; i < min(len(matches), start+visible); i++ {
		rows = append(rows, m.trackLine(width, i, m.catalog[matches[i]]))
	}
	for len(rows) < height-1 {
		rows = append(rows, line(t.base, width, ""))
	}
	if len(rows) < height {
		rows = append(rows, line(t.muted, width, "  ↑↓ Browse  ·  ↵ Play  ·  Tab Menu  ·  ? Help"))
	}
	return strings.Join(rows, "\n")
}

func (m model) helpPanel(width, height int) string {
	lines := []string{
		"  Keyboard shortcuts", "  ↑↓ / j k    Move selection", "  Enter       Open or play",
		"  Space       Play / pause", "  /           Search local library", "  Tab         Switch menu / tracks",
		"  n / p       Next / previous", "  + / -       Volume", "  Esc / ?     Close help", "  q / Ctrl+C  Quit",
	}
	rows := make([]string, height)
	for i := range rows {
		if i < len(lines) {
			rows[i] = line(m.theme.panel, width, lines[i])
		} else {
			rows[i] = line(m.theme.panel, width, "")
		}
	}
	return strings.Join(rows, "\n")
}
