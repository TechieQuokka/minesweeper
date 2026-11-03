package ui

import (
	"math/rand/v2"
	"testing"

	"github.com/TechieQuokka/minesweeper/internal/game"
)

// testApp returns an app at scale 1 with a seeded game whose first click at
// the board center has already been made.
func testApp(t *testing.T) *app {
	t.Helper()
	a := &app{level: game.Medium}
	a.restart(false)
	a.game = game.New(game.Medium, game.WithRand(rand.New(rand.NewPCG(7, 7))))
	a.m = newMetrics(1, a.level.Width, a.level.Height)
	return a
}

func (a *app) cellCenter(x, y int) (int, int) {
	return a.m.boardX + x*a.m.cell + a.m.cell/2, a.m.boardY + y*a.m.cell + a.m.cell/2
}

// frame feeds one input frame, mimicking Update.
func (a *app) frame(in mouseInput) {
	a.cursorX, a.cursorY = in.x, in.y
	a.handleMouse(in)
	a.reportEnd()
}

func at(a *app, x, y int) mouseInput {
	px, py := a.cellCenter(x, y)
	return mouseInput{x: px, y: py}
}

func (a *app) leftClick(x, y int) {
	in := at(a, x, y)
	in.left, in.leftDown = true, true
	a.frame(in)
	in = at(a, x, y)
	in.leftUp = true
	a.frame(in)
}

func (a *app) rightClick(x, y int) {
	in := at(a, x, y)
	in.right, in.rightDown = true, true
	a.frame(in)
	in = at(a, x, y)
	in.rightUp = true
	a.frame(in)
}

func TestLeftClickOpensOnRelease(t *testing.T) {
	a := testApp(t)
	in := at(a, 8, 8)
	in.left, in.leftDown = true, true
	a.frame(in)
	if a.game.Cell(8, 8).State == game.Opened {
		t.Fatal("cell opened on press, want on release")
	}
	if a.faceKind() != faceSurprised {
		t.Error("face should be surprised while pressing on the board")
	}
	if p := a.pressedCells(); len(p) != 1 || !p[[2]int{8, 8}] {
		t.Errorf("pressed cells = %v, want only (8,8)", p)
	}
	in = at(a, 8, 8)
	in.leftUp = true
	a.frame(in)
	if a.game.Cell(8, 8).State != game.Opened {
		t.Fatal("cell not opened on release")
	}
}

func TestLeftReleaseOffBoardCancels(t *testing.T) {
	a := testApp(t)
	in := at(a, 8, 8)
	in.left, in.leftDown = true, true
	a.frame(in)
	a.frame(mouseInput{x: 1, y: 1, leftUp: true})
	if a.game.Status() != game.Ready {
		t.Fatal("release off board opened a cell")
	}
}

func TestPressMovesWithCursor(t *testing.T) {
	a := testApp(t)
	in := at(a, 1, 1)
	in.left, in.leftDown = true, true
	a.frame(in)
	in = at(a, 3, 4)
	in.left = true
	a.frame(in)
	if p := a.pressedCells(); !p[[2]int{3, 4}] || p[[2]int{1, 1}] {
		t.Errorf("pressed cells = %v, want (3,4)", p)
	}
	in = at(a, 3, 4)
	in.leftUp = true
	a.frame(in)
	if a.game.Cell(3, 4).State != game.Opened {
		t.Error("release should open the cell under the cursor")
	}
}

func TestRightClickTogglesOnPress(t *testing.T) {
	a := testApp(t)
	in := at(a, 2, 2)
	in.right, in.rightDown = true, true
	a.frame(in)
	if s := a.game.Cell(2, 2).State; s != game.Flagged {
		t.Fatalf("state after right press = %v, want Flagged", s)
	}
	in = at(a, 2, 2)
	in.rightUp = true
	a.frame(in)
	if s := a.game.Cell(2, 2).State; s != game.Flagged {
		t.Fatalf("state after right release = %v, want Flagged", s)
	}
	a.rightClick(2, 2)
	a.rightClick(2, 2)
	if s := a.game.Cell(2, 2).State; s != game.Hidden {
		t.Fatalf("state after full cycle = %v, want Hidden", s)
	}
}

func TestBothButtonsChordAndSuppressLeftRelease(t *testing.T) {
	a := testApp(t)
	in := at(a, 5, 5)
	in.left, in.leftDown = true, true
	a.frame(in)
	in = at(a, 5, 5)
	in.left, in.right, in.rightDown = true, true, true
	a.frame(in)
	if !a.chording {
		t.Fatal("left then right should start a chord")
	}
	if s := a.game.Cell(5, 5).State; s != game.Hidden {
		t.Fatalf("right press during left hold toggled mark: %v", s)
	}
	if p := a.pressedCells(); len(p) != 9 {
		t.Errorf("chord pressed %d cells, want 9", len(p))
	}
	// Release right first: chord fires (no-op on a hidden cell).
	in = at(a, 5, 5)
	in.left, in.rightUp = true, true
	a.frame(in)
	// Then left: must not open the cell.
	in = at(a, 5, 5)
	in.leftUp = true
	a.frame(in)
	if a.game.Status() != game.Ready {
		t.Fatal("left release after chord opened a cell")
	}
	// Input works normally afterwards.
	a.leftClick(5, 5)
	if a.game.Status() == game.Ready {
		t.Fatal("click after chord was ignored")
	}
}

// findNumber returns an opened number cell that still has hidden neighbors.
func findNumber(t *testing.T, g *game.Game) (int, int) {
	t.Helper()
	lv := g.Level()
	for y := range lv.Height {
		for x := range lv.Width {
			c := g.Cell(x, y)
			if c.State != game.Opened || c.Adjacent == 0 {
				continue
			}
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					if g.InBounds(x+dx, y+dy) && g.Cell(x+dx, y+dy).State == game.Hidden && !g.Cell(x+dx, y+dy).Mine {
						return x, y
					}
				}
			}
		}
	}
	t.Fatal("no suitable number cell")
	return 0, 0
}

func TestMiddleClickChords(t *testing.T) {
	a := testApp(t)
	// Open cells until the board has a frontier.
	a.leftClick(8, 8)
	for i := 0; a.game.Status() == game.Playing && i < 5; i++ {
		if c := a.game.Cell(i, 0); !c.Mine {
			a.leftClick(i, 0)
		}
	}
	g := a.game
	x, y := findNumber(t, g)
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if g.InBounds(x+dx, y+dy) && g.Cell(x+dx, y+dy).Mine && g.Cell(x+dx, y+dy).State == game.Hidden {
				a.rightClick(x+dx, y+dy)
			}
		}
	}
	in := at(a, x, y)
	in.middle, in.midDn = true, true
	a.frame(in)
	in = at(a, x, y)
	in.midUp = true
	a.frame(in)
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			nx, ny := x+dx, y+dy
			if g.InBounds(nx, ny) && !g.Cell(nx, ny).Mine && g.Cell(nx, ny).State != game.Opened {
				t.Errorf("(%d,%d) not opened by chord", nx, ny)
			}
		}
	}
}

func faceInput(a *app) mouseInput {
	fx, fy := a.faceCenter()
	return mouseInput{x: int(fx), y: int(fy)}
}

func TestFaceRestart(t *testing.T) {
	a := testApp(t)
	a.leftClick(8, 8)
	old := a.game

	in := faceInput(a)
	in.left, in.leftDown = true, true
	a.frame(in)
	if !a.facePressed || a.faceKind() != faceNormal {
		t.Fatal("face should be pressed and not surprised")
	}
	in = faceInput(a)
	in.leftUp = true
	a.frame(in)
	if a.game == old || a.game.Status() != game.Ready {
		t.Fatal("face click did not restart")
	}
}

func TestFaceReleaseElsewhereCancels(t *testing.T) {
	a := testApp(t)
	old := a.game
	in := faceInput(a)
	in.left, in.leftDown = true, true
	a.frame(in)
	out := at(a, 0, 0)
	out.leftUp = true
	a.frame(out)
	if a.game != old {
		t.Fatal("releasing away from the face restarted")
	}
	if a.game.Status() != game.Ready {
		t.Fatal("face press leaked into the board")
	}
}

func TestReportEndOnce(t *testing.T) {
	a := testApp(t)
	calls := 0
	var gotWon bool
	a.onEnd = func(won bool, _ int) { calls++; gotWon = won }
	a.leftClick(8, 8)
	lv := a.game.Level()
	for y := range lv.Height {
		for x := range lv.Width {
			if a.game.Cell(x, y).Mine && !a.game.Over() {
				a.leftClick(x, y)
			}
		}
	}
	a.leftClick(0, 0)
	a.frame(mouseInput{})
	if calls != 1 || gotWon {
		t.Fatalf("onEnd calls = %d won = %v, want 1 call with loss", calls, gotWon)
	}
	if a.faceKind() != faceDead {
		t.Error("face should be dead after loss")
	}
	if p := a.pressedCells(); p != nil {
		t.Error("no cells should be pressed after game over")
	}
}

func TestTileFor(t *testing.T) {
	tests := []struct {
		c        game.Cell
		lost     bool
		exploded bool
		want     tile
	}{
		{game.Cell{State: game.Hidden}, false, false, tileHidden},
		{game.Cell{State: game.Question}, false, false, tileQuestion},
		{game.Cell{State: game.Flagged}, false, false, tileFlag},
		{game.Cell{State: game.Opened}, false, false, tileOpened},
		{game.Cell{State: game.Opened, Adjacent: 3}, false, false, tileNumber1 + 2},
		{game.Cell{Mine: true, State: game.Hidden}, false, false, tileHidden},
		{game.Cell{Mine: true, State: game.Hidden}, true, false, tileMine},
		{game.Cell{Mine: true, State: game.Question}, true, false, tileMine},
		{game.Cell{Mine: true, State: game.Flagged}, true, false, tileFlag},
		{game.Cell{State: game.Flagged}, true, false, tileWrongFlag},
		{game.Cell{Mine: true, State: game.Opened}, true, true, tileMineExploded},
	}
	for i, tt := range tests {
		if got := tileFor(tt.c, tt.lost, tt.exploded); got != tt.want {
			t.Errorf("case %d: tileFor(%+v) = %v, want %v", i, tt.c, got, tt.want)
		}
	}
}

func TestFormatCounter(t *testing.T) {
	for n, want := range map[int]string{0: "000", 7: "007", 40: "040", 999: "999", 1500: "999", -2: "-02", -150: "-99"} {
		if got := formatCounter(n); got != want {
			t.Errorf("formatCounter(%d) = %q, want %q", n, got, want)
		}
	}
}
