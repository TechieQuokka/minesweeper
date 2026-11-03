package game

import (
	"math/rand/v2"
	"time"
)

// CellState is the player-visible state of a cell.
type CellState int

const (
	Hidden CellState = iota
	Flagged
	Question
	Opened
)

// Cell is a single board square.
type Cell struct {
	Mine     bool
	Adjacent int // number of mines in the 8 neighbors
	State    CellState
}

// Status is the overall game state.
type Status int

const (
	Ready   Status = iota // no cell opened yet; mines not placed
	Playing               // at least one cell opened
	Won
	Lost
)

// MaxSeconds is the value at which the timer stops counting.
const MaxSeconds = 999

// Game holds the state of a single round.
type Game struct {
	level    Level
	cells    []Cell
	status   Status
	rng      *rand.Rand
	now      func() time.Time
	start    time.Time
	end      time.Time
	opened   int
	flags    int
	exploded int // index of the mine that ended the game, or -1
	placed   bool
}

// Option customizes a Game.
type Option func(*Game)

// WithRand sets the random source used for mine placement.
func WithRand(r *rand.Rand) Option { return func(g *Game) { g.rng = r } }

// WithClock sets the clock used by the timer.
func WithClock(now func() time.Time) Option { return func(g *Game) { g.now = now } }

// New creates a game for the given level. Mines are placed on the first open.
func New(level Level, opts ...Option) *Game {
	g := &Game{
		level:    level,
		cells:    make([]Cell, level.Width*level.Height),
		rng:      rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())),
		now:      time.Now,
		exploded: -1,
	}
	for _, o := range opts {
		o(g)
	}
	return g
}

// Level returns the game's level.
func (g *Game) Level() Level { return g.level }

// Status returns the current game status.
func (g *Game) Status() Status { return g.status }

// Over reports whether the game has been won or lost.
func (g *Game) Over() bool { return g.status == Won || g.status == Lost }

// InBounds reports whether (x, y) is on the board.
func (g *Game) InBounds(x, y int) bool {
	return x >= 0 && y >= 0 && x < g.level.Width && y < g.level.Height
}

// Cell returns the cell at (x, y). It panics if out of bounds.
func (g *Game) Cell(x, y int) Cell { return g.cells[g.index(x, y)] }

// Exploded returns the coordinates of the mine that ended the game.
func (g *Game) Exploded() (x, y int, ok bool) {
	if g.exploded < 0 {
		return 0, 0, false
	}
	return g.exploded % g.level.Width, g.exploded / g.level.Width, true
}

// MinesLeft returns total mines minus placed flags. It may be negative.
func (g *Game) MinesLeft() int { return g.level.Mines - g.flags }

// Seconds returns the timer value: 0 before the first open, then 1 at the
// first open and +1 every second, capped at MaxSeconds and frozen on game end.
func (g *Game) Seconds() int {
	if g.status == Ready {
		return 0
	}
	t := g.now()
	if g.Over() {
		t = g.end
	}
	s := int(t.Sub(g.start)/time.Second) + 1
	return min(s, MaxSeconds)
}

// Open opens the cell at (x, y). Flagged and already opened cells are ignored.
func (g *Game) Open(x, y int) {
	if g.Over() || !g.InBounds(x, y) {
		return
	}
	i := g.index(x, y)
	if s := g.cells[i].State; s == Flagged || s == Opened {
		return
	}
	if g.status == Ready {
		if !g.placed {
			g.placeMines(i)
		}
		g.status = Playing
		g.start = g.now()
	}
	g.open(i)
	g.checkWin()
}

// ToggleMark cycles a non-opened cell through hidden -> flagged -> question -> hidden.
func (g *Game) ToggleMark(x, y int) {
	if g.Over() || !g.InBounds(x, y) {
		return
	}
	c := &g.cells[g.index(x, y)]
	switch c.State {
	case Hidden:
		c.State = Flagged
		g.flags++
	case Flagged:
		c.State = Question
		g.flags--
	case Question:
		c.State = Hidden
	}
}

// Chord opens all non-flagged neighbors of an opened number cell when the
// number of adjacent flags equals its number.
func (g *Game) Chord(x, y int) {
	if g.status != Playing || !g.InBounds(x, y) {
		return
	}
	c := g.cells[g.index(x, y)]
	if c.State != Opened || c.Adjacent == 0 {
		return
	}
	flags := 0
	g.eachNeighbor(x, y, func(n int) {
		if g.cells[n].State == Flagged {
			flags++
		}
	})
	if flags != c.Adjacent {
		return
	}
	g.eachNeighbor(x, y, func(n int) {
		if g.status != Playing {
			return
		}
		if s := g.cells[n].State; s == Hidden || s == Question {
			g.open(n)
		}
	})
	g.checkWin()
}

func (g *Game) index(x, y int) int { return y*g.level.Width + x }

func (g *Game) eachNeighbor(x, y int, fn func(i int)) {
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			nx, ny := x+dx, y+dy
			if (dx != 0 || dy != 0) && g.InBounds(nx, ny) {
				fn(g.index(nx, ny))
			}
		}
	}
}

// placeMines distributes mines uniformly over all cells except safe.
func (g *Game) placeMines(safe int) {
	n := len(g.cells)
	candidates := make([]int, 0, n-1)
	for i := range n {
		if i != safe {
			candidates = append(candidates, i)
		}
	}
	mines := min(g.level.Mines, len(candidates))
	// Partial Fisher-Yates shuffle: the first `mines` entries are the picks.
	for k := range mines {
		j := k + g.rng.IntN(len(candidates)-k)
		candidates[k], candidates[j] = candidates[j], candidates[k]
		g.cells[candidates[k]].Mine = true
	}
	g.countAdjacent()
}

// countAdjacent fills in Adjacent for every cell and marks mines as placed.
func (g *Game) countAdjacent() {
	g.placed = true
	for i := range g.cells {
		x, y := i%g.level.Width, i/g.level.Width
		g.eachNeighbor(x, y, func(n int) {
			if g.cells[n].Mine {
				g.cells[i].Adjacent++
			}
		})
	}
}

// open opens cell i and flood-fills from zero cells.
func (g *Game) open(i int) {
	if g.cells[i].Mine {
		g.cells[i].State = Opened
		g.lose(i)
		return
	}
	stack := []int{i}
	for len(stack) > 0 {
		cur := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		c := &g.cells[cur]
		if c.State == Opened || c.State == Flagged {
			continue
		}
		c.State = Opened
		g.opened++
		if c.Adjacent != 0 {
			continue
		}
		g.eachNeighbor(cur%g.level.Width, cur/g.level.Width, func(n int) {
			if s := g.cells[n].State; s == Hidden || s == Question {
				stack = append(stack, n)
			}
		})
	}
}

func (g *Game) lose(i int) {
	g.status = Lost
	g.exploded = i
	g.end = g.now()
}

func (g *Game) checkWin() {
	if g.status != Playing || g.opened != len(g.cells)-g.level.Mines {
		return
	}
	g.status = Won
	g.end = g.now()
	for i := range g.cells {
		if g.cells[i].Mine && g.cells[i].State != Flagged {
			g.cells[i].State = Flagged
			g.flags++
		}
	}
}
