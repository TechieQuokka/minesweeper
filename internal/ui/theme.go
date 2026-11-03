package ui

import (
	"image/color"
	"math"
)

// Dark-only palette.
var (
	colBackground = rgb(0x17, 0x18, 0x1c)
	colPanel      = rgb(0x22, 0x24, 0x2b)
	colHidden     = rgb(0x3b, 0x3f, 0x4b)
	colHiddenEdge = rgb(0x46, 0x4b, 0x58) // subtle top highlight on hidden cells
	colOpened     = rgb(0x2a, 0x2d, 0x35)
	colDisplay    = rgb(0x12, 0x13, 0x16)
	colDisplayFg  = rgb(0xff, 0x8a, 0x5c)
	colButton     = rgb(0x30, 0x33, 0x3c)
	colButtonDown = rgb(0x26, 0x28, 0x2f)
	colText       = rgb(0xc9, 0xce, 0xd6)
	colFlag       = rgb(0xff, 0x63, 0x5a)
	colPole       = rgb(0xe4, 0xe7, 0xec)
	colMine       = rgb(0xd0, 0xd4, 0xdc)
	colMineDim    = rgb(0x6b, 0x70, 0x7a)
	colExploded   = rgb(0xe5, 0x53, 0x4b)
	colExplodedFg = rgb(0x1c, 0x1d, 0x21)
	colWrong      = rgb(0xff, 0x6b, 0x6b)
	colFace       = rgb(0xf4, 0xc5, 0x42)
	colFaceInk    = rgb(0x1b, 0x1c, 0x20)
	colFaceRim    = rgb(0xc9, 0x9a, 0x1e) // app icon only: keeps the face crisp at small sizes
)

// numberColors[n] is the color of the digit n (1..8).
var numberColors = [9]color.RGBA{
	{},
	rgb(0x5a, 0xa9, 0xff), // 1 blue
	rgb(0x5f, 0xd3, 0x8d), // 2 green
	rgb(0xff, 0x6b, 0x6b), // 3 red
	rgb(0xb1, 0x8c, 0xff), // 4 purple
	rgb(0xff, 0xb4, 0x54), // 5 orange
	rgb(0x4f, 0xd1, 0xc5), // 6 teal
	rgb(0xe6, 0xe6, 0xe6), // 7 white
	rgb(0x9a, 0xa0, 0xa6), // 8 gray
}

func rgb(r, g, b uint8) color.RGBA { return color.RGBA{r, g, b, 0xff} }

// Layout constants in device-independent pixels.
const (
	dipPad      = 14
	dipBarH     = 52
	dipGap      = 10
	dipCell     = 28
	dipDisplayW = 66
	dipDisplayH = 34
	dipFace     = 40
)

// metrics holds layout values in device pixels for a given scale factor.
type metrics struct {
	scale    float64
	pad      int
	barH     int
	gap      int
	cell     int
	displayW int
	displayH int
	face     int
	boardX   int
	boardY   int
	width    int
	height   int
}

func newMetrics(scale float64, cols, rows int) metrics {
	px := func(v float64) int { return int(math.Round(v * scale)) }
	m := metrics{
		scale:    scale,
		pad:      px(dipPad),
		barH:     px(dipBarH),
		gap:      px(dipGap),
		cell:     px(dipCell),
		displayW: px(dipDisplayW),
		displayH: px(dipDisplayH),
		face:     px(dipFace),
	}
	m.boardX = m.pad
	m.boardY = m.pad + m.barH + m.gap
	m.width = 2*m.pad + cols*m.cell
	m.height = m.boardY + rows*m.cell + m.pad
	return m
}

// windowSize returns the window size in device-independent pixels.
func windowSize(cols, rows int) (w, h int) {
	return 2*dipPad + cols*dipCell, dipPad + dipBarH + dipGap + rows*dipCell + dipPad
}
