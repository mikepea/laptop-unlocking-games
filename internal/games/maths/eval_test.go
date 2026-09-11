package maths

import (
	"fmt"
	"strconv"
	"strings"
)

// An exact-rational expression evaluator, written for the tests only.
//
// Every level builds its prompt and its answer separately, so the mistake that
// matters is showing one sum and marking another. The defence is to read what
// is on screen and work it out again, with code that shares nothing with the
// generator. That is what this is: a tokeniser and a recursive-descent parser
// over the notation the game actually prints -- "x" for times, "^" for a
// power, "sqrt(n)" for a root, fractions and decimals as they are written.
//
// It is exact, so 0.1 + 0.2 really is 0.3 here, and a level whose answer is
// 1/3 is compared against 1/3 rather than against 0.333.

// rat is an exact rational, always reduced, always with a positive
// denominator.
type rat struct{ n, d int }

func newRat(n, d int) rat {
	if d == 0 {
		panic("rat: zero denominator")
	}
	if d < 0 {
		n, d = -n, -d
	}
	if g := gcd(n, d); g > 1 {
		n, d = n/g, d/g
	}
	return rat{n, d}
}

func whole(n int) rat { return rat{n, 1} }

func (a rat) add(b rat) rat { return newRat(a.n*b.d+b.n*a.d, a.d*b.d) }
func (a rat) sub(b rat) rat { return newRat(a.n*b.d-b.n*a.d, a.d*b.d) }
func (a rat) mul(b rat) rat { return newRat(a.n*b.n, a.d*b.d) }

func (a rat) div(b rat) (rat, error) {
	if b.n == 0 {
		return rat{}, fmt.Errorf("division by zero")
	}
	return newRat(a.n*b.d, a.d*b.n), nil
}

func (a rat) eq(b rat) bool { return a.n == b.n && a.d == b.d }

func (a rat) String() string {
	if a.d == 1 {
		return strconv.Itoa(a.n)
	}
	return fmt.Sprintf("%d/%d", a.n, a.d)
}

// ratOf is an Answer as a rational, so a re-derived value can be compared
// against what the game will mark correct.
func ratOf(a Answer) rat {
	n, d := a.Value()
	return newRat(n, d)
}

// evalExpr works out an expression written the way the game prints one.
func evalExpr(s string) (rat, error) {
	tokens, err := tokenise(s)
	if err != nil {
		return rat{}, err
	}
	p := &parser{tokens: tokens}
	v, err := p.expr()
	if err != nil {
		return rat{}, err
	}
	if p.pos != len(p.tokens) {
		return rat{}, fmt.Errorf("%q: trailing %q", s, p.tokens[p.pos])
	}
	return v, nil
}

// tokenise splits an expression into numbers, operators, brackets and the one
// word the notation uses.
func tokenise(s string) ([]string, error) {
	var out []string
	for i := 0; i < len(s); {
		c := s[i]
		switch {
		case c == ' ':
			i++
		case c >= '0' && c <= '9':
			j := i
			for j < len(s) && (s[j] >= '0' && s[j] <= '9' || s[j] == '.') {
				j++
			}
			out, i = append(out, s[i:j]), j
		case c >= 'a' && c <= 'z':
			j := i
			for j < len(s) && s[j] >= 'a' && s[j] <= 'z' {
				j++
			}
			word := s[i:j]
			if word != "x" && word != "sqrt" {
				return nil, fmt.Errorf("%q: unexpected word %q", s, word)
			}
			out, i = append(out, word), j
		case strings.ContainsRune("+-/^()", rune(c)):
			out, i = append(out, string(c)), i+1
		default:
			return nil, fmt.Errorf("%q: unexpected character %q", s, string(c))
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%q: nothing to evaluate", s)
	}
	return out, nil
}

type parser struct {
	tokens []string
	pos    int
}

func (p *parser) peek() string {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos]
	}
	return ""
}

func (p *parser) take() string {
	t := p.peek()
	p.pos++
	return t
}

// expr is a run of terms joined by plus and minus, left to right.
func (p *parser) expr() (rat, error) {
	v, err := p.term()
	if err != nil {
		return rat{}, err
	}
	for p.peek() == "+" || p.peek() == "-" {
		op := p.take()
		rhs, err := p.term()
		if err != nil {
			return rat{}, err
		}
		if op == "+" {
			v = v.add(rhs)
		} else {
			v = v.sub(rhs)
		}
	}
	return v, nil
}

// term is a run of factors joined by times and divide, which bind tighter.
func (p *parser) term() (rat, error) {
	v, err := p.unary()
	if err != nil {
		return rat{}, err
	}
	for p.peek() == "x" || p.peek() == "/" {
		op := p.take()
		rhs, err := p.unary()
		if err != nil {
			return rat{}, err
		}
		if op == "x" {
			v = v.mul(rhs)
			continue
		}
		if v, err = v.div(rhs); err != nil {
			return rat{}, err
		}
	}
	return v, nil
}

// unary is a minus sign in front of something, as in "(-3)" and "2^-3".
func (p *parser) unary() (rat, error) {
	if p.peek() == "-" {
		p.take()
		v, err := p.unary()
		if err != nil {
			return rat{}, err
		}
		return whole(0).sub(v), nil
	}
	return p.power()
}

// power binds tighter than everything but brackets, and its exponent must be
// a whole number: nothing the game prints raises anything to a fraction.
func (p *parser) power() (rat, error) {
	base, err := p.primary()
	if err != nil {
		return rat{}, err
	}
	if p.peek() != "^" {
		return base, nil
	}
	p.take()
	e, err := p.unary()
	if err != nil {
		return rat{}, err
	}
	if e.d != 1 {
		return rat{}, fmt.Errorf("exponent %s is not whole", e)
	}
	v := whole(1)
	for i := 0; i < abs(e.n); i++ {
		v = v.mul(base)
	}
	if e.n < 0 {
		return whole(1).div(v)
	}
	return v, nil
}

func (p *parser) primary() (rat, error) {
	switch t := p.take(); {
	case t == "(":
		v, err := p.expr()
		if err != nil {
			return rat{}, err
		}
		if p.take() != ")" {
			return rat{}, fmt.Errorf("missing closing bracket")
		}
		return v, nil
	case t == "sqrt":
		if p.take() != "(" {
			return rat{}, fmt.Errorf("sqrt without a bracket")
		}
		v, err := p.expr()
		if err != nil {
			return rat{}, err
		}
		if p.take() != ")" {
			return rat{}, fmt.Errorf("missing closing bracket after sqrt")
		}
		root, err := exactRoot(v)
		if err != nil {
			return rat{}, err
		}
		return root, nil
	case t == "":
		return rat{}, fmt.Errorf("ran off the end of the expression")
	default:
		return parseNumeric(t)
	}
}

// exactRoot is the square root, and an error if there is not a whole one. A
// level that shows an irrational root has to ask for it in pieces, and one
// that does not is a bug worth failing on.
func exactRoot(v rat) (rat, error) {
	n, d := intRoot(v.n), intRoot(v.d)
	if n == 0 || d == 0 {
		return rat{}, fmt.Errorf("sqrt(%s) is not a whole root", v)
	}
	return newRat(n, d), nil
}

func intRoot(n int) int {
	for i := 0; i*i <= n; i++ {
		if i*i == n {
			return i
		}
	}
	return 0
}

// parseNumeric reads one number token: "7" or "0.375".
func parseNumeric(t string) (rat, error) {
	if intPart, fracPart, cut := strings.Cut(t, "."); cut {
		w, err1 := strconv.Atoi(intPart)
		f, err2 := strconv.Atoi(fracPart)
		if err1 != nil || err2 != nil {
			return rat{}, fmt.Errorf("bad decimal %q", t)
		}
		scale := pow(10, len(fracPart))
		return newRat(w*scale+f, scale), nil
	}
	n, err := strconv.Atoi(t)
	if err != nil {
		return rat{}, fmt.Errorf("bad number %q", t)
	}
	return whole(n), nil
}
