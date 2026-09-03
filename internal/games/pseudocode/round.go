package pseudocode

import (
	"fmt"
	"math"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"
)

// round is the scoring core: programs, answers and the clock. No terminal.
//
// Same shape as the maths round deliberately. Each game owning its own is the
// pattern here — spelling and typing do the same — because what a "question"
// is differs every time, and sharing it would mean a type that fits none of
// them well.
type round struct {
	level    Level
	programs []Program

	idx     int
	correct int
	// missed keeps the programs answered wrong, in order, for the "worth
	// another look" list on the results screen.
	missed []Program

	started   bool
	startedAt time.Time
	endedAt   time.Time
	done      bool

	now func() time.Time
}

func newRound(l Level, r *rand.Rand) *round {
	ps := make([]Program, l.Questions)
	for i := range ps {
		ps[i] = l.Gen(r)
	}
	return &round{level: l, programs: ps, now: time.Now}
}

// current is the program being traced.
//
// After the final answer idx is one past the end, and the launcher still
// renders this model once before swapping it out, so clamp rather than panic.
func (rd *round) current() Program {
	if len(rd.programs) == 0 {
		return Program{}
	}
	if rd.idx >= len(rd.programs) {
		return rd.programs[len(rd.programs)-1]
	}
	return rd.programs[rd.idx]
}

// start marks the clock as running. It is called on the first keystroke rather
// than when the level opens, so reading the first program costs nothing.
func (rd *round) start() {
	if !rd.started {
		rd.started = true
		rd.startedAt = rd.now()
	}
}

// submit grades an answer and moves on. A blank or unparseable answer is
// simply wrong, not an error.
func (rd *round) submit(answer string) bool {
	if rd.done {
		return false
	}
	rd.start()

	p := rd.current()
	got, err := strconv.Atoi(answer)
	ok := err == nil && got == p.Answer
	if ok {
		rd.correct++
	} else {
		rd.missed = append(rd.missed, p)
	}

	rd.idx++
	if rd.idx >= len(rd.programs) {
		rd.done = true
		rd.endedAt = rd.now()
	}
	return ok
}

func (rd *round) asked() int { return rd.idx }

func (rd *round) elapsed() time.Duration {
	if !rd.started {
		return 0
	}
	end := rd.endedAt
	if end.IsZero() {
		end = rd.now()
	}
	return end.Sub(rd.startedAt)
}

// accuracy is the share answered right so far, 0..1.
func (rd *round) accuracy() float64 {
	if rd.idx == 0 {
		return 1
	}
	return float64(rd.correct) / float64(rd.idx)
}

func (rd *round) passed() bool {
	return rd.done && rd.accuracy() >= rd.level.MinAccuracy
}

func (rd *round) score() int { return scoreOf(rd.correct, rd.idx, rd.elapsed()) }

// notes lists the missed programs on one line each, with what they printed.
func (rd *round) notes() []string {
	if len(rd.missed) == 0 {
		return nil
	}
	const maxNotes = 4 // a traced program is wider than a sum; fewer fit
	out := make([]string, 0, maxNotes)
	for _, p := range rd.missed {
		if len(out) == maxNotes {
			out = append(out, fmt.Sprintf("...and %d more", len(rd.missed)-maxNotes))
			break
		}
		out = append(out, fmt.Sprintf("%s  prints %d", oneLine(p), p.Answer))
	}
	return out
}

// oneLine squashes a program onto a single line for the results screen, where
// there is no room to lay it out properly.
func oneLine(p Program) string {
	parts := make([]string, 0, len(p.Lines))
	for _, l := range p.Lines {
		parts = append(parts, strings.TrimSpace(l))
	}
	return strings.Join(parts, " · ")
}

// minRate is the shortest window a pace is measured over, so answering the
// first question instantly does not report an absurd rate.
const minRate = 5 * time.Second

// scoreOf turns a round into points.
//
// Same shape as the maths game, but worth more per answer and capped at a
// lower pace: tracing four lines is slower than recalling 7 x 8, and rewarding
// speed here would push exactly the guessing the game is trying to replace.
func scoreOf(correct, asked int, elapsed time.Duration) int {
	if correct == 0 || asked == 0 {
		return 0
	}
	points := 9 * correct

	if elapsed < minRate {
		elapsed = minRate
	}
	qpm := float64(asked) / elapsed.Minutes()
	points += int(math.Round(math.Min(qpm, 12)))

	if accuracy := float64(correct) / float64(asked); accuracy > 0.8 {
		points += int(math.Round((accuracy - 0.8) * 250))
	}
	return points
}
