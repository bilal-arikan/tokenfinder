// Package hotkey registers a global keyboard shortcut using a hidden
// message-only window. Must be called from the UI thread that runs the
// Windows message loop.
package hotkey

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/lxn/win"
	"golang.org/x/sys/windows"
)

// Modifier flags for RegisterHotKey.
const (
	ModAlt      uint32 = 0x0001
	ModControl  uint32 = 0x0002
	ModShift    uint32 = 0x0004
	ModWin      uint32 = 0x0008
	ModNoRepeat uint32 = 0x4000
)

var (
	user32               = windows.NewLazySystemDLL("user32.dll")
	procRegisterHotKey   = user32.NewProc("RegisterHotKey")
	procUnregisterHotKey = user32.NewProc("UnregisterHotKey")
	className            = "TokenFinderHotkeyWindow"
	callbacks            = map[int]func(){}
	classRegistered      bool
	wndProcPtr           uintptr
)

// Manager owns one registered hotkey.
type Manager struct {
	hwnd win.HWND
	id   int
}

func wndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	if msg == win.WM_HOTKEY {
		if cb, ok := callbacks[int(wParam)]; ok && cb != nil {
			cb()
		}
		return 0
	}
	return win.DefWindowProc(hwnd, msg, wParam, lParam)
}

func ensureClass() error {
	if classRegistered {
		return nil
	}
	wndProcPtr = syscall.NewCallback(wndProc)
	name, _ := syscall.UTF16PtrFromString(className)
	wc := win.WNDCLASSEX{
		LpfnWndProc:   wndProcPtr,
		HInstance:     win.GetModuleHandle(nil),
		LpszClassName: name,
	}
	wc.CbSize = uint32(unsafe.Sizeof(wc))
	if win.RegisterClassEx(&wc) == 0 {
		return fmt.Errorf("RegisterClassEx failed")
	}
	classRegistered = true
	return nil
}

// Register binds modifiers+vk to callback. id must be unique per process.
func Register(id int, modifiers, vk uint32, callback func()) (*Manager, error) {
	if err := ensureClass(); err != nil {
		return nil, err
	}
	name, _ := syscall.UTF16PtrFromString(className)
	hwnd := win.CreateWindowEx(0, name, name, 0, 0, 0, 0, 0, win.HWND_MESSAGE, 0, win.GetModuleHandle(nil), nil)
	if hwnd == 0 {
		return nil, fmt.Errorf("CreateWindowEx failed")
	}
	ret, _, err := procRegisterHotKey.Call(uintptr(hwnd), uintptr(id), uintptr(modifiers), uintptr(vk))
	if ret == 0 {
		win.DestroyWindow(hwnd)
		return nil, fmt.Errorf("RegisterHotKey failed: %v", err)
	}
	callbacks[id] = callback
	return &Manager{hwnd: hwnd, id: id}, nil
}

// Unregister releases the hotkey and destroys the hidden window.
func (m *Manager) Unregister() {
	if m == nil {
		return
	}
	procUnregisterHotKey.Call(uintptr(m.hwnd), uintptr(m.id))
	delete(callbacks, m.id)
	win.DestroyWindow(m.hwnd)
}
