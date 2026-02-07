package ui

import (
	_ "embed"
	"fmt"
	"image/color"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

// ============================================================================
// GPU Shader Management — lazy compilation and gradient strip texture builder
// ============================================================================

//go:embed shaders/gradient_linear.kage
var linearGradientShaderSrc []byte

// gradientStripWidth is the resolution of the 1D gradient lookup texture.
// 256 texels provide 8-bit colour precision, matching CSS rendering quality.
const gradientStripWidth = 256

// --- Linear gradient shader (singleton, lazy-compiled) ----------------------

var (
	linearGradientShader     *ebiten.Shader
	linearGradientShaderOnce sync.Once
)

// getLinearGradientShader returns the compiled Kage shader for linear gradients.
// The shader is compiled once on first use and cached for the process lifetime.
func getLinearGradientShader() *ebiten.Shader {
	linearGradientShaderOnce.Do(func() {
		s, err := ebiten.NewShader(linearGradientShaderSrc)
		if err != nil {
			panic(fmt.Sprintf("ui: failed to compile linear gradient shader: %v", err))
		}
		linearGradientShader = s
	})
	return linearGradientShader
}

// --- Gradient strip texture builder -----------------------------------------

// buildGradientStrip rasterises the given colour stops into a 1×height=1 Ebiten
// image of the specified width.  Each texel at position x corresponds to
// t = x / (width-1) in the gradient.  The result uses premultiplied alpha as
// required by Ebitengine's rendering pipeline.
func buildGradientStrip(stops []ColorStop, width int) *ebiten.Image {
	img := ebiten.NewImage(width, 1)
	pix := make([]byte, width*4)

	for x := 0; x < width; x++ {
		t := float64(x) / float64(width-1)
		c := interpolateGradient(stops, t)

		// Convert to premultiplied RGBA bytes for WritePixels.
		r, g, b, a := c.RGBA() // premultiplied 16-bit
		pix[x*4+0] = uint8(r >> 8)
		pix[x*4+1] = uint8(g >> 8)
		pix[x*4+2] = uint8(b >> 8)
		pix[x*4+3] = uint8(a >> 8)
	}

	img.WritePixels(pix)
	return img
}

// ensureGradientStrip lazily builds and caches the 1D strip texture on the
// Gradient struct.  Thread-safety is not required because rendering runs on the
// single-threaded Ebiten game loop.
func (g *Gradient) ensureGradientStrip() *ebiten.Image {
	if g.strip == nil {
		g.strip = buildGradientStrip(g.ColorStops, gradientStripWidth)
	}
	return g.strip
}

// clearGradientStrip discards the cached strip texture.  Call this if colour
// stops are mutated after initial rendering (uncommon in practice).
func (g *Gradient) clearGradientStrip() {
	if g.strip != nil {
		g.strip.Deallocate()
		g.strip = nil
	}
}

// --- Premultiplied alpha colour helper --------------------------------------

// colorToPremultiplied converts any color.Color to premultiplied RGBA bytes.
// Used for gradient strip WritePixels which expects premultiplied data.
func colorToPremultiplied(c color.Color) (r, g, b, a uint8) {
	cr, cg, cb, ca := c.RGBA() // already premultiplied 16-bit
	return uint8(cr >> 8), uint8(cg >> 8), uint8(cb >> 8), uint8(ca >> 8)
}
