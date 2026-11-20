package mock

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/SAB1825/music-tui/internal/domain"
)

// Catalog returns the built-in, fictional music library.
func Catalog() []domain.Track {
	tracks := []domain.Track{
		{Title: "Copper Horizon", Artist: "Velvet Transit", Album: "After the Last Train", Duration: 211 * time.Second},
		{Title: "Still Air", Artist: "Hotel Meridian", Album: "Low Season", Duration: 248 * time.Second},
		{Title: "Soft Signal", Artist: "Kite Assembly", Album: "Signal Studies", Duration: 194 * time.Second},
		{Title: "Paper Moons", Artist: "Juniper Lines", Album: "Night Letters", Duration: 225 * time.Second},
		{Title: "Streetlights in Rain", Artist: "Velvet Transit", Album: "After the Last Train", Duration: 263 * time.Second},
		{Title: "A Room in Amber", Artist: "Hotel Meridian", Album: "Low Season", Duration: 187 * time.Second},
		{Title: "Daybreak Relay", Artist: "Kite Assembly", Album: "Signal Studies", Duration: 309 * time.Second},
		{Title: "Last Light", Artist: "Juniper Lines", Album: "Night Letters", Duration: 202 * time.Second},
	}
	for i := range tracks {
		tracks[i].Provider, tracks[i].ID = "mock", strconv.Itoa(i+1)
	}
	return tracks
}

// Provider exposes the same offline catalog through the search capability.
type Provider struct{}

func (Provider) Search(ctx context.Context, query, _ string) (domain.TrackPage, error) {
	if err := ctx.Err(); err != nil {
		return domain.TrackPage{}, err
	}
	query = strings.ToLower(strings.TrimSpace(query))
	var results domain.TrackPage
	for _, track := range Catalog() {
		if strings.Contains(strings.ToLower(track.Title+" "+track.Artist+" "+track.Album), query) {
			results.Tracks = append(results.Tracks, track)
		}
	}
	return results, nil
}
