package ui

import (
	"image"
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
	case "ease-in", "easein":
		return EaseInQuad
	case "ease-out", "easeout":
		return EaseOutQuad
	case "ease-in-out", "easeinout":
		return EaseInOutQuad
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
	strip      *ebiten.Image // cached 1D gradient strip texture (lazy, see shader.go)
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
// Uses a GPU shader with a 1D gradient strip texture for O(1) draw calls regardless of size.
func DrawGradient(screen *ebiten.Image, r Rect, g *Gradient) {
	if g == nil || len(g.ColorStops) < 2 {
		return
	}

	// CSS angle to math direction vector.
	// CSS: 0deg points upward (bottom-to-top), 90deg points right.
	// angleRad = (cssAngle - 90) * pi/180 gives: cos=dx, sin=dy of gradient direction.
	angleRad := (g.Angle - 90) * math.Pi / 180
	cosA := math.Cos(angleRad)
	sinA := math.Sin(angleRad)

	// Half-dimensions for center-relative corner projection
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

	// Precompute shader uniforms: t = GradA*dstPos.x + GradB*dstPos.y + GradC
	// maps destination pixel coords (including GeoM translation) directly to t ∈ [0,1].
	gradA := cosA / dotRange
	gradB := sinA / dotRange
	gradC := -(hw*cosA+hh*sinA+minDot)/dotRange - gradA*r.X - gradB*r.Y

	shader := getLinearGradientShader()
	strip := g.ensureGradientStrip()

	w, h := int(r.W), int(r.H)
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}

	op := &ebiten.DrawRectShaderOptions{}
	op.GeoM.Translate(r.X, r.Y)
	op.Images[0] = strip
	op.Uniforms = map[string]any{
		"GradA": float32(gradA),
		"GradB": float32(gradB),
		"GradC": float32(gradC),
	}
	screen.DrawRectShader(w, h, shader, op)
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

// DrawBoxShadow draws a box shadow effect around a rectangle.
// Outset shadows use a GPU SDF shader with analytical Gaussian blur (single draw call).
// Inset shadows fall back to a CPU multi-layer approach for now.
func DrawBoxShadow(screen *ebiten.Image, r Rect, shadow *BoxShadow, borderRadius float64) {
	if shadow == nil {
		return
	}

	if shadow.Inset {
		// Inset shadow: keep CPU fallback for now
		drawBoxShadowInsetCPU(screen, r, shadow, borderRadius)
		return
	}

	shader := getBoxShadowShader()
	sigma := shadow.Blur / 2
	if sigma < 0.001 {
		sigma = 0.001
	}

	expand := shadow.Blur*3 + math.Abs(shadow.Spread)

	// Shadow shape dimensions (element + spread)
	shapeHalfW := r.W/2 + shadow.Spread
	shapeHalfH := r.H/2 + shadow.Spread
	cornerRadius := borderRadius + shadow.Spread
	if cornerRadius < 0 {
		cornerRadius = 0
	}
	maxR := min(shapeHalfW, shapeHalfH)
	if cornerRadius > maxR {
		cornerRadius = maxR
	}

	// Draw rect (expanded to cover blur extent)
	drawX := r.X + shadow.OffsetX - expand
	drawY := r.Y + shadow.OffsetY - expand
	drawW := r.W + expand*2
	drawH := r.H + expand*2

	// Shadow center in destination coords
	centerX := r.X + r.W/2 + shadow.OffsetX
	centerY := r.Y + r.H/2 + shadow.OffsetY

	sr, sg, sb, sa := shadow.Color.RGBA()

	w, h := int(drawW), int(drawH)
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}

	op := &ebiten.DrawRectShaderOptions{}
	op.GeoM.Translate(drawX, drawY)
	op.Uniforms = map[string]any{
		"CenterX": float32(centerX),
		"CenterY": float32(centerY),
		"HalfW":   float32(shapeHalfW),
		"HalfH":   float32(shapeHalfH),
		"Radius":  float32(cornerRadius),
		"Sigma":   float32(sigma),
		"ShadowR": float32(sr) / 0xffff,
		"ShadowG": float32(sg) / 0xffff,
		"ShadowB": float32(sb) / 0xffff,
		"ShadowA": float32(sa) / 0xffff,
	}
	screen.DrawRectShader(w, h, shader, op)
}

// drawBoxShadowInsetCPU draws an inset box shadow using CPU multi-layer approach.
// This is a fallback for inset shadows until a dedicated GPU shader is implemented.
func drawBoxShadowInsetCPU(screen *ebiten.Image, r Rect, shadow *BoxShadow, borderRadius float64) {
	// Shadow color with alpha for blur
	sr, sg, sb, sa := shadow.Color.RGBA()
	baseAlpha := float64(sa) / 0xffff

	// Draw multiple layers for blur effect
	blurSteps := int(shadow.Blur)
	if blurSteps < 1 {
		blurSteps = 1
	}
	if blurSteps > 20 {
		blurSteps = 20 // limit for performance
	}

	for i := blurSteps; i >= 0; i-- {
		// Alpha decreases as we go further out
		layerAlpha := baseAlpha * (1.0 - float64(i)/float64(blurSteps+1)) * 0.5

		// Expand size for each layer
		expand := float64(i) + shadow.Spread

		layerRect := Rect{
			X: r.X + shadow.OffsetX - expand,
			Y: r.Y + shadow.OffsetY - expand,
			W: r.W + expand*2,
			H: r.H + expand*2,
		}

		shadowColor := color.RGBA{
			R: uint8(sr >> 8),
			G: uint8(sg >> 8),
			B: uint8(sb >> 8),
			A: uint8(layerAlpha * 255),
		}

		DrawRoundedRectPath(screen, layerRect, borderRadius+expand, shadowColor)
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
func drawRoundedRectStrokeEx(screen *ebiten.Image, r Rect, radTL, radTR, radBR, radBL float64, strokeWidth float64, clr color.Color) {
	if strokeWidth <= 0 {
		return
	}

	// Clamp each radius
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

// DrawRadialGradient draws a radial gradient in the given rectangle.
// Uses a GPU shader with a 1D gradient strip texture for proper elliptical rendering.
func DrawRadialGradient(screen *ebiten.Image, r Rect, g *Gradient) {
	if g == nil || len(g.ColorStops) < 2 {
		return
	}

	shader := getRadialGradientShader()
	strip := g.ensureGradientStrip()

	w, h := int(r.W), int(r.H)
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}

	op := &ebiten.DrawRectShaderOptions{}
	op.GeoM.Translate(r.X, r.Y)
	op.Images[0] = strip
	op.Uniforms = map[string]any{
		"CenterX": float32(r.X + r.W/2),
		"CenterY": float32(r.Y + r.H/2),
		"RadiusX": float32(r.W / 2),
		"RadiusY": float32(r.H / 2),
	}
	screen.DrawRectShader(w, h, shader, op)
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

// NewBackdropFilter creates a BackdropFilter with identity defaults.
func NewBackdropFilter() *BackdropFilter {
	return &BackdropFilter{Brightness: 1, Saturate: 1}
}

// backdropFilterIsDefault returns true when the filter would have no visible effect.
func backdropFilterIsDefault(bf *BackdropFilter) bool {
	return bf.Blur == 0 && bf.Brightness == 1 && bf.Saturate == 1
}

// filterIsDefault returns true when a CSS filter would have no visible effect.
func filterIsDefault(f *Filter) bool {
	return f.Blur == 0 && f.Brightness == 1 && f.Contrast == 1 &&
		f.Saturate == 1 && f.Grayscale == 0 && f.Sepia == 0 &&
		f.HueRotate == 0 && f.Invert == 0
}

// ApplyBackdropFilter captures the screen region behind a widget, applies
// blur/brightness/saturate effects, and composites the result back.
func ApplyBackdropFilter(screen *ebiten.Image, r Rect, bf *BackdropFilter) {
	if bf == nil || backdropFilterIsDefault(bf) {
		return
	}

	x0, y0 := int(r.X), int(r.Y)
	x1, y1 := int(r.X+r.W), int(r.Y+r.H)
	w, h := x1-x0, y1-y0
	if w <= 0 || h <= 0 {
		return
	}

	// Capture what's currently on screen behind this widget
	captured := screen.SubImage(image.Rect(x0, y0, x1, y1)).(*ebiten.Image)

	// Copy captured region into an offscreen buffer (origin-shifted)
	buf := ebiten.NewImage(w, h)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(-x0), float64(-y0))
	buf.DrawImage(captured, op)

	// 2-pass separable Gaussian blur
	if bf.Blur > 0 {
		shader := getBackdropBlurShader()
		if shader != nil {
			tmp := ebiten.NewImage(w, h)

			// Horizontal pass: buf → tmp
			sop := &ebiten.DrawRectShaderOptions{}
			sop.Images[0] = buf
			sop.Uniforms = map[string]interface{}{
				"Sigma":     float32(bf.Blur),
				"Direction": [2]float32{1, 0},
			}
			tmp.DrawRectShader(w, h, shader, sop)

			// Vertical pass: tmp → buf
			buf.Clear()
			sop2 := &ebiten.DrawRectShaderOptions{}
			sop2.Images[0] = tmp
			sop2.Uniforms = map[string]interface{}{
				"Sigma":     float32(bf.Blur),
				"Direction": [2]float32{0, 1},
			}
			buf.DrawRectShader(w, h, shader, sop2)

			tmp.Deallocate()
		}
	}

	// Apply brightness/saturate using the CSS filter shader
	if bf.Brightness != 1 || bf.Saturate != 1 {
		f := &Filter{
			Brightness: bf.Brightness,
			Contrast:   1,
			Saturate:   bf.Saturate,
			Grayscale:  0,
			Sepia:      0,
			HueRotate:  0,
			Invert:     0,
		}
		filtered := applyCSSFilter(buf, f)
		if filtered != buf {
			buf.Deallocate()
			buf = filtered
		}
	}

	// Composite back onto screen at the original position
	compOp := &ebiten.DrawImageOptions{}
	compOp.GeoM.Translate(float64(x0), float64(y0))
	screen.DrawImage(buf, compOp)

	buf.Deallocate()
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

// ============================================================================
// Gradient with border-radius clipping (tessellate + UV-mapped gradient texture)
// ============================================================================

// DrawGradientClipped draws a gradient clipped to a rounded rectangle path.
// Uses tessellation of the rounded rect path with UV-mapped gradient texture.
// Falls back to the fast GPU shader path when no rounding is needed.
func DrawGradientClipped(screen *ebiten.Image, r Rect, g *Gradient, radTL, radTR, radBR, radBL float64) {
	if g == nil || len(g.ColorStops) < 2 {
		return
	}

	// If no rounding, use fast GPU shader path
	if radTL <= 0 && radTR <= 0 && radBR <= 0 && radBL <= 0 {
		if g.Type == GradientRadial {
			DrawRadialGradient(screen, r, g)
		} else {
			DrawGradient(screen, r, g)
		}
		return
	}

	// Build rounded rect path (same as DrawRoundedRectPathEx)
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

	// Tessellate path
	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
	if len(vs) == 0 {
		return
	}

	// Build gradient texture
	gradW := int(r.W) + 1
	gradH := int(r.H) + 1
	if gradW < 1 {
		gradW = 1
	}
	if gradH < 1 {
		gradH = 1
	}

	var gradTex *ebiten.Image
	if g.Type == GradientRadial {
		gradTex = buildRadialGradientTexture(g, gradW, gradH)
	} else {
		gradTex = buildLinearGradientTexture(g, gradW, gradH)
	}
	if gradTex == nil {
		return
	}
	defer gradTex.Deallocate()

	// UV-map vertices to gradient texture
	tw := float32(gradTex.Bounds().Dx())
	th := float32(gradTex.Bounds().Dy())

	for i := range vs {
		vs[i].SrcX = (vs[i].DstX - float32(r.X)) / float32(r.W) * tw
		vs[i].SrcY = (vs[i].DstY - float32(r.Y)) / float32(r.H) * th
		vs[i].ColorR = 1
		vs[i].ColorG = 1
		vs[i].ColorB = 1
		vs[i].ColorA = 1
	}

	screen.DrawTriangles(vs, is, gradTex, &ebiten.DrawTrianglesOptions{AntiAlias: true})
}

// buildLinearGradientTexture creates a 2D texture for a CSS linear gradient.
// Used by DrawGradientClipped for UV-mapped tessellated rendering with border-radius.
func buildLinearGradientTexture(g *Gradient, w, h int) *ebiten.Image {
	if w <= 0 || h <= 0 {
		return nil
	}

	angleRad := (g.Angle - 90) * math.Pi / 180
	cosA := math.Cos(angleRad)
	sinA := math.Sin(angleRad)

	hw, hh := float64(w)/2, float64(h)/2
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

	img := ebiten.NewImage(w, h)
	pix := make([]byte, w*h*4)

	for py := 0; py < h; py++ {
		for px := 0; px < w; px++ {
			rx := float64(px) - hw
			ry := float64(py) - hh
			dot := rx*cosA + ry*sinA
			t := (dot - minDot) / dotRange
			if t < 0 {
				t = 0
			}
			if t > 1 {
				t = 1
			}
			c := interpolateGradient(g.ColorStops, t)
			cr, cg, cb, ca := c.RGBA()
			off := (py*w + px) * 4
			pix[off+0] = uint8(cr >> 8)
			pix[off+1] = uint8(cg >> 8)
			pix[off+2] = uint8(cb >> 8)
			pix[off+3] = uint8(ca >> 8)
		}
	}
	img.WritePixels(pix)
	return img
}

// buildRadialGradientTexture creates a 2D texture for a CSS radial gradient.
// Used by DrawGradientClipped for UV-mapped tessellated rendering with border-radius.
func buildRadialGradientTexture(g *Gradient, w, h int) *ebiten.Image {
	if w <= 0 || h <= 0 {
		return nil
	}

	img := ebiten.NewImage(w, h)
	pix := make([]byte, w*h*4)

	cx, cy := float64(w)/2, float64(h)/2
	rx, ry := cx, cy
	if rx <= 0 {
		rx = 1
	}
	if ry <= 0 {
		ry = 1
	}

	for py := 0; py < h; py++ {
		for px := 0; px < w; px++ {
			dx := (float64(px) - cx) / rx
			dy := (float64(py) - cy) / ry
			t := math.Sqrt(dx*dx + dy*dy)
			if t > 1 {
				t = 1
			}
			c := interpolateGradient(g.ColorStops, t)
			cr, cg, cb, ca := c.RGBA()
			off := (py*w + px) * 4
			pix[off+0] = uint8(cr >> 8)
			pix[off+1] = uint8(cg >> 8)
			pix[off+2] = uint8(cb >> 8)
			pix[off+3] = uint8(ca >> 8)
		}
	}
	img.WritePixels(pix)
	return img
}
