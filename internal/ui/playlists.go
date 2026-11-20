package ui

import "slices"

type sessionPlaylist struct {
	name    string
	track   int
	deleted bool
}

func (m model) pageName(p page) string {
	if p >= firstPlaylist {
		return m.playlists[p-firstPlaylist].name
	}
	return p.String()
}

func (m *model) addPlaylist(name string) {
	p := firstPlaylist + page(len(m.playlists))
	m.playlists = append(slices.Clone(m.playlists), sessionPlaylist{name: name, track: m.player.Index()})
	m.open(p)
}

func (m *model) removePlaylist() {
	removed := m.page
	if removed < firstPlaylist {
		return
	}
	m.playlists = slices.Clone(m.playlists)
	m.playlists[removed-firstPlaylist].deleted = true
	history := slices.DeleteFunc(slices.Clone(m.history[:m.historyIndex+1]), func(p page) bool { return p == removed })
	if len(history) == 0 {
		history = []page{home}
	}
	m.history, m.historyIndex = history, len(history)-1
	m.page = history[m.historyIndex]
	m.sidebar = m.page
	m.resetPage()
}
