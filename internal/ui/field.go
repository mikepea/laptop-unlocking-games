package ui

import "strings"

// Field is a single-line text input, shared by every game that asks the player
// to type an answer. It is deliberately tiny: no selection, no cursor movement,
// no history. Answers here are a few characters long and the backspace key is
// the only editing anyone needs.
type Field struct {
	value []rune

	// Max caps the length. Zero means no limit.
	Max int
	// Filter rejects characters that do not belong in this answer. Nil accepts
	// anything printable.
	Filter func(rune) bool
}

// Insert appends a character if it passes the filter and there is room.
func (f *Field) Insert(r rune) {
	if r < ' ' || r == 127 {
		return
	}
	if f.Filter != nil && !f.Filter(r) {
		return
	}
	if f.Max > 0 && len(f.value) >= f.Max {
		return
	}
	f.value = append(f.value, r)
}

// Backspace removes the last character.
func (f *Field) Backspace() {
	if len(f.value) > 0 {
		f.value = f.value[:len(f.value)-1]
	}
}

// Clear empties the field, keeping the allocation for the next answer.
func (f *Field) Clear() { f.value = f.value[:0] }

// Value is what has been typed.
func (f *Field) Value() string { return string(f.value) }

// Len is the number of characters typed.
func (f *Field) Len() int { return len(f.value) }

// Empty reports whether nothing has been typed.
func (f *Field) Empty() bool { return len(f.value) == 0 }

// Render draws the field with a block cursor, padded to width so the layout
// around it does not shift as characters are added.
func (f *Field) Render(width int) string {
	var b strings.Builder
	b.WriteString(Selected.Render(string(f.value)))
	b.WriteString(Cursor.Render(" "))

	if pad := width - len(f.value) - 1; pad > 0 {
		b.WriteString(Dim.Render(strings.Repeat(".", pad)))
	}
	return b.String()
}

// Digits accepts the characters a number can be made of. Minus is allowed so a
// player can start typing a negative answer and find out for themselves that it
// was not one.
func Digits(r rune) bool { return (r >= '0' && r <= '9') || r == '-' }

// Letters accepts letters, apostrophes and hyphens: everything a spelling can
// contain and nothing else.
func Letters(r rune) bool {
	switch {
	case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z':
		return true
	case r == '\'' || r == '-':
		return true
	}
	return false
}
