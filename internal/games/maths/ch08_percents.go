package maths

import (
	"fmt"
	"math/rand/v2"
)

// Chapter 8 is percents, which the book is careful to present as nothing new:
// a percent is a hundredth, so every question here is a fraction question or
// a decimal question that has been given a different hat.
//
// Every number is chosen so the answer comes out whole. A percent question
// that lands on 13.333 teaches rounding, which is chapter 6, and hides
// whether the percent was understood.

// wholePercents are the ones a child does in their head: multiples of five,
// nothing at either extreme.
func wholePercent(r *rand.Rand) int { return (r.IntN(18) + 1) * 5 } // 5..90

var percentLevels = []Level{
	{
		Title:       "Percent of a Number",
		Hint:        "Ten percent is a tenth. Build the rest out of tenths and halves of them.",
		Questions:   8,
		MinAccuracy: 0.75,
		Gen: func(r *rand.Rand) Question {
			pct := wholePercent(r)
			// A multiple of twenty, so a multiple of five percent of it is
			// always a whole number.
			n := (r.IntN(10) + 1) * 20
			return Question{
				Prompt: fmt.Sprintf("%d%% of %d", pct, n),
				Answer: Int(pct * n / 100),
			}
		},
	},
	{
		Title:       "Percents, Decimals and Fractions",
		Hint:        "Percent means out of a hundred. That is all it means.",
		Questions:   8,
		MinAccuracy: 0.75,
		Gen: func(r *rand.Rand) Question {
			pct := wholePercent(r)
			switch r.IntN(3) {
			case 0:
				return Question{
					Prompt: fmt.Sprintf("%d%% as a decimal", pct),
					Answer: Decimal(pct, 2),
				}
			case 1:
				return Question{
					Prompt: fmt.Sprintf("%s as a percent", formatDecimal(pct, 2)),
					Answer: Int(pct),
				}
			default:
				num, den := pct, 100
				if g := gcd(num, den); g > 1 {
					num, den = num/g, den/g
				}
				return Question{
					Prompt: fmt.Sprintf("%d%% as a fraction in simplest form", pct),
					Answer: Fraction(num, den),
				}
			}
		},
	},
	{
		Title:       "Up and Down by a Percent",
		Hint:        "Work out the change first, then add it on or take it off.",
		Questions:   8,
		MinAccuracy: 0.70,
		Gen: func(r *rand.Rand) Question {
			n := (r.IntN(10) + 1) * 20
			pct := wholePercent(r)
			change := pct * n / 100
			if r.IntN(2) == 0 {
				return Question{
					Context: fmt.Sprintf("A comic costs %dp and goes up by %d%%", n, pct),
					Prompt:  "What does it cost now, in pence?",
					Answer:  Int(n + change),
				}
			}
			return Question{
				Context: fmt.Sprintf("A comic costs %dp and has %d%% off", n, pct),
				Prompt:  "What does it cost now, in pence?",
				Answer:  Int(n - change),
			}
		},
	},
	{
		Title:       "Finding the Whole",
		Hint:        "You are told the part. Work out one percent, then all hundred of them.",
		Questions:   8,
		MinAccuracy: 0.70,
		Gen: func(r *rand.Rand) Question {
			whole := (r.IntN(10) + 1) * 20
			pct := wholePercent(r)
			return Question{
				Context: fmt.Sprintf("%d%% of a number is %d", pct, pct*whole/100),
				Prompt:  "What is the number?",
				Answer:  Int(whole),
			}
		},
	},
	{
		Title:       "How Much Did It Change",
		Hint:        "The change goes over what it started at, not what it ended at.",
		Questions:   8,
		MinAccuracy: 0.65,
		Gen: func(r *rand.Rand) Question {
			from := (r.IntN(10) + 1) * 20
			pct := wholePercent(r)
			change := pct * from / 100
			if r.IntN(2) == 0 {
				return Question{
					Context: fmt.Sprintf("The price goes from %dp to %dp", from, from+change),
					Prompt:  "What is the increase, as a percent?",
					Answer:  Int(pct),
				}
			}
			return Question{
				Context: fmt.Sprintf("The price goes from %dp to %dp", from, from-change),
				Prompt:  "What is the decrease, as a percent?",
				Answer:  Int(pct),
			}
		},
	},
}
