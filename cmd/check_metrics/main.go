package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/basicfont"
)

func main() {
	face := text.NewGoXFace(basicfont.Face7x13)
	m := face.Metrics()
	fmt.Printf("%+v\n", m)
}
