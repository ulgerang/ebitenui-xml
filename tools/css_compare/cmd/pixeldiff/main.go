package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: pixeldiff <img1.png> <img2.png> <diff_output.png>")
		os.Exit(1)
	}

	img1 := loadPNG(os.Args[1])
	img2 := loadPNG(os.Args[2])

	w := minInt(img1.Bounds().Dx(), img2.Bounds().Dx())
	h := minInt(img1.Bounds().Dy(), img2.Bounds().Dy())

	diff := image.NewRGBA(image.Rect(0, 0, w, h))
	totalPixels := w * h
	diffPixels := 0
	totalDelta := 0.0

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r1, g1, b1, _ := img1.At(x, y).RGBA()
			r2, g2, b2, _ := img2.At(x, y).RGBA()

			dr := absInt(int(r1>>8) - int(r2>>8))
			dg := absInt(int(g1>>8) - int(g2>>8))
			db := absInt(int(b1>>8) - int(b2>>8))

			delta := float64(dr+dg+db) / 3.0
			totalDelta += delta

			if delta > 10 { // threshold
				diffPixels++
				intensity := uint8(minInt(255, int(delta*3)))
				diff.Set(x, y, color.RGBA{intensity, 0, intensity, 255})
			} else {
				// Dim original
				r, g, b, _ := img1.At(x, y).RGBA()
				diff.Set(x, y, color.RGBA{uint8(r >> 9), uint8(g >> 9), uint8(b >> 9), 255})
			}
		}
	}

	savePNG(os.Args[3], diff)

	pct := float64(diffPixels) / float64(totalPixels) * 100
	avgDelta := totalDelta / float64(totalPixels)
	fmt.Printf("DIFF_PIXELS=%d\n", diffPixels)
	fmt.Printf("TOTAL_PIXELS=%d\n", totalPixels)
	fmt.Printf("DIFF_PCT=%.2f\n", pct)
	fmt.Printf("AVG_DELTA=%.2f\n", avgDelta)
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func loadPNG(path string) image.Image {
	f, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening %s: %v\n", path, err)
		os.Exit(1)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding %s: %v\n", path, err)
		os.Exit(1)
	}
	return img
}

func savePNG(path string, img image.Image) {
	f, err := os.Create(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating %s: %v\n", path, err)
		os.Exit(1)
	}
	defer f.Close()
	png.Encode(f, img)
}
