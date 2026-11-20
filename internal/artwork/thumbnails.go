package artwork

import (
	"bytes"
	"context"
	"errors"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"net/url"
	"time"
)

const thumbnailLimit = 2 << 20

// Thumbnail stores original RGB samples, independent of the current layout.
type Thumbnail [12][3]uint8

type Thumbnails struct{ client *http.Client }

func NewThumbnails(transport http.RoundTripper) Thumbnails {
	return Thumbnails{client: &http.Client{Transport: transport, Timeout: 4 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
}

func (t Thumbnails) Fetch(ctx context.Context, src string) (Thumbnail, error) {
	u, err := url.Parse(src)
	if err != nil || u.Scheme != "https" || u.User != nil ||
		(u.Host != "i.ytimg.com" && u.Host != "img.youtube.com" && u.Host != "i9.ytimg.com") {
		return Thumbnail{}, errors.New("unsupported thumbnail source")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, src, nil)
	if err != nil {
		return Thumbnail{}, errors.New("invalid thumbnail request")
	}
	resp, err := t.client.Do(req)
	if err != nil {
		return Thumbnail{}, errors.New("thumbnail request failed")
	}
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, thumbnailLimit+1))
	closeErr := resp.Body.Close()
	if resp.StatusCode != http.StatusOK || readErr != nil || closeErr != nil || len(body) > thumbnailLimit {
		return Thumbnail{}, errors.New("thumbnail unavailable")
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(body))
	if err != nil || cfg.Width < 1 || cfg.Height < 1 || cfg.Width > 2048 || cfg.Height > 2048 {
		return Thumbnail{}, errors.New("invalid thumbnail dimensions")
	}
	img, _, err := image.Decode(bytes.NewReader(body))
	if err != nil {
		return Thumbnail{}, errors.New("invalid thumbnail image")
	}
	var result Thumbnail
	b := img.Bounds()
	for y := range 2 {
		for x := range 6 {
			r, g, bv, _ := img.At(b.Min.X+(2*x+1)*b.Dx()/12, b.Min.Y+(2*y+1)*b.Dy()/4).RGBA()
			result[y*6+x] = [3]uint8{uint8(r >> 8), uint8(g >> 8), uint8(bv >> 8)}
		}
	}
	return result, nil
}
