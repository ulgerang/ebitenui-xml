package main

import (
	_ "embed"
	"image/color"
	"log"

	"github.com/ulgerang/ebitenui-xml/ui"
	"github.com/hajimehoshi/bitmapfont/v4"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

//go:embed assets/layout.xml
var layoutXML string

//go:embed assets/styles.json
var stylesJSON string

const (
	screenWidth  = 640
	screenHeight = 480
)

type Game struct {
	ui *ui.UI
}

func NewGame() (*Game, error) {
	g := &Game{}

	// Create UI manager
	g.ui = ui.New(screenWidth, screenHeight)

	// Load font using Ebiten's built-in bitmap font
	fontData := text.NewGoXFace(bitmapfont.FaceEA)
	g.ui.DefaultFontFace = fontData

	// Load styles first
	if err := g.ui.LoadStyles(stylesJSON); err != nil {
		return nil, err
	}

	// Load layout
	if err := g.ui.LoadLayout(layoutXML); err != nil {
		return nil, err
	}

	// Set up event handlers
	g.setupEventHandlers()

	// Set progress bar colors
	g.setupProgressBars()

	return g, nil
}

func (g *Game) setupEventHandlers() {
	// New Game button - with pulse animation
	if btn := g.ui.GetButton("btn-new"); btn != nil {
		btn.OnClick(func() {
			log.Println("New Game clicked!")
			btn.PlayAnimation("pulse")
			if status := g.ui.GetWidget("status"); status != nil {
				if t, ok := status.(*ui.Text); ok {
					t.Content = "Starting new game..."
					t.PlayAnimation("fadeIn")
				}
			}
		})
	}

	// Load Game button - with bounce animation
	if btn := g.ui.GetButton("btn-load"); btn != nil {
		btn.OnClick(func() {
			log.Println("Load Game clicked!")
			btn.PlayAnimation("bounce")
			if status := g.ui.GetWidget("status"); status != nil {
				if t, ok := status.(*ui.Text); ok {
					t.Content = "Loading game..."
				}
			}
		})
	}

	// Settings button - with shake animation
	if btn := g.ui.GetButton("btn-settings"); btn != nil {
		btn.OnClick(func() {
			log.Println("Settings clicked!")
			btn.PlayAnimation("shake")
			if status := g.ui.GetWidget("status"); status != nil {
				if t, ok := status.(*ui.Text); ok {
					t.Content = "Opening settings..."
				}
			}
		})
	}

	// Quit button - with wobble animation
	if btn := g.ui.GetButton("btn-quit"); btn != nil {
		btn.OnClick(func() {
			log.Println("Quit clicked!")
			btn.PlayAnimation("wobble")
			// In a real app, you'd call os.Exit(0) or similar
		})

	}
}

func (g *Game) setupProgressBars() {
	// Health bar - red/green
	if hp := g.ui.GetProgressBar("hp-bar"); hp != nil {
		hp.FillColor = color.RGBA{76, 175, 80, 255} // Green
	}

	// Mana bar - blue
	if mp := g.ui.GetProgressBar("mp-bar"); mp != nil {
		mp.FillColor = color.RGBA{33, 150, 243, 255} // Blue
	}

	// Experience bar - yellow/gold
	if exp := g.ui.GetProgressBar("exp-bar"); exp != nil {
		exp.FillColor = color.RGBA{255, 193, 7, 255} // Gold
	}
}

func (g *Game) Update() error {
	g.ui.Update()

	// Example: animate progress bars
	// if hp := g.ui.GetProgressBar("hp-bar"); hp != nil {
	// 	hp.Value = (hp.Value + 0.001)
	// 	if hp.Value > 1 {
	// 		hp.Value = 0
	// 	}
	// }

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Clear background
	screen.Fill(color.RGBA{15, 15, 26, 255})

	// Draw UI
	g.ui.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Ebiten XML UI Demo")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game, err := NewGame()
	if err != nil {
		log.Fatalf("Failed to create game: %v", err)
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
