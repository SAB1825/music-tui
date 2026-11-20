package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

type homeLayout struct {
	coverWidth, cardHeight, discoveryHeight, discoveryRows int
}

func homeDimensions(width, height int) homeLayout {
	switch {
	case width >= 110 && height >= 32:
		return homeLayout{20, 11, 7, 2}
	case height >= 23:
		return homeLayout{16, 9, 7, 1}
	default:
		return homeLayout{12, 6, 4, 1}
	}
}

func (m model) homePanel(width, height int) string {
	t := m.theme
	d := homeDimensions(width, height)
	header := []string{
		line(t.base, width, ""),
		line(t.accent.Background(t.base.GetBackground()), width, "  Still up, Night Owl?"),
		line(t.muted.Background(t.base.GetBackground()), width, "  Your late-night soundtrack"),
	}
	rows := append(header, strings.Split(m.featureCard(width, d), "\n")...)
	rows = append(rows, line(t.base, width, ""), m.filters(width))
	rows = append(rows, strings.Split(m.discovery(width, d), "\n")...)
	rows = padRows(rows, t.base, width, height-1)
	return strings.Join(append(rows, line(t.dim, width, "  ←→ Browse cards  ·  Enter Play  ·  c Filters  ·  Tab Menu  ·  ? Help")), "\n")
}

func (m model) featureCard(width int, d homeLayout) string {
	t := m.theme
	track, ok := m.player.Track()
	if !ok {
		return strings.Join(padRows([]string{line(t.card, width, "  No track queued")}, t.card, width, d.cardHeight), "\n")
	}
	spectrumWidth := 0
	if width >= 70 {
		spectrumWidth = min(36, width/3)
	}
	art := m.cover(m.player.Index(), d.coverWidth, d.cardHeight-2)
	metadataWidth := width - d.coverWidth - spectrumWidth - 10
	x, y := m.layout().sidebar+4+d.coverWidth, m.layout().toolbar+4
	if d.cardHeight >= 9 {
		y++
	}
	info := []string{
		label(t.card, metadataWidth, "  "+m.playerStatus()+"  ·  "+track.Album),
		label(t.card.Bold(true), metadataWidth, "  "+track.Title),
		label(t.muted.Background(t.card.GetBackground()), metadataWidth, "  "+track.Artist),
		m.progress(t.card, metadataWidth, track.Duration, x, y+3),
	}
	if d.cardHeight == 6 {
		info[2] = m.featureControls(metadataWidth, x, y+2, track.Artist)
	}
	if d.cardHeight >= 9 {
		info = append(info, m.featureControls(metadataWidth, x, y+4, ""))
		info = append([]string{line(t.card, metadataWidth, "")}, info...)
	}
	info = padRows(info, t.card, metadataWidth, d.cardHeight-2)
	art = line(t.card, d.coverWidth, "") + "\n" + art + "\n" + line(t.card, d.coverWidth, "")
	meta := line(t.card, metadataWidth, "") + "\n" + strings.Join(info, "\n") + "\n" + line(t.card, metadataWidth, "")
	blocks := []string{solid(t.base, 2, d.cardHeight), solid(t.card, 2, d.cardHeight), art, meta}
	if spectrumWidth > 0 {
		blocks = append(blocks, line(t.card, spectrumWidth, "")+"\n"+
			m.spectrum(spectrumWidth, d.cardHeight-2)+"\n"+line(t.card, spectrumWidth, ""))
	}
	blocks = append(blocks, solid(t.card, 4, d.cardHeight), solid(t.base, 2, d.cardHeight))
	return lipgloss.JoinHorizontal(lipgloss.Top, blocks...)
}

func solid(style lipgloss.Style, width, height int) string {
	row := line(style, width, "")
	return strings.TrimSuffix(strings.Repeat(row+"\n", height), "\n")
}

func (m model) filters(width int) string {
	t := m.theme
	all := categories()
	slots := min(len(all), max(1, (width-2)/9))
	first := min(max(0, m.category-slots+1), len(all)-slots)
	var row strings.Builder
	row.WriteString(line(t.base, 2, ""))
	for i := first; i < first+slots; i++ {
		style := t.base
		if m.category == i {
			style = t.selected
			if m.focus == categoryFocus {
				style = t.button
			}
		}
		m.region(filterAction, i, m.layout().sidebar+2+(i-first)*9, m.layout().toolbar+4+homeDimensions(width, m.bodyHeight()).cardHeight, 8, 1)
		style = m.hoverStyle(style, filterAction, i)
		row.WriteString(line(style.Align(lipgloss.Center), 8, all[i]))
		row.WriteString(line(t.base, 1, ""))
	}
	return line(t.base, width, row.String())
}

func discoveryColumns(width int) int { return min(4, max(1, (width-2)/25)) }

func discoveryCardWidth(width, columns int) int { return (width - 2 - (columns-1)*2) / columns }

func discoveryCover(width int, d homeLayout) coverSize {
	if d.discoveryHeight <= 4 {
		return coverSize{8, d.discoveryHeight}
	}
	return coverSize{discoveryCardWidth(width, discoveryColumns(width)), d.discoveryHeight - 2}
}

func (m model) discovery(width int, d homeLayout) string {
	t := m.theme
	matches := m.matches()
	if len(matches) == 0 {
		return line(t.muted, width, "  No tracks in this category")
	}
	columns := discoveryColumns(width)
	visible := columns * d.discoveryRows
	start := m.trackStart(len(matches), visible)
	cardWidth := discoveryCardWidth(width, columns)
	var rows []string
	for first := start; first < min(len(matches), start+visible); first += columns {
		var cards []string
		cards = append(cards, solid(t.base, 2, d.discoveryHeight))
		for i := first; i < min(len(matches), first+columns); i++ {
			if i > first {
				cards = append(cards, solid(t.base, 2, d.discoveryHeight))
			}
			x := m.layout().sidebar + 2 + (i-first)*(cardWidth+2)
			y := m.layout().toolbar + 5 + d.cardHeight + ((first-start)/columns)*(d.discoveryHeight+1)
			m.region(trackAction, i, x, y, cardWidth, d.discoveryHeight)
			cards = append(cards, m.discoveryCard(matches[i], cardWidth, d.discoveryHeight, i))
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, cards...))
		if first+columns < min(len(matches), start+visible) {
			rows = append(rows, line(t.base, width, ""))
		}
	}
	return strings.Join(rows, "\n")
}

func (m model) discoveryCard(index, width, height int, selection int) string {
	t := m.theme
	track := m.catalog[index]
	style, marker := t.base, "  "
	if selection == m.selected {
		style, marker = t.elevated, "› "
		if m.focus == tracksFocus {
			style = t.selected
		}
	}
	style = m.hoverStyle(style, trackAction, selection)
	if m.player.Playing() && m.player.Index() == index && marker == "  " {
		marker = "▶ "
	}
	if height <= 4 {
		cover := m.cover(index, 8, height)
		text := padRows([]string{line(style, width-8, ""), label(style.Bold(true), width-8, marker+track.Title),
			label(style, width-8, "  "+track.Artist)}, style, width-8, height)
		return lipgloss.JoinHorizontal(lipgloss.Top, cover, strings.Join(text, "\n"))
	}
	art := m.cover(index, width, height-2)
	return strings.Join([]string{art, label(style.Bold(true), width, marker+track.Title),
		label(style, width, "  "+track.Artist)}, "\n")
}

func (m model) featureControls(width, x, y int, artist string) string {
	t := m.theme
	return line(t.card, 2, "") +
		m.control(t.card, prevAction, x+2, y, 3, " ‹ ") +
		m.control(t.button, playAction, x+5, y, 3, " "+m.playSymbol()+" ") +
		m.control(t.card, nextAction, x+8, y, 3, " › ") +
		m.control(t.card, favoriteAction, x+11, y, 4, "  "+m.favorite()) +
		label(t.muted.Background(t.card.GetBackground()), width-15, artist)
}
