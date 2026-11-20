// Package ui implements the terminal interface.
package ui

import (
	"slices"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/SAB1825/music-tui/internal/config"
	"github.com/SAB1825/music-tui/internal/domain"
	"github.com/SAB1825/music-tui/internal/playback"
	"github.com/SAB1825/music-tui/internal/visualizer"
)

type focus int

const (
	tracksFocus focus = iota
	sidebarFocus
	searchFocus
	categoryFocus
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
	dialog           dialog
	menuOffset       int
	playlists        []sessionPlaylist
	hover            action
	screen           screenFrame
	drawing          *screenFrame
	clock            uint64
	history          []page
	historyIndex     int
	category         int
	likes, recent    []int
	covers           map[coverSize][4]string
	frame            []float64
	spectrumEnabled  bool
	spectrumError    string
	spectrumConfig   config.Visualizer
	spectrumEvents   <-chan visualizer.Event
}

// New builds the application model with the supplied in-memory library.
func New(tracks []domain.Track) tea.Model {
	theme := newTheme()
	search := textinput.New()
	search.SetStyles(theme.inputStyles())
	search.Prompt = ""
	search.Placeholder = "Search your library"
	search.CharLimit = 128
	return model{
		catalog: slices.Clone(tracks),
		player:  playback.NewSimulation(tracks),
		search:  search,
		theme:   theme,
		history: []page{home},
	}
}

func (m model) Init() tea.Cmd { return receiveSpectrum(m.spectrumEvents) }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	next, cmd := m.update(msg)
	if _, motion := msg.(tea.MouseMotionMsg); !motion || next.hover != m.hover {
		next.screen = next.draw()
	}
	return next, cmd
}

func (m model) update(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.search.SetWidth(max(1, m.searchWidth()-4))
		m.search.SetCursor(m.search.Position())
		m.prepareCovers()
		m.hover = action{}
		m.menuOffset = clampScroll(m.menuOffset, len(m.menuRows()), m.bodyHeight()-2)
		m.dialog.input.SetWidth(max(1, min(64, m.width)-8))
	case playbackTick:
		if uint64(msg) == m.clock && m.player.Playing() {
			m.player.Advance(time.Second)
			m.rememberTrack()
			if m.player.Playing() {
				cmd = tick(m.clock)
			}
		}
	case spectrumEvent:
		if !msg.open {
			m.spectrumEvents = nil
			m.frame = nil
			break
		}
		m.frame = msg.Frame
		if msg.Err != nil {
			m.spectrumError = msg.Err.Error()
		}
		cmd = receiveSpectrum(m.spectrumEvents)
	case tea.MouseMsg:
		return m.updateMouse(msg)
	case tea.KeyPressMsg:
		m.hover = action{}
		return m.updateKey(msg)
	default:
		if m.dialog.kind != dialogNone {
			return m.updateDialogInput(msg)
		}
		if m.focus == searchFocus {
			cmd = m.updateSearch(msg)
		}
	}
	return m, cmd
}

func tick(clock uint64) tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg { return playbackTick(clock) })
}

func (m model) updateKey(key tea.KeyPressMsg) (model, tea.Cmd) {
	name := key.String()
	if name == "ctrl+c" {
		return m, tea.Quit
	}
	if m.dialog.kind != dialogNone {
		return m.dialogKey(key)
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

	playing := m.player.Playing()
	restart := false
	switch name {
	case "?":
		cmd := m.showDialog(dialogHelp)
		return m, cmd
	case "esc":
		if m.search.Value() == "" && m.focus == tracksFocus {
			m.travel(-1)
		} else {
			m.resetPage()
		}
	case "[":
		m.travel(-1)
	case "]":
		m.travel(1)
	case ",":
		cmd := m.showDialog(dialogSettings)
		return m, cmd
	case "delete":
		if m.page >= firstPlaylist && m.search.Value() == "" {
			cmd := m.showDialog(dialogDelete)
			return m, cmd
		}
	case "u":
		m.open(queue)
	case "f":
		m.toggleFavorite()
	case "s":
		m.player.ToggleShuffle()
	case "r":
		m.player.ToggleRepeat()
	case "c":
		if m.page == home && m.search.Value() == "" {
			m.focus = categoryFocus
		}
	case "/":
		m.focus = searchFocus
		cmd := m.search.Focus()
		return m, cmd
	case "tab", "shift+tab":
		if m.focus == sidebarFocus || m.focus == categoryFocus {
			m.focus = tracksFocus
		} else {
			m.focus = sidebarFocus
			m.sidebar = m.page
			m.revealMenuSelection()
		}
	case "up":
		m.move(-m.verticalStep())
	case "down":
		m.move(m.verticalStep())
	case "k", "left", "h":
		m.move(-1)
	case "j", "right", "l":
		m.move(1)
	case "enter":
		if m.focus == sidebarFocus {
			cmd := m.choosePage(m.sidebar)
			return m, cmd
		} else if m.focus == categoryFocus {
			m.focus = tracksFocus
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
	case "shift+left":
		m.player.Seek(m.player.Position() - 5*time.Second)
	case "shift+right":
		m.player.Seek(m.player.Position() + 5*time.Second)
	case "+", "=":
		m.player.AdjustVolume(5)
	case "-":
		m.player.AdjustVolume(-5)
	}
	if restart || m.player.Playing() != playing {
		m.rememberTrack()
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

func (m model) verticalStep() int {
	l := m.layout()
	if m.focus == tracksFocus && m.discoveryVisible() && homeDimensions(l.content(), l.body).discoveryRows > 1 {
		return discoveryColumns(l.content())
	}
	return 1
}

func (m *model) move(step int) {
	if m.focus == sidebarFocus {
		m.moveMenu(step)
		return
	}
	if m.focus == categoryFocus {
		m.setCategory(min(len(categories())-1, max(0, m.category+step)))
		return
	}
	m.selected = min(max(0, len(m.matches())-1), max(0, m.selected+step))
	if m.selected < m.offset {
		m.offset = m.selected
	} else if rows := m.trackRows(m.bodyHeight()); rows > 0 && m.selected >= m.offset+rows {
		m.offset = m.selected - rows + 1
	}
}

func (m *model) setCategory(i int) {
	m.category = i
	m.selected, m.offset = 0, 0
}
