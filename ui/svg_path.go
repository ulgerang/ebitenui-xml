package ui

import (
	"image/color"
	"math"
	"strconv"
	"strings"
	"unicode"

	"github.com/hajimehoshi/ebiten/v2/vector"
)

// PathCommand represents a single SVG path command
type PathCommand struct {
	Command  rune
	Args     []float64
	Absolute bool
}

// ParsePathData parses SVG path data string into a vector.Path
// Supports: M, L, H, V, C, S, Q, T, A, Z (and lowercase relative versions)
func ParsePathData(d string) *vector.Path {
	if d == "" {
		return nil
	}

	commands := tokenizePathData(d)
	if len(commands) == 0 {
		return nil
	}

	var path vector.Path
	var currentX, currentY float64
	var startX, startY float64 // For Z command
	var lastControlX, lastControlY float64
	var lastCommand rune

	for _, cmd := range commands {
		args := cmd.Args
		isAbs := cmd.Absolute

		switch cmd.Command {
		case 'M', 'm': // MoveTo
			for i := 0; i+1 < len(args); i += 2 {
				x, y := args[i], args[i+1]
				if !isAbs {
					x += currentX
					y += currentY
				}
				if i == 0 {
					path.MoveTo(float32(x), float32(y))
					startX, startY = x, y
				} else {
					// Subsequent pairs are treated as LineTo
					path.LineTo(float32(x), float32(y))
				}
				currentX, currentY = x, y
			}

		case 'L', 'l': // LineTo
			for i := 0; i+1 < len(args); i += 2 {
				x, y := args[i], args[i+1]
				if !isAbs {
					x += currentX
					y += currentY
				}
				path.LineTo(float32(x), float32(y))
				currentX, currentY = x, y
			}

		case 'H', 'h': // Horizontal LineTo
			for _, x := range args {
				if !isAbs {
					x += currentX
				}
				path.LineTo(float32(x), float32(currentY))
				currentX = x
			}

		case 'V', 'v': // Vertical LineTo
			for _, y := range args {
				if !isAbs {
					y += currentY
				}
				path.LineTo(float32(currentX), float32(y))
				currentY = y
			}

		case 'C', 'c': // Cubic Bezier
			for i := 0; i+5 < len(args); i += 6 {
				x1, y1 := args[i], args[i+1]
				x2, y2 := args[i+2], args[i+3]
				x, y := args[i+4], args[i+5]
				if !isAbs {
					x1 += currentX
					y1 += currentY
					x2 += currentX
					y2 += currentY
					x += currentX
					y += currentY
				}
				path.CubicTo(float32(x1), float32(y1), float32(x2), float32(y2), float32(x), float32(y))
				lastControlX, lastControlY = x2, y2
				currentX, currentY = x, y
			}

		case 'S', 's': // Smooth Cubic Bezier
			for i := 0; i+3 < len(args); i += 4 {
				// Reflect last control point
				x1, y1 := currentX, currentY
				if lastCommand == 'C' || lastCommand == 'c' || lastCommand == 'S' || lastCommand == 's' {
					x1 = 2*currentX - lastControlX
					y1 = 2*currentY - lastControlY
				}
				x2, y2 := args[i], args[i+1]
				x, y := args[i+2], args[i+3]
				if !isAbs {
					x2 += currentX
					y2 += currentY
					x += currentX
					y += currentY
				}
				path.CubicTo(float32(x1), float32(y1), float32(x2), float32(y2), float32(x), float32(y))
				lastControlX, lastControlY = x2, y2
				currentX, currentY = x, y
			}

		case 'Q', 'q': // Quadratic Bezier
			for i := 0; i+3 < len(args); i += 4 {
				x1, y1 := args[i], args[i+1]
				x, y := args[i+2], args[i+3]
				if !isAbs {
					x1 += currentX
					y1 += currentY
					x += currentX
					y += currentY
				}
				path.QuadTo(float32(x1), float32(y1), float32(x), float32(y))
				lastControlX, lastControlY = x1, y1
				currentX, currentY = x, y
			}

		case 'T', 't': // Smooth Quadratic Bezier
			for i := 0; i+1 < len(args); i += 2 {
				// Reflect last control point
				x1, y1 := currentX, currentY
				if lastCommand == 'Q' || lastCommand == 'q' || lastCommand == 'T' || lastCommand == 't' {
					x1 = 2*currentX - lastControlX
					y1 = 2*currentY - lastControlY
				}
				x, y := args[i], args[i+1]
				if !isAbs {
					x += currentX
					y += currentY
				}
				path.QuadTo(float32(x1), float32(y1), float32(x), float32(y))
				lastControlX, lastControlY = x1, y1
				currentX, currentY = x, y
			}

		case 'A', 'a': // Arc
			for i := 0; i+6 < len(args); i += 7 {
				rx, ry := args[i], args[i+1]
				xAxisRotation := args[i+2]
				largeArcFlag := args[i+3] != 0
				sweepFlag := args[i+4] != 0
				x, y := args[i+5], args[i+6]
				if !isAbs {
					x += currentX
					y += currentY
				}

				// Convert arc to bezier curves
				arcToBezier(&path, currentX, currentY, rx, ry, xAxisRotation, largeArcFlag, sweepFlag, x, y)
				currentX, currentY = x, y
			}

		case 'Z', 'z': // ClosePath
			path.Close()
			currentX, currentY = startX, startY
		}

		lastCommand = cmd.Command
	}

	return &path
}

// tokenizePathData tokenizes the path data string into commands
func tokenizePathData(d string) []PathCommand {
	var commands []PathCommand
	var currentCmd rune
	var currentArgs []float64
	var isAbs bool

	// State machine for parsing
	var numBuf strings.Builder
	inNumber := false
	hasDecimal := false
	hasExponent := false

	flushNumber := func() {
		if numBuf.Len() > 0 {
			s := numBuf.String()
			if v, err := strconv.ParseFloat(s, 64); err == nil {
				currentArgs = append(currentArgs, v)
			}
			numBuf.Reset()
			inNumber = false
			hasDecimal = false
			hasExponent = false
		}
	}

	flushCommand := func() {
		if currentCmd != 0 {
			commands = append(commands, PathCommand{
				Command:  unicode.ToUpper(currentCmd),
				Args:     currentArgs,
				Absolute: isAbs,
			})
		}
		currentArgs = nil
	}

	for _, r := range d {
		switch {
		case unicode.IsLetter(r):
			flushNumber()
			flushCommand()
			currentCmd = r
			isAbs = unicode.IsUpper(r)

		case unicode.IsDigit(r):
			inNumber = true
			numBuf.WriteRune(r)

		case r == '.':
			if hasDecimal {
				// Start new number
				flushNumber()
			}
			inNumber = true
			hasDecimal = true
			numBuf.WriteRune(r)

		case r == '-' || r == '+':
			if inNumber && !hasExponent {
				// This is a new number
				flushNumber()
			}
			inNumber = true
			numBuf.WriteRune(r)

		case r == 'e' || r == 'E':
			if inNumber {
				hasExponent = true
				numBuf.WriteRune(r)
			}

		case r == ',' || unicode.IsSpace(r):
			flushNumber()
		}
	}

	flushNumber()
	flushCommand()

	return commands
}

// arcToBezier converts an SVG arc to cubic bezier curves
// Based on: https://www.w3.org/TR/SVG/implnote.html#ArcImplementationNotes
func arcToBezier(path *vector.Path, x1, y1, rx, ry, xAxisRotation float64, largeArc, sweep bool, x2, y2 float64) {
	// Handle degenerate cases
	if rx == 0 || ry == 0 {
		path.LineTo(float32(x2), float32(y2))
		return
	}

	// Absolute values
	rx = math.Abs(rx)
	ry = math.Abs(ry)

	// Convert angle to radians
	phi := xAxisRotation * math.Pi / 180

	// Compute center parameterization
	dx := (x1 - x2) / 2
	dy := (y1 - y2) / 2

	cosPhi := math.Cos(phi)
	sinPhi := math.Sin(phi)

	x1p := cosPhi*dx + sinPhi*dy
	y1p := -sinPhi*dx + cosPhi*dy

	// Correct out-of-range radii
	lambda := (x1p*x1p)/(rx*rx) + (y1p*y1p)/(ry*ry)
	if lambda > 1 {
		sqrtLambda := math.Sqrt(lambda)
		rx *= sqrtLambda
		ry *= sqrtLambda
	}

	// Compute center point
	rxSq := rx * rx
	rySq := ry * ry
	x1pSq := x1p * x1p
	y1pSq := y1p * y1p

	radical := (rxSq*rySq - rxSq*y1pSq - rySq*x1pSq) / (rxSq*y1pSq + rySq*x1pSq)
	if radical < 0 {
		radical = 0
	}
	radical = math.Sqrt(radical)

	if largeArc == sweep {
		radical = -radical
	}

	cxp := radical * rx * y1p / ry
	cyp := -radical * ry * x1p / rx

	cx := cosPhi*cxp - sinPhi*cyp + (x1+x2)/2
	cy := sinPhi*cxp + cosPhi*cyp + (y1+y2)/2

	// Compute angles
	theta1 := angle(1, 0, (x1p-cxp)/rx, (y1p-cyp)/ry)
	dTheta := angle((x1p-cxp)/rx, (y1p-cyp)/ry, (-x1p-cxp)/rx, (-y1p-cyp)/ry)

	if !sweep && dTheta > 0 {
		dTheta -= 2 * math.Pi
	} else if sweep && dTheta < 0 {
		dTheta += 2 * math.Pi
	}

	// Split arc into segments
	segments := int(math.Ceil(math.Abs(dTheta) / (math.Pi / 2)))
	if segments < 1 {
		segments = 1
	}

	segmentAngle := dTheta / float64(segments)

	for i := 0; i < segments; i++ {
		startAngle := theta1 + float64(i)*segmentAngle
		endAngle := startAngle + segmentAngle

		arcSegmentToBezier(path, cx, cy, rx, ry, phi, startAngle, endAngle)
	}
}

func arcSegmentToBezier(path *vector.Path, cx, cy, rx, ry, phi, startAngle, endAngle float64) {
	// Compute control points
	alpha := math.Sin(endAngle-startAngle) * (math.Sqrt(4+3*math.Pow(math.Tan((endAngle-startAngle)/2), 2)) - 1) / 3

	cosPhi := math.Cos(phi)
	sinPhi := math.Sin(phi)

	cosStart := math.Cos(startAngle)
	sinStart := math.Sin(startAngle)
	cosEnd := math.Cos(endAngle)
	sinEnd := math.Sin(endAngle)

	// Start point
	x1 := cx + rx*cosPhi*cosStart - ry*sinPhi*sinStart
	y1 := cy + rx*sinPhi*cosStart + ry*cosPhi*sinStart

	// Derivatives at start
	dx1 := -rx*cosPhi*sinStart - ry*sinPhi*cosStart
	dy1 := -rx*sinPhi*sinStart + ry*cosPhi*cosStart

	// End point
	x2 := cx + rx*cosPhi*cosEnd - ry*sinPhi*sinEnd
	y2 := cy + rx*sinPhi*cosEnd + ry*cosPhi*sinEnd

	// Derivatives at end
	dx2 := -rx*cosPhi*sinEnd - ry*sinPhi*cosEnd
	dy2 := -rx*sinPhi*sinEnd + ry*cosPhi*cosEnd

	// Control points
	cp1x := x1 + alpha*dx1
	cp1y := y1 + alpha*dy1
	cp2x := x2 - alpha*dx2
	cp2y := y2 - alpha*dy2

	path.CubicTo(float32(cp1x), float32(cp1y), float32(cp2x), float32(cp2y), float32(x2), float32(y2))
}

func angle(ux, uy, vx, vy float64) float64 {
	n := math.Sqrt(ux*ux+uy*uy) * math.Sqrt(vx*vx+vy*vy)
	if n == 0 {
		return 0
	}
	c := (ux*vx + uy*vy) / n
	c = math.Min(1, math.Max(-1, c))
	angle := math.Acos(c)
	if ux*vy-uy*vx < 0 {
		angle = -angle
	}
	return angle
}

// Common SVG icon path data for convenience
var CommonIcons = map[string]string{
	// Navigation
	"arrow-left":    "M15 18l-6-6 6-6",
	"arrow-right":   "M9 18l6-6-6-6",
	"arrow-up":      "M18 15l-6-6-6 6",
	"arrow-down":    "M6 9l6 6 6-6",
	"chevron-left":  "M15 18l-6-6 6-6",
	"chevron-right": "M9 18l6-6-6-6",

	// Actions
	"check":    "M20 6L9 17l-5-5",
	"x":        "M18 6L6 18M6 6l12 12",
	"plus":     "M12 5v14M5 12h14",
	"minus":    "M5 12h14",
	"menu":     "M3 12h18M3 6h18M3 18h18",
	"search":   "M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z",
	"settings": "M12 15a3 3 0 100-6 3 3 0 000 6z",

	// Media
	"play":   "M5 3l14 9-14 9V3z",
	"pause":  "M6 4h4v16H6V4zm8 0h4v16h-4V4z",
	"stop":   "M6 6h12v12H6V6z",
	"volume": "M11 5L6 9H2v6h4l5 4V5z",

	// UI
	"home":     "M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6",
	"user":     "M20 21v-2a4 4 0 00-4-4H8a4 4 0 00-4 4v2M12 11a4 4 0 100-8 4 4 0 000 8z",
	"heart":    "M20.84 4.61a5.5 5.5 0 00-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 00-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 000-7.78z",
	"star":     "M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z",
	"bookmark": "M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z",

	// File
	"file":   "M13 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V9l-7-7z",
	"folder": "M22 19a2 2 0 01-2 2H4a2 2 0 01-2-2V5a2 2 0 012-2h5l2 3h9a2 2 0 012 2v11z",
	"trash":  "M3 6h18M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2",
	"edit":   "M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z",

	// Status
	"info":    "M12 16v-4m0-4h.01M22 12a10 10 0 11-20 0 10 10 0 0120 0z",
	"warning": "M12 9v2m0 4h.01M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z",
	"error":   "M12 8v4m0 4h.01M22 12a10 10 0 11-20 0 10 10 0 0120 0z",
	"success": "M9 12l2 2 4-4m6 2a10 10 0 11-20 0 10 10 0 0120 0z",

	// Game-specific
	"sword":  "M14.5 5.5L18 9l-8 8-4-4 8-8zM5 21l3-3",
	"shield": "M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z",
	"coin":   "M12 22a10 10 0 100-20 10 10 0 000 20zM12 6v12M8 12h8",
	"potion": "M9 3h6v2h-6V3zM7 12v6a2 2 0 002 2h6a2 2 0 002-2v-6l-2-5H9l-2 5z",
	"gem":    "M12 2L2 7l10 15 10-15-10-5zM2 7h20",
}

// CreateIconPath creates a simple icon path with stroke-only style
func CreateIconPath(iconName string, stroke color.Color, strokeWidth float64) *SVGPath {
	d, ok := CommonIcons[iconName]
	if !ok {
		return nil
	}
	return &SVGPath{
		D:           d,
		Fill:        nil,
		Stroke:      stroke,
		StrokeWidth: strokeWidth,
		Opacity:     1,
	}
}

// CreateIconSVG creates a complete SVG document for an icon
func CreateIconSVG(iconName string, size float64, stroke color.Color, strokeWidth float64) *SVGDocument {
	path := CreateIconPath(iconName, stroke, strokeWidth)
	if path == nil {
		return nil
	}
	return &SVGDocument{
		Width:    size,
		Height:   size,
		ViewBox:  ViewBox{MinX: 0, MinY: 0, Width: 24, Height: 24},
		Elements: []SVGElement{path},
	}
}
