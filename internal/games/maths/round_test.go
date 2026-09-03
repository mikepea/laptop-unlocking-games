package maths

import (
	"errors"
	"math/rand/v2"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/mikepea/laptop-unlocking-games/internal/games"
)

var errBadPrompt = errors.New("prompt is not an arithmetic question")

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
			return Question{Prompt: fmtQ(n, "+", 1), Answer: n + 1}
		},
	}
}

func TestGeneratedQuestionsAgreeWithTheirPrompts(t *testing.T) {
	// The prompt and the answer are produced separately in every generator, so
	// a wrong operator would show one sum and mark another. Re-derive the
	// answer from the prompt and check they match.
	r := games.NewTestRand(7)
	for _, l := range Levels {
		for i := 0; i < 300; i++ {
			q := l.Gen(r)
			want, err := evalPrompt(q.Prompt)
			if err != nil {
				t.Fatalf("level %q produced an unparseable prompt %q: %v", l.Title, q.Prompt, err)
			}
			if want != q.Answer {
				t.Fatalf("level %q: prompt %q says %d, question claims %d", l.Title, q.Prompt, want, q.Answer)
			}
			if q.Answer < 0 {
				t.Fatalf("level %q produced a negative answer: %q = %d", l.Title, q.Prompt, q.Answer)
			}
		}
	}
}

// evalPrompt works out "7 x 8" independently of the code that built it.
func evalPrompt(prompt string) (int, error) {
	parts := strings.Fields(prompt)
	if len(parts) != 3 {
		return 0, errBadPrompt
	}
	a, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}
	b, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, err
	}
	switch parts[1] {
	case "+":
		return a + b, nil
	case "-":
		return a - b, nil
	case "x":
		return a * b, nil
	case "/":
		if b == 0 || a%b != 0 {
			return 0, errBadPrompt
		}
		return a / b, nil
	}
	return 0, errBadPrompt
}

func TestDivisionIsAlwaysExact(t *testing.T) {
	// "Sharing Out" builds its dividend by multiplying, so it can never ask
	// for a remainder. Guard that, because it is easy to break.
	r := games.NewTestRand(11)
	sharing := Levels[5]
	if sharing.Title != "Sharing Out" {
		t.Fatalf("expected Levels[5] to be Sharing Out, got %q", sharing.Title)
	}
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
	if len(rd.missed) != 1 || rd.missed[0].Answer != 3 {
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
		rd.submit("-1")
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
	for i, l := range Levels {
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
		if l.Gen == nil {
			t.Errorf("level %q has no generator", l.Title)
		}
	}
}
