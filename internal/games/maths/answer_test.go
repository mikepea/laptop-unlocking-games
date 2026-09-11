package maths

import "testing"

// The answer type is the one piece of this package every level depends on. If
// it accepts something it should not, a child is told a wrong answer is right;
// if it refuses something it should accept, they are told a right one is
// wrong, which is worse.

func TestAnswerAcceptsTheValueHoweverItIsWritten(t *testing.T) {
	tests := []struct {
		name   string
		answer Answer
		yes    []string
		no     []string
	}{
		{
			name:   "a whole number",
			answer: Int(12),
			yes:    []string{"12", " 12 ", "24/2"},
			no:     []string{"", "13", "1.2", "banana", "-12", "12x"},
		},
		{
			name:   "a negative whole number",
			answer: Int(-7),
			yes:    []string{"-7"},
			no:     []string{"7", "-", "--7"},
		},
		{
			name:   "a fraction wants lowest terms",
			answer: Fraction(3, 4),
			yes:    []string{"3/4"},
			no:     []string{"6/8", "0.75", "4/3", "3"},
		},
		{
			name:   "a fraction that came out whole",
			answer: Fraction(4, 2),
			yes:    []string{"2", "2/1"},
			no:     []string{"4/2", "2.0"},
		},
		{
			name:   "a decimal refuses the fraction of the same value",
			answer: Decimal(25, 2),
			yes:    []string{"0.25", ".25", "0.250", "0.2500"},
			no:     []string{"1/4", "25", "0.2", "."},
		},
		{
			name:   "a decimal that came out whole",
			answer: Decimal(400, 2),
			yes:    []string{"4", "4.0", "4.00"},
			no:     []string{"400", "0.04"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for _, in := range tc.yes {
				if !tc.answer.Matches(in) {
					t.Errorf("%s refused %q, which is the right answer", tc.answer, in)
				}
			}
			for _, in := range tc.no {
				if tc.answer.Matches(in) {
					t.Errorf("%s accepted %q, which is not", tc.answer, in)
				}
			}
		})
	}
}

func TestAnswerWritesItselfTheWayItWantsTyping(t *testing.T) {
	tests := []struct {
		answer Answer
		want   string
	}{
		{Int(0), "0"},
		{Int(-7), "-7"},
		{Fraction(18, 24), "3/4"},  // reduced on the way in
		{Fraction(-2, -4), "1/2"},  // and the signs tidied up
		{Fraction(5, -10), "-1/2"}, // a negative denominator moves to the top
		{Fraction(6, 3), "2"},      // a fraction that is really a whole number
		{Decimal(375, 3), "0.375"},
		{Decimal(30, 2), "0.3"}, // trailing noughts are not written by hand
		{Decimal(1250, 2), "12.5"},
		{Decimal(0, 2), "0"},
		{Decimal(-25, 1), "-2.5"},
		{Decimal(7, 0), "7"},
	}
	for _, tc := range tests {
		if got := tc.answer.String(); got != tc.want {
			t.Errorf("answer written as %q, want %q", got, tc.want)
		}
		// Whatever it writes, it has to take back.
		if !tc.answer.Matches(tc.answer.String()) {
			t.Errorf("answer %q is not accepted by its own marker", tc.answer)
		}
	}
}

func TestParseNumberTellsTheThreeShapesApart(t *testing.T) {
	tests := []struct {
		in       string
		num, den int
		form     Form
		ok       bool
	}{
		{in: "7", num: 7, den: 1, form: FormWhole, ok: true},
		{in: "-7", num: -7, den: 1, form: FormWhole, ok: true},
		{in: "3/4", num: 3, den: 4, form: FormFrac, ok: true},
		{in: "-3/4", num: -3, den: 4, form: FormFrac, ok: true},
		{in: "0.25", num: 25, den: 100, form: FormDec, ok: true},
		{in: ".25", num: 25, den: 100, form: FormDec, ok: true},
		{in: "-0.25", num: -25, den: 100, form: FormDec, ok: true},
		// Everything a keyboard can produce that is not a number.
		{in: ""}, {in: "   "}, {in: "-"}, {in: "."}, {in: "/"},
		{in: "3."}, {in: "1/0"}, {in: "1.2.3"}, {in: "1/2/3"},
		{in: "3-4"}, {in: "3/-4"}, {in: "12345678901234"},
	}
	for _, tc := range tests {
		num, den, form, ok := parseNumber(tc.in)
		if ok != tc.ok {
			t.Errorf("parseNumber(%q) ok = %v, want %v", tc.in, ok, tc.ok)
			continue
		}
		if !ok {
			continue
		}
		if num != tc.num || den != tc.den || form != tc.form {
			t.Errorf("parseNumber(%q) = %d/%d %v, want %d/%d %v",
				tc.in, num, den, form, tc.num, tc.den, tc.form)
		}
	}
}

// Every character the field lets through has to reach the marker without
// crashing it. The field is deliberately permissive so a typo can be
// corrected, which means the marker sees a lot of nonsense.
func TestMarkerSurvivesAnythingTheFieldAllows(t *testing.T) {
	const allowed = "0123456789-/."
	answers := []Answer{Int(3), Fraction(3, 4), Decimal(25, 2)}
	for _, a := range answers {
		for _, x := range allowed {
			for _, y := range allowed {
				for _, z := range allowed {
					a.Matches(string([]rune{x, y, z}))
				}
			}
		}
	}
}
