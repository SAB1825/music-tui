// Package config reads the optional XDG configuration file.
package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// Visualizer configures CAVA's local output monitor.
type Visualizer struct {
	Enabled   bool   `toml:"enabled"`
	Backend   string `toml:"backend"`
	Bars      int    `toml:"bars"`
	Framerate int    `toml:"framerate"`
	Source    string `toml:"source"`
}

type file struct {
	Visualizer Visualizer `toml:"visualizer"`
}

// Defaults enables the local monitor with a compact spectrum.
func Defaults() Visualizer {
	return Visualizer{Enabled: true, Backend: "cava", Bars: 24, Framerate: 20, Source: "auto"}
}

// Validate checks the supported backend and display limits.
func (v Visualizer) Validate() error {
	switch {
	case v.Backend != "cava":
		return fmt.Errorf("visualizer.backend must be cava")
	case v.Bars < 8 || v.Bars > 64:
		return fmt.Errorf("visualizer.bars must be between 8 and 64")
	case v.Framerate < 15 || v.Framerate > 25:
		return fmt.Errorf("visualizer.framerate must be between 15 and 25")
	case v.Source != "auto" && !monitorName(v.Source):
		return fmt.Errorf("visualizer.source must be auto or a PulseAudio .monitor source")
	default:
		return nil
	}
}

func monitorName(name string) bool {
	if !strings.HasSuffix(name, ".monitor") || len(name) <= len(".monitor") || len(name) > 255 {
		return false
	}
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', strings.ContainsRune("._-", r):
		default:
			return false
		}
	}
	return true
}

// Load reads config.toml without creating or changing it.
func Load() (Visualizer, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return Visualizer{}, fmt.Errorf("locate configuration: %w", err)
	}
	path := filepath.Join(dir, "music-tui", "config.toml")
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return Defaults(), nil
	}
	if err != nil {
		return Visualizer{}, fmt.Errorf("open %s: %w", path, err)
	}
	v, readErr := decode(f)
	if err := errors.Join(readErr, f.Close()); err != nil {
		return Visualizer{}, fmt.Errorf("read %s: %w", path, err)
	}
	return v, nil
}

func decode(r io.Reader) (Visualizer, error) {
	cfg := file{Visualizer: Defaults()}
	if err := toml.NewDecoder(r).DisallowUnknownFields().Decode(&cfg); err != nil {
		return Visualizer{}, err
	}
	return cfg.Visualizer, cfg.Visualizer.Validate()
}
