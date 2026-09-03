package maths

import (
	"fmt"
	"math/rand/v2"
	"strings"
)

// The algebra half of the curriculum. It sits on top of the number facts, but
// deliberately keeps the numbers small: the skill being learned is what a
// letter means, not arithmetic under pressure.
//
// Every answer is still an int. That is the constraint that keeps the whole
// game — one field, one number, instant marking — working unchanged, and it is
// why "simplify this" is asked as a short sequence of steps rather than as one
// box you type "6(x + 3y)" into. Each step is a win on its own, which is the
// same reason a round is twelve short questions instead of one long one.
//
// Notation is the conventional 3x rather than 3*x: this is the form he will
// meet at school. The pseudo-code game is where the * form belongs, and seeing
// both is worth something.

// term renders a coefficient against a variable the way it is written by hand:
// 1x is x, -1x is -x, everything else keeps its number.
func term(n int, v string) string {
	switch n {
	case 1:
		return v
	case -1:
		return "-" + v
	default:
		return fmt.Sprintf("%d%s", n, v)
	}
}

// expr joins terms with the plus signs written out, e.g. "3x + 18y".
func expr(parts ...string) string { return strings.Join(parts, " + ") }

// gcd is Euclid, used to check a factorisation is the fullest one available.
func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	if a < 0 {
		return -a
	}
	return a
}

var algebraLevels = []Level{
	{
		Title:       "Missing Number",
		Hint:        "The same sums as before, with a hole in them. What fills it?",
		Questions:   12,
		MinAccuracy: 0.75,
		Gen: func(r *rand.Rand) Question {
			// Addition and subtraction only, and the missing number moves
			// around: always putting the hole in the same place teaches the
			// position rather than the idea.
			a := r.IntN(12) + 2
			b := r.IntN(12) + 2
			switch r.IntN(3) {
			case 0:
				return Question{Context: fmt.Sprintf("%d + ? = %d", a, a+b), Prompt: "?", Answer: b}
			case 1:
				return Question{Context: fmt.Sprintf("? + %d = %d", b, a+b), Prompt: "?", Answer: a}
			default:
				return Question{Context: fmt.Sprintf("%d - ? = %d", a+b, a), Prompt: "?", Answer: b}
			}
		},
	},
	{
		Title:       "One Step",
		Hint:        "x is just the missing number wearing a letter. Undo what was done to it.",
		Questions:   12,
		MinAccuracy: 0.75,
		Gen: func(r *rand.Rand) Question {
			x := r.IntN(12) + 1
			switch r.IntN(4) {
			case 0:
				b := r.IntN(10) + 1
				return Question{Context: fmt.Sprintf("x + %d = %d", b, x+b), Prompt: "x", Answer: x}
			case 1:
				// Take away no more than there is: x - 9 = -2 drags negative
				// numbers into a level that is only meant to be teaching what
				// a letter stands for.
				b := r.IntN(x) + 1
				return Question{Context: fmt.Sprintf("x - %d = %d", b, x-b), Prompt: "x", Answer: x}
			case 2:
				a := r.IntN(9) + 2
				return Question{Context: fmt.Sprintf("%s = %d", term(a, "x"), a*x), Prompt: "x", Answer: x}
			default:
				a := r.IntN(9) + 2
				return Question{Context: fmt.Sprintf("x / %d = %d", a, x), Prompt: "x", Answer: a * x}
			}
		},
	},
	{
		Title:       "Two Steps",
		Hint:        "Take the loose number off first, then undo the times.",
		Questions:   10,
		MinAccuracy: 0.70,
		Gen: func(r *rand.Rand) Question {
			x := r.IntN(10) + 1
			a := r.IntN(5) + 2 // 2..6
			b := r.IntN(12) + 1
			if r.IntN(2) == 0 {
				// The adding case can use any b; the taking-away case below
				// cannot, or the right-hand side goes negative.
				return Question{
					Context: fmt.Sprintf("%s + %d = %d", term(a, "x"), b, a*x+b),
					Prompt:  "x", Answer: x,
				}
			}
			b = r.IntN(a*x) + 1
			return Question{
				Context: fmt.Sprintf("%s - %d = %d", term(a, "x"), b, a*x-b),
				Prompt:  "x", Answer: x,
			}
		},
	},
	{
		Title:       "Putting Numbers In",
		Hint:        "You are told what x is. Swap it in and work it out.",
		Questions:   12,
		MinAccuracy: 0.75,
		Gen: func(r *rand.Rand) Question {
			x := r.IntN(9) + 2
			a := r.IntN(6) + 2
			b := r.IntN(10) + 1
			if r.IntN(2) == 0 {
				return Question{
					Context: fmt.Sprintf("x = %d", x),
					Prompt:  fmt.Sprintf("%s + %d", term(a, "x"), b), Answer: a*x + b,
				}
			}
			y := r.IntN(7) + 2
			c := r.IntN(5) + 2
			return Question{
				Context: fmt.Sprintf("x = %d, y = %d", x, y),
				Prompt:  expr(term(a, "x"), term(c, "y")), Answer: a*x + c*y,
			}
		},
	},
	{
		Title:       "Collecting Up",
		Hint:        "Add up the xs. They are all the same thing, so they go together.",
		Questions:   12,
		MinAccuracy: 0.75,
		Gen: func(r *rand.Rand) Question {
			a := r.IntN(8) + 1
			b := r.IntN(8) + 1
			// A y in the middle makes the point that only like terms collect.
			if r.IntN(2) == 0 {
				c := r.IntN(6) + 1
				return Question{
					Context: fmt.Sprintf("%s = ?x", expr(term(a, "x"), term(c, "y"), term(b, "x"))),
					Prompt:  "?", Answer: a + b,
				}
			}
			return Question{
				Context: fmt.Sprintf("%s = ?x", expr(term(a, "x"), term(b, "x"))),
				Prompt:  "?", Answer: a + b,
			}
		},
	},
	{
		Title:       "Simplify and Factorise",
		Hint:        "Collect the letters up, then find what divides into both.",
		Questions:   8, // two problems of four steps
		MinAccuracy: 0.70,
		GenSteps:    simplifySteps,
	},
}

// simplifySteps walks one expression from four scattered terms to a factorised
// form, as four questions that each want a single number:
//
//	6y + 3x + 12y + 3x     how many x?          -> 6
//	                       how many y?          -> 18
//	6x + 18y               what divides both?   -> 6
//	6x + 18y = 6(x + ?y)   what is ?            -> 3
//
// Built backwards from the factorisation so it always comes out whole: pick
// the common factor and what is left inside the brackets, multiply up, then
// break the totals into two terms each.
func simplifySteps(r *rand.Rand) []Question {
	g := r.IntN(5) + 2 // common factor, 2..6

	// What is left inside the brackets must share no factor of its own.
	// 6(2x + 4y) is really 12(x + 2y), so step three would be asking for the
	// biggest common factor while marking a smaller one correct -- the game
	// would tell him he was wrong when he was right, which is the worst thing
	// a game like this can do. Also reject ix == iy, since 6(x + y) hides
	// which number came from where.
	var ix, iy int
	for {
		ix, iy = r.IntN(4)+1, r.IntN(4)+1
		if ix != iy && gcd(ix, iy) == 1 {
			break
		}
	}
	totalX, totalY := g*ix, g*iy

	// Split each total into the two terms that will be scattered through the
	// expression. Both totals are at least 2, so both splits have room.
	a := r.IntN(totalX-1) + 1
	d := totalX - a
	b := r.IntN(totalY-1) + 1
	c := totalY - b

	// Shuffled, so collecting is real work rather than reading left to right.
	parts := []string{term(a, "x"), term(b, "y"), term(c, "y"), term(d, "x")}
	r.Shuffle(len(parts), func(i, j int) { parts[i], parts[j] = parts[j], parts[i] })

	scattered := expr(parts...)
	collected := expr(term(totalX, "x"), term(totalY, "y"))

	return []Question{
		{Context: scattered, Prompt: "How many x altogether?", Answer: totalX},
		{Context: scattered, Prompt: "How many y altogether?", Answer: totalY},
		{Context: collected, Prompt: "What is the biggest number that divides both?", Answer: g},
		{
			Context: fmt.Sprintf("%s = %d(%s + ?y)", collected, g, term(ix, "x")),
			Prompt:  "?", Answer: iy,
		},
	}
}
