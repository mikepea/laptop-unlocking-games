package maths

// Small whole-number helpers shared across the chapters. They are here rather
// than in the chapter that first needed them because number theory, fractions
// and factorising all want the same three functions.

// gcd is Euclid. It is always positive, and gcd(0, n) is n.
func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	if a < 0 {
		return -a
	}
	return a
}

// lcm is the lowest common multiple. lcm(0, n) is 0.
func lcm(a, b int) int {
	if a == 0 || b == 0 {
		return 0
	}
	n := a / gcd(a, b) * b
	if n < 0 {
		return -n
	}
	return n
}

// abs is the absolute value of an int.
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// pow is b to the e, for small whole exponents. e must not be negative.
func pow(b, e int) int {
	n := 1
	for i := 0; i < e; i++ {
		n *= b
	}
	return n
}
