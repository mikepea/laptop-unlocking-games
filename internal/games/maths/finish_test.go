package maths

import (
	"math/rand/v2"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mikepea/laptop-unlocking-games/internal/profile"
)

// Bubble Tea renders the model returned from Update before the FinishedMsg
// produced by games.Finish is delivered and the launcher swaps this model out.
// So View always runs at least once on a round that is already over, with idx
// one past the last question. Both ways of finishing a round must survive that.

// startLevel0 goes in through the menus, so the two-screen walk from the
// contents page to a level is covered as well.
func startLevel0(t *testing.T) *model {
	t.Helper()
	g := &Game{Rand: rand.New(rand.NewPCG(1, 2))}
	m := g.New(&profile.Profile{}).(*model)
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter}) // contents: open the warm-up
	m = next.(*model)
	if m.mode != modeLevels {
		t.Fatalf("enter on the contents page left mode %v", m.mode)
	}
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter}) // levels: play the first
	m = next.(*model)
	if m.mode != modePlay {
		t.Fatalf("enter on the levels page left mode %v", m.mode)
	}
	return m
}

// openLevel starts any level, with the whole curriculum unlocked.
func openLevel(t *testing.T, chapter, level int) *model {
	t.Helper()
	m := &model{
		mode:    modeLevels,
		cleared: len(Levels),
		chapter: chapter,
		cursor:  level,
		rng:     rand.New(rand.NewPCG(7, 9)),
	}
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
		m = answer(t, m, m.rd.current().Answer.String())
	}
	if got := m.rd.idx; got != len(m.rd.questions) {
		t.Fatalf("idx = %d, want %d", got, len(m.rd.questions))
	}
	_ = m.View()
}

func TestViewAfterFinishingWithAWrongAnswer(t *testing.T) {
	m := startLevel0(t)

	// Answer everything wrong. The first pass sends every question round
	// again, so the round ends on the last of the second attempts -- which
	// is the one whose correction screen must not go back to modePlay.
	for {
		m = answer(t, m, "9999")
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

func TestCurrentIsStableOnceTheRoundIsOver(t *testing.T) {
	m := startLevel0(t)
	last := m.rd.questions[len(m.rd.questions)-1]
	for !m.rd.done {
		m = answer(t, m, m.rd.current().Answer.String())
	}
	if got := m.rd.current(); got != last {
		t.Errorf("current() after finishing = %v, want the last question %v", got, last)
	}
}

// TestEveryLevelCanBePlayedToTheEnd walks every level of every chapter by
// typing the answer it prints, and renders each screen on the way. A level
// that cannot be finished, or whose printed answer the field will not accept,
// fails here rather than in front of a child.
func TestEveryLevelCanBePlayedToTheEnd(t *testing.T) {
	for ci, c := range Chapters {
		for li, l := range c.Levels {
			t.Run(l.Title, func(t *testing.T) {
				m := openLevel(t, ci, li)
				if m.mode != modePlay {
					t.Fatalf("level %q did not open", l.Title)
				}
				for !m.rd.done {
					_ = m.View()
					typed := m.rd.current().Answer.String()
					m = answer(t, m, typed)
					if m.mode == modeCorrection {
						t.Fatalf("level %q refused its own answer %q to %q",
							l.Title, typed, m.missedQ.Prompt)
					}
				}
				_ = m.View()
				if !m.rd.passed() {
					t.Errorf("level %q did not pass on a perfect round", l.Title)
				}
				// The flat index the launcher stores has to be this level's.
				if got := m.result().RoundIndex; got != LevelIndex(ci, li) {
					t.Errorf("level %q reports index %d, want %d", l.Title, got, LevelIndex(ci, li))
				}
			})
		}
	}
}
