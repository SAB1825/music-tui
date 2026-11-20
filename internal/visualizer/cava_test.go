package visualizer

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	"github.com/SAB1825/music-tui/internal/config"
)

func TestReadFrames(t *testing.T) {
	var frames [][]float64
	data := iotest.OneByteReader(strings.NewReader("250;1000;\n250;500\n750;900;"))
	if err := readFrames(data, 2, func(frame []float64) { frames = append(frames, frame) }); err != nil {
		t.Fatal(err)
	}
	if len(frames) != 2 || !slices.Equal(frames[0], []float64{.25, 1}) || !slices.Equal(frames[1], []float64{.25, .925}) {
		t.Fatalf("unexpected complete frames and decay: %v", frames)
	}
	for _, line := range []string{"1", "1;2;3", "1;xyz", "1;-3", "1;1001", "1;;2", ""} {
		if _, err := parseFrame([]byte(line), 2); err == nil {
			t.Errorf("accepted malformed frame %q", line)
		}
	}
	if err := readFrames(io.LimitReader(strings.NewReader("100;200"), 5), 2, func([]float64) {
		t.Error("published a partial frame")
	}); err != nil {
		t.Fatal(err)
	}
}

func TestMissingExecutable(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	events := Start(t.Context(), config.Defaults())
	event := <-events
	if !errors.Is(event.Err, exec.ErrNotFound) {
		t.Fatalf("expected missing executable, got %v", event.Err)
	}
	if _, open := <-events; open {
		t.Fatal("backend did not close its output after failure")
	}
}

func TestProcessShutdown(t *testing.T) {
	shell, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("shell unavailable")
	}
	tmp := t.TempDir()
	t.Setenv("TMPDIR", tmp)
	t.Setenv("CAVA_TEST_PID", filepath.Join(tmp, "pid"))
	executable := filepath.Join(tmp, "fake-cava")
	script := "#!" + shell + "\nprintf '%s' \"$$\" > \"$CAVA_TEST_PID\"\nprintf '1000;0;0;0;0;0;0;0;\\n'\nexec sleep 30\n"
	if err := os.WriteFile(executable, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	cfg := config.Defaults()
	cfg.Bars = 8
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	events := start(ctx, cfg, executable)
	event := <-events
	if event.Err != nil || len(event.Frame) != 8 || event.Frame[0] != 1 {
		t.Fatalf("unexpected process event: %+v", event)
	}
	pidText, err := os.ReadFile(filepath.Join(tmp, "pid"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := strconv.Atoi(string(pidText)); err != nil {
		t.Fatal(err)
	}
	cancel()
	select {
	case _, open := <-events:
		if open {
			t.Fatal("unexpected frame after cancellation")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("backend failed to shut down")
	}
	if _, err := os.Stat(filepath.Join("/proc", string(pidText))); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("subprocess still exists: %v", err)
	}
	if configs, err := filepath.Glob(filepath.Join(tmp, "*.ini")); err != nil || len(configs) != 0 {
		t.Fatalf("temporary configs left over: %v, %v", configs, err)
	}
}

func TestIdleFramesAreCoalesced(t *testing.T) {
	events := 0
	err := readFrames(strings.NewReader(strings.Repeat("0;0;\n", 30)), 2, func([]float64) { events++ })
	if err != nil || events != 1 {
		t.Fatalf("expected one idle transition, got %d events and %v", events, err)
	}
}
