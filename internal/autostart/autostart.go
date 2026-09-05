// Package autostart toggles the HKCU Run registry entry for the app.
package autostart

import (
	"os"

	"golang.org/x/sys/windows/registry"
)

const (
	runKey    = `Software\Microsoft\Windows\CurrentVersion\Run`
	valueName = "TokenFinder"
)

// Enabled reports whether the app is registered to start with Windows.
func Enabled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()
	_, _, err = k.GetStringValue(valueName)
	return err == nil
}

// Enable registers the current executable in the Run key.
func Enable() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	return k.SetStringValue(valueName, `"`+exe+`"`)
}

// Disable removes the Run key entry. Missing entry is not an error.
func Disable() error {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	if err := k.DeleteValue(valueName); err != nil && err != registry.ErrNotExist {
		return err
	}
	return nil
}
