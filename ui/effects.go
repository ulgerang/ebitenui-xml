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

// DrawGradient draws a gradient in the given rectangle
func DrawGradient(screen *ebiten.Image, r Rect, g *Gradient) {
	if g == nil || len(g.ColorStops) < 2 {
		return
	}

	// Simple linear gradient implementation (horizontal)
	for x := 0.0; x < r.W; x++ {
		t := x / r.W

		// Find color at position t
		clr := interpolateGradient(g.ColorStops, t)

		vector.DrawFilledRect(screen,
			float32(r.X+x), float32(r.Y),
			1, float32(r.H),
			clr, false)
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

// DrawRoundedRectPath draws a rounded rectangle with proper corners
func DrawRoundedRectPath(screen *ebiten.Image, r Rect, radius float64, clr color.Color) {
	if radius <= 0 {
		vector.DrawFilledRect(screen,
			float32(r.X), float32(r.Y),
			float32(r.W), float32(r.H),
			clr, false)
		return
	}

	// Clamp radius to half of smallest dimension
	maxRadius := min(r.W, r.H) / 2
	if radius > maxRadius {
		radius = maxRadius
	}

	// Draw using path for proper rounded corners
	path := &vector.Path{}

	rad := float32(radius)
	x, y := float32(r.X), float32(r.Y)
	w, h := float32(r.W), float32(r.H)

	// Start from top-left corner (after the curve)
	path.MoveTo(x+rad, y)

	// Top edge
	path.LineTo(x+w-rad, y)
	// Top-right corner
	path.QuadTo(x+w, y, x+w, y+rad)

	// Right edge
	path.LineTo(x+w, y+h-rad)
	// Bottom-right corner
	path.QuadTo(x+w, y+h, x+w-rad, y+h)

	// Bottom edge
	path.LineTo(x+rad, y+h)
	// Bottom-left corner
	path.QuadTo(x, y+h, x, y+h-rad)

	// Left edge
	path.LineTo(x, y+rad)
	// Top-left corner
	path.QuadTo(x, y, x+rad, y)

	path.Close()

	// Fill the path
	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)

	// Set color for all vertices
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

// DrawBoxShadow draws a box shadow effect around a rectangle
func DrawBoxShadow(screen *ebiten.Image, r Rect, shadow *BoxShadow, borderRadius float64) {
	if shadow == nil {
		return
	}

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

// drawRoundedRectStroke draws a stroked rounded rectangle
func drawRoundedRectStroke(screen *ebiten.Image, r Rect, radius float64, strokeWidth float64, clr color.Color) {
	if strokeWidth <= 0 {
		return
	}

	// Clamp radius
	maxRadius := min(r.W, r.H) / 2
	if radius > maxRadius {
		radius = maxRadius
	}

	path := &vector.Path{}
	rad := float32(radius)
	x, y := float32(r.X), float32(r.Y)
	w, h := float32(r.W), float32(r.H)

	path.MoveTo(x+rad, y)
	path.LineTo(x+w-rad, y)
	path.QuadTo(x+w, y, x+w, y+rad)
	path.LineTo(x+w, y+h-rad)
	path.QuadTo(x+w, y+h, x+w-rad, y+h)
	path.LineTo(x+rad, y+h)
	path.QuadTo(x, y+h, x, y+h-rad)
	path.LineTo(x, y+rad)
	path.QuadTo(x, y, x+rad, y)
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
