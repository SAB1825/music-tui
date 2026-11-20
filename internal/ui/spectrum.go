package ui

import (
	"strings"
)

func (m model) spectrum(width, height int) string {
	t := m.theme
	rows := []string{
		label(t.accent.Background(t.inset.GetBackground()), width, "  SPECTRUM"),
	}
	bars := spectrumBars(m.frame, min(32, max(0, width-4)))
	glyphs := []rune(" ▁▂▃▄▅▆▇█")
	plotHeight := max(0, height-2)
	style := t.accent.Background(t.inset.GetBackground())
	for level := plotHeight - 1; level >= 0; level-- {
		var row strings.Builder
		row.WriteString("  ")
		for _, value := range bars {
			row.WriteRune(glyphs[barHeight(value, plotHeight, level)])
		}
		if level == 0 && strings.TrimSpace(row.String()) == "" {
			rows = append(rows, line(t.dim.Background(t.inset.GetBackground()), width, "  "+strings.Repeat("╌", min(32, max(0, width-4)))))
			continue
		}
		rows = append(rows, line(style, width, row.String()))
	}
	status := "No output signal"
	switch {
	case m.spectrumError != "":
		status = "Unavailable · [,]"
	case !m.spectrumEnabled:
		status = "Disabled · [,]"
	case m.frame == nil:
		status = "Waiting for output"
	default:
		for _, value := range m.frame {
			if value > 0 {
				status = "System output · CAVA"
				break
			}
		}
	}
	rows = append(rows, label(t.muted.Background(t.inset.GetBackground()), width, "  "+status))
	return strings.Join(padRows(rows, t.inset, width, height), "\n")
}

func spectrumBars(frame []float64, count int) []float64 {
	count = min(count, len(frame))
	if count <= 0 {
		return nil
	}
	bars := make([]float64, count)
	for i := range bars {
		for _, value := range frame[i*len(frame)/count : (i+1)*len(frame)/count] {
			bars[i] = max(bars[i], value)
		}
	}
	return bars
}

func barHeight(value float64, height, level int) int {
	return min(8, max(0, int(value*float64(height*8)+.5)-level*8))
}
