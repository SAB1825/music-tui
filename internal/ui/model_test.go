package ui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/SAB1825/music-tui/internal/providers/mock"
)

func press(m model, code rune, text string) (model, tea.Cmd) {
	next, cmd := m.Update(tea.KeyPressMsg{Code: code, Text: text})
	return next.(model), cmd
}

func TestNavigationAndPlayback(t *testing.T) {
	m := New(mock.Catalog()).(model)
	m, _ = press(m, tea.KeyTab, "")
	m, _ = press(m, tea.KeyDown, "")
	m, _ = press(m, tea.KeyEnter, "")
	if m.page != library || m.focus != tracksFocus {
		t.Fatal("enter should open the focused sidebar item")
	}
	m, _ = press(m, 'j', "j")
	m, cmd := press(m, tea.KeyEnter, "")
	track, _ := m.player.Track()
	if track.Title != "Still Air" || !m.player.Playing() || cmd == nil {
		t.Fatalf("enter should start the selected track, got %+v", track)
	}
	oldTick := playbackTick(m.clock)
	m, _ = press(m, 'n', "n")
	next, _ := m.Update(oldTick)
	m = next.(model)
	if m.player.Position() != 0 {
		t.Fatal("stale tick advanced a newly selected track")
	}
	next, cmd = m.Update(playbackTick(m.clock))
	m = next.(model)
	if m.player.Position() != time.Second || cmd == nil {
		t.Fatal("active tick should advance playback and schedule the next tick")
	}
	m, cmd = press(m, ' ', " ")
	if m.player.Playing() || cmd != nil {
		t.Fatal("space should pause playback")
	}
}

func TestSearchFocusAndPaste(t *testing.T) {
	m := New(mock.Catalog()).(model)
	m, _ = press(m, '/', "/")
	m, cmd := press(m, 'q', "q")
	if m.search.Value() != "q" || !m.search.Focused() {
		t.Fatal("q should be entered as text in search")
	}
	if cmd != nil {
		if _, quit := cmd().(tea.QuitMsg); quit {
			t.Fatal("q quit while searching")
		}
	}
	m, _ = press(m, tea.KeyEsc, "")
	m, _ = press(m, tea.KeyDown, "")
	m, _ = press(m, '/', "/")
	next, _ := m.Update(tea.PasteMsg{Content: "Copper Horizon"})
	m = next.(model)
	m, _ = press(m, tea.KeyEnter, "")
	if m.search.Focused() || m.selected != 0 || len(m.matches()) != 1 {
		t.Fatal("pasted filter should reset selection and enter should return to results")
	}
	m, _ = press(m, tea.KeyEnter, "")
	track, _ := m.player.Track()
	if track.Title != "Copper Horizon" || !m.player.Playing() {
		t.Fatal("enter should play the filtered result")
	}
}

func TestLayoutFitsTerminal(t *testing.T) {
	for _, size := range [][2]int{{1, 1}, {8, 4}, {20, 5}, {32, 10}, {65, 13}, {66, 13}, {80, 24}, {120, 35}, {160, 45}} {
		for _, mode := range []string{"tracks", "sidebar", "search", "help"} {
			w, h := size[0], size[1]
			t.Run(fmt.Sprintf("%dx%d/%s", w, h, mode), func(t *testing.T) {
				m := New(mock.Catalog()).(model)
				next, _ := m.Update(tea.WindowSizeMsg{Width: w, Height: h})
				m = next.(model)
				switch mode {
				case "sidebar":
					m, _ = press(m, tea.KeyTab, "")
				case "search":
					m, _ = press(m, '/', "/")
					next, _ := m.Update(tea.PasteMsg{Content: "月 音 é 界界界界界界界界界界界界"})
					m = next.(model)
				case "help":
					m, _ = press(m, '?', "?")
				}
				lines := strings.Split(m.View().Content, "\n")
				if len(lines) != h {
					t.Fatalf("rendered %d rows, want %d", len(lines), h)
				}
				for i, line := range lines {
					if width := lipgloss.Width(line); width > w {
						t.Fatalf("row %d rendered %d cells, want at most %d", i, width, w)
					}
				}
				dock := strings.Join(lines[h-m.playerHeight():], "\n")
				if w >= 20 && !strings.Contains(dock, "Paused") {
					t.Fatal("player is missing from the bottom of the view")
				}
			})
		}
	}
}

func TestSelectedTrackVisibleAfterResize(t *testing.T) {
	m := New(mock.Catalog()).(model)
	next, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 34})
	m = next.(model)
	for range len(m.catalog) - 1 {
		m, _ = press(m, 'j', "j")
	}
	next, _ = m.Update(tea.WindowSizeMsg{Width: 40, Height: 10})
	m = next.(model)
	if !strings.Contains(m.View().Content, "Last Light") {
		t.Fatal("resize hid the selected track")
	}
}
