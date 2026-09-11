package maths

import (
	"fmt"
	"math/rand/v2"
)

// Chapter 9 is square roots: undoing a square, and tidying up what is left
// when it does not undo exactly.
//
// A root is written sqrt(72), which is the form the pseudo-code game uses too
// and the only one that survives a console font. An irrational answer cannot
// be typed into the field, so the levels that meet one ask for it in pieces:
// sqrt(72) is 6 lots of sqrt(2), and both of those are numbers you can type.

// composites are the answers "Multiplying Roots" is built from: every one of
// them splits into a pair of roots more than one way.
var composites = []int{4, 6, 8, 9, 10, 12, 14, 15, 16, 18, 20}

// squares is the table worth knowing, and the one the levels are built from.
var squares = []int{1, 4, 9, 16, 25, 36, 49, 64, 81, 100, 121, 144, 169, 196, 225}

var rootLevels = []Level{
	{
		Title:       "Perfect Squares",
		Hint:        "What times itself makes this? Learn the table up to 15 and most of it is sight-reading.",
		Questions:   12,
		MinAccuracy: 0.75,
		Gen: func(r *rand.Rand) Question {
			n := r.IntN(len(squares)-1) + 2 // 2..15; sqrt(1) is not a question
			if r.IntN(3) == 0 {
				return Question{Prompt: fmt.Sprintf("%d squared", n), Answer: Int(n * n)}
			}
			return Question{Prompt: fmt.Sprintf("sqrt(%d)", n*n), Answer: Int(n)}
		},
	},
	{
		Title:       "Roots Between Whole Numbers",
		Hint:        "Find the square just below it and the square just above.",
		Questions:   10,
		MinAccuracy: 0.70,
		Gen: func(r *rand.Rand) Question {
			// Strictly between two consecutive squares, so there is a whole
			// number either side and neither of them is the answer itself.
			i := r.IntN(len(squares) - 2)
			lo, hi := squares[i], squares[i+1]
			n := r.IntN(hi-lo-1) + lo + 1
			if r.IntN(2) == 0 {
				return Question{
					Prompt: fmt.Sprintf("Biggest whole number below sqrt(%d)?", n),
					Answer: Int(i + 1),
				}
			}
			return Question{
				Prompt: fmt.Sprintf("Smallest whole number above sqrt(%d)?", n),
				Answer: Int(i + 2),
			}
		},
	},
	{
		Title:       "Multiplying Roots",
		Hint:        "sqrt(a) times sqrt(b) is sqrt(a times b). Multiply inside, then take the root.",
		Questions:   12,
		MinAccuracy: 0.70,
		Gen: func(r *rand.Rand) Question {
			// Built backwards from the answer: split its square into the two
			// numbers that go under the roots, so the product is a perfect
			// square and the answer is a whole number the field can hold.
			//
			// The answer is never prime. A prime has only one split, so the
			// question would always come out sqrt(p) x sqrt(p), which is the
			// definition of a square root rather than practice at using one.
			m := composites[r.IntN(len(composites))]
			var factors []int
			for d := 2; d < m*m; d++ {
				if m*m%d == 0 {
					factors = append(factors, d)
				}
			}
			a := factors[r.IntN(len(factors))]
			return Question{
				Prompt: fmt.Sprintf("sqrt(%d) x sqrt(%d)", a, m*m/a),
				Answer: Int(m),
			}
		},
	},
	{
		Title:       "Tidying Roots Up",
		Hint:        "Pull out the biggest square you can find. What is left stays under the root.",
		Questions:   9,
		MinAccuracy: 0.65,
		GenSteps:    simplifyRootSteps,
	},
}

// simplifyRootSteps takes one root apart into a whole number times a smaller
// root, a step at a time:
//
//	sqrt(72)   biggest square dividing 72 -> 36
//	           sqrt(36)                   -> 6
//	           sqrt(72) = 6 x sqrt(?)     -> 2
func simplifyRootSteps(r *rand.Rand) []Question {
	// Pick the answer first: outside times the root of inside, where inside
	// is square-free so "the biggest square" really is the one taken out.
	squareFree := []int{2, 3, 5, 6, 7, 10, 11, 13, 14, 15}
	inside := squareFree[r.IntN(len(squareFree))]
	outside := r.IntN(4) + 2 // 2..5
	square := outside * outside
	n := square * inside

	return []Question{
		{
			Context: fmt.Sprintf("sqrt(%d)", n),
			Prompt:  fmt.Sprintf("Biggest square number that divides %d?", n),
			Answer:  Int(square),
		},
		{
			Context: fmt.Sprintf("sqrt(%d) = sqrt(%d) x sqrt(%d)", n, square, inside),
			Prompt:  fmt.Sprintf("sqrt(%d)", square),
			Answer:  Int(outside),
		},
		{
			Context: fmt.Sprintf("sqrt(%d) = %d x sqrt(?)", n, outside),
			Prompt:  "?", Answer: Int(inside),
		},
	}
}
