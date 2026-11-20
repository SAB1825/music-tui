package main

import (
	"context"
	"errors"
	"log/slog"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/SAB1825/music-tui/internal/config"
	"github.com/SAB1825/music-tui/internal/providers/mock"
	"github.com/SAB1825/music-tui/internal/ui"
	"github.com/SAB1825/music-tui/internal/visualizer"
)

func main() {
	if err := run(); err != nil && !errors.Is(err, tea.ErrInterrupted) {
		slog.Error("run music TUI", "error", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	cfg, configErr := config.Load()
	var events <-chan visualizer.Event
	if configErr == nil && cfg.Enabled {
		events = visualizer.Start(ctx, cfg)
	}
	defer func() {
		cancel()
		if events != nil {
			for range events {
			}
		}
	}()
	_, err := tea.NewProgram(ui.WithVisualizer(mock.Catalog(), cfg, events, configErr)).Run()
	return err
}
