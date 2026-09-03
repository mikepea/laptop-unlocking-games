// Package maths is the number-facts game: addition, subtraction, times tables
// and division, against the clock.
package maths

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
const GameID = "maths"

// Game is the maths sprint.
type Game struct {
	// Rand is the source of questions. Nil means seed from the clock.
	Rand *rand.Rand
}

// New returns the maths game for registration with the launcher.
func New() *Game { return &Game{} }

func (g *Game) ID() string    { return GameID }
func (g *Game) Title() string { return "Maths Sprint" }
func (g *Game) Blurb() string { return "Number facts against the clock." }
func (g *Game) Stage() string { return "arcade" }

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
	// modeCorrection holds the right answer on screen after a wrong one. It is
	// the only place the game deliberately slows down: an answer that flashed
	// past would teach nothing.
	modeCorrection
)

type model struct {
	mode mode

	cleared int
	cursor  int

	rng   *rand.Rand
	rd    *round
	field ui.Field

	// missedQ is the question being corrected in modeCorrection.
	missedQ Question
	// streak is consecutive right answers, for a bit of encouragement.
	streak int
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
	case modeCorrection:
		return m.updateCorrection(key)
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
		m.rd = newRound(Levels[m.cursor], m.rng)
		m.field = ui.Field{Max: 4, Filter: ui.Digits}
		m.mode = modePlay
	}
	return m, nil
}

func (m *model) updatePlay(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		return m, games.Finish(m.result())
	case tea.KeyBackspace:
		m.field.Backspace()
	case tea.KeyEnter:
		// An empty answer is not submitted: it is almost always a stray Enter,
		// and marking it wrong would be a cruel way to find that out.
		if m.field.Empty() {
			return m, nil
		}
		q := m.rd.current()
		if m.rd.submit(m.field.Value()) {
			m.streak++
			m.field.Clear()
			if m.rd.done {
				return m, games.Finish(m.result())
			}
			return m, nil
		}
		m.streak = 0
		m.missedQ = q
		m.mode = modeCorrection
	case tea.KeyRunes:
		m.rd.start()
		for _, r := range msg.Runes {
			m.field.Insert(r)
		}
	}
	return m, nil
}

func (m *model) updateCorrection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		return m, games.Finish(m.result())
	case tea.KeyEnter, tea.KeySpace:
		m.field.Clear()
		m.mode = modePlay
		if m.rd.done {
			return m, games.Finish(m.result())
		}
	}
	return m, nil
}

func (m *model) result() games.Result {
	return games.Result{
		GameID:     GameID,
		Completed:  m.rd.passed(),
		Round:      m.rd.level.Title,
		RoundIndex: m.cursor,
		Score:      m.rd.score(),
		Points:     m.rd.score(),
		Accuracy:   m.rd.accuracy(),
		Duration:   m.rd.elapsed(),
		Notes:      m.rd.notes(),
	}
}

func (m *model) View() string {
	switch m.mode {
	case modeSelect:
		return m.viewSelect()
	case modeCorrection:
		return m.viewCorrection()
	default:
		return m.viewPlay()
	}
}

func (m *model) viewSelect() string {
	var b strings.Builder
	b.WriteString(ui.Title.Render("Maths Sprint"))
	b.WriteString("\n")
	b.WriteString(ui.Subtitle.Render("Pass a level to open the next one."))
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

		title := fmt.Sprintf("%-26s", l.Title)
		if i == m.cursor {
			title = ui.Selected.Render(title)
		} else if !unlocked {
			title = ui.Locked.Render(title)
		}
		fmt.Fprintf(&b, "%s %s %s %s\n", marker, badge, title,
			ui.Dim.Render(fmt.Sprintf("%d questions", l.Questions)))
	}

	b.WriteString("\n")
	if LevelUnlocked(m.cursor, m.cleared) {
		b.WriteString(ui.Muted.Render(Levels[m.cursor].Hint))
	} else {
		b.WriteString(ui.Dim.Render("Pass the level above to open this one."))
	}
	b.WriteString(ui.Help.Render("up/down choose · enter play · esc back"))
	return b.String()
}

// header is the line every in-round screen starts with.
func (m *model) header() string {
	return ui.Title.Render(m.rd.level.Title) +
		ui.Dim.Render(fmt.Sprintf("   question %d of %d",
			min(m.rd.idx+1, len(m.rd.questions)), len(m.rd.questions)))
}

// scoreLine is the running tally and progress bar.
func (m *model) scoreLine() string {
	frac := float64(m.rd.idx) / float64(len(m.rd.questions))
	line := fmt.Sprintf("  %s %s   %s",
		ui.Dim.Render("right"),
		ui.Selected.Render(fmt.Sprintf("%d/%d", m.rd.correct, m.rd.idx)),
		ui.Bar(20, frac))
	if m.streak >= 3 {
		line += ui.Gold.Render(fmt.Sprintf("   %d in a row!", m.streak))
	}
	return line
}

func (m *model) viewPlay() string {
	var b strings.Builder
	b.WriteString(m.header())
	b.WriteString("\n")
	b.WriteString(ui.Subtitle.Render(m.rd.level.Hint))
	b.WriteString("\n\n\n")

	// No padding after the field: an answer can be one digit or three, and a
	// row of dots would suggest there are more to type.
	fmt.Fprintf(&b, "      %s %s %s\n",
		ui.Selected.Render(m.rd.current().Prompt),
		ui.Dim.Render("="),
		m.field.Render(0))

	b.WriteString("\n\n")
	b.WriteString(m.scoreLine())
	b.WriteString(ui.Help.Render("enter answer · esc finish early"))
	return b.String()
}

func (m *model) viewCorrection() string {
	var b strings.Builder
	b.WriteString(m.header())
	b.WriteString("\n")
	b.WriteString(ui.Subtitle.Render(m.rd.level.Hint))
	b.WriteString("\n\n\n")

	fmt.Fprintf(&b, "      %s %s %s\n",
		ui.Bad.Render(m.missedQ.Prompt),
		ui.Dim.Render("="),
		ui.Good.Render(fmt.Sprintf("%d", m.missedQ.Answer)))
	b.WriteString("\n")
	b.WriteString(ui.Muted.Render("      Not quite. Have a look at that one."))

	b.WriteString("\n\n")
	b.WriteString(m.scoreLine())
	b.WriteString(ui.Help.Render("enter carry on"))
	return b.String()
}
