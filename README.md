# laptop-unlocking-games

A laptop that opens up as it is earned.

It boots to a TTY running one Go binary — a typing trainer. Playing earns
points, points unlock rungs of a ladder, and the ladder ends at a desktop with
Steam on it. Everything above the first rung is hidden until it is reached.

## The games

| game | opens at | teaches |
| ---- | -------- | ------- |
| **Typing Trainer** | free | home row through full sentences, with live WPM and accuracy |
| **Maths Sprint** | 250 | addition, subtraction, times tables and division, then algebra: solving for x, substitution, collecting and factorising |
| **Spelling Bee** | 550 | a word appears, vanishes, and has to be written from memory |
| **Code Breaker** | 900 | deduction: crack a hidden code from exact/near feedback |
| **Shell Quest** | 2000 | `ls`, `cd`, `cat`, `pwd` and hidden files, in a pretend filesystem |

Shell Quest sits directly below the **shell** rung on purpose: by the time the
real prompt is handed over, the commands should already be familiar.

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
