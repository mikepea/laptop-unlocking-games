package typing

import (
	"testing"
	"time"
)

// lesson returns a two-line lesson with a known character count.
func lesson() Lesson {
	return Lesson{
		Title:       "Test",
		MinAccuracy: 0.9,
		Lines:       []string{"abc", "de"},
	}
}

// typeString feeds a whole string through the session.
func typeString(s *session, text string) {
	for _, r := range text {
		s.typeRune(r)
	}
}

func TestSessionClockStartsOnFirstKeystroke(t *testing.T) {
	s := newSession(lesson())
	now := time.Unix(1000, 0)
	s.now = func() time.Time { return now }

	if got := s.elapsed(); got != 0 {
		t.Fatalf("elapsed before typing = %v, want 0", got)
	}

	s.typeRune('a')
	now = now.Add(30 * time.Second)
	if got := s.elapsed(); got != 30*time.Second {
		t.Fatalf("elapsed = %v, want 30s", got)
	}
}

func TestLineCompleteRequiresEveryCharacterCorrect(t *testing.T) {
	s := newSession(lesson())

	typeString(s, "abx")
	if s.lineComplete() {
		t.Fatal("line with a wrong character reported complete")
	}
	if s.lineClean() {
		t.Fatal("line with a wrong character reported clean")
	}

	s.backspace()
	s.typeRune('c')
	if !s.lineComplete() {
		t.Fatal("fully correct line not reported complete")
	}
}

func TestTypeRuneIgnoresInputPastEndOfLine(t *testing.T) {
	s := newSession(lesson())
	typeString(s, "abcdef")

	if len(s.typed) != 3 {
		t.Fatalf("typed %d characters, want 3 (line length)", len(s.typed))
	}
	if s.keystrokes != 3 {
		t.Fatalf("counted %d keystrokes, want 3", s.keystrokes)
	}
}

func TestAccuracyCountsFirstAttemptsOnly(t *testing.T) {
	s := newSession(lesson())

	// Four keystrokes, three of them right: the corrected mistake still counts
	// against accuracy, which is the point.
	typeString(s, "ab")
	s.typeRune('x')
	s.backspace()
	s.typeRune('c')

	if s.keystrokes != 4 {
		t.Fatalf("keystrokes = %d, want 4", s.keystrokes)
	}
	if got, want := s.accuracy(), 0.75; got != want {
		t.Fatalf("accuracy = %v, want %v", got, want)
	}
}

func TestAdvanceMovesThroughLinesThenFinishes(t *testing.T) {
	s := newSession(lesson())
	now := time.Unix(1000, 0)
	s.now = func() time.Time { return now }

	typeString(s, "abc")
	if done := s.advance(); done {
		t.Fatal("advance finished the lesson after the first of two lines")
	}
	if s.lineIdx != 1 {
		t.Fatalf("lineIdx = %d, want 1", s.lineIdx)
	}
	if len(s.typed) != 0 {
		t.Fatalf("typed buffer not reset on advance: %q", string(s.typed))
	}

	typeString(s, "de")
	if done := s.advance(); !done {
		t.Fatal("advance did not finish the lesson after the last line")
	}
	if !s.done {
		t.Fatal("session not marked done")
	}
	if s.linesFinished() != 2 {
		t.Fatalf("linesFinished = %d, want 2", s.linesFinished())
	}
}

func TestPassedRequiresFinishingAndAccuracy(t *testing.T) {
	// Perfect run: passes.
	s := newSession(lesson())
	typeString(s, "abc")
	s.advance()
	typeString(s, "de")
	s.advance()
	if !s.passed() {
		t.Fatalf("perfect run did not pass (accuracy %v)", s.accuracy())
	}

	// Same lesson, finished but sloppy: below the 90% bar.
	sloppy := newSession(lesson())
	typeString(sloppy, "xyz")
	for i := 0; i < 3; i++ {
		sloppy.backspace()
	}
	typeString(sloppy, "abc")
	sloppy.advance()
	typeString(sloppy, "de")
	sloppy.advance()
	if sloppy.passed() {
		t.Fatalf("run at %v accuracy passed a 0.9 lesson", sloppy.accuracy())
	}

	// Abandoned mid-lesson: never passes, however clean.
	abandoned := newSession(lesson())
	typeString(abandoned, "abc")
	if abandoned.passed() {
		t.Fatal("unfinished lesson reported as passed")
	}
}

func TestWPMUsesFiveCharacterWords(t *testing.T) {
	s := newSession(Lesson{Title: "T", MinAccuracy: 0.9, Lines: []string{"abcdefghij"}})
	now := time.Unix(0, 0)
	s.now = func() time.Time { return now }

	typeString(s, "abcdefghij")
	now = now.Add(30 * time.Second)

	// Ten correct characters is two words; two words in half a minute is 4 wpm.
	if got, want := s.wpm(), 4.0; got != want {
		t.Fatalf("wpm = %v, want %v", got, want)
	}
}

func TestScoreOf(t *testing.T) {
	tests := []struct {
		name     string
		lines    int
		wpm      float64
		accuracy float64
		want     int
	}{
		{"nothing finished scores nothing", 0, 60, 1.0, 0},
		{"lines alone", 4, 0, 0, 20},
		{"speed adds linearly", 4, 30, 0, 50},
		{"accuracy pays only above 80%", 4, 0, 0.80, 20},
		{"perfect accuracy caps the bonus", 4, 0, 1.0, 70},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := scoreOf(tc.lines, tc.wpm, tc.accuracy); got != tc.want {
				t.Fatalf("scoreOf(%d, %v, %v) = %d, want %d", tc.lines, tc.wpm, tc.accuracy, got, tc.want)
			}
		})
	}
}

func TestLessonUnlocked(t *testing.T) {
	if !LessonUnlocked(0, 0) {
		t.Fatal("first lesson should always be open")
	}
	if LessonUnlocked(1, 0) {
		t.Fatal("second lesson open with nothing cleared")
	}
	if !LessonUnlocked(1, 1) {
		t.Fatal("second lesson closed after clearing the first")
	}
}

func TestCuratedLessonsAreWellFormed(t *testing.T) {
	for i, l := range Lessons {
		if l.Title == "" {
			t.Errorf("lesson %d has no title", i)
		}
		if len(l.Lines) == 0 {
			t.Errorf("lesson %q has no lines", l.Title)
		}
		if l.MinAccuracy <= 0 || l.MinAccuracy > 1 {
			t.Errorf("lesson %q has MinAccuracy %v, want 0 < x <= 1", l.Title, l.MinAccuracy)
		}
		for j, line := range l.Lines {
			if line == "" {
				t.Errorf("lesson %q line %d is empty", l.Title, j)
			}
			// 48 columns keeps a line inside an 80-column TTY with framing.
			if n := len([]rune(line)); n > 48 {
				t.Errorf("lesson %q line %d is %d characters, want <= 48", l.Title, j, n)
			}
			// Every character has to be reachable on a UK/US keyboard. An em
			// dash or a curly quote sneaking in makes a line impossible to
			// finish, and the player cannot skip it.
			for _, r := range line {
				if r < ' ' || r > '~' {
					t.Errorf("lesson %q line %d contains %q, which is not on a keyboard", l.Title, j, r)
				}
			}
		}
	}
}
