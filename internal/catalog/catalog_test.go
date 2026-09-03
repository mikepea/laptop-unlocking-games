package catalog

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mikepea/laptop-unlocking-games/internal/profile"
	"github.com/mikepea/laptop-unlocking-games/internal/unlocks"
)

func TestEveryGameIsGatedByARealStage(t *testing.T) {
	// A typo in Stage() would leave the game permanently visible-but-unlocked
	// with a nonsense cost, which is exactly the sort of thing nobody notices
	// until a child asks why they cannot play it.
	for _, g := range Registry().All() {
		if _, ok := unlocks.ByID(g.Stage()); !ok {
			t.Errorf("game %q is gated by stage %q, which is not on the ladder", g.ID(), g.Stage())
		}
	}
}

func TestGameIdentityIsWellFormed(t *testing.T) {
	seen := map[string]bool{}
	for _, g := range Registry().All() {
		if g.ID() == "" {
			t.Error("a game has an empty ID")
		}
		if seen[g.ID()] {
			t.Errorf("duplicate game ID %q", g.ID())
		}
		seen[g.ID()] = true

		if g.Title() == "" {
			t.Errorf("game %q has no title", g.ID())
		}
		if g.Blurb() == "" {
			t.Errorf("game %q has no blurb", g.ID())
		}
	}
}

func TestEveryGameBuildsAModelAndStarts(t *testing.T) {
	p := profile.New("kiddo")
	for _, g := range Registry().All() {
		m := g.New(p)
		if m == nil {
			t.Fatalf("game %q returned a nil model", g.ID())
		}

		// Every game opens on a chooser, so View must work before any input.
		if m.View() == "" {
			t.Errorf("game %q rendered nothing on open", g.ID())
		}

		// Enter starts the first round in every game. Rendering that screen
		// is a smoke test for the play view: a nil round or a bad index in a
		// format string panics here rather than in front of a child.
		m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		if m.View() == "" {
			t.Errorf("game %q rendered nothing after starting a round", g.ID())
		}
	}
}

func TestOnlyTheFirstStageIsFree(t *testing.T) {
	// Every game past the first should cost something, or the ladder stops
	// being a ladder.
	free := 0
	for _, g := range Registry().All() {
		stage, ok := unlocks.ByID(g.Stage())
		if ok && stage.Cost == 0 {
			free++
		}
	}
	if free != 1 {
		t.Fatalf("%d games are free to play, want exactly 1", free)
	}
}
