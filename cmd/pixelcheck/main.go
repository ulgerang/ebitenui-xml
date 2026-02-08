package main

import (
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
)

const (
	CellW   = 200
	CellH   = 150
	cols    = 5
	padX    = 10
	padY    = 30
	marginX = 5
	marginY = 5
)

func cellRegion(i int) image.Rectangle {
	col := i % cols
	row := i / cols
	x := padX + col*(CellW+marginX*2) + marginX
	y := padX + row*(CellH+marginY*2+padY) + padY + marginY
	return image.Rect(x, y, x+CellW, y+CellH)
}

func loadPNG(path string) image.Image {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		panic(err)
	}
	return img
}

func main() {
	ebitenImg := loadPNG("cmd/css_testloop/output/ebiten.png")
	browserImg := loadPNG("cmd/css_testloop/output/browser.png")

	tests := []struct {
		idx  int
		name string
	}{
		{1, "bg-gradient-h (90deg)"},
		{2, "bg-gradient-v (180deg)"},
		{3, "bg-gradient-diag (45deg)"},
	}

	for _, t := range tests {
		region := cellRegion(t.idx)
		fmt.Printf("\n=== %s (idx=%d) cell=(%d,%d)-(%d,%d) ===\n",
			t.name, t.idx, region.Min.X, region.Min.Y, region.Max.X, region.Max.Y)

		// Sample pixels along the vertical center and horizontal center
		cx := (region.Min.X + region.Max.X) / 2
		cy := (region.Min.Y + region.Max.Y) / 2

		// Sample horizontal line through center
		fmt.Println("  Horizontal line (y=center):")
		for _, frac := range []float64{0.0, 0.25, 0.5, 0.75, 1.0} {
			x := region.Min.X + int(frac*float64(CellW-1))
			r1, g1, b1, _ := ebitenImg.At(x, cy).RGBA()
			r2, g2, b2, _ := browserImg.At(x, cy).RGBA()
			fmt.Printf("    x=%-4d ebiten=(%3d,%3d,%3d) browser=(%3d,%3d,%3d)\n",
				x, r1>>8, g1>>8, b1>>8, r2>>8, g2>>8, b2>>8)
		}

		// Sample vertical line through center
		fmt.Println("  Vertical line (x=center):")
		for _, frac := range []float64{0.0, 0.25, 0.5, 0.75, 1.0} {
			y := region.Min.Y + int(frac*float64(CellH-1))
			r1, g1, b1, _ := ebitenImg.At(cx, y).RGBA()
			r2, g2, b2, _ := browserImg.At(cx, y).RGBA()
			fmt.Printf("    y=%-4d ebiten=(%3d,%3d,%3d) browser=(%3d,%3d,%3d)\n",
				y, r1>>8, g1>>8, b1>>8, r2>>8, g2>>8, b2>>8)
		}

		// Overall diff percentage
		diff, total := 0, 0
		for y := region.Min.Y; y < region.Max.Y; y++ {
			for x := region.Min.X; x < region.Max.X; x++ {
				total++
				r1, g1, b1, _ := ebitenImg.At(x, y).RGBA()
				r2, g2, b2, _ := browserImg.At(x, y).RGBA()
				dr := math.Abs(float64(r1>>8) - float64(r2>>8))
				dg := math.Abs(float64(g1>>8) - float64(g2>>8))
				db := math.Abs(float64(b1>>8) - float64(b2>>8))
				delta := (dr + dg + db) / 3.0
				if delta > 10 {
					diff++
				}
			}
		}
		fmt.Printf("  Diff: %.1f%%\n", float64(diff)/float64(total)*100)
	}
}
