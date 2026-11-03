package ui

import (
	"image"
	"image/png"
	"math/rand/v2"
	"os"
	"path/filepath"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/TechieQuokka/minesweeper/internal/game"
)

// TestSnapshots renders representative game states to PNG files for visual
// review. It opens a window, so it only runs when MINE_SNAPSHOT_DIR is set:
//
//	MINE_SNAPSHOT_DIR=/some/dir go test ./internal/ui -run Snapshots
func TestSnapshots(t *testing.T) {
	dir := os.Getenv("MINE_SNAPSHOT_DIR")
	if dir == "" {
		t.Skip("MINE_SNAPSHOT_DIR not set")
	}

	type scene struct {
		name  string
		setup func(a *app)
	}
	seeded := func(lv game.Level) *game.Game {
		return game.New(lv, game.WithRand(rand.New(rand.NewPCG(7, 7))))
	}
	scenes := []scene{
		{"fresh-medium", func(a *app) {}},
		{"playing-medium", func(a *app) {
			g := seeded(game.Medium)
			g.Open(8, 8)
			marks := 0
			for y := range 16 {
				for x := range 16 {
					c := g.Cell(x, y)
					if c.Mine && c.State == game.Hidden && marks < 6 {
						g.ToggleMark(x, y)
						if marks%3 == 2 {
							g.ToggleMark(x, y) // question
						}
						marks++
					}
				}
			}
			a.game = g
		}},
		{"lost-medium", func(a *app) {
			g := seeded(game.Medium)
			g.Open(8, 8)
			flagged := 0
			for y := range 16 {
				for x := range 16 {
					c := g.Cell(x, y)
					if c.State == game.Hidden && flagged < 4 {
						g.ToggleMark(x, y) // some right, some wrong
						flagged++
					}
				}
			}
			for y := range 16 {
				for x := range 16 {
					if c := g.Cell(x, y); c.Mine && c.State == game.Hidden && !g.Over() {
						g.Open(x, y)
					}
				}
			}
			a.game = g
		}},
		{"won-easy", func(a *app) {
			a.level = game.Easy
			g := seeded(game.Easy)
			g.Open(4, 4)
			for y := range 9 {
				for x := range 9 {
					if !g.Cell(x, y).Mine {
						g.Open(x, y)
					}
				}
			}
			a.game = g
		}},
		{"fresh-hard", func(a *app) { a.level = game.Hard; a.restart(false) }},
	}

	h := &snapshotHarness{t: t, dir: dir, scenes: make([]func(*app), len(scenes)), names: make([]string, len(scenes))}
	for i, s := range scenes {
		h.scenes[i], h.names[i] = s.setup, s.name
	}
	ebiten.SetWindowSize(windowSize(game.Hard.Width, game.Hard.Height))
	if err := ebiten.RunGame(h); err != nil && err != ebiten.Termination {
		t.Fatal(err)
	}
}

type snapshotHarness struct {
	t      *testing.T
	dir    string
	scenes []func(*app)
	names  []string
	i      int
}

func (h *snapshotHarness) Update() error {
	if h.i >= len(h.scenes) {
		return ebiten.Termination
	}
	a := newApp(game.Medium, nil)
	h.scenes[h.i](a)
	a.LayoutF(0, 0)
	a.m = newMetrics(a.m.scale, a.level.Width, a.level.Height)

	img := ebiten.NewImage(a.m.width, a.m.height)
	a.Draw(img)
	rgba := image.NewRGBA(image.Rect(0, 0, a.m.width, a.m.height))
	img.ReadPixels(rgba.Pix)
	img.Deallocate()

	f, err := os.Create(filepath.Join(h.dir, h.names[h.i]+".png"))
	if err != nil {
		return err
	}
	defer f.Close()
	if err := png.Encode(f, rgba); err != nil {
		return err
	}
	h.t.Logf("wrote %s (scale %.2f)", f.Name(), a.m.scale)
	h.i++
	return nil
}

func (h *snapshotHarness) Draw(*ebiten.Image) {}

func (h *snapshotHarness) Layout(w, hh int) (int, int) { return w, hh }
