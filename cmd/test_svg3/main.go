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
	uiManager *ui.UI
}

func NewGame() *Game {
	g := &Game{}
	g.uiManager = ui.New(screenWidth, screenHeight)

	// Simple styles
	stylesJSON := `{
		"#root": {
			"background": "#1E1E28",
			"padding": {"top": 50, "right": 50, "bottom": 50, "left": 50},
			"direction": "row",
			"gap": 50
		},
		".icon-box": {
			"width": 100,
			"height": 100,
			"background": "#323246",
			"borderRadius": 10,
			"direction": "column",
			"justify": "center",
			"align": "center"
		},
		"svg": {
			"width": 48,
			"height": 48
		}
	}`

	if err := g.uiManager.LoadStyles(stylesJSON); err != nil {
		log.Printf("Failed to load styles: %v", err)
	}

	// Simple layout - just icons in boxes
	xmlLayout := `
	<panel id="root">
		<panel class="icon-box">
			<icon icon="arrow-left" stroke="#4A90D9" stroke-width="2.5"/>
		</panel>
		<panel class="icon-box">
			<icon icon="check" stroke="#4CAF50" stroke-width="2.5"/>
		</panel>
		<panel class="icon-box">
			<icon icon="heart" stroke="#E91E63" stroke-width="2"/>
		</panel>
		<panel class="icon-box">
			<icon icon="star" stroke="#FFC107" stroke-width="2"/>
		</panel>
	</panel>
	`

	if err := g.uiManager.LoadLayout(xmlLayout); err != nil {
		log.Printf("Failed to load layout: %v", err)
	}

	return g
}

func (g *Game) Update() error {
	g.uiManager.Update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{30, 30, 40, 255})
	g.uiManager.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("SVG Simple Test")

	game := NewGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
