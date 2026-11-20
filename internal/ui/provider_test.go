package ui

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/SAB1825/music-tui/internal/domain"
	"github.com/SAB1825/music-tui/internal/providers"
	"github.com/SAB1825/music-tui/internal/providers/mock"
)

type testCatalog struct {
	calls          []string
	requestContext context.Context
}

func (c *testCatalog) Search(ctx context.Context, q, token string) (domain.TrackPage, error) {
	c.calls = append(c.calls, q+"|"+token)
	c.requestContext = ctx
	if q == "quota" {
		return domain.TrackPage{}, &providers.Error{Kind: providers.Quota}
	}
	if token != "" {
		return domain.TrackPage{Tracks: []domain.Track{{ID: "one"}, {ID: "two", Title: "More music", Provider: "youtube"}}}, nil
	}
	return domain.TrackPage{NextToken: "next", Tracks: []domain.Track{{
		ID: "one", Title: "Real video", Creator: "Example channel", Duration: 3*time.Minute + time.Second, Provider: "youtube",
		ExternalURL: "https://www.youtube.com/watch?v=one",
	}}}, nil
}

func (c *testCatalog) Playlist(_ context.Context, id string) (domain.Playlist, error) {
	return domain.Playlist{ID: id, Title: "Public test playlist", Description: "Test description", ExternalURL: "https://www.youtube.com/playlist?list=fixture"}, nil
}

func (c *testCatalog) PlaylistTracks(_ context.Context, id, token string) (domain.TrackPage, error) {
	c.calls = append(c.calls, id+"|"+token)
	return domain.TrackPage{Tracks: []domain.Track{{ID: "one", Title: "Video on playlist", Provider: "youtube"}}}, nil
}

func testOnline(t *testing.T, c *testCatalog, w, h int) model {
	t.Helper()
	var r providers.Registry
	r.Register(providers.Entry{ID: "youtube", Name: "YouTube", Catalog: c})
	r.Register(providers.Entry{ID: "mock", Name: "Mock", Catalog: mock.Provider{}})
	return sized(WithProviders(New(mock.Catalog()), t.Context(), r, "auto").(model), w, h)
}

func submitSearch(m model, query string) (model, tea.Cmd) {
	m, _ = press(m, '/', "/")
	m.search.SetValue(query)
	return press(m, tea.KeyEnter, "")
}

func TestOnlineSearchNavigation(t *testing.T) {
	for _, size := range [][2]int{{160, 45}, {120, 35}, {80, 24}} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			c := &testCatalog{}
			m := testOnline(t, c, size[0], size[1])
			if strings.Contains(m.View().Content, "Copper Horizon") {
				t.Fatal("online start rendered mock data")
			}
			m, cmd := submitSearch(m, "Nina Simone")
			if cmd == nil || len(c.calls) != 0 || !strings.Contains(m.View().Content, "Searching YouTube") {
				t.Fatal("expected an asynchronous loading state")
			}
			result := cmd()
			next, _ := m.Update(result)
			m = next.(model)
			rows := strings.Split(m.View().Content, "\n")
			if len(rows) != size[1] {
				t.Fatalf("online view has %d rows, want %d", len(rows), size[1])
			}
			for _, line := range rows {
				if lipgloss.Width(line) > size[0] {
					t.Fatal("online view overflowed")
				}
			}
			for _, name := range []string{"Real video", "Example channel", "3:01", "YouTube"} {
				if !strings.Contains(m.View().Content, name) {
					t.Errorf("missing %q", name)
				}
			}
			m, again := submitSearch(m, "Nina Simone")
			if again != nil || len(c.calls) != 1 {
				t.Fatal("duplicate submission incurred a request")
			}
			target := findTarget(t, m, moreAction, 0)
			next, cmd = m.Update(tea.MouseClickMsg{X: target.bounds.Min.X, Y: target.bounds.Min.Y, Button: tea.MouseLeft})
			if cmd == nil {
				t.Fatal("More click didn't request the next page")
			}
			m = next.(model)
			next, _ = m.Update(cmd())
			m = next.(model)
			if len(m.online.tracks) != 2 || m.online.next != "" {
				t.Fatal("pagination failed to append and deduplicate")
			}
			m = clickTarget(t, m, trackAction, 0)
			m, cmd = press(m, tea.KeyEnter, "")
			if cmd == nil || m.player.Playing() {
				t.Fatal("online track must offer browser opening without playing the mock queue")
			}
			m, _ = press(m, 'v', "v")
			if m.dialog.kind != dialogProviders {
				t.Fatal("provider dialog did not open")
			}
			m = clickTarget(t, m, dialogButtonAction, 1)
			if m.remote() || len(m.online.tracks) != 0 || !strings.Contains(m.View().Content, "Copper Horizon") {
				t.Fatal("returning to mock failed")
			}
		})
	}
}

func TestSupersededResultsAndPlaylist(t *testing.T) {
	c := &testCatalog{}
	m := testOnline(t, c, 120, 35)
	m, old := submitSearch(m, "old")
	m, current := submitSearch(m, "new")
	older := old().(catalogResult)
	if c.requestContext.Err() == nil {
		t.Fatal("old command should be canceled")
	}
	next, _ := m.Update(current())
	m = next.(model)
	next, _ = m.Update(older)
	m = next.(model)
	if m.online.query != "new" || len(m.online.tracks) != 1 {
		t.Fatal("obsolete result replaced current search")
	}
	m, cmd := submitSearch(m, "quota")
	next, _ = m.Update(cmd())
	m = next.(model)
	if m.dialog.kind != dialogCatalogError || !strings.Contains(m.View().Content, "quota exceeded") {
		t.Fatal("sanitized quota message missing")
	}
	m, _ = press(m, tea.KeyEsc, "")
	m, _ = press(m, 'g', "g")
	next, _ = m.Update(tea.PasteMsg{Content: "fixture"})
	m = next.(model)
	m, cmd = press(m, tea.KeyEnter, "")
	if cmd == nil {
		t.Fatal("playlist submission did not start")
	}
	next, _ = m.Update(cmd())
	m = next.(model)
	if m.online.playlist.ID != "fixture" || !strings.Contains(m.View().Content, "Video on playlist") {
		t.Fatal("playlist not displayed")
	}
}
