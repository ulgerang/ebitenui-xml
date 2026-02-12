package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/ulgerang/ebitenui-xml/ui"
)

func main() {
	xmlData, _ := os.ReadFile("cmd/showcase/layout.xml")
	styleData, _ := os.ReadFile("cmd/showcase/styles.json")

	engine := ui.New(800, 600)
	fontData, _ := os.ReadFile("C:/Windows/Fonts/segoeui.ttf")
	source, _ := text.NewGoTextFaceSource(bytes.NewReader(fontData))
	engine.DefaultFont = source

	engine.LoadLayout(string(xmlData))
	engine.LoadStyles(string(styleData))

	// Force layout
	ui.LayoutWidget(engine.Root())

	dump(engine.Root(), 0)
}

func dump(w ui.Widget, depth int) {
	r := w.ComputedRect()
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}
	fmt.Printf("%s[%s#%s] rect=(%.1f, %.1f, %.1f, %.1f) style_h=%.1f flex_g=%.1f\n",
		indent, w.Type(), w.ID(), r.X, r.Y, r.W, r.H, w.Style().Height, w.Style().FlexGrow)

	for _, child := range w.Children() {
		dump(child, depth+1)
	}
}
