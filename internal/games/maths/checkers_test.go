package maths

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// answerCheckers maps every level to the code that re-derives its answers.
// TestEveryLevelAgreesWithWhatItShows fails on a level that is missing here,
// so this map is the list of levels as much as Chapters is.
var answerCheckers = map[string]checker{
	// Warm-up: plain sums, so the prompt is the whole question.
	"Times Tables":       checkExpression,
	"Sharing Out":        checkExpression,
	"Everything At Once": checkExpression,

	// Chapter 1: still expressions, now with brackets and minus signs.
	"Clever Adding":       checkExpression,
	"Clever Multiplying":  checkExpression,
	"Breaking Numbers Up": checkExpression,
	"Below Zero":          checkExpression,
	"Signs That Multiply": checkExpression,
	"Order of Operations": checkExpression,

	// Chapter 2: the "?" questions put the answer back into the rule on
	// screen and check the two sides still agree.
	"Powers":             checkExpression,
	"Multiplying Powers": checkExpression,
	"Dividing Powers":    checkExpression,
	"Powers of Powers":   checkExpression,
	"Zero and Below":     checkExpression,

	// Chapter 3: sentences, so each one is read apart and worked out the
	// slow, obvious way.
	"Primes":                   checkPrimes,
	"Multiples":                checkMultiples,
	"Divisibility Tests":       checkDivisibility,
	"Prime Factorisation":      checkFactorisation,
	"Counting Divisors":        checkCountingDivisors,
	"Biggest Common Factor":    checkCommonFactor,
	"Smallest Common Multiple": checkCommonMultiple,

	// Chapter 4: the answer has to be a fraction in lowest terms, and the
	// checker says so as well as checking the value.
	"Equivalent Fractions":  checkExpression,
	"Simplest Form":         checkFractionForm,
	"Adding Fractions":      checkFractionForm,
	"Taking Fractions Away": checkFractionForm,
	"Multiplying Fractions": checkFractionForm,
	"Dividing Fractions":    checkDividingFractions,
	"Mixed Numbers":         checkFractionForm,
	"Which Is Bigger":       checkWhichIsBigger,

	// Chapter 5: equations, checked by putting the answer back in.
	"Missing Number":         checkSolvedEquation,
	"One Step":               checkSolvedEquation,
	"Two Steps":              checkSolvedEquation,
	"Letters on Both Sides":  checkSolvedEquation,
	"Putting Numbers In":     checkSubstitution,
	"Collecting Up":          checkCollecting,
	"Inequalities":           checkInequality,
	"Simplify and Factorise": checkSimplifyStep,

	// Chapter 6.
	"Place Value":          checkPlaceValue,
	"Adding Decimals":      checkDecimalForm,
	"Multiplying Decimals": checkDecimalForm,
	"Dividing Decimals":    checkDecimalForm,
	"Swapping Forms":       checkSwappingForms,
	"Rounding":             checkRounding,

	// Chapter 7: word problems, read apart into their numbers.
	"Ratios":                   checkRatios,
	"Scaling Up":               checkScalingUp,
	"Changing Units":           checkChangingUnits,
	"Rates":                    checkRates,
	"Speed, Distance and Time": checkSpeed,

	// Chapter 8.
	"Percent of a Number":              checkPercentOf,
	"Percents, Decimals and Fractions": checkPercentForms,
	"Up and Down by a Percent":         checkPercentChange,
	"Finding the Whole":                checkFindingWhole,
	"How Much Did It Change":           checkPercentDifference,

	// Chapter 9.
	"Perfect Squares":             checkWordyExpression,
	"Roots Between Whole Numbers": checkRootBounds,
	"Multiplying Roots":           checkMultiplyingRoots,
	"Tidying Roots Up":            checkTidyingRoots,
}

// ---------------------------------------------------------------------------
// Chapter 3. Nothing in here uses the helpers the generators use.

// primeByTrialDivision is a second opinion on isPrime, written the other way
// round: a number is prime if nothing below it divides it.
func primeByTrialDivision(n int) bool {
	if n < 2 {
		return false
	}
	for d := 2; d < n; d++ {
		if n%d == 0 {
			return false
		}
	}
	return true
}

func checkMultiples(t *testing.T, level string, q Question) {
	t.Helper()
	n := intsIn(t, q.Prompt)
	switch {
	case strings.HasPrefix(q.Prompt, "Smallest multiple of"):
		step, above := n[0], n[1]
		k := above + 1
		for k%step != 0 {
			k++
		}
		wantInt(t, level, q, k)
	case strings.HasPrefix(q.Prompt, "How many multiples of"):
		step, lo, hi := n[0], n[1], n[2]
		count := 0
		for v := lo; v <= hi; v++ {
			if v%step == 0 {
				count++
			}
		}
		wantInt(t, level, q, count)
	default: // "What is the 4th multiple of 7?"
		which, step := n[0], n[1]
		total := 0
		for i := 0; i < which; i++ {
			total += step
		}
		wantInt(t, level, q, total)
	}
}

func checkDivisibility(t *testing.T, level string, q Question) {
	t.Helper()
	if strings.HasPrefix(q.Prompt, "What do the digits") {
		n := intsIn(t, q.Context)[0]
		sum := 0
		for _, c := range strconv.Itoa(n) {
			sum += int(c - '0')
		}
		wantInt(t, level, q, sum)
		return
	}
	// "What is the remainder when 3746 is divided by 9?"
	n := intsIn(t, q.Prompt)
	wantInt(t, level, q, n[0]%n[1])
}

func checkPrimes(t *testing.T, level string, q Question) {
	t.Helper()
	n := intsIn(t, q.Prompt)
	switch {
	case strings.HasPrefix(q.Prompt, "What is the next prime"):
		k := n[0] + 1
		for !primeByTrialDivision(k) {
			k++
		}
		wantInt(t, level, q, k)
	case strings.HasPrefix(q.Prompt, "How many primes"):
		count := 0
		for v := n[0] + 1; v < n[1]; v++ {
			if primeByTrialDivision(v) {
				count++
			}
		}
		wantInt(t, level, q, count)
	default: // "What is the smallest prime that divides 91?"
		d := 2
		for n[0]%d != 0 {
			d++
		}
		if !primeByTrialDivision(d) {
			t.Fatalf("%s: %q, but %d is not prime", level, q.Prompt, d)
		}
		wantInt(t, level, q, d)
	}
}

func checkFactorisation(t *testing.T, level string, q Question) {
	t.Helper()
	n := intsIn(t, q.Context)[0]
	p := intsIn(t, q.Prompt)[0]
	count := 0
	for n%p == 0 {
		n, count = n/p, count+1
	}
	wantInt(t, level, q, count)
}

func checkCountingDivisors(t *testing.T, level string, q Question) {
	t.Helper()
	n := intsIn(t, q.Prompt)[0]
	count := 0
	for d := 1; d <= n; d++ {
		if n%d == 0 {
			count++
		}
	}
	wantInt(t, level, q, count)

	// The factorisation shown above the question has to be that number's.
	lhs, rhs, ok := strings.Cut(q.Context, " = ")
	if !ok {
		t.Fatalf("%s: context %q is not an equation", level, q.Context)
	}
	left, err := evalExpr(lhs)
	if err != nil {
		t.Fatalf("%s: %v", level, err)
	}
	right, err := evalExpr(rhs)
	if err != nil {
		t.Fatalf("%s: %v", level, err)
	}
	if !left.eq(right) {
		t.Errorf("%s: context claims %q, but that is %s and %s", level, q.Context, left, right)
	}
}

func checkCommonFactor(t *testing.T, level string, q Question) {
	t.Helper()
	n := intsIn(t, q.Prompt)
	best := 1
	for d := 1; d <= n[0] && d <= n[1]; d++ {
		if n[0]%d == 0 && n[1]%d == 0 {
			best = d
		}
	}
	wantInt(t, level, q, best)
}

func checkCommonMultiple(t *testing.T, level string, q Question) {
	t.Helper()
	n := intsIn(t, q.Prompt)
	m := 1
	for m%n[0] != 0 || m%n[1] != 0 {
		m++
	}
	wantInt(t, level, q, m)
}

// ---------------------------------------------------------------------------
// Chapter 4.

var fractionPattern = regexp.MustCompile(`\d+/\d+`)

// checkDividingFractions cannot go through the expression evaluator: on
// screen "3/4 divided by 2/5" has three slashes meaning two different things,
// and reading them left to right gives the wrong answer. So the two sides are
// split apart first.
func checkDividingFractions(t *testing.T, level string, q Question) {
	t.Helper()
	lhs, rhs, ok := strings.Cut(q.Prompt, " divided by ")
	if !ok {
		t.Fatalf("%s: %q does not say what is divided by what", level, q.Prompt)
	}
	a, err := evalExpr(lhs)
	if err != nil {
		t.Fatalf("%s: %v", level, err)
	}
	b, err := evalExpr(rhs)
	if err != nil {
		t.Fatalf("%s: %v", level, err)
	}
	v, err := a.div(b)
	if err != nil {
		t.Fatalf("%s: %q divides by zero", level, q.Prompt)
	}
	want(t, level, q, v)
	wantForm(t, level, q, FormFrac)
}

func checkWhichIsBigger(t *testing.T, level string, q Question) {
	t.Helper()
	shown := fractionPattern.FindAllString(q.Prompt, -1)
	if len(shown) != 2 {
		t.Fatalf("%s: %q does not offer two fractions", level, q.Prompt)
	}
	a, err := evalExpr(shown[0])
	if err != nil {
		t.Fatalf("%s: %v", level, err)
	}
	b, err := evalExpr(shown[1])
	if err != nil {
		t.Fatalf("%s: %v", level, err)
	}
	if a.eq(b) {
		t.Fatalf("%s: %q, but they are the same size", level, q.Prompt)
	}
	bigger := a
	if b.n*a.d > a.n*b.d {
		bigger = b
	}
	want(t, level, q, bigger)
	// The answer has to be one of the two on screen, written as it is
	// written there: being told "5/8" when you typed what you were shown
	// would be indistinguishable from a bug.
	if q.Answer.String() != shown[0] && q.Answer.String() != shown[1] {
		t.Errorf("%s: %q answers %s, which is neither of the fractions shown",
			level, q.Prompt, q.Answer)
	}
}

// From here on: the chapters whose questions are equations with letters in,
// or sentences. Same promise as above -- nothing in here shares code with the
// generator it is checking.

// ---------------------------------------------------------------------------
// Chapter 5.

// answerInt is the answer as a whole number, failing the test if it is not
// one. Every answer in chapter 5 is.
func answerInt(t *testing.T, level string, q Question) int {
	t.Helper()
	n, ok := q.Answer.Whole()
	if !ok {
		t.Fatalf("%s: %q answers %s, which is not a whole number", level, q.Prompt, q.Answer)
	}
	return n
}

func checkSolvedEquation(t *testing.T, level string, q Question) {
	t.Helper()
	checkEquation(t, level, q.Context, answerInt(t, level, q))
	if n := answerInt(t, level, q); n < 0 {
		t.Errorf("%s: negative answer in %q", level, q.Context)
	}
}

func checkSubstitution(t *testing.T, level string, q Question) {
	t.Helper()
	vars := map[string]int{}
	for _, assign := range strings.Split(q.Context, ", ") {
		name, value, ok := strings.Cut(assign, " = ")
		if !ok {
			t.Fatalf("%s: bad context %q", level, q.Context)
		}
		n, err := strconv.Atoi(value)
		if err != nil {
			t.Fatalf("%s: bad value in %q: %v", level, q.Context, err)
		}
		vars[name] = n
	}
	got, err := evalTerms(strings.Fields(q.Prompt), vars)
	if err != nil {
		t.Fatalf("%s: prompt %q with %q: %v", level, q.Prompt, q.Context, err)
	}
	wantInt(t, level, q, got)
}

// sumCoefficients adds up the coefficients of one letter in an expression
// like "3x + 5y + 2x", ignoring the terms in other letters.
func sumCoefficients(expr, v string) int {
	total := 0
	for _, tok := range strings.Fields(expr) {
		m := coefficient.FindStringSubmatch(tok)
		if m == nil || m[2] != v {
			continue
		}
		n := 1
		if m[1] != "" {
			n, _ = strconv.Atoi(m[1])
		}
		total += n
	}
	return total
}

func checkCollecting(t *testing.T, level string, q Question) {
	t.Helper()
	lhs, _, ok := strings.Cut(q.Context, " = ")
	if !ok {
		t.Fatalf("%s: bad context %q", level, q.Context)
	}
	wantInt(t, level, q, sumCoefficients(lhs, "x"))
}

// checkInequality tries the answer and the number next to it. An inequality
// question is only right if the answer fits and the one past it does not --
// checking just the answer would pass a level that was always one out.
func checkInequality(t *testing.T, level string, q Question) {
	t.Helper()
	lhs, rhs, sign := q.Context, "", ""
	for _, s := range []string{" < ", " > "} {
		if l, r, ok := strings.Cut(q.Context, s); ok {
			lhs, rhs, sign = l, r, strings.TrimSpace(s)
		}
	}
	if sign == "" {
		t.Fatalf("%s: context %q is not an inequality", level, q.Context)
	}
	limit, err := strconv.Atoi(rhs)
	if err != nil {
		t.Fatalf("%s: bad limit in %q: %v", level, q.Context, err)
	}

	holds := func(x int) bool {
		v, err := evalTerms(strings.Fields(lhs), map[string]int{"x": x})
		if err != nil {
			t.Fatalf("%s: left of %q: %v", level, q.Context, err)
		}
		if sign == "<" {
			return v < limit
		}
		return v > limit
	}

	answer := answerInt(t, level, q)
	if !holds(answer) {
		t.Errorf("%s: %q answers %d, which does not satisfy it", level, q.Context, answer)
	}
	// One step further in the direction the question asked must fail, or the
	// answer was not the biggest or smallest after all.
	beyond := answer + 1
	if strings.Contains(q.Prompt, "smallest") {
		beyond = answer - 1
	}
	if holds(beyond) {
		t.Errorf("%s: %q answers %d, but %d satisfies it too", level, q.Context, answer, beyond)
	}
}

// checkSimplifyStep checks one step of a simplify-and-factorise problem. The
// four steps ask four different things, told apart by their prompts.
func checkSimplifyStep(t *testing.T, level string, q Question) {
	t.Helper()
	switch {
	case strings.HasPrefix(q.Prompt, "How many x"):
		wantInt(t, level, q, sumCoefficients(q.Context, "x"))
	case strings.HasPrefix(q.Prompt, "How many y"):
		wantInt(t, level, q, sumCoefficients(q.Context, "y"))
	case strings.HasPrefix(q.Prompt, "What is the biggest number"):
		x, y := sumCoefficients(q.Context, "x"), sumCoefficients(q.Context, "y")
		best := 1
		for d := 1; d <= x && d <= y; d++ {
			if x%d == 0 && y%d == 0 {
				best = d
			}
		}
		wantInt(t, level, q, best)
	default:
		// "6x + 18y = 6(x + ?y)": put the answer in and multiply the bracket
		// back out.
		lhs, rhs, ok := strings.Cut(q.Context, " = ")
		if !ok {
			t.Fatalf("%s: bad context %q", level, q.Context)
		}
		filled := strings.Replace(rhs, "?", q.Answer.String(), 1)
		open := strings.Index(filled, "(")
		if open < 0 || !strings.HasSuffix(filled, ")") {
			t.Fatalf("%s: %q is not a factorised form", level, rhs)
		}
		factor, err := strconv.Atoi(filled[:open])
		if err != nil {
			t.Fatalf("%s: bad factor in %q: %v", level, filled, err)
		}
		inside := filled[open+1 : len(filled)-1]
		for _, v := range []string{"x", "y"} {
			if got, out := sumCoefficients(lhs, v), factor*sumCoefficients(inside, v); got != out {
				t.Errorf("%s: %q with ? = %s gives %d%s on the left and %d%s on the right",
					level, q.Context, q.Answer, got, v, out, v)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Chapter 6.

func checkPlaceValue(t *testing.T, level string, q Question) {
	t.Helper()
	if !strings.HasPrefix(q.Prompt, "Which digit") {
		checkDecimalForm(t, level, q)
		return
	}
	// Read the digit straight out of the text on screen, counting along from
	// the point. That catches a value printed without the column it is being
	// asked about as well as a wrong answer.
	_, after, ok := strings.Cut(q.Context, ".")
	if !ok {
		t.Fatalf("%s: %q has no decimal point", level, q.Context)
	}
	column := 0
	for i, name := range placeNames {
		if strings.Contains(q.Prompt, name) {
			column = i
		}
	}
	if column >= len(after) {
		t.Fatalf("%s: asks for the %s of %q, which does not print that far",
			level, placeNames[column], q.Context)
	}
	wantInt(t, level, q, int(after[column]-'0'))
}

func checkSwappingForms(t *testing.T, level string, q Question) {
	t.Helper()
	checkWordyExpression(t, level, q)
	if strings.HasSuffix(q.Prompt, "as a decimal") {
		wantForm(t, level, q, FormDec)
		return
	}
	wantForm(t, level, q, FormFrac)
	if num, den := q.Answer.Value(); den > 1 && gcd(num, den) != 1 {
		t.Errorf("%s: %q answers %s, which is not in lowest terms", level, q.Prompt, q.Answer)
	}
}

func checkRounding(t *testing.T, level string, q Question) {
	t.Helper()
	// "12.3456 to 2 decimal places"
	value, tail, ok := strings.Cut(q.Prompt, " to ")
	if !ok {
		t.Fatalf("%s: %q does not say what to round to", level, q.Prompt)
	}
	v, err := evalExpr(value)
	if err != nil {
		t.Fatalf("%s: %v", level, err)
	}
	places := intsIn(t, tail)[0]

	// Scale up, add a half, take the whole part, scale back: rounding from
	// first principles in exact arithmetic.
	scale := whole(pow(10, places))
	shifted := v.mul(scale).add(newRat(1, 2))
	floored := shifted.n / shifted.d
	rounded, err := whole(floored).div(scale)
	if err != nil {
		t.Fatalf("%s: %v", level, err)
	}
	want(t, level, q, rounded)
	wantForm(t, level, q, FormDec)
}

// ---------------------------------------------------------------------------
// Chapter 7.

func checkRatios(t *testing.T, level string, q Question) {
	t.Helper()
	switch {
	case strings.Contains(q.Context, "shared in the ratio"):
		n := intsIn(t, q.Context) // total, then the two parts
		total, p, q2 := n[0], n[1], n[2]
		part := max(p, q2)
		if strings.Contains(q.Prompt, "smaller") {
			part = min(p, q2)
		}
		wantInt(t, level, q, total/(p+q2)*part)
	case q.Prompt == "?":
		// "12 : ? = 3 : 5"
		n := intsIn(t, q.Context)
		wantInt(t, level, q, n[0]*n[2]/n[1])
	default:
		// "3 cats for every 5 dogs. With 12 cats, how many dogs?"
		n := intsIn(t, q.Prompt)
		cats, dogs, have := n[0], n[1], n[2]
		wantInt(t, level, q, have/cats*dogs)
	}
}

func checkScalingUp(t *testing.T, level string, q Question) {
	t.Helper()
	var count, total int
	if strings.Contains(q.Context, "cost") {
		n := intsIn(t, q.Context) // "4 pens cost 60p"
		count, total = n[0], n[1]
	} else {
		n := intsIn(t, q.Context) // "A machine makes 60 cards in 4 minutes"
		total, count = n[0], n[1]
	}
	asked := intsIn(t, q.Prompt)[0]
	if total%count != 0 {
		t.Fatalf("%s: %q does not divide into a whole rate", level, q.Context)
	}
	wantInt(t, level, q, total/count*asked)
}

// unitPairs is the conversion table written out again on purpose. Checking
// the game's factors against the game's own table would check nothing.
var unitPairs = []struct {
	big, small string
	factor     int
}{
	{"m", "cm", 100},
	{"cm", "mm", 10},
	{"km", "m", 1000},
	{"kg", "g", 1000},
	{"litres", "ml", 1000},
	{"hours", "minutes", 60},
	{"minutes", "seconds", 60},
}

func checkChangingUnits(t *testing.T, level string, q Question) {
	t.Helper()
	// "How many cm in 3.5 m?" going down, "How many m is 350 cm?" going up.
	// Either way the unit asked for comes third and the unit given comes
	// last, so the pair says which direction this is.
	fields := strings.Fields(q.Prompt)
	if len(fields) != 6 {
		t.Fatalf("%s: %q is not a conversion question", level, q.Prompt)
	}
	asked := fields[2]
	given := strings.TrimSuffix(fields[len(fields)-1], "?")

	values := valuesIn(q.Prompt)
	if len(values) != 1 {
		t.Fatalf("%s: %q does not have exactly one value in it", level, q.Prompt)
	}

	for _, u := range unitPairs {
		switch {
		case u.big == given && u.small == asked:
			want(t, level, q, values[0].mul(whole(u.factor)))
			return
		case u.small == given && u.big == asked:
			got, err := values[0].div(whole(u.factor))
			if err != nil {
				t.Fatalf("%s: %v", level, err)
			}
			want(t, level, q, got)
			return
		}
	}
	t.Fatalf("%s: %q converts between %q and %q, which is not a pair of units",
		level, q.Prompt, given, asked)
}

func checkRates(t *testing.T, level string, q Question) {
	t.Helper()
	n := intsIn(t, q.Context) // "120 km in 2 hours", "5 pens cost 60p", "60 pages read in 5 days"
	total, count := n[0], n[1]
	if strings.Contains(q.Context, "cost") {
		total, count = n[1], n[0]
	}
	if total%count != 0 {
		t.Fatalf("%s: %q is not a whole rate", level, q.Context)
	}
	wantInt(t, level, q, total/count)
}

func checkSpeed(t *testing.T, level string, q Question) {
	t.Helper()
	n := intsIn(t, q.Context) // "A train goes 180 km in 3 hours[, so 60 km per hour]"
	distance, hours := n[0], n[1]
	if distance%hours != 0 {
		t.Fatalf("%s: %q is not a whole speed", level, q.Context)
	}
	speed := distance / hours

	switch {
	case strings.HasPrefix(q.Prompt, "How fast"):
		wantInt(t, level, q, speed)
	case strings.HasPrefix(q.Prompt, "How far"):
		wantInt(t, level, q, speed*intsIn(t, q.Prompt)[0])
	default: // "How many hours to go 420 km?"
		far := intsIn(t, q.Prompt)[0]
		if far%speed != 0 {
			t.Fatalf("%s: %q does not take a whole number of hours", level, q.Prompt)
		}
		wantInt(t, level, q, far/speed)
	}
}

// ---------------------------------------------------------------------------
// Chapter 8.

// percentOf is a percent of a number, exactly, as a rational.
func percentOf(pct, n int) rat { return newRat(pct*n, 100) }

func checkPercentOf(t *testing.T, level string, q Question) {
	t.Helper()
	n := intsIn(t, q.Prompt) // "15% of 80"
	want(t, level, q, percentOf(n[0], n[1]))
}

func checkPercentForms(t *testing.T, level string, q Question) {
	t.Helper()
	switch {
	case strings.HasSuffix(q.Prompt, "as a percent"):
		// "0.45 as a percent"
		want(t, level, q, valuesIn(q.Prompt)[0].mul(whole(100)))
		wantForm(t, level, q, FormWhole)
	case strings.HasSuffix(q.Prompt, "as a decimal"):
		want(t, level, q, percentOf(intsIn(t, q.Prompt)[0], 1))
		wantForm(t, level, q, FormDec)
	default:
		want(t, level, q, percentOf(intsIn(t, q.Prompt)[0], 1))
		wantForm(t, level, q, FormFrac)
		if num, den := q.Answer.Value(); den > 1 && gcd(num, den) != 1 {
			t.Errorf("%s: %q answers %s, not in lowest terms", level, q.Prompt, q.Answer)
		}
	}
}

func checkPercentChange(t *testing.T, level string, q Question) {
	t.Helper()
	n := intsIn(t, q.Context) // "A coat costs 80p and goes up by 25%"
	start, pct := n[0], n[1]
	change := percentOf(pct, start)
	if strings.Contains(q.Context, "off") {
		want(t, level, q, whole(start).sub(change))
		return
	}
	want(t, level, q, whole(start).add(change))
}

func checkFindingWhole(t *testing.T, level string, q Question) {
	t.Helper()
	n := intsIn(t, q.Context) // "25% of a number is 20"
	pct, part := n[0], n[1]
	want(t, level, q, newRat(part*100, pct))
}

func checkPercentDifference(t *testing.T, level string, q Question) {
	t.Helper()
	n := intsIn(t, q.Context) // "The price goes from 80p to 100p"
	from, to := n[0], n[1]
	want(t, level, q, newRat(abs(to-from)*100, from))
}

// ---------------------------------------------------------------------------
// Chapter 9.

func checkRootBounds(t *testing.T, level string, q Question) {
	t.Helper()
	n := intsIn(t, q.Prompt)[0]
	if strings.HasPrefix(q.Prompt, "Biggest") {
		k := 0
		for (k+1)*(k+1) < n {
			k++
		}
		wantInt(t, level, q, k)
		return
	}
	k := 0
	for k*k <= n {
		k++
	}
	wantInt(t, level, q, k)
}

func checkMultiplyingRoots(t *testing.T, level string, q Question) {
	t.Helper()
	// "sqrt(8) x sqrt(2)": neither root is whole on its own, so they have to
	// be multiplied together before the root is taken.
	n := intsIn(t, q.Prompt)
	if len(n) != 2 {
		t.Fatalf("%s: %q is not two roots multiplied", level, q.Prompt)
	}
	product := n[0] * n[1]
	root := 0
	for root*root < product {
		root++
	}
	if root*root != product {
		t.Fatalf("%s: %q leaves sqrt(%d), which is not whole", level, q.Prompt, product)
	}
	wantInt(t, level, q, root)
}

func checkTidyingRoots(t *testing.T, level string, q Question) {
	t.Helper()
	switch {
	case strings.HasPrefix(q.Prompt, "Biggest square number"):
		n := intsIn(t, q.Prompt)[0]
		best := 1
		for d := 1; d*d <= n; d++ {
			if n%(d*d) == 0 {
				best = d * d
			}
		}
		wantInt(t, level, q, best)
	case strings.HasPrefix(q.Prompt, "sqrt("):
		checkExpression(t, level, q)
	default:
		// "sqrt(72) = 6 x sqrt(?)": square the outside, times what is left,
		// and it has to come back to the number under the original root.
		n := intsIn(t, q.Context) // the number under the root, then the outside
		inside := answerInt(t, level, q)
		if got := n[1] * n[1] * inside; got != n[0] {
			t.Errorf("%s: %q with ? = %d gives sqrt(%d), not sqrt(%d)",
				level, q.Context, inside, got, n[0])
		}
	}
}
