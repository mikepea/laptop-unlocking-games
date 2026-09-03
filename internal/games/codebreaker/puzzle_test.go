package codebreaker

import (
	"strings"
	"testing"

	"github.com/mikepea/laptop-unlocking-games/internal/games"
)

func TestFeedback(t *testing.T) {
	// The whole game lives in this function. Every case here is one a player
	// will actually hit, and the repeats cases are the ones that are easy to
	// get subtly wrong.
	tests := []struct {
		name                string
		secret, guess       []int
		wantExact, wantNear int
	}{
		{"all wrong", []int{1, 2, 3}, []int{4, 5, 6}, 0, 0},
		{"all exact", []int{1, 2, 3}, []int{1, 2, 3}, 3, 0},
		{"all near", []int{1, 2, 3}, []int{3, 1, 2}, 0, 3},
		{"one of each", []int{1, 2, 3, 4}, []int{1, 3, 5, 6}, 1, 1},
		{"an exact match is never also near", []int{1, 1, 2}, []int{1, 3, 3}, 1, 0},
		{"guess repeats a digit the secret has once", []int{1, 2, 3}, []int{1, 1, 1}, 1, 0},
		{"secret repeats a digit the guess has once", []int{1, 1, 1}, []int{1, 2, 3}, 1, 0},
		{"two in the secret, two in the guess, none in place", []int{1, 1, 2, 2}, []int{2, 2, 1, 1}, 0, 4},
		{"two in the secret, three in the guess", []int{1, 1, 5, 6}, []int{2, 2, 1, 1}, 0, 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			exact, near := feedback(tc.secret, tc.guess, 8)
			if exact != tc.wantExact || near != tc.wantNear {
				t.Fatalf("feedback(%v, %v) = (%d exact, %d near), want (%d, %d)",
					tc.secret, tc.guess, exact, near, tc.wantExact, tc.wantNear)
			}
		})
	}
}

func TestFeedbackNeverExceedsTheCodeLength(t *testing.T) {
	// Whatever the guess, exact + near cannot describe more digits than there
	// are. A regression here would show a child impossible clues.
	r := games.NewTestRand(5)
	for _, l := range Levels {
		for i := 0; i < 300; i++ {
			secret := makeCode(l, r)
			guess := makeCode(l, r)
			exact, near := feedback(secret, guess, l.Symbols)
			if exact+near > l.Length {
				t.Fatalf("level %q: feedback(%v, %v) = %d + %d, more than %d digits",
					l.Title, secret, guess, exact, near, l.Length)
			}
		}
	}
}

func TestMakeCodeRespectsTheLevel(t *testing.T) {
	r := games.NewTestRand(9)
	for _, l := range Levels {
		for i := 0; i < 200; i++ {
			code := makeCode(l, r)
			if len(code) != l.Length {
				t.Fatalf("level %q made a code of length %d, want %d", l.Title, len(code), l.Length)
			}
			seen := map[int]bool{}
			for _, d := range code {
				if d < 1 || d > l.Symbols {
					t.Fatalf("level %q made digit %d, want 1..%d", l.Title, d, l.Symbols)
				}
				if !l.Repeats && seen[d] {
					t.Fatalf("level %q forbids repeats but made %v", l.Title, code)
				}
				seen[d] = true
			}
		}
	}
}

func TestSolvingEndsThePuzzle(t *testing.T) {
	pz := newPuzzle(Levels[0], games.NewTestRand(4))
	g := pz.guess(pz.secret)

	if g.Exact != len(pz.secret) {
		t.Fatalf("guessing the secret gave %d exact, want %d", g.Exact, len(pz.secret))
	}
	if !pz.solved || !pz.done() {
		t.Fatal("guessing the secret did not solve the puzzle")
	}
	if pz.used() != 1 {
		t.Fatalf("used = %d, want 1", pz.used())
	}
	if got, want := pz.efficiency(), 1.0; got != want {
		t.Fatalf("efficiency on a first-guess solve = %v, want %v", got, want)
	}
	if pz.notes() != nil {
		t.Fatalf("a solved puzzle produced notes: %v", pz.notes())
	}
}

func TestRunningOutOfGuessesEndsThePuzzleAndRevealsTheCode(t *testing.T) {
	l := Levels[0]
	pz := newPuzzle(l, games.NewTestRand(4))

	// A guess that cannot be right: every digit outside the alphabet.
	wrong := make([]int, l.Length)
	for i := range wrong {
		wrong[i] = 0
	}
	for i := 0; i < l.Guesses; i++ {
		pz.guess(wrong)
	}

	if pz.solved {
		t.Fatal("puzzle reported solved after only wrong guesses")
	}
	if !pz.out || !pz.done() {
		t.Fatal("puzzle did not end after the last guess")
	}
	if pz.remaining() != 0 {
		t.Fatalf("remaining = %d, want 0", pz.remaining())
	}
	if pz.score() != 0 {
		t.Fatalf("score = %d on an unsolved puzzle, want 0", pz.score())
	}
	if got, want := pz.efficiency(), 0.0; got != want {
		t.Fatalf("efficiency = %v on an unsolved puzzle, want %v", got, want)
	}

	notes := pz.notes()
	if len(notes) != 1 {
		t.Fatalf("notes = %v, want one line revealing the code", notes)
	}
	if want := (Guess{Digits: pz.secret}).String(); !strings.Contains(notes[0], want) {
		t.Fatalf("notes %q does not reveal the code %q", notes[0], want)
	}
}

func TestEfficiencyFallsAsGuessesAreUsed(t *testing.T) {
	l := Levels[1]
	first := newPuzzle(l, games.NewTestRand(4))
	first.guess(first.secret)

	later := newPuzzle(l, games.NewTestRand(4))
	wrong := make([]int, l.Length)
	later.guess(wrong)
	later.guess(wrong)
	later.guess(later.secret)

	if !(first.efficiency() > later.efficiency()) {
		t.Fatalf("first-guess efficiency %v is not above third-guess %v",
			first.efficiency(), later.efficiency())
	}
	if !(first.score() > later.score()) {
		t.Fatalf("first-guess score %d is not above third-guess %d", first.score(), later.score())
	}
}

func TestScoreOfPaysNothingForAnUnsolvedCode(t *testing.T) {
	if got := scoreOf(Levels[0], 3, false); got != 0 {
		t.Fatalf("scoreOf(unsolved) = %d, want 0", got)
	}
	if got := scoreOf(Levels[0], Levels[0].Guesses, true); got <= 0 {
		t.Fatalf("scoreOf(solved on the last guess) = %d, want something positive", got)
	}
}

func TestHarderLevelsPayMore(t *testing.T) {
	// Same number of guesses used, harder level: more points.
	easy := scoreOf(Levels[0], 3, true)
	hard := scoreOf(Levels[4], 3, true)
	if !(hard > easy) {
		t.Fatalf("five digits scored %d, three digits scored %d; want the harder level higher", hard, easy)
	}
}

func TestGuessString(t *testing.T) {
	if got, want := (Guess{Digits: []int{1, 4, 2, 6}}).String(), "1 4 2 6"; got != want {
		t.Fatalf("String() = %q, want %q", got, want)
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
		if l.Length < 1 {
			t.Errorf("level %q has length %d", l.Title, l.Length)
		}
		if l.Symbols < 2 || l.Symbols > 9 {
			t.Errorf("level %q uses %d symbols; the input field only accepts 1-9", l.Title, l.Symbols)
		}
		if !l.Repeats && l.Length > l.Symbols {
			t.Errorf("level %q wants %d distinct digits from an alphabet of %d", l.Title, l.Length, l.Symbols)
		}
		if l.Guesses < 1 {
			t.Errorf("level %q allows %d guesses", l.Title, l.Guesses)
		}
	}
}
