package artwork

import (
	"slices"
	"testing"
)

func TestRaster(t *testing.T) {
	cover := Raster(0, 20, 12)
	if len(cover) != 240 || slices.Min(cover) != 0 || slices.Max(cover) != 3 {
		t.Fatal("expected a bounded four-tone illustration")
	}
	if slices.Equal(cover, Raster(1, 20, 12)) || !slices.Equal(cover, Raster(0, 20, 12)) {
		t.Fatal("cover should be deterministic and vary by design")
	}
	if Raster(0, 0, 5) != nil {
		t.Fatal("empty cover should have no cells")
	}
}
