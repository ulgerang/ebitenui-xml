package ui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// NineSlice represents a 9-slice (9-patch) image for scalable UI elements
// The image is divided into 9 regions that scale differently:
//
//	+---+-------+---+
//	| 1 |   2   | 3 |  <- corners (1,3,7,9) don't scale
//	+---+-------+---+
//	| 4 |   5   | 6 |  <- edges (2,8) scale horizontally, (4,6) scale vertically
//	+---+-------+---+
//	| 7 |   8   | 9 |  <- center (5) scales both ways
//	+---+-------+---+
type NineSlice struct {
	image *ebiten.Image

	// Slice boundaries (in pixels from each edge)
	Left   int
	Right  int
	Top    int
	Bottom int

	// Cached sub-images
	topLeft     *ebiten.Image
	top         *ebiten.Image
	topRight    *ebiten.Image
	left        *ebiten.Image
	center      *ebiten.Image
	right       *ebiten.Image
	bottomLeft  *ebiten.Image
	bottom      *ebiten.Image
	bottomRight *ebiten.Image
}

// NewNineSlice creates a new 9-slice from an image with specified borders
func NewNineSlice(img *ebiten.Image, left, right, top, bottom int) *NineSlice {
	ns := &NineSlice{
		image:  img,
		Left:   left,
		Right:  right,
		Top:    top,
		Bottom: bottom,
	}
	ns.slice()
	return ns
}

// slice creates the 9 sub-images
func (ns *NineSlice) slice() {
	if ns.image == nil {
		return
	}

	bounds := ns.image.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	// Calculate inner region boundaries
	innerLeft := ns.Left
	innerRight := w - ns.Right
	innerTop := ns.Top
	innerBottom := h - ns.Bottom

	// Top row
	ns.topLeft = subImage(ns.image, 0, 0, innerLeft, innerTop)
	ns.top = subImage(ns.image, innerLeft, 0, innerRight, innerTop)
	ns.topRight = subImage(ns.image, innerRight, 0, w, innerTop)

	// Middle row
	ns.left = subImage(ns.image, 0, innerTop, innerLeft, innerBottom)
	ns.center = subImage(ns.image, innerLeft, innerTop, innerRight, innerBottom)
	ns.right = subImage(ns.image, innerRight, innerTop, w, innerBottom)

	// Bottom row
	ns.bottomLeft = subImage(ns.image, 0, innerBottom, innerLeft, h)
	ns.bottom = subImage(ns.image, innerLeft, innerBottom, innerRight, h)
	ns.bottomRight = subImage(ns.image, innerRight, innerBottom, w, h)
}

// subImage extracts a sub-image
func subImage(img *ebiten.Image, x0, y0, x1, y1 int) *ebiten.Image {
	if x1 <= x0 || y1 <= y0 {
		return nil
	}
	return img.SubImage(image.Rect(x0, y0, x1, y1)).(*ebiten.Image)
}

// Draw draws the 9-slice at the target rectangle
func (ns *NineSlice) Draw(screen *ebiten.Image, x, y, width, height float64, colorScale *ebiten.ColorScale) {
	if ns.image == nil {
		return
	}

	left := float64(ns.Left)
	right := float64(ns.Right)
	top := float64(ns.Top)
	bottom := float64(ns.Bottom)

	// Calculate middle region size
	middleW := width - left - right
	middleH := height - top - bottom

	if middleW < 0 {
		middleW = 0
	}
	if middleH < 0 {
		middleH = 0
	}

	// Draw corners (no scaling)
	ns.drawPart(screen, ns.topLeft, x, y, left, top, colorScale)
	ns.drawPart(screen, ns.topRight, x+width-right, y, right, top, colorScale)
	ns.drawPart(screen, ns.bottomLeft, x, y+height-bottom, left, bottom, colorScale)
	ns.drawPart(screen, ns.bottomRight, x+width-right, y+height-bottom, right, bottom, colorScale)

	// Draw edges (scale in one direction)
	ns.drawPart(screen, ns.top, x+left, y, middleW, top, colorScale)
	ns.drawPart(screen, ns.bottom, x+left, y+height-bottom, middleW, bottom, colorScale)
	ns.drawPart(screen, ns.left, x, y+top, left, middleH, colorScale)
	ns.drawPart(screen, ns.right, x+width-right, y+top, right, middleH, colorScale)

	// Draw center (scale both ways)
	ns.drawPart(screen, ns.center, x+left, y+top, middleW, middleH, colorScale)
}

// drawPart draws a part of the 9-slice with scaling
func (ns *NineSlice) drawPart(screen, part *ebiten.Image, x, y, w, h float64, colorScale *ebiten.ColorScale) {
	if part == nil || w <= 0 || h <= 0 {
		return
	}

	bounds := part.Bounds()
	srcW := float64(bounds.Dx())
	srcH := float64(bounds.Dy())

	if srcW <= 0 || srcH <= 0 {
		return
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(w/srcW, h/srcH)
	op.GeoM.Translate(x, y)

	if colorScale != nil {
		op.ColorScale = *colorScale
	}

	screen.DrawImage(part, op)
}

// MinSize returns the minimum size this 9-slice can be drawn at
func (ns *NineSlice) MinSize() (int, int) {
	return ns.Left + ns.Right, ns.Top + ns.Bottom
}
