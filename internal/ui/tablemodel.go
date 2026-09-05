package ui

import (
	"strings"

	"github.com/lxn/walk"

	"tokenfinder/internal/store"
)

// Column indexes of the entry table.
const (
	colName = iota
	colKind
	colValue
	colUpdated
	colNote
)

// noteMaxChars caps the note preview shown in the table.
const noteMaxChars = 48

// entryModel adapts []store.Entry to walk.TableModel.
type entryModel struct {
	walk.TableModelBase
	items []store.Entry
}

func newEntryModel() *entryModel { return &entryModel{} }

func (m *entryModel) RowCount() int { return len(m.items) }

func (m *entryModel) Value(row, col int) interface{} {
	e := m.items[row]
	switch col {
	case colName:
		return e.Name
	case colKind:
		return e.Kind
	case colValue:
		return e.Masked()
	case colUpdated:
		return e.UpdatedAt.Format("2006-01-02 15:04")
	case colNote:
		return notePreview(e.Note)
	}
	return ""
}

// notePreview collapses a note to one line and truncates it with an ellipsis.
func notePreview(note string) string {
	first, _, multi := strings.Cut(strings.ReplaceAll(note, "\r\n", "\n"), "\n")
	line := strings.TrimSpace(first)
	if multi {
		line += " …"
	}
	r := []rune(line)
	if len(r) > noteMaxChars {
		return string(r[:noteMaxChars-1]) + "…"
	}
	return line
}

// SetItems replaces the rows and notifies the view.
func (m *entryModel) SetItems(items []store.Entry) {
	m.items = items
	m.PublishRowsReset()
}

// At returns the entry at row, or false when out of range.
func (m *entryModel) At(row int) (store.Entry, bool) {
	if row < 0 || row >= len(m.items) {
		return store.Entry{}, false
	}
	return m.items[row], true
}
