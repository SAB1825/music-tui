package ui

import (
	"image"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

type dialogKind uint8

const (
	dialogNone dialogKind = iota
	dialogHelp
	dialogSettings
	dialogError
	dialogPlaylist
	dialogDelete
)

type dialog struct {
	kind             dialogKind
	selected, scroll int
	input            textinput.Model
}

type dialogContent struct {
	title   string
	lines   []string
	buttons []string
	input   bool
}

func (m *model) showDialog(kind dialogKind) tea.Cmd {
	m.dialog = dialog{kind: kind}
	m.hover = action{}
	if kind == dialogPlaylist {
		m.dialog.input = textinput.New()
		m.dialog.input.SetStyles(m.theme.inputStyles())
		m.dialog.input.Prompt = ""
		m.dialog.input.Placeholder = "Playlist name"
		m.dialog.input.CharLimit = 60
		m.dialog.input.SetWidth(max(1, min(64, m.width)-8))
		return m.dialog.input.Focus()
	}
	return nil
}

func (m *model) dismissDialog() tea.Cmd {
	m.dialog = dialog{}
	m.hover = action{}
	if m.focus == searchFocus {
		return m.search.Focus()
	}
	return nil
}

func (m model) dialogContent() dialogContent {
	switch m.dialog.kind {
	case dialogSettings:
		buttons := []string{"Close"}
		if m.spectrumError != "" {
			buttons = []string{"Close", "Details"}
		}
		return dialogContent{"Settings", m.settingsDetails(), buttons, false}
	case dialogError:
		return dialogContent{"Spectrum unavailable", diagnosticLines(m.spectrumError, max(1, min(m.width, 64)-6)), []string{"Back"}, false}
	case dialogPlaylist:
		return dialogContent{"Create playlist", []string{"Session playlist with the cued track, when available."}, []string{"Cancel", "Create"}, true}
	case dialogDelete:
		return dialogContent{"Delete playlist?", []string{"Remove “" + m.pageName(m.page) + "” from this session."}, []string{"Cancel", "Delete"}, false}
	default:
		return dialogContent{"Keyboard shortcuts", []string{
			"↑↓ / j k     Navigate",
			"←→ / h l     Browse cards and filters",
			"Enter        Open or play",
			"Space        Play / pause",
			"/            Search",
			"Tab          Switch menu / tracks",
			"c            Focus filters",
			"[ / ]        Back / forward",
			"n / p        Next / previous",
			"Shift ←/→    Seek 5 seconds",
			"+ / -        Volume",
			"f / s / r    Favorite / shuffle / repeat",
			"u / ,        Queue / settings",
			"Delete       Remove an open session playlist",
			"Esc          Close dialog",
			"q / Ctrl+C   Quit",
		}, []string{"Close"}, false}
	}
}

func (m model) dialogGeometry(d dialogContent) (width, border, padding, inner, visible, height int) {
	width = min(m.width, 64)
	if width >= 12 && m.height >= 4 {
		border, padding = 1, 2
	}
	inner = max(1, width-2*(border+padding))
	fixed := 2
	if d.input {
		fixed++
	}
	visible = max(0, min(len(d.lines), m.height-2*border-fixed))
	height = min(m.height, visible+fixed+2*border)
	return
}

func (m model) overlay(base string) string {
	d := m.dialogContent()
	w, border, pad, inner, visible, height := m.dialogGeometry(d)
	x, y := (m.width-w)/2, (m.height-height)/2
	m.drawing.dialog = image.Rect(x, y, x+w, y+height)
	dx, dy := x+border+pad, y+border
	t := m.theme
	closeWidth := min(3, max(0, inner-1))
	rows := []string{label(t.elevated.Bold(true), inner-closeWidth, d.title) + m.control(t.elevated, dialogCloseAction, dx+inner-closeWidth, dy, closeWidth, " × ")}
	first := clampScroll(m.dialog.scroll, len(d.lines), visible)
	for _, value := range d.lines[first : first+visible] {
		rows = append(rows, label(t.muted.Background(t.elevated.GetBackground()), inner, value))
	}
	if d.input && len(rows) < height-2*border-1 {
		m.region(dialogInputAction, 0, dx, dy+len(rows), inner, 1)
		style := t.panel
		if m.dialog.selected == 0 {
			style = t.selected
		}
		rows = append(rows, label(style, inner, m.dialog.input.View()))
	}
	if len(rows) < height-2*border {
		var buttons strings.Builder
		for i, title := range d.buttons {
			start, end := i*inner/len(d.buttons), (i+1)*inner/len(d.buttons)
			style := t.elevated
			if end-start < lipgloss.Width(title)+2 {
				title = compactDialogButton(title)
			}
			prefix := "  "
			if i == m.dialogButtonIndex() {
				style, prefix = t.button, "› "
			}
			if m.buttonEnabled(i) {
				m.region(dialogButtonAction, i, dx+start, dy+len(rows), end-start, 1)
				style = m.hoverStyle(style, dialogButtonAction, i)
			} else {
				style = t.disabled.Background(t.elevated.GetBackground())
			}
			if end-start < lipgloss.Width(title)+2 {
				prefix = ""
			}
			buttons.WriteString(label(style, end-start, prefix+title))
		}
		rows = append(rows, buttons.String())
	}
	rows = padRows(rows, t.elevated, inner, height-2*border)
	style := t.elevated
	if border > 0 {
		style = style.Padding(0, pad).Border(lipgloss.RoundedBorder()).
			BorderForeground(t.focusBorder).BorderBackground(t.elevated.GetBackground())
	}
	panel := style.Render(strings.Join(rows, "\n"))
	canvas := lipgloss.NewCanvas(m.width, m.height)
	canvas.Compose(lipgloss.NewCompositor(lipgloss.NewLayer(base), lipgloss.NewLayer(panel).X(x).Y(y).Z(1)))
	out := strings.Split(canvas.Render(), "\n")
	return strings.Join(padRows(out, t.base, m.width, m.height), "\n")
}

func (m model) dialogButtonIndex() int {
	if m.dialog.kind == dialogPlaylist {
		return m.dialog.selected - 1
	}
	return m.dialog.selected
}

func (m model) buttonEnabled(i int) bool {
	return m.dialog.kind != dialogPlaylist || i != 1 || strings.TrimSpace(m.dialog.input.Value()) != ""
}

func (m model) activateDialog(i int) (model, tea.Cmd) {
	if !m.buttonEnabled(i) {
		return m, nil
	}
	switch {
	case m.dialog.kind == dialogSettings && i == 1:
		cmd := m.showDialog(dialogError)
		return m, cmd
	case m.dialog.kind == dialogError:
		cmd := m.showDialog(dialogSettings)
		return m, cmd
	case m.dialog.kind == dialogPlaylist && i == 1:
		m.addPlaylist(strings.TrimSpace(m.dialog.input.Value()))
	case m.dialog.kind == dialogDelete && i == 1:
		m.removePlaylist()
	}
	cmd := m.dismissDialog()
	return m, cmd
}

func (m model) clickDialog(a action, inside bool) (model, tea.Cmd) {
	switch a.kind {
	case dialogCloseAction:
		cmd := m.dismissDialog()
		return m, cmd
	case dialogButtonAction:
		return m.activateDialog(a.index)
	case dialogInputAction:
		m.dialog.selected = 0
		cmd := m.dialog.input.Focus()
		return m, cmd
	default:
		if !inside && m.dialog.kind != dialogDelete && m.dialog.kind != dialogPlaylist {
			cmd := m.dismissDialog()
			return m, cmd
		}
	}
	return m, nil
}

func (m model) dialogKey(key tea.KeyPressMsg) (model, tea.Cmd) {
	switch key.String() {
	case "esc":
		cmd := m.dismissDialog()
		return m, cmd
	case "?":
		if m.dialog.kind == dialogHelp {
			cmd := m.dismissDialog()
			return m, cmd
		} else {
			return m.updateDialogInput(key)
		}
	case "q":
		if m.dialog.kind == dialogHelp {
			return m, tea.Quit
		}
		return m.updateDialogInput(key)
	case "enter":
		i := m.dialogButtonIndex()
		if i < 0 {
			i = 1
		}
		return m.activateDialog(i)
	case "tab", "shift+tab":
		d := m.dialogContent()
		count := len(d.buttons)
		if d.input {
			count++
		}
		step := 1
		if key.String() == "shift+tab" {
			step = -1
		}
		m.dialog.selected = (m.dialog.selected + step + count) % count
		if d.input {
			if m.dialog.selected == 0 {
				cmd := m.dialog.input.Focus()
				return m, cmd
			}
			m.dialog.input.Blur()
		}
	case "up", "down":
		step := 1
		if key.String() == "up" {
			step = -1
		}
		m.scrollDialog(step)
	default:
		return m.updateDialogInput(key)
	}
	return m, nil
}

func (m *model) scrollDialog(step int) {
	d := m.dialogContent()
	_, _, _, _, visible, _ := m.dialogGeometry(d)
	m.dialog.scroll = clampScroll(m.dialog.scroll+step, len(d.lines), visible)
}

func (m model) updateDialogInput(msg tea.Msg) (model, tea.Cmd) {
	if m.dialog.kind == dialogPlaylist && m.dialog.input.Focused() {
		var cmd tea.Cmd
		m.dialog.input, cmd = m.dialog.input.Update(msg)
		return m, cmd
	}
	return m, nil
}

func diagnosticLines(value string, width int) []string {
	text := strconv.QuoteToGraphic(value)
	return strings.Split(ansi.Wrap("Error: "+text[1:len(text)-1], width, ""), "\n")
}

func compactDialogButton(title string) string {
	switch title {
	case "Cancel":
		return "No"
	case "Delete":
		return "Yes"
	case "Create":
		return "Add"
	case "Details":
		return "Info"
	case "Close":
		return "×"
	default:
		return title
	}
}
