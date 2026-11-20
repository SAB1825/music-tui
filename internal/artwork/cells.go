// Package artwork supplies terminal-cell cover illustrations.
package artwork

import "math"

// Raster samples one of four original cover designs into a row-major palette.
func Raster(design, width, height int) []uint8 {
	if width <= 0 || height <= 0 {
		return nil
	}
	cells := make([]uint8, width*height)
	for row := range height {
		y := (float64(row) + .5) / float64(height)
		for column := range width {
			x := (float64(column) + .5) / float64(width)
			cells[row*width+column] = shade(design%4, x, y)
		}
	}
	return cells
}

func shade(design int, x, y float64) uint8 {
	moonX, moonY := .63, .34
	if design%2 == 1 {
		moonX, moonY = .33, .33
	}
	if math.Hypot((x-moonX)*1.25, y-moonY) < .18 {
		return 0
	}
	if y > .70+.12*math.Sin(x*7+float64(design)) {
		return 3
	}
	if y > .55+.17*math.Sin(x*9+float64(design)*2) {
		return 2
	}
	return 1
}
