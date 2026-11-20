package ui

import (
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
)

type menuRow struct {
	name    string
	page    page
	section bool
}

func (m model) menuRows() []menuRow {
	rows := []menuRow{
		{name: "BROWSE", section: true}, {page: home, name: "⌂  Home"}, {page: library, name: "▤  Library"},
		{page: radio, name: "○  Radio · off"}, {page: downloads, name: "↓  Downloads · off"},
		{name: "PLAYLISTS", section: true}, {page: createPlaylist, name: "+  Create playlist"},
		{page: liked, name: "♥  Liked Songs"}, {page: onRepeat, name: "On Repeat"}, {page: recent, name: "Recently Played"},
		{page: night, name: "Night Drive"}, {page: daybreak, name: "Daybreak Tapes"},
	}
	for i, p := range m.playlists {
		if !p.deleted {
			rows = append(rows, menuRow{name: p.name, page: firstPlaylist + page(i)})
		}
	}
	return rows
}

func (m model) sidebarPanel(width, height int) string {
	t := m.theme
	if height < 3 {
		return strings.Join(padRows([]string{label(t.selected, width, "› "+m.pageName(m.sidebar))}, t.sidebar, width, height), "\n")
	}
	menu := m.menuRows()
	visible := height - 2
	first := clampScroll(m.menuOffset, len(menu), visible)
	rows := make([]string, 0, height)
	for _, row := range menu[first:min(len(menu), first+visible)] {
		if row.section {
			rows = append(rows, label(t.dim.Background(t.sidebar.GetBackground()).Bold(true), width, "  "+row.name))
		} else {
			if !row.page.unavailable() {
				m.region(navigateAction, int(row.page), 0, m.layout().toolbar+len(rows), width, 1)
			}
			rows = append(rows, m.menuItem(width, row))
		}
	}
	rows = padRows(rows, t.sidebar, width, visible)
	m.region(navigateAction, int(settings), 0, m.layout().toolbar+height-1, width, 1)
	rows = append(rows, line(t.sidebar, width, ""),
		m.menuItem(width, menuRow{page: settings, name: "⚙  Settings"}))
	return strings.Join(rows, "\n")
}

func (m model) menuItem(width int, row menuRow) string {
	t := m.theme
	style, mark := t.sidebar, "  "
	if row.page.unavailable() {
		style = t.disabled.Background(t.sidebar.GetBackground())
	}
	if row.page == m.page {
		style = t.elevated.Bold(true)
	}
	if m.focus == sidebarFocus && row.page == m.sidebar {
		style, mark = t.selected, "› "
	}
	style = m.hoverStyle(style, navigateAction, int(row.page))
	if row.page >= onRepeat && row.page <= daybreak && width >= 14 {
		cover := m.cover(int(row.page-onRepeat), 3, 1)
		return lipgloss.JoinHorizontal(lipgloss.Top, line(style, 2, mark), cover,
			label(style, width-5, "  "+row.name))
	}
	return label(style, width, mark+row.name)
}

func (m *model) revealMenuSelection() {
	if m.sidebar == queue {
		m.sidebar = home
	}
	for i, row := range m.menuRows() {
		if !row.section && row.page == m.sidebar {
			visible := max(1, m.bodyHeight()-2)
			m.menuOffset = min(i, max(m.menuOffset, i-visible+1))
			return
		}
	}
}

func (m *model) moveMenu(step int) {
	var pages []page
	for _, row := range m.menuRows() {
		if !row.section {
			pages = append(pages, row.page)
		}
	}
	pages = append(pages, settings)
	i := slices.Index(pages, m.sidebar)
	m.sidebar = pages[min(len(pages)-1, max(0, i+step))]
	m.revealMenuSelection()
}
