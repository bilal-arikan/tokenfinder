// Package singleinstance guarantees only one copy of the app runs per user.
package singleinstance

import (
	"errors"

	"golang.org/x/sys/windows"
)

// Acquire creates a named mutex. already is true when another instance owns it.
func Acquire(name string) (release func(), already bool, err error) {
	n, err := windows.UTF16PtrFromString("Local\\" + name)
	if err != nil {
		return nil, false, err
	}
	h, err := windows.CreateMutex(nil, false, n)
	if errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
		if h != 0 {
			windows.CloseHandle(h)
		}
		return func() {}, true, nil
	}
	if err != nil {
		return nil, false, err
	}
	return func() { windows.CloseHandle(h) }, false, nil
}
