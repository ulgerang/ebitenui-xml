package ui

import (
	"encoding/xml"
	"image/color"
	"io"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// SVGDocument represents a parsed SVG document
type SVGDocument struct {
	Width    float64
	Height   float64
	ViewBox  ViewBox
	Elements []SVGElement
	Font     text.Face              // optional font for <text> rendering
	Defs     map[string]interface{} // id -> *SVGLinearGradient or *SVGRadialGradient
}

// SetFont sets the font face used for SVG <text> element rendering.
func (doc *SVGDocument) SetFont(f text.Face) {
	doc.Font = f
}

// ViewBox represents SVG viewBox attribute
type ViewBox struct {
	MinX   float64
	MinY   float64
	Width  float64
	Height float64
}

// SVGElement is the interface for all SVG elements
type SVGElement interface {
	Draw(screen *ebiten.Image, offsetX, offsetY, scaleX, scaleY float64)
}

// SVGGroup represents <g> element
type SVGGroup struct {
	Transform SVGTransform
	Elements  []SVGElement
	Fill      color.Color
	Stroke    color.Color
	StrokeW   float64
	Opacity   float64
}

// SVGGroup Draw implementation
func (g *SVGGroup) Draw(screen *ebiten.Image, offsetX, offsetY, scaleX, scaleY float64) {
	needsOffscreen := !g.Transform.IsSimple() || g.Opacity < 1.0

	if !needsOffscreen {
		// Simple transform, full opacity: draw children directly to screen
		newOffsetX := offsetX + g.Transform.TranslateX*scaleX
		newOffsetY := offsetY + g.Transform.TranslateY*scaleY
		newScaleX := scaleX * g.Transform.ScaleX
		newScaleY := scaleY * g.Transform.ScaleY

		for _, elem := range g.Elements {
			elem.Draw(screen, newOffsetX, newOffsetY, newScaleX, newScaleY)
		}
	} else {
		// Need offscreen compositing for complex transform and/or group opacity.
		bounds := screen.Bounds()
		offscreen := ebiten.NewImage(bounds.Dx(), bounds.Dy())

		if g.Transform.IsSimple() {
			// Simple transform + opacity: draw children to offscreen, composite with alpha
			newOffsetX := g.Transform.TranslateX * scaleX
			newOffsetY := g.Transform.TranslateY * scaleY
			newScaleX := scaleX * g.Transform.ScaleX
			newScaleY := scaleY * g.Transform.ScaleY

			for _, elem := range g.Elements {
				elem.Draw(offscreen, newOffsetX, newOffsetY, newScaleX, newScaleY)
			}

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(offsetX, offsetY)
			if g.Opacity < 1.0 {
				op.ColorScale.ScaleAlpha(float32(g.Opacity))
			}
			screen.DrawImage(offscreen, op)
		} else {
			// Complex transform (rotate, skew, matrix): render to offscreen at full
			// screen resolution, then composite with a GeoM that conjugates the SVG
			// transform with the viewBox scale.
			//
			// Pipeline for a point (x,y) in SVG coords of a child element:
			//   1. Child renders at pixel (x*scaleX, y*scaleY) on the offscreen buffer.
			//   2. GeoM converts pixel coords back to SVG coords: Scale(1/s).
			//   3. The SVG-space transform T is applied.
			//   4. Result is scaled back to pixels: Scale(s).
			//   5. Offset positions the result on the target screen.
			//
			// Effective GeoM = Translate(off) · Scale(s) · T_svg · Scale(1/s)
			for _, elem := range g.Elements {
				elem.Draw(offscreen, 0, 0, scaleX, scaleY)
			}

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(1/scaleX, 1/scaleY)
			svgGeoM := g.Transform.ToGeoM()
			op.GeoM.Concat(svgGeoM)
			op.GeoM.Scale(scaleX, scaleY)
			op.GeoM.Translate(offsetX, offsetY)

			if g.Opacity < 1.0 {
				op.ColorScale.ScaleAlpha(float32(g.Opacity))
			}
			screen.DrawImage(offscreen, op)
		}
		offscreen.Deallocate()
	}
}

// SVGTransform represents transform attribute
type SVGTransform struct {
	TranslateX float64
	TranslateY float64
	ScaleX     float64
	ScaleY     float64
	Rotate     float64
	OriginX    float64 // Rotation origin X
	OriginY    float64 // Rotation origin Y
	SkewX      float64 // Skew angle X in radians
	SkewY      float64 // Skew angle Y in radians
	HasMatrix  bool
	Matrix     [6]float64 // a, b, c, d, e, f
}

func NewSVGTransform() SVGTransform {
	return SVGTransform{ScaleX: 1, ScaleY: 1}
}

// IsSimple returns true if the transform only contains translate and scale.
func (t *SVGTransform) IsSimple() bool {
	return t.Rotate == 0 && t.SkewX == 0 && t.SkewY == 0 && !t.HasMatrix
}

// ToGeoM converts the SVGTransform to an ebiten.GeoM.
// This function returns the GeoM for the group's *own* transform,
// without applying inherited scales or offsets.
func (t *SVGTransform) ToGeoM() ebiten.GeoM {
	var geoM ebiten.GeoM

	// Apply matrix transform first if present
	if t.HasMatrix {
		// Matrix transform: [a c e]
		//                   [b d f]
		// Ebiten GeoM:     [E00 E01 E02]
		//                  [E10 E11 E12]
		// So, E00=a, E10=b, E01=c, E11=d, E02=e, E12=f
		geoM.SetElement(0, 0, t.Matrix[0]) // a
		geoM.SetElement(1, 0, t.Matrix[1]) // b
		geoM.SetElement(0, 1, t.Matrix[2]) // c
		geoM.SetElement(1, 1, t.Matrix[3]) // d
		geoM.SetElement(0, 2, t.Matrix[4]) // e
		geoM.SetElement(1, 2, t.Matrix[5]) // f
	}

	// Apply skewX
	if t.SkewX != 0 {
		skewXMatrix := ebiten.GeoM{}
		skewXMatrix.SetElement(0, 1, math.Tan(t.SkewX)) // tan(angle)
		geoM.Concat(skewXMatrix)
	}

	// Apply skewY
	if t.SkewY != 0 {
		skewYMatrix := ebiten.GeoM{}
		skewYMatrix.SetElement(1, 0, math.Tan(t.SkewY)) // tan(angle)
		geoM.Concat(skewYMatrix)
	}

	// Apply rotation around origin
	if t.Rotate != 0 {
		// Translate to origin
		geoM.Translate(-t.OriginX, -t.OriginY)
		// Rotate
		geoM.Rotate(t.Rotate)
		// Translate back
		geoM.Translate(t.OriginX, t.OriginY)
	}

	// Apply scale
	geoM.Scale(t.ScaleX, t.ScaleY)

	// Apply translate
	geoM.Translate(t.TranslateX, t.TranslateY)

	return geoM
}

// ============================================================================
// SVG Gradient Types
// ============================================================================

// SVGGradientStop represents a <stop> element within a gradient definition.
type SVGGradientStop struct {
	Offset  float64     // 0-1 (parsed from "0%"-"100%" or 0.0-1.0)
	Color   color.Color // stop-color
	Opacity float64     // stop-opacity, default 1
}

// SVGLinearGradient represents a <linearGradient> element.
// Coordinates are in objectBoundingBox units (0-1) by default.
type SVGLinearGradient struct {
	ID             string
	X1, Y1, X2, Y2 float64
	Stops          []SVGGradientStop
}

// SVGRadialGradient represents a <radialGradient> element.
// Coordinates are in objectBoundingBox units (0-1) by default.
type SVGRadialGradient struct {
	ID     string
	CX, CY float64 // center, default 0.5
	R      float64 // radius, default 0.5
	Stops  []SVGGradientStop
}

// ============================================================================
// SVG Clip Path Types
// ============================================================================

// SVGClipPath represents a <clipPath> element containing clipping shapes.
type SVGClipPath struct {
	ID       string
	Elements []SVGElement
}

// SVGClippedElement wraps an SVGElement with a clip path for masked rendering.
// The clip is applied using offscreen compositing with destination-in blending.
type SVGClippedElement struct {
	Child      SVGElement
	ClipPathID string       // "url(#id)" parsed to "id", resolved after parsing
	ClipPath   *SVGClipPath // resolved clip path definition
}

func (c *SVGClippedElement) Draw(screen *ebiten.Image, offsetX, offsetY, scaleX, scaleY float64) {
	if c.ClipPath == nil || len(c.ClipPath.Elements) == 0 {
		c.Child.Draw(screen, offsetX, offsetY, scaleX, scaleY)
		return
	}

	bounds := screen.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	// 1. Render the child element to an offscreen buffer
	content := ebiten.NewImage(w, h)
	c.Child.Draw(content, offsetX, offsetY, scaleX, scaleY)

	// 2. Render the clip path shapes as a mask on a separate buffer.
	//    Any non-transparent pixel in the mask acts as "keep" region.
	mask := ebiten.NewImage(w, h)
	for _, clipElem := range c.ClipPath.Elements {
		clipElem.Draw(mask, offsetX, offsetY, scaleX, scaleY)
	}

	// 3. Destination-in blend: keep content pixels only where mask has alpha > 0
	content.DrawImage(mask, &ebiten.DrawImageOptions{
		Blend: ebiten.BlendDestinationIn,
	})

	// 4. Composite the clipped content onto the original screen
	screen.DrawImage(content, &ebiten.DrawImageOptions{})

	content.Deallocate()
	mask.Deallocate()
}

// SVGRect represents <rect> element
type SVGRect struct {
	X, Y, Width, Height float64
	RX, RY              float64 // rounded corners
	Fill                color.Color
	Stroke              color.Color
	StrokeWidth         float64
	Opacity             float64
	FillOpacity         float64
	StrokeOpacity       float64
	FillGradientID      string      // "url(#id)" parsed to "id"
	FillGradient        interface{} // resolved *SVGLinearGradient or *SVGRadialGradient
}

func (r *SVGRect) Draw(screen *ebiten.Image, offsetX, offsetY, scaleX, scaleY float64) {
	x := offsetX + r.X*scaleX
	y := offsetY + r.Y*scaleY
	w := r.Width * scaleX
	h := r.Height * scaleY
	rx := r.RX * scaleX
	ry := r.RY * scaleY

	// Use larger radius for drawing
	radius := math.Max(rx, ry)

	// Gradient fill
	if r.FillGradient != nil {
		var path vector.Path
		if radius > 0 {
			r32 := float32(radius)
			maxR := float32(math.Min(w, h) / 2)
			if r32 > maxR {
				r32 = maxR
			}
			fx, fy, fw, fh := float32(x), float32(y), float32(w), float32(h)
			path.MoveTo(fx+r32, fy)
			path.LineTo(fx+fw-r32, fy)
			path.ArcTo(fx+fw, fy, fx+fw, fy+r32, r32)
			path.LineTo(fx+fw, fy+fh-r32)
			path.ArcTo(fx+fw, fy+fh, fx+fw-r32, fy+fh, r32)
			path.LineTo(fx+r32, fy+fh)
			path.ArcTo(fx, fy+fh, fx, fy+fh-r32, r32)
			path.LineTo(fx, fy+r32)
			path.ArcTo(fx, fy, fx+r32, fy, r32)
			path.Close()
		} else {
			fx, fy, fw, fh := float32(x), float32(y), float32(w), float32(h)
			path.MoveTo(fx, fy)
			path.LineTo(fx+fw, fy)
			path.LineTo(fx+fw, fy+fh)
			path.LineTo(fx, fy+fh)
			path.Close()
		}
		svgGradientFillPath(screen, &path, r.FillGradient, r.Opacity*r.FillOpacity)
	} else if r.Fill != nil {
		fillColor := applyOpacity(r.Fill, r.Opacity*r.FillOpacity)
		if radius > 0 {
			DrawRoundedRectPath(screen, Rect{X: x, Y: y, W: w, H: h}, radius, fillColor)
		} else {
			vector.DrawFilledRect(screen, float32(x), float32(y), float32(w), float32(h), fillColor, true)
		}
	}

	// Stroke
	if r.Stroke != nil && r.StrokeWidth > 0 {
		strokeOpacity := r.Opacity * r.StrokeOpacity
		sw := float32(r.StrokeWidth * scaleX)
		if strokeOpacity < 1.0 {
			// SVG spec: stroke is painted as a single opaque layer, then composited
			// with opacity. This prevents corner overlap double-blending artifacts.
			bounds := screen.Bounds()
			offscreen := ebiten.NewImage(bounds.Dx(), bounds.Dy())
			if radius > 0 {
				svgDrawRoundedRectStroke(offscreen, x, y, w, h, radius, r.Stroke, sw)
			} else {
				drawRectStroke(offscreen, x, y, w, h, r.Stroke, sw)
			}
			op := &ebiten.DrawImageOptions{}
			op.ColorScale.ScaleAlpha(float32(strokeOpacity))
			screen.DrawImage(offscreen, op)
			offscreen.Deallocate()
		} else {
			strokeColor := applyOpacity(r.Stroke, strokeOpacity)
			if radius > 0 {
				svgDrawRoundedRectStroke(screen, x, y, w, h, radius, strokeColor, sw)
			} else {
				drawRectStroke(screen, x, y, w, h, strokeColor, sw)
			}
		}
	}
}

// SVGCircle represents <circle> element
type SVGCircle struct {
	CX, CY, R      float64
	Fill           color.Color
	Stroke         color.Color
	StrokeWidth    float64
	Opacity        float64
	FillOpacity    float64
	StrokeOpacity  float64
	FillGradientID string
	FillGradient   interface{}
}

func (c *SVGCircle) Draw(screen *ebiten.Image, offsetX, offsetY, scaleX, scaleY float64) {
	cx := offsetX + c.CX*scaleX
	cy := offsetY + c.CY*scaleY
	r := c.R * math.Min(scaleX, scaleY)

	// Gradient fill
	if c.FillGradient != nil {
		// Build circle path using cubic bezier approximation
		var path vector.Path
		k := float32(0.5522847498) // (4/3)*tan(pi/8)
		fcx, fcy, fr := float32(cx), float32(cy), float32(r)
		path.MoveTo(fcx+fr, fcy)
		path.CubicTo(fcx+fr, fcy+k*fr, fcx+k*fr, fcy+fr, fcx, fcy+fr)
		path.CubicTo(fcx-k*fr, fcy+fr, fcx-fr, fcy+k*fr, fcx-fr, fcy)
		path.CubicTo(fcx-fr, fcy-k*fr, fcx-k*fr, fcy-fr, fcx, fcy-fr)
		path.CubicTo(fcx+k*fr, fcy-fr, fcx+fr, fcy-k*fr, fcx+fr, fcy)
		path.Close()
		svgGradientFillPath(screen, &path, c.FillGradient, c.Opacity*c.FillOpacity)
	} else if c.Fill != nil {
		fillColor := applyOpacity(c.Fill, c.Opacity*c.FillOpacity)
		vector.DrawFilledCircle(screen, float32(cx), float32(cy), float32(r), fillColor, true)
	}

	// Stroke
	if c.Stroke != nil && c.StrokeWidth > 0 {
		strokeColor := applyOpacity(c.Stroke, c.Opacity*c.StrokeOpacity)
		sw := float32(c.StrokeWidth * scaleX)
		vector.StrokeCircle(screen, float32(cx), float32(cy), float32(r), sw, strokeColor, true)
	}
}

// SVGEllipse represents <ellipse> element
type SVGEllipse struct {
	CX, CY, RX, RY float64
	Fill           color.Color
	Stroke         color.Color
	StrokeWidth    float64
	Opacity        float64
	FillOpacity    float64
	StrokeOpacity  float64
	FillGradientID string
	FillGradient   interface{}
}

func (e *SVGEllipse) Draw(screen *ebiten.Image, offsetX, offsetY, scaleX, scaleY float64) {
	cx := offsetX + e.CX*scaleX
	cy := offsetY + e.CY*scaleY
	rx := e.RX * scaleX
	ry := e.RY * scaleY

	// Draw ellipse using path
	var path vector.Path
	path.MoveTo(float32(cx+rx), float32(cy))

	// Approximate ellipse with bezier curves
	const segments = 4
	for i := 0; i < segments; i++ {
		theta1 := float64(i) * math.Pi * 2 / segments
		theta2 := float64(i+1) * math.Pi * 2 / segments

		// Control point factor for cubic bezier approximation of arc
		k := 0.5522847498 // (4/3)*tan(pi/8)

		x1 := cx + rx*math.Cos(theta1)
		y1 := cy + ry*math.Sin(theta1)
		x2 := cx + rx*math.Cos(theta2)
		y2 := cy + ry*math.Sin(theta2)

		// Control points
		cp1x := x1 - k*rx*math.Sin(theta1)
		cp1y := y1 + k*ry*math.Cos(theta1)
		cp2x := x2 + k*rx*math.Sin(theta2)
		cp2y := y2 - k*ry*math.Cos(theta2)

		path.CubicTo(float32(cp1x), float32(cp1y), float32(cp2x), float32(cp2y), float32(x2), float32(y2))
	}
	path.Close()

	// Gradient fill
	if e.FillGradient != nil {
		svgGradientFillPath(screen, &path, e.FillGradient, e.Opacity*e.FillOpacity)
	} else if e.Fill != nil {
		fillColor := applyOpacity(e.Fill, e.Opacity*e.FillOpacity)
		vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
		applyColorToVertices(vs, fillColor)
		screen.DrawTriangles(vs, is, whiteImage, &ebiten.DrawTrianglesOptions{AntiAlias: true, FillRule: ebiten.FillRuleNonZero})
	}

	// Stroke
	if e.Stroke != nil && e.StrokeWidth > 0 {
		strokeColor := applyOpacity(e.Stroke, e.Opacity*e.StrokeOpacity)
		sw := float32(e.StrokeWidth * scaleX)
		sop := &vector.StrokeOptions{Width: sw, LineCap: vector.LineCapRound, LineJoin: vector.LineJoinRound}
		vs, is := path.AppendVerticesAndIndicesForStroke(nil, nil, sop)
		applyColorToVertices(vs, strokeColor)
		screen.DrawTriangles(vs, is, whiteImage, &ebiten.DrawTrianglesOptions{AntiAlias: true, FillRule: ebiten.FillRuleNonZero})
	}
}

// SVGLine represents <line> element
type SVGLine struct {
	X1, Y1, X2, Y2 float64
	Stroke         color.Color
	StrokeWidth    float64
	Opacity        float64
	StrokeOpacity  float64
	StrokeLineCap  string
}

func (l *SVGLine) Draw(screen *ebiten.Image, offsetX, offsetY, scaleX, scaleY float64) {
	if l.Stroke == nil || l.StrokeWidth <= 0 {
		return
	}

	x1 := offsetX + l.X1*scaleX
	y1 := offsetY + l.Y1*scaleY
	x2 := offsetX + l.X2*scaleX
	y2 := offsetY + l.Y2*scaleY

	strokeColor := applyOpacity(l.Stroke, l.Opacity*l.StrokeOpacity)
	sw := float32(l.StrokeWidth * scaleX)

	lineCap := vector.LineCapButt
	switch l.StrokeLineCap {
	case "round":
		lineCap = vector.LineCapRound
	case "square":
		lineCap = vector.LineCapSquare
	}

	vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), sw, strokeColor, true)
	_ = lineCap // TODO: Apply line cap when vector package supports it
}

// SVGPolyline represents <polyline> element
type SVGPolyline struct {
	Points         []Point
	Fill           color.Color
	Stroke         color.Color
	StrokeWidth    float64
	Opacity        float64
	FillOpacity    float64
	StrokeOpacity  float64
	FillGradientID string
	FillGradient   interface{}
}

type Point struct {
	X, Y float64
}

func (p *SVGPolyline) Draw(screen *ebiten.Image, offsetX, offsetY, scaleX, scaleY float64) {
	if len(p.Points) < 2 {
		return
	}

	var path vector.Path
	first := p.Points[0]
	path.MoveTo(float32(offsetX+first.X*scaleX), float32(offsetY+first.Y*scaleY))

	for i := 1; i < len(p.Points); i++ {
		pt := p.Points[i]
		path.LineTo(float32(offsetX+pt.X*scaleX), float32(offsetY+pt.Y*scaleY))
	}

	// Stroke
	if p.Stroke != nil && p.StrokeWidth > 0 {
		strokeColor := applyOpacity(p.Stroke, p.Opacity*p.StrokeOpacity)
		sw := float32(p.StrokeWidth * scaleX)
		sop := &vector.StrokeOptions{Width: sw, LineCap: vector.LineCapRound, LineJoin: vector.LineJoinRound}
		vs, is := path.AppendVerticesAndIndicesForStroke(nil, nil, sop)
		applyColorToVertices(vs, strokeColor)
		screen.DrawTriangles(vs, is, whiteImage, &ebiten.DrawTrianglesOptions{AntiAlias: true, FillRule: ebiten.FillRuleNonZero})
	}
}

// SVGPolygon represents <polygon> element
type SVGPolygon struct {
	Points         []Point
	Fill           color.Color
	Stroke         color.Color
	StrokeWidth    float64
	Opacity        float64
	FillOpacity    float64
	StrokeOpacity  float64
	FillRule       string // "nonzero" or "evenodd"
	FillGradientID string
	FillGradient   interface{}
}

func (p *SVGPolygon) Draw(screen *ebiten.Image, offsetX, offsetY, scaleX, scaleY float64) {
	if len(p.Points) < 3 {
		return
	}

	var path vector.Path
	first := p.Points[0]
	path.MoveTo(float32(offsetX+first.X*scaleX), float32(offsetY+first.Y*scaleY))

	for i := 1; i < len(p.Points); i++ {
		pt := p.Points[i]
		path.LineTo(float32(offsetX+pt.X*scaleX), float32(offsetY+pt.Y*scaleY))
	}
	path.Close()

	// Gradient fill
	if p.FillGradient != nil {
		svgGradientFillPath(screen, &path, p.FillGradient, p.Opacity*p.FillOpacity)
	} else if p.Fill != nil {
		fillColor := applyOpacity(p.Fill, p.Opacity*p.FillOpacity)
		vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
		applyColorToVertices(vs, fillColor)
		dto := &ebiten.DrawTrianglesOptions{AntiAlias: true}
		if p.FillRule == "evenodd" {
			dto.FillRule = ebiten.FillRuleEvenOdd
		} else {
			dto.FillRule = ebiten.FillRuleNonZero
		}
		screen.DrawTriangles(vs, is, whiteImage, dto)
	}

	// Stroke
	if p.Stroke != nil && p.StrokeWidth > 0 {
		strokeColor := applyOpacity(p.Stroke, p.Opacity*p.StrokeOpacity)
		sw := float32(p.StrokeWidth * scaleX)
		sop := &vector.StrokeOptions{Width: sw, LineCap: vector.LineCapRound, LineJoin: vector.LineJoinRound}
		vs, is := path.AppendVerticesAndIndicesForStroke(nil, nil, sop)
		applyColorToVertices(vs, strokeColor)
		screen.DrawTriangles(vs, is, whiteImage, &ebiten.DrawTrianglesOptions{AntiAlias: true, FillRule: ebiten.FillRuleNonZero})
	}
}

// SVGPath represents <path> element
type SVGPath struct {
	D              string // path data
	Fill           color.Color
	Stroke         color.Color
	StrokeWidth    float64
	Opacity        float64
	FillOpacity    float64
	StrokeOpacity  float64
	FillRule       string // "nonzero" or "evenodd"
	FillGradientID string
	FillGradient   interface{}
}

func (p *SVGPath) Draw(screen *ebiten.Image, offsetX, offsetY, scaleX, scaleY float64) {
	path := ParsePathDataScaled(p.D, offsetX, offsetY, scaleX, scaleY)
	if path == nil {
		return
	}

	// Gradient fill
	if p.FillGradient != nil {
		svgGradientFillPath(screen, path, p.FillGradient, p.Opacity*p.FillOpacity)
	} else if p.Fill != nil {
		fillColor := applyOpacity(p.Fill, p.Opacity*p.FillOpacity)
		vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
		if len(vs) > 0 {
			applyColorToVertices(vs, fillColor)
			dto := &ebiten.DrawTrianglesOptions{AntiAlias: true}
			if p.FillRule == "evenodd" {
				dto.FillRule = ebiten.FillRuleEvenOdd
			} else {
				dto.FillRule = ebiten.FillRuleNonZero
			}
			screen.DrawTriangles(vs, is, whiteImage, dto)
		}
	}

	// Stroke
	if p.Stroke != nil && p.StrokeWidth > 0 {
		strokeColor := applyOpacity(p.Stroke, p.Opacity*p.StrokeOpacity)
		sw := float32(p.StrokeWidth * scaleX)
		sop := &vector.StrokeOptions{Width: sw, LineCap: vector.LineCapRound, LineJoin: vector.LineJoinRound}
		vs, is := path.AppendVerticesAndIndicesForStroke(nil, nil, sop)
		if len(vs) > 0 {
			applyColorToVertices(vs, strokeColor)
			screen.DrawTriangles(vs, is, whiteImage, &ebiten.DrawTrianglesOptions{AntiAlias: true, FillRule: ebiten.FillRuleNonZero})
		}
	}
}

// SVGUse represents <use> element which references another element by ID.
type SVGUse struct {
	X, Y  float64    // offset position
	RefID string     // referenced element ID (from href="#id")
	Ref   SVGElement // resolved reference
}

// Draw renders the referenced element at the (X, Y) offset.
func (u *SVGUse) Draw(screen *ebiten.Image, offsetX, offsetY, scaleX, scaleY float64) {
	if u.Ref == nil {
		return
	}
	u.Ref.Draw(screen, offsetX+u.X*scaleX, offsetY+u.Y*scaleY, scaleX, scaleY)
}

// LoadSVG loads an SVG file from disk
func LoadSVG(filename string) (*SVGDocument, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ParseSVG(f)
}

// ParseSVG parses SVG from a reader
func ParseSVG(r io.Reader) (*SVGDocument, error) {
	decoder := xml.NewDecoder(r)
	doc := &SVGDocument{
		Defs: make(map[string]interface{}),
	}

	var currentGroup *SVGGroup
	var groupStack []*SVGGroup
	var inheritedFill, inheritedStroke color.Color
	var inheritedStrokeWidth float64 = 1

	// Gradient parsing state
	var inDefs bool
	var currentLinearGrad *SVGLinearGradient
	var currentRadialGrad *SVGRadialGradient
	var currentClipPath *SVGClipPath

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch t := token.(type) {
		case xml.StartElement:
			attrs := attrMap(t.Attr)

			switch t.Name.Local {
			case "svg":
				doc.Width = parseFloat(attrs["width"], 100)
				doc.Height = parseFloat(attrs["height"], 100)
				if vb, ok := attrs["viewBox"]; ok {
					doc.ViewBox = parseViewBox(vb)
				} else {
					doc.ViewBox = ViewBox{Width: doc.Width, Height: doc.Height}
				}

			case "defs":
				inDefs = true

			case "linearGradient":
				currentLinearGrad = &SVGLinearGradient{
					ID: attrs["id"],
					X1: parseFloat(attrs["x1"], 0),
					Y1: parseFloat(attrs["y1"], 0),
					X2: parseFloat(attrs["x2"], 1),
					Y2: parseFloat(attrs["y2"], 0),
				}

			case "radialGradient":
				currentRadialGrad = &SVGRadialGradient{
					ID: attrs["id"],
					CX: parseFloat(attrs["cx"], 0.5),
					CY: parseFloat(attrs["cy"], 0.5),
					R:  parseFloat(attrs["r"], 0.5),
				}

			case "stop":
				stop := SVGGradientStop{
					Offset:  parseSVGStopOffset(attrs["offset"]),
					Color:   parseSVGColor(attrs["stop-color"], color.Black),
					Opacity: parseFloat(attrs["stop-opacity"], 1),
				}
				if currentLinearGrad != nil {
					currentLinearGrad.Stops = append(currentLinearGrad.Stops, stop)
				} else if currentRadialGrad != nil {
					currentRadialGrad.Stops = append(currentRadialGrad.Stops, stop)
				}

			case "clipPath":
				currentClipPath = &SVGClipPath{
					ID: attrs["id"],
				}

			case "g":
				newGroup := &SVGGroup{
					Transform: parseSVGTransform(attrs["transform"]),
					Fill:      parseSVGColor(attrs["fill"], inheritedFill),
					Stroke:    parseSVGColor(attrs["stroke"], inheritedStroke),
					StrokeW:   parseFloat(attrs["stroke-width"], inheritedStrokeWidth),
					Opacity:   parseFloat(attrs["opacity"], 1),
				}
				if currentGroup != nil {
					groupStack = append(groupStack, currentGroup)
				}
				currentGroup = newGroup
				inheritedFill = newGroup.Fill
				inheritedStroke = newGroup.Stroke
				inheritedStrokeWidth = newGroup.StrokeW

			case "rect":
				fillColor, gradID := parseSVGFill(attrs["fill"], inheritedFill)
				elem := &SVGRect{
					X:              parseFloat(attrs["x"], 0),
					Y:              parseFloat(attrs["y"], 0),
					Width:          parseFloat(attrs["width"], 0),
					Height:         parseFloat(attrs["height"], 0),
					RX:             parseFloat(attrs["rx"], 0),
					RY:             parseFloat(attrs["ry"], 0),
					Fill:           fillColor,
					Stroke:         parseSVGColor(attrs["stroke"], inheritedStroke),
					StrokeWidth:    parseFloat(attrs["stroke-width"], inheritedStrokeWidth),
					Opacity:        parseFloat(attrs["opacity"], 1),
					FillOpacity:    parseFloat(attrs["fill-opacity"], 1),
					StrokeOpacity:  parseFloat(attrs["stroke-opacity"], 1),
					FillGradientID: gradID,
				}
				addShapeElement(doc, currentGroup, currentClipPath, attrs["clip-path"], elem, inDefs, attrs["id"])

			case "circle":
				fillColor, gradID := parseSVGFill(attrs["fill"], inheritedFill)
				elem := &SVGCircle{
					CX:             parseFloat(attrs["cx"], 0),
					CY:             parseFloat(attrs["cy"], 0),
					R:              parseFloat(attrs["r"], 0),
					Fill:           fillColor,
					Stroke:         parseSVGColor(attrs["stroke"], inheritedStroke),
					StrokeWidth:    parseFloat(attrs["stroke-width"], inheritedStrokeWidth),
					Opacity:        parseFloat(attrs["opacity"], 1),
					FillOpacity:    parseFloat(attrs["fill-opacity"], 1),
					StrokeOpacity:  parseFloat(attrs["stroke-opacity"], 1),
					FillGradientID: gradID,
				}
				addShapeElement(doc, currentGroup, currentClipPath, attrs["clip-path"], elem, inDefs, attrs["id"])

			case "ellipse":
				fillColor, gradID := parseSVGFill(attrs["fill"], inheritedFill)
				elem := &SVGEllipse{
					CX:             parseFloat(attrs["cx"], 0),
					CY:             parseFloat(attrs["cy"], 0),
					RX:             parseFloat(attrs["rx"], 0),
					RY:             parseFloat(attrs["ry"], 0),
					Fill:           fillColor,
					Stroke:         parseSVGColor(attrs["stroke"], inheritedStroke),
					StrokeWidth:    parseFloat(attrs["stroke-width"], inheritedStrokeWidth),
					Opacity:        parseFloat(attrs["opacity"], 1),
					FillOpacity:    parseFloat(attrs["fill-opacity"], 1),
					StrokeOpacity:  parseFloat(attrs["stroke-opacity"], 1),
					FillGradientID: gradID,
				}
				addShapeElement(doc, currentGroup, currentClipPath, attrs["clip-path"], elem, inDefs, attrs["id"])

			case "line":
				elem := &SVGLine{
					X1:            parseFloat(attrs["x1"], 0),
					Y1:            parseFloat(attrs["y1"], 0),
					X2:            parseFloat(attrs["x2"], 0),
					Y2:            parseFloat(attrs["y2"], 0),
					Stroke:        parseSVGColor(attrs["stroke"], inheritedStroke),
					StrokeWidth:   parseFloat(attrs["stroke-width"], inheritedStrokeWidth),
					Opacity:       parseFloat(attrs["opacity"], 1),
					StrokeOpacity: parseFloat(attrs["stroke-opacity"], 1),
					StrokeLineCap: attrs["stroke-linecap"],
				}
				addShapeElement(doc, currentGroup, currentClipPath, attrs["clip-path"], elem, inDefs, attrs["id"])

			case "polyline":
				fillColor, gradID := parseSVGFill(attrs["fill"], nil) // polyline default no fill
				elem := &SVGPolyline{
					Points:         parsePoints(attrs["points"]),
					Fill:           fillColor,
					Stroke:         parseSVGColor(attrs["stroke"], inheritedStroke),
					StrokeWidth:    parseFloat(attrs["stroke-width"], inheritedStrokeWidth),
					Opacity:        parseFloat(attrs["opacity"], 1),
					FillOpacity:    parseFloat(attrs["fill-opacity"], 1),
					StrokeOpacity:  parseFloat(attrs["stroke-opacity"], 1),
					FillGradientID: gradID,
				}
				addShapeElement(doc, currentGroup, currentClipPath, attrs["clip-path"], elem, inDefs, attrs["id"])

			case "polygon":
				fillColor, gradID := parseSVGFill(attrs["fill"], inheritedFill)
				elem := &SVGPolygon{
					Points:         parsePoints(attrs["points"]),
					Fill:           fillColor,
					Stroke:         parseSVGColor(attrs["stroke"], inheritedStroke),
					StrokeWidth:    parseFloat(attrs["stroke-width"], inheritedStrokeWidth),
					Opacity:        parseFloat(attrs["opacity"], 1),
					FillOpacity:    parseFloat(attrs["fill-opacity"], 1),
					StrokeOpacity:  parseFloat(attrs["stroke-opacity"], 1),
					FillRule:       attrs["fill-rule"],
					FillGradientID: gradID,
				}
				addShapeElement(doc, currentGroup, currentClipPath, attrs["clip-path"], elem, inDefs, attrs["id"])

			case "path":
				fillColor, gradID := parseSVGFill(attrs["fill"], inheritedFill)
				elem := &SVGPath{
					D:              attrs["d"],
					Fill:           fillColor,
					Stroke:         parseSVGColor(attrs["stroke"], inheritedStroke),
					StrokeWidth:    parseFloat(attrs["stroke-width"], inheritedStrokeWidth),
					Opacity:        parseFloat(attrs["opacity"], 1),
					FillOpacity:    parseFloat(attrs["fill-opacity"], 1),
					StrokeOpacity:  parseFloat(attrs["stroke-opacity"], 1),
					FillRule:       attrs["fill-rule"],
					FillGradientID: gradID,
				}
				addShapeElement(doc, currentGroup, currentClipPath, attrs["clip-path"], elem, inDefs, attrs["id"])

			case "use":
				href := attrs["href"]
				if href == "" {
					href = attrs["xlink:href"]
				}
				refID := strings.TrimPrefix(href, "#")
				elem := &SVGUse{
					X:     parseFloat(attrs["x"], 0),
					Y:     parseFloat(attrs["y"], 0),
					RefID: refID,
				}
				addElement(doc, currentGroup, elem)
			}

		case xml.EndElement:
			switch t.Name.Local {
			case "g":
				// Pop group from stack
				if currentGroup != nil {
					addElement(doc, nil, currentGroup)
					if len(groupStack) > 0 {
						currentGroup = groupStack[len(groupStack)-1]
						groupStack = groupStack[:len(groupStack)-1]
						inheritedFill = currentGroup.Fill
						inheritedStroke = currentGroup.Stroke
						inheritedStrokeWidth = currentGroup.StrokeW
					} else {
						currentGroup = nil
						inheritedFill = nil
						inheritedStroke = nil
						inheritedStrokeWidth = 1
					}
				}

			case "defs":
				inDefs = false

			case "linearGradient":
				if currentLinearGrad != nil && currentLinearGrad.ID != "" {
					doc.Defs[currentLinearGrad.ID] = currentLinearGrad
				}
				currentLinearGrad = nil

			case "radialGradient":
				if currentRadialGrad != nil && currentRadialGrad.ID != "" {
					doc.Defs[currentRadialGrad.ID] = currentRadialGrad
				}
				currentRadialGrad = nil

			case "clipPath":
				if currentClipPath != nil {
					// Ensure all clip shapes have opaque fill for mask rendering
					for _, elem := range currentClipPath.Elements {
						ensureClipFill(elem)
					}
					if currentClipPath.ID != "" {
						doc.Defs[currentClipPath.ID] = currentClipPath
					}
				}
				currentClipPath = nil
			}
		}
	}

	// If we're still in a group, add it to doc
	if currentGroup != nil {
		doc.Elements = append(doc.Elements, currentGroup)
	}

	// Resolve gradient references
	resolveGradients(doc)

	// Resolve clip-path references
	resolveClipPaths(doc)

	return doc, nil
}

// ParseSVGString parses SVG from a string
func ParseSVGString(s string) (*SVGDocument, error) {
	return ParseSVG(strings.NewReader(s))
}

// Draw renders the SVG document to the screen
func (doc *SVGDocument) Draw(screen *ebiten.Image, x, y, width, height float64) {
	if doc.ViewBox.Width <= 0 || doc.ViewBox.Height <= 0 {
		return
	}

	scaleX := width / doc.ViewBox.Width
	scaleY := height / doc.ViewBox.Height

	offsetX := x - doc.ViewBox.MinX*scaleX
	offsetY := y - doc.ViewBox.MinY*scaleY

	for _, elem := range doc.Elements {
		elem.Draw(screen, offsetX, offsetY, scaleX, scaleY)
	}
}

// Helper functions

func attrMap(attrs []xml.Attr) map[string]string {
	m := make(map[string]string)
	for _, a := range attrs {
		m[a.Name.Local] = a.Value
	}
	return m
}

func parseFloat(s string, def float64) float64 {
	if s == "" {
		return def
	}
	// Remove units like "px", "pt", etc.
	s = strings.TrimSuffix(s, "px")
	s = strings.TrimSuffix(s, "pt")
	s = strings.TrimSuffix(s, "em")
	s = strings.TrimSpace(s)

	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return def
	}
	return v
}

func parseViewBox(s string) ViewBox {
	parts := strings.Fields(s)
	if len(parts) < 4 {
		return ViewBox{}
	}
	return ViewBox{
		MinX:   parseFloat(parts[0], 0),
		MinY:   parseFloat(parts[1], 0),
		Width:  parseFloat(parts[2], 0),
		Height: parseFloat(parts[3], 0),
	}
}

func parseSVGTransform(s string) SVGTransform {
	t := NewSVGTransform()
	if s == "" {
		return t
	}

	// Parse transform functions
	for _, fn := range strings.Split(s, ")") {
		fn = strings.TrimSpace(fn)
		if fn == "" {
			continue
		}

		parts := strings.SplitN(fn, "(", 2)
		if len(parts) != 2 {
			continue
		}

		name := strings.TrimSpace(parts[0])
		args := parseTransformArgs(parts[1])

		switch name {
		case "translate":
			if len(args) >= 1 {
				t.TranslateX = args[0]
			}
			if len(args) >= 2 {
				t.TranslateY = args[1]
			}
		case "scale":
			if len(args) >= 1 {
				t.ScaleX = args[0]
				t.ScaleY = args[0]
			}
			if len(args) >= 2 {
				t.ScaleY = args[1]
			}
		case "rotate":
			if len(args) >= 1 {
				t.Rotate = args[0] * math.Pi / 180
			}
			if len(args) >= 3 {
				t.OriginX = args[1]
				t.OriginY = args[2]
			}
		case "skewX":
			if len(args) >= 1 {
				t.SkewX = args[0] * math.Pi / 180
			}
		case "skewY":
			if len(args) >= 1 {
				t.SkewY = args[0] * math.Pi / 180
			}
		case "matrix":
			if len(args) == 6 {
				t.HasMatrix = true
				for i := 0; i < 6; i++ {
					t.Matrix[i] = args[i]
				}
			}
		}
	}

	return t
}

func parseTransformArgs(s string) []float64 {
	s = strings.ReplaceAll(s, ",", " ")
	parts := strings.Fields(s)
	result := make([]float64, 0, len(parts))
	for _, p := range parts {
		v, err := strconv.ParseFloat(p, 64)
		if err == nil {
			result = append(result, v)
		}
	}
	return result
}

func parseSVGColor(s string, inherited color.Color) color.Color {
	if s == "" {
		return inherited
	}
	if s == "none" || s == "transparent" {
		return nil
	}
	if s == "currentColor" || s == "inherit" {
		return inherited
	}
	return parseColor(s)
}

// parseSVGFill parses a fill attribute value, detecting url(#id) gradient references.
// Returns (color, gradientID). If gradientID is non-empty, color will be nil.
func parseSVGFill(s string, inherited color.Color) (color.Color, string) {
	if s == "" {
		return inherited, ""
	}
	// Detect gradient reference: fill="url(#gradientID)"
	if strings.HasPrefix(s, "url(#") {
		id := strings.TrimPrefix(s, "url(#")
		id = strings.TrimSuffix(id, ")")
		id = strings.TrimSpace(id)
		return nil, id
	}
	return parseSVGColor(s, inherited), ""
}

// parseSVGStopOffset parses a <stop> offset attribute.
// Accepts "0%"-"100%" or 0.0-1.0 numeric values.
func parseSVGStopOffset(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	if strings.HasSuffix(s, "%") {
		s = strings.TrimSuffix(s, "%")
		v, _ := strconv.ParseFloat(s, 64)
		return v / 100
	}
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

// resolveGradients walks all elements in the document and links
// FillGradientID strings to actual gradient definition pointers.
func resolveGradients(doc *SVGDocument) {
	if len(doc.Defs) == 0 {
		return
	}
	resolveGradientElements(doc.Defs, doc.Elements)
}

func resolveGradientElements(defs map[string]interface{}, elements []SVGElement) {
	for _, elem := range elements {
		switch e := elem.(type) {
		case *SVGRect:
			if e.FillGradientID != "" {
				e.FillGradient = defs[e.FillGradientID]
			}
		case *SVGCircle:
			if e.FillGradientID != "" {
				e.FillGradient = defs[e.FillGradientID]
			}
		case *SVGEllipse:
			if e.FillGradientID != "" {
				e.FillGradient = defs[e.FillGradientID]
			}
		case *SVGPolygon:
			if e.FillGradientID != "" {
				e.FillGradient = defs[e.FillGradientID]
			}
		case *SVGPolyline:
			if e.FillGradientID != "" {
				e.FillGradient = defs[e.FillGradientID]
			}
		case *SVGPath:
			if e.FillGradientID != "" {
				e.FillGradient = defs[e.FillGradientID]
			}
		case *SVGGroup:
			resolveGradientElements(defs, e.Elements)
		case *SVGUse:
			if e.RefID != "" {
				if ref, ok := defs[e.RefID]; ok {
					if svgElem, ok := ref.(SVGElement); ok {
						e.Ref = svgElem
					}
				}
			}
		}
	}
}

func parsePoints(s string) []Point {
	s = strings.ReplaceAll(s, ",", " ")
	parts := strings.Fields(s)
	points := make([]Point, 0, len(parts)/2)

	for i := 0; i+1 < len(parts); i += 2 {
		x, _ := strconv.ParseFloat(parts[i], 64)
		y, _ := strconv.ParseFloat(parts[i+1], 64)
		points = append(points, Point{X: x, Y: y})
	}
	return points
}

func addElement(doc *SVGDocument, group *SVGGroup, elem SVGElement) {
	if group != nil {
		group.Elements = append(group.Elements, elem)
	} else {
		doc.Elements = append(doc.Elements, elem)
	}
}

// addShapeElement routes a parsed shape element to the correct destination:
//   - If inside a <clipPath> definition, adds to the clip path's element list
//   - If the shape has a clip-path="url(#id)" attribute, wraps it in SVGClippedElement
//   - Otherwise, adds normally via addElement
func addShapeElement(doc *SVGDocument, group *SVGGroup, clipPath *SVGClipPath, clipPathAttr string, elem SVGElement, inDefs bool, id string) {
	if inDefs && id != "" {
		doc.Defs[id] = elem
		return
	}
	if clipPath != nil {
		clipPath.Elements = append(clipPath.Elements, elem)
		return
	}
	clipID := parseSVGClipPathRef(clipPathAttr)
	if clipID != "" {
		wrapped := &SVGClippedElement{
			Child:      elem,
			ClipPathID: clipID,
		}
		addElement(doc, group, wrapped)
		return
	}
	addElement(doc, group, elem)
}

// parseSVGClipPathRef extracts an id from a clip-path="url(#id)" attribute value.
func parseSVGClipPathRef(s string) string {
	if !strings.HasPrefix(s, "url(#") {
		return ""
	}
	id := strings.TrimPrefix(s, "url(#")
	id = strings.TrimSuffix(id, ")")
	return strings.TrimSpace(id)
}

// ensureClipFill ensures a clip path shape has an opaque fill so it renders
// as a solid mask region. SVG spec says clip shapes default to black fill,
// but our parser may leave fill as nil if not explicitly set.
func ensureClipFill(elem SVGElement) {
	switch e := elem.(type) {
	case *SVGRect:
		if e.Fill == nil {
			e.Fill = color.White
		}
	case *SVGCircle:
		if e.Fill == nil {
			e.Fill = color.White
		}
	case *SVGEllipse:
		if e.Fill == nil {
			e.Fill = color.White
		}
	case *SVGPath:
		if e.Fill == nil {
			e.Fill = color.White
		}
	case *SVGPolygon:
		if e.Fill == nil {
			e.Fill = color.White
		}
	case *SVGPolyline:
		if e.Fill == nil {
			e.Fill = color.White
		}
	}
}

// resolveClipPaths walks all elements and links ClipPathID strings to
// actual *SVGClipPath definition pointers from doc.Defs.
func resolveClipPaths(doc *SVGDocument) {
	if len(doc.Defs) == 0 {
		return
	}
	resolveClipPathElements(doc.Defs, doc.Elements)
}

func resolveClipPathElements(defs map[string]interface{}, elements []SVGElement) {
	for _, elem := range elements {
		switch e := elem.(type) {
		case *SVGClippedElement:
			if e.ClipPathID != "" {
				if cp, ok := defs[e.ClipPathID].(*SVGClipPath); ok {
					e.ClipPath = cp
				}
			}
		case *SVGGroup:
			resolveClipPathElements(defs, e.Elements)
		}
	}
}

// Note: applyOpacity is defined in widget.go

func applyColorToVertices(vs []ebiten.Vertex, c color.Color) {
	r, g, b, a := c.RGBA()
	rf := float32(r) / 0xffff
	gf := float32(g) / 0xffff
	bf := float32(b) / 0xffff
	af := float32(a) / 0xffff

	for i := range vs {
		vs[i].ColorR = rf
		vs[i].ColorG = gf
		vs[i].ColorB = bf
		vs[i].ColorA = af
	}
}

func drawRectStroke(screen *ebiten.Image, x, y, w, h float64, c color.Color, strokeWidth float32) {
	var path vector.Path
	path.MoveTo(float32(x), float32(y))
	path.LineTo(float32(x+w), float32(y))
	path.LineTo(float32(x+w), float32(y+h))
	path.LineTo(float32(x), float32(y+h))
	path.Close()

	sop := &vector.StrokeOptions{
		Width:    strokeWidth,
		LineJoin: vector.LineJoinMiter, // SVG default is miter
	}
	vs, is := path.AppendVerticesAndIndicesForStroke(nil, nil, sop)
	applyColorToVertices(vs, c)
	screen.DrawTriangles(vs, is, whiteImage, &ebiten.DrawTrianglesOptions{AntiAlias: true, FillRule: ebiten.FillRuleNonZero})
}

func svgDrawRoundedRectStroke(screen *ebiten.Image, x, y, w, h, radius float64, c color.Color, strokeWidth float32) {
	var path vector.Path

	// Clamp radius
	maxRadius := math.Min(w, h) / 2
	if radius > maxRadius {
		radius = maxRadius
	}

	r := float32(radius)

	// Start at top-left after corner
	path.MoveTo(float32(x)+r, float32(y))

	// Top edge and top-right corner
	path.LineTo(float32(x+w)-r, float32(y))
	path.ArcTo(float32(x+w), float32(y), float32(x+w), float32(y)+r, r)

	// Right edge and bottom-right corner
	path.LineTo(float32(x+w), float32(y+h)-r)
	path.ArcTo(float32(x+w), float32(y+h), float32(x+w)-r, float32(y+h), r)

	// Bottom edge and bottom-left corner
	path.LineTo(float32(x)+r, float32(y+h))
	path.ArcTo(float32(x), float32(y+h), float32(x), float32(y+h)-r, r)

	// Left edge and top-left corner
	path.LineTo(float32(x), float32(y)+r)
	path.ArcTo(float32(x), float32(y), float32(x)+r, float32(y), r)

	path.Close()

	sop := &vector.StrokeOptions{Width: strokeWidth, LineCap: vector.LineCapRound, LineJoin: vector.LineJoinRound}
	vs, is := path.AppendVerticesAndIndicesForStroke(nil, nil, sop)
	applyColorToVertices(vs, c)
	screen.DrawTriangles(vs, is, whiteImage, &ebiten.DrawTrianglesOptions{AntiAlias: true, FillRule: ebiten.FillRuleNonZero})
}

// transformPath is deprecated - transforms are now applied at the vertex level in applyPathTransform
// This function is kept for backward compatibility but just returns the original path
func transformPath(path *vector.Path, offsetX, offsetY, scaleX, scaleY float64) *vector.Path {
	return path
}

// applyPathTransform applies offset and scale to vertices
func applyPathTransform(vs []ebiten.Vertex, offsetX, offsetY, scaleX, scaleY float64) {
	for i := range vs {
		vs[i].DstX = float32(offsetX) + vs[i].DstX*float32(scaleX)
		vs[i].DstY = float32(offsetY) + vs[i].DstY*float32(scaleY)
	}
}

// ============================================================================
// SVG Gradient Fill Rendering
// ============================================================================

// svgStopsToColorStops converts SVGGradientStop slice to ColorStop slice,
// applying stop-opacity into the colour's alpha channel.
func svgStopsToColorStops(stops []SVGGradientStop) []ColorStop {
	result := make([]ColorStop, len(stops))
	for i, s := range stops {
		c := s.Color
		if s.Opacity < 1 {
			c = applyOpacity(c, s.Opacity)
		}
		result[i] = ColorStop{Color: c, Position: s.Offset}
	}
	return result
}

// buildSVGLinearGradientTexture creates a 2D texture by sampling the linear
// gradient defined in objectBoundingBox coordinates (x1,y1)→(x2,y2).
func buildSVGLinearGradientTexture(grad *SVGLinearGradient, w, h int) *ebiten.Image {
	if w <= 0 || h <= 0 {
		return nil
	}
	stops := svgStopsToColorStops(grad.Stops)
	img := ebiten.NewImage(w, h)
	pix := make([]byte, w*h*4)

	dx := grad.X2 - grad.X1
	dy := grad.Y2 - grad.Y1
	lenSq := dx*dx + dy*dy
	if lenSq == 0 {
		lenSq = 1 // avoid division by zero — fallback to first stop
	}

	for py := 0; py < h; py++ {
		for px := 0; px < w; px++ {
			// Normalise pixel to objectBoundingBox [0,1]
			var nx, ny float64
			if w > 1 {
				nx = float64(px) / float64(w-1)
			}
			if h > 1 {
				ny = float64(py) / float64(h-1)
			}

			// Project onto gradient line
			t := ((nx-grad.X1)*dx + (ny-grad.Y1)*dy) / lenSq
			if t < 0 {
				t = 0
			}
			if t > 1 {
				t = 1
			}

			c := interpolateGradient(stops, t)
			r, g, b, a := c.RGBA()
			off := (py*w + px) * 4
			pix[off+0] = uint8(r >> 8)
			pix[off+1] = uint8(g >> 8)
			pix[off+2] = uint8(b >> 8)
			pix[off+3] = uint8(a >> 8)
		}
	}
	img.WritePixels(pix)
	return img
}

// buildSVGRadialGradientTexture creates a 2D texture by sampling the radial
// gradient defined in objectBoundingBox coordinates with centre (cx,cy) and radius r.
func buildSVGRadialGradientTexture(grad *SVGRadialGradient, w, h int) *ebiten.Image {
	if w <= 0 || h <= 0 {
		return nil
	}
	stops := svgStopsToColorStops(grad.Stops)
	img := ebiten.NewImage(w, h)
	pix := make([]byte, w*h*4)

	radius := grad.R
	if radius <= 0 {
		radius = 0.5
	}

	for py := 0; py < h; py++ {
		for px := 0; px < w; px++ {
			var nx, ny float64
			if w > 1 {
				nx = float64(px) / float64(w-1)
			}
			if h > 1 {
				ny = float64(py) / float64(h-1)
			}

			ddx := (nx - grad.CX) / radius
			ddy := (ny - grad.CY) / radius
			dist := math.Sqrt(ddx*ddx + ddy*ddy)
			if dist > 1 {
				dist = 1
			}

			c := interpolateGradient(stops, dist)
			r, g, b, a := c.RGBA()
			off := (py*w + px) * 4
			pix[off+0] = uint8(r >> 8)
			pix[off+1] = uint8(g >> 8)
			pix[off+2] = uint8(b >> 8)
			pix[off+3] = uint8(a >> 8)
		}
	}
	img.WritePixels(pix)
	return img
}

// drawSVGGradientFill renders a gradient-filled shape using the tessellated
// fill vertices. It maps vertex positions to gradient-texture UV coordinates
// based on the vertex bounding box.
func drawSVGGradientFill(screen *ebiten.Image, vs []ebiten.Vertex, is []uint16, gradTex *ebiten.Image, opacity float64) {
	if len(vs) == 0 || gradTex == nil {
		return
	}

	// Find vertex bounding box
	minX, minY := float32(math.MaxFloat32), float32(math.MaxFloat32)
	maxX, maxY := float32(-math.MaxFloat32), float32(-math.MaxFloat32)
	for _, v := range vs {
		if v.DstX < minX {
			minX = v.DstX
		}
		if v.DstY < minY {
			minY = v.DstY
		}
		if v.DstX > maxX {
			maxX = v.DstX
		}
		if v.DstY > maxY {
			maxY = v.DstY
		}
	}

	bw := maxX - minX
	bh := maxY - minY
	if bw <= 0 || bh <= 0 {
		return
	}

	tw := float32(gradTex.Bounds().Dx())
	th := float32(gradTex.Bounds().Dy())
	opf := float32(opacity)

	// Map vertex positions to texture UV coordinates
	for i := range vs {
		vs[i].SrcX = (vs[i].DstX - minX) / bw * tw
		vs[i].SrcY = (vs[i].DstY - minY) / bh * th
		vs[i].ColorR = opf
		vs[i].ColorG = opf
		vs[i].ColorB = opf
		vs[i].ColorA = opf
	}

	screen.DrawTriangles(vs, is, gradTex, &ebiten.DrawTrianglesOptions{AntiAlias: true})
}

// svgBuildGradientTex builds the appropriate gradient texture for the given
// gradient definition and pixel dimensions.
func svgBuildGradientTex(fillGradient interface{}, w, h int) *ebiten.Image {
	switch g := fillGradient.(type) {
	case *SVGLinearGradient:
		return buildSVGLinearGradientTexture(g, w, h)
	case *SVGRadialGradient:
		return buildSVGRadialGradientTexture(g, w, h)
	}
	return nil
}

// svgGradientFillPath tessellates the given path, builds a gradient texture
// matching the vertex bounding box, and draws the textured fill.
// Returns true if gradient was drawn, false if no gradient is set.
func svgGradientFillPath(screen *ebiten.Image, path *vector.Path, fillGradient interface{}, opacity float64) bool {
	if fillGradient == nil {
		return false
	}
	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
	if len(vs) == 0 {
		return true // gradient set but nothing to draw
	}

	// Compute bounding box for texture dimensions
	minX, minY := float32(math.MaxFloat32), float32(math.MaxFloat32)
	maxX, maxY := float32(-math.MaxFloat32), float32(-math.MaxFloat32)
	for _, v := range vs {
		if v.DstX < minX {
			minX = v.DstX
		}
		if v.DstY < minY {
			minY = v.DstY
		}
		if v.DstX > maxX {
			maxX = v.DstX
		}
		if v.DstY > maxY {
			maxY = v.DstY
		}
	}
	tw := int(maxX-minX) + 1
	th := int(maxY-minY) + 1
	if tw < 1 {
		tw = 1
	}
	if th < 1 {
		th = 1
	}

	gradTex := svgBuildGradientTex(fillGradient, tw, th)
	if gradTex == nil {
		return true
	}
	drawSVGGradientFill(screen, vs, is, gradTex, opacity)
	gradTex.Deallocate()
	return true
}
