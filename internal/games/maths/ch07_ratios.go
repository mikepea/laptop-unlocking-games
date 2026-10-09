package maths

import (
	"fmt"
	"math/rand/v2"
)

// Chapter 7 is ratios, conversions and rates, which the book treats as one
// idea seen three times: two quantities that keep the same relationship while
// both of them change.
//
// These are word problems, so the prompts are sentences. Money is in pence
// and lengths are metric, because the pound sign is not on the console font
// and a mixed-unit question is a vocabulary test wearing a maths costume.

// countable is the stuff the word problems are about. Plural only: "4 pens
// cost 60p" needs no article and no singular form.
var countable = []string{"pens", "apples", "stickers", "marbles", "cards", "buns"}

// conversion is one pair of metric units and the factor between them. Every
// factor is a multiple of ten, which is what makes a value with one decimal
// place convert to a whole number.
type conversion struct {
	big, small string
	factor     int
}

var conversions = []conversion{
	{"m", "cm", 100},
	{"cm", "mm", 10},
	{"km", "m", 1000},
	{"kg", "g", 1000},
	{"litres", "ml", 1000},
	{"hours", "minutes", 60},
	{"minutes", "seconds", 60},
}

var ratioLevels = []Level{
	{
		Title:       "Ratios",
		Hint:        "Add the parts of the ratio up. That many shares, then count out each one.",
		Questions:   8,
		MinAccuracy: 0.70,
		Gen: func(r *rand.Rand) Question {
			// Coprime parts, so the ratio on screen is already in its
			// simplest form and the number of shares is unambiguous.
			var p, q int
			for {
				p, q = r.IntN(6)+1, r.IntN(6)+1
				if p != q && gcd(p, q) == 1 {
					break
				}
			}
			k := r.IntN(9) + 2

			switch r.IntN(3) {
			case 0:
				which, share := "bigger", max(p, q)*k
				if r.IntN(2) == 0 {
					which, share = "smaller", min(p, q)*k
				}
				return Question{
					Context: fmt.Sprintf("%d shared in the ratio %d:%d", (p+q)*k, p, q),
					Prompt:  fmt.Sprintf("How big is the %s part?", which),
					Answer:  Int(share),
				}
			case 1:
				return Question{
					Context: fmt.Sprintf("%d : ? = %d : %d", p*k, p, q),
					Prompt:  "?", Answer: Int(q * k),
				}
			default:
				return Question{
					Prompt: fmt.Sprintf("%d cats for every %d dogs. With %d cats, how many dogs?", p, q, p*k),
					Answer: Int(q * k),
				}
			}
		},
	},
	{
		Title:       "Scaling Up",
		Hint:        "Work out what one of them costs or takes, then multiply.",
		Questions:   8,
		MinAccuracy: 0.70,
		Gen: func(r *rand.Rand) Question {
			each := r.IntN(18) + 3 // 3..20
			n1 := r.IntN(5) + 2
			// A different number from the one given, or there is nothing to
			// scale and the answer is already on screen.
			n2 := n1
			for n2 == n1 {
				n2 = r.IntN(8) + 2
			}
			if r.IntN(2) == 0 {
				thing := countable[r.IntN(len(countable))]
				return Question{
					Context: fmt.Sprintf("%d %s cost %dp", n1, thing, n1*each),
					Prompt:  fmt.Sprintf("What do %d cost, in pence?", n2),
					Answer:  Int(n2 * each),
				}
			}
			return Question{
				Context: fmt.Sprintf("A machine makes %d cards in %d minutes", n1*each, n1),
				Prompt:  fmt.Sprintf("How many in %d minutes?", n2),
				Answer:  Int(n2 * each),
			}
		},
	},
	{
		Title:       "Changing Units",
		Hint:        "Going to the smaller unit, multiply. Going to the bigger one, divide.",
		Questions:   8,
		MinAccuracy: 0.70,
		Gen: func(r *rand.Rand) Question {
			c := conversions[r.IntN(len(conversions))]
			value := randDec(r, 9, 1)
			if r.IntN(2) == 0 {
				return Question{
					Prompt: fmt.Sprintf("How many %s in %s %s?", c.small, value, c.big),
					Answer: Int(value.scaled * c.factor / 10),
				}
			}
			// Built from the answer, so the division always comes out.
			return Question{
				Prompt: fmt.Sprintf("How many %s is %d %s?", c.big, value.scaled*c.factor/10, c.small),
				Answer: Decimal(value.scaled, value.places),
			}
		},
	},
	{
		Title:       "Rates",
		Hint:        "How much for one? Divide the total by how many there were.",
		Questions:   8,
		MinAccuracy: 0.70,
		Gen: func(r *rand.Rand) Question {
			each := r.IntN(20) + 2
			n := r.IntN(9) + 2
			switch r.IntN(3) {
			case 0:
				return Question{
					Context: fmt.Sprintf("%d km in %d hours", each*n, n),
					Prompt:  "How many km in one hour?",
					Answer:  Int(each),
				}
			case 1:
				thing := countable[r.IntN(len(countable))]
				return Question{
					Context: fmt.Sprintf("%d %s cost %dp", n, thing, each*n),
					Prompt:  "How much is one, in pence?",
					Answer:  Int(each),
				}
			default:
				return Question{
					Context: fmt.Sprintf("%d pages read in %d days", each*n, n),
					Prompt:  "How many pages a day?",
					Answer:  Int(each),
				}
			}
		},
	},
	{
		Title:       "Speed, Distance and Time",
		Hint:        "Speed is distance over time. Rearrange it for whichever one is missing.",
		Questions:   8,
		MinAccuracy: 0.70,
		GenSteps:    speedSteps,
	},
}

// speedSteps takes one journey and asks for each of the three quantities in
// turn, so the same relationship gets used forwards and both ways back.
//
//	A train goes 180 km in 3 hours.   speed -> 60, then a distance, then a time
func speedSteps(r *rand.Rand) []Question {
	speed := (r.IntN(9) + 4) * 10 // 40..120 km/h, a plausible train
	hours := r.IntN(4) + 2        // 2..5
	laterHours := r.IntN(6) + 2
	farther := (r.IntN(6) + 2) * speed

	journey := fmt.Sprintf("A train goes %d km in %d hours", speed*hours, hours)
	return []Question{
		{Context: journey, Prompt: "How fast is it going, in km per hour?", Answer: Int(speed)},
		{
			Context: fmt.Sprintf("%s, so %d km per hour", journey, speed),
			Prompt:  fmt.Sprintf("How far would it get in %d hours?", laterHours),
			Answer:  Int(speed * laterHours),
		},
		{
			Context: fmt.Sprintf("%s, so %d km per hour", journey, speed),
			Prompt:  fmt.Sprintf("How many hours to go %d km?", farther),
			Answer:  Int(farther / speed),
		},
	}
}
