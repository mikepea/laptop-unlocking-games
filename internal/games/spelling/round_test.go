package spelling

import (
	"strings"
	"testing"
	"time"

	"github.com/mikepea/laptop-unlocking-games/internal/games"
)

func testLevel() Level {
	return Level{
		Title:         "Test",
		Count:         4,
		MinAccuracy:   0.75,
		ShowForMillis: 1000,
		Words: []Word{
			{"alpha", "The alpha wolf led the pack."},
			{"bravo", "The crowd shouted bravo at the end."},
			{"charlie", "We called the puppy charlie."},
			{"delta", "The river spreads out into a delta."},
			{"echo", "Her voice came back as an echo."},
			{"foxtrot", "They danced a slow foxtrot."},
		},
	}
}

func TestRoundDrawsDistinctWordsFromThePool(t *testing.T) {
	rd := newRound(testLevel(), games.NewTestRand(3))

	if len(rd.words) != 4 {
		t.Fatalf("drew %d words, want 4", len(rd.words))
	}
	seen := map[string]bool{}
	pool := map[string]bool{}
	for _, w := range testLevel().Words {
		pool[w.Text] = true
	}
	for _, w := range rd.words {
		if seen[w.Text] {
			t.Errorf("word %q drawn twice in one round", w.Text)
		}
		seen[w.Text] = true
		if !pool[w.Text] {
			t.Errorf("word %q is not in the level's pool", w.Text)
		}
	}
}

func TestRoundHandlesAPoolSmallerThanTheCount(t *testing.T) {
	l := testLevel()
	l.Count = 99
	rd := newRound(l, games.NewTestRand(3))

	if len(rd.words) != len(l.Words) {
		t.Fatalf("drew %d words from a pool of %d", len(rd.words), len(l.Words))
	}
}

func TestSubmitIgnoresCaseAndSurroundingSpace(t *testing.T) {
	rd := newRound(testLevel(), games.NewTestRand(3))
	want := rd.current().Text

	if !rd.submit("  " + strings.ToUpper(want) + "  ") {
		t.Fatalf("%q in capitals with spaces was graded wrong", want)
	}
	if rd.correct != 1 {
		t.Fatalf("correct = %d, want 1", rd.correct)
	}
}

func TestSubmitRecordsWhatWasWritten(t *testing.T) {
	rd := newRound(testLevel(), games.NewTestRand(3))
	want := rd.current().Text

	if rd.submit("definitelywrong") {
		t.Fatal("a wrong spelling was graded correct")
	}
	if len(rd.missed) != 1 {
		t.Fatalf("missed = %+v, want one entry", rd.missed)
	}
	if rd.missed[0].Want != want || rd.missed[0].Got != "definitelywrong" {
		t.Fatalf("missed[0] = %+v, want both spellings recorded", rd.missed[0])
	}

	notes := rd.notes()
	if len(notes) != 1 || !strings.Contains(notes[0], want) || !strings.Contains(notes[0], "definitelywrong") {
		t.Fatalf("notes = %v, want both spellings side by side", notes)
	}
}

func TestNotesSayWhenAnAnswerWasBlank(t *testing.T) {
	rd := newRound(testLevel(), games.NewTestRand(3))
	rd.submit("   ")

	notes := rd.notes()
	if len(notes) != 1 || !strings.Contains(notes[0], "blank") {
		t.Fatalf("notes = %v, want a blank-answer note", notes)
	}
}

func TestRoundFinishesAndGradesPass(t *testing.T) {
	rd := newRound(testLevel(), games.NewTestRand(3))
	for !rd.done {
		rd.submit(rd.current().Text)
	}
	if !rd.passed() {
		t.Fatalf("a perfect round did not pass (accuracy %v)", rd.accuracy())
	}
	if rd.notes() != nil {
		t.Fatalf("a perfect round produced notes: %v", rd.notes())
	}

	// Half right is below the 0.75 bar.
	sloppy := newRound(testLevel(), games.NewTestRand(3))
	for i := 0; !sloppy.done; i++ {
		if i%2 == 0 {
			sloppy.submit(sloppy.current().Text)
			continue
		}
		sloppy.submit("wrong")
	}
	if sloppy.passed() {
		t.Fatalf("a round at %v accuracy passed a 0.75 level", sloppy.accuracy())
	}
}

func TestShowForGrowsWithWordLength(t *testing.T) {
	l := testLevel()
	l.Words = []Word{{"cat", "The cat sat on the wall."}}
	l.Count = 1
	short := newRound(l, games.NewTestRand(3))

	l.Words = []Word{{"extraordinary", "It was an extraordinary sight."}}
	long := newRound(l, games.NewTestRand(3))

	if !(long.showFor() > short.showFor()) {
		t.Fatalf("long word shows for %v, short for %v; want longer", long.showFor(), short.showFor())
	}
	if short.showFor() != time.Duration(l.ShowForMillis)*time.Millisecond+sentenceExtra {
		t.Fatalf("a short word got %v, want the level's base of %dms", short.showFor(), l.ShowForMillis)
	}
}

func TestScoreOfHasNoSpeedComponent(t *testing.T) {
	tests := []struct {
		name           string
		correct, asked int
		want           int
	}{
		{"nothing right scores nothing", 0, 10, 0},
		{"perfect", 10, 10, 80 + 50},
		{"accuracy pays only above 80%", 8, 10, 64},
		{"just above the bar", 9, 10, 72 + 25},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := scoreOf(tc.correct, tc.asked); got != tc.want {
				t.Fatalf("scoreOf(%d, %d) = %d, want %d", tc.correct, tc.asked, got, tc.want)
			}
		})
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
		if len(l.Words) < l.Count {
			t.Errorf("level %q asks for %d words from a pool of %d", l.Title, l.Count, len(l.Words))
		}
		if l.ShowForMillis <= 0 {
			t.Errorf("level %q shows words for %dms", l.Title, l.ShowForMillis)
		}
		if l.MinAccuracy <= 0 || l.MinAccuracy > 1 {
			t.Errorf("level %q has MinAccuracy %v", l.Title, l.MinAccuracy)
		}

		seen := map[string]bool{}
		for _, w := range l.Words {
			if seen[w.Text] {
				t.Errorf("level %q lists %q twice", l.Title, w.Text)
			}
			seen[w.Text] = true

			// Every word has to be typeable with the letters-only field, or it
			// cannot be answered however well it is spelled.
			for _, r := range w.Text {
				if !isTypeable(r) {
					t.Errorf("level %q word %q contains %q, which the answer field rejects", l.Title, w.Text, r)
				}
			}
		}
	}
}

// isTypeable mirrors ui.Letters, which is the filter on the answer field.
func isTypeable(r rune) bool {
	switch {
	case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z':
		return true
	case r == '\'' || r == '-':
		return true
	}
	return false
}

// A finished round is still rendered once before the launcher swaps the model
// out, so current() must stay in bounds past the last word rather than
// panicking. This is the bug that hit the maths game.
func TestCurrentIsStableOnceTheRoundIsOver(t *testing.T) {
	rd := newRound(testLevel(), games.NewTestRand(3))
	last := rd.words[len(rd.words)-1]

	for !rd.done {
		rd.submit(rd.current().Text)
	}
	if rd.idx != len(rd.words) {
		t.Fatalf("idx = %d, want %d", rd.idx, len(rd.words))
	}
	if got := rd.current(); got != last {
		t.Errorf("current() after finishing = %q, want the last word %q", got, last)
	}
	// showFor also calls current(); it must not panic either.
	_ = rd.showFor()
}

// The sentence is the whole point of the change: without it "Sound Alikes" is
// unanswerable, since "weight" and "wait" look identical when shown alone. A
// sentence that has drifted from its word would light up nothing, so check
// every one of them.
func TestEveryWordAppearsInItsOwnSentence(t *testing.T) {
	for _, l := range Levels {
		for _, w := range l.Words {
			start, end, ok := wordSpan(w.Sentence, w.Text)
			if !ok {
				t.Errorf("level %q: %q does not appear as a whole word in %q",
					l.Title, w.Text, w.Sentence)
				continue
			}
			if got := strings.ToLower(w.Sentence[start:end]); got != strings.ToLower(w.Text) {
				t.Errorf("level %q: highlight for %q lands on %q in %q",
					l.Title, w.Text, w.Sentence, got)
			}
		}
	}
}

func TestEverySentenceIsPresentAndReadable(t *testing.T) {
	for _, l := range Levels {
		for _, w := range l.Words {
			if w.Sentence == "" {
				t.Errorf("level %q: %q has no sentence", l.Title, w.Text)
			}
			// The sentence is shown on one line inside the frame.
			if len(w.Sentence) > 60 {
				t.Errorf("level %q: sentence for %q is %d chars: %q",
					l.Title, w.Text, len(w.Sentence), w.Sentence)
			}
		}
	}
}

func TestWordSpanMatchesWholeWordsOnly(t *testing.T) {
	if _, _, ok := wordSpan("The town was quiet.", "own"); ok {
		t.Error(`"own" matched inside "town"`)
	}
	start, end, ok := wordSpan("Suddenly the lights went out.", "suddenly")
	if !ok || start != 0 || end != len("Suddenly") {
		t.Errorf("span = %d,%d,%v; want the leading capitalised word", start, end, ok)
	}
	if _, _, ok := wordSpan("Nothing to find here.", "absent"); ok {
		t.Error("a missing word reported a span")
	}
}
