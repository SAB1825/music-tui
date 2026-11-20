package main

import (
	"context"
	"errors"
	"log/slog"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/SAB1825/music-tui/internal/config"
	"github.com/SAB1825/music-tui/internal/providers"
	"github.com/SAB1825/music-tui/internal/providers/mock"
	"github.com/SAB1825/music-tui/internal/providers/youtube"
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
	if configErr == nil && cfg.Visualizer.Enabled {
		events = visualizer.Start(ctx, cfg.Visualizer)
	}
	defer func() {
		cancel()
		if events != nil {
			for range events {
			}
		}
	}()
	registry := providerRegistry(cfg.Provider, configErr)
	initial := ui.WithVisualizer(mock.Catalog(), cfg.Visualizer, events, configErr)
	_, err := tea.NewProgram(ui.WithProviders(initial, ctx, registry, cfg.Provider.Default)).Run()
	return err
}

func providerRegistry(cfg config.Provider, configErr error) providers.Registry {
	var registry providers.Registry
	entry := providers.Entry{ID: "youtube", Name: "YouTube"}
	switch {
	case configErr != nil:
		entry.Unavailable = "Check the application configuration in Settings."
	case !cfg.YouTube.Enabled:
		entry.Unavailable = "YouTube is disabled in config.toml."
	default:
		key, err := config.YouTubeKey(".env")
		switch {
		case err != nil:
			entry.Unavailable = err.Error()
		case key == "":
			entry.Unavailable = "Set YOUTUBE_API_KEY in your environment or local .env file."
		default:
			entry.Catalog = youtube.New(key, "", nil, cfg.YouTube.Timeout(), cfg.YouTube.SearchLimit)
		}
	}
	registry.Register(entry)
	registry.Register(providers.Entry{ID: "mock", Name: "Mock", Catalog: mock.Provider{}})
	return registry
}
