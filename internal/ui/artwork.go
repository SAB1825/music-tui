package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/SAB1825/music-tui/internal/artwork"
)

type coverSize struct{ width, height int }

func (m *model) prepareCovers() {
	l := m.layout()
	d := homeDimensions(l.content(), l.body)
	sizes := []coverSize{{3, 1}, {playerCoverWidth(l.width), max(1, l.player-1)}, {d.coverWidth, d.cardHeight - 2}}
	sizes = append(sizes, discoveryCover(l.content(), d))
	m.covers = make(map[coverSize][4]string, len(sizes))
	for _, size := range sizes {
		if _, exists := m.covers[size]; exists {
			continue
		}
		var variants [4]string
		for i := range variants {
			variants[i] = paintCover(m.theme.covers[i], i, size)
		}
		m.covers[size] = variants
	}
}

func (m model) cover(index, width, height int) string {
	if variants, ok := m.covers[coverSize{width, height}]; ok {
		return variants[max(0, index)%len(variants)]
	}
	return solid(m.theme.elevated, width, height)
}

func paintCover(palette [4]string, design int, size coverSize) string {
	cells := artwork.Raster(design, size.width, size.height*2)
	var styles [4][4]lipgloss.Style
	for top := range palette {
		for bottom := range palette {
			styles[top][bottom] = lipgloss.NewStyle().
				Foreground(lipgloss.Color(palette[top])).Background(lipgloss.Color(palette[bottom]))
		}
	}
	var cover strings.Builder
	for y := range size.height {
		if y > 0 {
			cover.WriteByte('\n')
		}
		top := cells[y*2*size.width:]
		bottom := top[size.width:]
		for x := 0; x < size.width; {
			start := x
			upper, lower := top[x], bottom[x]
			for x < size.width && top[x] == upper && bottom[x] == lower {
				x++
			}
			cover.WriteString(styles[upper][lower].Render(strings.Repeat("▀", x-start)))
		}
	}
	return cover.String()
}
