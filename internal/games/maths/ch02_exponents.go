package maths

import (
	"fmt"
	"math/rand/v2"
)

// Chapter 2 is exponents: what a power means, and the three rules that fall
// out of writing one down in full. The rules are the point, so most of these
// levels ask for the power rather than the value -- 2^3 x 2^4 = 2^7 is the
// thing being learned, and working out that it is 128 as well is arithmetic
// that gets in the way of noticing 3 + 4 = 7.
//
// Powers are written 2^5, which is what he will type and near enough to what
// the book prints.

// power renders b^e.
func power(b, e int) string { return fmt.Sprintf("%d^%d", b, e) }

// safePowers are base/largest-exponent pairs whose value stays under about a
// thousand, so "Powers" is a question about what a power means and not a long
// multiplication.
var safePowers = [][2]int{
	{2, 10}, {3, 6}, {4, 5}, {5, 4}, {6, 3}, {7, 3}, {8, 3}, {9, 3}, {10, 3},
}

var exponentLevels = []Level{
	{
		Title:       "Powers",
		Hint:        "2^5 means five 2s multiplied together. Not 2 times 5.",
		Questions:   8,
		MinAccuracy: 0.75,
		Gen: func(r *rand.Rand) Question {
			p := safePowers[r.IntN(len(safePowers))]
			base, e := p[0], r.IntN(p[1]-1)+2
			return Question{Prompt: power(base, e), Answer: Int(pow(base, e))}
		},
	},
	{
		Title:       "Multiplying Powers",
		Hint:        "Same base: write both out in full and count. The powers add.",
		Questions:   8,
		MinAccuracy: 0.75,
		Gen: func(r *rand.Rand) Question {
			base := r.IntN(8) + 2
			e1, e2 := r.IntN(5)+1, r.IntN(5)+1
			ctx := fmt.Sprintf("%s x %s = %s", power(base, e1), power(base, e2), power(base, e1+e2))
			// Every so often ask what it comes to as well, but only when the
			// answer is a number worth writing down.
			if v := pow(base, e1+e2); v <= 1024 && r.IntN(4) == 0 {
				return Question{Context: ctx, Prompt: power(base, e1+e2), Answer: Int(v)}
			}
			return Question{
				Context: fmt.Sprintf("%s x %s = %d^?", power(base, e1), power(base, e2), base),
				Prompt:  "?", Answer: Int(e1 + e2),
			}
		},
	},
	{
		Title:       "Dividing Powers",
		Hint:        "Same base again. The powers take away.",
		Questions:   8,
		MinAccuracy: 0.75,
		Gen: func(r *rand.Rand) Question {
			base := r.IntN(8) + 2
			e2 := r.IntN(4) + 1
			e1 := e2 + r.IntN(5) + 1
			return Question{
				Context: fmt.Sprintf("%s / %s = %d^?", power(base, e1), power(base, e2), base),
				Prompt:  "?", Answer: Int(e1 - e2),
			}
		},
	},
	{
		Title:       "Powers of Powers",
		Hint:        "(2^3)^4 is four lots of three 2s. The powers multiply.",
		Questions:   8,
		MinAccuracy: 0.70,
		Gen: func(r *rand.Rand) Question {
			base := r.IntN(8) + 2
			m, n := r.IntN(4)+2, r.IntN(4)+2
			if r.IntN(3) == 0 {
				// A product raised to a power: each factor takes the power.
				// Asking for one of them makes the point without needing the
				// answer to be an expression.
				other := base + r.IntN(5) + 1
				return Question{
					Context: fmt.Sprintf("(%d x %d)^%d = %d^? x %s", base, other, n, base, power(other, n)),
					Prompt:  "?", Answer: Int(n),
				}
			}
			return Question{
				Context: fmt.Sprintf("(%s)^%d = %d^?", power(base, m), n, base),
				Prompt:  "?", Answer: Int(m * n),
			}
		},
	},
	{
		Title:       "Zero and Below",
		Hint:        "Anything to the power nought is 1. A negative power is one over the power.",
		Questions:   8,
		MinAccuracy: 0.70,
		Gen: func(r *rand.Rand) Question {
			base := r.IntN(8) + 2
			switch r.IntN(4) {
			case 0:
				return Question{Prompt: power(base, 0), Answer: Int(1)}
			case 1:
				return Question{Prompt: power(base, 1), Answer: Int(base)}
			case 2:
				// The reason the rule has to be that: keep taking one off the
				// power and you keep dividing, so past zero you are dividing
				// into a fraction.
				e := r.IntN(3) + 1
				return Question{
					Prompt: fmt.Sprintf("%d^-%d", base, e),
					Answer: Fraction(1, pow(base, e)),
				}
			default:
				e := r.IntN(3) + 2
				return Question{
					Context: fmt.Sprintf("%d^-%d = 1/?", base, e),
					Prompt:  "?", Answer: Int(pow(base, e)),
				}
			}
		},
	},
}
