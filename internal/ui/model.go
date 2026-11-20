// Package ui implements the terminal interface.
package ui

import (
	"slices"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/SAB1825/music-tui/internal/domain"
	"github.com/SAB1825/music-tui/internal/playback"
)

type page int

const (
	home page = iota
	library
)

func (p page) String() string {
	if p == library {
		return "Library"
	}
	return "Home"
}

type focus int

const (
	tracksFocus focus = iota
	sidebarFocus
	searchFocus
)

type playbackTick uint64

type model struct {
	catalog          []domain.Track
	player           playback.Simulation
	search           textinput.Model
	theme            theme
	width, height    int
	page, sidebar    page
	focus            focus
	selected, offset int
	help             bool
	clock            uint64
}

// New builds the application model with the supplied in-memory library.
func New(tracks []domain.Track) tea.Model {
	search := textinput.New()
	search.Prompt = ""
	search.Placeholder = "Search your library"
	search.CharLimit = 128
	return model{
		catalog: slices.Clone(tracks),
		player:  playback.NewSimulation(tracks),
		search:  search,
		theme:   newTheme(),
	}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.search.SetWidth(max(1, m.width-4))
		m.search.SetCursor(m.search.Position())
	case playbackTick:
		if uint64(msg) == m.clock && m.player.Playing() {
			m.player.Advance(time.Second)
			if m.player.Playing() {
				cmd = tick(m.clock)
			}
		}
	case tea.KeyPressMsg:
		return m.updateKey(msg)
	default:
		if m.focus == searchFocus {
			cmd = m.updateSearch(msg)
		}
	}
	return m, cmd
}

func tick(clock uint64) tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg { return playbackTick(clock) })
}

func (m model) updateKey(key tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	name := key.String()
	if name == "ctrl+c" {
		return m, tea.Quit
	}
	if m.focus == searchFocus {
		switch name {
		case "esc":
			m.search.Reset()
			m.selected, m.offset = 0, 0
			fallthrough
		case "enter", "tab", "shift+tab":
			m.search.Blur()
			m.focus = tracksFocus
			return m, nil
		}
		cmd := m.updateSearch(key)
		return m, cmd
	}
	if name == "q" {
		return m, tea.Quit
	}
	if m.help {
		if name == "esc" || name == "?" {
			m.help = false
		}
		return m, nil
	}

	playing := m.player.Playing()
	restart := false
	switch name {
	case "?":
		m.help = true
	case "esc":
		m.search.Reset()
		m.selected, m.offset = 0, 0
		m.focus = tracksFocus
	case "/":
		m.focus = searchFocus
		cmd := m.search.Focus()
		return m, cmd
	case "tab", "shift+tab":
		if m.focus == sidebarFocus {
			m.focus = tracksFocus
		} else {
			m.focus = sidebarFocus
			m.sidebar = m.page
		}
	case "up", "k":
		m.move(-1)
	case "down", "j":
		m.move(1)
	case "enter":
		if m.focus == sidebarFocus {
			m.page, m.focus = m.sidebar, tracksFocus
			m.search.Reset()
			m.selected, m.offset = 0, 0
		} else if matches := m.matches(); m.selected < len(matches) {
			m.player.Play(matches[m.selected])
			restart = true
		}
	case "space":
		m.player.Toggle()
	case "n":
		m.player.Skip(1)
		restart = playing
	case "p":
		m.player.Skip(-1)
		restart = playing
	case "+", "=":
		m.player.AdjustVolume(5)
	case "-":
		m.player.AdjustVolume(-5)
	}
	if restart || m.player.Playing() != playing {
		// A tick already in flight must not advance a new play session.
		m.clock++
		if m.player.Playing() {
			return m, tick(m.clock)
		}
	}
	return m, nil
}

func (m *model) updateSearch(msg tea.Msg) tea.Cmd {
	before := m.search.Value()
	var cmd tea.Cmd
	m.search, cmd = m.search.Update(msg)
	if before != m.search.Value() {
		m.selected, m.offset = 0, 0
	}
	return cmd
}

func (m *model) move(step int) {
	if m.focus == sidebarFocus {
		m.sidebar = page(min(int(library), max(int(home), int(m.sidebar)+step)))
		return
	}
	m.selected = min(max(0, len(m.matches())-1), max(0, m.selected+step))
	if m.selected < m.offset {
		m.offset = m.selected
	} else if rows := m.trackRows(m.bodyHeight()); rows > 0 && m.selected >= m.offset+rows {
		m.offset = m.selected - rows + 1
	}
}

func (m model) matches() []int {
	query := strings.ToLower(strings.TrimSpace(m.search.Value()))
	var matches []int
	for i, track := range m.catalog {
		if strings.Contains(strings.ToLower(track.Title+" "+track.Artist+" "+track.Album), query) {
			matches = append(matches, i)
		}
	}
	return matches
}
