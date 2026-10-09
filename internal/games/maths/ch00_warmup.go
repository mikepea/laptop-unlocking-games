package maths

import "math/rand/v2"

// The warm-up is not in the book. It is the number facts the book assumes you
// already have, and it stays at the front deliberately -- an easy win to start
// on is worth more than a tight syllabus. Plain adding and taking away were
// dropped as too easy, and the times tables are one level that leans on the
// ones that need the practice.
//
// Every prompt here is a plain "a op b" sum, which is a promise round_test.go
// checks.
var warmUpLevels = []Level{
	{
		Title:       "Times Tables",
		Hint:        "All of them, but mostly the sixes, sevens, elevens and twelves.",
		Questions:   8,
		MinAccuracy: 0.75,
		Gen:         timesTable(tableWeights),
	},
	{
		Title:       "Sharing Out",
		Hint:        "Division. Ask yourself what times what makes the big number.",
		Questions:   8,
		MinAccuracy: 0.75,
		Gen: func(r *rand.Rand) Question {
			b := r.IntN(9) + 2  // 2..10, never divide by one
			n := r.IntN(10) + 1 // 1..10
			return Question{Prompt: fmtQ(b*n, "/", b), Answer: Int(n)}
		},
	},
	{
		Title:       "Everything At Once",
		Hint:        "All four, mixed up. Read each one carefully.",
		Questions:   8,
		MinAccuracy: 0.75,
		Gen: func(r *rand.Rand) Question {
			switch r.IntN(4) {
			case 0:
				a, b := r.IntN(20)+1, r.IntN(20)+1
				return Question{Prompt: fmtQ(a, "+", b), Answer: Int(a + b)}
			case 1:
				a := r.IntN(29) + 2
				b := r.IntN(a) + 1
				return Question{Prompt: fmtQ(a, "-", b), Answer: Int(a - b)}
			case 2:
				a, b := r.IntN(12)+1, r.IntN(12)+1
				return Question{Prompt: fmtQ(a, "x", b), Answer: Int(a * b)}
			default:
				b := r.IntN(11) + 2
				n := r.IntN(12) + 1
				return Question{Prompt: fmtQ(b*n, "/", b), Answer: Int(n)}
			}
		},
	},
}

// tableWeights is how often each table is drawn, relative to the others. The
// sixes, sevens, elevens and twelves come up three times as often as the rest,
// so they make up about two thirds of a round.
var tableWeights = map[int]int{
	2: 1, 3: 1, 4: 1, 5: 1, 6: 3, 7: 3, 8: 1, 9: 1, 10: 1, 11: 3, 12: 3,
}

// timesTable returns a generator drawing the table from weights and the other
// operand from 1..12.
func timesTable(weights map[int]int) func(*rand.Rand) Question {
	// Expand into a flat pool once, in table order so a seeded rand is
	// deterministic (map iteration order is not).
	var pool []int
	for t := 1; t <= 12; t++ {
		for range weights[t] {
			pool = append(pool, t)
		}
	}
	return func(r *rand.Rand) Question {
		a := r.IntN(12) + 1
		b := pool[r.IntN(len(pool))]
		// Half the time, put the table first. Knowing 8 x 3 is not the same
		// skill as knowing 3 x 8 until it is.
		if r.IntN(2) == 0 {
			a, b = b, a
		}
		return Question{Prompt: fmtQ(a, "x", b), Answer: Int(a * b)}
	}
}
