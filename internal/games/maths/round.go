package maths

import (
	"fmt"
	"math"
	"math/rand/v2"
	"strconv"
	"time"
)

func fmtQ(a int, op string, b int) string { return fmt.Sprintf("%d %s %d", a, op, b) }

// round is the scoring core: questions, answers and the clock. No terminal.
type round struct {
	level     Level
	questions []Question

	idx     int
	correct int
	// missed keeps the questions that were answered wrong, in order, for the
	// "worth another look" list on the results screen.
	missed []Question

	started   bool
	startedAt time.Time
	endedAt   time.Time
	done      bool

	now func() time.Time
}

func newRound(l Level, r *rand.Rand) *round {
	var qs []Question
	if l.GenSteps != nil {
		// Whole problems only. Stopping halfway through the working would
		// leave an expression collected but never factorised, so a round runs
		// slightly over Questions rather than cut one short.
		for len(qs) < l.Questions {
			qs = append(qs, l.GenSteps(r)...)
		}
	} else {
		qs = make([]Question, l.Questions)
		for i := range qs {
			qs[i] = l.Gen(r)
		}
	}
	return &round{level: l, questions: qs, now: time.Now}
}

// current is the question being asked.
//
// After the final answer idx is one past the end. Bubble Tea renders the model
// it was handed back before the FinishedMsg from games.Finish is delivered and
// the launcher swaps this model out, so View always runs at least once on a
// finished round. Return the last question rather than panicking.
func (rd *round) current() Question {
	if len(rd.questions) == 0 {
		return Question{}
	}
	if rd.idx >= len(rd.questions) {
		return rd.questions[len(rd.questions)-1]
	}
	return rd.questions[rd.idx]
}

// start marks the clock as running. It is called on the first keystroke rather
// than when the level opens, so reading the first question costs nothing.
func (rd *round) start() {
	if !rd.started {
		rd.started = true
		rd.startedAt = rd.now()
	}
}

// submit grades an answer and moves on. It reports whether the answer was
// right; a blank or unparseable answer is simply wrong, not an error.
func (rd *round) submit(answer string) bool {
	if rd.done {
		return false
	}
	rd.start()

	q := rd.current()
	got, err := strconv.Atoi(answer)
	ok := err == nil && got == q.Answer
	if ok {
		rd.correct++
	} else {
		rd.missed = append(rd.missed, q)
	}

	rd.idx++
	if rd.idx >= len(rd.questions) {
		rd.done = true
		rd.endedAt = rd.now()
	}
	return ok
}

// asked is how many questions have been answered.
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

// accuracy is the share of questions answered right so far, 0..1.
func (rd *round) accuracy() float64 {
	if rd.idx == 0 {
		return 1
	}
	return float64(rd.correct) / float64(rd.idx)
}

// passed reports whether the round was finished and met its accuracy bar.
func (rd *round) passed() bool {
	return rd.done && rd.accuracy() >= rd.level.MinAccuracy
}

func (rd *round) score() int { return scoreOf(rd.correct, rd.idx, rd.elapsed()) }

// notes lists the missed questions with their answers, so the results screen
// can show what to look at again.
func (rd *round) notes() []string {
	if len(rd.missed) == 0 {
		return nil
	}
	// Anything longer than this is a wall of text rather than a lesson.
	const maxNotes = 6
	out := make([]string, 0, maxNotes)
	for _, q := range rd.missed {
		if len(out) == maxNotes {
			out = append(out, fmt.Sprintf("...and %d more", len(rd.missed)-maxNotes))
			break
		}
		out = append(out, fmt.Sprintf("%s = %d", q.Prompt, q.Answer))
	}
	return out
}

// minRate is the shortest window a pace is measured over, so answering the
// first question instantly does not report an absurd rate.
const minRate = 5 * time.Second

// scoreOf turns a round into points.
//
// Same shape as the typing game: correct answers are the floor, pace adds a
// capped bonus, and accuracy pays only above 80% so guessing fast is never
// worth more than working it out.
func scoreOf(correct, asked int, elapsed time.Duration) int {
	if correct == 0 || asked == 0 {
		return 0
	}
	points := 6 * correct

	if elapsed < minRate {
		elapsed = minRate
	}
	// Questions per minute, capped so a lucky streak cannot run away with it.
	qpm := float64(asked) / elapsed.Minutes()
	points += int(math.Round(math.Min(qpm, 30)))

	if accuracy := float64(correct) / float64(asked); accuracy > 0.8 {
		points += int(math.Round((accuracy - 0.8) * 250))
	}
	return points
}
