package ui

import (
	"image/color"
	"math"
	"time"
)

// ============================================================================
// CSS Transition Engine
// ============================================================================

// TransitionEngine handles CSS-like property transitions between widget states.
// It interpolates float64 properties (opacity, borderRadius, etc.) and
// color.Color properties (backgroundColor, borderColor, textColor) over time.
// When a state change occurs mid-transition, the current interpolated value
// becomes the new start value for a smooth reversal (matching CSS behavior).
type TransitionEngine struct {
	floatTransitions map[string]*floatTransition
	colorTransitions map[string]*colorTransition
}

// floatTransition represents an active transition for a single float64 property.
type floatTransition struct {
	startValue float64
	endValue   float64
	startTime  time.Time
	duration   time.Duration
	delay      time.Duration
	easing     EasingFunc
}

// colorTransition represents an active transition for a single color property.
type colorTransition struct {
	startR, startG, startB, startA float64 // NRGBA [0,1]
	endR, endG, endB, endA         float64 // NRGBA [0,1]
	startTime                      time.Time
	duration                       time.Duration
	delay                          time.Duration
	easing                         EasingFunc
}

// NewTransitionEngine creates a new TransitionEngine.
func NewTransitionEngine() *TransitionEngine {
	return &TransitionEngine{
		floatTransitions: make(map[string]*floatTransition),
		colorTransitions: make(map[string]*colorTransition),
	}
}

// ============================================================================
// Property Mappings
// ============================================================================

// cssPropertyNormalize maps CSS property names (kebab-case and camelCase)
// to their normalized internal keys.
var cssPropertyNormalize = map[string]string{
	// Float properties
	"opacity":                    "opacity",
	"border-radius":              "borderRadius",
	"borderRadius":               "borderRadius",
	"border-width":               "borderWidth",
	"borderWidth":                "borderWidth",
	"font-size":                  "fontSize",
	"fontSize":                   "fontSize",
	"width":                      "width",
	"height":                     "height",
	"gap":                        "gap",
	"top":                        "top",
	"right":                      "right",
	"bottom":                     "bottom",
	"left":                       "left",
	"border-top-width":           "borderTopWidth",
	"borderTopWidth":             "borderTopWidth",
	"border-right-width":         "borderRightWidth",
	"borderRightWidth":           "borderRightWidth",
	"border-bottom-width":        "borderBottomWidth",
	"borderBottomWidth":          "borderBottomWidth",
	"border-left-width":          "borderLeftWidth",
	"borderLeftWidth":            "borderLeftWidth",
	"border-top-left-radius":     "borderTopLeftRadius",
	"borderTopLeftRadius":        "borderTopLeftRadius",
	"border-top-right-radius":    "borderTopRightRadius",
	"borderTopRightRadius":       "borderTopRightRadius",
	"border-bottom-left-radius":  "borderBottomLeftRadius",
	"borderBottomLeftRadius":     "borderBottomLeftRadius",
	"border-bottom-right-radius": "borderBottomRightRadius",
	"borderBottomRightRadius":    "borderBottomRightRadius",
	"letter-spacing":             "letterSpacing",
	"letterSpacing":              "letterSpacing",
	"line-height":                "lineHeight",
	"lineHeight":                 "lineHeight",
	"padding-top":                "paddingTop",
	"paddingTop":                 "paddingTop",
	"padding-right":              "paddingRight",
	"paddingRight":               "paddingRight",
	"padding-bottom":             "paddingBottom",
	"paddingBottom":              "paddingBottom",
	"padding-left":               "paddingLeft",
	"paddingLeft":                "paddingLeft",
	"margin-top":                 "marginTop",
	"marginTop":                  "marginTop",
	"margin-right":               "marginRight",
	"marginRight":                "marginRight",
	"margin-bottom":              "marginBottom",
	"marginBottom":               "marginBottom",
	"margin-left":                "marginLeft",
	"marginLeft":                 "marginLeft",
	"outline-offset":             "outlineOffset",
	"outlineOffset":              "outlineOffset",

	// Color properties
	"background":       "backgroundColor",
	"background-color": "backgroundColor",
	"backgroundColor":  "backgroundColor",
	"border-color":     "borderColor",
	"borderColor":      "borderColor",
	"color":            "textColor",
	"text-color":       "textColor",
	"textColor":        "textColor",
}

// allFloatProperties lists all transitionable float64 properties.
var allFloatProperties = []string{
	"opacity", "borderRadius", "borderWidth", "fontSize",
	"width", "height", "gap", "top", "right", "bottom", "left",
	"borderTopWidth", "borderRightWidth", "borderBottomWidth", "borderLeftWidth",
	"borderTopLeftRadius", "borderTopRightRadius", "borderBottomLeftRadius", "borderBottomRightRadius",
	"letterSpacing", "lineHeight", "outlineOffset",
	"paddingTop", "paddingRight", "paddingBottom", "paddingLeft",
	"marginTop", "marginRight", "marginBottom", "marginLeft",
}

// allColorProperties lists all transitionable color properties.
var allColorProperties = []string{
	"backgroundColor", "borderColor", "textColor",
}

// colorPropertySet is a lookup set for color properties.
var colorPropertySet = map[string]bool{
	"backgroundColor": true,
	"borderColor":     true,
	"textColor":       true,
}

// ============================================================================
// Property Getters and Setters
// ============================================================================

// getFloatValue extracts a float64 property from a Style by normalized name.
func getFloatValue(s *Style, prop string) float64 {
	switch prop {
	case "opacity":
		return s.Opacity
	case "borderRadius":
		return s.BorderRadius
	case "borderWidth":
		return s.BorderWidth
	case "fontSize":
		return s.FontSize
	case "width":
		return s.Width
	case "height":
		return s.Height
	case "gap":
		return s.Gap
	case "top":
		return s.Top
	case "right":
		return s.Right
	case "bottom":
		return s.Bottom
	case "left":
		return s.Left
	case "borderTopWidth":
		return s.BorderTopWidth
	case "borderRightWidth":
		return s.BorderRightWidth
	case "borderBottomWidth":
		return s.BorderBottomWidth
	case "borderLeftWidth":
		return s.BorderLeftWidth
	case "borderTopLeftRadius":
		return s.BorderTopLeftRadius
	case "borderTopRightRadius":
		return s.BorderTopRightRadius
	case "borderBottomLeftRadius":
		return s.BorderBottomLeftRadius
	case "borderBottomRightRadius":
		return s.BorderBottomRightRadius
	case "letterSpacing":
		return s.LetterSpacing
	case "lineHeight":
		return s.LineHeight
	case "outlineOffset":
		return s.OutlineOffset
	case "paddingTop":
		return s.Padding.Top
	case "paddingRight":
		return s.Padding.Right
	case "paddingBottom":
		return s.Padding.Bottom
	case "paddingLeft":
		return s.Padding.Left
	case "marginTop":
		return s.Margin.Top
	case "marginRight":
		return s.Margin.Right
	case "marginBottom":
		return s.Margin.Bottom
	case "marginLeft":
		return s.Margin.Left
	}
	return 0
}

// setFloatValue sets a float64 property on a Style by normalized name.
func setFloatValue(s *Style, prop string, v float64) {
	switch prop {
	case "opacity":
		s.Opacity = v
	case "borderRadius":
		s.BorderRadius = v
	case "borderWidth":
		s.BorderWidth = v
	case "fontSize":
		s.FontSize = v
	case "width":
		s.Width = v
	case "height":
		s.Height = v
	case "gap":
		s.Gap = v
	case "top":
		s.Top = v
	case "right":
		s.Right = v
	case "bottom":
		s.Bottom = v
	case "left":
		s.Left = v
	case "borderTopWidth":
		s.BorderTopWidth = v
	case "borderRightWidth":
		s.BorderRightWidth = v
	case "borderBottomWidth":
		s.BorderBottomWidth = v
	case "borderLeftWidth":
		s.BorderLeftWidth = v
	case "borderTopLeftRadius":
		s.BorderTopLeftRadius = v
	case "borderTopRightRadius":
		s.BorderTopRightRadius = v
	case "borderBottomLeftRadius":
		s.BorderBottomLeftRadius = v
	case "borderBottomRightRadius":
		s.BorderBottomRightRadius = v
	case "letterSpacing":
		s.LetterSpacing = v
	case "lineHeight":
		s.LineHeight = v
	case "outlineOffset":
		s.OutlineOffset = v
	case "paddingTop":
		s.Padding.Top = v
	case "paddingRight":
		s.Padding.Right = v
	case "paddingBottom":
		s.Padding.Bottom = v
	case "paddingLeft":
		s.Padding.Left = v
	case "marginTop":
		s.Margin.Top = v
	case "marginRight":
		s.Margin.Right = v
	case "marginBottom":
		s.Margin.Bottom = v
	case "marginLeft":
		s.Margin.Left = v
	}
}

// getColorValue extracts a color.Color property from a Style by normalized name.
func getColorValue(s *Style, prop string) color.Color {
	switch prop {
	case "backgroundColor":
		return s.BackgroundColor
	case "borderColor":
		return s.BorderColor
	case "textColor":
		return s.TextColor
	}
	return nil
}

// setColorValue sets a color.Color property on a Style by normalized name.
func setColorValue(s *Style, prop string, c color.Color) {
	switch prop {
	case "backgroundColor":
		s.BackgroundColor = c
	case "borderColor":
		s.BorderColor = c
	case "textColor":
		s.TextColor = c
	}
}

// ============================================================================
// Color Helpers
// ============================================================================

// colorToNRGBA converts a color.Color to normalized [0,1] NRGBA (non-premultiplied) components.
// nil colors are treated as transparent black (0,0,0,0).
func colorToNRGBA(c color.Color) (r, g, b, a float64) {
	if c == nil {
		return 0, 0, 0, 0
	}
	nrgba := color.NRGBAModel.Convert(c).(color.NRGBA)
	return float64(nrgba.R) / 255, float64(nrgba.G) / 255, float64(nrgba.B) / 255, float64(nrgba.A) / 255
}

// nrgbaToColor creates a color.NRGBA from normalized [0,1] components.
func nrgbaToColor(r, g, b, a float64) color.Color {
	return color.NRGBA{
		R: uint8(transClamp(r*255, 0, 255)),
		G: uint8(transClamp(g*255, 0, 255)),
		B: uint8(transClamp(b*255, 0, 255)),
		A: uint8(transClamp(a*255, 0, 255)),
	}
}

// transClamp clamps a value to [min, max].
func transClamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// colorsEqual checks if two colors are visually identical.
func colorsEqual(a, b color.Color) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	ar, ag, ab, aa := a.RGBA()
	br, bg, bb, ba := b.RGBA()
	return ar == br && ag == bg && ab == bb && aa == ba
}

// ============================================================================
// Cubic Bezier Easing
// ============================================================================

// CubicBezierEasing creates an easing function from CSS cubic-bezier(x1, y1, x2, y2).
// Uses Newton-Raphson iteration to solve the parametric curve.
func CubicBezierEasing(x1, y1, x2, y2 float64) EasingFunc {
	return func(x float64) float64 {
		if x <= 0 {
			return 0
		}
		if x >= 1 {
			return 1
		}
		t := solveCubicBezierX(x, x1, x2)
		return evalCubicBezier(t, y1, y2)
	}
}

// evalCubicBezier evaluates a 1D cubic bezier at parameter t with
// control points (0, p1, p2, 1).
// B(t) = 3(1-t)²t·p1 + 3(1-t)t²·p2 + t³
func evalCubicBezier(t, p1, p2 float64) float64 {
	u := 1 - t
	return 3*u*u*t*p1 + 3*u*t*t*p2 + t*t*t
}

// evalCubicBezierDerivative evaluates the derivative of a 1D cubic bezier.
// B'(t) = 3(1-t)²·p1 + 6(1-t)t·(p2-p1) + 3t²·(1-p2)
func evalCubicBezierDerivative(t, p1, p2 float64) float64 {
	u := 1 - t
	return 3*u*u*p1 + 6*u*t*(p2-p1) + 3*t*t*(1-p2)
}

// solveCubicBezierX finds parameter t for a given x-coordinate using Newton-Raphson.
func solveCubicBezierX(x, x1, x2 float64) float64 {
	t := x // initial guess
	for i := 0; i < 8; i++ {
		xEst := evalCubicBezier(t, x1, x2)
		dx := xEst - x
		if math.Abs(dx) < 1e-7 {
			break
		}
		dxdt := evalCubicBezierDerivative(t, x1, x2)
		if math.Abs(dxdt) < 1e-7 {
			break
		}
		t -= dx / dxdt
		t = transClamp(t, 0, 1)
	}
	return t
}

// Standard CSS easing functions as cubic-bezier curves.
var (
	// EaseCSSEase is the CSS "ease" function: cubic-bezier(0.25, 0.1, 0.25, 1.0)
	EaseCSSEase = CubicBezierEasing(0.25, 0.1, 0.25, 1.0)

	// EaseCSSEaseIn is the CSS "ease-in" function: cubic-bezier(0.42, 0, 1.0, 1.0)
	EaseCSSEaseIn = CubicBezierEasing(0.42, 0, 1.0, 1.0)

	// EaseCSSEaseOut is the CSS "ease-out" function: cubic-bezier(0, 0, 0.58, 1.0)
	EaseCSSEaseOut = CubicBezierEasing(0, 0, 0.58, 1.0)

	// EaseCSSEaseInOut is the CSS "ease-in-out" function: cubic-bezier(0.42, 0, 0.58, 1.0)
	EaseCSSEaseInOut = CubicBezierEasing(0.42, 0, 0.58, 1.0)
)

// ============================================================================
// Transition Lifecycle
// ============================================================================

// StartTransitions detects property changes between oldStyle and newStyle,
// and starts transitions for properties declared in the transition list.
// If a transition is already in progress for a property, the current
// interpolated value becomes the new start value (smooth reversal).
func (te *TransitionEngine) StartTransitions(oldStyle, newStyle *Style, declarations []Transition) {
	if len(declarations) == 0 {
		return
	}

	now := time.Now()

	for _, decl := range declarations {
		dur := time.Duration(decl.Duration * float64(time.Second))
		delay := time.Duration(decl.Delay * float64(time.Second))
		easing := decl.Easing
		if easing == nil {
			easing = EaseCSSEase // CSS default
		}
		if dur <= 0 {
			continue
		}

		if decl.Property == "all" {
			// Transition all transitionable properties
			for _, prop := range allFloatProperties {
				te.maybeStartFloat(prop, oldStyle, newStyle, now, dur, delay, easing)
			}
			for _, prop := range allColorProperties {
				te.maybeStartColor(prop, oldStyle, newStyle, now, dur, delay, easing)
			}
			continue
		}

		// Normalize the property name
		norm := decl.Property
		if n, ok := cssPropertyNormalize[decl.Property]; ok {
			norm = n
		}

		if colorPropertySet[norm] {
			te.maybeStartColor(norm, oldStyle, newStyle, now, dur, delay, easing)
		} else {
			te.maybeStartFloat(norm, oldStyle, newStyle, now, dur, delay, easing)
		}

		// Expand shorthand properties
		te.expandShorthand(decl.Property, oldStyle, newStyle, now, dur, delay, easing)
	}
}

// expandShorthand handles CSS shorthand properties that map to multiple sub-properties.
func (te *TransitionEngine) expandShorthand(prop string, oldStyle, newStyle *Style, now time.Time, dur, delay time.Duration, easing EasingFunc) {
	switch prop {
	case "padding":
		for _, sub := range []string{"paddingTop", "paddingRight", "paddingBottom", "paddingLeft"} {
			te.maybeStartFloat(sub, oldStyle, newStyle, now, dur, delay, easing)
		}
	case "margin":
		for _, sub := range []string{"marginTop", "marginRight", "marginBottom", "marginLeft"} {
			te.maybeStartFloat(sub, oldStyle, newStyle, now, dur, delay, easing)
		}
	case "border-radius", "borderRadius":
		for _, sub := range []string{"borderTopLeftRadius", "borderTopRightRadius", "borderBottomLeftRadius", "borderBottomRightRadius"} {
			te.maybeStartFloat(sub, oldStyle, newStyle, now, dur, delay, easing)
		}
	case "border-width", "borderWidth":
		for _, sub := range []string{"borderTopWidth", "borderRightWidth", "borderBottomWidth", "borderLeftWidth"} {
			te.maybeStartFloat(sub, oldStyle, newStyle, now, dur, delay, easing)
		}
	}
}

// maybeStartFloat starts a float transition if the property value changed.
// If a transition is already in progress, uses the current interpolated value as start.
func (te *TransitionEngine) maybeStartFloat(prop string, oldStyle, newStyle *Style, now time.Time, dur, delay time.Duration, easing EasingFunc) {
	oldVal := getFloatValue(oldStyle, prop)
	newVal := getFloatValue(newStyle, prop)

	// If there's an active transition, use its current interpolated value as start
	if existing, ok := te.floatTransitions[prop]; ok {
		currentVal := te.interpolateFloat(existing, now)
		if currentVal != nil {
			oldVal = *currentVal
		}
	}

	// Only start if values differ
	if oldVal == newVal {
		delete(te.floatTransitions, prop)
		return
	}

	te.floatTransitions[prop] = &floatTransition{
		startValue: oldVal,
		endValue:   newVal,
		startTime:  now,
		duration:   dur,
		delay:      delay,
		easing:     easing,
	}
}

// maybeStartColor starts a color transition if the property value changed.
// If a transition is already in progress, uses the current interpolated value as start.
func (te *TransitionEngine) maybeStartColor(prop string, oldStyle, newStyle *Style, now time.Time, dur, delay time.Duration, easing EasingFunc) {
	oldColor := getColorValue(oldStyle, prop)
	newColor := getColorValue(newStyle, prop)

	// If colors are equal, no transition needed
	if colorsEqual(oldColor, newColor) {
		delete(te.colorTransitions, prop)
		return
	}

	or, og, ob, oa := colorToNRGBA(oldColor)
	nr, ng, nb, na := colorToNRGBA(newColor)

	// If there's an active transition, use its current interpolated value as start
	if existing, ok := te.colorTransitions[prop]; ok {
		cr, cg, cb, ca := te.interpolateColor(existing, now)
		if cr != nil {
			or, og, ob, oa = *cr, *cg, *cb, *ca
		}
	}

	te.colorTransitions[prop] = &colorTransition{
		startR: or, startG: og, startB: ob, startA: oa,
		endR: nr, endG: ng, endB: nb, endA: na,
		startTime: now,
		duration:  dur,
		delay:     delay,
		easing:    easing,
	}
}

// ============================================================================
// Interpolation
// ============================================================================

// interpolateFloat returns the current interpolated value for a float transition.
// Returns nil if the transition has completed.
func (te *TransitionEngine) interpolateFloat(t *floatTransition, now time.Time) *float64 {
	elapsed := now.Sub(t.startTime) - t.delay
	if elapsed < 0 {
		// Still in delay period — return start value
		v := t.startValue
		return &v
	}

	progress := float64(elapsed) / float64(t.duration)
	if progress >= 1 {
		return nil // completed
	}

	if t.easing != nil {
		progress = t.easing(progress)
	}

	v := t.startValue + (t.endValue-t.startValue)*progress
	return &v
}

// interpolateColor returns the current interpolated NRGBA components for a color transition.
// Returns nil values if the transition has completed.
func (te *TransitionEngine) interpolateColor(t *colorTransition, now time.Time) (r, g, b, a *float64) {
	elapsed := now.Sub(t.startTime) - t.delay
	if elapsed < 0 {
		// Still in delay period — return start value
		return &t.startR, &t.startG, &t.startB, &t.startA
	}

	progress := float64(elapsed) / float64(t.duration)
	if progress >= 1 {
		return nil, nil, nil, nil // completed
	}

	if t.easing != nil {
		progress = t.easing(progress)
	}

	rr := t.startR + (t.endR-t.startR)*progress
	gg := t.startG + (t.endG-t.startG)*progress
	bb := t.startB + (t.endB-t.startB)*progress
	aa := t.startA + (t.endA-t.startA)*progress
	return &rr, &gg, &bb, &aa
}

// Apply applies active transitions to a style, returning a new style with
// interpolated values. If no transitions are active, returns the input style unchanged.
func (te *TransitionEngine) Apply(style *Style) *Style {
	if !te.IsActive() {
		return style
	}

	result := style.Clone()
	now := time.Now()
	var completedFloats []string
	var completedColors []string

	// Apply float transitions
	for prop, t := range te.floatTransitions {
		v := te.interpolateFloat(t, now)
		if v != nil {
			setFloatValue(result, prop, *v)
		} else {
			// Transition completed — use end value and mark for cleanup
			setFloatValue(result, prop, t.endValue)
			completedFloats = append(completedFloats, prop)
		}
	}

	// Apply color transitions
	for prop, t := range te.colorTransitions {
		r, g, b, a := te.interpolateColor(t, now)
		if r != nil {
			setColorValue(result, prop, nrgbaToColor(*r, *g, *b, *a))
		} else {
			// Transition completed — use end value and mark for cleanup
			setColorValue(result, prop, nrgbaToColor(t.endR, t.endG, t.endB, t.endA))
			completedColors = append(completedColors, prop)
		}
	}

	// Cleanup completed transitions
	for _, prop := range completedFloats {
		delete(te.floatTransitions, prop)
	}
	for _, prop := range completedColors {
		delete(te.colorTransitions, prop)
	}

	return result
}

// IsActive returns true if any transition is currently in progress.
func (te *TransitionEngine) IsActive() bool {
	return len(te.floatTransitions) > 0 || len(te.colorTransitions) > 0
}
