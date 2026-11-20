package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestYouTubeKeyPrecedence(t *testing.T) {
	t.Setenv("YOUTUBE_API_KEY", "")
	if err := os.Unsetenv("YOUTUBE_API_KEY"); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), ".env")
	if value, err := YouTubeKey(path); err != nil || value != "" {
		t.Fatal("missing file should produce no credential")
	}
	if err := os.WriteFile(path, []byte("OTHER=test\nexport YOUTUBE_API_KEY = 'file-value' # example\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if value, err := YouTubeKey(path); err != nil || value != "file-value" {
		t.Fatal("dotenv assignment was not loaded")
	}
	t.Setenv("YOUTUBE_API_KEY", "environment-value")
	if value, err := YouTubeKey(path); err != nil || value != "environment-value" {
		t.Fatal("environment did not take precedence")
	}
	t.Setenv("YOUTUBE_API_KEY", "")
	if value, err := YouTubeKey(path); err != nil || value != "" {
		t.Fatal("explicit empty environment variable did not disable credentials")
	}
}

func TestProviderSettings(t *testing.T) {
	cfg, err := decode(strings.NewReader("[visualizer]\nenabled=false\n[provider]\ndefault='mock'\n[provider.youtube]\nsearch_limit=17"))
	if err != nil || cfg.Provider.Default != "mock" || !cfg.Provider.YouTube.Enabled || cfg.Provider.YouTube.SearchLimit != 17 || cfg.Visualizer.Enabled {
		t.Fatalf("configuration merge failed: %v", err)
	}
	for _, line := range []string{"search_limit=51", "request_timeout='5ms'", "api_key='must-not-leak'"} {
		_, err := decode(strings.NewReader("[provider.youtube]\n" + line))
		if err == nil || strings.Contains(err.Error(), "must-not-leak") {
			t.Fatal("incorrect or unsafe validation error")
		}
	}
}
