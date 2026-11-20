package ui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/SAB1825/music-tui/internal/providers/mock"
)

func sized(m model, w, h int) model {
	next, _ := m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	return next.(model)
}

func TestPresentationBreakpoints(t *testing.T) {
	for _, tc := range []struct {
		w, h, sidebar, body, cover, columns int
	}{
		{160, 45, 33, 37, 20, 4},
		{120, 35, 25, 27, 16, 3},
		{80, 24, 0, 16, 12, 3},
	} {
		m := sized(New(mock.Catalog()).(model), tc.w, tc.h)
		l := m.layout()
		if l.sidebar != tc.sidebar || l.body != tc.body || l.body+l.toolbar+l.player != tc.h ||
			homeDimensions(l.content(), l.body).coverWidth != tc.cover || discoveryColumns(l.content()) != tc.columns {
			t.Fatalf("unexpected layout for %dx%d: %+v", tc.w, tc.h, l)
		}
		if view := m.render(); !strings.Contains(view, "Still up, Night Owl?") || !strings.Contains(view, "Energize") {
			t.Fatal("expected a feature and filters at the supported size")
		}
	}
}

func TestDiscoveryAndHistory(t *testing.T) {
	m := sized(New(mock.Catalog()).(model), 80, 24)
	m, _ = press(m, 'c', "c")
	m, _ = press(m, tea.KeyRight, "")
	if m.focus != categoryFocus || m.category != 1 || len(m.matches()) != 2 {
		t.Fatal("category should filter discovery")
	}
	m, _ = press(m, tea.KeyEnter, "")
	m, _ = press(m, tea.KeyRight, "")
	m, _ = press(m, tea.KeyEnter, "")
	if m.player.Index() != 6 || len(m.recent) != 1 || m.recent[0] != 6 {
		t.Fatal("expected the second discovery card to play")
	}
	m, _ = press(m, 'f', "f")
	if m.favorite() != "♥" {
		t.Fatal("favorite shortcut did not update the playing track")
	}
	m, _ = press(m, 'u', "u")
	m, _ = press(m, '[', "[")
	if m.page != home {
		t.Fatal("back should return to Home")
	}
	m, _ = press(m, ']', "]")
	if m.page != queue {
		t.Fatal("forward should reopen Queue")
	}
	m, _ = press(m, ',', ",")
	if m.dialog.kind != dialogSettings || m.page != queue || !strings.Contains(m.render(), "ARTWORK") {
		t.Fatal("settings shortcut did not open diagnostics")
	}
}

func TestCoverCacheAndTruncation(t *testing.T) {
	m := sized(New(mock.Catalog()).(model), 160, 45)
	size := coverSize{20, 9}
	cover := m.cover(0, size.width, size.height)
	w, h := lipgloss.Size(cover)
	if w != size.width || h != size.height || !strings.Contains(cover, "▀") || m.cover(1, w, h) == cover {
		t.Fatal("expected cached original artwork in separate palettes")
	}
	m = sized(m, 80, 24)
	if _, exists := m.covers[size]; exists {
		t.Fatal("resize should replace the prior geometry cache")
	}
	if formatted := label(m.theme.base, 6, "界界界界abc"); lipgloss.Width(formatted) != 6 || !strings.Contains(formatted, "…") {
		t.Fatalf("unexpected Unicode truncation: %q", formatted)
	}
}

func TestSpectrumSizing(t *testing.T) {
	frame := []float64{0, .5, 1, .25}
	bars := spectrumBars(frame, 2)
	if len(bars) != 2 || bars[0] != .5 || bars[1] != 1 || len(spectrumBars(frame, 30)) != len(frame) {
		t.Fatalf("unexpected regrouped frequency bars: %v", bars)
	}
	for _, tc := range []struct {
		value                 float64
		size, level, expected int
	}{
		{0, 3, 0, 0}, {.5, 3, 0, 8}, {.5, 3, 1, 4}, {.5, 3, 2, 0}, {1, 3, 2, 8},
	} {
		if got := barHeight(tc.value, tc.size, tc.level); got != tc.expected {
			t.Fatalf("bar height for %+v: %d", tc, got)
		}
	}
}
