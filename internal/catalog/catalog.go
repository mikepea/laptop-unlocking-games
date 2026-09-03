// Package catalog is the one place that knows which games this build ships.
//
// It exists so the binary and the tests agree: a game added here is registered
// everywhere, and the invariants that hold across the whole set — unique IDs,
// every gating stage present on the ladder — can be checked in one place.
package catalog

import (
	"github.com/mikepea/laptop-unlocking-games/internal/games"
	"github.com/mikepea/laptop-unlocking-games/internal/games/codebreaker"
	"github.com/mikepea/laptop-unlocking-games/internal/games/maths"
	"github.com/mikepea/laptop-unlocking-games/internal/games/pseudocode"
	"github.com/mikepea/laptop-unlocking-games/internal/games/shellquest"
	"github.com/mikepea/laptop-unlocking-games/internal/games/spelling"
	"github.com/mikepea/laptop-unlocking-games/internal/games/typing"
)

// Registry returns every game, in the order they were added to the project.
// The launcher sorts them for display, so this order does not decide the menu.
func Registry() *games.Registry {
	return games.NewRegistry(
		typing.New(),
		maths.New(),
		spelling.New(),
		codebreaker.New(),
		pseudocode.New(),
		shellquest.New(),
	)
}
