package ui

import (
	"fmt"
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"
	"github.com/SAB1825/music-tui/internal/config"
	"github.com/SAB1825/music-tui/internal/domain"
	"github.com/SAB1825/music-tui/internal/visualizer"
)

type spectrumEvent struct {
	visualizer.Event
	open bool
}

// WithVisualizer attaches the optional backend to the initial UI model.
func WithVisualizer(tracks []domain.Track, cfg config.Visualizer, events <-chan visualizer.Event, configErr error) tea.Model {
	m := New(tracks).(model)
	m.spectrumEnabled, m.spectrumEvents = cfg.Enabled, events
	m.spectrumConfig = cfg
	if configErr != nil {
		m.spectrumError = configErr.Error()
	}
	return m
}

func receiveSpectrum(events <-chan visualizer.Event) tea.Cmd {
	if events == nil {
		return nil
	}
	return func() tea.Msg {
		event, ok := <-events
		return spectrumEvent{Event: event, open: ok}
	}
}

func (m model) settingsPanel(width, height int) string {
	t := m.theme
	status := "Disabled in config.toml"
	if m.spectrumEnabled {
		monitor := m.spectrumConfig.Source
		if monitor == "auto" {
			monitor = "default output monitor"
		}
		status = fmt.Sprintf("CAVA · %d bars at %d fps · %s",
			m.spectrumConfig.Bars, m.spectrumConfig.Framerate, monitor)
	}
	rows := []string{
		line(t.heading, width, "  Settings"),
		line(t.base, width, ""),
		line(t.accent, width, "  AUDIO SPECTRUM"),
		label(t.muted, width, "  "+status),
		label(t.muted, width, "  Captures shared system output. This player simulates playback."),
	}
	if m.spectrumError != "" {
		if !m.spectrumEnabled {
			rows[3] = label(t.warning, width, "  Visualizer configuration unavailable")
		}
		details := strings.FieldsFunc(m.spectrumError, unicode.IsControl)
		for _, detail := range details {
			rows = append(rows, label(t.warning, width, "  "+detail))
		}
	}
	rows = append(rows,
		label(t.dim, width, "  Configure in $XDG_CONFIG_HOME/music-tui/config.toml and restart."),
		line(t.base, width, ""),
		line(t.accent, width, "  ARTWORK"),
		label(t.muted, width, "  Terminal-cell covers"),
	)
	return strings.Join(padRows(rows, t.base, width, height), "\n")
}
