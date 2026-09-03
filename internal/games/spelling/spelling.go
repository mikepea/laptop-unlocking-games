// Package spelling is the spelling bee: a word appears, it disappears, and you
// write it down from memory.
package spelling

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mikepea/laptop-unlocking-games/internal/games"
	"github.com/mikepea/laptop-unlocking-games/internal/profile"
	"github.com/mikepea/laptop-unlocking-games/internal/ui"
)

// GameID is the registry key and profile stats key for this game.
const GameID = "spelling"

// Game is the spelling bee.
type Game struct {
	// Rand chooses the words. Nil means seed from the clock.
	Rand *rand.Rand
}

// New returns the spelling game for registration with the launcher.
func New() *Game { return &Game{} }

func (g *Game) ID() string    { return GameID }
func (g *Game) Title() string { return "Spelling Bee" }
func (g *Game) Blurb() string { return "Read it, lose it, write it down." }
func (g *Game) Stage() string { return "spelling" }

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
	// modeShow has the word on screen and the keyboard ignored.
	modeShow
	modeType
	// modeCorrection puts the right spelling next to what was written.
	modeCorrection
)

// tickMsg drives the reveal countdown.
type tickMsg time.Time

// tickEvery is how often the countdown bar redraws while a word is showing.
const tickEvery = 100 * time.Millisecond

func tick() tea.Cmd {
	return tea.Tick(tickEvery, func(t time.Time) tea.Msg { return tickMsg(t) })
}

type model struct {
	mode mode

	cleared int
	cursor  int

	rng   *rand.Rand
	rd    *round
	field ui.Field

	// hideAt is when the current word stops being visible.
	hideAt time.Time
	// showTotal is the full reveal duration, for the countdown bar.
	showTotal time.Duration

	missed missedWord
	streak int
}

func (m *model) Init() tea.Cmd { return nil }

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		if m.mode != modeShow {
			// A tick left over from a word that has already been hidden.
			return m, nil
		}
		if !time.Now().Before(m.hideAt) {
			m.mode = modeType
			return m, nil
		}
		return m, tick()

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		switch m.mode {
		case modeSelect:
			return m.updateSelect(msg)
		case modeShow:
			return m.updateShow(msg)
		case modeCorrection:
			return m.updateCorrection(msg)
		default:
			return m.updateType(msg)
		}
	}
	return m, nil
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
		m.field = ui.Field{Max: 24, Filter: ui.Letters}
		m.rd.start()
		return m, m.reveal()
	}
	return m, nil
}

// reveal shows the current word and starts the countdown to hiding it.
func (m *model) reveal() tea.Cmd {
	m.showTotal = m.rd.showFor()
	m.hideAt = time.Now().Add(m.showTotal)
	m.mode = modeShow
	return tick()
}

func (m *model) updateShow(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		return m, games.Finish(m.result())
	case tea.KeyEnter, tea.KeySpace:
		// Hide it early. Some players would rather have the challenge than
		// the extra second, and letting them choose costs nothing.
		m.mode = modeType
	}
	return m, nil
}

func (m *model) updateType(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		return m, games.Finish(m.result())
	case tea.KeyBackspace:
		m.field.Backspace()
	case tea.KeyEnter:
		if m.field.Empty() {
			return m, nil
		}
		want := m.rd.current()
		got := m.field.Value()
		if m.rd.submit(got) {
			m.streak++
			m.field.Clear()
			if m.rd.done {
				return m, games.Finish(m.result())
			}
			return m, m.reveal()
		}
		m.streak = 0
		m.missed = missedWord{Want: want, Got: got}
		m.mode = modeCorrection
	case tea.KeyRunes:
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
		if m.rd.done {
			return m, games.Finish(m.result())
		}
		return m, m.reveal()
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
	case modeShow:
		return m.viewShow()
	case modeCorrection:
		return m.viewCorrection()
	default:
		return m.viewType()
	}
}

func (m *model) viewSelect() string {
	var b strings.Builder
	b.WriteString(ui.Title.Render("Spelling Bee"))
	b.WriteString("\n")
	b.WriteString(ui.Subtitle.Render("The word appears, then vanishes. Write it down."))
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

		title := fmt.Sprintf("%-22s", l.Title)
		if i == m.cursor {
			title = ui.Selected.Render(title)
		} else if !unlocked {
			title = ui.Locked.Render(title)
		}
		fmt.Fprintf(&b, "%s %s %s %s\n", marker, badge, title,
			ui.Dim.Render(fmt.Sprintf("%d words", l.Count)))
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

func (m *model) header() string {
	return ui.Title.Render(m.rd.level.Title) +
		ui.Dim.Render(fmt.Sprintf("   word %d of %d",
			min(m.rd.idx+1, len(m.rd.words)), len(m.rd.words)))
}

func (m *model) scoreLine() string {
	frac := float64(m.rd.idx) / float64(len(m.rd.words))
	line := fmt.Sprintf("  %s %s   %s",
		ui.Dim.Render("right"),
		ui.Selected.Render(fmt.Sprintf("%d/%d", m.rd.correct, m.rd.idx)),
		ui.Bar(20, frac))
	if m.streak >= 3 {
		line += ui.Gold.Render(fmt.Sprintf("   %d in a row!", m.streak))
	}
	return line
}

func (m *model) viewShow() string {
	var b strings.Builder
	b.WriteString(m.header())
	b.WriteString("\n")
	b.WriteString(ui.Subtitle.Render("Look at it. Really look at it."))
	b.WriteString("\n\n\n")

	fmt.Fprintf(&b, "      %s\n\n", ui.Selected.Render(m.rd.current()))

	// The bar drains rather than fills: it is time running out, not progress.
	remaining := time.Until(m.hideAt)
	frac := 0.0
	if m.showTotal > 0 && remaining > 0 {
		frac = float64(remaining) / float64(m.showTotal)
	}
	fmt.Fprintf(&b, "      %s\n", ui.Bar(20, frac))

	b.WriteString("\n")
	b.WriteString(m.scoreLine())
	b.WriteString(ui.Help.Render("enter hide it now · esc finish early"))
	return b.String()
}

func (m *model) viewType() string {
	var b strings.Builder
	b.WriteString(m.header())
	b.WriteString("\n")
	b.WriteString(ui.Subtitle.Render(m.rd.level.Hint))
	b.WriteString("\n\n\n")

	// Deliberately unpadded: a field sized to the word would give away how
	// many letters it has, which is half the answer.
	fmt.Fprintf(&b, "      %s %s\n", ui.Dim.Render("spell it:"), m.field.Render(0))

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

	fmt.Fprintf(&b, "      %s %s\n", ui.Dim.Render("you wrote "), ui.Bad.Render(m.missed.Got))
	fmt.Fprintf(&b, "      %s %s\n", ui.Dim.Render("it was    "), ui.Good.Render(m.missed.Want))

	b.WriteString("\n")
	b.WriteString(ui.Muted.Render("      Look at where they differ."))

	b.WriteString("\n\n")
	b.WriteString(m.scoreLine())
	b.WriteString(ui.Help.Render("enter carry on"))
	return b.String()
}
