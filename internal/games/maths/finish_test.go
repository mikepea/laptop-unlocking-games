package maths

import (
	"math/rand/v2"
	"strconv"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mikepea/laptop-unlocking-games/internal/profile"
)

// Bubble Tea renders the model returned from Update before the FinishedMsg
// produced by games.Finish is delivered and the launcher swaps this model out.
// So View always runs at least once on a round that is already over, with idx
// one past the last question. Both ways of finishing a round must survive that.

func startLevel0(t *testing.T) *model {
	t.Helper()
	g := &Game{Rand: rand.New(rand.NewPCG(1, 2))}
	m := g.New(&profile.Profile{}).(*model)
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter}) // select screen: play level 0
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
	if got := m.rd.idx; got != len(m.rd.questions) {
		t.Fatalf("idx = %d, want %d", got, len(m.rd.questions))
	}
	_ = m.View()
}

func TestViewAfterFinishingWithAWrongAnswer(t *testing.T) {
	m := startLevel0(t)
	for !m.rd.done {
		if m.rd.idx == len(m.rd.questions)-1 {
			m = answer(t, m, "9999") // last question, deliberately wrong
			break
		}
		m = answer(t, m, strconv.Itoa(m.rd.current().Answer))
	}
	_ = m.View() // correction screen

	// Dismissing the correction must finish the round, not return to play with
	// no question left to show.
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	after := next.(*model)
	if after.mode == modePlay {
		t.Error("returned to modePlay after the round finished")
	}
	_ = after.View()
}

func TestCurrentIsStableOnceTheRoundIsOver(t *testing.T) {
	m := startLevel0(t)
	last := m.rd.questions[len(m.rd.questions)-1]
	for !m.rd.done {
		m = answer(t, m, strconv.Itoa(m.rd.current().Answer))
	}
	if got := m.rd.current(); got != last {
		t.Errorf("current() after finishing = %v, want the last question %v", got, last)
	}
}
