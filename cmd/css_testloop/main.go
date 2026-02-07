// css_testloop renders all CSS test cases in a grid using ebitenui-xml,
// captures a screenshot after stabilization, then exits.
//
// Usage:
//
//	go run ./cmd/css_testloop -mode render -out ebiten.png
//	go run ./cmd/css_testloop -mode html   -out reference.html
//	go run ./cmd/css_testloop -mode compare -browser browser.png -ebiten ebiten.png -out report.html
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"

	"github.com/hajimehoshi/bitmapfont/v4"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/ulgerang/ebitenui-xml/ui"
)

var (
	mode       = flag.String("mode", "render", "render|html|compare")
	outPath    = flag.String("out", "ebiten.png", "output file path")
	browserPng = flag.String("browser", "", "browser screenshot for compare mode")
	ebitenPng  = flag.String("ebiten", "", "ebiten screenshot for compare mode")
)

// grid layout constants
const (
	cols    = 5
	padX    = 10
	padY    = 30 // extra space for label
	marginX = 5
	marginY = 5
)

func gridSize(n int) (int, int) {
	rows := int(math.Ceil(float64(n) / float64(cols)))
	w := cols*(CellW+marginX*2) + padX*2
	h := rows*(CellH+marginY*2+padY) + padX*2
	return w, h
}

// CellRegion returns pixel rect for test case index i in the grid.
func CellRegion(i, totalW int) image.Rectangle {
	col := i % cols
	row := i / cols
	x := padX + col*(CellW+marginX*2) + marginX
	y := padX + row*(CellH+marginY*2+padY) + padY + marginY
	return image.Rect(x, y, x+CellW, y+CellH)
}

// ── Ebiten Rendering ──

type Game struct {
	ui       []*ui.UI
	cases    []CSSTestCase
	font     text.Face
	w, h     int
	frames   int
	captured bool
}

func NewGame(cases []CSSTestCase) *Game {
	w, h := gridSize(len(cases))
	font := text.NewGoXFace(bitmapfont.FaceEA)

	g := &Game{
		cases: cases,
		font:  font,
		w:     w,
		h:     h,
	}

	// Build a UI instance per test case
	for _, tc := range cases {
		u := ui.New(float64(CellW), float64(CellH))
		u.DefaultFontFace = font
		if err := u.LoadStyles(tc.Styles); err != nil {
			log.Printf("WARN: %s styles: %v", tc.ID, err)
			continue
		}
		if err := u.LoadLayout(tc.XML); err != nil {
			log.Printf("WARN: %s layout: %v", tc.ID, err)
			continue
		}
		g.ui = append(g.ui, u)
	}
	return g
}

func (g *Game) Update() error {
	g.frames++
	// Wait 10 frames for rendering to stabilize, then capture
	if g.frames == 10 && !g.captured {
		g.captured = true
		g.captureScreenshot()
		return ebiten.Termination
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{18, 18, 30, 255})

	for i, u := range g.ui {
		if u == nil {
			continue
		}
		region := CellRegion(i, g.w)

		// Render test case into a temp image
		tmp := ebiten.NewImage(CellW, CellH)
		tmp.Fill(color.RGBA{18, 18, 30, 255})
		u.Draw(tmp)

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(region.Min.X), float64(region.Min.Y))
		screen.DrawImage(tmp, op)
		tmp.Deallocate()

		// Draw label above cell
		labelOp := &text.DrawOptions{}
		labelOp.GeoM.Translate(float64(region.Min.X), float64(region.Min.Y-14))
		labelOp.ColorScale.ScaleWithColor(color.RGBA{180, 180, 200, 255})
		text.Draw(screen, g.cases[i].ID, g.font, labelOp)
	}
}

func (g *Game) Layout(_, _ int) (int, int) { return g.w, g.h }

func (g *Game) captureScreenshot() {
	img := ebiten.NewImage(g.w, g.h)
	img.Fill(color.RGBA{18, 18, 30, 255})
	g.Draw(img)

	// Convert to standard image
	bounds := image.Rect(0, 0, g.w, g.h)
	rgba := image.NewRGBA(bounds)
	for y := 0; y < g.h; y++ {
		for x := 0; x < g.w; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}

	f, err := os.Create(*outPath)
	if err != nil {
		log.Fatalf("create %s: %v", *outPath, err)
	}
	defer f.Close()
	png.Encode(f, rgba)
	log.Printf("Saved ebiten screenshot: %s (%dx%d)", *outPath, g.w, g.h)
	img.Deallocate()
}

func main() {
	flag.Parse()
	cases := AllTestCases()

	switch *mode {
	case "render":
		game := NewGame(cases)
		ebiten.SetWindowSize(game.w, game.h)
		ebiten.SetWindowTitle("CSS TestLoop - EbitenUI-XML")
		if err := ebiten.RunGame(game); err != nil && err != ebiten.Termination {
			log.Fatal(err)
		}

	case "html":
		if err := GenerateHTML(cases, *outPath); err != nil {
			log.Fatalf("html generation: %v", err)
		}

	case "compare":
		if *browserPng == "" || *ebitenPng == "" {
			fmt.Println("Usage: -mode compare -browser <browser.png> -ebiten <ebiten.png> -out <report.html>")
			os.Exit(1)
		}
		results := CompareImages(*browserPng, *ebitenPng, cases)
		if err := GenerateReport(results, *outPath); err != nil {
			log.Fatalf("report: %v", err)
		}

	default:
		fmt.Printf("Unknown mode: %s (use render|html|compare)\n", *mode)
		os.Exit(1)
	}
}
