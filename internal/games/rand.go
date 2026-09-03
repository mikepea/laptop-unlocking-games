package games

import (
	"math/rand/v2"
	"time"
)

// NewRand returns a randomness source seeded from the clock.
//
// Every game that generates its own content takes a *rand.Rand rather than
// calling the package-level functions, so a test can pin the seed and get the
// same questions every run.
func NewRand() *rand.Rand {
	seed := uint64(time.Now().UnixNano())
	return rand.New(rand.NewPCG(seed, seed>>32))
}

// NewTestRand returns a source with a fixed seed, for tests and for anywhere
// reproducible content matters.
func NewTestRand(seed uint64) *rand.Rand {
	return rand.New(rand.NewPCG(seed, seed*2654435761))
}
