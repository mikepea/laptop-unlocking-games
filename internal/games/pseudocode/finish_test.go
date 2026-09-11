package pseudocode

import (
	"math/rand/v2"
	"strconv"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mikepea/laptop-unlocking-games/internal/profile"
)

// The launcher renders the model returned from Update before the FinishedMsg
// is delivered and it is swapped out, so View always runs once on a round that
// is already over. This is the bug that took the maths game down; guard it
// here rather than find out the same way twice.

func startLevel0(t *testing.T) *model {
	t.Helper()
	g := &Game{Rand: rand.New(rand.NewPCG(3, 4))}
	m := g.New(&profile.Profile{}).(*model)
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	return next.(*model)
}

func answer(t *testing.T, m *model, s string) *model {
	t.Helper()
	for _, r := range s {
		next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = next.(*model)
	}
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	return next.(*model)
}

func TestViewAfterFinishingWithACorrectAnswer(t *testing.T) {
	m := startLevel0(t)
	for !m.rd.done {
		m = answer(t, m, strconv.Itoa(m.rd.current().Answer))
	}
	_ = m.View()
}

func TestViewAfterFinishingWithAWrongAnswer(t *testing.T) {
	m := startLevel0(t)

	// Answer everything wrong. The first pass sends every program round
	// again, so the round ends on the last of the second attempts -- which
	// is the one whose correction screen must not go back to modePlay.
	for {
		m = answer(t, m, "99999")
		if m.mode != modeCorrection {
			t.Fatalf("a wrong answer did not stop at the correction screen")
		}
		_ = m.View()

		finished := m.rd.done
		next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		m = next.(*model)
		if finished {
			break
		}
		if m.mode != modePlay {
			t.Fatalf("dismissing a correction mid-round left mode %v", m.mode)
		}
	}

	if m.mode == modePlay {
		t.Error("returned to modePlay after the round finished")
	}
	_ = m.View()
}
