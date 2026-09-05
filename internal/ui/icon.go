package ui

import (
	"image"
	"image/color"

	"github.com/lxn/walk"
)

// rsrc numbers resources sequentially: the manifest takes id 1, so the
// RT_GROUP_ICON of assets/icon.ico gets id 2 (or 1 if built without manifest).
var iconResourceIDs = []int{2, 1}

// loadAppIcon returns the multi-size icon embedded via rsrc (assets/icon.ico).
// Falls back to a procedurally drawn key when the resource is missing, e.g.
// in a plain `go build` without rsrc.syso.
func loadAppIcon(dpi int) (*walk.Icon, error) {
	for _, id := range iconResourceIDs {
		if icon, err := walk.NewIconFromResourceIdWithSize(id, walk.Size{Width: 32, Height: 32}); err == nil {
			return icon, nil
		}
	}
	return drawFallbackIcon(dpi)
}

// drawFallbackIcon draws a small key glyph on a violet disc.
func drawFallbackIcon(dpi int) (*walk.Icon, error) {
	const size = 32
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	disc := color.RGBA{R: 0x8b, G: 0x6c, B: 0xff, A: 0xff}
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
			head := inCircle(fx, fy, 11, 16, 5.5) && !inCircle(fx, fy, 11, 16, 2.2)
			shaft := inRect(x, y, 15, 14, 26, 17)
			teeth := inRect(x, y, 20, 18, 22, 22) || inRect(x, y, 24, 18, 26, 21)
			if head || shaft || teeth {
				c = key
			}
			img.SetRGBA(x, y, c)
		}
	}

	return walk.NewIconFromImageForDPI(img, dpi)
}
