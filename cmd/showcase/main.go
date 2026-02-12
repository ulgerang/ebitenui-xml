package main

import (
	"bytes"
	"image"
	"image/png"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/ulgerang/ebitenui-xml/ui"
)

const (
	showcaseWidth  = 960
	showcaseHeight = 640
	outputPNG      = "cmd/showcase/ebiten_showcase.png"
	layoutPath     = "cmd/showcase/layout.xml"
	stylePath      = "cmd/showcase/styles.json"
)

type game struct {
	engine         *ui.UI
	frames         int
	captured       bool
	capturePending bool
}

func (g *game) Update() error {
	g.engine.Update()
	g.frames++
	if g.frames >= 10 && !g.capturePending && !g.captured {
		g.capturePending = true
	}
	if g.captured {
		return ebiten.Termination
	}
	return nil
}

func (g *game) Draw(screen *ebiten.Image) {
	g.engine.Draw(screen)

	if g.capturePending && !g.captured {
		if err := saveImage(screen, outputPNG, showcaseWidth, showcaseHeight); err != nil {
			log.Printf("capture error: %v", err)
		} else {
			log.Printf("Saved %s", outputPNG)
		}
		g.captured = true
		g.capturePending = false
	}
}

func (g *game) Layout(_, _ int) (int, int) {
	return showcaseWidth, showcaseHeight
}

func saveImage(src *ebiten.Image, path string, w, h int) error {
	rgba := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			rgba.Set(x, y, src.At(x, y))
		}
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return png.Encode(f, rgba)
}

func loadFace(paths []string) *text.GoTextFaceSource {
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		src, err := text.NewGoTextFaceSource(bytes.NewReader(data))
		if err == nil {
			return src
		}
	}
	return nil
}

func main() {
	layoutXML, err := os.ReadFile(layoutPath)
	if err != nil {
		log.Fatalf("read layout: %v", err)
	}
	styleJSON, err := os.ReadFile(stylePath)
	if err != nil {
		log.Fatalf("read styles: %v", err)
	}

	engine := ui.New(showcaseWidth, showcaseHeight)
	engine.DefaultFont = loadFace([]string{
		"C:/Windows/Fonts/segoeui.ttf",
		"C:/Windows/Fonts/arial.ttf",
		"C:/Windows/Fonts/malgun.ttf",
	})
	engine.DefaultBoldFont = loadFace([]string{
		"C:/Windows/Fonts/segoeuib.ttf",
		"C:/Windows/Fonts/arialbd.ttf",
		"C:/Windows/Fonts/malgunbd.ttf",
	})

	if err := engine.LoadLayout(string(layoutXML)); err != nil {
		log.Fatalf("load layout: %v", err)
	}
	if err := engine.LoadStyles(string(styleJSON)); err != nil {
		log.Fatalf("load styles: %v", err)
	}

	ebiten.SetWindowSize(showcaseWidth, showcaseHeight)
	ebiten.SetWindowTitle("Showcase Renderer")
	if err := ebiten.RunGame(&game{engine: engine}); err != nil && err != ebiten.Termination {
		log.Fatal(err)
	}
}
