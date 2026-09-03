package codebreaker

import (
	"math/rand/v2"
	"strings"
	"time"
)

// Level describes one difficulty of puzzle.
type Level struct {
	Title string
	Hint  string
	// Length is how many digits the code has.
	Length int
	// Symbols is the highest digit allowed; codes use 1..Symbols.
	Symbols int
	// Repeats allows the same digit more than once in a code.
	Repeats bool
	// Guesses is how many attempts the player gets.
	Guesses int
}

// Levels is the curriculum, easiest first.
var Levels = []Level{
	{
		Title:  "Three Digits",
		Hint:   "Three digits, 1 to 6, all different. Start anywhere.",
		Length: 3, Symbols: 6, Repeats: false, Guesses: 8,
	},
	{
		Title:  "Four Digits",
		Hint:   "One more digit. Keep track of what you have ruled out.",
		Length: 4, Symbols: 6, Repeats: false, Guesses: 10,
	},
	{
		Title:  "Repeats Allowed",
		Hint:   "A digit can now appear twice. That changes everything.",
		Length: 4, Symbols: 6, Repeats: true, Guesses: 10,
	},
	{
		Title:  "Wider Range",
		Hint:   "Digits 1 to 8 now. More to rule out, same number of guesses.",
		Length: 4, Symbols: 8, Repeats: true, Guesses: 12,
	},
	{
		Title:  "Five Digits",
		Hint:   "The full puzzle. Work it out, do not guess.",
		Length: 5, Symbols: 8, Repeats: true, Guesses: 12,
	},
}

// LevelUnlocked reports whether level i is open given how many are cleared.
func LevelUnlocked(i, cleared int) bool { return i <= cleared }

// Guess is one attempt and what it revealed.
type Guess struct {
	Digits []int
	// Exact is right digit in the right place.
	Exact int
	// Near is right digit in the wrong place.
	Near int
}

// String renders the digits with spaces between, e.g. "1 4 2 6".
func (g Guess) String() string {
	var b strings.Builder
	for i, d := range g.Digits {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteByte(byte('0' + d))
	}
	return b.String()
}

// puzzle is one code and the attempts made against it. No terminal.
type puzzle struct {
	level  Level
	secret []int

	history []Guess

	started   bool
	startedAt time.Time
	endedAt   time.Time

	solved bool
	// out is true when the guesses ran out without a solve.
	out bool

	now func() time.Time
}

func newPuzzle(l Level, r *rand.Rand) *puzzle {
	return &puzzle{level: l, secret: makeCode(l, r), now: time.Now}
}

// makeCode draws a secret for the level.
func makeCode(l Level, r *rand.Rand) []int {
	if l.Repeats {
		code := make([]int, l.Length)
		for i := range code {
			code[i] = r.IntN(l.Symbols) + 1
		}
		return code
	}
	// Without repeats, shuffle the available digits and take a prefix. This
	// cannot loop forever the way rejection sampling can.
	pool := make([]int, l.Symbols)
	for i := range pool {
		pool[i] = i + 1
	}
	r.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
	return pool[:l.Length]
}

func (p *puzzle) start() {
	if !p.started {
		p.started = true
		p.startedAt = p.now()
	}
}

// done reports whether the puzzle is over, either way.
func (p *puzzle) done() bool { return p.solved || p.out }

// used is how many guesses have been made.
func (p *puzzle) used() int { return len(p.history) }

// remaining is how many guesses are left.
func (p *puzzle) remaining() int { return p.level.Guesses - p.used() }

// guess grades an attempt and records it.
func (p *puzzle) guess(digits []int) Guess {
	p.start()

	exact, near := feedback(p.secret, digits, p.level.Symbols)
	g := Guess{Digits: append([]int(nil), digits...), Exact: exact, Near: near}
	p.history = append(p.history, g)

	switch {
	case exact == len(p.secret):
		p.solved = true
		p.endedAt = p.now()
	case p.used() >= p.level.Guesses:
		p.out = true
		p.endedAt = p.now()
	}
	return g
}

func (p *puzzle) elapsed() time.Duration {
	if !p.started {
		return 0
	}
	end := p.endedAt
	if end.IsZero() {
		end = p.now()
	}
	return end.Sub(p.startedAt)
}

// efficiency is how well the code was cracked, 0..1. Cracking it on the first
// guess scores 1; using every guess scores just above 0; failing scores 0.
func (p *puzzle) efficiency() float64 {
	if !p.solved {
		return 0
	}
	return float64(p.level.Guesses-p.used()+1) / float64(p.level.Guesses)
}

func (p *puzzle) score() int { return scoreOf(p.level, p.used(), p.solved) }

// notes explains the outcome for the results screen.
func (p *puzzle) notes() []string {
	if p.solved {
		return nil
	}
	if !p.out {
		return []string{"You left before cracking it."}
	}
	return []string{"The code was " + Guess{Digits: p.secret}.String()}
}

// feedback compares a guess against the secret and returns exact and near
// matches. It is the whole game in one function.
//
// A digit already counted as exact cannot also count as near, and a digit
// appearing twice in the guess but once in the secret earns one near at most —
// which is what makes the repeats level genuinely harder.
func feedback(secret, guess []int, symbols int) (exact, near int) {
	secretCounts := make([]int, symbols+1)
	guessCounts := make([]int, symbols+1)

	for i := range secret {
		switch {
		case i >= len(guess):
			secretCounts[secret[i]]++
		case secret[i] == guess[i]:
			exact++
		default:
			secretCounts[secret[i]]++
			guessCounts[guess[i]]++
		}
	}

	for d := 1; d <= symbols; d++ {
		near += min(secretCounts[d], guessCounts[d])
	}
	return exact, near
}

// scoreOf turns a finished puzzle into points.
//
// Unlike the other games there is no partial credit: a code is cracked or it is
// not. Deduction rewards patience, so the bonus is for guesses left over rather
// than for time taken.
func scoreOf(l Level, used int, solved bool) int {
	if !solved {
		return 0
	}
	base := 25 + 10*(l.Length-3) + 5*(l.Symbols-6)
	return base + 12*(l.Guesses-used)
}
