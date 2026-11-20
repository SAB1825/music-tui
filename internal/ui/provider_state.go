package ui

import (
	"context"
	"errors"
	"log/slog"
	"maps"
	"os"
	"os/exec"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"

	tea "charm.land/bubbletea/v2"
	"github.com/SAB1825/music-tui/internal/artwork"
	"github.com/SAB1825/music-tui/internal/domain"
	"github.com/SAB1825/music-tui/internal/providers"
)

const resultLimit = 200

type providerState struct {
	registry            providers.Registry
	active              providers.Entry
	ctx                 context.Context
	cancel              context.CancelFunc
	serial              uint64
	query, next         string
	tracks              []domain.Track
	playlist            domain.Playlist
	loading, paginating bool
	playlistRequest     bool
	browserBusy         bool
	lastNotice          string
	err                 string
	thumbs              artwork.Thumbnails
	images              map[string]string
	imageBusy           bool
	imageScope          uint64
	imageCancel         context.CancelFunc
}

type catalogResult struct {
	serial   uint64
	playlist domain.Playlist
	page     domain.TrackPage
	err      error
}

type thumbnailResult struct {
	scope  uint64
	images map[string]string
}
type browserResult struct {
	serial uint64
	failed bool
}

// WithProviders configures catalog discovery independently of the mock playback queue.
func WithProviders(initial tea.Model, ctx context.Context, registry providers.Registry, preferred string) tea.Model {
	m := initial.(model)
	m.online = providerState{ctx: ctx, registry: registry, active: registry.Select(preferred), thumbs: artwork.NewThumbnails(nil)}
	m.updateSearchPlaceholder()
	return m
}

func (m model) remote() bool { return m.online.active.ID != "" && m.online.active.ID != "mock" }
func (m model) providerName() string {
	if m.online.active.Name == "" {
		return "Mock"
	}
	return m.online.active.Name
}
func (m *model) updateSearchPlaceholder() {
	m.search.Placeholder = "Search your library"
	if m.remote() {
		m.search.Placeholder = "Search " + m.providerName() + " (Enter)"
	}
}

func (m *model) selectProvider(entry providers.Entry) {
	if entry.Catalog == nil || entry.ID == m.online.active.ID {
		return
	}
	m.clearCatalog()
	m.online.active = entry
	m.online.lastNotice = ""
	m.online.images = nil
	m.updateSearchPlaceholder()
	m.open(home)
	if m.remote() && m.player.Playing() {
		m.player.Toggle()
		m.clock++
	}
}

func (m *model) clearCatalog() {
	if m.online.cancel != nil {
		m.online.cancel()
		m.online.cancel = nil
	}
	m.online.serial++
	if m.online.imageCancel != nil {
		m.online.imageCancel()
		m.online.imageCancel = nil
	}
	m.online.imageScope++
	m.online.imageBusy = false
	m.online.images = nil
	m.online.query, m.online.next, m.online.err = "", "", ""
	m.online.tracks = nil
	m.online.playlist = domain.Playlist{}
	m.online.loading, m.online.paginating, m.online.playlistRequest = false, false, false
	m.selected, m.offset = 0, 0
}

func (m *model) startCatalog(query string, playlist, paginate bool) tea.Cmd {
	query = strings.TrimSpace(query)
	if query == "" || !m.remote() {
		return nil
	}
	s := &m.online
	if paginate {
		if s.loading || s.next == "" || len(s.tracks) >= resultLimit {
			return nil
		}
	} else {
		if s.query == query && s.playlistRequest == playlist && (s.loading || s.err == "") {
			return nil
		}
		m.clearCatalog()
	}
	ctx, cancel := context.WithCancel(s.ctx)
	s.cancel = cancel
	s.serial++
	s.query, s.loading, s.paginating, s.err = query, true, paginate, ""
	s.playlistRequest = playlist
	serial, token, entry, previous := s.serial, s.next, s.active, s.playlist
	return func() tea.Msg {
		defer cancel()
		result := catalogResult{serial: serial, playlist: previous}
		if playlist {
			source, ok := entry.Catalog.(providers.Playlists)
			if !ok {
				result.err = &providers.Error{Kind: providers.Unavailable, Operation: "playlists"}
				return result
			}
			if !paginate {
				result.playlist, result.err = source.Playlist(ctx, query)
			}
			if result.err == nil {
				result.page, result.err = source.PlaylistTracks(ctx, result.playlist.ID, token)
			}
		} else {
			result.page, result.err = entry.Catalog.Search(ctx, query, token)
		}
		return result
	}
}

func (m *model) applyCatalog(result catalogResult) {
	s := &m.online
	if result.serial != s.serial || !s.loading {
		return
	}
	s.loading = false
	s.cancel = nil
	if result.err != nil {
		s.err = userCatalogError(result.err)
		var failure *providers.Error
		if errors.As(result.err, &failure) {
			slog.Debug("catalog request failed", "provider", s.active.ID, "category", failure.Kind, "http_status", failure.Status)
			switch failure.Kind {
			case providers.Credentials, providers.Disabled, providers.Quota:
				if m.dialog.kind == dialogNone && s.lastNotice != s.err {
					s.lastNotice = s.err
					m.showDialog(dialogCatalogError)
				}
			}
		}
		return
	}
	s.playlist = result.playlist
	tracks := slices.Clone(s.tracks)
	seen := make(map[string]bool, len(tracks))
	for _, t := range tracks {
		seen[t.ID] = true
	}
	for _, t := range result.page.Tracks {
		if len(tracks) >= resultLimit {
			break
		}
		if s.playlist.ID == "" && seen[t.ID] {
			continue
		}
		seen[t.ID] = true
		tracks = append(tracks, t)
	}
	s.tracks = tracks
	if s.next == result.page.NextToken || len(tracks) >= resultLimit {
		s.next = ""
	} else {
		s.next = result.page.NextToken
	}
}

func userCatalogError(err error) string {
	var failure *providers.Error
	if errors.As(err, &failure) {
		switch failure.Kind {
		case providers.Credentials:
			return "The API key was rejected. Check its YouTube API access and restrictions."
		case providers.Disabled:
			return "YouTube Data API v3 is disabled for this key’s project."
		case providers.Quota:
			return "YouTube API quota exceeded. Try again after the quota resets."
		case providers.RateLimit:
			return "YouTube is rate limiting requests. Wait before trying again."
		case providers.Timeout:
			return "The request timed out. Try again."
		case providers.InvalidInput:
			return "YouTube rejected this request. Check the query or public playlist link."
		case providers.Unavailable:
			return "This resource is unavailable or is not a public playlist."
		case providers.InvalidResponse:
			return "YouTube returned an unreadable response."
		}
	}
	return "Could not reach the catalog. Check your connection and try again."
}

func (m *model) nextCatalogPage() tea.Cmd {
	return m.startCatalog(m.online.query, m.online.playlist.ID != "", true)
}

func (m *model) openCatalogURL(src string) tea.Cmd {
	if src == "" || m.online.browserBusy {
		return nil
	}
	m.online.browserBusy = true
	ctx, serial := m.online.ctx, m.online.serial
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if !strings.HasPrefix(src, "https://www.youtube.com/watch?") && !strings.HasPrefix(src, "https://www.youtube.com/playlist?") {
			return browserResult{serial, true}
		}
		name := "xdg-open"
		if runtime.GOOS == "darwin" {
			name = "open"
		}
		cmd := exec.CommandContext(ctx, name, src)
		for _, value := range os.Environ() {
			if !strings.HasPrefix(value, "YOUTUBE_API_KEY=") {
				cmd.Env = append(cmd.Env, value)
			}
		}
		return browserResult{serial, cmd.Run() != nil}
	}
}

func (m model) selectedURL() string {
	if m.selected < len(m.online.tracks) {
		return m.online.tracks[m.selected].ExternalURL
	}
	return ""
}

func (m *model) startThumbnails() tea.Cmd {
	s := &m.online
	if !m.remote() || s.imageBusy || len(s.images) >= 64 || s.ctx == nil || m.bodyHeight() < 5 || m.dialog.kind != dialogNone || (m.layout().sidebar == 0 && m.focus == sidebarFocus) {
		return nil
	}
	start := m.trackStart(len(s.tracks), m.bodyHeight()-4)
	var urls []string
	add := func(src string) {
		if _, exists := s.images[src]; src != "" && !exists && !slices.Contains(urls, src) && len(urls) < min(4, 64-len(s.images)) {
			urls = append(urls, src)
		}
	}
	add(s.playlist.ArtworkURL)
	for _, t := range s.tracks[start:min(len(s.tracks), start+max(0, m.bodyHeight()-4))] {
		add(t.ArtworkURL)
	}
	if len(urls) == 0 {
		return nil
	}
	s.imageBusy = true
	ctx, cancel := context.WithCancel(s.ctx)
	s.imageCancel = cancel
	scope, loader := s.imageScope, s.thumbs
	return func() tea.Msg {
		defer cancel()
		images := make(map[string]string)
		type sample struct{ src, rendered string }
		done := make(chan sample, len(urls))
		for _, src := range urls {
			go func() {
				pixels, err := loader.Fetch(ctx, src)
				rendered := ""
				if err == nil {
					rendered = paintThumbnail(pixels)
				}
				done <- sample{src, rendered}
			}()
		}
		for range urls {
			item := <-done
			images[item.src] = item.rendered
		}
		return thumbnailResult{scope, images}
	}
}

func (m *model) applyThumbnails(result thumbnailResult) {
	if result.scope != m.online.imageScope {
		return
	}
	m.online.imageCancel = nil
	m.online.imageBusy = false
	if !m.remote() {
		return
	}
	dst := maps.Clone(m.online.images)
	if dst == nil {
		dst = make(map[string]string)
	}
	maps.Copy(dst, result.images)
	m.online.images = dst
}

func paintThumbnail(img artwork.Thumbnail) string {
	tint := func(rgb [3]uint8) string {
		return strconv.Itoa(int(rgb[0])) + ";" + strconv.Itoa(int(rgb[1])) + ";" + strconv.Itoa(int(rgb[2]))
	}
	var out strings.Builder
	for x := range 6 {
		out.WriteString("\x1b[38;2;" + tint(img[x]) + ";48;2;" + tint(img[6+x]) + "m▀")
	}
	out.WriteString("\x1b[0m")
	return out.String()
}

func catalogText(value string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, value)
}
