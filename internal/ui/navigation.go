package ui

import (
	tea "charm.land/bubbletea/v2"
	"slices"
	"strings"
)

type page int

const (
	home page = iota
	library
	radio
	downloads
	createPlaylist
	liked
	onRepeat
	recent
	night
	daybreak
	settings
	queue
	publicPlaylist
	providersPage
	firstPlaylist
)

func (p page) String() string {
	return [...]string{
		"Home", "Library", "Radio", "Downloads", "Create playlist", "Liked Songs",
		"On Repeat", "Recently Played", "Night Drive", "Daybreak Tapes", "Settings", "Queue", "Public playlist", "Music providers",
	}[p]
}

func (p page) unavailable() bool { return p == radio || p == downloads }

func (m *model) choosePage(p page) tea.Cmd {
	switch p {
	case publicPlaylist:
		return m.showDialog(dialogPublicPlaylist)
	case providersPage:
		return m.showDialog(dialogProviders)
	case home:
		if m.remote() {
			m.clearCatalog()
		}
		m.open(p)
		return nil
	case settings:
		return m.showDialog(dialogSettings)
	case createPlaylist:
		return m.showDialog(dialogPlaylist)
	default:
		m.open(p)
		return nil
	}
}

func (m *model) open(p page) {
	if p == m.page {
		m.resetPage()
		return
	}
	m.history = append(slices.Clone(m.history[:m.historyIndex+1]), p)
	m.historyIndex++
	m.page = p
	m.resetPage()
}

func (m *model) travel(step int) {
	next := m.historyIndex + step
	if next >= 0 && next < len(m.history) {
		m.historyIndex = next
		m.page = m.history[next]
		m.resetPage()
	}
}

func (m *model) resetPage() {
	m.search.Reset()
	m.search.Blur()
	m.focus = tracksFocus
	m.selected, m.offset = 0, 0
}

func categories() []string {
	return []string{"All", "Energize", "Sleep", "Sad", "Romance", "Commute", "Focus", "Party"}
}

func moodTrack(category, index int) bool {
	masks := [...]uint8{0b11111111, 0b01000100, 0b10100010, 0b10010000, 0b00101000, 0b01010001, 0b00100110, 0b01000101}
	return category == 0 || masks[category]&(1<<uint(index%8)) != 0
}

func (m model) matches() []int {
	if m.remote() {
		indices := make([]int, len(m.online.tracks))
		for i := range indices {
			indices[i] = i
		}
		return indices
	}
	query := strings.ToLower(strings.TrimSpace(m.search.Value()))
	if query == "" && m.page >= firstPlaylist {
		if p := m.playlists[m.page-firstPlaylist]; p.track >= 0 && !p.deleted {
			return []int{p.track}
		}
		return nil
	}
	if query == "" && (m.page == settings || m.page.unavailable()) {
		return nil
	}
	if query == "" && m.page == recent {
		return m.recent
	}
	var matches []int
	for i, track := range m.catalog {
		if query != "" {
			if strings.Contains(strings.ToLower(track.Title+" "+track.Artist+" "+track.Album), query) {
				matches = append(matches, i)
			}
			continue
		}
		switch m.page {
		case home:
			if !moodTrack(m.category, i) {
				continue
			}
		case liked:
			if !slices.Contains(m.likes, i) {
				continue
			}
		case onRepeat:
			if i%3 != 0 {
				continue
			}
		case night:
			if i%4 > 1 {
				continue
			}
		case daybreak:
			if i%4 < 2 {
				continue
			}
		}
		matches = append(matches, i)
	}
	return matches
}

func (m *model) rememberTrack() {
	if !m.player.Playing() {
		return
	}
	i := m.player.Index()
	if len(m.recent) > 0 && m.recent[0] == i {
		return
	}
	m.recent = append([]int{i}, slices.DeleteFunc(slices.Clone(m.recent), func(previous int) bool { return previous == i })...)
}

func (m *model) toggleFavorite() {
	i := m.player.Index()
	if i < 0 {
		return
	}
	m.likes = slices.Clone(m.likes)
	if existing := slices.Index(m.likes, i); existing >= 0 {
		m.likes = slices.Delete(m.likes, existing, existing+1)
	} else {
		m.likes = append(m.likes, i)
	}
	if m.page == liked {
		m.selected = min(m.selected, max(0, len(m.matches())-1))
	}
}

func (m model) favorite() string {
	if slices.Contains(m.likes, m.player.Index()) {
		return "♥"
	}
	return "♡"
}
