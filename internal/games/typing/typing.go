// Package typing is the first game: a lesson-based typing trainer. It is the
// only thing on the laptop at the start, and the source of the points that open
// everything else.
package typing

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mikepea/laptop-unlocking-games/internal/games"
	"github.com/mikepea/laptop-unlocking-games/internal/profile"
	"github.com/mikepea/laptop-unlocking-games/internal/ui"
)

// GameID is the registry key and profile stats key for this game.
const GameID = "typing"

// Game is the typing trainer.
type Game struct{}

// New returns the typing game for registration with the launcher.
func New() *Game { return &Game{} }

func (g *Game) ID() string    { return GameID }
func (g *Game) Title() string { return "Typing Trainer" }
func (g *Game) Blurb() string { return "Learn the keys. Every point starts here." }
func (g *Game) Stage() string { return "typing" }

// NewModel builds a session model for one visit to the game.
func (g *Game) New(p *profile.Profile) tea.Model {
	return &model{cleared: p.Stats(GameID).LessonsCleared}
}

type mode int

const (
	modeSelect mode = iota
	modePlay
)

type model struct {
	mode mode

	// cleared is how many lessons the player has passed, which is also the
	// index of the first locked lesson.
	cleared int
	cursor  int

	sess *session

	width int
}

func (m *model) Init() tea.Cmd { return nil }

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		if m.mode == modeSelect {
			return m.updateSelect(msg)
		}
		return m.updatePlay(msg)
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
		if m.cursor < len(Lessons)-1 {
			m.cursor++
		}
	case "enter", " ":
		if !LessonUnlocked(m.cursor, m.cleared) {
			return m, nil
		}
		m.sess = newSession(Lessons[m.cursor])
		m.mode = modePlay
	}
	return m, nil
}

func (m *model) updatePlay(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		// Leaving early still banks the lines that were finished.
		return m, games.Finish(m.result())
	case tea.KeyBackspace:
		m.sess.backspace()
		return m, nil
	case tea.KeySpace:
		m.sess.typeRune(' ')
	case tea.KeyRunes:
		for _, r := range msg.Runes {
			m.sess.typeRune(r)
		}
	default:
		// Arrows, function keys and the like are not input here.
		return m, nil
	}

	if m.sess.lineComplete() {
		if done := m.sess.advance(); done {
			return m, games.Finish(m.result())
		}
	}
	return m, nil
}

// result packages the session for the launcher.
func (m *model) result() games.Result {
	s := m.sess
	return games.Result{
		GameID:     GameID,
		Completed:  s.passed(),
		Round:      s.lesson.Title,
		RoundIndex: m.cursor,
		Score:      s.score(),
		Points:     s.score(),
		Accuracy:   s.accuracy(),
		WPM:        s.wpm(),
		Duration:   s.elapsed(),
	}
}

func (m *model) View() string {
	if m.mode == modeSelect {
		return m.viewSelect()
	}
	return m.viewPlay()
}

func (m *model) viewSelect() string {
	var b strings.Builder
	b.WriteString(ui.Title.Render("Typing Trainer"))
	b.WriteString("\n")
	b.WriteString(ui.Subtitle.Render("Pass a lesson to open the next one."))
	b.WriteString("\n\n")

	for i, l := range Lessons {
		unlocked := LessonUnlocked(i, m.cleared)

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

		// The badge is styled separately, so only the plain text is padded —
		// pad the title here rather than formatting the whole row at once.
		title := fmt.Sprintf("%-22s", l.Title)
		if i == m.cursor {
			title = ui.Selected.Render(title)
		} else if !unlocked {
			title = ui.Locked.Render(title)
		}

		fmt.Fprintf(&b, "%s %s %s %s\n", marker, badge, title,
			ui.Dim.Render(fmt.Sprintf("%d lines", len(l.Lines))))
	}

	b.WriteString("\n")
	if l := Lessons[m.cursor]; LessonUnlocked(m.cursor, m.cleared) {
		b.WriteString(ui.Muted.Render(l.Hint))
	} else {
		b.WriteString(ui.Dim.Render("Pass the lesson above to open this one."))
	}

	b.WriteString(ui.Help.Render("up/down choose · enter play · esc back"))
	return b.String()
}

func (m *model) viewPlay() string {
	s := m.sess

	var b strings.Builder
	header := lipgloss.JoinHorizontal(lipgloss.Top,
		ui.Title.Render(s.lesson.Title),
		ui.Dim.Render(fmt.Sprintf("   line %d of %d", s.lineIdx+1, len(s.lesson.Lines))),
	)
	b.WriteString(header)
	b.WriteString("\n")
	b.WriteString(ui.Subtitle.Render(s.lesson.Hint))
	b.WriteString("\n\n")

	b.WriteString("  " + renderLine(s.target, s.typed))
	b.WriteString("\n\n")

	// Reserve the nudge line whether or not it has anything to say, so the
	// stats below never jump up and down mid-lesson.
	if len(s.typed) == len(s.target) && !s.lineClean() {
		b.WriteString(ui.Bad.Render("  Fix the red letters: backspace to go back."))
	}
	b.WriteString("\n\n")

	acc := s.accuracy()
	accStyle := ui.Good
	if acc < s.lesson.MinAccuracy {
		accStyle = ui.Warn
	}
	// The bar counts the line in progress as a fraction, so it creeps forward
	// as the player types rather than jumping once a line lands.
	progress := (float64(s.lineIdx) + float64(len(s.typed))/float64(len(s.target))) /
		float64(len(s.lesson.Lines))
	stats := fmt.Sprintf("  %s  %s   %s  %s   %s",
		ui.Dim.Render("wpm"), ui.Selected.Render(fmt.Sprintf("%3.0f", s.wpm())),
		ui.Dim.Render("accuracy"), accStyle.Render(fmt.Sprintf("%3.0f%%", acc*100)),
		ui.Bar(20, progress),
	)
	b.WriteString(stats)

	b.WriteString(ui.Help.Render("esc finish early · backspace fix"))
	return b.String()
}

// renderLine colours the target text against what has been typed: green for
// right, red for wrong, a block cursor on the next character, dim ahead.
func renderLine(target, typed []rune) string {
	var b strings.Builder
	for i, ch := range target {
		switch {
		case i < len(typed) && typed[i] == ch:
			b.WriteString(ui.Typed.Render(string(ch)))
		case i < len(typed):
			b.WriteString(ui.Wrong.Render(visible(ch)))
		case i == len(typed):
			b.WriteString(ui.Cursor.Render(visible(ch)))
		default:
			b.WriteString(ui.Pending.Render(string(ch)))
		}
	}
	return b.String()
}

// visible substitutes a mark for a space so the cursor and error highlights are
// not invisible. Underscore rather than ␣ because the Linux console font is not
// guaranteed to have the fancier glyph.
func visible(ch rune) string {
	if ch == ' ' {
		return "_"
	}
	return string(ch)
}
