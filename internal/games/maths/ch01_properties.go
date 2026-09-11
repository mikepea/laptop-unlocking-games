package maths

import (
	"fmt"
	"math/rand/v2"
	"strconv"
)

// Chapter 1 of the book is the one that looks like it can be skipped and
// cannot. It is about why the rules of arithmetic are what they are: that you
// may reorder a sum to make it easy, that a product can be broken apart and
// put back together, and what happens when you carry on counting past zero.
//
// The levels here are drills for the habits that come out of it. "Clever
// Adding" is not a harder sum than the warm-up, it is the same sum done in an
// order that makes it easy, and the hint says so.

// signed renders a number that is not the first thing on its line. A negative
// one gets brackets, which is how the book writes it and how it stays legible:
// "8 + (-13)" rather than "8 + -13".
func signed(n int) string {
	if n < 0 {
		return fmt.Sprintf("(%d)", n)
	}
	return strconv.Itoa(n)
}

var propertiesLevels = []Level{
	{
		Title:       "Clever Adding",
		Hint:        "Add them in any order you like. Look for two that make a ten.",
		Questions:   10,
		MinAccuracy: 0.75,
		Gen: func(r *rand.Rand) Question {
			// Two of the three terms end in digits that make ten, so the sum
			// is easy once they are put next to each other, and a slog if
			// they are not. They are never next to each other.
			units := r.IntN(9) + 1
			a := (r.IntN(5)+1)*10 + units
			b := (r.IntN(5)+1)*10 + (10 - units)
			c := (r.IntN(8) + 1) * 10
			return Question{
				Prompt: fmt.Sprintf("%d + %d + %d", a, c, b),
				Answer: Int(a + b + c),
			}
		},
	},
	{
		Title:       "Clever Multiplying",
		Hint:        "Multiply them in any order you like. Find the pair that makes a hundred.",
		Questions:   10,
		MinAccuracy: 0.70,
		Gen: func(r *rand.Rand) Question {
			// Same trick, one operation up: a pair that makes a round number,
			// kept apart so that spotting it is the work.
			pairs := [][2]int{{25, 4}, {4, 25}, {2, 50}, {50, 2}, {20, 5}, {5, 20}, {2, 5}, {5, 2}}
			p := pairs[r.IntN(len(pairs))]
			m := r.IntN(18) + 2
			return Question{
				Prompt: fmt.Sprintf("%d x %d x %d", p[0], m, p[1]),
				Answer: Int(p[0] * m * p[1]),
			}
		},
	},
	{
		Title:       "Breaking Numbers Up",
		Hint:        "The distributive property. Split the awkward number, do both halves, put them back.",
		Questions:   9,
		MinAccuracy: 0.70,
		GenSteps:    distributiveSteps,
	},
	{
		Title:       "Below Zero",
		Hint:        "Adding a negative goes down. Taking one away goes up.",
		Questions:   12,
		MinAccuracy: 0.70,
		Gen: func(r *rand.Rand) Question {
			a := r.IntN(12) + 1
			b := r.IntN(12) + 1
			switch r.IntN(4) {
			case 0:
				return Question{Prompt: fmt.Sprintf("-%d + %d", a, b), Answer: Int(b - a)}
			case 1:
				return Question{Prompt: fmt.Sprintf("%d - %d", a, a+b), Answer: Int(-b)}
			case 2:
				return Question{Prompt: fmt.Sprintf("%d + %s", a, signed(-b)), Answer: Int(a - b)}
			default:
				return Question{Prompt: fmt.Sprintf("%d - %s", a, signed(-b)), Answer: Int(a + b)}
			}
		},
	},
	{
		Title:       "Signs That Multiply",
		Hint:        "Two minus signs cancel each other out. One on its own does not.",
		Questions:   12,
		MinAccuracy: 0.70,
		Gen: func(r *rand.Rand) Question {
			a := r.IntN(9) + 2
			b := r.IntN(9) + 2
			// Each of the two numbers is independently negative or not, so
			// all four sign patterns come up, including the plain one.
			sa, sb := 1, 1
			if r.IntN(2) == 0 {
				sa = -1
			}
			if r.IntN(2) == 0 {
				sb = -1
			}
			if r.IntN(2) == 0 {
				return Question{
					Prompt: fmt.Sprintf("%d x %s", sa*a, signed(sb*b)),
					Answer: Int(sa * a * sb * b),
				}
			}
			// Division is built from the product so it is always exact.
			return Question{
				Prompt: fmt.Sprintf("%d / %s", sa*a*b, signed(sb*b)),
				Answer: Int(sa * a / sb),
			}
		},
	},
	{
		Title:       "Order of Operations",
		Hint:        "Brackets first, then times and divide, then plus and minus.",
		Questions:   12,
		MinAccuracy: 0.70,
		Gen: func(r *rand.Rand) Question {
			a := r.IntN(9) + 2
			b := r.IntN(8) + 2
			c := r.IntN(8) + 2
			d := r.IntN(8) + 2
			switch r.IntN(6) {
			case 0:
				return Question{Prompt: fmt.Sprintf("%d + %d x %d", a, b, c), Answer: Int(a + b*c)}
			case 1:
				return Question{Prompt: fmt.Sprintf("(%d + %d) x %d", a, b, c), Answer: Int((a + b) * c)}
			case 2:
				// Written so the division is exact: the dividend is built from
				// the divisor rather than picked and hoped for.
				return Question{Prompt: fmt.Sprintf("%d - %d / %d", a, b*c, c), Answer: Int(a - b)}
			case 3:
				return Question{Prompt: fmt.Sprintf("%d x %d + %d x %d", a, b, c, d), Answer: Int(a*b + c*d)}
			case 4:
				// Both addends are positive and sum to a multiple of c, so
				// the bracket divides exactly and nothing goes below zero.
				split := r.IntN(a*c-1) + 1
				return Question{Prompt: fmt.Sprintf("(%d + %d) / %d", a*c-split, split, c), Answer: Int(a)}
			default:
				return Question{Prompt: fmt.Sprintf("%d + %d x (%d - %d)", a, b, c+d, d), Answer: Int(a + b*c)}
			}
		},
	},
}

// distributiveSteps walks one product through the distributive property, in
// whichever direction the book is teaching it: breaking an awkward multiply
// into two easy ones, or spotting that two products share a factor and
// collecting them back into one.
//
//	7 x 104 = (7 x 100) + (7 x 4)      7 x 100 -> 700, 7 x 4 -> 28, 7 x 104 -> 728
//	13 x 7 + 13 x 3 = 13 x (7 + 3)     7 + 3   -> 10,  13 x 10 -> 130
func distributiveSteps(r *rand.Rand) []Question {
	a := r.IntN(8) + 2 // 2..9, small enough to do in your head

	if r.IntN(2) == 0 {
		// Split: a near-round number, either side of the round one.
		base := (r.IntN(3) + 1) * 100 // 100, 200, 300
		off := r.IntN(8) + 1
		if r.IntN(2) == 0 {
			whole := base + off
			return []Question{
				{Context: fmt.Sprintf("%d x %d = (%d x %d) + (%d x %d)", a, whole, a, base, a, off),
					Prompt: fmtQ(a, "x", base), Answer: Int(a * base)},
				{Context: fmt.Sprintf("%d x %d = (%d x %d) + (%d x %d)", a, whole, a, base, a, off),
					Prompt: fmtQ(a, "x", off), Answer: Int(a * off)},
				{Context: fmt.Sprintf("%d + %d", a*base, a*off),
					Prompt: fmtQ(a, "x", whole), Answer: Int(a * whole)},
			}
		}
		whole := base - off
		return []Question{
			{Context: fmt.Sprintf("%d x %d = (%d x %d) - (%d x %d)", a, whole, a, base, a, off),
				Prompt: fmtQ(a, "x", base), Answer: Int(a * base)},
			{Context: fmt.Sprintf("%d x %d = (%d x %d) - (%d x %d)", a, whole, a, base, a, off),
				Prompt: fmtQ(a, "x", off), Answer: Int(a * off)},
			{Context: fmt.Sprintf("%d - %d", a*base, a*off),
				Prompt: fmtQ(a, "x", whole), Answer: Int(a * whole)},
		}
	}

	// Collect: two products sharing a factor, where the other two make a ten.
	b := r.IntN(9) + 1
	c := 10*(r.IntN(2)+1) - b
	m := r.IntN(15) + 4 // 4..18, awkward on purpose
	scattered := fmt.Sprintf("%d x %d + %d x %d", m, b, m, c)
	return []Question{
		{Context: fmt.Sprintf("%s = %d x (%d + %d)", scattered, m, b, c),
			Prompt: fmtQ(b, "+", c), Answer: Int(b + c)},
		{Context: scattered,
			Prompt: fmt.Sprintf("%d x (%d + %d)", m, b, c), Answer: Int(m * (b + c))},
	}
}
