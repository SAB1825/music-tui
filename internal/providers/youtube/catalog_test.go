package youtube

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/SAB1825/music-tui/internal/providers"
)

type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func fixture(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	return New("fixture-key", "https://example.invalid/youtube/v3", roundTrip(func(r *http.Request) (*http.Response, error) {
		if r.Header.Get("X-Goog-Api-Key") != "fixture-key" || r.URL.Query().Has("key") {
			t.Error("key was not confined to the API header")
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w.Result(), nil
	}), 2*time.Second, 20)
}

func writeFixture(t *testing.T, w http.ResponseWriter, body string) {
	t.Helper()
	if _, err := fmt.Fprint(w, body); err != nil {
		t.Fatal(err)
	}
}

func TestSearchPreservesOrderAndBatches(t *testing.T) {
	var requests []string
	client := fixture(t, func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.URL.Path)
		q := r.URL.Query()
		if strings.HasSuffix(r.URL.Path, "/search") {
			if q.Get("q") != "Björk 音楽" || q.Get("videoCategoryId") != "10" || q.Get("type") != "video" ||
				q.Get("part") != "snippet" || q.Get("pageToken") != "cursor" || q.Get("maxResults") != "20" {
				t.Errorf("wrong search query: %v", q)
			}
			writeFixture(t, w, `{"nextPageToken":"next","items":[{"id":{"videoId":"b"},"snippet":{"title":"B &amp; C"}},{"id":{"videoId":"a"}},{"id":{"videoId":"a"}},{"id":{"videoId":"gone"},"snippet":{"title":"Missing"}}]}`)
		} else {
			if q.Get("id") != "b,a,gone" || q.Get("part") != "snippet,contentDetails,status" {
				t.Errorf("bad batch: %v", q)
			}
			writeFixture(t, w, `{"items":[{"id":"a","snippet":{"title":"Second"}},{"id":"b","snippet":{"channelTitle":"Creator","thumbnails":{"medium":{"url":"https://i.ytimg.com/vi/b/mqdefault.jpg"}}},"contentDetails":{"duration":"PT3M12S"}}]}`)
		}
	})
	page, err := client.Search(t.Context(), "Björk 音楽", "cursor")
	if err != nil || len(requests) != 2 || len(page.Tracks) != 3 || page.NextToken != "next" {
		t.Fatalf("result=%+v err=%v requests=%v", page, err, requests)
	}
	first := page.Tracks[0]
	if first.ID != "b" || first.Title != "B & C" || first.Creator != "Creator" || first.Artist != "" || first.Album != "" ||
		first.Duration != 192*time.Second || first.ArtworkURL == "" || first.ExternalURL == "" || page.Tracks[1].ID != "a" ||
		!page.Tracks[2].Unavailable || page.Tracks[2].ExternalURL != "" {
		t.Fatalf("bad mapping: %+v", page)
	}
}

func TestEmptySearchAndDuration(t *testing.T) {
	client := fixture(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/search") {
			t.Error("empty results should skip metadata")
		}
		writeFixture(t, w, `{"items":[]}`)
	})
	page, err := client.Search(t.Context(), "empty", "")
	if err != nil || len(page.Tracks) != 0 {
		t.Fatalf("empty: %+v %v", page, err)
	}
	for _, tc := range []struct {
		in   string
		want time.Duration
	}{
		{"PT5S", 5 * time.Second}, {"PT1H02M03.25S", time.Hour + 2*time.Minute + 3250*time.Millisecond},
		{"P1DT2H", 26 * time.Hour}, {"", 0}, {"P2M", 0}, {"PT9999999999999999S", 0},
	} {
		if got := parseDuration(tc.in); got != tc.want {
			t.Errorf("%s: got %v want %v", tc.in, got, tc.want)
		}
	}
}

func TestPlaylist(t *testing.T) {
	client := fixture(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		switch r.URL.Path {
		case "/youtube/v3/playlists":
			writeFixture(t, w, `{"items":[{"id":"PLTest","snippet":{"title":"Public set","description":"A playlist"},"contentDetails":{"itemCount":2},"status":{"privacyStatus":"public"}}]}`)
		case "/youtube/v3/playlistItems":
			if q.Get("playlistId") != "PLTest" {
				t.Errorf("wrong playlist: %v", q)
			}
			if q.Get("pageToken") == "last" {
				writeFixture(t, w, `{"items":[]}`)
				return
			}
			writeFixture(t, w, `{"nextPageToken":"last","items":[{"contentDetails":{"videoId":"one"}},{"snippet":{"title":"Removed"}}]}`)
		case "/youtube/v3/videos":
			writeFixture(t, w, `{"items":[{"id":"one","snippet":{"title":"First"},"contentDetails":{"duration":"PT1M"}}]}`)
		}
	})
	playlist, err := client.Playlist(t.Context(), "https://www.youtube.com/playlist?list=PLTest")
	if err != nil || playlist.Title != "Public set" || playlist.TrackCount == nil || *playlist.TrackCount != 2 {
		t.Fatalf("playlist=%+v %v", playlist, err)
	}
	page, err := client.PlaylistTracks(t.Context(), playlist.ID, "")
	if err != nil || page.NextToken != "last" || len(page.Tracks) != 2 || page.Tracks[0].Duration != time.Minute || !page.Tracks[1].Unavailable {
		t.Fatalf("items=%+v %v", page, err)
	}
	last, err := client.PlaylistTracks(t.Context(), playlist.ID, page.NextToken)
	if err != nil || len(last.Tracks) != 0 {
		t.Fatalf("empty page=%+v %v", last, err)
	}
	if _, err := client.Playlist(t.Context(), "https://unrelated.invalid/?list=PLTest"); err == nil {
		t.Fatal("accepted unrelated URL")
	}
}

func TestHTTPFailuresAreSanitized(t *testing.T) {
	for _, tc := range []struct {
		code  int
		body  string
		kind  providers.ErrorKind
		calls int
		wait  string
	}{
		{400, `{"error":{"message":"fixture-key","details":[{"reason":"API_KEY_INVALID"}]}}`, providers.Credentials, 1, ""},
		{403, `{"error":{"errors":[{"reason":"quotaExceeded"}]}}`, providers.Quota, 1, ""},
		{403, `{"error":{"details":[{"reason":"SERVICE_DISABLED"}]}}`, providers.Disabled, 1, ""},
		{429, `{}`, providers.RateLimit, 1, "30"},
		{503, `{}`, providers.Service, 3, ""},
		{200, `junk fixture-key`, providers.InvalidResponse, 1, ""},
		{404, `{}`, providers.Unavailable, 1, ""},
	} {
		t.Run(string(tc.kind), func(t *testing.T) {
			count := 0
			client := fixture(t, func(w http.ResponseWriter, _ *http.Request) {
				count++
				w.Header().Set("Retry-After", tc.wait)
				w.WriteHeader(tc.code)
				writeFixture(t, w, tc.body)
			})
			_, err := client.Search(t.Context(), "test", "")
			var failure *providers.Error
			if !errors.As(err, &failure) || failure.Kind != tc.kind || strings.Contains(err.Error(), "fixture-key") || count != tc.calls {
				t.Fatalf("wrong failure or retry count: %v (%d)", err, count)
			}
		})
	}
}

func TestCancellationTimeoutAndTransientRecovery(t *testing.T) {
	var calls atomic.Int32
	client := fixture(t, func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(429)
			return
		}
		writeFixture(t, w, `{"items":[]}`)
	})
	if _, err := client.Search(t.Context(), "retry", ""); err != nil || calls.Load() != 2 {
		t.Fatalf("retry failed: %v", err)
	}
	client = New("fixture-key", "https://example.invalid", roundTrip(func(r *http.Request) (*http.Response, error) {
		<-r.Context().Done()
		return nil, fmt.Errorf("private fixture-key details: %w", r.Context().Err())
	}), 20*time.Millisecond, 2)
	_, err := client.Search(t.Context(), "timeout", "")
	var failure *providers.Error
	if !errors.As(err, &failure) || failure.Kind != providers.Timeout {
		t.Fatalf("timeout: %v", err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := client.Search(ctx, "cancel", ""); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation: %v", err)
	}
}
