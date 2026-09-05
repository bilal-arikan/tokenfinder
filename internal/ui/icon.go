package ui

import (
	"image"
	"image/color"

	"github.com/lxn/walk"
)

// buildIcon draws a small key glyph on a blue disc so the app needs no
// external .ico asset.
func buildIcon(dpi int) (*walk.Icon, error) {
	const size = 32
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	disc := color.RGBA{R: 0x1f, G: 0x6f, B: 0xeb, A: 0xff}
	key := color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}

	inCircle := func(x, y, cx, cy, r float64) bool {
		dx, dy := x-cx, y-cy
		return dx*dx+dy*dy <= r*r
	}
	inRect := func(x, y, x0, y0, x1, y1 int) bool {
		return x >= x0 && x <= x1 && y >= y0 && y <= y1
	}

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			fx, fy := float64(x)+0.5, float64(y)+0.5

			if !inCircle(fx, fy, 16, 16, 15.5) {
				continue
			}
			c := disc

			// Key head: ring on the left.
			head := inCircle(fx, fy, 11, 16, 5.5) && !inCircle(fx, fy, 11, 16, 2.2)
			// Shaft: horizontal bar to the right.
			shaft := inRect(x, y, 15, 14, 26, 17)
			// Two teeth hanging from the shaft.
			teeth := inRect(x, y, 20, 18, 22, 22) || inRect(x, y, 24, 18, 26, 21)

			if head || shaft || teeth {
				c = key
			}
			img.SetRGBA(x, y, c)
		}
	}

	return walk.NewIconFromImageForDPI(img, dpi)
}
