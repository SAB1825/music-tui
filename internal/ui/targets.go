package ui

import (
	"image"
	"math"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type actionKind uint8

const (
	noAction actionKind = iota
	navigateAction
	trackAction
	filterAction
	searchAction
	backAction
	forwardAction
	menuAction
	playAction
	prevAction
	nextAction
	shuffleAction
	repeatAction
	queueAction
	seekAction
	volumeAction
	favoriteAction
	dialogCloseAction
	dialogButtonAction
	dialogInputAction
	deleteAction
)

type action struct {
	kind  actionKind
	index int
}

type target struct {
	bounds image.Rectangle
	action action
}

type screenFrame struct {
	content string
	targets []target
	dialog  image.Rectangle
}

func (f screenFrame) hit(x, y int) target {
	p := image.Pt(x, y)
	for _, item := range f.targets {
		if p.In(item.bounds) {
			return item
		}
	}
	return target{}
}

// Registration occurs alongside rendering so clipped and omitted controls have no targets.
func (m model) region(kind actionKind, index, x, y, w, h int) {
	if m.drawing == nil {
		return
	}
	r := image.Rect(x, y, x+max(0, w), y+max(0, h)).Intersect(image.Rect(0, 0, m.width, m.height))
	if !r.Empty() {
		m.drawing.targets = append(m.drawing.targets, target{r, action{kind, index}})
	}
}

func (m model) hoverStyle(style lipgloss.Style, kind actionKind, index int) lipgloss.Style {
	if m.hover == (action{kind, index}) {
		return style.Background(m.theme.hover.GetBackground()).Foreground(m.theme.foreground)
	}
	return style
}

func (m model) control(style lipgloss.Style, kind actionKind, x, y, w int, text string) string {
	switch kind {
	case playAction, prevAction, nextAction, favoriteAction:
		if m.player.Index() < 0 {
			return line(m.theme.disabled.Background(style.GetBackground()), w, text)
		}
	}
	m.region(kind, 0, x, y, w, 1)
	return line(m.hoverStyle(style, kind, 0), w, text)
}

func (m model) updateMouse(msg tea.MouseMsg) (model, tea.Cmd) {
	mouse := msg.Mouse()
	hit := m.screen.hit(mouse.X, mouse.Y)
	switch msg.(type) {
	case tea.MouseMotionMsg:
		m.hover = hit.action
		return m, nil
	case tea.MouseWheelMsg:
		step := 0
		switch mouse.Button {
		case tea.MouseWheelUp, tea.MouseWheelLeft:
			step = -1
		case tea.MouseWheelDown, tea.MouseWheelRight:
			step = 1
		}
		if step != 0 {
			m.hover = action{}
			m.scrollAt(mouse.X, mouse.Y, step)
		}
		return m, nil
	case tea.MouseClickMsg:
		if mouse.Button != tea.MouseLeft || !image.Pt(mouse.X, mouse.Y).In(image.Rect(0, 0, m.width, m.height)) {
			return m, nil
		}
	default:
		return m, nil
	}
	m.hover = hit.action
	if m.dialog.kind != dialogNone {
		return m.clickDialog(hit.action, image.Pt(mouse.X, mouse.Y).In(m.screen.dialog))
	}
	if m.focus == searchFocus && hit.action.kind != searchAction && !hit.action.opensDialog() {
		m.search.Blur()
		m.focus = tracksFocus
	}
	switch hit.action.kind {
	case navigateAction:
		cmd := m.choosePage(page(hit.action.index))
		return m, cmd
	case trackAction:
		m.selected, m.focus = hit.action.index, tracksFocus
	case filterAction:
		m.setCategory(hit.action.index)
		m.focus = categoryFocus
	case searchAction:
		if m.focus != searchFocus {
			m.focus = searchFocus
			cmd := m.search.Focus()
			return m, cmd
		}
	case menuAction:
		return m.updateKey(tea.KeyPressMsg{Code: tea.KeyTab})
	case backAction:
		m.travel(-1)
	case forwardAction:
		m.travel(1)
	case deleteAction:
		cmd := m.showDialog(dialogDelete)
		return m, cmd
	case seekAction:
		if track, ok := m.player.Track(); ok {
			m.player.Seek(time.Duration(float64(track.Duration) * cellFraction(mouse.X, hit.bounds)))
		}
	case volumeAction:
		m.player.AdjustVolume(int(math.Round(100*cellFraction(mouse.X, hit.bounds))) - m.player.Volume())
	default:
		if key := playbackActionKey(hit.action.kind); key != 0 {
			return m.updateKey(tea.KeyPressMsg{Code: key, Text: string(key)})
		}
	}
	return m, nil
}

func cellFraction(x int, bounds image.Rectangle) float64 {
	if bounds.Dx() < 2 {
		return .5
	}
	return min(1, max(0, float64(x-bounds.Min.X)/float64(bounds.Dx()-1)))
}

func playbackActionKey(kind actionKind) rune {
	switch kind {
	case playAction:
		return ' '
	case prevAction:
		return 'p'
	case nextAction:
		return 'n'
	case shuffleAction:
		return 's'
	case repeatAction:
		return 'r'
	case queueAction:
		return 'u'
	case favoriteAction:
		return 'f'
	default:
		return 0
	}
}

func (m *model) scrollAt(x, y, step int) {
	l := m.layout()
	if m.dialog.kind != dialogNone {
		if image.Pt(x, y).In(m.screen.dialog) {
			m.scrollDialog(step)
		}
		return
	}
	if x < 0 || x >= l.width || y < l.toolbar || y >= l.toolbar+l.body || l.toolbar == 0 {
		return
	}
	if x < l.sidebar || l.sidebar == 0 && m.focus == sidebarFocus {
		m.menuOffset = clampScroll(m.menuOffset+step, len(m.menuRows()), l.body-2)
		return
	}
	if m.focus == searchFocus {
		m.search.Blur()
	}
	m.focus = tracksFocus
	count, visible := len(m.matches()), m.trackRows(l.body)
	if count > visible && visible > 0 {
		m.offset = clampScroll(m.trackStart(count, visible)+step, count, visible)
		m.selected = min(m.offset+visible-1, max(m.offset, m.selected))
	}
}

func (m model) trackStart(count, visible int) int {
	return clampScroll(max(min(m.offset, m.selected), m.selected-visible+1), count, visible)
}

func clampScroll(offset, count, visible int) int {
	return min(max(0, count-max(1, visible)), max(0, offset))
}

func (a action) opensDialog() bool {
	return a.kind == navigateAction && (page(a.index) == settings || page(a.index) == createPlaylist)
}
