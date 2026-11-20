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
	t, w, h := m.theme, m.width, m.playerHeight()
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
			label(t.panel, w, fmt.Sprintf("␣ %s   p/n Skip   ± %d%%", m.playSymbol(), m.player.Volume()))
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
			line(t.rule, w, strings.Repeat("─", w)),
			label(t.panel.Bold(true), w, "  "+m.playerStatus()+" · "+title),
			m.progress(t.panel, w, track.Duration),
			label(t.panel, w, fmt.Sprintf("  ␣ %s  p/n  ± %d%%", m.playSymbol(), m.player.Volume())),
		}, t.panel, w, h), "\n")
	}
	height := h - 1
	artWidth := playerCoverWidth(w)
	metadata := []string{
		label(t.panel.Bold(true), leftW-artWidth, "  "+title+"  "+m.favorite()),
		label(t.muted.Background(t.panel.GetBackground()), leftW-artWidth, "  "+track.Artist),
		label(t.dim.Background(t.panel.GetBackground()), leftW-artWidth, "  "+m.playerStatus()+" · Mock Music"),
	}
	left := lipgloss.JoinHorizontal(lipgloss.Top, m.cover(max(0, m.player.Index()), artWidth, height),
		strings.Join(padRows(metadata, t.panel, leftW-artWidth, height), "\n"))
	center := []string{line(t.panel, centerW, ""), m.controls(centerW), m.progress(t.panel, centerW, track.Duration)}
	utilities := "  u Queue  ·  , Settings"
	if rightW < 24 {
		utilities = "  u Queue"
	}
	right := []string{
		line(t.panel, rightW, ""),
		label(t.panel.Bold(true), rightW, fmt.Sprintf("  Vol %3d%%  %s", m.player.Volume(), volumeMeter(m.player.Volume(), min(10, max(0, rightW-12))))),
		label(t.muted.Background(t.panel.GetBackground()), rightW, utilities),
	}
	dock := lipgloss.JoinHorizontal(lipgloss.Top, left,
		strings.Join(padRows(center, t.panel, centerW, height), "\n"),
		strings.Join(padRows(right, t.panel, rightW, height), "\n"))
	return line(t.rule, w, strings.Repeat("─", w)) + "\n" + dock
}

func (m model) controls(width int) string {
	t := m.theme
	state := func(on bool) string {
		if on {
			return "●"
		}
		return "○"
	}
	shuffle, repeat := state(m.player.Shuffle()), state(m.player.Repeat())
	side := (width - 3) / 2
	return t.panel.Width(side).Align(lipgloss.Right).Render(shuffle+"⇄   ‹  ") +
		line(t.button, 3, " "+m.playSymbol()+" ") +
		line(t.panel, width-side-3, "  ›   ↻"+repeat)
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

func (m model) progress(style lipgloss.Style, width int, duration time.Duration) string {
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
	bar := m.theme.accent.Background(style.GetBackground()).Render(strings.Repeat("━", filled)) +
		m.theme.dim.Background(style.GetBackground()).Render(strings.Repeat("─", n-filled))
	return style.Render("  "+times+" ") + bar + style.Render(" "+total+"  ")
}
