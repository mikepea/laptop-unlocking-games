package maths

import (
	"fmt"
	"math/rand/v2"
	"strings"
)

// Chapter 3 is number theory: which numbers go into which. It is the chapter
// that makes chapter 4 possible, because cancelling a fraction down is a
// question about common factors wearing a disguise.
//
// The prompts here are sentences rather than expressions. "What is the
// smallest multiple of 7 above 50?" cannot be written as a sum, and writing
// it as one would hide what is being asked.

// primesUnder100 is the sieve, written out. It is short, it never changes, and
// a table is easier to be sure of than a generator.
var primesUnder100 = []int{
	2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41, 43, 47,
	53, 59, 61, 67, 71, 73, 79, 83, 89, 97,
}

// isPrime is trial division. It is used to build questions, and the tests
// check the questions against a sieve built a different way.
func isPrime(n int) bool {
	if n < 2 {
		return false
	}
	for d := 2; d*d <= n; d++ {
		if n%d == 0 {
			return false
		}
	}
	return true
}

// digitSum adds up the digits of a number, which is the whole of the
// divisibility tests for 3 and 9.
func digitSum(n int) int {
	n = abs(n)
	sum := 0
	for n > 0 {
		sum += n % 10
		n /= 10
	}
	return sum
}

var numberTheoryLevels = []Level{
	{
		Title:       "Multiples",
		Hint:        "The multiples of 7 are 7, 14, 21 and on. Count in sevens.",
		Questions:   10,
		MinAccuracy: 0.75,
		Gen: func(r *rand.Rand) Question {
			n := r.IntN(10) + 3 // 3..12
			switch r.IntN(3) {
			case 0:
				above := (r.IntN(9) + 3) * 10 // 30..110
				next := (above/n + 1) * n
				return Question{
					Prompt: fmt.Sprintf("Smallest multiple of %d above %d?", n, above),
					Answer: Int(next),
				}
			case 1:
				top := (r.IntN(9) + 2) * 10 // 20..100
				return Question{
					Prompt: fmt.Sprintf("How many multiples of %d from 1 to %d?", n, top),
					Answer: Int(top / n),
				}
			default:
				k := r.IntN(8) + 2
				return Question{
					Prompt: fmt.Sprintf("What is the %s multiple of %d?", ordinal(k), n),
					Answer: Int(k * n),
				}
			}
		},
	},
	{
		Title:       "Divisibility Tests",
		Hint:        "Add the digits up. What that leaves over dividing by 3 or 9, the number leaves too.",
		Questions:   10,
		MinAccuracy: 0.70,
		GenSteps:    divisibilitySteps,
	},
	{
		Title:       "Primes",
		Hint:        "A prime has no divisors but itself and 1. Check 2, 3, 5, 7 and stop when they square past it.",
		Questions:   10,
		MinAccuracy: 0.70,
		Gen: func(r *rand.Rand) Question {
			switch r.IntN(3) {
			case 0:
				// Never the last one in the table: there has to be a next.
				i := r.IntN(len(primesUnder100) - 1)
				return Question{
					Prompt: fmt.Sprintf("What is the next prime after %d?", primesUnder100[i]),
					Answer: Int(primesUnder100[i+1]),
				}
			case 1:
				lo := (r.IntN(8) + 1) * 10 // 10..80
				hi := lo + 20
				count := 0
				for _, p := range primesUnder100 {
					if p > lo && p < hi {
						count++
					}
				}
				return Question{
					Prompt: fmt.Sprintf("How many primes are there between %d and %d?", lo, hi),
					Answer: Int(count),
				}
			default:
				// The smallest prime factor of a composite: the first thing
				// you look for, and the first step of a factorisation.
				n := 0
				for n < 4 || isPrime(n) {
					n = r.IntN(96) + 4
				}
				smallest := 2
				for n%smallest != 0 {
					smallest++
				}
				return Question{
					Prompt: fmt.Sprintf("What is the smallest prime that divides %d?", n),
					Answer: Int(smallest),
				}
			}
		},
	},
	{
		Title:       "Prime Factorisation",
		Hint:        "Split it until every piece is prime. A prime that is not in there at all appears nought times.",
		Questions:   9,
		MinAccuracy: 0.70,
		GenSteps:    factorisationSteps,
	},
	{
		Title:       "Counting Divisors",
		Hint:        "Add one to each power and multiply those together. 2^3 x 3^2 has 4 x 3 divisors.",
		Questions:   10,
		MinAccuracy: 0.70,
		Gen: func(r *rand.Rand) Question {
			e2, e3 := r.IntN(4)+1, r.IntN(3)+1
			n := pow(2, e2) * pow(3, e3)
			return Question{
				Context: fmt.Sprintf("%d = %s x %s", n, power(2, e2), power(3, e3)),
				Prompt:  fmt.Sprintf("How many divisors does %d have?", n),
				Answer:  Int((e2 + 1) * (e3 + 1)),
			}
		},
	},
	{
		Title:       "Biggest Common Factor",
		Hint:        "What goes into both? Take the primes they share.",
		Questions:   10,
		MinAccuracy: 0.70,
		Gen: func(r *rand.Rand) Question {
			// Built from a shared factor and two coprime halves, so the
			// answer is genuinely the biggest one and not merely a common one.
			g := r.IntN(11) + 2
			var p, q int
			for {
				p, q = r.IntN(9)+2, r.IntN(9)+2
				if p != q && gcd(p, q) == 1 {
					break
				}
			}
			return Question{
				Prompt: fmt.Sprintf("Biggest number that divides both %d and %d?", g*p, g*q),
				Answer: Int(g),
			}
		},
	},
	{
		Title:       "Smallest Common Multiple",
		Hint:        "Count up in both until they land on the same number.",
		Questions:   10,
		MinAccuracy: 0.70,
		Gen: func(r *rand.Rand) Question {
			// Two different numbers: "both 8 and 8" is not a question.
			a := r.IntN(10) + 3
			b := a
			for b == a {
				b = r.IntN(10) + 3
			}
			return Question{
				Prompt: fmt.Sprintf("Smallest number that both %d and %d divide into?", a, b),
				Answer: Int(lcm(a, b)),
			}
		},
	},
}

// ordinal writes 2 as "2nd". Only the small ones the levels use.
func ordinal(n int) string {
	suffix := "th"
	switch {
	case n%100 >= 11 && n%100 <= 13:
	case n%10 == 1:
		suffix = "st"
	case n%10 == 2:
		suffix = "nd"
	case n%10 == 3:
		suffix = "rd"
	}
	return fmt.Sprintf("%d%s", n, suffix)
}

// divisibilitySteps walks one number through the test for 9, which is also
// the test for 3: add the digits, and whatever that leaves over, the number
// leaves over too.
//
//	3746   digits add to -> 20    remainder on dividing by 9 -> 2
func divisibilitySteps(r *rand.Rand) []Question {
	n := r.IntN(8000) + 1000 // four digits, so adding them up is real work
	sum := digitSum(n)
	by := 9
	if r.IntN(2) == 0 {
		by = 3
	}
	return []Question{
		{
			Context: fmt.Sprintf("%d", n),
			Prompt:  "What do the digits add up to?",
			Answer:  Int(sum),
		},
		{
			Context: fmt.Sprintf("%d, digits adding to %d", n, sum),
			Prompt:  fmt.Sprintf("What is the remainder when %d is divided by %d?", n, by),
			Answer:  Int(n % by),
		},
	}
}

// factorisationSteps takes one number apart into its primes, asking for each
// power in turn. A prime that is not a factor is asked for too: the answer is
// nought, and knowing that is knowing what the question means.
func factorisationSteps(r *rand.Rand) []Question {
	primes := []int{2, 3, 5}
	var exps []int
	n := 0
	// Reject the numbers that make a dull question: a prime power on its own,
	// and anything small enough to see at a glance.
	for n < 24 || countNonZero(exps) < 2 {
		exps = []int{r.IntN(4), r.IntN(3), r.IntN(2)}
		n = 1
		for i, p := range primes {
			n *= pow(p, exps[i])
		}
	}

	// Each step carries the working so far, so by the last one the whole
	// factorisation is on screen and the round ends on something finished.
	var found []string
	qs := make([]Question, 0, len(primes))
	for i, p := range primes {
		ctx := fmt.Sprintf("%d", n)
		if len(found) > 0 {
			ctx = fmt.Sprintf("%d   so far: %s", n, strings.Join(found, " x "))
		}
		qs = append(qs, Question{
			Context: ctx,
			Prompt:  fmt.Sprintf("How many %ds in the prime factorisation?", p),
			Answer:  Int(exps[i]),
		})
		if exps[i] > 0 {
			found = append(found, power(p, exps[i]))
		}
	}
	return qs
}

func countNonZero(xs []int) int {
	n := 0
	for _, x := range xs {
		if x > 0 {
			n++
		}
	}
	return n
}
