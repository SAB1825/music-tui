package playback

import (
	"math/rand/v2"
	"slices"
	"time"

	"github.com/SAB1825/music-tui/internal/domain"
)

// Simulation tracks an in-memory queue and advances only when given elapsed time.
type Simulation struct {
	queue           []domain.Track
	index           int
	position        time.Duration
	volume          int
	playing         bool
	shuffle, repeat bool
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

// Index returns the current queue index, or -1 for an empty queue.
func (s Simulation) Index() int {
	if len(s.queue) == 0 {
		return -1
	}
	return s.index
}

func (s Simulation) Shuffle() bool   { return s.shuffle }
func (s Simulation) Repeat() bool    { return s.repeat }
func (s *Simulation) ToggleShuffle() { s.shuffle = !s.shuffle }
func (s *Simulation) ToggleRepeat()  { s.repeat = !s.repeat }

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

// Seek changes position within the currently cued track.
func (s *Simulation) Seek(position time.Duration) {
	if track, ok := s.Track(); ok {
		s.position = min(max(0, track.Duration), max(0, position))
	}
}

// AdjustVolume changes the simulated volume in percentage points.
func (s *Simulation) AdjustVolume(delta int) {
	s.volume = min(100, max(0, s.volume+delta))
}

// Advance carries elapsed time across tracks and stops at the end of the queue.
func (s *Simulation) Advance(elapsed time.Duration) {
	emptyTracks := 0
	for s.playing && elapsed > 0 {
		remaining := s.queue[s.index].Duration - s.position
		if elapsed < remaining {
			s.position += elapsed
			return
		}
		if remaining > 0 {
			elapsed -= remaining
			emptyTracks = 0
		} else {
			emptyTracks++
			if emptyTracks == len(s.queue) {
				s.playing = false
				return
			}
		}
		if s.index == len(s.queue)-1 && !s.repeat && (!s.shuffle || len(s.queue) < 2) {
			s.position = s.queue[s.index].Duration
			s.playing = false
			return
		}
		if s.shuffle && len(s.queue) > 1 && s.queue[s.index].Duration > 0 {
			s.Skip(1 + rand.IntN(len(s.queue)-1))
		} else {
			s.Skip(1)
		}
	}
}
