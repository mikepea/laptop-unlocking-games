package spelling

import (
	"fmt"
	"math"
	"math/rand/v2"
	"strings"
	"time"
)

// missedWord is a word that was spelled wrong, and what was written instead.
type missedWord struct {
	Want string
	Got  string
}

// round is the scoring core: which words, what was typed, and how long it took.
type round struct {
	level Level
	words []string

	idx     int
	correct int
	missed  []missedWord

	started   bool
	startedAt time.Time
	endedAt   time.Time
	done      bool

	now func() time.Time
}

// newRound draws Count distinct words from the level's pool.
func newRound(l Level, r *rand.Rand) *round {
	pool := make([]string, len(l.Words))
	copy(pool, l.Words)
	r.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })

	n := l.Count
	if n > len(pool) {
		n = len(pool)
	}
	return &round{level: l, words: pool[:n], now: time.Now}
}

// current is the word being asked.
func (rd *round) current() string { return rd.words[rd.idx] }

// showFor is how long the current word stays visible. Longer words get a
// little longer to read, because the challenge is spelling them, not glimpsing
// them.
func (rd *round) showFor() time.Duration {
	base := time.Duration(rd.level.ShowForMillis) * time.Millisecond
	if extra := len(rd.current()) - 6; extra > 0 {
		base += time.Duration(extra) * 150 * time.Millisecond
	}
	return base
}

func (rd *round) start() {
	if !rd.started {
		rd.started = true
		rd.startedAt = rd.now()
	}
}

// submit grades an answer and moves on. Comparison ignores case and
// surrounding space: the skill being tested is which letters, in which order.
func (rd *round) submit(answer string) bool {
	if rd.done {
		return false
	}
	rd.start()

	want := rd.current()
	got := strings.TrimSpace(answer)
	ok := strings.EqualFold(got, want)
	if ok {
		rd.correct++
	} else {
		rd.missed = append(rd.missed, missedWord{Want: want, Got: got})
	}

	rd.idx++
	if rd.idx >= len(rd.words) {
		rd.done = true
		rd.endedAt = rd.now()
	}
	return ok
}

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

func (rd *round) accuracy() float64 {
	if rd.idx == 0 {
		return 1
	}
	return float64(rd.correct) / float64(rd.idx)
}

func (rd *round) passed() bool {
	return rd.done && rd.accuracy() >= rd.level.MinAccuracy
}

func (rd *round) score() int { return scoreOf(rd.correct, rd.idx) }

// notes lists what was missed, alongside what was written instead. Seeing the
// two spellings side by side is the whole lesson.
func (rd *round) notes() []string {
	if len(rd.missed) == 0 {
		return nil
	}
	const maxNotes = 6
	out := make([]string, 0, maxNotes+1)
	for _, w := range rd.missed {
		if len(out) == maxNotes {
			out = append(out, fmt.Sprintf("...and %d more", len(rd.missed)-maxNotes))
			break
		}
		if w.Got == "" {
			out = append(out, fmt.Sprintf("%s (left blank)", w.Want))
			continue
		}
		out = append(out, fmt.Sprintf("%s - you wrote %s", w.Want, w.Got))
	}
	return out
}

// scoreOf turns a round into points.
//
// There is no speed bonus here on purpose. Spelling is recall, not pace, and
// paying for pace would push toward guessing before the word has been read.
func scoreOf(correct, asked int) int {
	if correct == 0 || asked == 0 {
		return 0
	}
	points := 8 * correct
	if accuracy := float64(correct) / float64(asked); accuracy > 0.8 {
		points += int(math.Round((accuracy - 0.8) * 250))
	}
	return points
}
