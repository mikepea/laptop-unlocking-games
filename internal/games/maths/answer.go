package maths

import (
	"fmt"
	"strconv"
	"strings"
)

// Answer is what a question wants typed.
//
// It used to be a plain int, which was enough while the game was number facts
// and solving for x. The book it now follows spends three chapters on
// fractions, decimals and percents, and asking for those as whole numbers
// ("give the numerator of the answer in lowest terms") teaches the game rather
// than the maths. So an answer is an exact rational plus a Form saying how it
// must be written down.
//
// Everything is exact: there is no float anywhere in here, so 0.1 + 0.2 is
// 0.3 and a child is never told they were wrong by a rounding error.
type Answer struct {
	// num/den is the value, always in lowest terms with den > 0.
	num, den int
	form     Form
	// text is the answer as it should be written, for the correction screen.
	text string
}

// Form is how an answer has to be written down.
type Form int

const (
	// FormWhole is an integer. Anything of equal value is accepted, so a child
	// who types 4/1 is not punished for it.
	FormWhole Form = iota
	// FormFrac wants a fraction, in lowest terms. A decimal of the same value
	// is refused: the level asking for it is teaching the fraction. A whole
	// number is fine when the fraction reduces to one.
	FormFrac
	// FormDec wants a decimal. A fraction of the same value is refused, for
	// the same reason.
	FormDec
)

// Int is a whole-number answer.
func Int(n int) Answer {
	return Answer{num: n, den: 1, form: FormWhole, text: strconv.Itoa(n)}
}

// Fraction is an answer that must be typed as a fraction in lowest terms. It
// reduces what it is handed, so a generator can pass 18/24 and let this work
// out that the answer on screen is 3/4.
func Fraction(num, den int) Answer {
	if den == 0 {
		panic("maths: Fraction with a zero denominator")
	}
	if den < 0 {
		num, den = -num, -den
	}
	if g := gcd(num, den); g > 1 {
		num, den = num/g, den/g
	}
	a := Answer{num: num, den: den, form: FormFrac, text: fmt.Sprintf("%d/%d", num, den)}
	if den == 1 {
		a.text = strconv.Itoa(num)
	}
	return a
}

// Decimal is an answer that must be typed as a decimal, given as scaled over
// ten to the places: Decimal(375, 3) is 0.375 and Decimal(-25, 1) is -2.5.
// Taking it scaled rather than as a float keeps the value exact.
func Decimal(scaled, places int) Answer {
	if places < 0 {
		panic("maths: Decimal with negative places")
	}
	den := pow(10, places)
	a := Answer{num: scaled, den: den, form: FormDec, text: formatDecimal(scaled, places)}
	if g := gcd(a.num, a.den); g > 1 {
		a.num, a.den = a.num/g, a.den/g
	}
	return a
}

// formatDecimal writes scaled/10^places the way it would be written by hand,
// with trailing zeros trimmed: 0.30 is 0.3, and 4.00 is 4.
func formatDecimal(scaled, places int) string {
	sign := ""
	if scaled < 0 {
		sign, scaled = "-", -scaled
	}
	digits := strconv.Itoa(scaled)
	if len(digits) <= places {
		digits = strings.Repeat("0", places-len(digits)+1) + digits
	}
	whole, frac := digits[:len(digits)-places], digits[len(digits)-places:]
	frac = strings.TrimRight(frac, "0")
	if frac == "" {
		return sign + whole
	}
	return sign + whole + "." + frac
}

// String is the answer as it should be written, for the correction screen and
// the list of ones worth another look.
func (a Answer) String() string { return a.text }

// Whole reports the answer as an int, and whether it is one. Chapter tests use
// it to re-derive an answer they can compute directly.
func (a Answer) Whole() (int, bool) { return a.num, a.den == 1 }

// Value is the answer as an exact fraction in lowest terms, for tests that
// need to compare two answers without going through their text.
func (a Answer) Value() (num, den int) { return a.num, a.den }

// Form is how the answer has to be written.
func (a Answer) Form() Form { return a.form }

// Equal reports whether two answers are the same number written the same way.
func (a Answer) Equal(b Answer) bool {
	return a.num == b.num && a.den == b.den && a.form == b.form
}

// Matches grades what was typed. A blank or unparseable entry is simply wrong.
//
// Grading is by value, not by string, so 0.50 and 0.5 both pass and so does a
// fraction written the long way round -- except where the Form says the shape
// of the answer is the thing being taught, in which case the wrong shape is
// wrong. That is deliberate: "1/4 + 1/4" answered 0.5 has not practised
// anything the level was for, and the correction screen shows 1/2, which says
// so better than an error message would.
func (a Answer) Matches(input string) bool {
	num, den, form, ok := parseNumber(input)
	if !ok {
		return false
	}
	if num*a.den != a.num*den {
		return false
	}
	switch a.form {
	case FormFrac:
		// A fraction must be in lowest terms; 6/8 is not an answer to "give
		// this in simplest form". A whole number is fine when it is the value.
		return form != FormDec && (form != FormFrac || gcd(num, den) == 1)
	case FormDec:
		return form != FormFrac
	}
	return true
}

// parseNumber reads what a player typed as an exact fraction, reporting which
// of the three shapes it was written in: "7", "3/4", "0.25" and ".25" are all
// numbers, with an optional leading minus.
func parseNumber(s string) (num, den int, form Form, ok bool) {
	s = strings.TrimSpace(s)
	// Longer than anything the field will hold, and long enough that the
	// arithmetic below cannot overflow.
	if s == "" || len(s) > 12 {
		return 0, 0, 0, false
	}

	neg := false
	if rest, cut := strings.CutPrefix(s, "-"); cut {
		neg, s = true, rest
	}
	if !wellFormedNumber(s) {
		return 0, 0, 0, false
	}
	sign := 1
	if neg {
		sign = -1
	}

	if top, bottom, cut := strings.Cut(s, "/"); cut {
		n, err1 := strconv.Atoi(top)
		d, err2 := strconv.Atoi(bottom)
		if err1 != nil || err2 != nil || d == 0 {
			return 0, 0, 0, false
		}
		return sign * n, d, FormFrac, true
	}

	if whole, frac, cut := strings.Cut(s, "."); cut {
		// "3." is not a number anyone means to have typed, and neither is "."
		if frac == "" {
			return 0, 0, 0, false
		}
		if whole == "" {
			whole = "0"
		}
		w, err1 := strconv.Atoi(whole)
		f, err2 := strconv.Atoi(frac)
		if err1 != nil || err2 != nil {
			return 0, 0, 0, false
		}
		scale := pow(10, len(frac))
		return sign * (w*scale + f), scale, FormDec, true
	}

	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, 0, 0, false
	}
	return sign * n, 1, FormWhole, true
}

// wellFormedNumber reports whether s (already stripped of any leading minus)
// is digits with at most one point and at most one slash, and is not
// punctuation alone. It runs before the Atoi calls below so they cannot be
// handed a stray sign from the middle of the string.
func wellFormedNumber(s string) bool {
	digits, points, slashes := 0, 0, 0
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9':
			digits++
		case r == '.':
			points++
		case r == '/':
			slashes++
		default:
			return false
		}
	}
	return digits > 0 && points <= 1 && slashes <= 1
}
