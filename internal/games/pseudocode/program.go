package pseudocode

import (
	"fmt"
	"math/rand/v2"
)

// Program is one puzzle: a few lines of code, and the number they print.
type Program struct {
	Lines []string
	// Answer is what the print line puts on the screen.
	Answer int
}

// Level is a set of programs of one kind. Levels are cleared in order.
type Level struct {
	Title string
	Hint  string
	// Questions is how many programs are traced in a round.
	Questions int
	// MinAccuracy is the share that must be right to pass.
	MinAccuracy float64
	// Gen makes one program. It is called once per question in a round.
	Gen func(r *rand.Rand) Program
}

// The language is four statements wide and never nests more than one level
// deep, because it is here to teach one idea at a time:
//
//	x = 5              put a number in a box
//	y = x + 3          work something out from a box
//	repeat 3 times:    do the indented line that many times
//	if x > 10:         do the indented line only if this is true
//	print x            show what is in the box
//
// Multiplication is written with a * on purpose. Maths Sprint writes 3x, and
// meeting both forms is worth something: they are the same idea, and knowing
// that is most of the bridge between the two games.

const indent = "  "

func assign(v string, n int) string     { return fmt.Sprintf("%s = %d", v, n) }
func op(dst, a, o string, b int) string { return fmt.Sprintf("%s = %s %s %d", dst, a, o, b) }
func printVar(v string) string          { return "print " + v }
func repeat(n int) string               { return fmt.Sprintf("repeat %d times:", n) }
func ifCmp(v, cmp string, n int) string { return fmt.Sprintf("if %s %s %d:", v, cmp, n) }
func body(line string) string           { return indent + line }

// apply is the generator working out its own answer as it builds the program.
// The test traces the rendered lines with a separate interpreter and checks
// the two agree, which is what catches a program that shows one thing and
// marks another.
func apply(a int, o string, b int) int {
	switch o {
	case "+":
		return a + b
	case "-":
		return a - b
	case "*":
		return a * b
	}
	panic("unknown operator " + o)
}

// Levels is the curriculum, easiest first.
var Levels = []Level{
	{
		Title:       "Boxes",
		Hint:        "A name, an equals sign, and a number goes in the box. print shows what is in it.",
		Questions:   10,
		MinAccuracy: 0.80,
		Gen: func(r *rand.Rand) Program {
			x := r.IntN(20) + 1
			y := r.IntN(20) + 1
			// Two boxes, so the answer depends on reading which one is printed
			// rather than on finding the only number on screen.
			if r.IntN(2) == 0 {
				return Program{Lines: []string{assign("x", x), assign("y", y), printVar("x")}, Answer: x}
			}
			return Program{Lines: []string{assign("x", x), assign("y", y), printVar("y")}, Answer: y}
		},
	},
	{
		Title:       "Working It Out",
		Hint:        "The right-hand side is worked out first, then it goes in the box on the left.",
		Questions:   10,
		MinAccuracy: 0.75,
		Gen: func(r *rand.Rand) Program {
			x := r.IntN(12) + 2
			o := []string{"+", "-", "*"}[r.IntN(3)]
			b := r.IntN(9) + 1
			if o == "-" {
				b = r.IntN(x) + 1 // no negative results at this level
			}
			if o == "*" {
				b = r.IntN(5) + 2
			}
			return Program{
				Lines:  []string{assign("x", x), op("y", "x", o, b), printVar("y")},
				Answer: apply(x, o, b),
			}
		},
	},
	{
		Title:       "Changing Your Mind",
		Hint:        "x = x + 2 is not a sum to solve. Take what is in x, add 2, put it back.",
		Questions:   12,
		MinAccuracy: 0.70,
		Gen: func(r *rand.Rand) Program {
			// The level this game exists for. In algebra "x = x + 2" is simply
			// false; here it is the most ordinary line there is, and that
			// collision is the thing worth spending a whole level on.
			x := r.IntN(12) + 3
			o := []string{"+", "-", "*"}[r.IntN(3)]
			b := r.IntN(8) + 1
			if o == "-" {
				b = r.IntN(x) + 1
			}
			if o == "*" {
				b = r.IntN(4) + 2
			}
			after := apply(x, o, b)

			// Sometimes change it twice, so the answer cannot be got by
			// treating the second line as the whole program.
			if r.IntN(2) == 0 {
				c := r.IntN(6) + 1
				return Program{
					Lines: []string{
						assign("x", x), op("x", "x", o, b), op("x", "x", "+", c), printVar("x"),
					},
					Answer: after + c,
				}
			}
			return Program{
				Lines:  []string{assign("x", x), op("x", "x", o, b), printVar("x")},
				Answer: after,
			}
		},
	},
	{
		Title:       "Two Boxes",
		Hint:        "Changing x after y was worked out does not go back and change y.",
		Questions:   10,
		MinAccuracy: 0.70,
		Gen: func(r *rand.Rand) Program {
			// The trap is deliberate: y is computed from x, then x moves. A
			// player who thinks of x as "the answer" rather than "a box right
			// now" will print the wrong one.
			x := r.IntN(10) + 2
			b := r.IntN(8) + 1
			c := r.IntN(6) + 1
			y := x + b
			return Program{
				Lines: []string{
					assign("x", x), op("y", "x", "+", b), op("x", "x", "+", c), printVar("y"),
				},
				Answer: y,
			}
		},
	},
	{
		Title:       "Round and Round",
		Hint:        "The indented line runs that many times. Count them off one at a time.",
		Questions:   10,
		MinAccuracy: 0.70,
		Gen: func(r *rand.Rand) Program {
			x := r.IntN(10) + 1
			n := r.IntN(4) + 2 // 2..5 times
			o := "+"
			b := r.IntN(6) + 1
			if r.IntN(3) == 0 {
				o = "*"
				b = 2 // doubling is the only multiply that stays small enough to trace
			}
			total := x
			for i := 0; i < n; i++ {
				total = apply(total, o, b)
			}
			return Program{
				Lines: []string{
					assign("x", x), repeat(n), body(op("x", "x", o, b)), printVar("x"),
				},
				Answer: total,
			}
		},
	},
	{
		Title:       "Only If",
		Hint:        "Check whether it is true first. If it is not, the indented line never happens.",
		Questions:   10,
		MinAccuracy: 0.70,
		Gen: func(r *rand.Rand) Program {
			x := r.IntN(20) + 1
			threshold := r.IntN(20) + 1
			cmp := []string{">", "<"}[r.IntN(2)]
			b := r.IntN(9) + 1

			taken := (cmp == ">" && x > threshold) || (cmp == "<" && x < threshold)
			answer := x
			if taken {
				answer = x + b
			}
			return Program{
				Lines: []string{
					assign("x", x), ifCmp("x", cmp, threshold), body(op("x", "x", "+", b)), printVar("x"),
				},
				Answer: answer,
			}
		},
	},
}

// LevelUnlocked reports whether level i is open given how many are cleared.
func LevelUnlocked(i, cleared int) bool { return i <= cleared }
