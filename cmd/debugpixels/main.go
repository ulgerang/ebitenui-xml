package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
)

// CellRegion returns pixel rect for test case index i in the grid.
func CellRegion(i int) image.Rectangle {
	cols := 5
	padX := 10
	padY := 30
	marginX := 5
	marginY := 5
	CellW := 200
	CellH := 150

	col := i % cols
	row := i / cols
	x := padX + col*(CellW+marginX*2) + marginX
	y := padX + row*(CellH+marginY*2+padY) + padY + marginY
	return image.Rect(x, y, x+CellW, y+CellH)
}

func loadPNG(path string) image.Image {
	f, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot open %s: %v\n", path, err)
		os.Exit(1)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot decode %s: %v\n", path, err)
		os.Exit(1)
	}
	return img
}

func px(img image.Image, x, y int) color.RGBA {
	r, g, b, a := img.At(x, y).RGBA()
	return color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
}

func fmtC(c color.RGBA) string {
	return fmt.Sprintf("rgba(%3d,%3d,%3d,%3d)", c.R, c.G, c.B, c.A)
}

func sampleAndPrint(label string, ebiten, browser image.Image, x, y int) {
	ec := px(ebiten, x, y)
	bc := px(browser, x, y)
	match := "OK"
	dr := int(ec.R) - int(bc.R)
	dg := int(ec.G) - int(bc.G)
	db := int(ec.B) - int(bc.B)
	if abs(dr) > 10 || abs(dg) > 10 || abs(db) > 10 {
		match = fmt.Sprintf("DIFF dr=%d dg=%d db=%d", dr, dg, db)
	}
	fmt.Printf("  %-20s (%4d,%4d) ebiten=%s browser=%s %s\n", label, x, y, fmtC(ec), fmtC(bc), match)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func main() {
	ebitenImg := loadPNG("cmd/css_testloop/output/ebiten.png")
	browserImg := loadPNG("cmd/css_testloop/output/browser.png")

	testCases := []struct {
		name  string
		index int
	}{
		{"bg-gradient-v", 2},
		{"bg-gradient-diag", 3},
		{"box-shadow", 7},
		{"flex-row", 8},
		{"flex-col", 9},
		{"justify-center", 10},
		{"padding", 15},
		{"opacity", 16},
		{"nested-layout", 17},
		{"overflow-hidden", 18},
		{"border-bg-combined", 19},
		{"flex-grow", 14},
	}

	for _, tc := range testCases {
		r := CellRegion(tc.index)
		fmt.Printf("\n=== %s (cell %d) region=(%d,%d)-(%d,%d) ===\n",
			tc.name, tc.index, r.Min.X, r.Min.Y, r.Max.X, r.Max.Y)

		switch tc.name {
		case "bg-gradient-v":
			// 180deg gradient: green (#2ecc71) at top → purple (#8e44ad) at bottom
			// Widget at (0,0) within cell, size 180x130
			sampleAndPrint("top-left", ebitenImg, browserImg, r.Min.X+5, r.Min.Y+5)
			sampleAndPrint("top-center", ebitenImg, browserImg, r.Min.X+90, r.Min.Y+5)
			sampleAndPrint("center", ebitenImg, browserImg, r.Min.X+90, r.Min.Y+65)
			sampleAndPrint("bottom-center", ebitenImg, browserImg, r.Min.X+90, r.Min.Y+125)
			sampleAndPrint("bottom-left", ebitenImg, browserImg, r.Min.X+5, r.Min.Y+125)
			// Check if gradient runs horizontally or vertically
			sampleAndPrint("mid-left", ebitenImg, browserImg, r.Min.X+5, r.Min.Y+65)
			sampleAndPrint("mid-right", ebitenImg, browserImg, r.Min.X+175, r.Min.Y+65)

		case "bg-gradient-diag":
			// 45deg gradient: orange (#f39c12) → blue (#2980b9)
			sampleAndPrint("top-left", ebitenImg, browserImg, r.Min.X+5, r.Min.Y+5)
			sampleAndPrint("center", ebitenImg, browserImg, r.Min.X+90, r.Min.Y+65)
			sampleAndPrint("bottom-right", ebitenImg, browserImg, r.Min.X+175, r.Min.Y+125)
			sampleAndPrint("top-right", ebitenImg, browserImg, r.Min.X+175, r.Min.Y+5)
			sampleAndPrint("bottom-left", ebitenImg, browserImg, r.Min.X+5, r.Min.Y+125)

		case "opacity":
			// Red box with opacity 0.5 on dark bg
			sampleAndPrint("center", ebitenImg, browserImg, r.Min.X+90, r.Min.Y+65)
			sampleAndPrint("top-left+5", ebitenImg, browserImg, r.Min.X+5, r.Min.Y+5)
			sampleAndPrint("outside(190,5)", ebitenImg, browserImg, r.Min.X+190, r.Min.Y+5)

		case "flex-row":
			// 3 boxes (40x40) with gap 8, padding 10, direction row
			// Children start at (10,10) with gap=8
			// First child: x=10, y=10, w=40, h=40
			sampleAndPrint("parent-bg(5,5)", ebitenImg, browserImg, r.Min.X+5, r.Min.Y+5)
			sampleAndPrint("child1-center(30,30)", ebitenImg, browserImg, r.Min.X+30, r.Min.Y+30)
			sampleAndPrint("child2-center(78,30)", ebitenImg, browserImg, r.Min.X+78, r.Min.Y+30)
			sampleAndPrint("child3-center(126,30)", ebitenImg, browserImg, r.Min.X+126, r.Min.Y+30)
			// Check if children are placed correctly
			sampleAndPrint("gap-area(55,30)", ebitenImg, browserImg, r.Min.X+55, r.Min.Y+30)
			sampleAndPrint("below-children(30,80)", ebitenImg, browserImg, r.Min.X+30, r.Min.Y+80)

		case "justify-center":
			// 2 boxes (30x30) centered, padding 10, bg #2c3e50
			sampleAndPrint("bg(5,5)", ebitenImg, browserImg, r.Min.X+5, r.Min.Y+5)
			sampleAndPrint("center(90,65)", ebitenImg, browserImg, r.Min.X+90, r.Min.Y+65)

		case "padding":
			// Parent #2c3e50, padding 20top/10right/20bot/10left, inner child flexGrow=1 #e74c3c
			// Inner should fill: x=10, y=20, w=160, h=90
			sampleAndPrint("outer-top-left(5,5)", ebitenImg, browserImg, r.Min.X+5, r.Min.Y+5)
			sampleAndPrint("inner-center(90,65)", ebitenImg, browserImg, r.Min.X+90, r.Min.Y+65)
			sampleAndPrint("inner-top-left(15,25)", ebitenImg, browserImg, r.Min.X+15, r.Min.Y+25)
			sampleAndPrint("left-pad(5,65)", ebitenImg, browserImg, r.Min.X+5, r.Min.Y+65)
			sampleAndPrint("top-pad(90,10)", ebitenImg, browserImg, r.Min.X+90, r.Min.Y+10)

		case "border-bg-combined":
			// BG #1abc9c, border 4px #f39c12, radius 16
			sampleAndPrint("center", ebitenImg, browserImg, r.Min.X+90, r.Min.Y+65)
			sampleAndPrint("top-edge(90,3)", ebitenImg, browserImg, r.Min.X+90, r.Min.Y+3)
			sampleAndPrint("left-edge(3,65)", ebitenImg, browserImg, r.Min.X+3, r.Min.Y+65)
			// Check outside the widget area (should be dark bg)
			sampleAndPrint("outside(185,5)", ebitenImg, browserImg, r.Min.X+185, r.Min.Y+5)

		case "box-shadow":
			// Widget with margin 15, so widget starts at (15,15), 160x110
			// Shadow: 4 4 12 0 rgba(0,0,0,0.5)
			sampleAndPrint("widget-center(95,70)", ebitenImg, browserImg, r.Min.X+95, r.Min.Y+70)
			sampleAndPrint("shadow-area(180,130)", ebitenImg, browserImg, r.Min.X+180, r.Min.Y+130)
			sampleAndPrint("no-shadow(10,10)", ebitenImg, browserImg, r.Min.X+10, r.Min.Y+10)

		case "flex-col":
			// 3 children height=25, gap=6, padding=10
			sampleAndPrint("child1(90,22)", ebitenImg, browserImg, r.Min.X+90, r.Min.Y+22)
			sampleAndPrint("child2(90,53)", ebitenImg, browserImg, r.Min.X+90, r.Min.Y+53)
			sampleAndPrint("child3(90,84)", ebitenImg, browserImg, r.Min.X+90, r.Min.Y+84)
			sampleAndPrint("bg(5,5)", ebitenImg, browserImg, r.Min.X+5, r.Min.Y+5)

		case "nested-layout":
			sampleAndPrint("bg(5,5)", ebitenImg, browserImg, r.Min.X+5, r.Min.Y+5)
			sampleAndPrint("top-left-child(30,30)", ebitenImg, browserImg, r.Min.X+30, r.Min.Y+30)
			sampleAndPrint("bot-child(90,115)", ebitenImg, browserImg, r.Min.X+90, r.Min.Y+115)

		case "overflow-hidden":
			// Parent 180x130 bg=#2c3e50, child 300x300 bg=#e74c3c, overflow hidden
			sampleAndPrint("center(90,65)", ebitenImg, browserImg, r.Min.X+90, r.Min.Y+65)
			sampleAndPrint("top-left(5,5)", ebitenImg, browserImg, r.Min.X+5, r.Min.Y+5)
			sampleAndPrint("outside(185,5)", ebitenImg, browserImg, r.Min.X+185, r.Min.Y+5)

		case "flex-grow":
			// 2 children: flexGrow 1 and 2, height=40, padding=10
			sampleAndPrint("child1-center(40,30)", ebitenImg, browserImg, r.Min.X+40, r.Min.Y+30)
			sampleAndPrint("child2-center(130,30)", ebitenImg, browserImg, r.Min.X+130, r.Min.Y+30)
			sampleAndPrint("bg(5,5)", ebitenImg, browserImg, r.Min.X+5, r.Min.Y+5)
		}
	}
}
