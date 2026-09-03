package pseudocode

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/mikepea/laptop-unlocking-games/internal/games"
)

// Every generator works its answer out as it builds the program, so the two
// can disagree — show one program and mark another. trace is a second,
// independent implementation: it reads the rendered lines the way the player
// does and reports what they print. If the two ever differ, the game is lying
// to him.

type vars map[string]int

// value resolves a token that is either a number or a variable.
func (v vars) value(tok string) (int, error) {
	if n, err := strconv.Atoi(tok); err == nil {
		return n, nil
	}
	if n, ok := v[tok]; ok {
		return n, nil
	}
	return 0, fmt.Errorf("unknown name %q", tok)
}

// evalRHS works out "5", "x + 3" or "x * 2".
func (v vars) evalRHS(toks []string) (int, error) {
	switch len(toks) {
	case 1:
		return v.value(toks[0])
	case 3:
		a, err := v.value(toks[0])
		if err != nil {
			return 0, err
		}
		b, err := v.value(toks[2])
		if err != nil {
			return 0, err
		}
		switch toks[1] {
		case "+":
			return a + b, nil
		case "-":
			return a - b, nil
		case "*":
			return a * b, nil
		}
		return 0, fmt.Errorf("unknown operator %q", toks[1])
	}
	return 0, fmt.Errorf("cannot evaluate %q", strings.Join(toks, " "))
}

// trace runs the program and returns what it printed.
func trace(lines []string) (int, error) {
	v := vars{}
	printed := 0
	seenPrint := false

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(line, indent) {
			return 0, fmt.Errorf("line %d is indented but nothing opened a block: %q", i, line)
		}
		toks := strings.Fields(line)
		if len(toks) == 0 {
			continue
		}

		// A header line ending in ":" owns the indented lines beneath it.
		if strings.HasSuffix(line, ":") {
			var block []string
			for j := i + 1; j < len(lines) && strings.HasPrefix(lines[j], indent); j++ {
				block = append(block, strings.TrimPrefix(lines[j], indent))
				i = j
			}
			if len(block) == 0 {
				return 0, fmt.Errorf("%q opens a block with nothing in it", line)
			}

			runs := 0
			switch toks[0] {
			case "repeat": // repeat N times:
				n, err := strconv.Atoi(toks[1])
				if err != nil {
					return 0, err
				}
				runs = n
			case "if": // if x > 10:
				subject, err := v.value(toks[1])
				if err != nil {
					return 0, err
				}
				want, err := strconv.Atoi(strings.TrimSuffix(toks[3], ":"))
				if err != nil {
					return 0, err
				}
				switch toks[2] {
				case ">":
					if subject > want {
						runs = 1
					}
				case "<":
					if subject < want {
						runs = 1
					}
				default:
					return 0, fmt.Errorf("unknown comparison %q", toks[2])
				}
			default:
				return 0, fmt.Errorf("unknown block %q", line)
			}

			for r := 0; r < runs; r++ {
				for _, inner := range block {
					it := strings.Fields(inner)
					if len(it) < 3 || it[1] != "=" {
						return 0, fmt.Errorf("block line is not an assignment: %q", inner)
					}
					n, err := v.evalRHS(it[2:])
					if err != nil {
						return 0, err
					}
					v[it[0]] = n
				}
			}
			continue
		}

		switch {
		case toks[0] == "print":
			n, err := v.value(toks[1])
			if err != nil {
				return 0, err
			}
			printed, seenPrint = n, true
		case len(toks) >= 3 && toks[1] == "=":
			n, err := v.evalRHS(toks[2:])
			if err != nil {
				return 0, err
			}
			v[toks[0]] = n
		default:
			return 0, fmt.Errorf("cannot run %q", line)
		}
	}
	if !seenPrint {
		return 0, fmt.Errorf("program never prints anything")
	}
	return printed, nil
}

func TestEveryProgramPrintsWhatItClaims(t *testing.T) {
	r := games.NewTestRand(31)
	for _, l := range Levels {
		for i := 0; i < 400; i++ {
			p := l.Gen(r)
			got, err := trace(p.Lines)
			if err != nil {
				t.Fatalf("level %q produced a program that will not run:\n%s\n%v",
					l.Title, strings.Join(p.Lines, "\n"), err)
			}
			if got != p.Answer {
				t.Fatalf("level %q:\n%s\nprints %d, question claims %d",
					l.Title, strings.Join(p.Lines, "\n"), got, p.Answer)
			}
		}
	}
}

// Negative numbers are a topic of their own and none of these levels teach it.
func TestNoNegativeNumbersAreShown(t *testing.T) {
	r := games.NewTestRand(32)
	for _, l := range Levels {
		for i := 0; i < 400; i++ {
			p := l.Gen(r)
			if p.Answer < 0 {
				t.Fatalf("level %q prints %d", l.Title, p.Answer)
			}
			for _, line := range p.Lines {
				for _, tok := range strings.Fields(line) {
					if n, err := strconv.Atoi(tok); err == nil && n < 0 {
						t.Fatalf("level %q shows a negative number in %q", l.Title, line)
					}
				}
			}
		}
	}
}

// The reassignment level exists to teach one thing: that "x = x + 2" takes
// what is in x, changes it, and puts it back. If a round of it could come out
// without a single line where the variable is on both sides, it would not be
// teaching that at all.
func TestReassignmentLevelAlwaysReassigns(t *testing.T) {
	var l Level
	for _, cand := range Levels {
		if cand.Title == "Changing Your Mind" {
			l = cand
		}
	}
	if l.Title == "" {
		t.Fatal("no level titled \"Changing Your Mind\"")
	}

	r := games.NewTestRand(33)
	for i := 0; i < 400; i++ {
		p := l.Gen(r)
		found := false
		for _, line := range p.Lines {
			toks := strings.Fields(line)
			// "x = x + 2": the name on the left appears again on the right.
			if len(toks) >= 4 && toks[1] == "=" && toks[2] == toks[0] {
				found = true
			}
		}
		if !found {
			t.Fatalf("no line changes a box using itself:\n%s", strings.Join(p.Lines, "\n"))
		}
	}
}

func TestLevelsAreWellFormed(t *testing.T) {
	for i, l := range Levels {
		if l.Title == "" {
			t.Errorf("level %d has no title", i)
		}
		if l.Hint == "" {
			t.Errorf("level %q has no hint", l.Title)
		}
		if l.Questions <= 0 {
			t.Errorf("level %q asks %d questions", l.Title, l.Questions)
		}
		if l.MinAccuracy <= 0 || l.MinAccuracy > 1 {
			t.Errorf("level %q has MinAccuracy %v", l.Title, l.MinAccuracy)
		}
		if l.Gen == nil {
			t.Errorf("level %q has no generator", l.Title)
		}
	}
}

// Programs have to fit an 80-column TTY with the frame around them.
func TestProgramsFitTheScreen(t *testing.T) {
	r := games.NewTestRand(34)
	for _, l := range Levels {
		for i := 0; i < 200; i++ {
			for _, line := range l.Gen(r).Lines {
				if len(line) > 40 {
					t.Errorf("level %q line is %d chars: %q", l.Title, len(line), line)
				}
			}
		}
	}
}
