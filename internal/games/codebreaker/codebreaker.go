// Package codebreaker is the deduction game: a hidden code, and feedback on
// each guess telling you how close you were without telling you where.
package codebreaker

import (
	"fmt"
	"math/rand/v2"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mikepea/laptop-unlocking-games/internal/games"
	"github.com/mikepea/laptop-unlocking-games/internal/profile"
	"github.com/mikepea/laptop-unlocking-games/internal/ui"
)

// GameID is the registry key and profile stats key for this game.
const GameID = "codebreaker"

// Game is the code breaker.
type Game struct {
	// Rand picks the secret. Nil means seed from the clock.
	Rand *rand.Rand
}

// New returns the code breaker for registration with the launcher.
func New() *Game { return &Game{} }

func (g *Game) ID() string    { return GameID }
func (g *Game) Title() string { return "Code Breaker" }
func (g *Game) Blurb() string { return "Work out the hidden code. No guessing." }
func (g *Game) Stage() string { return "codebreaker" }

// New builds a session model for one visit to the game.
func (g *Game) New(p *profile.Profile) tea.Model {
	r := g.Rand
	if r == nil {
		r = games.NewRand()
	}
	return &model{cleared: p.Stats(GameID).LessonsCleared, rng: r}
}

type mode int

const (
	modeSelect mode = iota
	modePlay
	// modeOver shows the outcome, and the code when it was not found.
	modeOver
)

type model struct {
	mode mode

	cleared int
	cursor  int

	rng   *rand.Rand
	pz    *puzzle
	field ui.Field
}

func (m *model) Init() tea.Cmd { return nil }

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	if key.Type == tea.KeyCtrlC {
		return m, tea.Quit
	}

	switch m.mode {
	case modeSelect:
		return m.updateSelect(key)
	case modeOver:
		if key.Type == tea.KeyEnter || key.Type == tea.KeyEsc || key.Type == tea.KeySpace {
			return m, games.Finish(m.result())
		}
		return m, nil
	default:
		return m.updatePlay(key)
	}
}

func (m *model) updateSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		return m, games.Exit()
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(Levels)-1 {
			m.cursor++
		}
	case "enter", " ":
		if !LevelUnlocked(m.cursor, m.cleared) {
			return m, nil
		}
		l := Levels[m.cursor]
		m.pz = newPuzzle(l, m.rng)
		m.field = ui.Field{Max: l.Length, Filter: symbolFilter(l.Symbols)}
		m.mode = modePlay
	}
	return m, nil
}

// symbolFilter accepts only the digits this level's codes are made of, so a
// player cannot type a 9 into a 1-to-6 puzzle and wonder why it never matches.
func symbolFilter(symbols int) func(rune) bool {
	return func(r rune) bool { return r >= '1' && r <= rune('0'+symbols) }
}

func (m *model) updatePlay(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		return m, games.Finish(m.result())
	case tea.KeyBackspace:
		m.field.Backspace()
	case tea.KeyEnter:
		// Only a full-length guess can be graded.
		if m.field.Len() != m.pz.level.Length {
			return m, nil
		}
		digits := make([]int, 0, m.field.Len())
		for _, r := range m.field.Value() {
			digits = append(digits, int(r-'0'))
		}
		m.pz.guess(digits)
		m.field.Clear()
		if m.pz.done() {
			m.mode = modeOver
		}
	case tea.KeyRunes:
		m.pz.start()
		for _, r := range msg.Runes {
			m.field.Insert(r)
		}
	}
	return m, nil
}

func (m *model) result() games.Result {
	return games.Result{
		GameID:     GameID,
		Completed:  m.pz.solved,
		Round:      m.pz.level.Title,
		RoundIndex: m.cursor,
		Score:      m.pz.score(),
		Points:     m.pz.score(),
		Accuracy:   m.pz.efficiency(),
		Duration:   m.pz.elapsed(),
		Notes:      m.pz.notes(),
	}
}

func (m *model) View() string {
	switch m.mode {
	case modeSelect:
		return m.viewSelect()
	case modeOver:
		return m.viewOver()
	default:
		return m.viewPlay()
	}
}

func (m *model) viewSelect() string {
	var b strings.Builder
	b.WriteString(ui.Title.Render("Code Breaker"))
	b.WriteString("\n")
	b.WriteString(ui.Subtitle.Render("Crack a level to open the next one."))
	b.WriteString("\n\n")

	for i, l := range Levels {
		unlocked := LevelUnlocked(i, m.cleared)

		marker := " "
		if i == m.cursor {
			marker = ">"
		}
		badge := ui.BadgeLock()
		switch {
		case i < m.cleared:
			badge = ui.BadgeDone()
		case unlocked:
			badge = ui.BadgeOpen()
		}

		title := fmt.Sprintf("%-18s", l.Title)
		if i == m.cursor {
			title = ui.Selected.Render(title)
		} else if !unlocked {
			title = ui.Locked.Render(title)
		}
		fmt.Fprintf(&b, "%s %s %s %s\n", marker, badge, title,
			ui.Dim.Render(fmt.Sprintf("%d digits, 1-%d, %d guesses", l.Length, l.Symbols, l.Guesses)))
	}

	b.WriteString("\n")
	if LevelUnlocked(m.cursor, m.cleared) {
		b.WriteString(ui.Muted.Render(Levels[m.cursor].Hint))
	} else {
		b.WriteString(ui.Dim.Render("Crack the level above to open this one."))
	}
	b.WriteString(ui.Help.Render("up/down choose · enter play · esc back"))
	return b.String()
}

// viewHistory renders the guesses made so far with their feedback.
func (m *model) viewHistory() string {
	var b strings.Builder
	for _, g := range m.pz.history {
		exact := ui.Dim.Render(fmt.Sprintf("%d exact", g.Exact))
		if g.Exact > 0 {
			exact = ui.Good.Render(fmt.Sprintf("%d exact", g.Exact))
		}
		near := ui.Dim.Render(fmt.Sprintf("%d near", g.Near))
		if g.Near > 0 {
			near = ui.Warn.Render(fmt.Sprintf("%d near", g.Near))
		}
		fmt.Fprintf(&b, "    %s    %s  %s\n", ui.Selected.Render(g.String()), exact, near)
	}
	return b.String()
}

// renderGuess draws the guess being typed as one slot per digit, spaced to line
// up exactly with the history above it. Without this the input reads "153"
// while every row above it reads "1 5 3", and comparing them is the game.
func (m *model) renderGuess() string {
	typed := []rune(m.field.Value())

	slots := make([]string, 0, m.pz.level.Length)
	for i := 0; i < m.pz.level.Length; i++ {
		switch {
		case i < len(typed):
			slots = append(slots, ui.Selected.Render(string(typed[i])))
		case i == len(typed):
			slots = append(slots, ui.Cursor.Render(" "))
		default:
			slots = append(slots, ui.Dim.Render("_"))
		}
	}
	return strings.Join(slots, " ")
}

func (m *model) viewPlay() string {
	var b strings.Builder
	b.WriteString(ui.Title.Render(m.pz.level.Title))
	b.WriteString(ui.Dim.Render(fmt.Sprintf("   %d guesses left", m.pz.remaining())))
	b.WriteString("\n")
	b.WriteString(ui.Subtitle.Render(m.pz.level.Hint))
	b.WriteString("\n\n")

	b.WriteString(m.viewHistory())

	// The input line sits where the next history row will appear, so the list
	// grows downward without the cursor jumping about.
	fmt.Fprintf(&b, "  > %s\n", m.renderGuess())

	b.WriteString("\n")
	b.WriteString(ui.Dim.Render("  exact = right digit, right place. near = right digit, wrong place."))
	b.WriteString(ui.Help.Render("type digits · enter guess · esc give up"))
	return b.String()
}

func (m *model) viewOver() string {
	var b strings.Builder
	b.WriteString(ui.Title.Render(m.pz.level.Title))
	b.WriteString("\n\n")

	if m.pz.solved {
		b.WriteString(ui.Good.Render(fmt.Sprintf("  Cracked it in %d %s.",
			m.pz.used(), plural(m.pz.used(), "guess", "guesses"))))
	} else {
		b.WriteString(ui.Bad.Render("  Out of guesses."))
		b.WriteString("\n")
		fmt.Fprintf(&b, "  %s %s", ui.Dim.Render("The code was"),
			ui.Selected.Render(Guess{Digits: m.pz.secret}.String()))
	}
	b.WriteString("\n\n")
	b.WriteString(m.viewHistory())

	b.WriteString(ui.Help.Render("enter continue"))
	return b.String()
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}
