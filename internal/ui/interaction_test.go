package ui

import (
	"image"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/SAB1825/music-tui/internal/domain"
	"github.com/SAB1825/music-tui/internal/providers/mock"
	"github.com/charmbracelet/x/ansi"
)

func findTarget(t *testing.T, m model, kind actionKind, index int) target {
	t.Helper()
	for _, hit := range m.screen.targets {
		if hit.action == (action{kind, index}) {
			return hit
		}
	}
	t.Fatalf("missing target %v/%d", kind, index)
	return target{}
}

func mouseClick(m model, x, y int) model {
	next, _ := m.Update(tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft})
	return next.(model)
}

func clickTarget(t *testing.T, m model, kind actionKind, index int) model {
	t.Helper()
	h := findTarget(t, m, kind, index)
	return mouseClick(m, h.bounds.Min.X, h.bounds.Min.Y)
}

func mouseWheel(m model, x, y int, button tea.MouseButton) model {
	next, _ := m.Update(tea.MouseWheelMsg{X: x, Y: y, Button: button})
	return next.(model)
}

func TestHitTestingAcrossSizes(t *testing.T) {
	for _, size := range []image.Point{{X: 160, Y: 45}, {X: 120, Y: 35}, {X: 80, Y: 24}} {
		m := sized(New(mock.Catalog()).(model), size.X, size.Y)
		for _, page := range []page{home, library} {
			m.open(page)
			m.screen = m.draw()
			for i, h := range m.screen.targets {
				if h.bounds.Empty() || !h.bounds.In(image.Rect(0, 0, size.X, size.Y)) {
					t.Fatalf("invalid target: %+v", h)
				}
				for _, other := range m.screen.targets[i+1:] {
					if h.bounds.Overlaps(other.bounds) {
						t.Fatalf("overlapping targets: %+v and %+v", h, other)
					}
				}
			}
			rendered := strings.Split(m.View().Content, "\n")
			h := findTarget(t, m, playAction, 0)
			if !strings.Contains(ansi.Strip(ansi.Cut(rendered[h.bounds.Min.Y], h.bounds.Min.X, h.bounds.Max.X)), "▶") {
				t.Fatal("play hitbox missed the rendered control")
			}
		}
	}
	f := screenFrame{targets: []target{{image.Rect(2, 3, 5, 7), action{trackAction, 1}}}}
	if f.hit(2, 3).action.kind != trackAction || f.hit(5, 6).action.kind != noAction || f.hit(4, 7).action.kind != noAction {
		t.Fatal("expected inclusive top-left and exclusive bottom-right")
	}
}

func TestMouseNavigationSearchAndResize(t *testing.T) {
	m := sized(New(mock.Catalog()).(model), 160, 45)
	for _, h := range m.screen.targets {
		if h.action == (action{navigateAction, int(radio)}) {
			t.Fatal("disabled item is clickable")
		}
	}
	m = clickTarget(t, m, navigateAction, int(library))
	m = clickTarget(t, m, trackAction, 2)
	if m.page != library || m.selected != 2 || m.player.Playing() {
		t.Fatal("click did not select a library track")
	}
	m, _ = press(m, tea.KeyEnter, "")
	if m.player.Index() != 2 || !m.player.Playing() {
		t.Fatal("Enter did not play the clicked track")
	}
	m = clickTarget(t, m, searchAction, 0)
	m = clickTarget(t, m, searchAction, 0)
	m, _ = press(m, 'n', "n")
	if m.search.Value() != "n" || m.player.Index() != 2 {
		t.Fatal("search click or text triggered a global action")
	}
	m = mouseClick(m, 0, 0)
	if m.search.Focused() || m.search.Value() != "n" {
		t.Fatal("click-away should retain the query")
	}
	m, _ = press(m, tea.KeyEsc, "")
	m = sized(m, 80, 24)
	m = clickTarget(t, m, menuAction, 0)
	m = clickTarget(t, m, navigateAction, int(home))
	m = clickTarget(t, m, filterAction, 1)
	m = clickTarget(t, m, trackAction, 1)
	if m.page != home || m.category != 1 || m.selected != 1 {
		t.Fatal("compact menu/cards did not use resized geometry")
	}
	m = clickTarget(t, m, backAction, 0)
	if m.page != library {
		t.Fatal("toolbar back did not navigate")
	}
	m = clickTarget(t, m, forwardAction, 0)
	if m.page != home {
		t.Fatal("toolbar forward did not navigate")
	}
}

func TestMousePlaybackControls(t *testing.T) {
	m := sized(New(mock.Catalog()).(model), 120, 35)
	m = clickTarget(t, m, playAction, 0)
	m = clickTarget(t, m, nextAction, 0)
	m = clickTarget(t, m, prevAction, 0)
	m = clickTarget(t, m, shuffleAction, 0)
	m = clickTarget(t, m, repeatAction, 0)
	if !m.player.Playing() || m.player.Index() != 0 || !m.player.Shuffle() || !m.player.Repeat() {
		t.Fatal("playback buttons failed")
	}
	h := findTarget(t, m, seekAction, 0)
	m = mouseClick(m, h.bounds.Max.X-1, h.bounds.Min.Y)
	track, _ := m.player.Track()
	if m.player.Position() != track.Duration {
		t.Fatal("end of progress did not seek to end")
	}
	h = findTarget(t, m, seekAction, 0)
	m = mouseClick(m, h.bounds.Min.X, h.bounds.Min.Y)
	if m.player.Position() != 0 {
		t.Fatal("start of progress did not rewind")
	}
	h = findTarget(t, m, volumeAction, 0)
	m = mouseClick(m, h.bounds.Max.X-1, h.bounds.Min.Y)
	if m.player.Volume() != 100 {
		t.Fatal("volume did not reach 100")
	}
	m = clickTarget(t, m, volumeAction, 0)
	if m.player.Volume() != 0 {
		t.Fatal("volume did not reach zero")
	}
	m = clickTarget(t, m, queueAction, 0)
	if m.page != queue {
		t.Fatal("queue button did not open queue")
	}
}

func TestPointerScrollingAndHover(t *testing.T) {
	var catalog []domain.Track
	for range 6 {
		catalog = append(catalog, mock.Catalog()...)
	}
	m := sized(New(catalog).(model), 120, 35)
	m = clickTarget(t, m, navigateAction, int(library))
	l := m.layout()
	m = mouseWheel(m, l.sidebar+1, l.toolbar+2, tea.MouseWheelDown)
	if m.offset != 1 || m.menuOffset != 0 {
		t.Fatal("library wheel should advance only the library")
	}
	for range 55 {
		m = mouseWheel(m, l.sidebar+1, l.toolbar+2, tea.MouseWheelDown)
	}
	if m.offset != len(catalog)-m.trackRows(l.body) {
		t.Fatal("library wheel exceeded bounds")
	}
	m = mouseWheel(m, l.sidebar+1, 0, tea.MouseWheelUp)
	if m.offset != len(catalog)-m.trackRows(l.body) {
		t.Fatal("toolbar wheel leaked into library")
	}
	m = sized(m, 120, 18)
	m = mouseWheel(m, 1, m.layout().toolbar, tea.MouseWheelDown)
	if m.menuOffset != 1 {
		t.Fatal("sidebar wheel ignored its viewport")
	}
	h := findTarget(t, m, navigateAction, int(library))
	msg := tea.MouseMotionMsg{X: h.bounds.Min.X, Y: h.bounds.Min.Y}
	next, _ := m.Update(msg)
	hovered := next.(model)
	if hovered.hover != h.action || hovered.View().Content == m.View().Content {
		t.Fatal("hover styling did not change")
	}
	next, _ = hovered.Update(msg)
	if next.(model).screen.content != hovered.screen.content {
		t.Fatal("stationary hover changed the frame")
	}
}

func TestDialogsCaptureInputAndRestoreFocus(t *testing.T) {
	m := sized(New(mock.Catalog()).(model), 120, 35)
	m = clickTarget(t, m, searchAction, 0)
	m = clickTarget(t, m, navigateAction, int(settings))
	if m.dialog.kind != dialogSettings || m.screen.dialog.Min.X != (m.width-m.screen.dialog.Dx())/2 ||
		!strings.Contains(m.View().Content, "╭") || !strings.Contains(m.View().Content, "╯") {
		t.Fatal("settings dialog is not centered and rounded")
	}
	m, _ = press(m, ' ', " ")
	if m.player.Playing() {
		t.Fatal("modal key passed through")
	}
	m = mouseClick(m, 1, 4)
	if m.dialog.kind != dialogNone || m.focus != searchFocus || !m.search.Focused() || m.page != home {
		t.Fatal("outside click failed to dismiss or passed through")
	}
	m, _ = press(m, tea.KeyEsc, "")
	m, _ = press(m, '?', "?")
	m = clickTarget(t, m, dialogCloseAction, 0)
	if m.dialog.kind != dialogNone {
		t.Fatal("close control failed")
	}
	m = clickTarget(t, m, navigateAction, int(createPlaylist))
	next, _ := m.Update(tea.PasteMsg{Content: "Test list"})
	m = next.(model)
	m = clickTarget(t, m, dialogButtonAction, 1)
	if m.page < firstPlaylist || m.pageName(m.page) != "Test list" || len(m.matches()) != 1 {
		t.Fatal("playlist confirmation failed")
	}
	m, _ = press(m, tea.KeyDelete, "")
	m = mouseClick(m, 0, 0)
	if m.dialog.kind != dialogDelete {
		t.Fatal("destructive dialog dismissed outside")
	}
	m, _ = press(m, tea.KeyEnter, "")
	if m.playlists[0].deleted || m.dialog.kind != dialogNone {
		t.Fatal("default action must cancel deletion")
	}
	m, _ = press(m, tea.KeyDelete, "")
	m, _ = press(m, tea.KeyTab, "")
	m, _ = press(m, tea.KeyEnter, "")
	if !m.playlists[0].deleted || m.page != home {
		t.Fatal("explicit confirmation did not delete playlist")
	}
}

func TestSmallDialogsKeepActionsVisible(t *testing.T) {
	for _, size := range []image.Point{{X: 80, Y: 24}, {X: 24, Y: 8}, {X: 14, Y: 5}} {
		m := sized(New(mock.Catalog()).(model), size.X, size.Y)
		m.showDialog(dialogDelete)
		m.addPlaylist("Session")
		m.screen = m.draw()
		rows := strings.Split(m.View().Content, "\n")
		for i, title := range []string{"Cancel", "Delete"} {
			h := findTarget(t, m, dialogButtonAction, i)
			text := ansi.Strip(ansi.Cut(rows[h.bounds.Min.Y], h.bounds.Min.X, h.bounds.Max.X))
			if size.X < 24 {
				title = compactDialogButton(title)
			}
			if !strings.Contains(text, title) {
				t.Fatalf("button clipped at %v: %q", size, text)
			}
		}
		for _, row := range rows {
			if lipgloss.Width(row) > size.X {
				t.Fatalf("dialog overflow at %v", size)
			}
		}
		m, _ = press(m, tea.KeyEsc, "")
		if m.dialog.kind != dialogNone {
			t.Fatal("Escape failed")
		}
	}
}

func TestChromeIsGrayscale(t *testing.T) {
	theme := newTheme()
	for _, style := range []lipgloss.Style{theme.base, theme.sidebar, theme.panel, theme.card, theme.inset,
		theme.elevated, theme.accent, theme.muted, theme.dim, theme.button, theme.selected, theme.hover, theme.disabled} {
		for _, c := range []interface {
			RGBA() (uint32, uint32, uint32, uint32)
		}{style.GetBackground(), style.GetForeground()} {
			r, g, b, _ := c.RGBA()
			if r != g || g != b {
				t.Fatal("chrome uses a non-gray color")
			}
		}
	}
	if cellFraction(6, image.Rect(2, 0, 11, 1)) != .5 || time.Duration(cellFraction(2, image.Rect(2, 0, 3, 1))*float64(time.Second)) != 500*time.Millisecond {
		t.Fatal("seek/volume fraction calculation is incorrect")
	}
}
