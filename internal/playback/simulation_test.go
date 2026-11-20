package playback

import (
	"testing"
	"time"

	"github.com/SAB1825/music-tui/internal/domain"
)

func TestAdvance(t *testing.T) {
	tracks := []domain.Track{{Title: "First", Duration: 2 * time.Second}, {Title: "Last", Duration: 3 * time.Second}}
	for _, tc := range []struct {
		elapsed  time.Duration
		title    string
		position time.Duration
		playing  bool
	}{
		{time.Second, "First", time.Second, true},
		{2 * time.Second, "Last", 0, true},
		{3 * time.Second, "Last", time.Second, true},
		{10 * time.Second, "Last", 3 * time.Second, false},
	} {
		t.Run(tc.elapsed.String(), func(t *testing.T) {
			s := NewSimulation(tracks)
			s.Toggle()
			s.Advance(tc.elapsed)
			track, _ := s.Track()
			if track.Title != tc.title || s.Position() != tc.position || s.Playing() != tc.playing {
				t.Fatalf("got %q at %v with playing=%t", track.Title, s.Position(), s.Playing())
			}
		})
	}
}

func TestPausedSkipAndVolume(t *testing.T) {
	s := NewSimulation([]domain.Track{{Title: "One"}, {Title: "Two"}})
	s.Skip(-1)
	s.Advance(time.Second)
	track, _ := s.Track()
	if track.Title != "Two" || s.Playing() || s.Position() != 0 {
		t.Fatal("wrapping backward should leave the next track cued and paused")
	}
	s.AdjustVolume(100)
	if s.Volume() != 100 {
		t.Fatal("volume exceeded upper bound")
	}
	s.AdjustVolume(-150)
	if s.Volume() != 0 {
		t.Fatal("volume fell below zero")
	}
	empty := NewSimulation(nil)
	empty.Toggle()
	empty.Skip(1)
	if _, ok := empty.Track(); ok || empty.Playing() {
		t.Fatal("empty queue should stay idle")
	}
}

func TestRepeatAndShuffle(t *testing.T) {
	s := NewSimulation([]domain.Track{{Duration: time.Second}, {Duration: time.Second}})
	s.Play(1)
	s.ToggleRepeat()
	s.Advance(1500 * time.Millisecond)
	if s.Index() != 0 || s.Position() != 500*time.Millisecond || !s.Repeat() {
		t.Fatal("repeat should wrap to the start")
	}
	s.ToggleRepeat()
	s.ToggleShuffle()
	s.Play(1)
	s.Advance(time.Second)
	if s.Index() != 0 || !s.Shuffle() || !s.Playing() {
		t.Fatal("shuffle should advance to a different track")
	}
}
