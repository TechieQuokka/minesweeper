package ui

import (
	"bytes"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/gomonobold"
)

var (
	fontBold = mustFont(gobold.TTF)
	fontMono = mustFont(gomonobold.TTF)
)

func mustFont(ttf []byte) *text.GoTextFaceSource {
	src, err := text.NewGoTextFaceSource(bytes.NewReader(ttf))
	if err != nil {
		panic(err)
	}
	return src
}

// fillRoundRect fills a rounded rectangle.
func fillRoundRect(dst *ebiten.Image, x, y, w, h, r float32, clr color.Color) {
	var p vector.Path
	r = min(r, w/2, h/2)
	p.MoveTo(x+r, y)
	p.LineTo(x+w-r, y)
	p.ArcTo(x+w, y, x+w, y+r, r)
	p.LineTo(x+w, y+h-r)
	p.ArcTo(x+w, y+h, x+w-r, y+h, r)
	p.LineTo(x+r, y+h)
	p.ArcTo(x, y+h, x, y+h-r, r)
	p.LineTo(x, y+r)
	p.ArcTo(x, y, x+r, y, r)
	p.Close()
	fillPath(dst, &p, clr)
}

func fillPath(dst *ebiten.Image, p *vector.Path, clr color.Color) {
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(clr)
	vector.FillPath(dst, p, nil, op)
}

func strokeLine(dst *ebiten.Image, x0, y0, x1, y1, width float32, clr color.Color) {
	var p vector.Path
	p.MoveTo(x0, y0)
	p.LineTo(x1, y1)
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(clr)
	vector.StrokePath(dst, &p, &vector.StrokeOptions{Width: width, LineCap: vector.LineCapRound}, op)
}

func strokeArc(dst *ebiten.Image, cx, cy, r, from, to, width float32, clr color.Color) {
	var p vector.Path
	p.Arc(cx, cy, r, from, to, vector.Clockwise)
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(clr)
	vector.StrokePath(dst, &p, &vector.StrokeOptions{Width: width, LineCap: vector.LineCapRound}, op)
}

// drawText draws s centered at (cx, cy). size is in device pixels.
func drawText(dst *ebiten.Image, s string, src *text.GoTextFaceSource, size, cx, cy float64, clr color.Color) {
	face := &text.GoTextFace{Source: src, Size: size}
	op := &text.DrawOptions{}
	op.GeoM.Translate(cx, cy)
	op.ColorScale.ScaleWithColor(clr)
	op.PrimaryAlign = text.AlignCenter
	op.SecondaryAlign = text.AlignCenter
	text.Draw(dst, s, face, op)
}

// tile identifies a pre-rendered cell image.
type tile int

const (
	tileHidden tile = iota
	tileOpened      // also used for pressed cells and 0
	tileFlag
	tileQuestion
	tileMine
	tileMineExploded
	tileWrongFlag
	tileNumber1 // tileNumber1+n-1 for n in 1..8
	tileCount   = tileNumber1 + 8
)

// buildTiles renders every cell image at the given cell size in pixels.
func buildTiles(size int) [tileCount]*ebiten.Image {
	var tiles [tileCount]*ebiten.Image
	s := float32(size)
	inset := max(1, s*0.05)
	radius := s * 0.16

	base := func(bg color.Color, hidden bool) *ebiten.Image {
		img := ebiten.NewImage(size, size)
		if hidden {
			// A slightly lighter rim on top gives a hint of depth without bevels.
			fillRoundRect(img, inset, inset, s-2*inset, s-2*inset, radius, colHiddenEdge)
			fillRoundRect(img, inset, inset+max(1, s*0.06), s-2*inset, s-2*inset-max(1, s*0.06), radius, bg)
		} else {
			fillRoundRect(img, inset, inset, s-2*inset, s-2*inset, radius, bg)
		}
		return img
	}

	tiles[tileHidden] = base(colHidden, true)
	tiles[tileOpened] = base(colOpened, false)

	tiles[tileFlag] = base(colHidden, true)
	drawFlag(tiles[tileFlag], s)

	tiles[tileQuestion] = base(colHidden, true)
	drawText(tiles[tileQuestion], "?", fontBold, float64(s)*0.6, float64(s)/2, float64(s)/2, colText)

	tiles[tileMine] = base(colOpened, false)
	drawMine(tiles[tileMine], s, colMine, colOpened)

	tiles[tileMineExploded] = base(colExploded, false)
	drawMine(tiles[tileMineExploded], s, colExplodedFg, colExploded)

	tiles[tileWrongFlag] = base(colOpened, false)
	drawMine(tiles[tileWrongFlag], s, colMineDim, colOpened)
	w := s * 0.09
	strokeLine(tiles[tileWrongFlag], s*0.24, s*0.24, s*0.76, s*0.76, w, colWrong)
	strokeLine(tiles[tileWrongFlag], s*0.76, s*0.24, s*0.24, s*0.76, w, colWrong)

	for n := 1; n <= 8; n++ {
		img := base(colOpened, false)
		drawText(img, string(rune('0'+n)), fontBold, float64(s)*0.6, float64(s)/2, float64(s)/2, numberColors[n])
		tiles[tileNumber1+tile(n-1)] = img
	}
	return tiles
}

func drawFlag(dst *ebiten.Image, s float32) {
	poleX := s * 0.40
	strokeLine(dst, poleX, s*0.24, poleX, s*0.74, max(1.5, s*0.06), colPole)
	strokeLine(dst, s*0.30, s*0.76, s*0.62, s*0.76, max(1.5, s*0.07), colPole)
	var p vector.Path
	p.MoveTo(poleX, s*0.22)
	p.LineTo(s*0.74, s*0.37)
	p.LineTo(poleX, s*0.52)
	p.Close()
	fillPath(dst, &p, colFlag)
}

func drawMine(dst *ebiten.Image, s float32, body, bg color.Color) {
	c := s / 2
	r := s * 0.2
	spike := r * 1.55
	w := max(1.5, s*0.07)
	for i := range 4 {
		a := float64(i) * math.Pi / 4
		dx, dy := float32(math.Cos(a))*spike, float32(math.Sin(a))*spike
		strokeLine(dst, c-dx, c-dy, c+dx, c+dy, w, body)
	}
	vector.FillCircle(dst, c, c, r, body, true)
	vector.FillCircle(dst, c-r*0.35, c-r*0.35, r*0.25, bg, true)
}

// faceKind is the expression of the restart button.
type faceKind int

const (
	faceNormal faceKind = iota
	faceSurprised
	faceCool
	faceDead
)

// drawFace draws the restart button centered at (cx, cy).
func drawFace(dst *ebiten.Image, cx, cy float32, size float32, kind faceKind, pressed bool) {
	bg := colButton
	if pressed {
		bg = colButtonDown
		cy += size * 0.03
	}
	fillRoundRect(dst, cx-size/2, cy-size/2, size, size, size*0.22, bg)

	r := size * 0.33
	vector.FillCircle(dst, cx, cy, r, colFace, true)

	ink := colFaceInk
	lw := max(1.5, size*0.05)
	eyeY := cy - r*0.25
	eyeDX := r * 0.38
	eyeR := r * 0.12

	switch kind {
	case faceNormal:
		vector.FillCircle(dst, cx-eyeDX, eyeY, eyeR, ink, true)
		vector.FillCircle(dst, cx+eyeDX, eyeY, eyeR, ink, true)
		strokeArc(dst, cx, cy+r*0.05, r*0.5, math.Pi*0.2, math.Pi*0.8, lw, ink)
	case faceSurprised:
		vector.FillCircle(dst, cx-eyeDX, eyeY, eyeR*1.2, ink, true)
		vector.FillCircle(dst, cx+eyeDX, eyeY, eyeR*1.2, ink, true)
		vector.StrokeCircle(dst, cx, cy+r*0.42, r*0.17, lw, ink, true)
	case faceCool:
		gw, gh := r*0.6, r*0.36
		fillRoundRect(dst, cx-eyeDX-gw/2, eyeY-gh/2, gw, gh, gh*0.4, ink)
		fillRoundRect(dst, cx+eyeDX-gw/2, eyeY-gh/2, gw, gh, gh*0.4, ink)
		strokeLine(dst, cx-r*0.85, eyeY-gh*0.2, cx+r*0.85, eyeY-gh*0.2, lw*0.8, ink)
		strokeArc(dst, cx, cy+r*0.05, r*0.5, math.Pi*0.2, math.Pi*0.8, lw, ink)
	case faceDead:
		d := eyeR * 1.4
		for _, ex := range []float32{cx - eyeDX, cx + eyeDX} {
			strokeLine(dst, ex-d, eyeY-d, ex+d, eyeY+d, lw*0.8, ink)
			strokeLine(dst, ex+d, eyeY-d, ex-d, eyeY+d, lw*0.8, ink)
		}
		strokeArc(dst, cx, cy+r*0.8, r*0.4, math.Pi*1.2, math.Pi*1.8, lw, ink)
	}
}

// formatCounter formats a 3-digit counter, e.g. 040, -02, 999.
func formatCounter(n int) string {
	if n < 0 {
		n = min(-n, 99)
		return "-" + string(rune('0'+n/10)) + string(rune('0'+n%10))
	}
	n = min(n, 999)
	return string([]rune{rune('0' + n/100), rune('0' + n/10%10), rune('0' + n%10)})
}
