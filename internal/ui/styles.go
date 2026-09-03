// Package ui holds the shared look of every unlock screen. Colours are chosen
// from the 256-colour cube; Lip Gloss downsamples them automatically on a plain
// Linux TTY, which is where this spends most of its life.
package ui

import "github.com/charmbracelet/lipgloss"

// Palette.
var (
	ColorAccent  = lipgloss.Color("39")  // bright blue
	ColorGood    = lipgloss.Color("42")  // green
	ColorBad     = lipgloss.Color("203") // red
	ColorWarn    = lipgloss.Color("214") // amber
	ColorMuted   = lipgloss.Color("245") // grey
	ColorDim     = lipgloss.Color("240") // darker grey
	ColorGold    = lipgloss.Color("220") // achievements
	ColorInverse = lipgloss.Color("236") // cursor background
)

// Shared styles.
var (
	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorAccent)

	Subtitle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	Muted = lipgloss.NewStyle().Foreground(ColorMuted)
	Dim   = lipgloss.NewStyle().Foreground(ColorDim)
	Good  = lipgloss.NewStyle().Foreground(ColorGood)
	Bad   = lipgloss.NewStyle().Foreground(ColorBad)
	Warn  = lipgloss.NewStyle().Foreground(ColorWarn)
	Gold  = lipgloss.NewStyle().Foreground(ColorGold).Bold(true)

	// Typed returns text the player got right.
	Typed = lipgloss.NewStyle().Foreground(ColorGood)
	// Wrong is a mistyped character, shown on the target text.
	Wrong = lipgloss.NewStyle().Foreground(ColorBad).Underline(true)
	// Pending is target text not yet reached.
	Pending = lipgloss.NewStyle().Foreground(ColorDim)
	// Cursor marks the character to type next.
	Cursor = lipgloss.NewStyle().Foreground(ColorAccent).Background(ColorInverse).Bold(true)

	// Selected is the highlighted row in a menu.
	Selected = lipgloss.NewStyle().Foreground(ColorAccent).Bold(true)
	// Locked is a menu row the player has not earned yet.
	Locked = lipgloss.NewStyle().Foreground(ColorDim)

	// Help is the key hint strip at the bottom of every screen.
	Help = lipgloss.NewStyle().Foreground(ColorDim).MarginTop(1)

	// Panel frames a block of content.
	Panel = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorDim).
		Padding(0, 2)
)

// Status badges. These are four ASCII characters wide on purpose: the columns
// beside them must line up, and the Linux console font is not guaranteed to
// have ✓, ★ or an emoji padlock — a missing glyph shows as a blank or a box.
func BadgeDone() string { return Good.Render("done") }
func BadgeOpen() string { return Warn.Render("open") }
func BadgeLock() string { return Dim.Render("lock") }

// Bar renders a fixed-width progress bar for a 0..1 fraction.
func Bar(width int, frac float64) string {
	if width < 1 {
		return ""
	}
	if frac < 0 {
		frac = 0
	}
	if frac > 1 {
		frac = 1
	}
	filled := int(frac*float64(width) + 0.5)
	out := make([]rune, 0, width)
	for i := 0; i < width; i++ {
		if i < filled {
			out = append(out, '█')
		} else {
			out = append(out, '░')
		}
	}
	return lipgloss.NewStyle().Foreground(ColorAccent).Render(string(out[:filled])) +
		Dim.Render(string(out[filled:]))
}
