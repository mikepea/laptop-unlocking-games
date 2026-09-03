package maths

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/mikepea/laptop-unlocking-games/internal/games"
)

// The algebra levels build their prompt and their answer separately, same as
// the arithmetic ones, so the same class of mistake is possible: show one
// equation and mark another. These tests re-derive the answer from what is on
// screen rather than trusting the generator.

var coefficient = regexp.MustCompile(`^(\d*)([a-z])$`)

// evalTerms works out a space-separated expression like "3x + 5" or "x / 4",
// with the letters given values. It is deliberately simple-minded: left to
// right, no precedence beyond binding * and / to the term before them, which
// is all the generated forms need.
func evalTerms(tokens []string, vars map[string]int) (int, error) {
	value := func(tok string) (int, error) {
		if n, err := strconv.Atoi(tok); err == nil {
			return n, nil
		}
		if m := coefficient.FindStringSubmatch(tok); m != nil {
			v, ok := vars[m[2]]
			if !ok {
				return 0, fmt.Errorf("unknown variable %q", m[2])
			}
			if m[1] == "" {
				return v, nil
			}
			c, err := strconv.Atoi(m[1])
			return c * v, err
		}
		if v, ok := vars[tok]; ok { // "?" and bare letters
			return v, nil
		}
		return 0, fmt.Errorf("unparseable term %q", tok)
	}

	total, err := value(tokens[0])
	if err != nil {
		return 0, err
	}
	for i := 1; i < len(tokens); i += 2 {
		if i+1 >= len(tokens) {
			return 0, fmt.Errorf("dangling operator %q", tokens[i])
		}
		rhs, err := value(tokens[i+1])
		if err != nil {
			return 0, err
		}
		switch tokens[i] {
		case "+":
			total = total + rhs
		case "-":
			total = total - rhs
		case "*":
			total = total * rhs
		case "/":
			if rhs == 0 || total%rhs != 0 {
				return 0, fmt.Errorf("%d / %d is not whole", total, rhs)
			}
			total = total / rhs
		default:
			return 0, fmt.Errorf("unknown operator %q", tokens[i])
		}
	}
	return total, nil
}

// checkEquation asserts that putting the answer back into the equation on
// screen makes both sides agree.
func checkEquation(t *testing.T, level, context string, answer int) {
	t.Helper()
	sides := strings.Split(context, " = ")
	if len(sides) != 2 {
		t.Fatalf("%s: %q is not an equation", level, context)
	}
	vars := map[string]int{"?": answer, "x": answer}
	lhs, err := evalTerms(strings.Fields(sides[0]), vars)
	if err != nil {
		t.Fatalf("%s: left of %q: %v", level, context, err)
	}
	rhs, err := evalTerms(strings.Fields(sides[1]), vars)
	if err != nil {
		t.Fatalf("%s: right of %q: %v", level, context, err)
	}
	if lhs != rhs {
		t.Errorf("%s: %q with the unknown = %d gives %d = %d", level, context, answer, lhs, rhs)
	}
}

func levelByTitle(t *testing.T, title string) Level {
	t.Helper()
	for _, l := range algebraLevels {
		if l.Title == title {
			return l
		}
	}
	t.Fatalf("no algebra level titled %q", title)
	return Level{}
}

func TestEquationLevelsBalanceWithTheirAnswer(t *testing.T) {
	r := games.NewTestRand(21)
	for _, title := range []string{"Missing Number", "One Step", "Two Steps"} {
		l := levelByTitle(t, title)
		for i := 0; i < 400; i++ {
			q := l.Gen(r)
			checkEquation(t, title, q.Context, q.Answer)
			if q.Answer < 0 {
				t.Fatalf("%s: negative answer in %q", title, q.Context)
			}
		}
	}
}

func TestSubstitutionMatchesTheValuesGiven(t *testing.T) {
	l := levelByTitle(t, "Putting Numbers In")
	r := games.NewTestRand(22)
	for i := 0; i < 400; i++ {
		q := l.Gen(r)
		vars := map[string]int{}
		for _, assign := range strings.Split(q.Context, ", ") {
			parts := strings.Split(assign, " = ")
			if len(parts) != 2 {
				t.Fatalf("bad context %q", q.Context)
			}
			n, err := strconv.Atoi(parts[1])
			if err != nil {
				t.Fatalf("bad value in %q: %v", q.Context, err)
			}
			vars[parts[0]] = n
		}
		got, err := evalTerms(strings.Fields(q.Prompt), vars)
		if err != nil {
			t.Fatalf("prompt %q with %q: %v", q.Prompt, q.Context, err)
		}
		if got != q.Answer {
			t.Errorf("%q where %s is %d, question claims %d", q.Prompt, q.Context, got, q.Answer)
		}
	}
}

func TestCollectingUpCountsEveryMatchingTerm(t *testing.T) {
	l := levelByTitle(t, "Collecting Up")
	r := games.NewTestRand(23)
	for i := 0; i < 400; i++ {
		q := l.Gen(r)
		lhs, _, ok := strings.Cut(q.Context, " = ")
		if !ok {
			t.Fatalf("bad context %q", q.Context)
		}
		total := 0
		for _, tok := range strings.Fields(lhs) {
			m := coefficient.FindStringSubmatch(tok)
			if m == nil || m[2] != "x" {
				continue // a "+", or a y term that must not be counted
			}
			n := 1
			if m[1] != "" {
				n, _ = strconv.Atoi(m[1])
			}
			total += n
		}
		if total != q.Answer {
			t.Errorf("%q has %d xs, question claims %d", q.Context, total, q.Answer)
		}
	}
}

// The third step of a simplify problem asks for "the biggest number that
// divides both". If the coefficients left inside the brackets share a factor
// of their own, the number the question wants is not in fact the biggest one,
// and the answer it marks correct is wrong.
func TestFactorisationIsTheFullestOne(t *testing.T) {
	l := levelByTitle(t, "Simplify and Factorise")
	r := games.NewTestRand(24)
	for i := 0; i < 500; i++ {
		steps := l.GenSteps(r)
		if len(steps) != 4 {
			t.Fatalf("expected 4 steps, got %d", len(steps))
		}
		totalX, totalY, factor := steps[0].Answer, steps[1].Answer, steps[2].Answer

		if want := gcd(totalX, totalY); factor != want {
			t.Fatalf("%q: biggest common factor of %d and %d is %d, question says %d",
				steps[2].Context, totalX, totalY, want, factor)
		}
		if got := steps[3].Answer; got*factor != totalY {
			t.Errorf("%q: %d x %d != %d", steps[3].Context, got, factor, totalY)
		}
	}
}

func TestSimplifyStepsShowTheirWorking(t *testing.T) {
	l := levelByTitle(t, "Simplify and Factorise")
	r := games.NewTestRand(25)
	for i := 0; i < 100; i++ {
		for _, q := range l.GenSteps(r) {
			if q.Context == "" {
				t.Errorf("step %q has nothing on screen to work from", q.Prompt)
			}
			if q.Answer <= 0 {
				t.Errorf("step %q has answer %d", q.Prompt, q.Answer)
			}
		}
	}
}

// A round must never stop part-way through a problem: half-finished working
// teaches nothing and looks like a bug.
func TestSteppedRoundsEndOnAWholeProblem(t *testing.T) {
	l := levelByTitle(t, "Simplify and Factorise")
	rd := newRound(l, games.NewTestRand(26))
	if len(rd.questions)%4 != 0 {
		t.Errorf("round has %d questions, not a whole number of 4-step problems", len(rd.questions))
	}
	if len(rd.questions) < l.Questions {
		t.Errorf("round has %d questions, fewer than the %d asked for", len(rd.questions), l.Questions)
	}
}

// Negative numbers are their own topic and none of these levels teach it. An
// equation is allowed to contain a minus sign as an operator ("x - 6 = 1") but
// never a negative value ("x - 9 = -2"), which would land unannounced on a
// player who is here to find out what a letter means.
func TestNoNegativeNumbersAreShown(t *testing.T) {
	r := games.NewTestRand(27)
	for _, l := range algebraLevels {
		for i := 0; i < 500; i++ {
			var qs []Question
			if l.GenSteps != nil {
				qs = l.GenSteps(r)
			} else {
				qs = []Question{l.Gen(r)}
			}
			for _, q := range qs {
				for _, text := range []string{q.Context, q.Prompt} {
					for _, tok := range strings.Fields(text) {
						if n, err := strconv.Atoi(tok); err == nil && n < 0 {
							t.Fatalf("level %q shows a negative number in %q", l.Title, text)
						}
					}
				}
			}
		}
	}
}
