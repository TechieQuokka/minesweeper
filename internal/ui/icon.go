package ui

import (
	"image"
	"image/color"
	"math"
)

// iconSizes are the window icon resolutions handed to the OS; it picks the
// closest one for the title bar, taskbar and app switcher.
var iconSizes = []int{16, 24, 32, 48, 64, 128, 256}

// windowIcons renders the app icon at every size in iconSizes.
func windowIcons() []image.Image {
	icons := make([]image.Image, len(iconSizes))
	for i, s := range iconSizes {
		icons[i] = Icon(s)
	}
	return icons
}

// Icon renders the app icon, the yellow smiley from the restart button on a
// dark rounded tile, as a size×size image. It is rasterized on the CPU so it
// can be used before the game loop starts and outside of Ebitengine.
func Icon(size int) *image.RGBA {
	const ss = 4 // supersampling grid per axis
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	s := float64(size)

	// Geometry in pixels, relative to the icon size.
	c := s / 2
	tileInset := s * 0.04
	tileHalf := c - tileInset
	tileRadius := s * 0.22
	faceR := s * 0.36
	rimW := max(1, s*0.025)
	eyeY := c - faceR*0.25
	eyeDX := faceR * 0.38
	eyeR := faceR * 0.13
	smileCY := c + faceR*0.05
	smileR := faceR * 0.5
	smileHalfW := max(0.75, faceR*0.065)
	smileFrom, smileTo := math.Pi*0.2, math.Pi*0.8

	for py := range size {
		for px := range size {
			var acc [4]float64 // premultiplied RGBA accumulated over samples
			for sy := range ss {
				for sx := range ss {
					x := float64(px) + (float64(sx)+0.5)/ss
					y := float64(py) + (float64(sy)+0.5)/ss

					var clr color.RGBA
					switch {
					case sdRoundSquare(x-c, y-c, tileHalf, tileRadius) > 0:
						continue // outside the tile: transparent
					case math.Hypot(x-c, y-c) > faceR:
						clr = colButton
					case math.Hypot(x-c, y-c) > faceR-rimW:
						clr = colFaceRim
					case math.Hypot(x-(c-eyeDX), y-eyeY) <= eyeR,
						math.Hypot(x-(c+eyeDX), y-eyeY) <= eyeR,
						distArc(x-c, y-smileCY, smileR, smileFrom, smileTo) <= smileHalfW:
						clr = colFaceInk
					default:
						clr = colFace
					}
					acc[0] += float64(clr.R)
					acc[1] += float64(clr.G)
					acc[2] += float64(clr.B)
					acc[3] += 255
				}
			}
			const n = ss * ss
			img.SetRGBA(px, py, color.RGBA{
				R: uint8(math.Round(acc[0] / n)),
				G: uint8(math.Round(acc[1] / n)),
				B: uint8(math.Round(acc[2] / n)),
				A: uint8(math.Round(acc[3] / n)),
			})
		}
	}
	return img
}

// sdRoundSquare is the signed distance from (x, y) to a square of half-size
// h with corner radius r, centered at the origin. Negative means inside.
func sdRoundSquare(x, y, h, r float64) float64 {
	qx := math.Abs(x) - h + r
	qy := math.Abs(y) - h + r
	return math.Hypot(max(qx, 0), max(qy, 0)) + min(max(qx, qy), 0) - r
}

// distArc is the distance from (x, y) to the arc of radius r centered at the
// origin, spanning angles from..to (radians, y pointing down).
func distArc(x, y, r, from, to float64) float64 {
	a := math.Atan2(y, x)
	if a < 0 {
		a += 2 * math.Pi
	}
	if a >= from && a <= to {
		return math.Abs(math.Hypot(x, y) - r)
	}
	d1 := math.Hypot(x-r*math.Cos(from), y-r*math.Sin(from))
	d2 := math.Hypot(x-r*math.Cos(to), y-r*math.Sin(to))
	return min(d1, d2)
}
