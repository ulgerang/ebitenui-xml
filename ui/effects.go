package ui

import (
	"image/color"
	"math"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// BoxShadow represents a CSS-like box shadow
type BoxShadow struct {
	OffsetX float64
	OffsetY float64
	Blur    float64
	Spread  float64
	Color   color.Color
	Inset   bool
}

// Transform represents CSS-like 2D transforms
type Transform struct {
	TranslateX float64
	TranslateY float64
	ScaleX     float64
	ScaleY     float64
	Rotate     float64 // in degrees
	OriginX    float64 // 0-1, where 0.5 is center
	OriginY    float64
}

// NewTransform creates a default transform (identity)
func NewTransform() *Transform {
	return &Transform{
		ScaleX:  1,
		ScaleY:  1,
		OriginX: 0.5,
		OriginY: 0.5,
	}
}

// Transition represents CSS-like transition properties
type Transition struct {
	Property string
	Duration float64 // in seconds
	Easing   EasingFunc
	Delay    float64
}

// EasingFunc is a function that maps t in [0,1] to output [0,1]
type EasingFunc func(t float64) float64

// Easing functions
var (
	EaseLinear = func(t float64) float64 { return t }

	EaseInQuad = func(t float64) float64 { return t * t }

	EaseOutQuad = func(t float64) float64 { return t * (2 - t) }

	EaseInOutQuad = func(t float64) float64 {
		if t < 0.5 {
			return 2 * t * t
		}
		return -1 + (4-2*t)*t
	}

	EaseInCubic = func(t float64) float64 { return t * t * t }

	EaseOutCubic = func(t float64) float64 {
		t--
		return t*t*t + 1
	}

	EaseInOutCubic = func(t float64) float64 {
		if t < 0.5 {
			return 4 * t * t * t
		}
		return (t-1)*(2*t-2)*(2*t-2) + 1
	}

	EaseOutElastic = func(t float64) float64 {
		if t == 0 || t == 1 {
			return t
		}
		return math.Pow(2, -10*t)*math.Sin((t*10-0.75)*(2*math.Pi)/3) + 1
	}

	EaseOutBounce = func(t float64) float64 {
		n1 := 7.5625
		d1 := 2.75
		if t < 1/d1 {
			return n1 * t * t
		} else if t < 2/d1 {
			t -= 1.5 / d1
			return n1*t*t + 0.75
		} else if t < 2.5/d1 {
			t -= 2.25 / d1
			return n1*t*t + 0.9375
		}
		t -= 2.625 / d1
		return n1*t*t + 0.984375
	}
)

// ParseEasing parses an easing function name
func ParseEasing(name string) EasingFunc {
	switch strings.ToLower(name) {
	case "linear":
		return EaseLinear
	case "ease":
		return EaseCSSEase
	case "ease-in", "easein":
		return EaseCSSEaseIn
	case "ease-out", "easeout":
		return EaseCSSEaseOut
	case "ease-in-out", "easeinout":
		return EaseCSSEaseInOut
	case "ease-in-cubic":
		return EaseInCubic
	case "ease-out-cubic":
		return EaseOutCubic
	case "ease-in-out-cubic":
		return EaseInOutCubic
	case "elastic":
		return EaseOutElastic
	case "bounce":
		return EaseOutBounce
	default:
		return EaseLinear
	}
}

// TextOverflow defines text overflow behavior
type TextOverflow int

const (
	TextOverflowVisible TextOverflow = iota
	TextOverflowClip
	TextOverflowEllipsis
)

// WhiteSpace defines white-space handling (like CSS)
type WhiteSpace int

const (
	WhiteSpaceNormal  WhiteSpace = iota // wrap text
	WhiteSpaceNowrap                    // no wrapping
	WhiteSpacePre                       // preserve whitespace
	WhiteSpacePreWrap                   // preserve + wrap
)

// Overflow defines overflow behavior
type Overflow int

const (
	OverflowVisible Overflow = iota
	OverflowHidden
	OverflowScroll
	OverflowAuto
)

// Position defines positioning mode (like CSS)
type Position int

const (
	PositionRelative Position = iota
	PositionAbsolute
	PositionFixed
)

// Display defines display mode
type Display int

const (
	DisplayBlock Display = iota
	DisplayFlex
	DisplayNone
	DisplayInline
)

// Extended CSS-like gradient support
type Gradient struct {
	Type       GradientType
	Angle      float64 // for linear gradients, in degrees
	ColorStops []ColorStop
}

type GradientType int

const (
	GradientLinear GradientType = iota
	GradientRadial
)

type ColorStop struct {
	Color    color.Color
	Position float64 // 0-1
}

// DrawGradient draws a linear gradient in the given rectangle with angle support.
// CSS angles: 0deg=bottom-to-top, 90deg=left-to-right, 180deg=top-to-bottom, 270deg=right-to-left.
func DrawGradient(screen *ebiten.Image, r Rect, g *Gradient) {
	if g == nil || len(g.ColorStops) < 2 {
		return
	}

	// CSS angle to math direction vector.
	// CSS: 0deg points upward (bottom-to-top), 90deg points right.
	// Math: we need the gradient-line direction in screen coords (Y-down).
	// angleRad = (cssAngle - 90) * pi/180 gives us: cos=dx, sin=dy of gradient direction.
	angleRad := (g.Angle - 90) * math.Pi / 180
	cosA := math.Cos(angleRad)
	sinA := math.Sin(angleRad)

	// Half-dimensions for center-relative coordinates
	hw, hh := r.W/2, r.H/2

	// Project all four corners onto the gradient direction to find the
	// full extent of the gradient line across the rectangle.
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

	// Normalize angle to [0, 360) for fast-path checks
	normAngle := math.Mod(g.Angle, 360)
	if normAngle < 0 {
		normAngle += 360
	}

	if normAngle == 90 || normAngle == 270 {
		// Horizontal gradient — draw vertical strips (1px wide, full height)
		for px := 0.0; px < r.W; px++ {
			rx := px - hw
			dot := rx * cosA
			t := (dot - minDot) / dotRange
			if t < 0 {
				t = 0
			}
			if t > 1 {
				t = 1
			}
			clr := interpolateGradient(g.ColorStops, t)
			vector.DrawFilledRect(screen, float32(r.X+px), float32(r.Y), 1, float32(r.H), clr, false)
		}
	} else if normAngle == 0 || normAngle == 180 {
		// Vertical gradient — draw horizontal strips (full width, 1px tall)
		for py := 0.0; py < r.H; py++ {
			ry := py - hh
			dot := ry * sinA
			t := (dot - minDot) / dotRange
			if t < 0 {
				t = 0
			}
			if t > 1 {
				t = 1
			}
			clr := interpolateGradient(g.ColorStops, t)
			vector.DrawFilledRect(screen, float32(r.X), float32(r.Y+py), float32(r.W), 1, clr, false)
		}
	} else {
		// Arbitrary angle — pixel-by-pixel rendering
		for py := 0.0; py < r.H; py++ {
			for px := 0.0; px < r.W; px++ {
				rx := px - hw
				ry := py - hh
				dot := rx*cosA + ry*sinA
				t := (dot - minDot) / dotRange
				if t < 0 {
					t = 0
				}
				if t > 1 {
					t = 1
				}
				clr := interpolateGradient(g.ColorStops, t)
				vector.DrawFilledRect(screen, float32(r.X+px), float32(r.Y+py), 1, 1, clr, false)
			}
		}
	}
}

// interpolateGradient finds the color at position t
func interpolateGradient(stops []ColorStop, t float64) color.Color {
	if t <= stops[0].Position {
		return stops[0].Color
	}
	if t >= stops[len(stops)-1].Position {
		return stops[len(stops)-1].Color
	}

	// Find surrounding stops
	for i := 0; i < len(stops)-1; i++ {
		if t >= stops[i].Position && t <= stops[i+1].Position {
			// Interpolate between stops[i] and stops[i+1]
			range_ := stops[i+1].Position - stops[i].Position
			localT := (t - stops[i].Position) / range_
			return lerpColor(stops[i].Color, stops[i+1].Color, localT)
		}
	}

	return stops[0].Color
}

// lerpColor linearly interpolates between two colors
func lerpColor(c1, c2 color.Color, t float64) color.Color {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()

	return color.RGBA{
		R: uint8((float64(r1>>8)*(1-t) + float64(r2>>8)*t)),
		G: uint8((float64(g1>>8)*(1-t) + float64(g2>>8)*t)),
		B: uint8((float64(b1>>8)*(1-t) + float64(b2>>8)*t)),
		A: uint8((float64(a1>>8)*(1-t) + float64(a2>>8)*t)),
	}
}

// ParseGradient parses a CSS-like gradient string
// e.g., "linear-gradient(90deg, #ff0000, #0000ff)"
func ParseGradient(s string) *Gradient {
	s = strings.TrimSpace(s)

	if strings.HasPrefix(s, "linear-gradient(") {
		return parseLinearGradient(s)
	}

	return nil
}

func parseLinearGradient(s string) *Gradient {
	// Remove "linear-gradient(" prefix and ")" suffix
	s = strings.TrimPrefix(s, "linear-gradient(")
	s = strings.TrimSuffix(s, ")")

	parts := strings.Split(s, ",")
	if len(parts) < 2 {
		return nil
	}

	g := &Gradient{
		Type:       GradientLinear,
		Angle:      90, // default to horizontal
		ColorStops: make([]ColorStop, 0),
	}

	startIdx := 0
	// Check if first part is an angle
	first := strings.TrimSpace(parts[0])
	if strings.HasSuffix(first, "deg") {
		angle := strings.TrimSuffix(first, "deg")
		g.Angle = parseFloatValue(angle)
		startIdx = 1
	}

	// Parse color stops
	numColors := len(parts) - startIdx
	for i := startIdx; i < len(parts); i++ {
		colorStr := strings.TrimSpace(parts[i])
		clr := parseColor(colorStr)
		if clr != nil {
			stopPos := float64(i-startIdx) / float64(numColors-1)
			g.ColorStops = append(g.ColorStops, ColorStop{
				Color:    clr,
				Position: stopPos,
			})
		}
	}

	if len(g.ColorStops) < 2 {
		return nil
	}

	return g
}

func parseFloatValue(s string) float64 {
	s = strings.TrimSpace(s)
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// DrawRoundedRectPath draws a rounded rectangle with the same radius on all corners.
func DrawRoundedRectPath(screen *ebiten.Image, r Rect, radius float64, clr color.Color) {
	DrawRoundedRectPathEx(screen, r, radius, radius, radius, radius, clr)
}

// DrawRoundedRectPathEx draws a filled rounded rectangle with independent per-corner radii.
// radTL = top-left, radTR = top-right, radBR = bottom-right, radBL = bottom-left.
func DrawRoundedRectPathEx(screen *ebiten.Image, r Rect, radTL, radTR, radBR, radBL float64, clr color.Color) {
	// Fast path: no rounding at all
	if radTL <= 0 && radTR <= 0 && radBR <= 0 && radBL <= 0 {
		vector.DrawFilledRect(screen,
			float32(r.X), float32(r.Y),
			float32(r.W), float32(r.H),
			clr, false)
		return
	}

	// Clamp each radius to half the smallest dimension
	maxRadius := min(r.W, r.H) / 2
	if radTL > maxRadius {
		radTL = maxRadius
	}
	if radTR > maxRadius {
		radTR = maxRadius
	}
	if radBR > maxRadius {
		radBR = maxRadius
	}
	if radBL > maxRadius {
		radBL = maxRadius
	}

	path := &vector.Path{}
	x, y := float32(r.X), float32(r.Y)
	w, h := float32(r.W), float32(r.H)
	rTL := float32(radTL)
	rTR := float32(radTR)
	rBR := float32(radBR)
	rBL := float32(radBL)

	// Start after top-left curve
	path.MoveTo(x+rTL, y)

	// Top edge → top-right corner
	path.LineTo(x+w-rTR, y)
	path.QuadTo(x+w, y, x+w, y+rTR)

	// Right edge → bottom-right corner
	path.LineTo(x+w, y+h-rBR)
	path.QuadTo(x+w, y+h, x+w-rBR, y+h)

	// Bottom edge → bottom-left corner
	path.LineTo(x+rBL, y+h)
	path.QuadTo(x, y+h, x, y+h-rBL)

	// Left edge → top-left corner
	path.LineTo(x, y+rTL)
	path.QuadTo(x, y, x+rTL, y)

	path.Close()

	// Fill the path
	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)

	cr, cg, cb, ca := clr.RGBA()
	for i := range vs {
		vs[i].ColorR = float32(cr) / 0xffff
		vs[i].ColorG = float32(cg) / 0xffff
		vs[i].ColorB = float32(cb) / 0xffff
		vs[i].ColorA = float32(ca) / 0xffff
	}

	op := &ebiten.DrawTrianglesOptions{}
	op.AntiAlias = true
	screen.DrawTriangles(vs, is, whiteImage, op)
}

// whiteImage is a 1x1 white image used for drawing colored shapes
var whiteImage = func() *ebiten.Image {
	img := ebiten.NewImage(1, 1)
	img.Fill(color.White)
	return img
}()

// min returns the smaller of two float64 values
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// ============================================================================
// Modern CSS Effects - Box Shadow, Text Shadow, Outline, Radial Gradient, etc.
// ============================================================================

// DrawBoxShadow draws a CSS-compliant box shadow using Gaussian blur via
// three-pass box blur approximation.  It rasterizes a rounded-rect alpha mask
// into an offscreen buffer, blurs it, multiplies by the shadow colour, and
// composites the result onto screen with premultiplied-alpha source-over.
func DrawBoxShadow(screen *ebiten.Image, r Rect, shadow *BoxShadow, borderRadius float64) {
	if shadow == nil {
		return
	}

	spread := shadow.Spread
	offsetX := shadow.OffsetX
	offsetY := shadow.OffsetY

	// Shadow shape dimensions (expanded by spread on each side).
	shapeW := r.W + 2*spread
	shapeH := r.H + 2*spread
	if shapeW <= 0 || shapeH <= 0 {
		return
	}
	shapeRadius := borderRadius + spread
	maxRad := math.Min(shapeW, shapeH) / 2
	if shapeRadius > maxRad {
		shapeRadius = maxRad
	}
	if shapeRadius < 0 {
		shapeRadius = 0
	}

	// CSS blur-radius → Gaussian sigma.
	sigma := shadow.Blur / 2.0
	if sigma < 0 {
		sigma = 0
	}

	// 3·sigma covers 99.7% of Gaussian energy.
	padding := int(math.Ceil(3 * sigma))
	if padding < 0 {
		padding = 0
	}

	bufW := int(math.Ceil(shapeW)) + 2*padding
	bufH := int(math.Ceil(shapeH)) + 2*padding
	if bufW <= 0 || bufH <= 0 {
		return
	}

	// Where the buffer's (0,0) falls on screen.
	screenX := r.X + offsetX - spread - float64(padding)
	screenY := r.Y + offsetY - spread - float64(padding)

	// ── 1. Rasterize rounded-rect alpha mask via SDF ──────────────
	alphaMap := make([]float64, bufW*bufH)

	hw := shapeW / 2
	hh := shapeH / 2
	cx := float64(padding) + hw
	cy := float64(padding) + hh
	innerHW := hw - shapeRadius
	innerHH := hh - shapeRadius

	for py := 0; py < bufH; py++ {
		for px := 0; px < bufW; px++ {
			fx := float64(px) + 0.5
			fy := float64(py) + 0.5
			dx := math.Abs(fx-cx) - innerHW
			dy := math.Abs(fy-cy) - innerHH
			if dx < 0 {
				dx = 0
			}
			if dy < 0 {
				dy = 0
			}
			dist := math.Sqrt(dx*dx+dy*dy) - shapeRadius
			if dist <= -0.5 {
				alphaMap[py*bufW+px] = 1.0
			} else if dist >= 0.5 {
				alphaMap[py*bufW+px] = 0.0
			} else {
				alphaMap[py*bufW+px] = 0.5 - dist
			}
		}
	}

	// ── 2. Three-pass separable box blur (≈ Gaussian) ─────────────
	if sigma > 0.5 {
		tmp := make([]float64, bufW*bufH)
		sizes := shadowBoxBlurSizes(sigma, 3)
		for _, sz := range sizes {
			rad := (sz - 1) / 2
			shadowBoxBlurH(alphaMap, tmp, bufW, bufH, rad)
			shadowBoxBlurV(tmp, alphaMap, bufW, bufH, rad)
		}
	}

	// ── 3. Convert to premultiplied-alpha RGBA pixels ─────────────
	sr, sg, sb, sa := shadow.Color.RGBA()
	colR := float64(sr) / 0xffff
	colG := float64(sg) / 0xffff
	colB := float64(sb) / 0xffff
	colA := float64(sa) / 0xffff

	pixels := make([]byte, bufW*bufH*4)
	for i := 0; i < bufW*bufH; i++ {
		m := alphaMap[i]
		pr := colR * m
		pg := colG * m
		pb := colB * m
		pa := colA * m
		if pr > 1 {
			pr = 1
		}
		if pg > 1 {
			pg = 1
		}
		if pb > 1 {
			pb = 1
		}
		if pa > 1 {
			pa = 1
		}
		pixels[i*4+0] = uint8(pr*255 + 0.5)
		pixels[i*4+1] = uint8(pg*255 + 0.5)
		pixels[i*4+2] = uint8(pb*255 + 0.5)
		pixels[i*4+3] = uint8(pa*255 + 0.5)
	}

	// ── 4. Upload and composite ───────────────────────────────────
	shadowImg := ebiten.NewImage(bufW, bufH)
	shadowImg.WritePixels(pixels)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(screenX, screenY)
	screen.DrawImage(shadowImg, op)
	shadowImg.Deallocate()
}

// shadowBoxBlurSizes returns n box widths (odd integers) whose iterated
// convolution best approximates a Gaussian with the given sigma.
// Reference: "Fast Almost-Gaussian Filtering" (W3C CSS filter specification).
func shadowBoxBlurSizes(sigma float64, n int) []int {
	wIdeal := math.Sqrt(12.0*sigma*sigma/float64(n) + 1.0)
	wl := int(math.Floor(wIdeal))
	if wl%2 == 0 {
		wl--
	}
	if wl < 1 {
		wl = 1
	}
	wu := wl + 2
	mIdeal := (12.0*sigma*sigma - float64(n)*float64(wl*wl) -
		4.0*float64(n)*float64(wl) - 3.0*float64(n)) /
		(-4.0*float64(wl) - 4.0)
	m := int(math.Round(mIdeal))
	if m < 0 {
		m = 0
	}
	if m > n {
		m = n
	}
	sizes := make([]int, n)
	for i := 0; i < n; i++ {
		if i < m {
			sizes[i] = wl
		} else {
			sizes[i] = wu
		}
	}
	return sizes
}

// shadowBoxBlurH performs a horizontal box blur pass.
// Edge pixels are clamped (extended-border policy) to avoid darkening
// at image borders.
func shadowBoxBlurH(src, dst []float64, w, h, radius int) {
	if radius <= 0 {
		copy(dst, src)
		return
	}
	invDiam := 1.0 / float64(2*radius+1)
	for y := 0; y < h; y++ {
		row := y * w
		var sum float64
		for x := -radius; x <= radius; x++ {
			xi := x
			if xi < 0 {
				xi = 0
			} else if xi >= w {
				xi = w - 1
			}
			sum += src[row+xi]
		}
		dst[row] = sum * invDiam
		for x := 1; x < w; x++ {
			addIdx := x + radius
			if addIdx >= w {
				addIdx = w - 1
			}
			subIdx := x - radius - 1
			if subIdx < 0 {
				subIdx = 0
			}
			sum += src[row+addIdx] - src[row+subIdx]
			dst[row+x] = sum * invDiam
		}
	}
}

// shadowBoxBlurV performs a vertical box blur pass with clamped edges.
func shadowBoxBlurV(src, dst []float64, w, h, radius int) {
	if radius <= 0 {
		copy(dst, src)
		return
	}
	invDiam := 1.0 / float64(2*radius+1)
	for x := 0; x < w; x++ {
		var sum float64
		for y := -radius; y <= radius; y++ {
			yi := y
			if yi < 0 {
				yi = 0
			} else if yi >= h {
				yi = h - 1
			}
			sum += src[yi*w+x]
		}
		dst[x] = sum * invDiam
		for y := 1; y < h; y++ {
			addIdx := y + radius
			if addIdx >= h {
				addIdx = h - 1
			}
			subIdx := y - radius - 1
			if subIdx < 0 {
				subIdx = 0
			}
			sum += src[addIdx*w+x] - src[subIdx*w+x]
			dst[y*w+x] = sum * invDiam
		}
	}
}

// TextShadow represents a CSS-like text shadow
type TextShadow struct {
	OffsetX float64
	OffsetY float64
	Blur    float64
	Color   color.Color
}

// Outline represents a CSS-like outline
type Outline struct {
	Width  float64
	Style  string // "solid", "dashed", "dotted"
	Color  color.Color
	Offset float64 // distance from border
}

// DrawOutline draws an outline around a rectangle
func DrawOutline(screen *ebiten.Image, r Rect, outline *Outline, borderRadius float64) {
	if outline == nil || outline.Width <= 0 {
		return
	}

	// Outline is drawn outside the border
	outlineRect := Rect{
		X: r.X - outline.Width - outline.Offset,
		Y: r.Y - outline.Width - outline.Offset,
		W: r.W + (outline.Width+outline.Offset)*2,
		H: r.H + (outline.Width+outline.Offset)*2,
	}

	drawRoundedRectStroke(screen, outlineRect, borderRadius+outline.Offset, outline.Width, outline.Color)
}

// drawRoundedRectStroke draws a stroked rounded rectangle with uniform radius.
func drawRoundedRectStroke(screen *ebiten.Image, r Rect, radius float64, strokeWidth float64, clr color.Color) {
	drawRoundedRectStrokeEx(screen, r, radius, radius, radius, radius, strokeWidth, clr)
}

// DrawRoundedRectStrokeEx draws a stroked rounded rectangle with per-corner radii (exported).
func DrawRoundedRectStrokeEx(screen *ebiten.Image, r Rect, radTL, radTR, radBR, radBL float64, strokeWidth float64, clr color.Color) {
	drawRoundedRectStrokeEx(screen, r, radTL, radTR, radBR, radBL, strokeWidth, clr)
}

// drawRoundedRectStrokeEx is the internal implementation for stroked rounded rects
// with independent per-corner radii.
//
// CSS border-box: the border sits entirely INSIDE the element rect.
// Ebiten stroke is centered on the path, so we inset the path by strokeWidth/2.
// Border radii are also reduced by strokeWidth/2 so the outer edge of the
// stroke matches the CSS border-radius.
func drawRoundedRectStrokeEx(screen *ebiten.Image, r Rect, radTL, radTR, radBR, radBL float64, strokeWidth float64, clr color.Color) {
	if strokeWidth <= 0 {
		return
	}

	// Inset the rect by half the stroke width (CSS border-box model).
	inset := strokeWidth / 2
	ir := Rect{
		X: r.X + inset,
		Y: r.Y + inset,
		W: r.W - strokeWidth,
		H: r.H - strokeWidth,
	}
	if ir.W <= 0 || ir.H <= 0 {
		return
	}

	// Adjust radii for the inset path so outer stroke edge matches CSS border-radius.
	radTL = math.Max(0, radTL-inset)
	radTR = math.Max(0, radTR-inset)
	radBR = math.Max(0, radBR-inset)
	radBL = math.Max(0, radBL-inset)

	// Clamp each radius
	maxRadius := min(ir.W, ir.H) / 2
	if radTL > maxRadius {
		radTL = maxRadius
	}
	if radTR > maxRadius {
		radTR = maxRadius
	}
	if radBR > maxRadius {
		radBR = maxRadius
	}
	if radBL > maxRadius {
		radBL = maxRadius
	}

	path := &vector.Path{}
	x, y := float32(ir.X), float32(ir.Y)
	w, h := float32(ir.W), float32(ir.H)
	rTL := float32(radTL)
	rTR := float32(radTR)
	rBR := float32(radBR)
	rBL := float32(radBL)

	path.MoveTo(x+rTL, y)
	path.LineTo(x+w-rTR, y)
	path.QuadTo(x+w, y, x+w, y+rTR)
	path.LineTo(x+w, y+h-rBR)
	path.QuadTo(x+w, y+h, x+w-rBR, y+h)
	path.LineTo(x+rBL, y+h)
	path.QuadTo(x, y+h, x, y+h-rBL)
	path.LineTo(x, y+rTL)
	path.QuadTo(x, y, x+rTL, y)
	path.Close()

	// Stroke the path
	opts := &vector.StrokeOptions{
		Width:    float32(strokeWidth),
		LineCap:  vector.LineCapRound,
		LineJoin: vector.LineJoinRound,
	}

	vs, is := path.AppendVerticesAndIndicesForStroke(nil, nil, opts)

	cr, cg, cb, ca := clr.RGBA()
	for i := range vs {
		vs[i].ColorR = float32(cr) / 0xffff
		vs[i].ColorG = float32(cg) / 0xffff
		vs[i].ColorB = float32(cb) / 0xffff
		vs[i].ColorA = float32(ca) / 0xffff
	}

	op := &ebiten.DrawTrianglesOptions{}
	op.AntiAlias = true
	screen.DrawTriangles(vs, is, whiteImage, op)
}

// DrawRadialGradient draws a radial gradient in the given rectangle
func DrawRadialGradient(screen *ebiten.Image, r Rect, g *Gradient) {
	if g == nil || len(g.ColorStops) < 2 {
		return
	}

	// Center of the radial gradient
	centerX := r.X + r.W/2
	centerY := r.Y + r.H/2
	maxRadius := max(r.W, r.H) / 2

	// Draw concentric circles from outside to inside
	steps := int(maxRadius)
	if steps > 100 {
		steps = 100 // Performance limit
	}

	for i := steps; i >= 0; i-- {
		t := float64(i) / float64(steps)
		currentRadius := maxRadius * t

		clr := interpolateGradient(g.ColorStops, t)

		// Draw a circle at this radius
		vector.DrawFilledCircle(screen,
			float32(centerX), float32(centerY),
			float32(currentRadius),
			clr, true)
	}
}

// max returns the larger of two float64 values
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// Filter represents CSS-like filter effects
type Filter struct {
	Blur       float64 // blur radius in pixels
	Brightness float64 // 0-2, where 1 is normal
	Contrast   float64 // 0-2, where 1 is normal
	Saturate   float64 // 0-2, where 1 is normal
	Grayscale  float64 // 0-1
	Sepia      float64 // 0-1
	HueRotate  float64 // degrees
	Invert     float64 // 0-1
}

// NewFilter creates a default filter (no effect)
func NewFilter() *Filter {
	return &Filter{
		Brightness: 1,
		Contrast:   1,
		Saturate:   1,
	}
}

// BackdropFilter represents CSS backdrop-filter for glassmorphism
type BackdropFilter struct {
	Blur       float64
	Brightness float64
	Saturate   float64
}

// GlassmorphismStyle creates a glassmorphism effect style
func GlassmorphismStyle(blurAmount float64, bgAlpha float64) (*Style, *BackdropFilter) {
	return &Style{
			BackgroundColor: color.RGBA{255, 255, 255, uint8(bgAlpha * 255)},
			BorderWidth:     1,
			BorderColor:     color.RGBA{255, 255, 255, 50},
			BorderRadius:    16,
		}, &BackdropFilter{
			Blur:       blurAmount,
			Brightness: 1.05,
			Saturate:   1.2,
		}
}

// ParseBoxShadow parses a CSS-like box-shadow string
// Format: "offsetX offsetY blur spread color [inset]"
// Example: "0 4px 6px rgba(0,0,0,0.3)"
func ParseBoxShadow(s string) *BoxShadow {
	s = strings.TrimSpace(s)
	if s == "" || s == "none" {
		return nil
	}

	shadow := &BoxShadow{}

	// Check for inset
	if strings.Contains(s, "inset") {
		shadow.Inset = true
		s = strings.Replace(s, "inset", "", 1)
	}

	// Find color (starts with # or rgb or named color)
	parts := strings.Fields(s)
	if len(parts) < 3 {
		return nil
	}

	// Parse values
	shadow.OffsetX = parsePxValue(parts[0])
	shadow.OffsetY = parsePxValue(parts[1])

	// Handle color that might be split across parts (like "rgba(0, 0, 0, 0.5)")
	colorStart := -1
	for i, part := range parts {
		if strings.HasPrefix(part, "#") || strings.HasPrefix(part, "rgb") ||
			strings.HasPrefix(part, "hsl") || getNamedColor(strings.ToLower(part)) != nil {
			colorStart = i
			break
		}
	}

	if colorStart == -1 {
		colorStart = len(parts) - 1
	}

	// Parse blur and spread
	if colorStart > 2 {
		shadow.Blur = parsePxValue(parts[2])
	}
	if colorStart > 3 {
		shadow.Spread = parsePxValue(parts[3])
	}

	// Parse color
	colorStr := strings.Join(parts[colorStart:], " ")
	shadow.Color = parseColor(colorStr)
	if shadow.Color == nil {
		shadow.Color = color.RGBA{0, 0, 0, 128}
	}

	return shadow
}

// parsePxValue parses a value like "10px" or "10"
func parsePxValue(s string) float64 {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "px")
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// ParseTextShadow parses a CSS text-shadow string
func ParseTextShadow(s string) *TextShadow {
	s = strings.TrimSpace(s)
	if s == "" || s == "none" {
		return nil
	}

	parts := strings.Fields(s)
	if len(parts) < 3 {
		return nil
	}

	shadow := &TextShadow{
		OffsetX: parsePxValue(parts[0]),
		OffsetY: parsePxValue(parts[1]),
	}

	// Check if third part is a number (blur) or color
	if len(parts) >= 4 {
		shadow.Blur = parsePxValue(parts[2])
		shadow.Color = parseColor(strings.Join(parts[3:], " "))
	} else {
		shadow.Color = parseColor(parts[2])
	}

	if shadow.Color == nil {
		shadow.Color = color.RGBA{0, 0, 0, 128}
	}

	return shadow
}

// ParseOutline parses a CSS outline string
// Format: "width style color"
// Example: "2px solid red"
func ParseOutline(s string) *Outline {
	s = strings.TrimSpace(s)
	if s == "" || s == "none" {
		return nil
	}

	parts := strings.Fields(s)
	if len(parts) < 2 {
		return nil
	}

	outline := &Outline{
		Width: parsePxValue(parts[0]),
		Style: "solid",
	}

	if len(parts) >= 2 {
		style := strings.ToLower(parts[1])
		if style == "solid" || style == "dashed" || style == "dotted" {
			outline.Style = style
			if len(parts) >= 3 {
				outline.Color = parseColor(strings.Join(parts[2:], " "))
			}
		} else {
			// parts[1] might be color
			outline.Color = parseColor(strings.Join(parts[1:], " "))
		}
	}

	if outline.Color == nil {
		outline.Color = color.Black
	}

	return outline
}

// DrawPerSideBorder draws borders with independent widths per side.
// Each side (top, right, bottom, left) is rendered as a filled rectangle.
// borderRadius is reserved for future rounded-corner clipping; currently unused.
func DrawPerSideBorder(screen *ebiten.Image, r Rect, top, right, bottom, left float64, clr color.Color, borderRadius float64) {
	if top > 0 {
		vector.DrawFilledRect(screen, float32(r.X), float32(r.Y), float32(r.W), float32(top), clr, false)
	}
	if bottom > 0 {
		vector.DrawFilledRect(screen, float32(r.X), float32(r.Y+r.H-bottom), float32(r.W), float32(bottom), clr, false)
	}
	if left > 0 {
		vector.DrawFilledRect(screen, float32(r.X), float32(r.Y), float32(left), float32(r.H), clr, false)
	}
	if right > 0 {
		vector.DrawFilledRect(screen, float32(r.X+r.W-right), float32(r.Y), float32(right), float32(r.H), clr, false)
	}
}
