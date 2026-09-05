package ui

import (
	"github.com/lxn/walk"

	"tokenfinder/internal/autostart"
)

func (a *App) setupTray() error {
	ni, err := walk.NewNotifyIcon(a.mw)
	if err != nil {
		return err
	}
	a.tray = ni

	if err := ni.SetIcon(a.icon); err != nil {
		return err
	}
	if err := ni.SetToolTip(WindowTitle + " " + Version + " (Ctrl+Alt+T)"); err != nil {
		return err
	}

	ni.MouseDown().Attach(func(x, y int, button walk.MouseButton) {
		if button == walk.LeftButton {
			a.togglePanel()
		}
	})

	menu := ni.ContextMenu().Actions()
	menu.Add(newAction("Open panel", a.showPanel))
	menu.Add(newAction("Add entry...", func() {
		a.showPanel()
		a.addEntry()
	}))
	menu.Add(walk.NewSeparatorAction())
	menu.Add(a.autostartAction())
	menu.Add(walk.NewSeparatorAction())
	menu.Add(newAction("About", a.showAbout))
	menu.Add(newAction("Exit", a.quit))

	return ni.SetVisible(true)
}

func (a *App) showAbout() {
	walk.MsgBox(a.mw, "About "+WindowTitle,
		WindowTitle+" "+Version+"\n\n"+
			"Encrypted tray vault for tokens, API keys and passwords.\n"+
			"Vault: "+a.store.Path()+"\n\n"+
			RepoURL,
		walk.MsgBoxIconInformation)
}

func newAction(text string, handler func()) *walk.Action {
	act := walk.NewAction()
	act.SetText(text)
	act.Triggered().Attach(handler)
	return act
}

// autostartAction is a checkable menu item bound to the HKCU Run key.
func (a *App) autostartAction() *walk.Action {
	act := walk.NewAction()
	act.SetText("Start with Windows")
	act.SetCheckable(true)
	act.SetChecked(autostart.Enabled())
	act.Triggered().Attach(func() {
		var err error
		if autostart.Enabled() {
			err = autostart.Disable()
		} else {
			err = autostart.Enable()
		}
		if err != nil {
			a.showError("Autostart change failed", err)
		}
		act.SetChecked(autostart.Enabled())
	})
	return act
}
