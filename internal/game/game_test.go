package game

import (
	"math/rand/v2"
	"testing"
	"time"
)

// layout builds a game from an ASCII board where '*' is a mine.
func layout(t *testing.T, rows ...string) *Game {
	t.Helper()
	lv := Level{Name: "test", Width: len(rows[0]), Height: len(rows)}
	for _, r := range rows {
		for _, ch := range r {
			if ch == '*' {
				lv.Mines++
			}
		}
	}
	g := New(lv)
	for y, r := range rows {
		for x, ch := range r {
			g.cells[g.index(x, y)].Mine = ch == '*'
		}
	}
	g.countAdjacent()
	return g
}

type fakeClock struct{ t time.Time }

func (c *fakeClock) now() time.Time          { return c.t }
func (c *fakeClock) advance(d time.Duration) { c.t = c.t.Add(d) }

func TestLevelByName(t *testing.T) {
	for _, name := range []string{"easy", "medium", "hard"} {
		if l, ok := LevelByName(name); !ok || l.Name != name {
			t.Errorf("LevelByName(%q) = %v, %v", name, l, ok)
		}
	}
	if _, ok := LevelByName("extreme"); ok {
		t.Error("LevelByName(extreme) should fail")
	}
	if DefaultLevel != Medium {
		t.Error("default level should be medium")
	}
}

func TestFirstClickIsSafeAndMineCountExact(t *testing.T) {
	for seed := range uint64(200) {
		for _, lv := range Levels {
			g := New(lv, WithRand(rand.New(rand.NewPCG(seed, seed))))
			x, y := int(seed)%lv.Width, int(seed/3)%lv.Height
			g.Open(x, y)
			if g.Cell(x, y).Mine {
				t.Fatalf("%s seed %d: first click hit a mine", lv.Name, seed)
			}
			if g.Status() == Lost {
				t.Fatalf("%s seed %d: lost on first click", lv.Name, seed)
			}
			mines := 0
			for i := range g.cells {
				if g.cells[i].Mine {
					mines++
				}
			}
			if mines != lv.Mines {
				t.Fatalf("%s seed %d: %d mines, want %d", lv.Name, seed, mines, lv.Mines)
			}
		}
	}
}

func TestFirstClickOnlyProtectsClickedCell(t *testing.T) {
	// With classic rules a neighbor of the first click may hold a mine.
	// On a nearly full board this must happen.
	lv := Level{Name: "dense", Width: 3, Height: 3, Mines: 8}
	g := New(lv)
	g.Open(1, 1)
	if g.Cell(1, 1).Mine {
		t.Fatal("clicked cell has a mine")
	}
	if g.Cell(1, 1).Adjacent != 8 {
		t.Fatalf("adjacent = %d, want 8", g.Cell(1, 1).Adjacent)
	}
	if g.Status() != Won {
		t.Fatalf("status = %v, want Won (only safe cell opened)", g.Status())
	}
}

func TestAdjacentCounts(t *testing.T) {
	g := layout(t,
		"*..",
		"...",
		"..*",
	)
	want := [][]int{
		{0, 1, 0},
		{1, 2, 1},
		{0, 1, 0},
	}
	for y := range 3 {
		for x := range 3 {
			c := g.Cell(x, y)
			if !c.Mine && c.Adjacent != want[y][x] {
				t.Errorf("(%d,%d) adjacent = %d, want %d", x, y, c.Adjacent, want[y][x])
			}
		}
	}
}

func TestFloodFill(t *testing.T) {
	g := layout(t,
		".....",
		".....",
		"...11",
		"...1*",
	)
	g.Open(0, 0)
	for y := range 4 {
		for x := range 5 {
			c := g.Cell(x, y)
			if c.Mine {
				if c.State == Opened {
					t.Error("mine opened by flood fill")
				}
				continue
			}
			if c.State != Opened {
				t.Errorf("(%d,%d) not opened", x, y)
			}
		}
	}
	if g.Status() != Won {
		t.Errorf("status = %v, want Won", g.Status())
	}
}

func TestFloodFillSkipsFlagsButOpensQuestions(t *testing.T) {
	g := layout(t,
		"....*",
		".....",
		".....",
	)
	g.ToggleMark(0, 2) // flag
	g.ToggleMark(1, 2) // flag
	g.ToggleMark(1, 2) // question
	g.Open(0, 0)
	if s := g.Cell(0, 2).State; s != Flagged {
		t.Errorf("flagged cell state = %v, want Flagged", s)
	}
	if s := g.Cell(1, 2).State; s != Opened {
		t.Errorf("question cell state = %v, want Opened", s)
	}
}

func TestToggleMarkCycle(t *testing.T) {
	g := layout(t, "*..", "...")
	want := []CellState{Flagged, Question, Hidden, Flagged}
	wantLeft := []int{0, 1, 1, 0}
	for i, w := range want {
		g.ToggleMark(2, 1)
		if s := g.Cell(2, 1).State; s != w {
			t.Fatalf("step %d: state = %v, want %v", i, s, w)
		}
		if l := g.MinesLeft(); l != wantLeft[i] {
			t.Fatalf("step %d: mines left = %d, want %d", i, l, wantLeft[i])
		}
	}
}

func TestMinesLeftCanGoNegative(t *testing.T) {
	g := layout(t, "*...")
	g.ToggleMark(1, 0)
	g.ToggleMark(2, 0)
	if l := g.MinesLeft(); l != -1 {
		t.Errorf("mines left = %d, want -1", l)
	}
}

func TestOpenIgnoresFlaggedAndOpenedCells(t *testing.T) {
	g := layout(t, "*..", "...")
	g.ToggleMark(0, 0)
	g.Open(0, 0)
	if g.Status() != Ready {
		t.Fatalf("status = %v, want Ready (flag blocks open)", g.Status())
	}
	g.Open(2, 1)
	g.ToggleMark(2, 1)
	if s := g.Cell(2, 1).State; s != Opened {
		t.Errorf("marking opened cell changed state to %v", s)
	}
}

func TestLoss(t *testing.T) {
	g := layout(t,
		"*.*",
		"...",
		"...",
	)
	g.Open(1, 2)
	g.ToggleMark(1, 0) // wrong flag
	g.Open(0, 0)
	if g.Status() != Lost {
		t.Fatalf("status = %v, want Lost", g.Status())
	}
	if x, y, ok := g.Exploded(); !ok || x != 0 || y != 0 {
		t.Errorf("exploded = %d,%d,%v want 0,0,true", x, y, ok)
	}
	// Further input is ignored.
	g.ToggleMark(2, 0)
	if s := g.Cell(2, 0).State; s != Hidden {
		t.Errorf("input after loss changed state to %v", s)
	}
}

func TestWinAutoFlags(t *testing.T) {
	g := layout(t,
		"*.",
		"..",
	)
	g.ToggleMark(0, 0)
	g.ToggleMark(0, 0) // question on the mine
	g.Open(1, 0)
	g.Open(0, 1)
	g.Open(1, 1)
	if g.Status() != Won {
		t.Fatalf("status = %v, want Won", g.Status())
	}
	if s := g.Cell(0, 0).State; s != Flagged {
		t.Errorf("mine state after win = %v, want Flagged", s)
	}
	if l := g.MinesLeft(); l != 0 {
		t.Errorf("mines left after win = %d, want 0", l)
	}
}

func TestChord(t *testing.T) {
	g := layout(t,
		"*...",
		"....",
		"..**",
	)
	g.Open(1, 0) // shows 1
	g.Chord(1, 0)
	if s := g.Cell(2, 0).State; s != Hidden {
		t.Fatal("chord without enough flags opened cells")
	}
	g.ToggleMark(0, 0)
	g.ToggleMark(2, 0)
	g.ToggleMark(2, 0) // question: chord should open it
	g.Chord(1, 0)
	for _, p := range [][2]int{{2, 0}, {0, 1}, {1, 1}, {2, 1}} {
		if s := g.Cell(p[0], p[1]).State; s != Opened {
			t.Errorf("(%d,%d) state = %v, want Opened", p[0], p[1], s)
		}
	}
	if g.Status() != Playing {
		t.Errorf("status = %v, want Playing", g.Status())
	}
}

func TestChordWithWrongFlagLoses(t *testing.T) {
	g := layout(t,
		"*..",
		"...",
	)
	g.Open(1, 1) // shows 1
	g.ToggleMark(2, 0)
	g.Chord(1, 1)
	if g.Status() != Lost {
		t.Fatalf("status = %v, want Lost", g.Status())
	}
	if x, y, _ := g.Exploded(); x != 0 || y != 0 {
		t.Errorf("exploded at %d,%d want 0,0", x, y)
	}
}

func TestChordIgnoresHiddenAndZeroCells(t *testing.T) {
	g := layout(t, "*...", "....")
	g.Chord(2, 0) // Ready: no-op
	if g.Status() != Ready {
		t.Fatal("chord before first open changed status")
	}
}

func TestTimer(t *testing.T) {
	clk := &fakeClock{t: time.Unix(1000, 0)}
	g := New(Level{Name: "t", Width: 4, Height: 4, Mines: 1})
	g.now = clk.now
	// place the mine deterministically
	g.cells[15].Mine = true
	g.countAdjacent()

	if s := g.Seconds(); s != 0 {
		t.Fatalf("before start = %d, want 0", s)
	}
	clk.advance(5 * time.Second)
	if s := g.Seconds(); s != 0 {
		t.Fatalf("idle before start = %d, want 0", s)
	}
	g.Open(3, 2) // shows a number, game continues
	if s := g.Seconds(); s != 1 {
		t.Fatalf("at start = %d, want 1", s)
	}
	clk.advance(2500 * time.Millisecond)
	if s := g.Seconds(); s != 3 {
		t.Fatalf("after 2.5s = %d, want 3", s)
	}
	g.Open(3, 3) // mine
	clk.advance(10 * time.Second)
	if s := g.Seconds(); s != 3 {
		t.Fatalf("after loss = %d, want frozen 3", s)
	}
}

func TestTimerCap(t *testing.T) {
	clk := &fakeClock{t: time.Unix(0, 0)}
	g := layout(t, "*..")
	g.now = clk.now
	g.Open(1, 0) // shows 1, game continues
	clk.advance(5000 * time.Second)
	if s := g.Seconds(); s != MaxSeconds {
		t.Errorf("seconds = %d, want %d", s, MaxSeconds)
	}
}

func TestWithClockOption(t *testing.T) {
	clk := &fakeClock{t: time.Unix(0, 0)}
	g := New(Easy, WithClock(clk.now))
	g.Open(0, 0)
	clk.advance(4 * time.Second)
	if g.Status() == Playing && g.Seconds() != 5 {
		t.Errorf("seconds = %d, want 5", g.Seconds())
	}
}
