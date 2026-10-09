# laptop-unlocking-games

A laptop that opens up as it is earned.

It boots to a TTY running one Go binary — a typing trainer. Playing earns
points, points unlock rungs of a ladder, and the ladder ends at a desktop with
Steam on it. Everything above the first rung is hidden until it is reached.

A point is a minute of screen time, and a well-played round of any game pays
three to five of them, so an hour of screen time is about an hour of play.

## The games

| game | opens at | teaches |
| ---- | -------- | ------- |
| **Typing Trainer** | free | home row through full sentences, with live WPM and accuracy |
| **Maths Sprint** | 8 | times tables (weighted towards 6, 7, 11 and 12) and division, then the first nine chapters of AoPS Prealgebra: negatives and order of operations, exponents, number theory, fractions, equations, decimals, ratios, percents, square roots |
| **Spelling Bee** | 18 | a sentence appears with one word picked out, both vanish, and the word has to be written from memory |
| **Code Breaker** | 30 | deduction: crack a hidden code from exact/near feedback |
| **Trace the Code** | 38 | read a short program and say what it prints: variables, reassignment, loops, if |
| **Shell Quest** | 67 | `ls`, `cd`, `cat`, `pwd` and hidden files, in a pretend filesystem |

Shell Quest sits directly below the **shell** rung on purpose: by the time the
real prompt is handed over, the commands should already be familiar.

### Maths Sprint and the book

Maths Sprint follows the Art of Problem Solving *Prealgebra* book. The level
menu is two screens: a contents page of chapters, then the levels inside one.
Chapter numbers are the book's, so "do chapter 4 tonight" means the same thing
at the kitchen table and on the laptop. A warm-up chapter of number facts sits
in front of chapter 1, and chapters 10 to 15 -- angles, area, triangles,
statistics and counting -- are deliberately not built: they want figures drawn,
and this runs on a virtual console.

Answers are typed as whole numbers, fractions (`3/4`, in lowest terms) or
decimals (`0.375`). A level that is teaching fractions refuses the decimal of
the same value, and the other way round.

Trace the Code picks up where Maths Sprint's algebra leaves off. "Putting
Numbers In" asks what `3x + 2` is when x is 4; Trace the Code asks the same
thing with the working written down a line at a time, then goes where algebra
cannot follow — a box whose contents change (`x = x + 2` is false in algebra
and ordinary in code), a line that runs four times, a line that never runs.

## Quick start

```sh
make run          # build and play
make check        # fmt, vet, test
make dist         # cross-compile for the laptop (linux/amd64)
make release      # dist/ plus an update manifest
```

Go 1.26.5, pinned in `.tool-versions`.

## Commands

```
unlock            start the launcher
unlock version    print build identity
unlock update     check for a newer build; -apply installs it
unlock profile    show where progress is stored, and a summary
```

Useful flags and their environment equivalents:

| flag            | env                   | meaning                                               |
| --------------- | --------------------- | ----------------------------------------------------- |
| `-profile`      | `UNLOCK_PROFILE`      | where progress lives, default `$XDG_DATA_HOME/unlock`  |
| `-name`         | `UNLOCK_PLAYER`       | player name, used on a first run                       |
| `-manifest-url` | `UNLOCK_MANIFEST_URL` | release manifest; empty disables update checks         |

## Layout

```
cmd/unlock          the binary: flag parsing and wiring
internal/catalog    the one place that knows which games ship
internal/launcher   the root Bubble Tea model — menu, results, progress
internal/games      the Game contract and registry
internal/games/*    one package per game: typing, maths, spelling,
                    codebreaker, shellquest
internal/profile    persistent progress, one JSON file, written atomically
internal/unlocks    the ladder from typing trainer to Steam
internal/achievements   badge definitions and rules
internal/points     the points ledger, with TaskBank stubbed behind an interface
internal/update     manifest fetch, checksum verification, in-place replace
internal/ui         shared Lip Gloss styles
deploy/             systemd units and the install script for the Arch box
```

## On the laptop

Target is Arch on x86_64. From a checkout on the machine:

```sh
make dist
sudo UNLOCK_MANIFEST_URL=https://.../manifest.json ./deploy/install.sh
```

That creates an unprivileged `player` account, autologs it in on tty1, launches
the game as its login shell, and enables a daily update timer. tty2 is left
alone so there is still a way in.

See [docs/architecture.md](docs/architecture.md) for how the pieces fit and how
to add the next game.

## Where this is going

- TaskBank as the real points ledger, so chores and games feed the same number.
- The upper rungs — shell, editor, browser, desktop, Steam — provisioned by the
  unlock rather than just described by it.
