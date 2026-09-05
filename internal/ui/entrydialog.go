package ui

import (
	"strings"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"

	"tokenfinder/internal/store"
)

// showEntryDialog edits e in place. Returns true when the user pressed OK.
func showEntryDialog(owner walk.Form, e *store.Entry) (bool, error) {
	var (
		dlg       *walk.Dialog
		db        *walk.DataBinder
		okBtn     *walk.PushButton
		cancelBtn *walk.PushButton
		revealBtn *walk.PushButton
		valueEdit *walk.LineEdit
		hidden    = true
	)

	title := "Add entry"
	if e.ID != "" {
		title = "Edit entry"
	}

	field := func(text string) Label {
		return Label{Text: text, TextColor: clrMuted}
	}
	input := func(assignTo **walk.LineEdit, bind string, password bool) LineEdit {
		return LineEdit{
			AssignTo:     assignTo,
			Text:         Bind(bind),
			Background:   SolidColorBrush{Color: clrSurface},
			TextColor:    clrText,
			PasswordMode: password,
		}
	}

	err := Dialog{
		AssignTo:      &dlg,
		Title:         title,
		DefaultButton: &okBtn,
		CancelButton:  &cancelBtn,
		MinSize:       Size{Width: 460, Height: 0},
		Background:    SolidColorBrush{Color: clrBg},
		Font:          Font{Family: fontFamily, PointSize: fontSize},
		Layout:        VBox{Margins: Margins{Left: 16, Top: 14, Right: 16, Bottom: 12}, Spacing: 10},
		DataBinder: DataBinder{
			AssignTo:   &db,
			DataSource: e,
		},
		Children: []Widget{
			Label{
				Text:      title,
				TextColor: clrAccent,
				Font:      Font{Family: fontFamily, PointSize: 13, Bold: true},
			},
			Composite{
				Layout: Grid{Columns: 3, MarginsZero: true, Spacing: 8},
				Children: []Widget{
					field("Name"),
					LineEdit{
						Text:       Bind("Name"),
						Background: SolidColorBrush{Color: clrSurface},
						TextColor:  clrText,
						ColumnSpan: 2,
					},

					field("Type"),
					ComboBox{Value: Bind("Kind"), Model: store.Kinds, ColumnSpan: 2},

					field("Value"),
					input(&valueEdit, "Value", true),
					PushButton{
						AssignTo: &revealBtn,
						Text:     "Show",
						MaxSize:  Size{Width: 60, Height: 0},
						OnClicked: func() {
							hidden = !hidden
							valueEdit.SetPasswordMode(hidden)
							if hidden {
								revealBtn.SetText("Show")
							} else {
								revealBtn.SetText("Hide")
							}
						},
					},

					field("Note"),
					TextEdit{
						Text:       Bind("Note"),
						Background: SolidColorBrush{Color: clrSurface},
						TextColor:  clrText,
						ColumnSpan: 2,
						MinSize:    Size{Width: 0, Height: 60},
						VScroll:    true,
					},
				},
			},
			Composite{
				Layout: HBox{MarginsZero: true, Spacing: 6},
				Children: []Widget{
					HSpacer{},
					PushButton{
						AssignTo: &okBtn,
						Text:     "Save",
						OnClicked: func() {
							if err := db.Submit(); err != nil {
								walk.MsgBox(dlg, title, err.Error(), walk.MsgBoxIconError)
								return
							}
							e.Name = strings.TrimSpace(e.Name)
							if e.Name == "" {
								walk.MsgBox(dlg, title, "Name is required.", walk.MsgBoxIconWarning)
								return
							}
							if e.Value == "" {
								walk.MsgBox(dlg, title, "Value is required.", walk.MsgBoxIconWarning)
								return
							}
							if e.Kind == "" {
								e.Kind = store.KindOther
							}
							dlg.Accept()
						},
					},
					PushButton{
						AssignTo:  &cancelBtn,
						Text:      "Cancel",
						OnClicked: func() { dlg.Cancel() },
					},
				},
			},
		},
	}.Create(owner)
	if err != nil {
		return false, err
	}

	darkenForm(dlg)
	return dlg.Run() == walk.DlgCmdOK, nil
}
