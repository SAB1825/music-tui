// Package providers defines the catalog capabilities available to the application.
package providers

import (
	"context"
	"slices"

	"github.com/SAB1825/music-tui/internal/domain"
)

type Catalog interface {
	Search(context.Context, string, string) (domain.TrackPage, error)
}

type Metadata interface {
	Videos(context.Context, []string) (map[string]domain.Track, error)
}

type Playlists interface {
	Playlist(context.Context, string) (domain.Playlist, error)
	PlaylistTracks(context.Context, string, string) (domain.TrackPage, error)
}

type Entry struct {
	ID, Name, Unavailable string
	Catalog               Catalog
}

type Registry struct{ entries []Entry }

func (r *Registry) Register(entry Entry) {
	i := slices.IndexFunc(r.entries, func(existing Entry) bool { return existing.ID == entry.ID })
	if i >= 0 {
		r.entries[i] = entry
		return
	}
	r.entries = append(r.entries, entry)
}

func (r Registry) Entries() []Entry { return slices.Clone(r.entries) }

func (r Registry) Resolve(id string) (Entry, bool) {
	for _, entry := range r.entries {
		if entry.ID == id {
			return entry, true
		}
	}
	return Entry{}, false
}

func (r Registry) Select(id string) Entry {
	if id == "auto" || id == "" {
		id = "youtube"
	}
	if entry, ok := r.Resolve(id); ok && entry.Catalog != nil {
		return entry
	}
	entry, _ := r.Resolve("mock")
	return entry
}
