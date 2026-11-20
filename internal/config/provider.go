package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

type Provider struct {
	Default string  `toml:"default"`
	YouTube YouTube `toml:"youtube"`
}

type YouTube struct {
	Enabled        bool   `toml:"enabled"`
	RequestTimeout string `toml:"request_timeout"`
	SearchLimit    int    `toml:"search_limit"`
}

func defaults() Config {
	return Config{Visualizer: Defaults(), Provider: Provider{
		Default: "auto", YouTube: YouTube{Enabled: true, RequestTimeout: "10s", SearchLimit: 20},
	}}
}

func (c Provider) validate() error {
	if c.Default != "auto" && c.Default != "mock" && c.Default != "youtube" {
		return errors.New("provider.default must be auto, mock, or youtube")
	}
	d, err := time.ParseDuration(c.YouTube.RequestTimeout)
	if err != nil || d < time.Second || d > time.Minute {
		return errors.New("provider.youtube.request_timeout must be between 1s and 1m")
	}
	if c.YouTube.SearchLimit < 1 || c.YouTube.SearchLimit > 50 {
		return errors.New("provider.youtube.search_limit must be between 1 and 50")
	}
	return nil
}

func (c YouTube) Timeout() time.Duration {
	d, _ := time.ParseDuration(c.RequestTimeout) // Validated during configuration loading.
	return d
}

// YouTubeKey reads a single dotenv assignment without sourcing shell code or
// adding credentials to the environment inherited by child processes.
func YouTubeKey(path string) (string, error) {
	const name = "YOUTUBE_API_KEY"
	if value, ok := os.LookupEnv(name); ok {
		return strings.TrimSpace(value), nil
	}
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", errors.New("could not open local .env")
	}
	var value string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(strings.TrimPrefix(scanner.Text(), "\uFEFF"))
		line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		field, right, ok := strings.Cut(line, "=")
		if !ok || strings.TrimSpace(field) != name {
			continue
		}
		right = strings.TrimSpace(right)
		if right != "" && (right[0] == '\'' || right[0] == '"') {
			end := strings.IndexByte(right[1:], right[0])
			if end < 0 {
				return "", errors.Join(errors.New("invalid YOUTUBE_API_KEY assignment in .env"), f.Close())
			}
			tail := strings.TrimSpace(right[end+2:])
			if tail != "" && !strings.HasPrefix(tail, "#") {
				return "", errors.Join(errors.New("invalid YOUTUBE_API_KEY assignment in .env"), f.Close())
			}
			value = right[1 : end+1]
		} else {
			for i, c := range right {
				if c == '#' && (i == 0 || right[i-1] == ' ' || right[i-1] == '\t') {
					right = right[:i]
					break
				}
			}
			value = strings.TrimSpace(right)
		}
	}
	if err := errors.Join(scanner.Err(), f.Close()); err != nil {
		return "", fmt.Errorf("could not read local .env")
	}
	return value, nil
}
