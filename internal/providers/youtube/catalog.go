package youtube

import (
	"context"
	"html"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/SAB1825/music-tui/internal/domain"
	"github.com/SAB1825/music-tui/internal/providers"
)

type list[T any] struct {
	Items         []T
	NextPageToken string
}

type thumbnail struct{ URL string }
type snippet struct {
	Title, ChannelTitle, VideoOwnerChannelTitle, Description string
	Thumbnails                                               map[string]thumbnail
	ResourceID                                               struct{ VideoID string }
}

type video struct {
	ID             string
	Snippet        snippet
	ContentDetails struct {
		Duration          string
		RegionRestriction struct{ Allowed *[]string }
	}
	Status struct{ PrivacyStatus, UploadStatus string }
}

func (c *Client) Search(ctx context.Context, query, token string) (domain.TrackPage, error) {
	q := url.Values{"part": {"snippet"}, "type": {"video"}, "videoCategoryId": {"10"}, "q": {strings.TrimSpace(query)}}
	if q.Get("q") == "" {
		return domain.TrackPage{}, nil
	}
	c.pageParams(q, token)
	var response list[struct {
		ID      struct{ VideoID string }
		Snippet snippet
	}]
	if err := c.get(ctx, "search", q, &response); err != nil {
		return domain.TrackPage{}, err
	}
	page := domain.TrackPage{NextToken: response.NextPageToken}
	seen := make(map[string]bool)
	for _, item := range response.Items {
		id := item.ID.VideoID
		if !validID(id) || seen[id] {
			continue
		}
		seen[id] = true
		page.Tracks = append(page.Tracks, fromSnippet(id, item.Snippet))
	}
	return c.enrich(ctx, page)
}

func (c *Client) Videos(ctx context.Context, ids []string) (map[string]domain.Track, error) {
	result := make(map[string]domain.Track)
	var unique []string
	seen := make(map[string]bool)
	for _, id := range ids {
		if validID(id) && !seen[id] {
			unique = append(unique, id)
			seen[id] = true
		}
	}
	for i := 0; i < len(unique); i += 50 {
		q := url.Values{"part": {"snippet,contentDetails,status"}, "id": {strings.Join(unique[i:min(i+50, len(unique))], ",")}}
		var response list[video]
		if err := c.get(ctx, "videos", q, &response); err != nil {
			return nil, err
		}
		for _, item := range response.Items {
			if !seen[item.ID] {
				continue
			}
			t := fromSnippet(item.ID, item.Snippet)
			t.Duration = parseDuration(item.ContentDetails.Duration)
			regions := item.ContentDetails.RegionRestriction.Allowed
			t.Unavailable = item.Status.PrivacyStatus == "private" ||
				(item.Status.UploadStatus != "" && item.Status.UploadStatus != "processed") ||
				(regions != nil && len(*regions) == 0)
			if t.Unavailable {
				t.ExternalURL = ""
			}
			result[t.ID] = t
		}
	}
	return result, nil
}

func (c *Client) enrich(ctx context.Context, page domain.TrackPage) (domain.TrackPage, error) {
	ids := make([]string, 0, len(page.Tracks))
	for _, t := range page.Tracks {
		ids = append(ids, t.ID)
	}
	details, err := c.Videos(ctx, ids)
	if err != nil {
		return domain.TrackPage{}, err
	}
	for i, t := range page.Tracks {
		if detail, ok := details[t.ID]; ok {
			if detail.Title == "" {
				detail.Title = t.Title
			}
			if detail.Creator == "" {
				detail.Creator = t.Creator
			}
			if detail.ArtworkURL == "" {
				detail.ArtworkURL = t.ArtworkURL
			}
			page.Tracks[i] = detail
		} else {
			t.Unavailable, t.ExternalURL = true, ""
			if t.Title == "" {
				t.Title = "Unavailable video"
			}
			page.Tracks[i] = t
		}
	}
	return page, nil
}

func (c *Client) pageParams(q url.Values, token string) {
	q.Set("maxResults", strconv.Itoa(c.limit))
	if token != "" {
		q.Set("pageToken", token)
	}
}

func fromSnippet(id string, s snippet) domain.Track {
	creator := s.VideoOwnerChannelTitle
	if creator == "" {
		creator = s.ChannelTitle
	}
	return domain.Track{Provider: "youtube", ID: id, Title: clean(s.Title), Creator: clean(creator), ArtworkURL: coverURL(s),
		ExternalURL: "https://www.youtube.com/watch?" + url.Values{"v": {id}}.Encode()}
}

func clean(value string) string {
	return strings.Join(strings.Fields(strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, html.UnescapeString(value))), " ")
}

func coverURL(s snippet) string {
	for _, size := range []string{"medium", "default", "high", "standard", "maxres"} {
		if src := s.Thumbnails[size].URL; src != "" {
			return src
		}
	}
	return ""
}

func validID(id string) bool {
	if len(id) == 0 || len(id) > 128 {
		return false
	}
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_', r == '-':
		default:
			return false
		}
	}
	return true
}

var durationPattern = regexp.MustCompile(`^P(?:(\d+)D)?(?:T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+(?:\.\d+)?)S)?)?$`)

func parseDuration(value string) time.Duration {
	parts := durationPattern.FindStringSubmatch(value)
	if len(parts) == 0 {
		return 0
	}
	var seconds float64
	for i, unit := range []float64{86400, 3600, 60, 1} {
		if parts[i+1] == "" {
			continue
		}
		n, err := strconv.ParseFloat(parts[i+1], 64)
		if err != nil {
			return 0
		}
		seconds += n * unit
	}
	if seconds >= float64(time.Duration(1<<63-1))/float64(time.Second) {
		return 0
	}
	return time.Duration(seconds * float64(time.Second))
}

func invalid(operation string) error {
	return &providers.Error{Kind: providers.InvalidInput, Operation: operation}
}
