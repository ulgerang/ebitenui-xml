package main

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/ulgerang/ebitenui-xml/ui"
)

const (
	screenWidth  = 800
	screenHeight = 600
)

type Game struct {
	ui       *ui.UI
	fontFace text.Face
}

func NewGame() (*Game, error) {
	g := &Game{}

	// Load font
	fontData, err := os.ReadFile("assets/fonts/NotoSansKR-Regular.ttf")
	if err != nil {
		// Try alternative path
		fontData, err = os.ReadFile("C:/Windows/Fonts/malgun.ttf")
		if err != nil {
			log.Printf("Warning: Could not load font: %v", err)
		}
	}

	if fontData != nil {
		source, err := text.NewGoTextFaceSource(bytes.NewReader(fontData))
		if err == nil {
			g.fontFace = &text.GoTextFace{
				Source: source,
				Size:   16,
			}
		}
	}

	// Create UI
	g.ui = ui.New(screenWidth, screenHeight)

	// IMPORTANT: JSON must use "styles" wrapper object!
	// Also padding must be an object, not a number
	styleJSON := `{
		"styles": {
			"ui": {
				"direction": "column",
				"width": 800,
				"height": 600,
				"background": "#1a1a2e",
				"padding": {"top": 20, "right": 20, "bottom": 20, "left": 20}
			},
			"#header": {
				"height": 60,
				"background": "rgba(0, 0, 0, 0.3)",
				"padding": {"top": 10, "right": 10, "bottom": 10, "left": 10},
				"margin": {"bottom": 20}
			},
			".title": {
				"fontSize": 24,
				"color": "#ffffff"
			},
			".form-panel": {
				"direction": "column",
				"gap": 16,
				"padding": {"top": 20, "right": 20, "bottom": 20, "left": 20},
				"background": "rgba(30, 30, 50, 0.9)",
				"borderRadius": 12,
				"flexGrow": 1
			},
			".form-row": {
				"direction": "row",
				"gap": 12,
				"align": "center",
				"height": 40
			},
			".label": {
				"width": 100,
				"fontSize": 14,
				"color": "#aaaaaa"
			},
			"input": {
				"flexGrow": 1,
				"height": 36,
				"background": "rgba(255, 255, 255, 0.1)",
				"borderRadius": 6,
				"borderWidth": 1,
				"border": "rgba(255, 255, 255, 0.2)"
			},
			"checkbox": {
				"fontSize": 14,
				"color": "#ffffff"
			},
			"slider": {
				"flexGrow": 1,
				"height": 24
			},
			"button": {
				"height": 40,
				"background": "#4169E1",
				"borderRadius": 8,
				"color": "#ffffff",
				"fontSize": 14
			},
			"#submit-btn": {
				"margin": {"top": 20}
			}
		}
	}`

	if err := g.ui.LoadStyles(styleJSON); err != nil {
		log.Printf("Warning: Failed to load styles: %v", err)
	}

	// Load layout
	layoutXML := `<ui id="root">
		<panel id="header">
			<text class="title">Extended UI Demo - Input Widgets Test</text>
		</panel>
		
		<panel class="form-panel">
			<panel class="form-row">
				<text class="label">Username:</text>
				<input id="username" placeholder="Enter your name..." />
			</panel>
			
			<panel class="form-row">
				<text class="label">Password:</text>
				<input id="password" placeholder="Enter password" password="true" />
			</panel>
			
			<panel class="form-row">
				<text class="label">Volume:</text>
				<slider id="volume" value="0.7" min="0" max="1" />
			</panel>
			
			<panel class="form-row">
				<checkbox id="music" checked="true">Enable Music</checkbox>
			</panel>
			
			<panel class="form-row">
				<checkbox id="sfx" checked="true">Enable Sound Effects</checkbox>
			</panel>
			
			<button id="submit-btn">Submit Settings</button>
		</panel>
	</ui>`

	if err := g.ui.LoadLayout(layoutXML); err != nil {
		return nil, fmt.Errorf("failed to load layout: %w", err)
	}

	// Setup widgets
	g.setupWidgets()

	return g, nil
}

func (g *Game) setupWidgets() {
	// Set font for text inputs
	if input := g.ui.GetTextInput("username"); input != nil && g.fontFace != nil {
		input.FontFace = g.fontFace
		input.OnSubmit = func(text string) {
			log.Printf("Username submitted: %s", text)
		}
	}

	if input := g.ui.GetTextInput("password"); input != nil && g.fontFace != nil {
		input.FontFace = g.fontFace
	}

	// Button
	if btn := g.ui.GetButton("submit-btn"); btn != nil {
		btn.OnClick(func() {
			if input := g.ui.GetTextInput("username"); input != nil {
				log.Printf("Submitted username: %s", input.Text)
			}
			if slider := g.ui.GetSlider("volume"); slider != nil {
				log.Printf("Volume: %.1f", slider.Value)
			}
			if cb := g.ui.GetCheckbox("music"); cb != nil {
				log.Printf("Music enabled: %v", cb.Checked)
			}
		})
	}

	// Slider
	if slider := g.ui.GetSlider("volume"); slider != nil {
		slider.OnChange = func(value float64) {
			log.Printf("Volume changed: %.2f", value)
		}
	}

	// Checkboxes
	if cb := g.ui.GetCheckbox("music"); cb != nil {
		cb.OnChange = func(checked bool) {
			log.Printf("Music: %v", checked)
		}
	}
}

func (g *Game) Update() error {
	g.ui.Update()

	// Handle input for focused text input
	if input := g.ui.GetTextInput("username"); input != nil && input.Focused {
		input.HandleInput()
	}
	if input := g.ui.GetTextInput("password"); input != nil && input.Focused {
		input.HandleInput()
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.ui.Draw(screen)
}

func (g *Game) Layout(w, h int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	game, err := NewGame()
	if err != nil {
		log.Fatal(err)
	}

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Ebiten UI - Input Widgets Test")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
