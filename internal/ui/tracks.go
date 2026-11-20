package ui

import (
	"fmt"
	"time"

	"github.com/SAB1825/music-tui/internal/domain"
	"github.com/charmbracelet/x/ansi"
)

func (m model) trackLine(width, index int, track domain.Track) string {
	style, marker := m.theme.base, " "
	if index == m.selected {
		marker = "›"
		style = m.theme.heading
		if m.focus == tracksFocus {
			style = m.theme.selected
		}
	}
	style = m.hoverStyle(style, trackAction, index)
	if m.page == queue && m.search.Value() == "" && index == m.player.Index() && index != m.selected {
		marker = "•"
	}
	if m.player.Playing() && track == m.catalog[m.player.Index()] && marker == " " {
		marker = "▶"
	}
	title := fmt.Sprintf("  %s %02d  %s", marker, index+1, track.Title)
	duration := formatTime(track.Duration)
	if width >= 64 {
		titleWidth := width * 42 / 100
		artistWidth := width * 24 / 100
		albumWidth := width - titleWidth - artistWidth - 7
		return line(style, titleWidth, title) +
			line(style, artistWidth, ansi.Truncate(track.Artist, artistWidth-1, "…")) +
			line(style, albumWidth, ansi.Truncate(track.Album, albumWidth-1, "…")) +
			line(style, 7, duration)
	}
	if width >= 30 {
		return line(style, width-7, ansi.Truncate(title+" · "+track.Artist, width-8, "…")) +
			line(style, 7, duration)
	}
	return line(style, width, title)
}

func formatTime(d time.Duration) string {
	return fmt.Sprintf("%d:%02d", d/time.Minute, d/time.Second%60)
}
