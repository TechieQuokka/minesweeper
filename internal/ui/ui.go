// Package ui renders a game.Game with Ebitengine and translates mouse and
// keyboard input into game actions.
package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/TechieQuokka/minesweeper/internal/game"
)

// EndFunc is called once per round when the game is won or lost.
type EndFunc func(won bool, seconds int)

// Run opens the game window and blocks until it is closed.
func Run(level game.Level, onEnd EndFunc) error {
	w, h := windowSize(level.Width, level.Height)
	ebiten.SetWindowSize(w, h)
	ebiten.SetWindowTitle("mine — " + level.Name)
	ebiten.SetWindowIcon(windowIcons())
	return ebiten.RunGameWithOptions(newApp(level, onEnd), &ebiten.RunGameOptions{
		// GNOME and other docks ignore the window icon and instead match
		// WM_CLASS against desktop entries; it must equal the desktop file
		// name (or its StartupWMClass) for the taskbar to show our icon.
		X11ClassName:    AppID,
		X11InstanceName: AppID,
	})
}

// AppID identifies the app to the desktop: it is the WM_CLASS of the window
// and the expected basename of the desktop entry and icon.
const AppID = "mine-minesweeper"

type app struct {
	level game.Level
	game  *game.Game
	onEnd EndFunc

	m     metrics
	tiles [tileCount]*ebiten.Image

	reported     bool // onEnd already called for this round
	leftPressing bool // left button held for a plain open
	chording     bool // both buttons or middle button held
	waitRelease  bool // ignore input until every button is released
	facePressed  bool // left button went down on the face button
	cursorX      int  // cursor position from the last Update, in screen pixels
	cursorY      int
}

// mouseInput is a snapshot of the mouse for one frame.
type mouseInput struct {
	x, y                       int
	left, right, middle        bool // held
	leftDown, rightDown, midDn bool // pressed this frame
	leftUp, rightUp, midUp     bool // released this frame
}

func (in mouseInput) anyHeld() bool { return in.left || in.right || in.middle }

func readMouse() mouseInput {
	x, y := ebiten.CursorPosition()
	return mouseInput{
		x: x, y: y,
		left:      ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft),
		right:     ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight),
		middle:    ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle),
		leftDown:  inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft),
		rightDown: inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight),
		midDn:     inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonMiddle),
		leftUp:    inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft),
		rightUp:   inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonRight),
		midUp:     inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonMiddle),
	}
}

func newApp(level game.Level, onEnd EndFunc) *app {
	a := &app{level: level, onEnd: onEnd}
	a.restart(false)
	return a
}

// restart starts a new round. If buttons are still held, input is ignored
// until they are released so the restarting click does not leak into the board.
func (a *app) restart(buttonsHeld bool) {
	a.game = game.New(a.level)
	a.reported = false
	a.leftPressing = false
	a.chording = false
	a.facePressed = false
	a.waitRelease = buttonsHeld
}

// LayoutF makes the logical screen match the window in device pixels so
// everything is drawn crisply on HiDPI displays.
func (a *app) LayoutF(_, _ float64) (float64, float64) {
	scale := ebiten.Monitor().DeviceScaleFactor()
	if a.tiles[0] == nil || scale != a.m.scale {
		a.m = newMetrics(scale, a.level.Width, a.level.Height)
		for _, t := range a.tiles {
			if t != nil {
				t.Deallocate()
			}
		}
		a.tiles = buildTiles(a.m.cell)
	}
	return float64(a.m.width), float64(a.m.height)
}

// Layout is required by ebiten.Game but unused because LayoutF is implemented.
func (a *app) Layout(w, h int) (int, int) {
	fw, fh := a.LayoutF(float64(w), float64(h))
	return int(fw), int(fh)
}

func (a *app) Update() error {
	in := readMouse()
	a.cursorX, a.cursorY = in.x, in.y
	if inpututil.IsKeyJustPressed(ebiten.KeyF2) {
		a.restart(in.anyHeld())
		return nil
	}
	a.handleMouse(in)
	a.reportEnd()
	return nil
}

// reportEnd calls onEnd once when the round finishes.
func (a *app) reportEnd() {
	if a.game.Over() && !a.reported {
		a.reported = true
		if a.onEnd != nil {
			a.onEnd(a.game.Status() == game.Won, a.game.Seconds())
		}
	}
}

func (a *app) handleMouse(in mouseInput) {
	left, right, middle := in.left, in.right, in.middle
	leftDown, rightDown, middleDown := in.leftDown, in.rightDown, in.midDn
	leftUp, rightUp, middleUp := in.leftUp, in.rightUp, in.midUp
	anyHeld := in.anyHeld()

	if a.waitRelease {
		if !anyHeld {
			a.waitRelease = false
		}
		return
	}

	cx, cy := in.x, in.y
	x, y, onBoard := a.cellAt(cx, cy)

	// Face button.
	if leftDown && !right && !middle && a.onFace(cx, cy) {
		a.facePressed = true
		return
	}
	if a.facePressed {
		if !left {
			a.facePressed = false
			if leftUp && a.onFace(cx, cy) {
				a.restart(false)
			}
		}
		return
	}

	// Presses.
	if leftDown {
		if right {
			a.chording = true
			a.leftPressing = false
		} else {
			a.leftPressing = true
		}
	}
	if rightDown {
		if left {
			a.chording = true
			a.leftPressing = false
		} else if onBoard {
			a.game.ToggleMark(x, y)
		}
	}
	if middleDown {
		a.chording = true
		a.leftPressing = false
	}

	// Releases.
	if a.chording && (leftUp || rightUp || middleUp) {
		if onBoard {
			a.game.Chord(x, y)
		}
		a.chording = false
		a.waitRelease = anyHeld
		return
	}
	if a.leftPressing && leftUp {
		a.leftPressing = false
		if onBoard {
			a.game.Open(x, y)
		}
	}

	// Recover from releases we never saw (e.g. focus lost while held).
	if !left {
		a.leftPressing = false
	}
	if !anyHeld {
		a.chording = false
	}
}

// cellAt maps a screen position to board coordinates.
func (a *app) cellAt(px, py int) (x, y int, ok bool) {
	rx, ry := px-a.m.boardX, py-a.m.boardY
	if rx < 0 || ry < 0 {
		return 0, 0, false
	}
	x, y = rx/a.m.cell, ry/a.m.cell
	return x, y, a.game.InBounds(x, y)
}

func (a *app) faceCenter() (float32, float32) {
	return float32(a.m.width) / 2, float32(a.m.pad) + float32(a.m.barH)/2
}

func (a *app) onFace(px, py int) bool {
	fx, fy := a.faceCenter()
	half := float32(a.m.face) / 2
	x, y := float32(px), float32(py)
	return x >= fx-half && x < fx+half && y >= fy-half && y < fy+half
}

func (a *app) Draw(screen *ebiten.Image) {
	screen.Fill(colBackground)
	a.drawBar(screen)
	a.drawBoard(screen)
}

func (a *app) drawBar(screen *ebiten.Image) {
	m := a.m
	pad, barH := float32(m.pad), float32(m.barH)
	boardW := float32(m.width) - 2*pad
	fillRoundRect(screen, pad, pad, boardW, barH, barH*0.2, colPanel)

	dw, dh := float32(m.displayW), float32(m.displayH)
	inner := (barH - dh) / 2
	cy := pad + barH/2
	fontSize := float64(dh) * 0.72

	// Mines left (left).
	lx := pad + inner
	fillRoundRect(screen, lx, cy-dh/2, dw, dh, dh*0.25, colDisplay)
	drawText(screen, formatCounter(a.game.MinesLeft()), fontMono, fontSize, float64(lx+dw/2), float64(cy), colDisplayFg)

	// Timer (right).
	rx := pad + boardW - inner - dw
	fillRoundRect(screen, rx, cy-dh/2, dw, dh, dh*0.25, colDisplay)
	drawText(screen, formatCounter(a.game.Seconds()), fontMono, fontSize, float64(rx+dw/2), float64(cy), colDisplayFg)

	// Face (center).
	fx, fy := a.faceCenter()
	pressed := a.facePressed && a.onFace(a.cursorX, a.cursorY)
	drawFace(screen, fx, fy, float32(m.face), a.faceKind(), pressed)
}

func (a *app) faceKind() faceKind {
	switch a.game.Status() {
	case game.Won:
		return faceCool
	case game.Lost:
		return faceDead
	}
	if a.waitRelease || a.facePressed {
		return faceNormal
	}
	if a.leftPressing || a.chording {
		if _, _, ok := a.cellAt(a.cursorX, a.cursorY); ok {
			return faceSurprised
		}
	}
	return faceNormal
}

func (a *app) drawBoard(screen *ebiten.Image) {
	m := a.m
	g := a.game
	lv := a.level

	// Panel behind the board.
	pad := float32(m.pad) * 0.4
	fillRoundRect(screen,
		float32(m.boardX)-pad, float32(m.boardY)-pad,
		float32(lv.Width*m.cell)+2*pad, float32(lv.Height*m.cell)+2*pad,
		float32(m.cell)*0.3, colPanel)

	pressed := a.pressedCells()
	ex, ey, exploded := g.Exploded()
	lost := g.Status() == game.Lost

	op := &ebiten.DrawImageOptions{}
	for y := range lv.Height {
		for x := range lv.Width {
			c := g.Cell(x, y)
			t := tileFor(c, lost, exploded && x == ex && y == ey)
			if pressed[[2]int{x, y}] {
				t = tileOpened
			}
			op.GeoM.Reset()
			op.GeoM.Translate(float64(m.boardX+x*m.cell), float64(m.boardY+y*m.cell))
			screen.DrawImage(a.tiles[t], op)
		}
	}
}

// pressedCells returns the hidden/question cells currently drawn as pressed.
func (a *app) pressedCells() map[[2]int]bool {
	if a.game.Over() || a.waitRelease || (!a.leftPressing && !a.chording) {
		return nil
	}
	x, y, ok := a.cellAt(a.cursorX, a.cursorY)
	if !ok {
		return nil
	}
	out := map[[2]int]bool{}
	add := func(x, y int) {
		if !a.game.InBounds(x, y) {
			return
		}
		if s := a.game.Cell(x, y).State; s == game.Hidden || s == game.Question {
			out[[2]int{x, y}] = true
		}
	}
	if a.chording {
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				add(x+dx, y+dy)
			}
		}
	} else {
		add(x, y)
	}
	return out
}

// tileFor picks the image for a cell.
func tileFor(c game.Cell, lost, exploded bool) tile {
	if exploded {
		return tileMineExploded
	}
	switch c.State {
	case game.Opened:
		if c.Mine {
			return tileMineExploded
		}
		if c.Adjacent > 0 {
			return tileNumber1 + tile(c.Adjacent-1)
		}
		return tileOpened
	case game.Flagged:
		if lost && !c.Mine {
			return tileWrongFlag
		}
		return tileFlag
	}
	if lost && c.Mine {
		return tileMine
	}
	if c.State == game.Question {
		return tileQuestion
	}
	return tileHidden
}
