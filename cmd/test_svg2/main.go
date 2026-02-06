package main

import (
	"fmt"
	"image/color"
	"log"

	"github.com/ulgerang/ebitenui-xml/ui"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 800
	screenHeight = 600
)

type Game struct {
}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{30, 30, 40, 255})

	// Test 1: Draw simple paths directly with vector at different positions
	colors := []color.Color{
		color.RGBA{255, 0, 0, 255},
		color.RGBA{0, 255, 0, 255},
		color.RGBA{0, 0, 255, 255},
		color.RGBA{255, 255, 0, 255},
	}

	for i := 0; i < 4; i++ {
		x := float32(100 + i*150)
		y := float32(100)

		// Draw a simple arrow
		var path vector.Path
		path.MoveTo(x, y+20)
		path.LineTo(x+30, y+20)
		path.LineTo(x+30, y+10)
		path.LineTo(x+50, y+25)
		path.LineTo(x+30, y+40)
		path.LineTo(x+30, y+30)
		path.LineTo(x, y+30)
		path.Close()

		op := &vector.StrokeOptions{Width: 2}
		vs, is := path.AppendVerticesAndIndicesForStroke(nil, nil, op)

		// Apply color
		r, g, b, a := colors[i].RGBA()
		for j := range vs {
			vs[j].ColorR = float32(r>>8) / 255
			vs[j].ColorG = float32(g>>8) / 255
			vs[j].ColorB = float32(b>>8) / 255
			vs[j].ColorA = float32(a>>8) / 255
		}

		screen.DrawTriangles(vs, is, whiteImage, &ebiten.DrawTrianglesOptions{AntiAlias: true})
	}

	// Test 2: Draw SVG icons using the UI framework at different positions
	iconNames := []string{"arrow-left", "check", "heart", "star"}
	iconColors := []color.Color{
		color.RGBA{255, 100, 100, 255},
		color.RGBA{100, 255, 100, 255},
		color.RGBA{255, 100, 200, 255},
		color.RGBA{255, 200, 100, 255},
	}

	for i, name := range iconNames {
		icon := ui.NewSVGIcon("test-" + name)
		icon.SetIcon(name, iconColors[i], 2.5)

		// Set computed rect (where the icon should appear)
		x := 100.0 + float64(i)*150.0
		y := 250.0
		icon.SetComputedRect(ui.Rect{X: x, Y: y, W: 64, H: 64})

		// Print the rect for debugging
		r := icon.ComputedRect()
		fmt.Printf("Icon %s: rect=(%f, %f, %f, %f)\n", name, r.X, r.Y, r.W, r.H)

		icon.Draw(screen)
	}

	// Test 3: Draw SVGDocument directly
	doc := ui.CreateIconSVG("coin", 24, color.RGBA{255, 215, 0, 255}, 2.5)
	if doc != nil {
		doc.Draw(screen, 100, 400, 64, 64)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

// whiteImage is needed for drawing triangles
var whiteImage = func() *ebiten.Image {
	img := ebiten.NewImage(1, 1)
	img.Fill(color.White)
	return img
}()

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("SVG Position Test")

	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
