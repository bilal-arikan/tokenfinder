package ui

import (
	"syscall"
	"unsafe"

	"github.com/lxn/walk"
	"github.com/lxn/win"
	"golang.org/x/sys/windows"

	"tokenfinder/internal/store"
)

// Dark palette. walk.Color is a COLORREF (0x00BBGGRR); walk.RGB builds it.
var (
	clrBg         = walk.RGB(0x11, 0x13, 0x1a) // window background
	clrSurface    = walk.RGB(0x1a, 0x1d, 0x27) // inputs, even rows
	clrSurfaceAlt = walk.RGB(0x1f, 0x23, 0x2f) // odd rows
	clrText       = walk.RGB(0xe8, 0xea, 0xf1)
	clrMuted      = walk.RGB(0x8a, 0x91, 0xa6)
	clrAccent     = walk.RGB(0x8b, 0x6c, 0xff) // violet
	clrSelection  = walk.RGB(0x33, 0x2a, 0x63) // selected row background

	clrKindToken    = walk.RGB(0x5e, 0xa8, 0xff)
	clrKindAPIKey   = walk.RGB(0x3d, 0xd6, 0x8c)
	clrKindPassword = walk.RGB(0xff, 0x8a, 0x5c)
)

const (
	fontFamily = "Segoe UI"
	monoFamily = "Consolas"
	fontSize   = 10
)

// kindColor maps an entry kind to its accent color.
func kindColor(kind string) walk.Color {
	switch kind {
	case store.KindToken:
		return clrKindToken
	case store.KindAPIKey:
		return clrKindAPIKey
	case store.KindPassword:
		return clrKindPassword
	}
	return clrMuted
}

// scaleDPI converts a 96-DPI pixel value to the primary monitor's DPI.
// Used for TableView sizes that walk expects in native pixels.
func scaleDPI(px int) int {
	hdc := win.GetDC(0)
	if hdc == 0 {
		return px
	}
	defer win.ReleaseDC(0, hdc)
	dpi := int(win.GetDeviceCaps(hdc, win.LOGPIXELSY))
	if dpi <= 0 {
		return px
	}
	return px * dpi / 96
}

var (
	dwmapi                    = windows.NewLazySystemDLL("dwmapi.dll")
	procDwmSetWindowAttribute = dwmapi.NewProc("DwmSetWindowAttribute")
)

// DWMWA_USE_IMMERSIVE_DARK_MODE. 20 on Windows 10 20H1+ and Windows 11,
// 19 on older Windows 10 builds; both are tried.
const (
	dwmwaDarkMode       = 20
	dwmwaDarkModeLegacy = 19
)

// darkTitleBar switches the non-client area (title bar, borders) to dark.
func darkTitleBar(hwnd win.HWND) {
	on := int32(1)
	size := uintptr(unsafe.Sizeof(on))
	if r, _, _ := procDwmSetWindowAttribute.Call(uintptr(hwnd), dwmwaDarkMode, uintptr(unsafe.Pointer(&on)), size); r != 0 {
		procDwmSetWindowAttribute.Call(uintptr(hwnd), dwmwaDarkModeLegacy, uintptr(unsafe.Pointer(&on)), size)
	}
}

// setTheme applies an undocumented-but-stable uxtheme dark sub-app name
// (the same ones Explorer uses) so scrollbars, buttons, edits and combo boxes
// draw in dark mode.
func setTheme(hwnd win.HWND, theme string) {
	win.SetWindowTheme(hwnd, syscall.StringToUTF16Ptr(theme), nil)
}

func className(hwnd win.HWND) string {
	buf := make([]uint16, 64)
	n, err := win.GetClassName(hwnd, &buf[0], len(buf))
	if err != nil || n <= 0 {
		return ""
	}
	return syscall.UTF16ToString(buf[:n])
}

// darkenTable themes the inner list views (walk keeps them private) and
// their headers, and paints the empty area below the rows dark.
func darkenTable(tv *walk.TableView) {
	for child := win.GetWindow(tv.Handle(), win.GW_CHILD); child != 0; child = win.GetWindow(child, win.GW_HWNDNEXT) {
		if className(child) != "SysListView32" {
			continue
		}
		setTheme(child, "DarkMode_Explorer")
		win.SendMessage(child, win.LVM_SETBKCOLOR, 0, uintptr(clrSurface))
		win.SendMessage(child, win.LVM_SETTEXTBKCOLOR, 0, uintptr(clrSurface))
		win.SendMessage(child, win.LVM_SETTEXTCOLOR, 0, uintptr(clrText))
		if hdr := win.HWND(win.SendMessage(child, win.LVM_GETHEADER, 0, 0)); hdr != 0 {
			setTheme(hdr, "DarkMode_ItemsView")
		}
	}
}

// darkenControls walks a container and applies dark control themes.
func darkenControls(c walk.Container) {
	children := c.Children()
	for i := 0; i < children.Len(); i++ {
		switch w := children.At(i).(type) {
		case *walk.PushButton:
			setTheme(w.Handle(), "DarkMode_Explorer")
		case *walk.ComboBox:
			setTheme(w.Handle(), "DarkMode_CFD")
		case *walk.LineEdit:
			setTheme(w.Handle(), "DarkMode_CFD")
		case *walk.TableView:
			darkenTable(w)
		case walk.Container:
			darkenControls(w)
		}
	}
}

// darkenForm applies the full dark treatment to a top-level window.
func darkenForm(f walk.Form) {
	darkTitleBar(f.Handle())
	darkenControls(f)
}
