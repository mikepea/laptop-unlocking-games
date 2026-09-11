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

// A wrong answer sends the question to the back of the queue for one more go.
// The point is that the correction screen has just shown the right answer, so
// the second attempt is a chance to use what was read there.

func TestAWrongAnswerSendsTheQuestionRound(t *testing.T) {
	rd := newRound(testLevel(), games.NewTestRand(1)) // 1+1, 2+1, 3+1, 4+1

	rd.submit("99") // 1 + 1, wrong
	if !rd.requeued {
		t.Fatal("a wrong answer on the first pass was not sent round again")
	}
	if len(rd.questions) != 5 {
		t.Fatalf("round has %d questions, want the missed one added back", len(rd.questions))
	}
	if got, want := rd.questions[4].Prompt, "1 + 1"; got != want {
		t.Errorf("question sent round is %q, want %q", got, want)
	}

	// Everything else right, so the only thing left is the second attempt.
	for i := 2; i <= 4; i++ {
		if !rd.submit(strconv.Itoa(i + 1)) {
			t.Fatalf("question %d graded wrong", i)
		}
	}
	if rd.done {
		t.Fatal("round finished without asking the missed question again")
	}
	if got := rd.current().Prompt; got != "1 + 1" {
		t.Fatalf("after the first pass the question is %q, want the missed one", got)
	}
	if !rd.submit("2") {
		t.Fatal("the second attempt was graded wrong")
	}
	if !rd.done {
		t.Fatal("round did not finish after the second attempt")
	}
	// Four right out of five asked: the second go counts, and so does the
	// fact that it took two.
	if got, want := rd.accuracy(), 0.8; got != want {
		t.Errorf("accuracy = %v, want %v", got, want)
	}
}

func TestAQuestionIsOnlySentRoundOnce(t *testing.T) {
	rd := newRound(testLevel(), games.NewTestRand(1))
	for i := 0; i < 4; i++ {
		rd.submit("99") // every first attempt wrong
	}
	if len(rd.questions) != 8 {
		t.Fatalf("round has %d questions, want 4 plus 4 second attempts", len(rd.questions))
	}
	for i := 0; i < 4; i++ {
		rd.submit("99") // every second attempt wrong too
		if rd.requeued {
			t.Fatal("a second attempt was sent round a third time")
		}
	}
	if !rd.done {
		t.Fatal("round did not finish once the second attempts were used up")
	}
	if len(rd.questions) != 8 {
		t.Fatalf("round grew to %d questions after the second pass", len(rd.questions))
	}
	if rd.accuracy() != 0 {
		t.Errorf("accuracy = %v, want 0 when nothing was ever right", rd.accuracy())
	}
}

func TestNotesListAMissedQuestionOnce(t *testing.T) {
	rd := newRound(testLevel(), games.NewTestRand(1))
	// Miss the first question twice and answer the rest correctly.
	rd.submit("99")
	for i := 2; i <= 4; i++ {
		rd.submit(strconv.Itoa(i + 1))
	}
	rd.submit("99")

	notes := rd.notes()
	if len(notes) != 1 {
		t.Fatalf("notes = %v, want the one missed question listed once", notes)
	}
	if !strings.Contains(notes[0], "1 + 1 = 2") {
		t.Errorf("note %q does not show the question and its answer", notes[0])
	}
}

// A round can never run away: every question is asked at most twice, so the
// worst case is exactly double the length it started at.
func TestARoundCanAtWorstDoubleInLength(t *testing.T) {
	for _, c := range Chapters {
		for _, l := range c.Levels {
			rd := newRound(l, games.NewTestRand(3))
			started := len(rd.questions)
			for !rd.done {
				rd.submit("-12345") // never right anywhere
			}
			if got := len(rd.questions); got != 2*started {
				t.Errorf("level %q: round of %d became %d, want %d",
					l.Title, started, got, 2*started)
			}
		}
	}
}
