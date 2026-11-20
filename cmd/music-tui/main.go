package main

import (
	"errors"
	"log/slog"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/SAB1825/music-tui/internal/providers/mock"
	"github.com/SAB1825/music-tui/internal/ui"
)

func main() {
	if _, err := tea.NewProgram(ui.New(mock.Catalog())).Run(); err != nil && !errors.Is(err, tea.ErrInterrupted) {
		slog.Error("run music TUI", "error", err)
		os.Exit(1)
	}
}
