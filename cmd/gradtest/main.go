package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

// Minimal gradient test — renders 3 gradients (90, 180, 45 deg) and checks pixels.

func main() {
	ebiten.SetWindowSize(600, 200)
	if err := ebiten.RunGame(&game{}); err != nil && err != ebiten.Termination {
		panic(err)
	}
}

type game struct {
	frame int
	done  bool
}

func (g *game) Update() error {
	g.frame++
	if g.frame == 5 && !g.done {
		g.done = true
		g.captureAndAnalyze()
		return ebiten.Termination
	}
	return nil
}

func (g *game) Layout(_, _ int) (int, int) { return 600, 200 }

func (g *game) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black)

	// Draw 3 gradient rects: 90deg, 180deg, 45deg
	angles := []float64{90, 180, 45}
	stops := []colorStop{
		{color.RGBA{231, 76, 60, 255}, 0},
		{color.RGBA{52, 152, 219, 255}, 1},
	}

	for i, angle := range angles {
		x := float64(i * 200)
		drawTestGradient(screen, x, 0, 180, 130, angle, stops)
	}
}

type colorStop struct {
	c   color.RGBA
	pos float64
}

func drawTestGradient(screen *ebiten.Image, rx, ry, rw, rh, cssAngle float64, stops []colorStop) {
	angleRad := (cssAngle - 90) * math.Pi / 180
	cosA := math.Cos(angleRad)
	sinA := math.Sin(angleRad)
	hw, hh := rw/2, rh/2

	corners := [4][2]float64{{-hw, -hh}, {hw, -hh}, {-hw, hh}, {hw, hh}}
	minDot, maxDot := math.Inf(1), math.Inf(-1)
	for _, c := range corners {
		dot := c[0]*cosA + c[1]*sinA
		if dot < minDot {
			minDot = dot
		}
		if dot > maxDot {
			maxDot = dot
		}
	}
	dotRange := maxDot - minDot
	if dotRange == 0 {
		dotRange = 1
	}

	gradA := cosA / dotRange
	gradB := sinA / dotRange
	gradC := -(hw*cosA + hh*sinA + minDot) / dotRange

	fmt.Printf("Angle=%.0f: cosA=%.6f sinA=%.6f hw=%.1f hh=%.1f\n", cssAngle, cosA, sinA, hw, hh)
	fmt.Printf("  minDot=%.4f maxDot=%.4f dotRange=%.4f\n", minDot, maxDot, dotRange)
	fmt.Printf("  gradA=%.8f gradB=%.8f gradC=%.8f\n", gradA, gradB, gradC)
	fmt.Printf("  t(0,0)=%.6f t(180,0)=%.6f t(0,130)=%.6f t(180,130)=%.6f\n",
		gradA*0+gradB*0+gradC,
		gradA*180+gradB*0+gradC,
		gradA*0+gradB*130+gradC,
		gradA*180+gradB*130+gradC)

	// Build strip
	strip := ebiten.NewImage(256, 1)
	pix := make([]byte, 256*4)
	for x := 0; x < 256; x++ {
		t := float64(x) / 255.0
		var r, g, b uint8
		if t <= stops[0].pos {
			r, g, b = stops[0].c.R, stops[0].c.G, stops[0].c.B
		} else if t >= stops[len(stops)-1].pos {
			r, g, b = stops[len(stops)-1].c.R, stops[len(stops)-1].c.G, stops[len(stops)-1].c.B
		} else {
			f := t
			r = uint8(float64(stops[0].c.R)*(1-f) + float64(stops[1].c.R)*f)
			g = uint8(float64(stops[0].c.G)*(1-f) + float64(stops[1].c.G)*f)
			b = uint8(float64(stops[0].c.B)*(1-f) + float64(stops[1].c.B)*f)
		}
		pix[x*4+0] = r
		pix[x*4+1] = g
		pix[x*4+2] = b
		pix[x*4+3] = 255
	}
	strip.WritePixels(pix)

	// Load shader
	shader, err := ebiten.NewShader([]byte(`//kage:unit pixels

package main

var GradA float
var GradB float
var GradC float

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	t := GradA*dstPos.x + GradB*dstPos.y + GradC
	t = clamp(t, 0.0, 1.0)
	origin := imageSrc0Origin()
	size := imageSrc0Size()
	fx := t * (size.x - 1.0)
	ix := floor(fx)
	frac := fx - ix
	c0 := imageSrc0At(origin + vec2(ix+0.5, 0.5))
	c1 := imageSrc0At(origin + vec2(min(ix+1.5, size.x-0.5), 0.5))
	return mix(c0, c1, frac) * color
}
`))
	if err != nil {
		panic(err)
	}

	w, h := int(rw), int(rh)
	sw, sh := float32(strip.Bounds().Dx()), float32(strip.Bounds().Dy())
	vertices := []ebiten.Vertex{
		{DstX: float32(rx), DstY: float32(ry), SrcX: 0, SrcY: 0, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		{DstX: float32(rx) + float32(w), DstY: float32(ry), SrcX: sw, SrcY: 0, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		{DstX: float32(rx), DstY: float32(ry) + float32(h), SrcX: 0, SrcY: sh, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		{DstX: float32(rx) + float32(w), DstY: float32(ry) + float32(h), SrcX: sw, SrcY: sh, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
	}
	indices := []uint16{0, 1, 2, 1, 3, 2}

	top := &ebiten.DrawTrianglesShaderOptions{}
	top.Images[0] = strip
	top.Uniforms = map[string]any{
		"GradA": float32(gradA),
		"GradB": float32(gradB),
		"GradC": float32(gradC),
	}
	screen.DrawTrianglesShader(vertices, indices, shader, top)
}

func (g *game) captureAndAnalyze() {
	img := ebiten.NewImage(600, 200)
	img.Fill(color.Black)
	g.Draw(img)

	// Check pixels
	labels := []string{"90deg", "180deg", "45deg"}
	for i := 0; i < 3; i++ {
		ox := i * 200
		fmt.Printf("\n=== %s ===\n", labels[i])
		for _, frac := range []float64{0.0, 0.25, 0.5, 0.75, 1.0} {
			x := ox + int(frac*179)
			y := 65 // vertical center
			r, g, b, _ := img.At(x, y).RGBA()
			fmt.Printf("  x=%d,y=%d → (%d,%d,%d)\n", x, y, r>>8, g>>8, b>>8)
		}
		for _, frac := range []float64{0.0, 0.25, 0.5, 0.75, 1.0} {
			x := ox + 90 // horizontal center
			y := int(frac * 129)
			r, g, b, _ := img.At(x, y).RGBA()
			fmt.Printf("  x=%d,y=%d → (%d,%d,%d)\n", x, y, r>>8, g>>8, b>>8)
		}
	}

	// Save for inspection
	rgba := image.NewRGBA(image.Rect(0, 0, 600, 200))
	for y := 0; y < 200; y++ {
		for x := 0; x < 600; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}
	f, _ := os.Create("gradient_test.png")
	png.Encode(f, rgba)
	f.Close()
	fmt.Println("\nSaved gradient_test.png")
}
