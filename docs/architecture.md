# Architecture

## The shape of the thing

One binary. It is the login shell on tty1, it is the menu, it is every game, and
it is its own updater. There is no daemon, no database, and no service to keep
running — the whole state of the world is a single JSON file in the player's
home directory.

That constraint is deliberate. A laptop that is off for three weeks and then
opened in a car with no wifi has to work exactly as well as one that is online.

```
  agetty --autologin ──► bash ──► exec unlock
                                    │
                                    ├── launcher (root tea.Model)
                                    │     ├── menu ── games.Registry
                                    │     ├── game ── one tea.Model per session
                                    │     ├── results ── achievements + unlocks
                                    │     └── progress ── the ladder
                                    │
                                    ├── profile.Store  ~/.local/share/unlock/profile.json
                                    ├── points.Backend  local now, TaskBank later
                                    └── update.Checker  background, silent on failure

  systemd timer ──► unlock update -apply ──► replaces /usr/local/bin/unlock
```

## Progress model

Three separate ideas, kept separate on purpose:

**Points** are the currency. They only ever go up (`PointsEarned`), and
`PointsSpent` tracks redemption against unlocks. Games award them; nothing takes
them away. `Profile.AwardPoints` silently ignores a negative award so a buggy
game can never cost a child their progress.

A point is a minute of screen time. Games do not award points directly: each
reports a `Score` on its own scale, and `launcher.apply` converts it with
`points.FromScore`, at `points.ScorePerPoint` (30) score to the point, rounded,
with a passed round always paying at least one. That keeps the exchange rate in
one place: change the constant and every game moves together. The ladder's
costs are in points, so a stage's cost is also roughly minutes of play.

**Achievements** are badges. They are one-shot, evaluated after every run
against the profile *as it stands after that run*, so a rule can look either at
the run itself (`r.WPM >= 35`) or at running totals (`totalPlays(p) >= 10`).

**Unlocks** are the ladder in `internal/unlocks`. Each stage has a lifetime
points cost. `launcher.syncUnlocks` records every stage the total has paid for,
which means adding a new rung to the ladder in a later release retroactively
grants it to a player who is already past its cost.

Within a game, progression is separate again: `GameStats.LessonsCleared` only
advances when the round passed is the round the player was actually up to.
Replaying an old lesson pays points but does not skip ahead, and neither does
passing a later lesson out of order.

## Adding a game

Implement `games.Game` (`ID`, `Title`, `Blurb`, `Stage`, `New`) and add it to
`internal/catalog`:

```go
return games.NewRegistry(typing.New(), ..., yourgame.New())
```

The catalog is the single list the binary and the tests both read, so a game
added there is covered by the cross-cutting checks: unique IDs, a gating stage
that actually exists on the ladder, and a play screen that renders.

`New` returns a fresh `tea.Model` for one session. It is handed the profile so
it can gate its own internal levels, but it must not write to it — the launcher
owns persistence. Hand control back with `games.Finish(result)` when a round
ends, or `games.Exit()` when the player backs out without playing.

Set `Stage()` to a rung of the ladder. A game whose stage is not yet unlocked
still appears in the menu, greyed out with its cost, because something visible
to aim at is the entire point.

Keep the scoring logic out of the `tea.Model`. Every game follows the same
split: a plain struct with an injectable clock and an injectable `*rand.Rand`
(`typing.session`, `maths.round`, `spelling.round`, `codebreaker.puzzle`,
`shellquest.shell`), tested without a terminal anywhere near it, and a model
reduced to key handling and rendering.

Content gets its own test. A generated question whose prompt and answer
disagree, a spelling the answer field cannot accept, a Shell Quest whose secret
appears in no file — each of these is unwinnable in a way a child would read as
their own failure, so each has an invariant test that re-derives the answer
independently of the code that produced it.

Maths Sprint is where that is done most thoroughly, because it has the most
generators. `maths.answerCheckers` maps every level to a function that reads
the question off the screen and works it out again — through an exact-rational
expression evaluator for the levels that print a sum, and by pulling the
numbers out of the sentence for the ones that print a word problem. A level
missing from that map fails the test, so a new level cannot be added without a
check. Nothing in the checkers shares code with the generator it checks: the
conversion factors, the prime tests and the divisor counts are all written out
a second time on purpose.

### Answers in Maths Sprint

`maths.Answer` is an exact rational plus a `Form` saying how it has to be
written: a whole number, a fraction in lowest terms, or a decimal. It is not a
float, so `0.1 + 0.2` is `0.3` and nobody is marked wrong by a rounding error.
The `Form` is what lets chapter 4 refuse `0.75` as an answer to `1/4 + 1/2`
and chapter 6 refuse `3/4` — in both cases the shape of the answer is the
thing being taught. Everywhere else `Form` is `FormWhole` and any equal value
passes.

### Chapters

Maths Sprint's levels are grouped into `maths.Chapters`, which follow the Art
of Problem Solving *Prealgebra* book, and flattened into `maths.Levels` for
progression. `GameStats.LessonsCleared` is an index into the flat list, so a
level may be appended to the end of a chapter but never inserted in front of
one already there: that would silently hand out, or take away, progress nobody
earned.

### Scoring shape

All the games score on the same curve, so a score -- and the points it
converts to -- means roughly the same thing wherever it came from: finishing is the floor, pace adds a **capped** bonus,
and accuracy pays only above 80% — guessing fast must never beat working it out.
Two games deviate deliberately. Spelling has no pace component at all, because
paying for speed would push toward answering before the word has been read. Code
breaker has no partial credit, because a code is cracked or it is not; its bonus
is for guesses left over.

### Second chances

A wrong answer does not end that question. It goes to the back of the queue
for one more go, after the correction screen has shown the right answer — the
difference between being told you got it wrong and being given a chance to use
what you were just told. Once only: a question missed on its second go needs a
person, not another go.

Maths Sprint, the Spelling Bee and Trace the Code all do this, each in its own
`round`, because each already owned one. Spelling reads the sentence out again
with the word, so the second attempt is the whole exercise; Trace the Code
leaves the program on screen next to what it prints, so the second attempt is
walking the same lines knowing where they end up.

Code Breaker does not, and should not. It has no queue of questions: there is
one code and eight to twelve guesses against it with feedback each time, so a
wrong guess is already the loop. A "second go" there would just be more
guesses, which `Level.Guesses` already sets. The typing trainer has no
discrete questions at all.

Both attempts count, so the accuracy denominator grows with the round. Twelve
questions with two missed and both recovered is 12 out of 14, not 12 out of 12
and not 10 out of 12 — a real second chance that still says it took two goes.
The consequence for code is that the length of the question slice is not fixed
for the life of a round; `round.firstPass` is.

## Rendering rules

This runs on a Linux virtual console, not a modern terminal emulator. Two
consequences:

- **ASCII only in anything rendered.** No emoji, no `✓`, no em dashes. The
  console font is not guaranteed to have them, and a missing glyph shows as a
  blank or a box. Status uses the four-character badges in `internal/ui`
  (`done` / `open` / `lock`), which also keeps columns aligned — a padlock emoji
  is two cells wide and a tick is one, which is what made the first version of
  the lesson list ragged.
- **Lesson content must be typeable.** Every character in a lesson line has to
  exist on the keyboard. There is a test for this; it exists because the first
  draft contained em dashes the player would have had no way to enter, and no
  way to skip.

The block characters in the progress bar are the one exception, and
`deploy/unlock-session.sh` sets a console font that has them.

## Points backends

`points.Backend` is `Award` plus `Balance`. Today the only implementation is
`points.Local`, which writes straight into the in-memory profile for the
launcher to persist.

`points.TaskBank` is the placeholder for the chore ledger. When it exists, fill
in the two methods and wire it up behind Local:

```go
Ledger: &points.Multi{Backends: []points.Backend{
    points.NewLocal(prof),
    &points.TaskBank{BaseURL: ..., Token: ..., PlayerID: ...},
}},
```

`Multi` treats the first backend as authoritative and the rest as best-effort
mirrors. That ordering matters: a TaskBank that is down, misconfigured, or
simply unreachable in the back of a car must never block play or lose an award.
Reconciliation is the open question — the award needs an idempotency key so a
retry over a flaky link cannot double-credit, and there needs to be a local
queue of awards that have not yet been mirrored.

## Updates

`unlock update` fetches a manifest, compares versions, and if there is something
newer for this GOOS/GOARCH, downloads it, verifies its SHA-256, and renames it
over the target. The manifest is generated from the real artifact by
`scripts/release.sh`, so the checksum cannot drift from the binary.

```json
{
  "version": "0.2.0",
  "notes": "adds the number game",
  "artifacts": {
    "linux/amd64": { "url": "https://...", "sha256": "...", "size": 7520418 }
  }
}
```

The checksum in the manifest is the only thing between the laptop and whatever
the network hands back, so the manifest must be served over HTTPS. It is not a
signature: anyone who can rewrite the manifest can rewrite the checksum with it.
If this ever leaves the house network, sign the manifest and pin a public key in
the binary.

Replacement is a rename onto the target path, which on Linux is atomic and safe
under a running process — the running image keeps the old inode until it exits.
Because the launcher is the login shell, the new build takes over the next time
the player quits and agetty logs them back in.

The launcher also checks in the background at startup and shows a one-line note
when something is available. Every failure path there is silent: an offline
laptop is the normal case, not an error worth interrupting anyone over.

## What is deliberately not built yet

- **The upper rungs do anything.** `unlocks.Ladder` currently only describes the
  stages; nothing provisions a shell, an editor or a desktop when one is
  reached. The declaration is separate from the effect precisely so the ladder
  can be tuned without touching any system integration.
- **Signed updates.** See above.
- **Anti-tamper.** The profile is hand-editable JSON and the account can reach
  it. This is a feature until it is a problem — and if a nine-year-old works out
  how to edit JSON to skip the ladder, that is arguably the ladder working.
