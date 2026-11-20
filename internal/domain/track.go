package domain

import "time"

// Track describes one catalog entry. Provider and ID together identify it.
type Track struct {
	Provider, ID                  string
	Title, Artist, Creator, Album string
	Duration                      time.Duration
	ArtworkURL, ExternalURL       string
	Unavailable                   bool
}

type Playlist struct {
	Provider, ID, Title, Description string
	ArtworkURL, ExternalURL          string
	TrackCount                       *int
}

type TrackPage struct {
	Tracks    []Track
	NextToken string
}
