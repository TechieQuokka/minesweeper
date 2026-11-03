// Command mine is a minesweeper game.
package main

import (
	"fmt"
	"os"

	"github.com/TechieQuokka/minesweeper/internal/ui"
)

func main() {
	cmd := parseArgs(os.Args[1:])
	if code, exit := handleNonPlay(cmd, os.Stdout, os.Stderr); exit {
		os.Exit(code)
	}

	onEnd := func(won bool, seconds int) {
		fmt.Println(resultLine(won, cmd.level.Name, seconds))
	}
	if err := ui.Run(cmd.level, onEnd); err != nil {
		fmt.Fprintf(os.Stderr, "mine: %v\n", err)
		os.Exit(1)
	}
}
