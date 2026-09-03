// Package pseudocode is the trace-the-program game: a few lines of code, and
// the question is what they print.
//
// It is the other half of the algebra in Maths Sprint. "Putting Numbers In"
// asks what 3x + 2 is when x is 4; this asks the same thing with the working
// written down a line at a time, and then keeps going into the places algebra
// cannot follow — a box whose contents change, a line that runs four times, a
// line that does not run at all.
package pseudocode

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
const GameID = "pseudocode"

// Game is the code-tracing puzzle.
type Game struct {
	// Rand is the source of programs. Nil means seed from the clock.
	Rand *rand.Rand
}

// New returns the game for registration with the launcher.
func New() *Game { return &Game{} }

func (g *Game) ID() string    { return GameID }
func (g *Game) Title() string { return "Trace the Code" }
func (g *Game) Blurb() string { return "Read the program. Say what it prints." }
func (g *Game) Stage() string { return "pseudocode" }

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
	// modeCorrection holds the right answer on screen after a wrong one, with
	// the program still visible so the trace can be walked again.
	modeCorrection
)

type model struct {
	mode mode

	cleared int
	cursor  int

	rng   *rand.Rand
	rd    *round
	field ui.Field

	// missedP is the program being corrected in modeCorrection.
	missedP Program
	streak  int
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
		m.field = ui.Field{Max: 5, Filter: ui.Digits}
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
		// An empty answer is a stray Enter, not a wrong answer.
		if m.field.Empty() {
			return m, nil
		}
		p := m.rd.current()
		if m.rd.submit(m.field.Value()) {
			m.streak++
			m.field.Clear()
			if m.rd.done {
				return m, games.Finish(m.result())
			}
			return m, nil
		}
		m.streak = 0
		m.missedP = p
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
		// Test done before going back to play: on the last program the round
		// is already over, and modePlay would render one that is not there.
		if m.rd.done {
			return m, games.Finish(m.result())
		}
		m.mode = modePlay
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
	b.WriteString(ui.Title.Render("Trace the Code"))
	b.WriteString("\n")
	b.WriteString(ui.Subtitle.Render("Read it like the computer does: one line at a time, top to bottom."))
	b.WriteString("\n\n")

	for i, l := range Levels {
		badge := ui.BadgeLock()
		switch {
		case i < m.cleared:
			badge = ui.BadgeDone()
		case LevelUnlocked(i, m.cleared):
			badge = ui.BadgeOpen()
		}
		title := ui.Locked.Render(l.Title)
		if LevelUnlocked(i, m.cleared) {
			title = ui.Selected.Render(l.Title)
		}
		marker := "  "
		if i == m.cursor {
			marker = ui.Selected.Render("> ")
		}
		fmt.Fprintf(&b, "%s%s  %s   %s\n", marker, badge, title,
			ui.Dim.Render(fmt.Sprintf("%d programs", l.Questions)))
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
		ui.Dim.Render(fmt.Sprintf("   program %d of %d",
			min(m.rd.idx+1, len(m.rd.programs)), len(m.rd.programs)))
}

// scoreLine is the running tally and progress bar.
func (m *model) scoreLine() string {
	frac := float64(m.rd.idx) / float64(len(m.rd.programs))
	line := fmt.Sprintf("  %s %s   %s",
		ui.Dim.Render("right"),
		ui.Selected.Render(fmt.Sprintf("%d/%d", m.rd.correct, m.rd.idx)),
		ui.Bar(20, frac))
	if m.streak >= 3 {
		line += ui.Gold.Render(fmt.Sprintf("   %d in a row!", m.streak))
	}
	return line
}

// renderProgram lays the code out as code: fixed left margin, blank line
// either side, indentation preserved so a loop body is visibly inside it.
func renderProgram(p Program, style func(...string) string) string {
	var b strings.Builder
	for _, line := range p.Lines {
		fmt.Fprintf(&b, "      %s\n", style(line))
	}
	return b.String()
}

func (m *model) viewPlay() string {
	var b strings.Builder
	b.WriteString(m.header())
	b.WriteString("\n")
	b.WriteString(ui.Subtitle.Render(m.rd.level.Hint))
	b.WriteString("\n\n")

	b.WriteString(renderProgram(m.rd.current(), ui.Selected.Render))
	b.WriteString("\n")

	fmt.Fprintf(&b, "      %s %s\n", ui.Dim.Render("it prints"), m.field.Render(0))

	b.WriteString("\n")
	b.WriteString(m.scoreLine())
	b.WriteString(ui.Help.Render("enter answer · esc finish early"))
	return b.String()
}

func (m *model) viewCorrection() string {
	var b strings.Builder
	b.WriteString(m.header())
	b.WriteString("\n")
	b.WriteString(ui.Subtitle.Render(m.rd.level.Hint))
	b.WriteString("\n\n")

	// The program stays on screen: the point of the pause is to walk the trace
	// again knowing where it ends up, not just to read a number.
	b.WriteString(renderProgram(m.missedP, ui.Muted.Render))
	b.WriteString("\n")

	fmt.Fprintf(&b, "      %s %s\n",
		ui.Dim.Render("it prints"),
		ui.Good.Render(fmt.Sprintf("%d", m.missedP.Answer)))
	b.WriteString("\n")
	b.WriteString(ui.Muted.Render("      Not quite. Walk it through one line at a time."))

	b.WriteString("\n\n")
	b.WriteString(m.scoreLine())
	b.WriteString(ui.Help.Render("enter carry on"))
	return b.String()
}
