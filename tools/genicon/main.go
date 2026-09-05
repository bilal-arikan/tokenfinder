// Command genicon renders the TokenFinder icon (a white key on a violet
// rounded square) at several sizes and writes assets/icon.ico plus
// docs/icon.png. No external dependencies; shapes are sampled with 4x4
// supersampling for anti-aliasing.
//
// Run from the repository root: go run ./tools/genicon
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
)

var sizes = []int{16, 20, 24, 32, 40, 48, 64, 128, 256}

const supersample = 4

// Palette matches internal/ui/theme.go.
var (
	gradTop    = rgb(0x9d, 0x82, 0xff)
	gradBottom = rgb(0x6a, 0x4b, 0xe6)
	keyColor   = rgb(0xff, 0xff, 0xff)
	shadow     = rgb(0x3a, 0x28, 0x8f)
)

func rgb(r, g, b uint8) color.RGBA { return color.RGBA{R: r, G: g, B: b, A: 0xff} }

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	frames := make([]*image.RGBA, 0, len(sizes))
	for _, s := range sizes {
		frames = append(frames, render(s))
	}

	if err := writeICO(filepath.Join(root, "assets", "icon.ico"), frames); err != nil {
		fail(err)
	}
	if err := writePNG(filepath.Join(root, "docs", "icon.png"), frames[len(frames)-1]); err != nil {
		fail(err)
	}
	fmt.Println("wrote assets/icon.ico and docs/icon.png")
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "genicon:", err)
	os.Exit(1)
}

// render draws one frame of the given pixel size.
func render(size int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	step := 1.0 / float64(size*supersample)
	for py := 0; py < size; py++ {
		for px := 0; px < size; px++ {
			var r, g, b, a float64
			for sy := 0; sy < supersample; sy++ {
				for sx := 0; sx < supersample; sx++ {
					u := (float64(px*supersample+sx) + 0.5) * step
					v := (float64(py*supersample+sy) + 0.5) * step
					c := shade(u, v)
					fa := float64(c.A) / 255
					r += float64(c.R) * fa
					g += float64(c.G) * fa
					b += float64(c.B) * fa
					a += fa
				}
			}
			n := float64(supersample * supersample)
			if a == 0 {
				continue
			}
			// Un-premultiply the averaged colour.
			img.SetRGBA(px, py, color.RGBA{
				R: uint8(r / a),
				G: uint8(g / a),
				B: uint8(b / a),
				A: uint8(a / n * 255),
			})
		}
	}
	return img
}

// shade returns the colour at unit coordinates (u, v) in [0,1].
func shade(u, v float64) color.RGBA {
	if !inRoundedSquare(u, v, 0.03, 0.24) {
		return color.RGBA{}
	}
	// Diagonal gradient background.
	t := clamp((u+v)/2, 0, 1)
	bg := lerp(gradTop, gradBottom, t)

	// Rotate the key 45 degrees so it points to the lower right.
	x, y := rotate(u, v, -math.Pi/4)

	// Soft shadow offset down-right.
	if keyShape(x-0.018, y+0.018) {
		bg = lerp(bg, shadow, 0.55)
	}
	if keyShape(x, y) {
		return keyColor
	}
	return bg
}

// keyShape tests the un-rotated key: ring on the left, shaft to the right,
// two teeth hanging down.
func keyShape(x, y float64) bool {
	const cy = 0.5
	ring := inCircle(x, y, 0.30, cy, 0.155) && !inCircle(x, y, 0.30, cy, 0.062)
	shaft := inRect(x, y, 0.40, cy-0.045, 0.80, cy+0.045)
	tooth1 := inRect(x, y, 0.62, cy+0.045, 0.685, cy+0.16)
	tooth2 := inRect(x, y, 0.735, cy+0.045, 0.80, cy+0.13)
	return ring || shaft || tooth1 || tooth2
}

func rotate(u, v, angle float64) (float64, float64) {
	du, dv := u-0.5, v-0.5
	s, c := math.Sincos(angle)
	return du*c - dv*s + 0.5, du*s + dv*c + 0.5
}

func inRoundedSquare(u, v, inset, radius float64) bool {
	lo, hi := inset, 1-inset
	if u < lo || u > hi || v < lo || v > hi {
		return false
	}
	cx := clamp(u, lo+radius, hi-radius)
	cy := clamp(v, lo+radius, hi-radius)
	return inCircle(u, v, cx, cy, radius)
}

func inCircle(x, y, cx, cy, r float64) bool {
	dx, dy := x-cx, y-cy
	return dx*dx+dy*dy <= r*r
}

func inRect(x, y, x0, y0, x1, y1 float64) bool {
	return x >= x0 && x <= x1 && y >= y0 && y <= y1
}

func clamp(x, lo, hi float64) float64 {
	return math.Max(lo, math.Min(hi, x))
}

func lerp(a, b color.RGBA, t float64) color.RGBA {
	m := func(x, y uint8) uint8 { return uint8(float64(x) + (float64(y)-float64(x))*t) }
	return color.RGBA{R: m(a.R, b.R), G: m(a.G, b.G), B: m(a.B, b.B), A: 0xff}
}

func writePNG(path string, img image.Image) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// writeICO stores every frame as a PNG-compressed ICO entry (supported since
// Windows Vista and by rsrc).
func writeICO(path string, frames []*image.RGBA) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var payloads [][]byte
	for _, fr := range frames {
		var buf bytes.Buffer
		if err := png.Encode(&buf, fr); err != nil {
			return err
		}
		payloads = append(payloads, buf.Bytes())
	}

	const headerSize, entrySize = 6, 16
	var out bytes.Buffer
	w := func(v any) { binary.Write(&out, binary.LittleEndian, v) }

	w(uint16(0)) // reserved
	w(uint16(1)) // type: icon
	w(uint16(len(frames)))

	offset := uint32(headerSize + entrySize*len(frames))
	for i, fr := range frames {
		s := fr.Bounds().Dx()
		dim := uint8(s)
		if s >= 256 {
			dim = 0
		}
		w(dim)       // width
		w(dim)       // height
		w(uint8(0))  // colour count
		w(uint8(0))  // reserved
		w(uint16(1)) // planes
		w(uint16(32))
		w(uint32(len(payloads[i])))
		w(offset)
		offset += uint32(len(payloads[i]))
	}
	for _, p := range payloads {
		out.Write(p)
	}
	return os.WriteFile(path, out.Bytes(), 0o644)
}
