package typing

import (
	"math"
	"time"
)

// session is the pure scoring core of the typing game: it knows nothing about
// terminals, which keeps the rules testable.
//
// The clock starts on the first keystroke, not when the lesson opens, so
// staring at the screen for a moment costs nothing.
type session struct {
	lesson Lesson

	lineIdx int
	target  []rune
	typed   []rune

	started    bool
	startedAt  time.Time
	finishedAt time.Time
	done       bool

	// keystrokes counts every character the player committed, including ones
	// they later deleted. hits counts the subset that were right first time.
	// Accuracy from these two numbers rewards getting it right, not fixing it.
	keystrokes int
	hits       int

	// committed is the character count of every line already finished.
	committed int

	now func() time.Time // injectable for tests
}

func newSession(l Lesson) *session {
	s := &session{lesson: l, now: time.Now}
	s.target = []rune(l.Lines[0])
	return s
}

// typeRune records one character of input.
func (s *session) typeRune(r rune) {
	if s.done || len(s.typed) >= len(s.target) {
		return
	}
	if !s.started {
		s.started = true
		s.startedAt = s.now()
	}
	s.keystrokes++
	if r == s.target[len(s.typed)] {
		s.hits++
	}
	s.typed = append(s.typed, r)
}

// backspace deletes the last character typed on the current line. It does not
// reach back into a line that has already been completed.
func (s *session) backspace() {
	if s.done || len(s.typed) == 0 {
		return
	}
	s.typed = s.typed[:len(s.typed)-1]
}

// lineClean reports whether every character typed so far on this line matches.
func (s *session) lineClean() bool {
	for i, r := range s.typed {
		if r != s.target[i] {
			return false
		}
	}
	return true
}

// lineComplete reports whether the current line is fully and correctly typed.
func (s *session) lineComplete() bool {
	return len(s.typed) == len(s.target) && s.lineClean()
}

// advance moves to the next line, or ends the lesson if this was the last one.
// It reports whether the lesson is now done.
func (s *session) advance() bool {
	if s.done {
		return true
	}
	s.committed += len(s.target)
	s.lineIdx++
	if s.lineIdx >= len(s.lesson.Lines) {
		s.done = true
		s.finishedAt = s.now()
		return true
	}
	s.target = []rune(s.lesson.Lines[s.lineIdx])
	s.typed = s.typed[:0]
	return false
}

// elapsed is the time spent typing.
func (s *session) elapsed() time.Duration {
	if !s.started {
		return 0
	}
	end := s.finishedAt
	if end.IsZero() {
		end = s.now()
	}
	return end.Sub(s.startedAt)
}

// correctChars is every character currently standing correct, across finished
// lines and the line in progress.
func (s *session) correctChars() int {
	n := s.committed
	for i, r := range s.typed {
		if r == s.target[i] {
			n++
		}
	}
	return n
}

// minRate is the shortest window the rate is computed over. Without it the
// first keystroke of a lesson divides by a few microseconds and the display
// reads five figures before settling.
const minRate = 2 * time.Second

// wpm uses the standard five-characters-per-word convention.
func (s *session) wpm() float64 {
	elapsed := s.elapsed()
	if elapsed <= 0 {
		return 0
	}
	if elapsed < minRate {
		elapsed = minRate
	}
	return (float64(s.correctChars()) / 5) / elapsed.Minutes()
}

// accuracy is the share of keystrokes that were right first time, 0..1.
func (s *session) accuracy() float64 {
	if s.keystrokes == 0 {
		return 1
	}
	return float64(s.hits) / float64(s.keystrokes)
}

// passed reports whether the lesson was cleared to its accuracy bar. A lesson
// only counts as passed if it was played to the end.
func (s *session) passed() bool {
	return s.done && s.accuracy() >= s.lesson.MinAccuracy
}

// score is the points-facing number shown to the player.
func (s *session) score() int {
	return scoreOf(s.linesFinished(), s.wpm(), s.accuracy())
}

// linesFinished is how many lines were typed out in full.
func (s *session) linesFinished() int { return s.lineIdx }

// scoreOf turns a run into points.
//
// The shape matters more than the constants: finishing lines is the floor,
// speed adds a linear bonus, and accuracy pays only above 80% so that hammering
// keys is never worth more than typing carefully.
func scoreOf(lines int, wpm, accuracy float64) int {
	if lines <= 0 {
		return 0
	}
	points := 5 * lines
	points += int(math.Round(wpm))
	if accuracy > 0.8 {
		points += int(math.Round((accuracy - 0.8) * 250))
	}
	return points
}
