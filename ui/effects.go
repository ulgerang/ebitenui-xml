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

	// strip is the cached 1D gradient lookup texture (256×1 pixels).
	// Built lazily on first GPU draw and reused across frames.
	strip *ebiten.Image
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
//
// This implementation uses a GPU Kage shader for single-draw-call rendering.
// The gradient is baked into a 256×1 lookup texture and the shader computes
// the gradient position per-pixel on the GPU.
func DrawGradient(screen *ebiten.Image, r Rect, g *Gradient) {
	if g == nil || len(g.ColorStops) < 2 {
		return
	}

	w := int(r.W)
	h := int(r.H)
	if w <= 0 || h <= 0 {
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

	// Precompute shader uniforms so the fragment shader only does:
	//   t = GradA * dstPos.x + GradB * dstPos.y + GradC
	//
	// Derivation (center-relative → absolute with GeoM translation):
	//   px_rel = dstPos.x - r.X - hw
	//   py_rel = dstPos.y - r.Y - hh
	//   dot    = cosA*px_rel + sinA*py_rel
	//   t      = (dot - minDot) / dotRange
	//          = cosA/dotRange * dstPos.x + sinA/dotRange * dstPos.y
	//            + (-(hw*cosA + hh*sinA + minDot) - cosA*r.X - sinA*r.Y) / dotRange
	gradA := cosA / dotRange
	gradB := sinA / dotRange
	gradC := (-(hw*cosA + hh*sinA + minDot) - cosA*r.X - sinA*r.Y) / dotRange

	// Ensure the 1D gradient strip texture is ready (lazy, cached).
	strip := g.ensureGradientStrip()

	shader := getLinearGradientShader()

	op := &ebiten.DrawRectShaderOptions{}
	op.GeoM.Translate(r.X, r.Y)
	op.Uniforms = map[string]any{
		"GradA": float32(gradA),
		"GradB": float32(gradB),
		"GradC": float32(gradC),
	}
	op.Images[0] = strip
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

// DrawBoxShadow draws a CSS-compliant box shadow using an SDF-based GPU shader.
// The shader analytically computes the Gaussian-blurred alpha of a rounded
// rectangle via erfc(d / (σ√2)), requiring only a single draw call and zero
// offscreen buffers.
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

	// 3·sigma covers 99.7% of Gaussian energy — extend draw rect by this amount.
	padding := math.Ceil(3 * sigma)

	// The shader rectangle covers the shadow shape plus blur padding.
	drawW := int(math.Ceil(shapeW + 2*padding))
	drawH := int(math.Ceil(shapeH + 2*padding))
	if drawW <= 0 || drawH <= 0 {
		return
	}

	// Where the draw rectangle starts on screen.
	drawX := r.X + offsetX - spread - padding
	drawY := r.Y + offsetY - spread - padding

	// Centre of the shadow shape in destination coordinates.
	centerX := drawX + float64(drawW)/2
	centerY := drawY + float64(drawH)/2

	// Shadow colour — extract premultiplied RGBA components.
	sr, sg, sb, sa := shadow.Color.RGBA()
	colR := float64(sr) / 0xffff
	colG := float64(sg) / 0xffff
	colB := float64(sb) / 0xffff
	colA := float64(sa) / 0xffff

	shader := getBoxShadowShader()

	op := &ebiten.DrawRectShaderOptions{}
	op.GeoM.Translate(drawX, drawY)
	op.Uniforms = map[string]any{
		"CenterX": float32(centerX),
		"CenterY": float32(centerY),
		"HalfW":   float32(shapeW / 2),
		"HalfH":   float32(shapeH / 2),
		"Radius":  float32(shapeRadius),
		"Sigma":   float32(sigma),
		"ShadowR": float32(colR),
		"ShadowG": float32(colG),
		"ShadowB": float32(colB),
		"ShadowA": float32(colA),
	}
	screen.DrawRectShader(drawW, drawH, shader, op)
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
