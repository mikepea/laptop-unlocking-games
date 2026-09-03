// Package launcher is the shell the laptop boots into: a menu of games, a
// record of what has been earned, and the ladder of things still locked.
package launcher

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mikepea/laptop-unlocking-games/internal/achievements"
	"github.com/mikepea/laptop-unlocking-games/internal/games"
	"github.com/mikepea/laptop-unlocking-games/internal/points"
	"github.com/mikepea/laptop-unlocking-games/internal/profile"
	"github.com/mikepea/laptop-unlocking-games/internal/ui"
	"github.com/mikepea/laptop-unlocking-games/internal/unlocks"
	"github.com/mikepea/laptop-unlocking-games/internal/update"
	"github.com/mikepea/laptop-unlocking-games/internal/version"
)

type state int

const (
	stateMenu state = iota
	stateGame
	stateResults
	stateProgress
)

type itemKind int

const (
	itemGame itemKind = iota
	itemProgress
	itemQuit
)

type menuItem struct {
	kind   itemKind
	game   games.Game
	label  string
	blurb  string
	locked bool
	// need is the stage that gates a locked item.
	need unlocks.Stage
}

// Model is the root Bubble Tea model.
type Model struct {
	store   *profile.Store
	prof    *profile.Profile
	reg     *games.Registry
	ledger  points.Backend
	checker *update.Checker

	state  state
	cursor int
	items  []menuItem

	active tea.Model

	// Results of the run just finished.
	lastResult  games.Result
	lastEarned  []achievements.Achievement
	lastUnlocks []unlocks.Stage

	pending *update.Available
	err     error

	width  int
	height int
}

// Options configures a launcher.
type Options struct {
	Store   *profile.Store
	Profile *profile.Profile
	Games   *games.Registry
	Ledger  points.Backend
	Checker *update.Checker
}

// New builds the launcher model.
func New(opts Options) *Model {
	m := &Model{
		store:   opts.Store,
		prof:    opts.Profile,
		reg:     opts.Games,
		ledger:  opts.Ledger,
		checker: opts.Checker,
	}
	// Stages already paid for may not be recorded yet — on a first run, or
	// after new stages are added to the ladder in a later build.
	m.syncUnlocks()
	m.rebuildMenu()
	return m
}

func (m *Model) Init() tea.Cmd {
	if m.checker == nil || !m.checker.Enabled() {
		return nil
	}
	return m.checkForUpdate
}

type updateCheckedMsg struct {
	available *update.Available
	err       error
}

func (m *Model) checkForUpdate() tea.Msg {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	a, err := m.checker.Check(ctx)
	return updateCheckedMsg{available: a, err: err}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		if m.active != nil {
			var cmd tea.Cmd
			m.active, cmd = m.active.Update(msg)
			return m, cmd
		}
		return m, nil

	case updateCheckedMsg:
		// A failed check is not worth interrupting anyone over; the laptop is
		// offline most of the time by design.
		m.pending = msg.available
		return m, nil

	case games.ExitMsg:
		m.active = nil
		m.state = stateMenu
		return m, nil

	case games.FinishedMsg:
		m.active = nil
		m.apply(msg.Result)
		m.state = stateResults
		return m, nil
	}

	if m.state == stateGame && m.active != nil {
		var cmd tea.Cmd
		m.active, cmd = m.active.Update(msg)
		return m, cmd
	}

	if key, ok := msg.(tea.KeyMsg); ok {
		return m.updateKey(key)
	}
	return m, nil
}

func (m *Model) updateKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyCtrlC {
		return m, tea.Quit
	}

	switch m.state {
	case stateResults, stateProgress:
		switch msg.String() {
		case "enter", "esc", " ", "q":
			m.state = stateMenu
		}
		return m, nil
	}

	switch msg.String() {
	case "q", "esc":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.items)-1 {
			m.cursor++
		}
	case "enter", " ":
		return m.activate()
	}
	return m, nil
}

func (m *Model) activate() (tea.Model, tea.Cmd) {
	if m.cursor >= len(m.items) {
		return m, nil
	}
	item := m.items[m.cursor]
	switch item.kind {
	case itemQuit:
		return m, tea.Quit
	case itemProgress:
		m.state = stateProgress
		return m, nil
	case itemGame:
		if item.locked {
			return m, nil
		}
		m.active = item.game.New(m.prof)
		m.state = stateGame
		cmds := []tea.Cmd{m.active.Init()}
		if m.width > 0 {
			// Hand the game the size it missed by not existing yet.
			var cmd tea.Cmd
			m.active, cmd = m.active.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

// apply folds a finished run into the profile: stats, points, achievements and
// stage unlocks, then saves.
func (m *Model) apply(r games.Result) {
	stats := m.prof.Stats(r.GameID)
	stats.Plays++
	stats.SecondsPlayed += int(r.Duration.Seconds())
	stats.LastPlayed = time.Now().UTC()
	if r.Score > stats.BestScore {
		stats.BestScore = r.Score
	}
	if r.WPM > stats.BestWPM {
		stats.BestWPM = r.WPM
	}
	if r.Completed && r.Accuracy > stats.BestAccuracy {
		stats.BestAccuracy = r.Accuracy
	}
	// Rounds are cleared in order, so a pass only advances progress when it is
	// the round the player was actually up to. Replaying an old lesson is
	// still worth points, just not a second step forward.
	if r.Completed && r.RoundIndex == stats.LessonsCleared {
		stats.LessonsCleared++
	}

	if r.Points > 0 && m.ledger != nil {
		award := points.Award{
			GameID: r.GameID,
			Round:  r.Round,
			Points: r.Points,
			Reason: fmt.Sprintf("%s: %s", r.GameID, r.Round),
			At:     time.Now().UTC(),
		}
		if err := m.ledger.Award(context.Background(), award); err != nil {
			m.err = err
		}
	}

	m.lastResult = r
	m.lastEarned = achievements.Evaluate(m.prof, r)
	m.lastUnlocks = m.syncUnlocks()

	if err := m.store.Save(m.prof); err != nil {
		m.err = err
	}
	m.rebuildMenu()
}

// syncUnlocks records every stage the lifetime point total has paid for,
// returning the ones that were not already held.
func (m *Model) syncUnlocks() []unlocks.Stage {
	var newly []unlocks.Stage
	reached, _, _ := unlocks.Reached(m.prof.PointsEarned)
	for _, s := range reached {
		if m.prof.Unlock(s.ID) {
			newly = append(newly, s)
		}
	}
	return newly
}

func (m *Model) rebuildMenu() {
	items := make([]menuItem, 0, len(m.reg.All())+2)
	for _, g := range m.reg.All() {
		stage, known := unlocks.ByID(g.Stage())
		locked := known && !m.prof.IsUnlocked(stage.ID)
		items = append(items, menuItem{
			kind:   itemGame,
			game:   g,
			label:  g.Title(),
			blurb:  g.Blurb(),
			locked: locked,
			need:   stage,
		})
	}
	items = append(items,
		menuItem{kind: itemProgress, label: "Progress", blurb: "Badges, points and what comes next."},
		menuItem{kind: itemQuit, label: "Quit", blurb: "Turn it off."},
	)
	m.items = items
	if m.cursor >= len(m.items) {
		m.cursor = len(m.items) - 1
	}
}

func (m *Model) View() string {
	switch m.state {
	case stateGame:
		if m.active != nil {
			return m.active.View()
		}
	case stateResults:
		return m.viewResults()
	case stateProgress:
		return m.viewProgress()
	}
	return m.viewMenu()
}

func (m *Model) header() string {
	var b strings.Builder
	name := m.prof.Name
	if name == "" {
		name = "player"
	}
	b.WriteString(ui.Title.Render("unlock"))
	b.WriteString(ui.Dim.Render(fmt.Sprintf("  %s's laptop", name)))
	b.WriteString(ui.Dim.Render(fmt.Sprintf("   %s", version.Version)))
	b.WriteString("\n\n")

	_, next, all := unlocks.Reached(m.prof.PointsEarned)
	b.WriteString(fmt.Sprintf("  %s %s",
		ui.Selected.Render(fmt.Sprintf("%d", m.prof.PointsEarned)),
		ui.Muted.Render("points"),
	))
	if !all {
		prev := 0
		for _, s := range unlocks.Ladder {
			if s.Cost < next.Cost {
				prev = s.Cost
			}
		}
		span := next.Cost - prev
		frac := 0.0
		if span > 0 {
			frac = float64(m.prof.PointsEarned-prev) / float64(span)
		}
		b.WriteString(ui.Dim.Render(fmt.Sprintf("   next: %s in %d", next.Title, next.Cost-m.prof.PointsEarned)))
		b.WriteString("\n  ")
		b.WriteString(ui.Bar(28, frac))
	} else {
		b.WriteString(ui.Gold.Render("   the whole ladder is yours"))
	}
	b.WriteString("\n")
	return b.String()
}

func (m *Model) viewMenu() string {
	var b strings.Builder
	b.WriteString(m.header())
	b.WriteString("\n")

	for i, item := range m.items {
		marker := " "
		if i == m.cursor {
			marker = ">"
		}
		label := fmt.Sprintf("  %s %-20s", marker, item.label)
		switch {
		case item.locked:
			b.WriteString(ui.Locked.Render(label + fmt.Sprintf("locked until %d points", item.need.Cost)))
		case i == m.cursor:
			b.WriteString(ui.Selected.Render(label) + ui.Muted.Render(item.blurb))
		default:
			b.WriteString(label + ui.Dim.Render(item.blurb))
		}
		b.WriteString("\n")
	}

	if m.pending != nil {
		b.WriteString("\n")
		b.WriteString(ui.Warn.Render(fmt.Sprintf("  An update is ready: %s", m.pending.Version)))
		b.WriteString("\n")
	}
	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(ui.Bad.Render(fmt.Sprintf("  %v", m.err)))
		b.WriteString("\n")
	}

	b.WriteString(ui.Help.Render("  up/down move · enter choose · q quit"))
	return b.String()
}

func (m *Model) viewResults() string {
	r := m.lastResult

	var b strings.Builder
	if r.Completed {
		b.WriteString(ui.Good.Render(fmt.Sprintf("  %s: passed!", r.Round)))
	} else {
		b.WriteString(ui.Warn.Render(fmt.Sprintf("  %s: stopped early", r.Round)))
	}
	b.WriteString("\n\n")

	row := func(k, v string) {
		b.WriteString(fmt.Sprintf("  %s %s\n", ui.Dim.Render(fmt.Sprintf("%-10s", k)), v))
	}
	row("words/min", ui.Selected.Render(fmt.Sprintf("%.0f", r.WPM)))
	row("accuracy", ui.Selected.Render(fmt.Sprintf("%.0f%%", r.Accuracy*100)))
	row("time", ui.Selected.Render(r.Duration.Round(time.Second).String()))
	row("points", ui.Gold.Render(fmt.Sprintf("+%d", r.Points)))

	if len(m.lastEarned) > 0 {
		b.WriteString("\n")
		b.WriteString(ui.Gold.Render("  New badges"))
		b.WriteString("\n")
		for _, a := range m.lastEarned {
			b.WriteString(fmt.Sprintf("  %s %s %s\n", ui.Gold.Render("*"), a.Title, ui.Dim.Render(a.Description)))
		}
	}

	if len(m.lastUnlocks) > 0 {
		b.WriteString("\n")
		b.WriteString(ui.Gold.Render("  Unlocked"))
		b.WriteString("\n")
		for _, s := range m.lastUnlocks {
			b.WriteString(fmt.Sprintf("  %s %s %s\n", ui.Gold.Render("+"), s.Title, ui.Dim.Render(s.Blurb)))
		}
	}

	b.WriteString(ui.Help.Render("  enter continue"))
	return b.String()
}

func (m *Model) viewProgress() string {
	var b strings.Builder
	b.WriteString(m.header())
	b.WriteString("\n")

	b.WriteString(ui.Title.Render("  The Ladder"))
	b.WriteString("\n")
	for _, s := range unlocks.Ladder {
		if m.prof.IsUnlocked(s.ID) {
			fmt.Fprintf(&b, "  %s %-18s %s\n", ui.BadgeDone(), s.Title, ui.Dim.Render(s.Blurb))
			continue
		}
		fmt.Fprintf(&b, "  %s %s\n", ui.BadgeLock(),
			ui.Locked.Render(fmt.Sprintf("%-18s %d points", s.Title, s.Cost)))
	}

	b.WriteString("\n")
	b.WriteString(ui.Title.Render("  Badges"))
	b.WriteString(ui.Dim.Render(fmt.Sprintf("  %d of %d", len(m.prof.Achievements), len(achievements.All))))
	b.WriteString("\n")
	for _, a := range achievements.All {
		if m.prof.HasAchievement(a.ID) {
			fmt.Fprintf(&b, "  %s %-18s %s\n", ui.Gold.Render("*"), a.Title, ui.Dim.Render(a.Description))
			continue
		}
		b.WriteString(ui.Locked.Render(fmt.Sprintf("  . %-18s %s\n", a.Title, a.Description)))
	}

	b.WriteString(ui.Help.Render("  esc back"))
	return b.String()
}
