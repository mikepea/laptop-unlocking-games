package maths

import (
	"fmt"
	"math/rand/v2"
)

// Chapter 6 is decimals, which the book introduces as fractions whose
// denominator is a power of ten and never lets you forget it. So this chapter
// runs both ways: arithmetic in decimals, and turning each form into the
// other.
//
// Nothing here is a float. A decimal is carried as a whole number of
// thousandths or hundredths and only written with a point at the last moment,
// so 0.1 + 0.2 comes out 0.3 and a child is never marked wrong by a rounding
// error they cannot see.

// dec is an exact decimal: the value is scaled over ten to the places.
type dec struct {
	scaled int
	places int
}

func (d dec) String() string { return formatDecimal(d.scaled, d.places) }

// at rewrites the value with more decimal places, for lining two of them up.
func (d dec) at(places int) int { return d.scaled * pow(10, places-d.places) }

// randDec draws a decimal below maxWhole with the given number of places,
// never zero.
func randDec(r *rand.Rand, maxWhole, places int) dec {
	return dec{scaled: r.IntN(maxWhole*pow(10, places)-1) + 1, places: places}
}

// randDecNotWhole is randDec with something after the point. A decimals level
// that asks "2 x 4" has wandered back into the warm-up.
func randDecNotWhole(r *rand.Rand, maxWhole, places int) dec {
	for {
		d := randDec(r, maxWhole, places)
		if d.scaled%pow(10, places) != 0 {
			return d
		}
	}
}

// placeNames are the columns after the point, in order.
var placeNames = []string{"tenths", "hundredths", "thousandths"}

// terminatingDenominators are the ones that divide a power of ten, so the
// decimal stops rather than recurring. Recurring decimals are in the book but
// cannot be typed into a single field, so they are left out.
var terminatingDenominators = []int{2, 4, 5, 8, 10, 16, 20, 25, 40, 50}

// placesFor is how many decimal places a fraction over den needs. It only
// answers for the denominators above, all of which divide 10000.
func placesFor(den int) int {
	for p := 1; p <= 4; p++ {
		if pow(10, p)%den == 0 {
			return p
		}
	}
	panic("maths: denominator does not terminate")
}

var decimalLevels = []Level{
	{
		Title:       "Place Value",
		Hint:        "After the point come tenths, then hundredths, then thousandths.",
		Questions:   8,
		MinAccuracy: 0.75,
		Gen: func(r *rand.Rand) Question {
			d := randDec(r, 90, 3)
			if r.IntN(2) == 0 {
				// The last digit has to be there to be asked about: a value
				// ending in nought prints without it, and "which digit is in
				// the thousandths of 47.38" is a trick question.
				for d.scaled%10 == 0 {
					d = randDec(r, 90, 3)
				}
				place := r.IntN(len(placeNames))
				digit := d.scaled / pow(10, 2-place) % 10
				return Question{
					Context: d.String(),
					Prompt:  fmt.Sprintf("Which digit is in the %s?", placeNames[place]),
					Answer:  Int(digit),
				}
			}
			// Times and divide by ten: the point stays still and the digits
			// slide, which is the whole idea of place value.
			if r.IntN(2) == 0 {
				return Question{
					Prompt: fmt.Sprintf("%s x 10", d),
					Answer: Decimal(d.scaled, d.places-1),
				}
			}
			return Question{
				Prompt: fmt.Sprintf("%s / 10", d),
				Answer: Decimal(d.scaled, d.places+1),
			}
		},
	},
	{
		Title:       "Adding Decimals",
		Hint:        "Line the points up under each other. Pad the short one with noughts.",
		Questions:   8,
		MinAccuracy: 0.75,
		Gen: func(r *rand.Rand) Question {
			p1, p2 := r.IntN(2)+1, r.IntN(2)+1
			a, b := randDecNotWhole(r, 40, p1), randDecNotWhole(r, 40, p2)
			places := max(p1, p2)
			if r.IntN(2) == 0 {
				return Question{
					Prompt: fmt.Sprintf("%s + %s", a, b),
					Answer: Decimal(a.at(places)+b.at(places), places),
				}
			}
			// Bigger one first: negative decimals belong with negative
			// numbers, not here.
			if a.at(places) < b.at(places) {
				a, b = b, a
			}
			return Question{
				Prompt: fmt.Sprintf("%s - %s", a, b),
				Answer: Decimal(a.at(places)-b.at(places), places),
			}
		},
	},
	{
		Title:       "Multiplying Decimals",
		Hint:        "Ignore the points, multiply, then put back as many places as you took out.",
		Questions:   8,
		MinAccuracy: 0.70,
		Gen: func(r *rand.Rand) Question {
			a := randDecNotWhole(r, 6, r.IntN(2)+1)
			if r.IntN(2) == 0 {
				whole := r.IntN(11) + 2
				return Question{
					Prompt: fmt.Sprintf("%s x %d", a, whole),
					Answer: Decimal(a.scaled*whole, a.places),
				}
			}
			// One place on the second one, so the answer never runs past
			// three: 3.67 x 2.04 is a written-method exercise, not a decimals
			// question.
			b := randDecNotWhole(r, 4, 1)
			return Question{
				Prompt: fmt.Sprintf("%s x %s", a, b),
				Answer: Decimal(a.scaled*b.scaled, a.places+b.places),
			}
		},
	},
	{
		Title:       "Dividing Decimals",
		Hint:        "Move both points the same way until you are dividing by a whole number.",
		Questions:   8,
		MinAccuracy: 0.70,
		Gen: func(r *rand.Rand) Question {
			// Built from the answer outwards, so it always comes out exactly
			// and never runs off the end of the field.
			if r.IntN(2) == 0 {
				quotient := randDecNotWhole(r, 6, r.IntN(2)+1)
				divisor := r.IntN(8) + 2
				return Question{
					Prompt: fmt.Sprintf("%s / %d", dec{quotient.scaled * divisor, quotient.places}, divisor),
					Answer: Decimal(quotient.scaled, quotient.places),
				}
			}
			// Dividing by a decimal, so the quotient keeps to one place and
			// the dividend stays readable.
			quotient := randDecNotWhole(r, 6, 1)
			divisor := randDecNotWhole(r, 4, 1)
			dividend := dec{quotient.scaled * divisor.scaled, quotient.places + divisor.places}
			return Question{
				Prompt: fmt.Sprintf("%s / %s", dividend, divisor),
				Answer: Decimal(quotient.scaled, quotient.places),
			}
		},
	},
	{
		Title:       "Swapping Forms",
		Hint:        "A decimal is a fraction over ten, a hundred or a thousand. Cancel it down.",
		Questions:   8,
		MinAccuracy: 0.70,
		Gen: func(r *rand.Rand) Question {
			den := terminatingDenominators[r.IntN(len(terminatingDenominators))]
			num := r.IntN(den-1) + 1
			if g := gcd(num, den); g > 1 {
				num, den = num/g, den/g
			}
			places := placesFor(den)
			scaled := num * pow(10, places) / den
			if r.IntN(2) == 0 {
				return Question{
					Prompt: fmt.Sprintf("%s as a decimal", frac(num, den)),
					Answer: Decimal(scaled, places),
				}
			}
			return Question{
				Prompt: fmt.Sprintf("%s as a fraction in simplest form", formatDecimal(scaled, places)),
				Answer: Fraction(num, den),
			}
		},
	},
	{
		Title:       "Rounding",
		Hint:        "Look at the next digit along. Five or more rounds up.",
		Questions:   8,
		MinAccuracy: 0.75,
		Gen: func(r *rand.Rand) Question {
			to := r.IntN(2) + 1 // round to one or two places
			var d dec
			for {
				d = randDec(r, 50, to+2)
				// The two digits being dropped have to be worth dropping, and
				// they must not be an exact half: that is a convention
				// argument, not a question, so skip it rather than mark a
				// reasonable answer wrong.
				if dropped := d.scaled % 100; dropped != 0 && dropped != 50 {
					break
				}
			}
			cut := pow(10, d.places-to)
			rounded := (d.scaled + cut/2) / cut
			return Question{
				Prompt: fmt.Sprintf("%s to %s", d, decimalPlaces(to)),
				Answer: Decimal(rounded, to),
			}
		},
	},
}

// decimalPlaces writes "1 decimal place" and "2 decimal places".
func decimalPlaces(n int) string {
	if n == 1 {
		return "1 decimal place"
	}
	return fmt.Sprintf("%d decimal places", n)
}
