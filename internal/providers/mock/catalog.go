package mock

import (
	"time"

	"github.com/SAB1825/music-tui/internal/domain"
)

// Catalog returns the built-in, fictional music library.
func Catalog() []domain.Track {
	return []domain.Track{
		{Title: "Copper Horizon", Artist: "Velvet Transit", Album: "After the Last Train", Duration: 211 * time.Second},
		{Title: "Still Air", Artist: "Hotel Meridian", Album: "Low Season", Duration: 248 * time.Second},
		{Title: "Soft Signal", Artist: "Kite Assembly", Album: "Signal Studies", Duration: 194 * time.Second},
		{Title: "Paper Moons", Artist: "Juniper Lines", Album: "Night Letters", Duration: 225 * time.Second},
		{Title: "Streetlights in Rain", Artist: "Velvet Transit", Album: "After the Last Train", Duration: 263 * time.Second},
		{Title: "A Room in Amber", Artist: "Hotel Meridian", Album: "Low Season", Duration: 187 * time.Second},
		{Title: "Daybreak Relay", Artist: "Kite Assembly", Album: "Signal Studies", Duration: 309 * time.Second},
		{Title: "Last Light", Artist: "Juniper Lines", Album: "Night Letters", Duration: 202 * time.Second},
	}
}
