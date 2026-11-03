// Package game implements minesweeper rules without any rendering or input
// dependencies, so it can be unit-tested in isolation.
package game

// Level describes a board configuration.
type Level struct {
	Name   string
	Width  int
	Height int
	Mines  int
}

// Predefined levels, matching the classic beginner/intermediate/expert boards.
var (
	Easy   = Level{Name: "easy", Width: 9, Height: 9, Mines: 10}
	Medium = Level{Name: "medium", Width: 16, Height: 16, Mines: 40}
	Hard   = Level{Name: "hard", Width: 30, Height: 16, Mines: 99}
)

// Levels lists all predefined levels in display order.
var Levels = []Level{Easy, Medium, Hard}

// DefaultLevel is used when no level is given on the command line.
var DefaultLevel = Medium

// LevelByName returns the predefined level with the given name.
func LevelByName(name string) (Level, bool) {
	for _, l := range Levels {
		if l.Name == name {
			return l, true
		}
	}
	return Level{}, false
}
