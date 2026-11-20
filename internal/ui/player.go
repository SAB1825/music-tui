package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

func (m model) playerBar() string {
	t, w := m.theme, m.width
	track, ok := m.player.Track()
	status := "Paused"
	if m.player.Playing() {
		status = "Playing"
	}
	if !ok {
		status = "Idle"
	}
	title := "No track queued"
	if ok {
		title = track.Title
	}
	now := "  " + status + "  ·  " + title
	if ok && w >= 48 {
		now += "  /  " + track.Artist
	}
	if m.playerHeight() == 1 {
		return line(t.panel.Bold(true), w, strings.TrimSpace(now))
	}
	controls := fmt.Sprintf("  Space Play/Pause  ·  p/n Skip  ·  +/- Volume %d%%", m.player.Volume())
	if w < 56 {
		controls = fmt.Sprintf("  ␣ Play/Pause  p/n  ± %d%%", m.player.Volume())
	}
	if w < 30 {
		controls = fmt.Sprintf("␣ Play p/n ± %d%%", m.player.Volume())
	}
	if m.playerHeight() == 2 {
		return line(t.panel.Bold(true), w, now) + "\n" + line(t.panel, w, controls)
	}

	times := formatTime(m.player.Position()) + " / " + formatTime(track.Duration)
	barWidth := min(44, w-ansi.StringWidth(times)-5)
	progress := "  " + times
	if barWidth > 0 {
		filled := 0
		if track.Duration > 0 {
			filled = int(float64(barWidth) * float64(m.player.Position()) / float64(track.Duration))
		}
		progress += "  " + strings.Repeat("━", filled) + strings.Repeat("─", barWidth-filled)
	}
	return strings.Join([]string{
		line(t.rule, w, strings.Repeat("─", w)),
		line(t.panel.Bold(true), w, now),
		line(t.accent, w, progress),
		line(t.panel, w, controls),
	}, "\n")
}
