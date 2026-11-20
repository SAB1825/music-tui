package youtube

import (
	"context"
	"net/url"
	"strings"

	"github.com/SAB1825/music-tui/internal/domain"
	"github.com/SAB1825/music-tui/internal/providers"
)

func playlistID(value string) string {
	value = strings.TrimSpace(value)
	if validID(value) {
		return value
	}
	u, err := url.Parse(value)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.User != nil {
		return ""
	}
	switch strings.ToLower(u.Host) {
	case "youtube.com", "www.youtube.com", "m.youtube.com", "music.youtube.com", "youtu.be":
		id := u.Query().Get("list")
		if validID(id) {
			return id
		}
	}
	return ""
}

func (c *Client) Playlist(ctx context.Context, value string) (domain.Playlist, error) {
	id := playlistID(value)
	if id == "" {
		return domain.Playlist{}, invalid("playlists")
	}
	q := url.Values{"part": {"snippet,contentDetails,status"}, "id": {id}}
	var response list[struct {
		ID             string
		Snippet        snippet
		ContentDetails struct{ ItemCount *int }
		Status         struct{ PrivacyStatus string }
	}]
	if err := c.get(ctx, "playlists", q, &response); err != nil {
		return domain.Playlist{}, err
	}
	for _, item := range response.Items {
		if item.ID == id && item.Status.PrivacyStatus == "public" {
			return domain.Playlist{Provider: "youtube", ID: id, Title: clean(item.Snippet.Title), Description: clean(item.Snippet.Description),
				ArtworkURL: coverURL(item.Snippet), TrackCount: item.ContentDetails.ItemCount,
				ExternalURL: "https://www.youtube.com/playlist?" + url.Values{"list": {id}}.Encode()}, nil
		}
	}
	return domain.Playlist{}, &providers.Error{Kind: providers.Unavailable, Operation: "playlists"}
}

func (c *Client) PlaylistTracks(ctx context.Context, id, token string) (domain.TrackPage, error) {
	if !validID(id) {
		return domain.TrackPage{}, invalid("playlistItems")
	}
	q := url.Values{"part": {"snippet,contentDetails"}, "playlistId": {id}}
	c.pageParams(q, token)
	var response list[struct {
		Snippet        snippet
		ContentDetails struct{ VideoID string }
	}]
	if err := c.get(ctx, "playlistItems", q, &response); err != nil {
		return domain.TrackPage{}, err
	}
	page := domain.TrackPage{NextToken: response.NextPageToken}
	for _, item := range response.Items {
		id := item.ContentDetails.VideoID
		if id == "" {
			id = item.Snippet.ResourceID.VideoID
		}
		if !validID(id) {
			page.Tracks = append(page.Tracks, domain.Track{Provider: "youtube", Title: "Unavailable video", Unavailable: true})
			continue
		}
		page.Tracks = append(page.Tracks, fromSnippet(id, item.Snippet))
	}
	return c.enrich(ctx, page)
}
