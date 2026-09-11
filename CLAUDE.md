# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```sh
make check                        # fmt (gofmt -l, fails if non-empty), vet, test — run before committing
make run                          # build to bin/ and play
make dist                         # CGO_ENABLED=0 linux/amd64 build into dist/
make release                      # dist/ plus dist/manifest.json via scripts/release.sh

go test ./internal/games/maths/   # one package
go test ./internal/games/maths/ -run TestSteppedRoundsEndOnAWholeProblem
go test ./... -count=1            # defeat the test cache
```

Go 1.26.5, pinned in `.tool-versions` (asdf). Version/commit/date are injected
through `-ldflags` into `internal/version`, so a plain `go build` produces a
binary that reports "dev" — use `make build` when the version matters.

## Read first

`README.md` (what the games are and how they ladder) and
`docs/architecture.md` (the design decisions, the points/achievements/unlocks
split, the scoring curve, the update mechanism, and what is deliberately not
built). Both are current and worth trusting. What follows is only the part a
change is likely to trip over.

## Architecture in one paragraph

One Go binary is the login shell, the menu, every game, and its own updater.
`cmd/unlock` parses flags and wires up a `profile.Store`, a `points.Backend`
and an `update.Checker`, then hands them to `launcher.New` — the root Bubble
Tea model, which swaps a per-session game model into `Model.active` and takes
control back on `games.FinishedMsg`. All state is one JSON file, written
atomically. No daemon, no database, no network required to play.

## Rules that are easy to break

**Games never write to the profile.** `Game.New(p *profile.Profile)` hands over
the profile so a game can gate its own internal levels (`p.Stats(GameID).LessonsCleared`);
persistence belongs to `launcher.apply`, which acts on the returned
`games.Result`. A game that mutates `p` will have its change silently
overwritten or double-counted.

**Adding a game means editing `internal/catalog`.** That list is what both the
binary and the cross-cutting tests read (`catalog_test.go` checks unique IDs, a
`Stage()` that exists on `unlocks.Ladder`, and that every game renders on open
*and* after Enter). A game registered anywhere else is untested.

**Keep scoring out of the `tea.Model`.** Every game splits the same way: a
plain struct holding the rules, with an injectable `now func() time.Time` and
an injectable `*rand.Rand` (`typing.session`, `maths.round`, `spelling.round`,
`codebreaker.puzzle`, `pseudocode.round`, `shellquest.shell`), tested with no
terminal involved; the model does key handling and rendering only. Use
`games.NewRand()` in production and `games.NewTestRand(seed)` in tests.

**ASCII only in anything rendered.** This runs on a Linux virtual console. No
emoji, no `✓`, no em dashes, no curly quotes — a missing glyph shows as a blank
or a box, and double-width characters wreck column alignment. Use the
four-character badges in `internal/ui` (`BadgeDone`/`BadgeOpen`/`BadgeLock`).
The progress-bar block characters are the single exception;
`deploy/unlock-session.sh` sets a console font that has them.

**Content needs an invariant test that re-derives the answer.** A generated
question whose prompt and answer disagree, a spelling word the answer field's
filter would reject, a lesson line containing a character not on a keyboard, a
Shell Quest secret present in no file — each is unwinnable in a way a child
reads as their own failure. See `maths/curriculum_test.go`,
`spelling/round_test.go` (`isTypeable`) and
`typing/session_test.go` (`TestCuratedLessonsAreWellFormed`) for the pattern.

**Adding a maths level means adding a checker.** `maths.Chapters` follows the
AoPS Prealgebra book; `maths.Levels` is that flattened, and
`GameStats.LessonsCleared` indexes the flat list, so append to the end of a
chapter and never insert. Every level must have an entry in
`maths.answerCheckers` (`curriculum_test.go`, `checkers_test.go`) or the
test suite fails. The checker re-reads the
question off the screen — via the exact-rational evaluator in `eval_test.go`
for anything that prints a sum, or by pulling numbers out of the sentence —
and must not share code with the generator.

**A maths answer is not an int.** `maths.Answer` is an exact rational plus a
`Form`: `Int(7)`, `Fraction(3, 4)` (typed in lowest terms) or
`Decimal(375, 3)` (0.375, kept as a scaled integer, never a float). A
fraction-form answer refuses the decimal of the same value and vice versa,
because in chapters 4 and 6 the shape of the answer is the thing being
taught. Answers must fit the eight-character field.

**A round grows as it is played.** In `maths`, `spelling` and `pseudocode` a
wrong answer puts its question back at the end of the queue for one more go
(`round.firstPass` is where the second attempts start), so the question slice
changes length mid-round and both attempts count towards accuracy. Nothing may
cache the round length. `codebreaker` is deliberately exempt — a wrong guess
is already its whole loop.

**Key handling normalises in one place.** `launcher.normaliseEnter` rewrites
ctrl+j (the line feed the keypad's Enter sends) to `tea.KeyEnter` before any
message reaches the menu or a game. Games switch on `tea.KeyEnter` and stay
unaware of it; do not re-handle it per game, or the next game added loses the
keypad.

**`View()` runs at least once on a finished round.** Bubble Tea renders the
model it was handed before `games.Finish`'s message swaps it out, so accessors
like `round.current()` must clamp rather than panic when the index is one past
the end. There are `finish_test.go` files in `maths` and `pseudocode` guarding
exactly this — copy them for a new game.

**Points only go up.** `Profile.AwardPoints` ignores non-positive awards by
design. `unlocks` costs are lifetime `PointsEarned`, not balance, so adding a
rung to the ladder retroactively grants it to anyone already past its cost —
`launcher.syncUnlocks` handles that on every run.

**Lesson progression is order-sensitive.** `GameStats.LessonsCleared` only
advances when the round passed is the round the player was up to; replaying an
earlier level pays points without skipping ahead.
