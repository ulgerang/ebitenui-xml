package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"time"
)

func main() {
	fmt.Println("=== Showcase Comparison Tool ===")

	// 1. Render Ebiten
	fmt.Println(">> Rendering Ebiten showcase...")
	cmd := exec.Command("go", "run", "cmd/showcase/main.go")
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error rendering Ebiten: %v\n", err)
		return
	}
	time.Sleep(1 * time.Second) // Wait for file to settle

	ebitenPath := "cmd/showcase/ebiten_showcase.png"
	browserPath := "cmd/showcase/browser_showcase.png"
	diffPath := "cmd/showcase/diff_showcase.png"

	// 2. Capture Browser (if _screenshot_css.cjs exists and is adapted)
	fmt.Println(">> Capturing Browser reference...")
	// We'll create a temporary screenshot script for the showcase
	js := `
const puppeteer = require('puppeteer');
(async () => {
  const browser = await puppeteer.launch({ headless: 'new', args: ['--no-sandbox'] });
  const page = await browser.newPage();
  await page.setViewport({ width: 800, height: 600, deviceScaleFactor: 1 });
  const url = 'file://' + process.cwd() + '/cmd/showcase/index.html';
  await page.goto(url, { waitUntil: 'networkidle0' });
  await page.screenshot({ path: 'cmd/showcase/browser_showcase.png' });
  await browser.close();
})();`
	os.WriteFile("shot_showcase.cjs", []byte(js), 0644)
	exec.Command("node", "shot_showcase.cjs").Run()
	os.Remove("shot_showcase.cjs")

	// 3. Compare
	fmt.Println(">> Comparing pixels...")
	eImg := loadPNG(ebitenPath)
	bImg := loadPNG(browserPath)

	if eImg == nil || bImg == nil {
		fmt.Println("Failed to load images.")
		return
	}

	diff, diffImg := compareImages(eImg, bImg)
	f, _ := os.Create(diffPath)
	png.Encode(f, diffImg)
	f.Close()

	fmt.Printf("Diff Result: %.2f%%\n", diff*100)
	if diff < 0.01 {
		fmt.Println("SUCCESS: Images match perfectly!")
	} else {
		fmt.Printf("FAILURE: Significant difference found. Check %s\n", diffPath)
	}
}

func loadPNG(path string) image.Image {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	img, _ := png.Decode(f)
	return img
}

func compareImages(img1, img2 image.Image) (float64, image.Image) {
	b1 := img1.Bounds()
	b2 := img2.Bounds()
	w := b1.Dx()
	h := b1.Dy()
	if b2.Dx() < w {
		w = b2.Dx()
	}
	if b2.Dy() < h {
		h = b2.Dy()
	}

	diffCount := 0
	out := image.NewRGBA(image.Rect(0, 0, w, h))

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c1 := img1.At(x, y)
			c2 := img2.At(x, y)
			r1, g1, b1, _ := c1.RGBA()
			r2, g2, b2, _ := c2.RGBA()

			dr := int(r1>>8) - int(r2>>8)
			dg := int(g1>>8) - int(g2>>8)
			db := int(b1>>8) - int(b2>>8)

			if abs(dr) > 5 || abs(dg) > 5 || abs(db) > 5 {
				diffCount++
				out.Set(x, y, color.RGBA{255, 0, 0, 255}) // Red for diff
			} else {
				out.Set(x, y, color.RGBA{uint8(r1 >> 8), uint8(g1 >> 8), uint8(b1 >> 8), 64}) // Faded original
			}
		}
	}

	return float64(diffCount) / float64(w*h), out
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
