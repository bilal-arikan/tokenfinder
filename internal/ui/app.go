// Package ui builds the tray icon and the secret picker panel using lxn/walk.
package ui

import (
	"time"

	"github.com/lxn/walk"

	"tokenfinder/internal/hotkey"
	"tokenfinder/internal/store"
)

const (
	// WindowTitle is used both for the panel and for FindWindow lookups.
	WindowTitle = "TokenFinder"

	clipboardClearAfter = 30 * time.Second

	hotkeyID = 1
	vkT      = 0x54 // virtual key code for 'T'
)

// App wires the store, the main window, the tray icon and the hotkey together.
type App struct {
	store *store.Store
	model *entryModel
	clip  clipboardGuard

	mw     *walk.MainWindow
	search *walk.LineEdit
	table  *walk.TableView
	status *walk.Label
	count  *walk.Label
	tray   *walk.NotifyIcon
	icon   *walk.Icon

	fontBold   *walk.Font
	fontMono   *walk.Font
	fontItalic *walk.Font

	quitting bool
}

// Run creates the UI, starts hidden in the tray and blocks until exit.
func Run(st *store.Store) error {
	a := &App{store: st, model: newEntryModel()}

	var err error
	if a.fontBold, err = walk.NewFont(fontFamily, fontSize, walk.FontBold); err != nil {
		return err
	}
	if a.fontMono, err = walk.NewFont(monoFamily, fontSize, 0); err != nil {
		return err
	}
	if a.fontItalic, err = walk.NewFont(fontFamily, fontSize, walk.FontItalic); err != nil {
		return err
	}

	if err := a.buildMainWindow(); err != nil {
		return err
	}

	icon, err := buildIcon(a.mw.DPI())
	if err != nil {
		return err
	}
	a.icon = icon
	a.mw.SetIcon(icon)

	if err := a.setupTray(); err != nil {
		return err
	}
	defer a.tray.Dispose()

	hk, err := hotkey.Register(hotkeyID, hotkey.ModControl|hotkey.ModAlt|hotkey.ModNoRepeat, vkT, a.togglePanel)
	if err != nil {
		a.tray.ShowInfo(WindowTitle, "Global hotkey Ctrl+Alt+T is unavailable: "+err.Error())
	} else {
		defer hk.Unregister()
	}

	a.refresh()
	a.mw.Run()
	return nil
}

// quit exits the message loop; Closing handler sees quitting=true and lets it through.
func (a *App) quit() {
	a.quitting = true
	walk.App().Exit(0)
}

func (a *App) showError(context string, err error) {
	walk.MsgBox(a.mw, WindowTitle, context+": "+err.Error(), walk.MsgBoxIconError)
}
