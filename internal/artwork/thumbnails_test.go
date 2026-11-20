package artwork

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type thumbnailTransport func(*http.Request) (*http.Response, error)

func (f thumbnailTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestThumbnailFetch(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	img.SetRGBA(0, 2, color.RGBA{R: 210, G: 90, B: 10, A: 255})
	var data bytes.Buffer
	if err := png.Encode(&data, img); err != nil {
		t.Fatal(err)
	}
	calls := 0
	c := NewThumbnails(thumbnailTransport(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Header.Get("X-Goog-Api-Key") != "" {
			t.Fatal("thumbnail request carried credentials")
		}
		w := httptest.NewRecorder()
		var payload []byte
		switch r.URL.Path {
		case "/redirect":
			http.Redirect(w, r, "https://unrelated.invalid", http.StatusFound)
		case "/bad":
			payload = []byte("invalid image")
		case "/large":
			payload = []byte(strings.Repeat("x", thumbnailLimit+1))
		default:
			payload = data.Bytes()
		}
		if payload != nil {
			if _, err := w.Write(payload); err != nil {
				t.Fatal(err)
			}
		}
		return w.Result(), nil
	}))
	art, err := c.Fetch(t.Context(), "https://i.ytimg.com/ok")
	if err != nil || art[0] != [3]uint8{210, 90, 10} {
		t.Fatalf("thumbnail sample: %v %v", art[0], err)
	}
	for _, src := range []string{"https://unrelated.invalid/image", "http://i.ytimg.com/test", "https://i.ytimg.com/redirect", "https://i.ytimg.com/bad", "https://i.ytimg.com/large"} {
		if _, err := c.Fetch(t.Context(), src); err == nil {
			t.Fatalf("accepted %s", src)
		}
	}
	if calls != 4 {
		t.Fatalf("unexpected deliveries: %d", calls)
	}
}
