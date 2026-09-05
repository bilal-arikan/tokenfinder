// TokenFinder: a Windows tray app that stores tokens, API keys and passwords
// in a DPAPI-encrypted vault and copies them to the clipboard on demand.
package main

//go:generate go run ./tools/genicon

import (
	"flag"
	"os"

	"github.com/lxn/walk"

	"tokenfinder/internal/singleinstance"
	"tokenfinder/internal/store"
	"tokenfinder/internal/ui"
)

const mutexName = "TokenFinder.SingleInstance"

func main() {
	vaultFlag := flag.String("vault", "", "path of the encrypted vault file (default: %APPDATA%\\TokenFinder\\vault.dat)")
	flag.Parse()

	release, already, err := singleinstance.Acquire(mutexName)
	if err != nil {
		fatal("Single instance check failed", err)
	}
	if already {
		// Second launch acts as "open panel" for the running instance.
		ui.ActivateExisting()
		return
	}
	defer release()

	path := *vaultFlag
	if path == "" {
		if path, err = store.DefaultVaultPath(); err != nil {
			fatal("Cannot resolve vault path", err)
		}
	}
	st, err := store.Open(path)
	if err != nil {
		fatal("Cannot open vault at "+path, err)
	}

	if err := ui.Run(st); err != nil {
		fatal("UI failed", err)
	}
}

func fatal(context string, err error) {
	walk.MsgBox(nil, ui.WindowTitle, context+": "+err.Error(), walk.MsgBoxIconError)
	os.Exit(1)
}
