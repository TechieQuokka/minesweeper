package main

import (
	"fmt"
	"io"

	"github.com/TechieQuokka/minesweeper/internal/game"
)

// version is overridable at build time: -ldflags "-X main.version=v1.2.3".
var version = "v1.0.0"

const usage = `usage: mine [easy|medium|hard]

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
`

type action int

const (
	actionPlay action = iota
	actionHelp
	actionVersion
	actionError
)

type command struct {
	action action
	level  game.Level
	err    string
}

// parseArgs interprets the command-line arguments (without the program name).
func parseArgs(args []string) command {
	if len(args) == 0 {
		return command{action: actionPlay, level: game.DefaultLevel}
	}
	if len(args) > 1 {
		return command{action: actionError, err: "too many arguments"}
	}
	switch arg := args[0]; arg {
	case "help":
		return command{action: actionHelp}
	case "--version", "-V":
		return command{action: actionVersion}
	default:
		if lv, ok := game.LevelByName(arg); ok {
			return command{action: actionPlay, level: lv}
		}
		return command{action: actionError, err: fmt.Sprintf("unknown argument: %q", arg)}
	}
}

// handleNonPlay prints help, version, or errors. It returns the exit code and
// whether the program should exit instead of starting a game.
func handleNonPlay(cmd command, stdout, stderr io.Writer) (code int, exit bool) {
	switch cmd.action {
	case actionHelp:
		fmt.Fprint(stdout, usage)
		return 0, true
	case actionVersion:
		fmt.Fprintf(stdout, "mine %s\n", version)
		return 0, true
	case actionError:
		fmt.Fprintf(stderr, "%s\nrun 'mine help' for usage\n", cmd.err)
		return 2, true
	}
	return 0, false
}

// resultLine formats the line printed to the terminal when a game ends.
func resultLine(won bool, level string, seconds int) string {
	if won {
		return fmt.Sprintf("cleared %s in %ds", level, seconds)
	}
	return fmt.Sprintf("boom! %s, %ds", level, seconds)
}
