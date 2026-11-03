# mine — Specification

A minesweeper game for Ubuntu, written in Go with [Ebitengine](https://ebitengine.org/).
Gameplay matches the classic Windows XP minesweeper rules; the visuals are an
original, modern, flat, dark-only design (no assets or look copied from the original).

## 1. CLI

Binary name: `mine`

```
mine                 start a medium game (default)
mine easy            9x9,   10 mines
mine medium          16x16, 40 mines
mine hard            30x16, 99 mines
mine help            print usage and exit
mine --version       print version and exit
mine -V              same as --version
```

- At most one argument is accepted.
- Any other argument (including `-h`, `--help`) is an error:
  ```
  unknown argument: "-h"
  run 'mine help' for usage
  ```
  Exit code `2`. More than one argument prints `too many arguments` with the
  same hint and exit code.
- Version string defaults to `v1.0.0`, overridable at build time with
  `-ldflags "-X main.version=..."`. Output format: `mine v1.0.0`.
- Difficulty can only be changed by restarting the program.

### Usage text (`mine help`)

```
usage: mine [easy|medium|hard]

levels:
  easy     9x9,   10 mines
  medium   16x16, 40 mines  (default)
  hard     30x16, 99 mines

other:
  mine help        show this help
  mine --version   show version (-V)

controls:
  left click       open cell
  right click      cycle flag -> ? -> blank
  both / middle    chord (open neighbors of a satisfied number)
  F2 or face       restart
```

### Result line

When a game ends (win or loss), one line is printed to stdout:

```
cleared hard in 142s
boom! hard, 87s
```

## 2. Gameplay rules (classic XP behavior)

### Board
| Level  | Width | Height | Mines |
|--------|-------|--------|-------|
| easy   | 9     | 9      | 10    |
| medium | 16    | 16     | 40    |
| hard   | 30    | 16     | 99    |

### First-click protection
- Mines are placed lazily on the first open action.
- Only the clicked cell is guaranteed safe (its neighbors may contain mines).

### Cell states
`hidden`, `flagged`, `question`, `opened`.

### Controls
- **Left click** on a hidden or question cell: open it. Flagged cells ignore left click.
- **Right click** on a non-opened cell: cycle `hidden -> flagged -> question -> hidden`.
  Question marks are always enabled.
- **Chord** (left+right together, or middle click) on an opened number cell:
  if the number of adjacent flags equals the number, open every adjacent
  hidden/question cell. Otherwise nothing happens.
- **Press feedback**: while a mouse button is held, the target cell (or the 3x3
  area for a chord) is drawn pressed. The action happens on release. Moving
  the cursor while held moves the pressed highlight. Releasing outside the board
  cancels.
- **F2** or clicking the face button: restart with the same level.
- Input on the board is ignored after the game ends (only restart works).

### Opening
- Opening a mine: loss.
- Opening a cell with 0 adjacent mines flood-opens its neighbors recursively.
  Flagged cells are not opened by flood fill; question cells are.

### Win
- Win when every non-mine cell is opened.
- On win, all remaining mines are auto-flagged and the mine counter shows 0.

### Loss
- All mines are revealed.
- The mine that was clicked is highlighted.
- Flags on non-mine cells are shown as wrong (crossed).
- Correct flags stay as flags.

## 3. Top bar

```
[ mines left ]     [ face ]     [ timer ]
```

- **Mines left**: total mines − flags placed. Can go negative. 3 digits
  (e.g. `040`, `-02`).
- **Face button**: restart. Expressions:
  - playing: normal
  - mouse button held on board: surprised
  - won: cool
  - lost: dead
  - face itself pressed: drawn pressed
- **Timer**: seconds since first open, starts at 1 on the first click
  (XP behavior), stops at 999, stops on win/loss. 3 digits.

## 4. Visual design

- Dark theme only. No light mode, no theme toggle.
- Modern flat style: no 3D bevels, no LED-style digits, no copied sprites.
- All graphics drawn in code (vector shapes / Ebitengine text); no external image assets.
- No animations.
- Number colors: distinct, readable on dark background, one color per 1–8.
- Window size fits the board; scaled for HiDPI via Ebitengine device scale factor.
- Window title: `mine — <level>`.

## 5. Architecture

```
cmd/mine/        main: CLI parsing, wiring, result line
internal/game/   pure game logic (no Ebitengine imports), fully unit-tested
internal/ui/     Ebitengine rendering + input, depends on internal/game
```

- `internal/game` is deterministic given a random source (injectable for tests).
- No persistence: nothing is written to disk.
- Dependencies: Go standard library + Ebitengine only.

## 6. Out of scope

Custom board sizes, high scores, light mode, animations, sound,
in-game menus, difficulty change without restart.

## 7. Build requirements (Ubuntu)

```
sudo apt install gcc libc6-dev libgl1-mesa-dev libxcursor-dev libxi-dev \
  libxinerama-dev libxrandr-dev libxxf86vm-dev libasound2-dev pkg-config
```
