package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

type menuRow struct {
	name    string
	page    page
	section bool
}

func menuRows() []menuRow {
	return []menuRow{
		{name: "BROWSE", section: true}, {page: home, name: "⌂  Home"}, {page: library, name: "▤  Library"},
		{page: radio, name: "○  Radio · off"}, {page: downloads, name: "↓  Downloads · off"},
		{name: "PLAYLISTS", section: true}, {page: createPlaylist, name: "+  Create playlist · off"},
		{page: liked, name: "♥  Liked Songs"}, {page: onRepeat, name: "On Repeat"}, {page: recent, name: "Recently Played"},
		{page: night, name: "Night Drive"}, {page: daybreak, name: "Daybreak Tapes"},
	}
}

func (m model) sidebarPanel(width, height int) string {
	t := m.theme
	if height < 3 {
		return strings.Join(padRows([]string{label(t.selected, width, "› "+m.sidebar.String())}, t.sidebar, width, height), "\n")
	}
	menu := menuRows()
	visible := height - 2
	first := 0
	if m.focus == sidebarFocus {
		for i, row := range menu {
			if !row.section && row.page == m.sidebar {
				first = max(0, i-visible+1)
				break
			}
		}
	}
	rows := make([]string, 0, height)
	for _, row := range menu[first:min(len(menu), first+visible)] {
		if row.section {
			rows = append(rows, label(t.dim.Background(t.sidebar.GetBackground()).Bold(true), width, "  "+row.name))
		} else {
			rows = append(rows, m.menuItem(width, row))
		}
	}
	rows = padRows(rows, t.sidebar, width, visible)
	rows = append(rows, line(t.rule.Background(t.sidebar.GetBackground()), width, strings.Repeat("─", width)),
		m.menuItem(width, menuRow{page: settings, name: "⚙  Settings"}))
	return strings.Join(rows, "\n")
}

func (m model) menuItem(width int, row menuRow) string {
	t := m.theme
	style, mark := t.sidebar, "  "
	if row.page.unavailable() {
		style = t.dim.Background(t.sidebar.GetBackground())
	}
	if row.page == m.page {
		style = t.elevated.Bold(true)
	}
	if m.focus == sidebarFocus && row.page == m.sidebar {
		style, mark = t.selected, "› "
	}
	if row.page >= onRepeat && row.page <= daybreak && width >= 14 {
		cover := m.cover(int(row.page-onRepeat), 3, 1)
		return lipgloss.JoinHorizontal(lipgloss.Top, line(style, 2, mark), cover,
			label(style, width-5, "  "+row.name))
	}
	return label(style, width, mark+row.name)
}
