// Package shellquest is the game that teaches the command line: a pretend
// filesystem, a handful of real commands, and a word hidden somewhere in it.
//
// It sits directly below the "shell" rung of the ladder on purpose. By the time
// a player is handed a real prompt, ls, cd, cat and pwd should already be
// muscle memory.
package shellquest

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mikepea/laptop-unlocking-games/internal/games"
	"github.com/mikepea/laptop-unlocking-games/internal/profile"
	"github.com/mikepea/laptop-unlocking-games/internal/ui"
)

// GameID is the registry key and profile stats key for this game.
const GameID = "shellquest"

// scrollback is how many output lines are kept on screen.
const scrollback = 14

// Game is the shell quest.
type Game struct{}

// New returns the shell quest for registration with the launcher.
func New() *Game { return &Game{} }

func (g *Game) ID() string    { return GameID }
func (g *Game) Title() string { return "Shell Quest" }
func (g *Game) Blurb() string { return "Find what is hidden. Learn the commands." }
func (g *Game) Stage() string { return "quest" }

// New builds a session model for one visit to the game.
func (g *Game) New(p *profile.Profile) tea.Model {
	return &model{cleared: p.Stats(GameID).LessonsCleared}
}

type mode int

const (
	modeSelect mode = iota
	modePlay
	modeSolved
)

type model struct {
	mode mode

	cleared int
	cursor  int

	sh    *shell
	field ui.Field
	lines []string
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
	case modeSolved:
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
		if m.cursor < len(Quests)-1 {
			m.cursor++
		}
	case "enter", " ":
		if !QuestUnlocked(m.cursor, m.cleared) {
			return m, nil
		}
		q := Quests[m.cursor]
		m.sh = newShell(q)
		// No filter: this is a command line, and typing the wrong thing into
		// it and reading the error is exactly the skill.
		m.field = ui.Field{Max: 60}
		m.lines = []string{
			ui.Muted.Render(q.Brief),
			"",
			ui.Dim.Render("Type help if you get stuck."),
		}
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
	case tea.KeySpace:
		m.field.Insert(' ')
	case tea.KeyEnter:
		input := m.field.Value()
		m.field.Clear()
		if strings.TrimSpace(input) == "clear" {
			m.sh.run(input)
			m.lines = nil
			return m, nil
		}
		// Echo the command with its prompt, so the transcript reads back like
		// a real session.
		m.append(ui.Dim.Render(m.sh.prompt()) + input)
		m.append(m.sh.run(input)...)
		if m.sh.solved {
			m.mode = modeSolved
		}
	case tea.KeyRunes:
		for _, r := range msg.Runes {
			m.field.Insert(r)
		}
	}
	return m, nil
}

// append adds output lines, trimming the scrollback from the top.
func (m *model) append(lines ...string) {
	m.lines = append(m.lines, lines...)
	if len(m.lines) > scrollback {
		m.lines = m.lines[len(m.lines)-scrollback:]
	}
}

func (m *model) result() games.Result {
	return games.Result{
		GameID:     GameID,
		Completed:  m.sh.solved,
		Round:      m.sh.quest.Title,
		RoundIndex: m.cursor,
		Score:      m.sh.score(),
		Points:     m.sh.score(),
		Accuracy:   m.sh.efficiency(),
		Duration:   m.sh.elapsed(),
		Notes:      m.sh.notes(),
	}
}

func (m *model) View() string {
	switch m.mode {
	case modeSelect:
		return m.viewSelect()
	case modeSolved:
		return m.viewSolved()
	default:
		return m.viewPlay()
	}
}

func (m *model) viewSelect() string {
	var b strings.Builder
	b.WriteString(ui.Title.Render("Shell Quest"))
	b.WriteString("\n")
	b.WriteString(ui.Subtitle.Render("A real command line, with something hidden in it."))
	b.WriteString("\n\n")

	for i, q := range Quests {
		unlocked := QuestUnlocked(i, m.cleared)

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

		title := fmt.Sprintf("%-20s", q.Title)
		if i == m.cursor {
			title = ui.Selected.Render(title)
		} else if !unlocked {
			title = ui.Locked.Render(title)
		}
		fmt.Fprintf(&b, "%s %s %s\n", marker, badge, title)
	}

	b.WriteString("\n")
	if QuestUnlocked(m.cursor, m.cleared) {
		b.WriteString(ui.Muted.Render(Quests[m.cursor].Hint))
	} else {
		b.WriteString(ui.Dim.Render("Finish the quest above to open this one."))
	}
	b.WriteString(ui.Help.Render("up/down choose · enter play · esc back"))
	return b.String()
}

func (m *model) viewPlay() string {
	var b strings.Builder
	b.WriteString(ui.Title.Render(m.sh.quest.Title))
	b.WriteString(ui.Dim.Render(fmt.Sprintf("   %d commands", m.sh.commands)))
	b.WriteString("\n\n")

	for _, l := range m.lines {
		b.WriteString("  " + l + "\n")
	}

	fmt.Fprintf(&b, "  %s%s\n", ui.Dim.Render(m.sh.prompt()), m.field.Render(0))

	b.WriteString(ui.Help.Render("  " + m.sh.quest.Hint + "\n  esc give up"))
	return b.String()
}

func (m *model) viewSolved() string {
	var b strings.Builder
	b.WriteString(ui.Title.Render(m.sh.quest.Title))
	b.WriteString("\n\n")
	b.WriteString(ui.Good.Render(fmt.Sprintf("  Found it, in %d commands.", m.sh.commands)))
	b.WriteString("\n")
	if m.sh.commands <= m.sh.quest.Par {
		b.WriteString(ui.Gold.Render(fmt.Sprintf("  Par was %d. Nicely done.", m.sh.quest.Par)))
	} else {
		b.WriteString(ui.Dim.Render(fmt.Sprintf("  Par is %d, if you want to try it again.", m.sh.quest.Par)))
	}
	b.WriteString("\n\n")
	b.WriteString(ui.Muted.Render("  Everything you just used is real. It all works on the"))
	b.WriteString("\n")
	b.WriteString(ui.Muted.Render("  actual machine, once you get there."))
	b.WriteString(ui.Help.Render("  enter continue"))
	return b.String()
}
