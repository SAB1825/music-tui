package ui

import (
	"fmt"

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

func (m model) settingsDetails() []string {
	status := "Disabled in config.toml"
	if m.spectrumEnabled {
		monitor := m.spectrumConfig.Source
		if monitor == "auto" {
			monitor = "default output monitor"
		}
		status = fmt.Sprintf("CAVA · %d bars at %d fps", m.spectrumConfig.Bars, m.spectrumConfig.Framerate)
		status += " · " + monitor
	}
	if m.spectrumError != "" {
		status = "Spectrum unavailable. Choose Details."
	}
	return []string{
		"AUDIO SPECTRUM", status, "Captures shared system output.",
		"Playback in this app is simulated.", "",
		"Use $XDG_CONFIG_HOME/music-tui/config.toml.",
		"", "ARTWORK", "Terminal-cell covers",
	}
}
