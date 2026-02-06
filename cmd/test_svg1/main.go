package main

import (
	"image/color"
	"log"

	"github.com/ulgerang/ebitenui-xml/ui"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	screenWidth  = 800
	screenHeight = 600
)

type Game struct {
	icons []*ui.SVGIcon
}

func NewGame() *Game {
	g := &Game{}

	// Create some SVG icons directly
	iconNames := []string{"arrow-left", "arrow-right", "check", "x", "heart", "star", "coin", "shield"}
	colors := []color.Color{
		color.RGBA{74, 144, 217, 255}, // arrow-left
		color.RGBA{74, 144, 217, 255}, // arrow-right
		color.RGBA{76, 175, 80, 255},  // check
		color.RGBA{244, 67, 54, 255},  // x
		color.RGBA{233, 30, 99, 255},  // heart
		color.RGBA{255, 193, 7, 255},  // star
		color.RGBA{255, 215, 0, 255},  // coin
		color.RGBA{63, 81, 181, 255},  // shield
	}

	for i, name := range iconNames {
		icon := ui.NewSVGIcon("icon-" + name)
		icon.SetIcon(name, colors[i], 2.5)

		// Calculate position
		row := i / 4
		col := i % 4
		x := 50.0 + float64(col)*100.0
		y := 100.0 + float64(row)*100.0

		// Set computed rect manually
		icon.SetComputedRect(ui.Rect{X: x, Y: y, W: 64, H: 64})

		g.icons = append(g.icons, icon)
	}

	return g
}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{30, 30, 40, 255})

	// Draw each icon
	for _, icon := range g.icons {
		icon.Draw(screen)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("SVG Direct Test")

	game := NewGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
