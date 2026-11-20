package ui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
)

func (m model) playerStatus() string {
	if m.player.Playing() {
		return "Playing"
	}
	if m.player.Index() >= 0 {
		return "Paused"
	}
	return "Idle"
}

func (m model) playSymbol() string {
	if m.player.Playing() {
		return "Ⅱ"
	}
	return "▶"
}

func (m model) playerBar() string {
	if m.remote() {
		return m.catalogBar()
	}
	t, w, h := m.theme, m.width, m.playerHeight()
	top := m.height - h
	track, ok := m.player.Track()
	title := "No track queued"
	if ok {
		title = track.Title
	}
	if h == 1 {
		return label(t.panel.Bold(true), w, m.playerStatus()+" · "+title)
	}
	if h == 2 {
		return label(t.panel.Bold(true), w, "  "+m.playerStatus()+" · "+title) + "\n" +
			m.compactControls(w, top+1)
	}
	leftW, rightW := w*29/100, w*23/100
	if w < 100 {
		leftW, rightW = w*34/100, w*21/100
	}
	if w < 52 {
		leftW, rightW = w, 0
	}
	centerW := w - leftW - rightW
	if centerW == 0 {
		return strings.Join(padRows([]string{
			line(t.panel, w, ""),
			label(t.panel.Bold(true), w, "  "+m.playerStatus()+" · "+title),
			m.progress(t.panel, w, track.Duration, 0, top+2),
			m.compactControls(w, top+3),
		}, t.panel, w, h), "\n")
	}
	height := h - 1
	artWidth := playerCoverWidth(w)
	titleWidth := leftW - artWidth - 3
	metadata := []string{
		label(t.panel.Bold(true), titleWidth, "  "+title) + m.control(t.panel, favoriteAction, artWidth+titleWidth, top+1, 3, " "+m.favorite()+" "),
		label(t.muted.Background(t.panel.GetBackground()), leftW-artWidth, "  "+track.Artist),
		label(t.dim.Background(t.panel.GetBackground()), leftW-artWidth, "  "+m.playerStatus()+" · Mock Music"),
	}
	left := lipgloss.JoinHorizontal(lipgloss.Top, m.cover(max(0, m.player.Index()), artWidth, height),
		strings.Join(padRows(metadata, t.panel, leftW-artWidth, height), "\n"))
	center := []string{line(t.panel, centerW, ""), m.controls(centerW, leftW, top+2),
		m.progress(t.panel, centerW, track.Duration, leftW, top+3)}
	rightX := leftW + centerW
	right := []string{
		line(t.panel, rightW, ""),
		m.volumeControl(rightW, rightX, top+2),
		m.control(t.panel, queueAction, rightX, top+3, rightW, "  u Queue"),
	}
	dock := lipgloss.JoinHorizontal(lipgloss.Top, left,
		strings.Join(padRows(center, t.panel, centerW, height), "\n"),
		strings.Join(padRows(right, t.panel, rightW, height), "\n"))
	return line(t.panel, w, "") + "\n" + dock
}

func (m model) controls(width, x, y int) string {
	t := m.theme
	state := func(on bool) string {
		if on {
			return "●"
		}
		return "○"
	}
	const controlsWidth = 19
	if width < controlsWidth {
		return m.compactControlsAt(width, x, y)
	}
	left := (width - controlsWidth) / 2
	start := x + left
	return line(t.panel, left, "") +
		m.control(t.panel, shuffleAction, start, y, 4, state(m.player.Shuffle())+"⇄") +
		m.control(t.panel, prevAction, start+4, y, 4, " ‹ ") +
		m.control(t.button, playAction, start+8, y, 3, " "+m.playSymbol()+" ") +
		m.control(t.panel, nextAction, start+11, y, 4, "  › ") +
		m.control(t.panel, repeatAction, start+15, y, 4, " ↻"+state(m.player.Repeat())) +
		line(t.panel, width-left-controlsWidth, "")
}

func (m model) compactControls(width, y int) string { return m.compactControlsAt(width, 0, y) }

func (m model) compactControlsAt(width, x, y int) string {
	t := m.theme
	if width < 9 {
		return m.control(t.button, playAction, x, y, width, m.playSymbol())
	}
	controls := m.control(t.panel, prevAction, x, y, 3, " ‹ ") +
		m.control(t.button, playAction, x+3, y, 3, " "+m.playSymbol()+" ") +
		m.control(t.panel, nextAction, x+6, y, 3, " › ")
	if width >= 23 {
		return controls + m.volumeControl(width-9, x+9, y)
	}
	return controls + line(t.panel, width-9, "")
}

func (m model) volumeControl(width, x, y int) string {
	t := m.theme
	prefix := fmt.Sprintf("  Vol %3d%%  ", m.player.Volume())
	start := lipgloss.Width(prefix)
	count := min(10, max(0, width-start))
	if count == 0 {
		return label(t.panel, width, prefix)
	}
	return line(t.panel, start, prefix) + m.control(t.panel, volumeAction, x+start, y, count, volumeMeter(m.player.Volume(), count)) +
		line(t.panel, width-start-count, "")
}

func playerCoverWidth(width int) int {
	if width < 100 {
		return 6
	}
	return 8
}

func volumeMeter(volume, cells int) string {
	filled := (volume*cells + 99) / 100
	return strings.Repeat("━", filled) + strings.Repeat("─", cells-filled)
}

func (m model) progress(style lipgloss.Style, width int, duration time.Duration, x, y int) string {
	times := formatTime(m.player.Position())
	total := formatTime(duration)
	n := width - lipgloss.Width(times) - lipgloss.Width(total) - 6
	if n <= 0 {
		return label(style, width, "  "+times+" / "+total)
	}
	filled := 0
	if duration > 0 {
		filled = min(n, max(0, int(float64(n)*float64(m.player.Position())/float64(duration))))
	}
	if duration > 0 {
		m.region(seekAction, 0, x+lipgloss.Width(times)+3, y, n, 1)
	}
	style = m.hoverStyle(style, seekAction, 0)
	bar := m.theme.accent.Background(style.GetBackground()).Render(strings.Repeat("━", filled)) +
		m.theme.dim.Background(style.GetBackground()).Render(strings.Repeat("─", n-filled))
	return style.Render("  "+times+" ") + bar + style.Render(" "+total+"  ")
}
