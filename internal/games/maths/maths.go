// Package maths is the number game: the number facts to warm up on, and then
// the first nine chapters of the Art of Problem Solving Prealgebra book, a
// chapter at a time, against the clock.
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
func (g *Game) Blurb() string { return "Number facts, then prealgebra, against the clock." }
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
	// modeChapters is the contents page: which chapter of the book.
	modeChapters mode = iota
	// modeLevels is the levels inside one chapter.
	modeLevels
	modePlay
	// modeCorrection holds the right answer on screen after a wrong one. It is
	// the only place the game deliberately slows down: an answer that flashed
	// past would teach nothing.
	modeCorrection
)

type model struct {
	mode mode

	// cleared is how many levels of the flat curriculum have been passed. The
	// two cursors below are a position in the menu; LevelIndex turns them
	// back into an index into that flat list.
	cleared int
	chapter int
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
	case modeChapters:
		return m.updateChapters(key)
	case modeLevels:
		return m.updateLevels(key)
	case modeCorrection:
		return m.updateCorrection(key)
	default:
		return m.updatePlay(key)
	}
}

func (m *model) updateChapters(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		return m, games.Exit()
	case "up", "k":
		if m.chapter > 0 {
			m.chapter--
		}
	case "down", "j":
		if m.chapter < len(Chapters)-1 {
			m.chapter++
		}
	case "enter", " ", "right", "l":
		if !ChapterUnlocked(m.chapter, m.cleared) {
			return m, nil
		}
		// Open on the first level not yet passed, so carrying on where you
		// left off costs no keystrokes. A finished chapter opens on its last
		// level, which is the one worth replaying.
		done := ChapterCleared(m.chapter, m.cleared)
		m.cursor = min(done, len(Chapters[m.chapter].Levels)-1)
		m.mode = modeLevels
	}
	return m, nil
}

func (m *model) updateLevels(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	chapter := Chapters[m.chapter]
	switch msg.String() {
	case "esc", "q", "left", "h":
		m.mode = modeChapters
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(chapter.Levels)-1 {
			m.cursor++
		}
	case "enter", " ":
		if !LevelUnlocked(LevelIndex(m.chapter, m.cursor), m.cleared) {
			return m, nil
		}
		m.rd = newRound(chapter.Levels[m.cursor], m.rng)
		// Wide enough for a fraction like 15/16 or a decimal like 0.0625, and
		// no wider: a field that takes ten characters suggests an answer that
		// needs them.
		m.field = ui.Field{Max: 8, Filter: ui.Number}
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
		// Test done BEFORE going back to play: getting the last question wrong
		// ends the round, and modePlay would render a question that is no
		// longer there.
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
		RoundIndex: LevelIndex(m.chapter, m.cursor),
		Score:      m.rd.score(),
		Points:     m.rd.score(),
		Accuracy:   m.rd.accuracy(),
		Duration:   m.rd.elapsed(),
		Notes:      m.rd.notes(),
	}
}

func (m *model) View() string {
	switch m.mode {
	case modeChapters:
		return m.viewChapters()
	case modeLevels:
		return m.viewLevels()
	case modeCorrection:
		return m.viewCorrection()
	default:
		return m.viewPlay()
	}
}

// badgeFor is the four-character marker in front of a menu row: passed, open
// to play, or still shut.
func badgeFor(done, unlocked bool) string {
	switch {
	case done:
		return ui.BadgeDone()
	case unlocked:
		return ui.BadgeOpen()
	}
	return ui.BadgeLock()
}

// marker is the arrow in front of the row the cursor is on.
func marker(selected bool) string {
	if selected {
		return ">"
	}
	return " "
}

// rowTitle pads a menu row's title to a column and styles it for its state.
func rowTitle(title string, width int, selected, unlocked bool) string {
	padded := fmt.Sprintf("%-*s", width, title)
	switch {
	case selected:
		return ui.Selected.Render(padded)
	case !unlocked:
		return ui.Locked.Render(padded)
	}
	return padded
}

func (m *model) viewChapters() string {
	var b strings.Builder
	b.WriteString(ui.Title.Render("Maths Sprint"))
	b.WriteString("\n")
	b.WriteString(ui.Subtitle.Render("Prealgebra, a chapter at a time. Pass a level to open the next."))
	b.WriteString("\n\n")

	for i, c := range Chapters {
		unlocked := ChapterUnlocked(i, m.cleared)
		done := ChapterCleared(i, m.cleared)

		// The warm-up has no number, so it is indented to keep every title
		// starting in the same column.
		name := "   " + c.Title
		if c.Num > 0 {
			name = fmt.Sprintf("%d. %s", c.Num, c.Title)
		}
		fmt.Fprintf(&b, "%s %s %s %s\n",
			marker(i == m.chapter),
			badgeFor(done == len(c.Levels), unlocked),
			rowTitle(name, 36, i == m.chapter, unlocked),
			ui.Dim.Render(fmt.Sprintf("%d/%d", done, len(c.Levels))))
	}

	b.WriteString("\n")
	if ChapterUnlocked(m.chapter, m.cleared) {
		b.WriteString(ui.Muted.Render(Chapters[m.chapter].Blurb))
	} else {
		b.WriteString(ui.Dim.Render("Finish the chapter above to open this one."))
	}
	b.WriteString(ui.Help.Render("up/down choose \u00b7 enter open \u00b7 esc back"))
	return b.String()
}

func (m *model) viewLevels() string {
	c := Chapters[m.chapter]

	var b strings.Builder
	name := c.Title
	if c.Num > 0 {
		name = fmt.Sprintf("Chapter %d. %s", c.Num, c.Title)
	}
	b.WriteString(ui.Title.Render(name))
	b.WriteString("\n")
	b.WriteString(ui.Subtitle.Render(c.Blurb))
	b.WriteString("\n\n")

	for i, l := range c.Levels {
		flat := LevelIndex(m.chapter, i)
		unlocked := LevelUnlocked(flat, m.cleared)
		fmt.Fprintf(&b, "%s %s %s %s\n",
			marker(i == m.cursor),
			badgeFor(flat < m.cleared, unlocked),
			rowTitle(l.Title, 28, i == m.cursor, unlocked),
			ui.Dim.Render(fmt.Sprintf("%d questions", l.Questions)))
	}

	b.WriteString("\n")
	if LevelUnlocked(LevelIndex(m.chapter, m.cursor), m.cleared) {
		b.WriteString(ui.Muted.Render(c.Levels[m.cursor].Hint))
	} else {
		b.WriteString(ui.Dim.Render("Pass the level above to open this one."))
	}
	b.WriteString(ui.Help.Render("up/down choose \u00b7 enter play \u00b7 esc chapters"))
	return b.String()
}

// askJoin is what sits between a question and its answer. "7 x 8" wants an
// equals sign; "How many x altogether?" is already a sentence and reads badly
// with one.
func askJoin(q Question) string {
	if strings.HasSuffix(q.Prompt, "?") {
		return " "
	}
	return ui.Dim.Render("=")
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

	q := m.rd.current()
	if q.Context != "" {
		fmt.Fprintf(&b, "      %s\n\n", ui.Muted.Render(q.Context))
	}

	// No padding after the field: an answer can be one digit or three, and a
	// row of dots would suggest there are more to type.
	fmt.Fprintf(&b, "      %s %s %s\n", ui.Selected.Render(q.Prompt), askJoin(q), m.field.Render(0))

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

	if m.missedQ.Context != "" {
		fmt.Fprintf(&b, "      %s\n\n", ui.Muted.Render(m.missedQ.Context))
	}
	fmt.Fprintf(&b, "      %s %s %s\n",
		ui.Bad.Render(m.missedQ.Prompt),
		askJoin(m.missedQ),
		ui.Good.Render(m.missedQ.Answer.String()))
	b.WriteString("\n")
	// Say that it is coming back. Otherwise the total on the header line
	// going up by one looks like the game losing count.
	note := "      Not quite. Have a look at that one."
	if m.rd.requeued {
		note = "      Not quite. Have a look -- this one comes round again at the end."
	}
	b.WriteString(ui.Muted.Render(note))

	b.WriteString("\n\n")
	b.WriteString(m.scoreLine())
	b.WriteString(ui.Help.Render("enter carry on"))
	return b.String()
}
