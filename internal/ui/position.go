package ui

import (
	"unsafe"

	"github.com/lxn/walk"
	"github.com/lxn/win"
)

const (
	spiGetWorkArea = 0x0030
	edgeMargin     = 12
)

// placeBottomRight anchors the window just above the tray corner, inside
// the work area (excludes the taskbar).
func placeBottomRight(mw *walk.MainWindow) {
	var rc win.RECT
	if !win.SystemParametersInfo(spiGetWorkArea, 0, unsafe.Pointer(&rc), 0) {
		return
	}
	b := mw.BoundsPixels()
	b.X = int(rc.Right) - b.Width - edgeMargin
	b.Y = int(rc.Bottom) - b.Height - edgeMargin
	if b.X < int(rc.Left) {
		b.X = int(rc.Left)
	}
	if b.Y < int(rc.Top) {
		b.Y = int(rc.Top)
	}
	mw.SetBoundsPixels(b)
}
