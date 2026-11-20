package domain

import "time"

// Track describes one catalog entry.
type Track struct {
	Title, Artist, Album string
	Duration             time.Duration
}
