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
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// SVGDocument represents a parsed SVG document
type SVGDocument struct {
	Width    float64
	Height   float64
	ViewBox  ViewBox
	Elements []SVGElement
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
}

// SVGTransform represents transform attribute
type SVGTransform struct {
	TranslateX float64
	TranslateY float64
	ScaleX     float64
	ScaleY     float64
	Rotate     float64
	OriginX    float64
	OriginY    float64
}

func NewSVGTransform() SVGTransform {
	return SVGTransform{ScaleX: 1, ScaleY: 1}
}

// SVGRect represents <rect> element
type SVGRect struct {
	X, Y, Width, Height float64
	RX, RY              float64 // rounded corners
	Fill                color.Color
	Stroke              color.Color
	StrokeWidth         float64
	Opacity             float64
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

	// Fill
	if r.Fill != nil {
		fillColor := applyOpacity(r.Fill, r.Opacity)
		if radius > 0 {
			DrawRoundedRectPath(screen, Rect{X: x, Y: y, W: w, H: h}, radius, fillColor)
		} else {
			vector.DrawFilledRect(screen, float32(x), float32(y), float32(w), float32(h), fillColor, true)
		}
	}

	// Stroke
	if r.Stroke != nil && r.StrokeWidth > 0 {
		strokeColor := applyOpacity(r.Stroke, r.Opacity)
		sw := float32(r.StrokeWidth * scaleX)
		if radius > 0 {
			svgDrawRoundedRectStroke(screen, x, y, w, h, radius, strokeColor, sw)
		} else {
			drawRectStroke(screen, x, y, w, h, strokeColor, sw)
		}
	}
}

// SVGCircle represents <circle> element
type SVGCircle struct {
	CX, CY, R   float64
	Fill        color.Color
	Stroke      color.Color
	StrokeWidth float64
	Opacity     float64
}

func (c *SVGCircle) Draw(screen *ebiten.Image, offsetX, offsetY, scaleX, scaleY float64) {
	cx := offsetX + c.CX*scaleX
	cy := offsetY + c.CY*scaleY
	r := c.R * math.Min(scaleX, scaleY)

	// Fill
	if c.Fill != nil {
		fillColor := applyOpacity(c.Fill, c.Opacity)
		vector.DrawFilledCircle(screen, float32(cx), float32(cy), float32(r), fillColor, true)
	}

	// Stroke
	if c.Stroke != nil && c.StrokeWidth > 0 {
		strokeColor := applyOpacity(c.Stroke, c.Opacity)
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

	// Fill
	if e.Fill != nil {
		fillColor := applyOpacity(e.Fill, e.Opacity)
		vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
		applyColorToVertices(vs, fillColor)
		screen.DrawTriangles(vs, is, whiteImage, &ebiten.DrawTrianglesOptions{AntiAlias: true})
	}

	// Stroke
	if e.Stroke != nil && e.StrokeWidth > 0 {
		strokeColor := applyOpacity(e.Stroke, e.Opacity)
		sw := float32(e.StrokeWidth * scaleX)
		sop := &vector.StrokeOptions{Width: sw, LineCap: vector.LineCapRound, LineJoin: vector.LineJoinRound}
		vs, is := path.AppendVerticesAndIndicesForStroke(nil, nil, sop)
		applyColorToVertices(vs, strokeColor)
		screen.DrawTriangles(vs, is, whiteImage, &ebiten.DrawTrianglesOptions{AntiAlias: true})
	}
}

// SVGLine represents <line> element
type SVGLine struct {
	X1, Y1, X2, Y2 float64
	Stroke         color.Color
	StrokeWidth    float64
	Opacity        float64
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

	strokeColor := applyOpacity(l.Stroke, l.Opacity)
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
	Points      []Point
	Fill        color.Color
	Stroke      color.Color
	StrokeWidth float64
	Opacity     float64
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
		strokeColor := applyOpacity(p.Stroke, p.Opacity)
		sw := float32(p.StrokeWidth * scaleX)
		sop := &vector.StrokeOptions{Width: sw, LineCap: vector.LineCapRound, LineJoin: vector.LineJoinRound}
		vs, is := path.AppendVerticesAndIndicesForStroke(nil, nil, sop)
		applyColorToVertices(vs, strokeColor)
		screen.DrawTriangles(vs, is, whiteImage, &ebiten.DrawTrianglesOptions{AntiAlias: true})
	}
}

// SVGPolygon represents <polygon> element
type SVGPolygon struct {
	Points      []Point
	Fill        color.Color
	Stroke      color.Color
	StrokeWidth float64
	Opacity     float64
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

	// Fill
	if p.Fill != nil {
		fillColor := applyOpacity(p.Fill, p.Opacity)
		vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
		applyColorToVertices(vs, fillColor)
		screen.DrawTriangles(vs, is, whiteImage, &ebiten.DrawTrianglesOptions{AntiAlias: true})
	}

	// Stroke
	if p.Stroke != nil && p.StrokeWidth > 0 {
		strokeColor := applyOpacity(p.Stroke, p.Opacity)
		sw := float32(p.StrokeWidth * scaleX)
		sop := &vector.StrokeOptions{Width: sw, LineCap: vector.LineCapRound, LineJoin: vector.LineJoinRound}
		vs, is := path.AppendVerticesAndIndicesForStroke(nil, nil, sop)
		applyColorToVertices(vs, strokeColor)
		screen.DrawTriangles(vs, is, whiteImage, &ebiten.DrawTrianglesOptions{AntiAlias: true})
	}
}

// SVGPath represents <path> element
type SVGPath struct {
	D           string // path data
	Fill        color.Color
	Stroke      color.Color
	StrokeWidth float64
	Opacity     float64
	FillRule    string // "nonzero" or "evenodd"
}

func (p *SVGPath) Draw(screen *ebiten.Image, offsetX, offsetY, scaleX, scaleY float64) {
	path := ParsePathData(p.D)
	if path == nil {
		return
	}

	// Fill
	if p.Fill != nil {
		fillColor := applyOpacity(p.Fill, p.Opacity)
		vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
		if len(vs) > 0 {
			applyPathTransform(vs, offsetX, offsetY, scaleX, scaleY)
			applyColorToVertices(vs, fillColor)
			screen.DrawTriangles(vs, is, whiteImage, &ebiten.DrawTrianglesOptions{AntiAlias: true})
		}
	}

	// Stroke
	if p.Stroke != nil && p.StrokeWidth > 0 {
		strokeColor := applyOpacity(p.Stroke, p.Opacity)
		sw := float32(p.StrokeWidth * scaleX)
		sop := &vector.StrokeOptions{Width: sw, LineCap: vector.LineCapRound, LineJoin: vector.LineJoinRound}
		vs, is := path.AppendVerticesAndIndicesForStroke(nil, nil, sop)
		if len(vs) > 0 {
			applyPathTransform(vs, offsetX, offsetY, scaleX, scaleY)
			applyColorToVertices(vs, strokeColor)
			screen.DrawTriangles(vs, is, whiteImage, &ebiten.DrawTrianglesOptions{AntiAlias: true})
		}
	}
}

// SVGGroup Draw implementation
func (g *SVGGroup) Draw(screen *ebiten.Image, offsetX, offsetY, scaleX, scaleY float64) {
	// Apply group transform
	newOffsetX := offsetX + g.Transform.TranslateX*scaleX
	newOffsetY := offsetY + g.Transform.TranslateY*scaleY
	newScaleX := scaleX * g.Transform.ScaleX
	newScaleY := scaleY * g.Transform.ScaleY

	for _, elem := range g.Elements {
		elem.Draw(screen, newOffsetX, newOffsetY, newScaleX, newScaleY)
	}
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
	doc := &SVGDocument{}

	var currentGroup *SVGGroup
	var groupStack []*SVGGroup
	var inheritedFill, inheritedStroke color.Color
	var inheritedStrokeWidth float64 = 1

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

			case "g":
				newGroup := &SVGGroup{
					Transform: parseSVGTransform(attrs["transform"]),
					Fill:      parseSVGColor(attrs["fill"], inheritedFill),
					Stroke:    parseSVGColor(attrs["stroke"], inheritedStroke),
					StrokeW:   parseFloat(attrs["stroke-width"], inheritedStrokeWidth),
				}
				if currentGroup != nil {
					groupStack = append(groupStack, currentGroup)
				}
				currentGroup = newGroup
				inheritedFill = newGroup.Fill
				inheritedStroke = newGroup.Stroke
				inheritedStrokeWidth = newGroup.StrokeW

			case "rect":
				elem := &SVGRect{
					X:           parseFloat(attrs["x"], 0),
					Y:           parseFloat(attrs["y"], 0),
					Width:       parseFloat(attrs["width"], 0),
					Height:      parseFloat(attrs["height"], 0),
					RX:          parseFloat(attrs["rx"], 0),
					RY:          parseFloat(attrs["ry"], 0),
					Fill:        parseSVGColor(attrs["fill"], inheritedFill),
					Stroke:      parseSVGColor(attrs["stroke"], inheritedStroke),
					StrokeWidth: parseFloat(attrs["stroke-width"], inheritedStrokeWidth),
					Opacity:     parseFloat(attrs["opacity"], 1),
				}
				// Handle fill-opacity and stroke-opacity
				if fo, ok := attrs["fill-opacity"]; ok {
					elem.Opacity = parseFloat(fo, 1)
				}
				addElement(doc, currentGroup, elem)

			case "circle":
				elem := &SVGCircle{
					CX:          parseFloat(attrs["cx"], 0),
					CY:          parseFloat(attrs["cy"], 0),
					R:           parseFloat(attrs["r"], 0),
					Fill:        parseSVGColor(attrs["fill"], inheritedFill),
					Stroke:      parseSVGColor(attrs["stroke"], inheritedStroke),
					StrokeWidth: parseFloat(attrs["stroke-width"], inheritedStrokeWidth),
					Opacity:     parseFloat(attrs["opacity"], 1),
				}
				addElement(doc, currentGroup, elem)

			case "ellipse":
				elem := &SVGEllipse{
					CX:          parseFloat(attrs["cx"], 0),
					CY:          parseFloat(attrs["cy"], 0),
					RX:          parseFloat(attrs["rx"], 0),
					RY:          parseFloat(attrs["ry"], 0),
					Fill:        parseSVGColor(attrs["fill"], inheritedFill),
					Stroke:      parseSVGColor(attrs["stroke"], inheritedStroke),
					StrokeWidth: parseFloat(attrs["stroke-width"], inheritedStrokeWidth),
					Opacity:     parseFloat(attrs["opacity"], 1),
				}
				addElement(doc, currentGroup, elem)

			case "line":
				elem := &SVGLine{
					X1:            parseFloat(attrs["x1"], 0),
					Y1:            parseFloat(attrs["y1"], 0),
					X2:            parseFloat(attrs["x2"], 0),
					Y2:            parseFloat(attrs["y2"], 0),
					Stroke:        parseSVGColor(attrs["stroke"], inheritedStroke),
					StrokeWidth:   parseFloat(attrs["stroke-width"], inheritedStrokeWidth),
					Opacity:       parseFloat(attrs["opacity"], 1),
					StrokeLineCap: attrs["stroke-linecap"],
				}
				addElement(doc, currentGroup, elem)

			case "polyline":
				elem := &SVGPolyline{
					Points:      parsePoints(attrs["points"]),
					Fill:        parseSVGColor(attrs["fill"], nil), // polyline default no fill
					Stroke:      parseSVGColor(attrs["stroke"], inheritedStroke),
					StrokeWidth: parseFloat(attrs["stroke-width"], inheritedStrokeWidth),
					Opacity:     parseFloat(attrs["opacity"], 1),
				}
				addElement(doc, currentGroup, elem)

			case "polygon":
				elem := &SVGPolygon{
					Points:      parsePoints(attrs["points"]),
					Fill:        parseSVGColor(attrs["fill"], inheritedFill),
					Stroke:      parseSVGColor(attrs["stroke"], inheritedStroke),
					StrokeWidth: parseFloat(attrs["stroke-width"], inheritedStrokeWidth),
					Opacity:     parseFloat(attrs["opacity"], 1),
				}
				addElement(doc, currentGroup, elem)

			case "path":
				elem := &SVGPath{
					D:           attrs["d"],
					Fill:        parseSVGColor(attrs["fill"], inheritedFill),
					Stroke:      parseSVGColor(attrs["stroke"], inheritedStroke),
					StrokeWidth: parseFloat(attrs["stroke-width"], inheritedStrokeWidth),
					Opacity:     parseFloat(attrs["opacity"], 1),
					FillRule:    attrs["fill-rule"],
				}
				addElement(doc, currentGroup, elem)
			}

		case xml.EndElement:
			if t.Name.Local == "g" {
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
			}
		}
	}

	// If we're still in a group, add it to doc
	if currentGroup != nil {
		doc.Elements = append(doc.Elements, currentGroup)
	}

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
	// Top
	vector.StrokeLine(screen, float32(x), float32(y), float32(x+w), float32(y), strokeWidth, c, true)
	// Right
	vector.StrokeLine(screen, float32(x+w), float32(y), float32(x+w), float32(y+h), strokeWidth, c, true)
	// Bottom
	vector.StrokeLine(screen, float32(x+w), float32(y+h), float32(x), float32(y+h), strokeWidth, c, true)
	// Left
	vector.StrokeLine(screen, float32(x), float32(y+h), float32(x), float32(y), strokeWidth, c, true)
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
	screen.DrawTriangles(vs, is, whiteImage, &ebiten.DrawTrianglesOptions{AntiAlias: true})
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
