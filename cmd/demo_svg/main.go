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

	// Create UI manager
	uiManager := ui.New(float64(screenWidth), float64(screenHeight))
	g.uiManager = uiManager

	// Load styles as JSON
	stylesJSON := `{
		"panel": {
			"background": "#1E1E28",
			"padding": {"top": 20, "right": 20, "bottom": 20, "left": 20},
			"direction": "column",
			"gap": 15
		},
		".icon-row": {
			"direction": "row",
			"gap": 20,
			"align": "center",
			"padding": {"top": 15, "right": 15, "bottom": 15, "left": 15},
			"background": "#28283A",
			"borderRadius": 10
		},
		".icon-container": {
			"width": 80,
			"height": 80,
			"background": "#323246",
			"borderRadius": 12,
			"direction": "column",
			"justify": "center",
			"align": "center",
			"gap": 5
		},
		"svg": {
			"width": 32,
			"height": 32
		},
		".icon-label": {
			"fontSize": 12,
			"textAlign": "center"
		},
		".title": {
			"fontSize": 24
		},
		".subtitle": {
			"fontSize": 14
		}
	}`

	if err := uiManager.LoadStyles(stylesJSON); err != nil {
		log.Printf("Failed to load styles: %v", err)
	}

	// Build UI layout
	xmlLayout := `
	<panel id="root">
		<text class="title">SVG Icons Demo</text>
		<text class="subtitle">Built-in icon library for game UI</text>
		
		<!-- Navigation Icons -->
		<panel class="icon-row">
			<panel class="icon-container">
				<icon icon="arrow-left" stroke="#4A90D9" stroke-width="2.5"/>
				<text class="icon-label">Left</text>
			</panel>
			<panel class="icon-container">
				<icon icon="arrow-right" stroke="#4A90D9" stroke-width="2.5"/>
				<text class="icon-label">Right</text>
			</panel>
			<panel class="icon-container">
				<icon icon="arrow-up" stroke="#4A90D9" stroke-width="2.5"/>
				<text class="icon-label">Up</text>
			</panel>
			<panel class="icon-container">
				<icon icon="arrow-down" stroke="#4A90D9" stroke-width="2.5"/>
				<text class="icon-label">Down</text>
			</panel>
		</panel>

		<!-- Action Icons -->
		<panel class="icon-row">
			<panel class="icon-container">
				<icon icon="check" stroke="#4CAF50" stroke-width="2.5"/>
				<text class="icon-label">Check</text>
			</panel>
			<panel class="icon-container">
				<icon icon="x" stroke="#F44336" stroke-width="2.5"/>
				<text class="icon-label">Close</text>
			</panel>
			<panel class="icon-container">
				<icon icon="plus" stroke="#2196F3" stroke-width="2.5"/>
				<text class="icon-label">Add</text>
			</panel>
			<panel class="icon-container">
				<icon icon="minus" stroke="#FF9800" stroke-width="2.5"/>
				<text class="icon-label">Remove</text>
			</panel>
		</panel>

		<!-- UI Icons -->
		<panel class="icon-row">
			<panel class="icon-container">
				<icon icon="menu" stroke="#9C27B0" stroke-width="2.5"/>
				<text class="icon-label">Menu</text>
			</panel>
			<panel class="icon-container">
				<icon icon="search" stroke="#00BCD4" stroke-width="2.5"/>
				<text class="icon-label">Search</text>
			</panel>
			<panel class="icon-container">
				<icon icon="settings" stroke="#607D8B" stroke-width="2.5"/>
				<text class="icon-label">Settings</text>
			</panel>
			<panel class="icon-container">
				<icon icon="home" stroke="#8BC34A" stroke-width="2.5"/>
				<text class="icon-label">Home</text>
			</panel>
		</panel>

		<!-- Game Icons -->
		<panel class="icon-row">
			<panel class="icon-container">
				<icon icon="heart" stroke="#E91E63" stroke-width="2"/>
				<text class="icon-label">Heart</text>
			</panel>
			<panel class="icon-container">
				<icon icon="star" stroke="#FFC107" stroke-width="2"/>
				<text class="icon-label">Star</text>
			</panel>
			<panel class="icon-container">
				<icon icon="shield" stroke="#3F51B5" stroke-width="2"/>
				<text class="icon-label">Shield</text>
			</panel>
			<panel class="icon-container">
				<icon icon="coin" stroke="#FFD700" stroke-width="2"/>
				<text class="icon-label">Coin</text>
			</panel>
		</panel>
	</panel>
	`

	// Parse and build UI
	if err := uiManager.LoadLayout(xmlLayout); err != nil {
		log.Printf("Failed to parse XML: %v", err)
	}

	return g
}

func (g *Game) Update() error {
	g.uiManager.Update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{20, 20, 30, 255})
	g.uiManager.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Ebiten SVG Icons Demo")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
