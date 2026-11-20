package playback

import (
	"slices"
	"time"

	"github.com/SAB1825/music-tui/internal/domain"
)

// Simulation tracks an in-memory queue and advances only when given elapsed time.
type Simulation struct {
	queue    []domain.Track
	index    int
	position time.Duration
	volume   int
	playing  bool
}

// NewSimulation cues the first track of a private queue.
func NewSimulation(tracks []domain.Track) Simulation {
	return Simulation{queue: slices.Clone(tracks), volume: 70}
}

// Track returns the currently cued track, if any.
func (s Simulation) Track() (domain.Track, bool) {
	if len(s.queue) == 0 {
		return domain.Track{}, false
	}
	return s.queue[s.index], true
}

func (s Simulation) Position() time.Duration { return s.position }
func (s Simulation) Volume() int             { return s.volume }
func (s Simulation) Playing() bool           { return s.playing }

// Play starts the selected catalog index from the beginning.
func (s *Simulation) Play(index int) {
	if index >= 0 && index < len(s.queue) {
		s.index, s.position, s.playing = index, 0, true
	}
}

// Toggle pauses or resumes the cued track.
func (s *Simulation) Toggle() {
	if track, ok := s.Track(); ok {
		if s.position >= track.Duration {
			s.position = 0
		}
		s.playing = !s.playing
	}
}

// Skip wraps through the queue while preserving the play/pause state.
func (s *Simulation) Skip(step int) {
	if count := len(s.queue); count > 0 {
		s.index = ((s.index+step%count)%count + count) % count
		s.position = 0
	}
}

// AdjustVolume changes the simulated volume in percentage points.
func (s *Simulation) AdjustVolume(delta int) {
	s.volume = min(100, max(0, s.volume+delta))
}

// Advance carries elapsed time across tracks and stops at the end of the queue.
func (s *Simulation) Advance(elapsed time.Duration) {
	for s.playing && elapsed > 0 {
		remaining := s.queue[s.index].Duration - s.position
		if elapsed < remaining {
			s.position += elapsed
			return
		}
		elapsed -= remaining
		if s.index == len(s.queue)-1 {
			s.position = s.queue[s.index].Duration
			s.playing = false
			return
		}
		s.index++
		s.position = 0
	}
}
