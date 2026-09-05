package ui

import (
	"sync/atomic"
	"time"

	"github.com/lxn/walk"
)

// clipboardGuard copies text to the clipboard and clears it again after a
// delay, but only if the clipboard still holds the same text.
type clipboardGuard struct {
	seq atomic.Int64
}

// Copy places text on the clipboard. onCleared runs on the UI thread when the
// automatic clear actually happened.
func (g *clipboardGuard) Copy(w *walk.MainWindow, text string, clearAfter time.Duration, onCleared func()) error {
	if err := walk.Clipboard().SetText(text); err != nil {
		return err
	}
	id := g.seq.Add(1)

	go func() {
		time.Sleep(clearAfter)
		w.Synchronize(func() {
			// A newer copy superseded this one; leave the clipboard alone.
			if g.seq.Load() != id {
				return
			}
			cur, err := walk.Clipboard().Text()
			if err != nil || cur != text {
				return
			}
			if err := walk.Clipboard().Clear(); err != nil {
				return
			}
			if onCleared != nil {
				onCleared()
			}
		})
	}()
	return nil
}
