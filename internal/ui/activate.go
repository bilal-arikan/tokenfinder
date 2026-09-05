package ui

import (
	"syscall"

	"github.com/lxn/win"
)

// ActivateExisting shows the panel of an already running instance.
// Returns false when no such window exists.
func ActivateExisting() bool {
	title, err := syscall.UTF16PtrFromString(WindowTitle)
	if err != nil {
		return false
	}
	hwnd := win.FindWindow(nil, title)
	if hwnd == 0 {
		return false
	}
	win.ShowWindow(hwnd, win.SW_SHOW)
	win.SetForegroundWindow(hwnd)
	return true
}
