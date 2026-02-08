//go:build ignore

package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"os"
)

func main() {
	ebitenImg, err := loadImg("cmd/css_testloop/output/ebiten.png")
	if err != nil {
		log.Fatal(err)
	}
	browserImg, err := loadImg("cmd/css_testloop/output/browser.png")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Index 2 (Vertical Gradient: #2ecc71 to #8e44ad)")
	// Box 2: (435, 45, 615, 175)
	sample(ebitenImg, browserImg, "Top-Center", 435+90, 45+5)
	sample(ebitenImg, browserImg, "Bottom-Center", 435+90, 45+130-5)

	fmt.Println("\nIndex 3 (45deg Gradient: #f39c12 to #2980b9)")
	// Box 3: (645, 45, 825, 175)
	sample(ebitenImg, browserImg, "Top-Left", 645+5, 45+5)
	sample(ebitenImg, browserImg, "Top-Right", 645+180-5, 45+5)
	sample(ebitenImg, browserImg, "Bottom-Left", 645+5, 45+130-5)
	sample(ebitenImg, browserImg, "Bottom-Right", 645+180-5, 45+130-5)
}

func loadImg(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	return img, err
}

func sample(ebitenImg, browserImg image.Image, label string, x, y int) {
	eC := ebitenImg.At(x, y)
	bC := browserImg.At(x, y)
	fmt.Printf("%s at (%d, %d):\n", label, x, y)
	fmt.Printf("  Ebiten:  %v\n", colorToHex(eC))
	fmt.Printf("  Browser: %v\n", colorToHex(bC))
}

func colorToHex(c color.Color) string {
	r, g, b, a := c.RGBA()
	return fmt.Sprintf("#%02x%02x%02x (a:%02x)", r>>8, g>>8, b>>8, a>>8)
}
