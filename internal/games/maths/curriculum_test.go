package maths

import (
	"math/rand/v2"
	"regexp"
	"strings"
	"testing"

	"github.com/mikepea/laptop-unlocking-games/internal/games"
)

// Every level in the curriculum has to have its answers re-derived from what
// it puts on screen, by code that shares nothing with the generator. A
// question whose prompt and answer disagree is unwinnable in a way a child
// reads as their own failure, and it is the single most likely bug in a file
// of sixty generators.
//
// answerCheckers is where that promise is kept. The test below fails on a
// level that is not in it, so adding a level without a check is not something
// that can be forgotten.

// checker re-derives the answer to one question. It is given the level title
// for error messages only.
type checker func(t *testing.T, level string, q Question)

// howManyQuestions is how many go through each checker. Generators branch on
// the random source, so a few hundred is what it takes to see every shape.
const howManyQuestions = 300

func TestEveryLevelAgreesWithWhatItShows(t *testing.T) {
	for ci, c := range Chapters {
		for li, l := range c.Levels {
			check, ok := answerCheckers[l.Title]
			if !ok {
				t.Errorf("level %q has no entry in answerCheckers", l.Title)
				continue
			}
			t.Run(l.Title, func(t *testing.T) {
				r := games.NewTestRand(uint64(1000 + ci*100 + li))
				for i := 0; i < howManyQuestions; i++ {
					for _, q := range questionsFrom(l, r) {
						check(t, l.Title, q)
					}
				}
			})
		}
	}
}

// questionsFrom asks a level for one question, or for one whole multi-step
// problem, whichever kind it is.
func questionsFrom(l Level, r *rand.Rand) []Question {
	if l.GenSteps != nil {
		return l.GenSteps(r)
	}
	return []Question{l.Gen(r)}
}

// TestEveryQuestionIsTypeableAndShowable is the check that does not care what
// the answer is, only that the question can be got onto a console and the
// answer can be got into the field.
func TestEveryQuestionIsTypeableAndShowable(t *testing.T) {
	for _, c := range Chapters {
		for _, l := range c.Levels {
			r := games.NewTestRand(99)
			for i := 0; i < 200; i++ {
				for _, q := range questionsFrom(l, r) {
					if q.Prompt == "" {
						t.Fatalf("level %q produced a question with no prompt", l.Title)
					}
					for _, text := range []string{q.Context, q.Prompt, q.Answer.String()} {
						if bad := nonASCII(text); bad != "" {
							t.Fatalf("level %q shows %q, which is not ASCII", l.Title, bad)
						}
						if strings.Contains(text, "%!") {
							t.Fatalf("level %q has a broken format string in %q", l.Title, text)
						}
					}
					// The field is eight characters wide. An answer that does
					// not fit cannot be typed at all.
					if n := len(q.Answer.String()); n > 8 {
						t.Fatalf("level %q wants %q typed, which is %d characters",
							l.Title, q.Answer, n)
					}
					// The answer as it is written on the correction screen has
					// to be one the marker accepts. Anything else tells a
					// child the right answer and then refuses it.
					if !q.Answer.Matches(q.Answer.String()) {
						t.Fatalf("level %q shows the answer as %q but will not accept it",
							l.Title, q.Answer)
					}
				}
			}
		}
	}
}

func nonASCII(s string) string {
	for _, r := range s {
		if r > 127 {
			return string(r)
		}
	}
	return ""
}

// ---------------------------------------------------------------------------
// Reading what is on screen.

var numberPattern = regexp.MustCompile(`\d+(?:\.\d+)?`)

// valuesIn pulls the numbers out of a line of text, in the order they are
// written. It has no idea what they mean; the checkers do.
func valuesIn(s string) []rat {
	var out []rat
	for _, tok := range numberPattern.FindAllString(s, -1) {
		v, err := parseNumeric(tok)
		if err != nil {
			continue
		}
		out = append(out, v)
	}
	return out
}

// intsIn is valuesIn where every number is whole, which is most of them.
func intsIn(t *testing.T, s string) []int {
	t.Helper()
	var out []int
	for _, v := range valuesIn(s) {
		if v.d != 1 {
			t.Fatalf("expected whole numbers in %q, found %s", s, v)
		}
		out = append(out, v.n)
	}
	return out
}

// want asserts that the answer to q is the value worked out independently.
func want(t *testing.T, level string, q Question, v rat) {
	t.Helper()
	if got := ratOf(q.Answer); !got.eq(v) {
		t.Errorf("%s: %q / %q answers %s, but works out to %s",
			level, q.Context, q.Prompt, got, v)
	}
}

func wantInt(t *testing.T, level string, q Question, n int) {
	t.Helper()
	want(t, level, q, whole(n))
}

// wantForm asserts the answer has to be written in a particular shape, which
// is the whole point of the levels that set one.
func wantForm(t *testing.T, level string, q Question, form Form) {
	t.Helper()
	if q.Answer.Form() != form {
		t.Errorf("%s: %q wants form %v, got %v", level, q.Prompt, form, q.Answer.Form())
	}
}

// ---------------------------------------------------------------------------
// The checkers that read an expression straight off the screen.

// checkExpression evaluates the prompt and compares. For a prompt that is
// only "?", it puts the answer back into the equation in the context and
// checks that both sides come out the same, which is the same promise made
// the other way round.
func checkExpression(t *testing.T, level string, q Question) {
	t.Helper()
	if strings.TrimSpace(q.Prompt) != "?" {
		v, err := evalExpr(q.Prompt)
		if err != nil {
			t.Fatalf("%s: cannot read prompt %q: %v", level, q.Prompt, err)
		}
		want(t, level, q, v)
		return
	}

	filled := strings.Replace(q.Context, "?", q.Answer.String(), 1)
	lhs, rhs, ok := strings.Cut(filled, " = ")
	if !ok {
		t.Fatalf("%s: prompt is ? but context %q is not an equation", level, q.Context)
	}
	left, err := evalExpr(lhs)
	if err != nil {
		t.Fatalf("%s: cannot read %q: %v", level, lhs, err)
	}
	right, err := evalExpr(rhs)
	if err != nil {
		t.Fatalf("%s: cannot read %q: %v", level, rhs, err)
	}
	if !left.eq(right) {
		t.Errorf("%s: %q with ? = %s gives %s = %s", level, q.Context, q.Answer, left, right)
	}
}

// wordyOperators are the places a level spells an operator out because the
// symbol would be ambiguous on screen. Undoing them turns the prompt back
// into something the evaluator can read.
var wordyOperators = strings.NewReplacer(
	" divided by ", " / ",
	" of ", " x ",
	" and ", " + ",
	" squared", " ^2",
	" in simplest form", "",
	" as a top-heavy fraction", "",
	" as a decimal", "",
	" as a fraction", "",
)

// checkWordyExpression is checkExpression for the prompts that are written
// out in words.
func checkWordyExpression(t *testing.T, level string, q Question) {
	t.Helper()
	checkExpression(t, level, Question{
		Context: wordyOperators.Replace(q.Context),
		Prompt:  wordyOperators.Replace(q.Prompt),
		Answer:  q.Answer,
	})
}

// checkFractionForm is checkWordyExpression plus the promise that the answer
// has to be typed as a fraction. Chapter 4 is not chapter 6.
func checkFractionForm(t *testing.T, level string, q Question) {
	t.Helper()
	checkWordyExpression(t, level, q)
	if strings.TrimSpace(q.Prompt) == "?" {
		return // the "?" questions in chapter 4 ask for a whole number
	}
	wantForm(t, level, q, FormFrac)
	if num, den := q.Answer.Value(); den > 1 && gcd(num, den) != 1 {
		t.Errorf("%s: %q answers %s, which is not in lowest terms", level, q.Prompt, q.Answer)
	}
}

// checkDecimalForm is the same promise for chapter 6.
func checkDecimalForm(t *testing.T, level string, q Question) {
	t.Helper()
	checkWordyExpression(t, level, q)
	wantForm(t, level, q, FormDec)
}
