package ui

import (
	"fmt"
	"os"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"

	"tokenfinder/internal/store"
)

func (a *App) buildMainWindow() error {
	err := MainWindow{
		AssignTo:   &a.mw,
		Title:      WindowTitle,
		Visible:    false,
		Size:       Size{Width: 760, Height: 460},
		MinSize:    Size{Width: 480, Height: 300},
		Background: SolidColorBrush{Color: clrBg},
		Font:       Font{Family: fontFamily, PointSize: fontSize},
		Layout:     VBox{Margins: Margins{Left: 14, Top: 12, Right: 14, Bottom: 10}, Spacing: 8},
		Children: []Widget{
			// Header: title on the left, entry counter on the right.
			Composite{
				Layout: HBox{MarginsZero: true},
				Children: []Widget{
					Label{
						Text:      "◆ TokenFinder",
						TextColor: clrAccent,
						Font:      Font{Family: fontFamily, PointSize: 15, Bold: true},
					},
					HSpacer{},
					Label{AssignTo: &a.count, Text: "", TextColor: clrMuted},
				},
			},
			// Thin accent stripe under the header.
			Composite{
				Background: SolidColorBrush{Color: clrAccent},
				MinSize:    Size{Width: 0, Height: 2},
				MaxSize:    Size{Width: 0, Height: 2},
			},
			LineEdit{
				AssignTo:      &a.search,
				CueBanner:     "Search by name, type or note...",
				Background:    SolidColorBrush{Color: clrSurface},
				TextColor:     clrText,
				OnTextChanged: a.refresh,
				OnKeyDown:     a.onSearchKey,
			},
			TableView{
				AssignTo:            &a.table,
				Model:               a.model,
				LastColumnStretched: true,
				CustomHeaderHeight:  scaleDPI(30),
				CustomRowHeight:     scaleDPI(26),
				StyleCell:           a.styleCell,
				Columns: []TableViewColumn{
					{Title: "Name", Width: 170},
					{Title: "Type", Width: 85},
					{Title: "Value", Width: 150},
					{Title: "Updated", Width: 115},
					{Title: "Note", Width: 160},
				},
				ContextMenuItems: []MenuItem{
					Action{Text: "Copy value\tEnter", OnTriggered: a.copySelected},
					Action{Text: "Copy name\tCtrl+Enter", OnTriggered: a.copyName},
					Action{Text: "Copy note", OnTriggered: a.copyNote},
					Separator{},
					Action{Text: "Edit...", OnTriggered: a.editEntry},
					Action{Text: "Delete", OnTriggered: a.deleteEntry},
				},
				OnItemActivated: a.copySelected,
				OnKeyDown:       a.onTableKey,
			},
			Composite{
				Layout: HBox{MarginsZero: true, Spacing: 6},
				Children: []Widget{
					PushButton{Text: "Copy", OnClicked: a.copySelected, ToolTipText: "Copy value (Enter)"},
					PushButton{Text: "Copy Name", OnClicked: a.copyName, ToolTipText: "Copy name (Ctrl+Enter)"},
					PushButton{Text: "Add", OnClicked: a.addEntry},
					PushButton{Text: "Edit", OnClicked: a.editEntry},
					PushButton{Text: "Delete", OnClicked: a.deleteEntry},
					HSpacer{},
					PushButton{Text: "Hide", OnClicked: a.hidePanel},
				},
			},
			Label{
				AssignTo:  &a.status,
				Text:      "Ctrl+Alt+T toggles this panel. Enter copies value, Ctrl+Enter copies name.",
				TextColor: clrMuted,
			},
		},
	}.Create()
	if err != nil {
		return err
	}

	darkenForm(a.mw)

	// Closing the window only hides it; the tray "Exit" action really quits.
	a.mw.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		if !a.quitting {
			*canceled = true
			a.hidePanel()
		}
	})
	return nil
}

// styleCell paints rows in the dark palette: selection in violet, kinds in
// their accent color, secret values in a muted monospace face.
func (a *App) styleCell(style *walk.CellStyle) {
	row := style.Row()
	if row < 0 {
		a.paintHeader(style)
		return
	}
	e, ok := a.model.At(row)
	if !ok {
		return
	}
	selected := row == a.table.CurrentIndex()

	switch {
	case selected:
		style.BackgroundColor = clrSelection
	case row%2 == 1:
		style.BackgroundColor = clrSurfaceAlt
	default:
		style.BackgroundColor = clrSurface
	}

	switch style.Col() {
	case colName:
		style.TextColor = clrText
		if selected {
			style.Font = a.fontBold
		}
	case colKind:
		style.TextColor = kindColor(e.Kind)
	case colValue:
		style.TextColor = clrMuted
		style.Font = a.fontMono
	case colNote:
		style.TextColor = clrMuted
		style.Font = a.fontItalic
	default:
		style.TextColor = clrMuted
	}
}

// paintHeader is called by walk after the default header item paint (row -1).
// Default drawing uses the light theme's dark text, so it is painted over.
func (a *App) paintHeader(style *walk.CellStyle) {
	canvas := style.Canvas()
	if canvas == nil {
		return
	}
	b := style.BoundsPixels()

	bg, err := walk.NewSolidColorBrush(clrSurfaceAlt)
	if err != nil {
		return
	}
	defer bg.Dispose()
	canvas.FillRectanglePixels(bg, b)

	line, err := walk.NewCosmeticPen(walk.PenSolid, clrAccent)
	if err == nil {
		defer line.Dispose()
		canvas.DrawLinePixels(line, walk.Point{X: b.X, Y: b.Y + b.Height - 1}, walk.Point{X: b.X + b.Width, Y: b.Y + b.Height - 1})
	}

	col := style.Col()
	if col < 0 || col >= a.table.Columns().Len() {
		return
	}
	pad := scaleDPI(6)
	textBounds := walk.Rectangle{X: b.X + pad, Y: b.Y, Width: b.Width - 2*pad, Height: b.Height - 1}
	canvas.DrawTextPixels(a.table.Columns().At(col).Title(), a.fontBold, clrMuted, textBounds,
		walk.TextLeft|walk.TextVCenter|walk.TextSingleLine|walk.TextEndEllipsis)
}

func (a *App) onSearchKey(key walk.Key) {
	switch key {
	case walk.KeyReturn:
		a.copyByModifier()
	case walk.KeyDown:
		a.table.SetFocus()
		if a.table.CurrentIndex() < 0 && a.model.RowCount() > 0 {
			a.table.SetCurrentIndex(0)
		}
	case walk.KeyEscape:
		a.hidePanel()
	}
}

func (a *App) onTableKey(key walk.Key) {
	switch key {
	case walk.KeyReturn:
		a.copyByModifier()
	case walk.KeyDelete:
		a.deleteEntry()
	case walk.KeyEscape:
		a.hidePanel()
	}
}

// refresh reloads the table from the store using the current search text.
func (a *App) refresh() {
	items := a.store.Search(a.search.Text())
	a.model.SetItems(items)
	if len(items) > 0 {
		a.table.SetCurrentIndex(0)
	}
	total := len(a.store.List())
	switch {
	case os.Getenv("TOKENFINDER_DEBUG") != "":
		a.status.SetText(fmt.Sprintf("vault=%s total=%d shown=%d q=%q args=%q", a.store.Path(), total, len(items), a.search.Text(), os.Args))
	case total == 0:
		// Make it obvious which file is in use when nothing shows up.
		a.status.SetText("Vault is empty: " + a.store.Path())
	}
	if len(items) == total {
		a.count.SetText(fmt.Sprintf("%d secrets", total))
	} else {
		a.count.SetText(fmt.Sprintf("%d of %d", len(items), total))
	}
}

func (a *App) showPanel() {
	// Always open with a clean filter so a stale search never hides entries.
	if a.search.Text() != "" {
		a.search.SetText("") // triggers refresh via OnTextChanged
	} else {
		a.refresh()
	}
	placeBottomRight(a.mw)
	a.mw.Show()
	win.SetForegroundWindow(a.mw.Handle())
	a.search.SetFocus()
	a.search.SetTextSelection(0, -1)
}

func (a *App) hidePanel() {
	a.mw.Hide()
}

func (a *App) togglePanel() {
	if a.mw.Visible() {
		a.hidePanel()
	} else {
		a.showPanel()
	}
}

// selectedEntry returns the highlighted row, falling back to the first row.
func (a *App) selectedEntry() (store.Entry, bool) {
	idx := a.table.CurrentIndex()
	if idx < 0 {
		idx = 0
	}
	return a.model.At(idx)
}

// copyByModifier maps Enter to the value and Ctrl+Enter to the name.
func (a *App) copyByModifier() {
	if walk.ControlDown() {
		a.copyName()
	} else {
		a.copySelected()
	}
}

// copySelected copies the secret value, hides the panel and schedules the
// clipboard wipe.
func (a *App) copySelected() {
	e, ok := a.selectedEntry()
	if !ok {
		return
	}
	if err := a.clip.Copy(a.mw, e.Value, clipboardClearAfter, func() {
		a.status.SetText("Clipboard cleared.")
	}); err != nil {
		a.showError("Copy failed", err)
		return
	}
	a.status.SetText("Copied value of \"" + e.Name + "\". Clipboard clears in 30s.")
	a.hidePanel()
	a.tray.ShowInfo(WindowTitle, "Copied value of \""+e.Name+"\" to clipboard.")
}

// copyName copies the entry name. Names are not secret, so the panel stays
// open and the clipboard is not wiped.
func (a *App) copyName() {
	e, ok := a.selectedEntry()
	if !ok {
		return
	}
	if err := walk.Clipboard().SetText(e.Name); err != nil {
		a.showError("Copy failed", err)
		return
	}
	a.status.SetText("Copied name \"" + e.Name + "\".")
}

// copyNote copies the full note text (the table only shows a preview).
func (a *App) copyNote() {
	e, ok := a.selectedEntry()
	if !ok {
		return
	}
	if e.Note == "" {
		a.status.SetText("\"" + e.Name + "\" has no note.")
		return
	}
	if err := walk.Clipboard().SetText(e.Note); err != nil {
		a.showError("Copy failed", err)
		return
	}
	a.status.SetText("Copied note of \"" + e.Name + "\".")
}

func (a *App) addEntry() {
	e := store.Entry{Kind: store.KindToken}
	ok, err := showEntryDialog(a.mw, &e)
	if err != nil {
		a.showError("Dialog failed", err)
		return
	}
	if !ok {
		return
	}
	if err := a.store.Upsert(e); err != nil {
		a.showError("Save failed", err)
		return
	}
	a.refresh()
}

func (a *App) editEntry() {
	e, ok := a.selectedEntry()
	if !ok {
		return
	}
	ok, err := showEntryDialog(a.mw, &e)
	if err != nil {
		a.showError("Dialog failed", err)
		return
	}
	if !ok {
		return
	}
	if err := a.store.Upsert(e); err != nil {
		a.showError("Save failed", err)
		return
	}
	a.refresh()
}

func (a *App) deleteEntry() {
	e, ok := a.selectedEntry()
	if !ok {
		return
	}
	answer := walk.MsgBox(a.mw, "Delete entry",
		"Delete \""+e.Name+"\"? This cannot be undone.",
		walk.MsgBoxYesNo|walk.MsgBoxIconWarning|walk.MsgBoxDefButton2)
	if answer != win.IDYES {
		return
	}
	if err := a.store.Delete(e.ID); err != nil {
		a.showError("Delete failed", err)
		return
	}
	a.refresh()
}
