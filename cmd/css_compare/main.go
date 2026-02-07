// css_compare runs the ebitenui-xml demo with an ERTP debug server attached,
// enabling automated screenshot capture for CSS comparison testing.
//
// Usage:
//
//	go run ./cmd/css_compare -layout assets/layout.xml -styles assets/styles.json
//
// Once running, the ERTP server listens on :9222 for screenshot requests.
package main

import (
	_ "embed"
	"flag"
	"fmt"
	"image/color"
	"log"
	"os"

	"github.com/hajimehoshi/bitmapfont/v4"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	debug "github.com/ulgerang/ebiten-ertp/debug"
	"github.com/ulgerang/ebitenui-xml/ui"
)

var (
	layoutPath = flag.String("layout", "assets/layout.xml", "Path to layout XML")
	stylesPath = flag.String("styles", "assets/styles.json", "Path to styles JSON")
	port       = flag.String("port", ":9222", "ERTP debug server port")
	width      = flag.Int("width", 640, "Window width")
	height     = flag.Int("height", 480, "Window height")
)

type Game struct {
	ui          *ui.UI
	debugServer *debug.Server
}

func NewGame(layoutXML, stylesJSON string, w, h int) (*Game, error) {
	g := &Game{}

	// Create UI
	g.ui = ui.New(float64(w), float64(h))
	fontData := text.NewGoXFace(bitmapfont.FaceEA)
	g.ui.DefaultFontFace = fontData

	// Load styles
	if err := g.ui.LoadStyles(stylesJSON); err != nil {
		return nil, fmt.Errorf("load styles: %w", err)
	}

	// Load layout
	if err := g.ui.LoadLayout(layoutXML); err != nil {
		return nil, fmt.Errorf("load layout: %w", err)
	}

	// Set up progress bars
	if hp := g.ui.GetProgressBar("hp-bar"); hp != nil {
		hp.FillColor = color.RGBA{76, 175, 80, 255}
	}
	if mp := g.ui.GetProgressBar("mp-bar"); mp != nil {
		mp.FillColor = color.RGBA{33, 150, 243, 255}
	}
	if exp := g.ui.GetProgressBar("exp-bar"); exp != nil {
		exp.FillColor = color.RGBA{255, 193, 7, 255}
	}

	// Create ERTP debug server
	g.debugServer = debug.New()

	return g, nil
}

func (g *Game) Update() error {
	g.ui.Update()
	g.debugServer.UpdateTick()
	g.debugServer.Input.Update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{15, 15, 26, 255})
	g.ui.Draw(screen)
	g.debugServer.CaptureFrame(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return *width, *height
}

func main() {
	flag.Parse()

	// Read files
	layoutXML, err := readFile(*layoutPath)
	if err != nil {
		log.Fatalf("Failed to read layout: %v", err)
	}
	stylesJSON, err := readFile(*stylesPath)
	if err != nil {
		log.Fatalf("Failed to read styles: %v", err)
	}

	game, err := NewGame(layoutXML, stylesJSON, *width, *height)
	if err != nil {
		log.Fatalf("Failed to create game: %v", err)
	}

	// Start ERTP debug server
	if err := game.debugServer.Start(*port); err != nil {
		log.Fatalf("Failed to start debug server: %v", err)
	}
	defer game.debugServer.Stop()

	log.Printf("🎮 CSS Compare harness running (%dx%d)", *width, *height)
	log.Printf("📡 ERTP debug server: http://localhost%s", *port)
	log.Printf("📸 Screenshot: http://localhost%s/screenshot", *port)

	ebiten.SetWindowSize(*width, *height)
	ebiten.SetWindowTitle("CSS Compare - EbitenUI-XML")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

func readFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
