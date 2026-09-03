// Package games defines the contract every aido game implements and the
// registry the launcher reads to build its menu.
package games

import (
	"sort"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mikepea/aido-linux-unlocker/internal/profile"
)

// Result is what a game reports back to the launcher when the player leaves it.
// A game that was abandoned still reports whatever was achieved up to that
// point, with Completed false.
type Result struct {
	GameID    string
	Completed bool

	// Round names what was just played — a lesson, a level, a wave. Shown on
	// the results screen and used for achievement rules.
	Round      string
	RoundIndex int

	Score    int
	Points   int
	Accuracy float64 // 0..1
	WPM      float64
	Duration time.Duration
}

// FinishedMsg is sent by a game to hand control back to the launcher.
type FinishedMsg struct{ Result Result }

// Finish is the command a game returns when it is done.
func Finish(r Result) tea.Cmd {
	return func() tea.Msg { return FinishedMsg{Result: r} }
}

// ExitMsg is sent by a game the player backed out of without playing. The
// launcher returns straight to the menu with no results screen.
type ExitMsg struct{}

// Exit is the command a game returns when the player leaves without playing.
func Exit() tea.Cmd {
	return func() tea.Msg { return ExitMsg{} }
}

// Game is a playable item in the launcher menu.
//
// New returns a fresh tea.Model for one session. The model is handed the
// player's profile so it can gate its own internal progression (which lessons
// or levels are open); it must not write to the profile — the launcher owns
// persistence and applies whatever comes back in the Result.
type Game interface {
	ID() string
	Title() string
	Blurb() string

	// Stage is the unlocks.Stage ID that gates this game. Games at a stage the
	// player has not reached are shown in the menu, locked, so there is always
	// something visible to aim at.
	Stage() string

	New(p *profile.Profile) tea.Model
}

// Registry is the set of games this build knows about.
type Registry struct {
	games []Game
}

// NewRegistry returns a registry holding the given games.
func NewRegistry(games ...Game) *Registry {
	r := &Registry{}
	for _, g := range games {
		r.Register(g)
	}
	return r
}

// Register adds a game, replacing any existing game with the same ID.
func (r *Registry) Register(g Game) {
	for i, existing := range r.games {
		if existing.ID() == g.ID() {
			r.games[i] = g
			return
		}
	}
	r.games = append(r.games, g)
}

// All returns every registered game, ordered by title.
func (r *Registry) All() []Game {
	out := make([]Game, len(r.games))
	copy(out, r.games)
	sort.Slice(out, func(i, j int) bool { return out[i].Title() < out[j].Title() })
	return out
}

// ByID looks a game up.
func (r *Registry) ByID(id string) (Game, bool) {
	for _, g := range r.games {
		if g.ID() == id {
			return g, true
		}
	}
	return nil, false
}
