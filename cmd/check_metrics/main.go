package main

import (
	"fmt"
	"github.com/hajimehoshi/bitmapfont/v4"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

func main() {
	face := text.NewGoXFace(bitmapfont.FaceEA)
	m := face.Metrics()
	fmt.Printf("%+v\n", m)
}
