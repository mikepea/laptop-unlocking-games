package maths

import "math/rand/v2"

// Question is one thing to answer.
type Question struct {
	// Context is optional working shown above the question: the equation being
	// solved, the expression being collected up. Empty for plain arithmetic,
	// where the prompt says everything.
	Context string
	// Prompt is shown without the answer, e.g. "7 x 8" or "x". A prompt ending
	// in "?" is a sentence and is rendered as one; anything else is treated as
	// the left-hand side of an equals sign.
	Prompt string
	Answer int
}

// Level is a set of questions of one kind. Levels are cleared in order.
type Level struct {
	Title string
	Hint  string
	// Questions is how many are asked in a round.
	Questions int
	// MinAccuracy is the share that must be right to pass.
	MinAccuracy float64
	// Gen makes one question. It is called once per question in a round.
	Gen func(r *rand.Rand) Question
	// GenSteps makes one multi-step problem, as the sequence of questions that
	// walks through it. Levels set either this or Gen, never both. A round
	// never splits a problem across its end, so a level using this can run a
	// question or two past Questions rather than leave working half-done.
	GenSteps func(r *rand.Rand) []Question
}

// Levels is the curriculum, easiest first: the number facts, then algebra
// built on top of them. Times tables and division stay at the front
// deliberately — an easy win to start on is worth more than a tight syllabus.
var Levels = append(append([]Level{}, arithmeticLevels...), algebraLevels...)

// arithmeticLevels are the number facts. Every prompt here is a plain
// "a op b" sum, which is a promise round_test.go checks.
var arithmeticLevels = []Level{
	{
		Title:       "Adding Up",
		Hint:        "Sums up to 20. Count on from the bigger number.",
		Questions:   12,
		MinAccuracy: 0.75,
		Gen: func(r *rand.Rand) Question {
			a := r.IntN(10) + 1
			b := r.IntN(10) + 1
			return Question{Prompt: fmtQ(a, "+", b), Answer: a + b}
		},
	},
	{
		Title:       "Taking Away",
		Hint:        "The answer is never below zero. Count back.",
		Questions:   12,
		MinAccuracy: 0.75,
		Gen: func(r *rand.Rand) Question {
			a := r.IntN(19) + 2
			b := r.IntN(a) + 1
			return Question{Prompt: fmtQ(a, "-", b), Answer: a - b}
		},
	},
	{
		Title:       "Twos, Fives and Tens",
		Hint:        "The friendly times tables. Learn these and the rest follow.",
		Questions:   12,
		MinAccuracy: 0.80,
		Gen:         timesTable(2, 5, 10),
	},
	{
		Title:       "Threes, Fours and Sixes",
		Hint:        "Six is just double three. Four is double two.",
		Questions:   12,
		MinAccuracy: 0.80,
		Gen:         timesTable(3, 4, 6),
	},
	{
		Title:       "Sevens, Eights and Nines",
		Hint:        "The hard ones. Nine times a number is ten times it, take one lot away.",
		Questions:   12,
		MinAccuracy: 0.75,
		Gen:         timesTable(7, 8, 9),
	},
	{
		Title:       "Sharing Out",
		Hint:        "Division. Ask yourself what times what makes the big number.",
		Questions:   12,
		MinAccuracy: 0.75,
		Gen: func(r *rand.Rand) Question {
			b := r.IntN(9) + 2  // 2..10, never divide by one
			n := r.IntN(10) + 1 // 1..10
			return Question{Prompt: fmtQ(b*n, "/", b), Answer: n}
		},
	},
	{
		Title:       "Everything At Once",
		Hint:        "All four, mixed up. Read each one carefully.",
		Questions:   15,
		MinAccuracy: 0.75,
		Gen: func(r *rand.Rand) Question {
			switch r.IntN(4) {
			case 0:
				a, b := r.IntN(20)+1, r.IntN(20)+1
				return Question{Prompt: fmtQ(a, "+", b), Answer: a + b}
			case 1:
				a := r.IntN(29) + 2
				b := r.IntN(a) + 1
				return Question{Prompt: fmtQ(a, "-", b), Answer: a - b}
			case 2:
				a, b := r.IntN(12)+1, r.IntN(12)+1
				return Question{Prompt: fmtQ(a, "x", b), Answer: a * b}
			default:
				b := r.IntN(11) + 2
				n := r.IntN(12) + 1
				return Question{Prompt: fmtQ(b*n, "/", b), Answer: n}
			}
		},
	},
}

// timesTable returns a generator drawing the second operand from the given
// tables and the first from 1..12.
func timesTable(tables ...int) func(*rand.Rand) Question {
	return func(r *rand.Rand) Question {
		a := r.IntN(12) + 1
		b := tables[r.IntN(len(tables))]
		// Half the time, put the table first. Knowing 8 x 3 is not the same
		// skill as knowing 3 x 8 until it is.
		if r.IntN(2) == 0 {
			a, b = b, a
		}
		return Question{Prompt: fmtQ(a, "x", b), Answer: a * b}
	}
}

// LevelUnlocked reports whether level i is open given how many are cleared.
func LevelUnlocked(i, cleared int) bool { return i <= cleared }
