package config

import (
	"strings"
	"testing"
)

func TestDecodeVisualizer(t *testing.T) {
	for _, tc := range []struct {
		input string
		ok    bool
	}{
		{"[visualizer]\nenabled = false\nbars = 16 # compact", true},
		{"[visualizer]\nbackend = 'other'", false},
		{"[visualizer]\nframerate = 30", false},
		{"[visualizer]\nbars = 0", false},
		{"[visualizer]\nsource = 'alsa_input.microphone'", false},
		{"[visualizer]\nsource = 'fixture.monitor'", true},
		{"[visualizer]\nenabledd = true", false},
	} {
		loaded, err := decode(strings.NewReader(tc.input))
		cfg := loaded.Visualizer
		if (err == nil) != tc.ok {
			t.Fatalf("%q: got %v", tc.input, err)
		}
		if strings.Contains(tc.input, "enabled = false") && (cfg.Enabled || cfg.Bars != 16 || cfg.Framerate != 20 || cfg.Backend != "cava") {
			t.Fatalf("decoded configuration with defaults: %+v", cfg)
		}
	}
}
