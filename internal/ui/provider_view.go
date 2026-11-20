package ui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

func (m model) onlinePanel(w, h int) string {
	t, s := m.theme, m.online
	title := "Discover on " + m.providerName()
	if s.playlist.ID != "" {
		title = s.playlist.Title
	} else if s.query != "" {
		title = "Search results"
	}
	rows := []string{label(t.heading, w, "  "+catalogText(title))}
	description := "Search music videos or open a public playlist."
	switch {
	case s.loading && !s.paginating && s.playlistRequest:
		description = "Loading public playlist…"
	case s.loading && !s.paginating:
		description = "Searching " + m.providerName() + "…"
	case s.loading:
		description = "Loading next page…"
	case s.err != "":
		description = s.err
	case s.playlist.ID != "":
		description = "Public playlist · " + m.providerName()
		if s.playlist.TrackCount != nil {
			description += fmt.Sprintf(" · %d videos", *s.playlist.TrackCount)
		}
	case s.query != "":
		description = fmt.Sprintf("%d videos · %s · %s", len(s.tracks), m.providerName(), s.query)
	}
	if img := s.images[s.playlist.ArtworkURL]; img != "" && w >= 15 {
		rows = append(rows, line(t.base, 2, "")+img+label(t.muted, w-8, "  "+catalogText(description)))
	} else {
		rows = append(rows, label(t.muted, w, "  "+catalogText(description)))
	}
	if h < 3 {
		return strings.Join(padRows(rows, t.base, w, h), "\n")
	}
	if len(s.tracks) == 0 {
		switch {
		case s.err != "":
			rows = append(rows, label(t.dim, w, "  / Search again  ·  v Music providers"))
		case !s.loading && s.query != "":
			rows = append(rows, label(t.dim, w, "  No results found. Try a different search."))
		case s.query == "":
			rows = append(rows, label(t.dim, w, "  / Search  ·  g Public playlist  ·  v Music providers"))
		}
		if s.query == "" && w >= 80 && h >= 12 {
			rows = append(rows, strings.Split(m.spectrum(min(32, w), 5), "\n")...)
		}
	}
	start := m.trackStart(len(s.tracks), max(0, h-4))
	for i := start; i < min(len(s.tracks), start+max(0, h-4)); i++ {
		track := s.tracks[i]
		style, marker := t.base, "  "
		if i == m.selected {
			style, marker = t.selected, "› "
		}
		style = m.hoverStyle(style, trackAction, i)
		creator := track.Creator
		if creator != "" {
			creator = " · " + creator
		}
		duration := ""
		if track.Unavailable {
			duration = "Unavailable"
		} else if track.Duration > 0 {
			duration = formatTime(track.Duration)
		}
		meta := m.providerName()
		if duration != "" {
			meta += " · " + duration
		}
		m.region(trackAction, i, m.layout().sidebar, m.layout().toolbar+len(rows), w, 1)
		img := s.images[track.ArtworkURL]
		if img == "" {
			img = line(style, 6, " ▧")
		}
		if w >= 36 {
			right := min(26, lipgloss.Width(meta)+2)
			rows = append(rows, line(style, 2, marker)+img+label(style, w-right-8, "  "+catalogText(track.Title+creator))+label(style, right, meta))
		} else {
			rows = append(rows, label(style, w, marker+catalogText(track.Title)))
		}
	}
	rows = padRows(rows, t.base, w, h-1)
	var footer string
	x, y := m.layout().sidebar, m.layout().toolbar+h-1
	if s.next != "" && !s.loading && w >= 16 {
		footer = m.control(t.elevated, moreAction, x, y, 16, "  +  More") +
			label(t.dim, w-16, "  ↑↓ Select  ·  Enter Open in YouTube")
	} else {
		footer = label(t.dim, w, "  ↑↓ Select  ·  Enter Open in YouTube  ·  / Search")
	}
	return strings.Join(append(rows, footer), "\n")
}

func (m model) catalogBar() string {
	t, w, h := m.theme, m.width, m.playerHeight()
	rows := []string{label(t.panel, w, "  "+m.providerName()+" · Online catalog")}

	if m.selected < len(m.online.tracks) && h >= 2 {
		track := m.online.tracks[m.selected]
		rows = append(rows, label(t.panel.Bold(true), w, "  "+catalogText(track.Title)))
		if w >= 20 && h >= 3 && track.ExternalURL != "" {
			rows = append(rows, m.control(t.button, openVideoAction, 0, m.height-h+len(rows), 20, "  Open in YouTube")+line(t.panel, w-20, ""))
		}
	}
	if len(rows) < h && w >= 26 && m.online.playlist.ExternalURL != "" {
		rows = append(rows, m.control(t.panel, openPlaylistAction, 0, m.height-h+len(rows), 26, "  Open playlist in YouTube")+
			label(t.muted, w-26, "  "+catalogText(m.online.playlist.Description)))
	}
	if len(rows) < h {
		rows = append(rows, label(t.panel, w, "  Videos open in your browser"))
	}
	return strings.Join(padRows(rows, t.panel, w, h), "\n")
}
