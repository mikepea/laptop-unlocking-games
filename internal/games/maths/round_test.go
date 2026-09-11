package maths

import (
	"math/rand/v2"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/mikepea/laptop-unlocking-games/internal/games"
)

// testLevel yields the predictable questions 1+1, 2+1, 3+1, 4+1, so a test can
// answer them without knowing anything about the real generators.
func testLevel() Level {
	n := 0
	return Level{
		Title:       "Test",
		Questions:   4,
		MinAccuracy: 0.75,
		Gen: func(_ *rand.Rand) Question {
			n++
			return Question{Prompt: fmtQ(n, "+", 1), Answer: Int(n + 1)}
		},
	}
}

func TestDivisionIsAlwaysExact(t *testing.T) {
	// "Sharing Out" builds its dividend by multiplying, so it can never ask
	// for a remainder. Guard that, because it is easy to break.
	r := games.NewTestRand(11)
	sharing := levelIn(t, "Warm-Up", "Sharing Out")
	for i := 0; i < 500; i++ {
		q := sharing.Gen(r)
		parts := strings.Fields(q.Prompt)
		a, _ := strconv.Atoi(parts[0])
		b, _ := strconv.Atoi(parts[2])
		if b == 0 {
			t.Fatal("division by zero generated")
		}
		if a%b != 0 {
			t.Fatalf("%q does not divide exactly", q.Prompt)
		}
	}
}

// levelIn finds a level by chapter and title, so a test does not depend on
// where in the curriculum it happens to sit.
func levelIn(t *testing.T, chapter, title string) Level {
	t.Helper()
	for _, c := range Chapters {
		if c.Title != chapter {
			continue
		}
		for _, l := range c.Levels {
			if l.Title == title {
				return l
			}
		}
	}
	t.Fatalf("no level %q in chapter %q", title, chapter)
	return Level{}
}

func TestSubmitGradesAndAdvances(t *testing.T) {
	rd := newRound(testLevel(), games.NewTestRand(1))

	if !rd.submit("2") { // 1 + 1
		t.Fatal("correct answer graded as wrong")
	}
	if rd.submit("99") { // 2 + 1
		t.Fatal("wrong answer graded as right")
	}
	if rd.asked() != 2 {
		t.Fatalf("asked = %d, want 2", rd.asked())
	}
	if got, want := rd.accuracy(), 0.5; got != want {
		t.Fatalf("accuracy = %v, want %v", got, want)
	}
	if n, _ := rd.missed[0].Answer.Whole(); len(rd.missed) != 1 || n != 3 {
		t.Fatalf("missed = %+v, want the 2 + 1 question", rd.missed)
	}
}

func TestSubmitTreatsRubbishAsWrongNotAsAnError(t *testing.T) {
	rd := newRound(testLevel(), games.NewTestRand(1))

	if rd.submit("") {
		t.Error("an empty answer was graded correct")
	}
	if rd.submit("banana") {
		t.Error("a non-numeric answer was graded correct")
	}
	if rd.asked() != 2 {
		t.Fatalf("asked = %d, want both attempts counted", rd.asked())
	}
}

func TestRoundFinishesAndGradesPass(t *testing.T) {
	rd := newRound(testLevel(), games.NewTestRand(1))
	for i := 1; i <= 4; i++ {
		rd.submit(strconv.Itoa(i + 1))
	}
	if !rd.done {
		t.Fatal("round did not finish after every question")
	}
	if !rd.passed() {
		t.Fatalf("a perfect round did not pass (accuracy %v)", rd.accuracy())
	}
	if rd.notes() != nil {
		t.Fatalf("a perfect round produced notes: %v", rd.notes())
	}
}

func TestUnfinishedRoundNeverPasses(t *testing.T) {
	rd := newRound(testLevel(), games.NewTestRand(1))
	rd.submit("2")
	if rd.passed() {
		t.Fatal("a round abandoned after one question reported as passed")
	}
}

func TestNotesAreCappedAndListAnswers(t *testing.T) {
	l := testLevel()
	l.Questions = 10
	rd := newRound(l, games.NewTestRand(1))
	for i := 0; i < 10; i++ {
		rd.submit("-99")
	}

	notes := rd.notes()
	if len(notes) != 7 {
		t.Fatalf("got %d notes, want 6 plus an 'and more' line: %v", len(notes), notes)
	}
	if !strings.Contains(notes[0], "=") {
		t.Errorf("note %q does not show the answer", notes[0])
	}
	if !strings.Contains(notes[6], "more") {
		t.Errorf("last note %q does not say how many were left out", notes[6])
	}
}

func TestScoreOf(t *testing.T) {
	tests := []struct {
		name           string
		correct, asked int
		elapsed        time.Duration
		want           int
	}{
		{"nothing right scores nothing", 0, 12, time.Minute, 0},
		{"perfect and quick", 12, 12, time.Minute, 72 + 12 + 50},
		{"accuracy pays only above 80%", 8, 10, time.Minute, 48 + 10},
		{"pace is capped", 12, 12, time.Second, 72 + 30 + 50},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := scoreOf(tc.correct, tc.asked, tc.elapsed); got != tc.want {
				t.Fatalf("scoreOf(%d, %d, %v) = %d, want %d", tc.correct, tc.asked, tc.elapsed, got, tc.want)
			}
		})
	}
}

func TestClockStartsOnFirstInput(t *testing.T) {
	rd := newRound(testLevel(), games.NewTestRand(1))
	now := time.Unix(1000, 0)
	rd.now = func() time.Time { return now }

	if rd.elapsed() != 0 {
		t.Fatalf("elapsed before playing = %v, want 0", rd.elapsed())
	}
	rd.start()
	now = now.Add(45 * time.Second)
	if rd.elapsed() != 45*time.Second {
		t.Fatalf("elapsed = %v, want 45s", rd.elapsed())
	}
}

func TestCuratedLevelsAreWellFormed(t *testing.T) {
	seen := map[string]bool{}
	for i, l := range Levels {
		// answerCheckers is keyed by title, and so is every test that looks a
		// level up. Two levels sharing a title would silently leave one of
		// them unchecked.
		if seen[l.Title] {
			t.Errorf("two levels are called %q", l.Title)
		}
		seen[l.Title] = true
		if l.Title == "" {
			t.Errorf("level %d has no title", i)
		}
		if l.Hint == "" {
			t.Errorf("level %q has no hint", l.Title)
		}
		if l.Questions <= 0 {
			t.Errorf("level %q asks %d questions", l.Title, l.Questions)
		}
		if l.MinAccuracy <= 0 || l.MinAccuracy > 1 {
			t.Errorf("level %q has MinAccuracy %v", l.Title, l.MinAccuracy)
		}
		if l.Gen == nil && l.GenSteps == nil {
			t.Errorf("level %q has no generator", l.Title)
		}
		if l.Gen != nil && l.GenSteps != nil {
			t.Errorf("level %q has both a Gen and a GenSteps", l.Title)
		}
	}
}
