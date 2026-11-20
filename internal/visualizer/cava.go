// Package visualizer manages the raw CAVA audio spectrum backend.
package visualizer

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/SAB1825/music-tui/internal/config"
)

// Event carries an owned spectrum frame or the final backend error.
type Event struct {
	Frame []float64
	Err   error
}

// Start opens a single latest-event stream. Cancel ctx to stop CAVA and close
// the stream after its process and configuration have been cleaned up.
func Start(ctx context.Context, cfg config.Visualizer) <-chan Event {
	return start(ctx, cfg, "cava")
}

func start(ctx context.Context, cfg config.Visualizer, executable string) <-chan Event {
	events := make(chan Event, 1)
	go func() {
		defer close(events)
		publish := func(event Event) {
			select {
			case events <- event:
				return
			default:
			}
			select {
			case <-events:
			default:
			}
			events <- event
		}
		if err := run(ctx, cfg, executable, func(frame []float64) { publish(Event{Frame: frame}) }); err != nil && ctx.Err() == nil {
			publish(Event{Err: err})
		}
	}()
	return events
}

func run(ctx context.Context, cfg config.Visualizer, executable string, emit func([]float64)) (result error) {
	if err := cfg.Validate(); err != nil {
		return err
	}
	if !cfg.Enabled {
		return nil
	}
	binary, err := exec.LookPath(executable)
	if err != nil {
		return fmt.Errorf("find CAVA executable: %w", err)
	}
	f, err := os.CreateTemp("", "music-tui-cava-*.ini")
	if err != nil {
		return fmt.Errorf("create CAVA configuration: %w", err)
	}
	defer func() {
		if err := os.Remove(f.Name()); err != nil {
			result = errors.Join(result, fmt.Errorf("remove CAVA configuration: %w", err))
		}
	}()
	_, err = f.WriteString(cavaConfig(cfg))
	if err := errors.Join(err, f.Close()); err != nil {
		return fmt.Errorf("write CAVA configuration: %w", err)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, "-p", f.Name())
	cmd.WaitDelay = 2 * time.Second
	var stderr diagnostic
	cmd.Stderr = &stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("open CAVA stdout: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start CAVA: %w", err)
	}
	parseErr := readFrames(stdout, cfg.Bars, emit)
	if parseErr != nil {
		cancel()
	}
	waitErr := cmd.Wait()
	if parseErr != nil {
		return parseErr
	}
	if waitErr == nil {
		waitErr = errors.New("raw output closed")
	}
	if details := strings.TrimSpace(stderr.String()); details != "" {
		return fmt.Errorf("CAVA: %w (%s)", waitErr, details)
	}
	return fmt.Errorf("CAVA: %w", waitErr)
}

func cavaConfig(cfg config.Visualizer) string {
	return fmt.Sprintf(`[general]
bars = %d
framerate = %d
sleep_timer = 0

[input]
method = pulse
source = %s

[output]
method = raw
channels = mono
raw_target = /dev/stdout
data_format = ascii
ascii_max_range = %d
bar_delimiter = 59
frame_delimiter = 10
`, cfg.Bars, cfg.Framerate, cfg.Source, fullScale)
}

type diagnostic struct{ strings.Builder }

func (d *diagnostic) Write(p []byte) (int, error) {
	room := max(0, 1024-d.Len())
	d.Builder.Write(p[:min(room, len(p))])
	return len(p), nil
}
