package maths

import (
	"fmt"
	"math/rand/v2"
	"strconv"
)

// Chapter 4 is fractions, and it is the reason the answer stopped being an
// int. Every level here wants a fraction typed as a fraction -- "3/4", in
// lowest terms -- and refuses the decimal of the same value, because a child
// who answers 0.75 to "1/4 + 1/2" has practised a calculator, not a fraction.
//
// Denominators stay in the times tables he already has. The work is meant to
// be finding the common denominator, not multiplying 13 by 17.

// friendlyDenominators are the ones whose common denominators stay small.
var friendlyDenominators = []int{2, 3, 4, 5, 6, 8, 9, 10, 12}

// frac renders a fraction the way it is typed. A denominator of one is just
// the number.
func frac(n, d int) string {
	if d == 1 {
		return strconv.Itoa(n)
	}
	return fmt.Sprintf("%d/%d", n, d)
}

// properFraction draws a fraction below one, with a friendly denominator and
// always in lowest terms.
//
// Lowest terms matters more than it looks: every level here marks the answer
// against a reduced fraction, so a question that showed 4/6 would be telling a
// child that 4/6 is a fraction while refusing it as an answer. Show only what
// would be accepted.
func properFraction(r *rand.Rand) (num, den int) {
	for {
		den = friendlyDenominators[r.IntN(len(friendlyDenominators))]
		num = r.IntN(den-1) + 1
		if gcd(num, den) == 1 {
			return num, den
		}
	}
}

var fractionLevels = []Level{
	{
		Title:       "Equivalent Fractions",
		Hint:        "Times the top and the bottom by the same thing and the fraction is unchanged.",
		Questions:   8,
		MinAccuracy: 0.75,
		Gen: func(r *rand.Rand) Question {
			n, d := properFraction(r)
			k := r.IntN(6) + 2
			if r.IntN(2) == 0 {
				return Question{
					Context: fmt.Sprintf("%s = ?/%d", frac(n, d), d*k),
					Prompt:  "?", Answer: Int(n * k),
				}
			}
			return Question{
				Context: fmt.Sprintf("%s = %d/?", frac(n, d), n*k),
				Prompt:  "?", Answer: Int(d * k),
			}
		},
	},
	{
		Title:       "Simplest Form",
		Hint:        "Cancel until nothing goes into the top and the bottom but 1.",
		Questions:   8,
		MinAccuracy: 0.75,
		Gen: func(r *rand.Rand) Question {
			n, d := properFraction(r)
			// Scale a fraction that is already in lowest terms back up, so
			// there is definitely something to cancel.
			k := r.IntN(5) + 2
			return Question{
				Prompt: fmt.Sprintf("%s in simplest form", frac(n*k, d*k)),
				Answer: Fraction(n, d),
			}
		},
	},
	{
		Title:       "Adding Fractions",
		Hint:        "Same bottom number first, then add the tops. Top-heavy answers are fine.",
		Questions:   8,
		MinAccuracy: 0.70,
		Gen: func(r *rand.Rand) Question {
			n1, d1 := properFraction(r)
			n2, d2 := properFraction(r)
			return Question{
				Prompt: fmt.Sprintf("%s + %s", frac(n1, d1), frac(n2, d2)),
				Answer: Fraction(n1*d2+n2*d1, d1*d2),
			}
		},
	},
	{
		Title:       "Taking Fractions Away",
		Hint:        "Same bottom number again. The answer is never below zero, and always in simplest form.",
		Questions:   8,
		MinAccuracy: 0.70,
		Gen: func(r *rand.Rand) Question {
			var n1, d1, n2, d2 int
			for {
				n1, d1 = properFraction(r)
				n2, d2 = properFraction(r)
				// Put the bigger one first. Negative fractions are a topic of
				// their own and this level is not it, and an answer of nought
				// is not a fraction question at all.
				if n1*d2 < n2*d1 {
					n1, d1, n2, d2 = n2, d2, n1, d1
				}
				if n1*d2 != n2*d1 {
					break
				}
			}
			return Question{
				Prompt: fmt.Sprintf("%s - %s", frac(n1, d1), frac(n2, d2)),
				Answer: Fraction(n1*d2-n2*d1, d1*d2),
			}
		},
	},
	{
		Title:       "Multiplying Fractions",
		Hint:        "Tops together, bottoms together, then cancel down. Answer in simplest form.",
		Questions:   8,
		MinAccuracy: 0.75,
		Gen: func(r *rand.Rand) Question {
			n1, d1 := properFraction(r)
			if r.IntN(3) == 0 {
				// A fraction of a whole number: the same operation, and the
				// one that turns up in every word problem.
				whole := d1 * (r.IntN(6) + 2)
				return Question{
					Prompt: fmt.Sprintf("%s of %d", frac(n1, d1), whole),
					Answer: Fraction(n1*whole, d1),
				}
			}
			n2, d2 := properFraction(r)
			return Question{
				Prompt: fmt.Sprintf("%s x %s", frac(n1, d1), frac(n2, d2)),
				Answer: Fraction(n1*n2, d1*d2),
			}
		},
	},
	{
		Title:       "Dividing Fractions",
		Hint:        "Turn the second one upside down and multiply. Answer in simplest form.",
		Questions:   8,
		MinAccuracy: 0.70,
		Gen: func(r *rand.Rand) Question {
			n1, d1 := properFraction(r)
			n2, d2 := properFraction(r)
			// "3/4 / 2/5" has two slashes meaning two different things, so
			// this one is written out in words.
			return Question{
				Prompt: fmt.Sprintf("%s divided by %s", frac(n1, d1), frac(n2, d2)),
				Answer: Fraction(n1*d2, d1*n2),
			}
		},
	},
	{
		Title:       "Mixed Numbers",
		Hint:        "Whole ones times the bottom, add the top. Or the other way round. Simplest form.",
		Questions:   8,
		MinAccuracy: 0.70,
		Gen: func(r *rand.Rand) Question {
			whole := r.IntN(5) + 1
			n, d := properFraction(r)
			top := whole*d + n
			if r.IntN(2) == 0 {
				return Question{
					Prompt: fmt.Sprintf("%d and %s as a top-heavy fraction", whole, frac(n, d)),
					Answer: Fraction(top, d),
				}
			}
			// Going the other way, the whole-number part is the question:
			// the remainder is on screen already.
			return Question{
				Context: fmt.Sprintf("%s = ? and %s", frac(top, d), frac(n, d)),
				Prompt:  "?", Answer: Int(whole),
			}
		},
	},
	{
		Title:       "Which Is Bigger",
		Hint:        "Put them over the same bottom number, then look at the tops. Type the bigger one.",
		Questions:   8,
		MinAccuracy: 0.70,
		Gen: func(r *rand.Rand) Question {
			var n1, d1, n2, d2 int
			for {
				n1, d1 = properFraction(r)
				n2, d2 = properFraction(r)
				// Equal values have no bigger one, and a pair with the same
				// bottom number is not a question.
				if n1*d2 != n2*d1 && d1 != d2 {
					break
				}
			}
			bigNum, bigDen := n1, d1
			if n2*d1 > n1*d2 {
				bigNum, bigDen = n2, d2
			}
			return Question{
				Prompt: fmt.Sprintf("Which is bigger, %s or %s?", frac(n1, d1), frac(n2, d2)),
				Answer: Fraction(bigNum, bigDen),
			}
		},
	},
}
