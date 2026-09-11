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
	for _, l := range equationLevels {
		if l.Title == title {
			return l
		}
	}
	t.Fatalf("no chapter 5 level titled %q", title)
	return Level{}
}

// The tests below are the ones that look across a whole multi-step problem.
// Per-question checking lives in answerCheckers, which covers every level in
// every chapter.

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
		totalX := answerInt(t, "Simplify and Factorise", steps[0])
		totalY := answerInt(t, "Simplify and Factorise", steps[1])
		factor := answerInt(t, "Simplify and Factorise", steps[2])

		if want := gcd(totalX, totalY); factor != want {
			t.Fatalf("%q: biggest common factor of %d and %d is %d, question says %d",
				steps[2].Context, totalX, totalY, want, factor)
		}
		if got := answerInt(t, "Simplify and Factorise", steps[3]); got*factor != totalY {
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
			if n, _ := q.Answer.Whole(); n <= 0 {
				t.Errorf("step %q has answer %s", q.Prompt, q.Answer)
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

// Negative numbers are chapter 1's topic, and chapter 5 does not revisit
// them. An equation is allowed to contain a minus sign as an operator
// ("x - 6 = 1") but never a negative value ("x - 9 = -2"), which would land
// unannounced on a player who is here to find out what a letter means.
func TestChapterFiveShowsNoNegativeNumbers(t *testing.T) {
	r := games.NewTestRand(27)
	for _, l := range equationLevels {
		for i := 0; i < 500; i++ {
			for _, q := range questionsFrom(l, r) {
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
