package ui

import (
	"regexp"
	"strconv"
	"strings"
)

// ============================================================================
// CSS Variables System
// ============================================================================

// CSSVariables holds CSS custom properties (variables)
type CSSVariables struct {
	vars map[string]string
}

// NewCSSVariables creates a new CSS variables container
func NewCSSVariables() *CSSVariables {
	return &CSSVariables{
		vars: make(map[string]string),
	}
}

// Set sets a CSS variable
func (v *CSSVariables) Set(name, value string) {
	// Ensure name starts with --
	if !strings.HasPrefix(name, "--") {
		name = "--" + name
	}
	v.vars[name] = value
}

// Get retrieves a CSS variable value
func (v *CSSVariables) Get(name string) string {
	if !strings.HasPrefix(name, "--") {
		name = "--" + name
	}
	return v.vars[name]
}

// GetWithFallback retrieves a CSS variable with a fallback value
func (v *CSSVariables) GetWithFallback(name, fallback string) string {
	if val := v.Get(name); val != "" {
		return val
	}
	return fallback
}

// Delete removes a CSS variable
func (v *CSSVariables) Delete(name string) {
	if !strings.HasPrefix(name, "--") {
		name = "--" + name
	}
	delete(v.vars, name)
}

// Resolve resolves var() references in a string
func (v *CSSVariables) Resolve(s string) string {
	// Match var(--name) or var(--name, fallback)
	re := regexp.MustCompile(`var\(\s*(--[\w-]+)\s*(?:,\s*([^)]+))?\)`)

	return re.ReplaceAllStringFunc(s, func(match string) string {
		submatches := re.FindStringSubmatch(match)
		if len(submatches) >= 2 {
			varName := submatches[1]
			fallback := ""
			if len(submatches) >= 3 {
				fallback = strings.TrimSpace(submatches[2])
			}
			if val := v.Get(varName); val != "" {
				return val
			}
			return fallback
		}
		return match
	})
}

// Clone creates a copy of the variables
func (v *CSSVariables) Clone() *CSSVariables {
	clone := NewCSSVariables()
	for k, val := range v.vars {
		clone.vars[k] = val
	}
	return clone
}

// Merge merges another set of variables (other takes precedence)
func (v *CSSVariables) Merge(other *CSSVariables) {
	if other == nil {
		return
	}
	for k, val := range other.vars {
		v.vars[k] = val
	}
}

// ============================================================================
// % Unit Support
// ============================================================================

// SizeValue represents a size that can be px, %, vh, vw, etc.
type SizeValue struct {
	Value float64
	Unit  SizeUnit
}

// SizeUnit represents the unit of a size value
type SizeUnit int

const (
	UnitPx      SizeUnit = iota // Pixels (default)
	UnitPercent                 // Percentage of parent
	UnitVw                      // Viewport width
	UnitVh                      // Viewport height
	UnitEm                      // Font size
	UnitRem                     // Root font size
	UnitAuto                    // Auto sizing
)

// ParseSizeValue parses a size value string (e.g., "50%", "100px", "10vh")
func ParseSizeValue(s string) SizeValue {
	s = strings.TrimSpace(s)

	if s == "auto" {
		return SizeValue{Unit: UnitAuto}
	}

	// Check for units
	if strings.HasSuffix(s, "%") {
		val, _ := strconv.ParseFloat(strings.TrimSuffix(s, "%"), 64)
		return SizeValue{Value: val, Unit: UnitPercent}
	}
	if strings.HasSuffix(s, "vw") {
		val, _ := strconv.ParseFloat(strings.TrimSuffix(s, "vw"), 64)
		return SizeValue{Value: val, Unit: UnitVw}
	}
	if strings.HasSuffix(s, "vh") {
		val, _ := strconv.ParseFloat(strings.TrimSuffix(s, "vh"), 64)
		return SizeValue{Value: val, Unit: UnitVh}
	}
	if strings.HasSuffix(s, "rem") {
		val, _ := strconv.ParseFloat(strings.TrimSuffix(s, "rem"), 64)
		return SizeValue{Value: val, Unit: UnitRem}
	}
	if strings.HasSuffix(s, "em") {
		val, _ := strconv.ParseFloat(strings.TrimSuffix(s, "em"), 64)
		return SizeValue{Value: val, Unit: UnitEm}
	}
	if strings.HasSuffix(s, "px") {
		val, _ := strconv.ParseFloat(strings.TrimSuffix(s, "px"), 64)
		return SizeValue{Value: val, Unit: UnitPx}
	}

	// Default: treat as pixels
	val, _ := strconv.ParseFloat(s, 64)
	return SizeValue{Value: val, Unit: UnitPx}
}

// Resolve resolves the size value to pixels
func (sv SizeValue) Resolve(context SizeContext) float64 {
	switch sv.Unit {
	case UnitPx:
		return sv.Value
	case UnitPercent:
		return sv.Value / 100 * context.ParentSize
	case UnitVw:
		return sv.Value / 100 * context.ViewportWidth
	case UnitVh:
		return sv.Value / 100 * context.ViewportHeight
	case UnitEm:
		return sv.Value * context.FontSize
	case UnitRem:
		return sv.Value * context.RootFontSize
	case UnitAuto:
		return 0 // Auto is handled specially by layout
	default:
		return sv.Value
	}
}

// IsAuto checks if the size is auto
func (sv SizeValue) IsAuto() bool {
	return sv.Unit == UnitAuto
}

// SizeContext provides context for resolving relative sizes
type SizeContext struct {
	ParentSize     float64 // Parent dimension (width or height)
	ViewportWidth  float64
	ViewportHeight float64
	FontSize       float64
	RootFontSize   float64
}

// ============================================================================
// Extended Style with CSS Variables and Relative Units
// ============================================================================

// ExtendedStyle adds CSS variable and relative unit support
type ExtendedStyle struct {
	*Style

	// Raw string values (may contain var() or relative units)
	WidthRaw         string `json:"widthRaw"`
	HeightRaw        string `json:"heightRaw"`
	MinWidthRaw      string `json:"minWidthRaw"`
	MinHeightRaw     string `json:"minHeightRaw"`
	MaxWidthRaw      string `json:"maxWidthRaw"`
	MaxHeightRaw     string `json:"maxHeightRaw"`
	PaddingTopRaw    string `json:"paddingTopRaw"`
	PaddingRightRaw  string `json:"paddingRightRaw"`
	PaddingBottomRaw string `json:"paddingBottomRaw"`
	PaddingLeftRaw   string `json:"paddingLeftRaw"`
	MarginTopRaw     string `json:"marginTopRaw"`
	MarginRightRaw   string `json:"marginRightRaw"`
	MarginBottomRaw  string `json:"marginBottomRaw"`
	MarginLeftRaw    string `json:"marginLeftRaw"`
	GapRaw           string `json:"gapRaw"`
	FontSizeRaw      string `json:"fontSizeRaw"`

	// Parsed size values
	parsedWidth     *SizeValue
	parsedHeight    *SizeValue
	parsedMinWidth  *SizeValue
	parsedMinHeight *SizeValue
	parsedMaxWidth  *SizeValue
	parsedMaxHeight *SizeValue
}

// ResolveWithContext resolves all relative values using the given context
func (es *ExtendedStyle) ResolveWithContext(ctx SizeContext, vars *CSSVariables) {
	if es.Style == nil {
		return
	}

	resolveSize := func(raw string) float64 {
		if raw == "" {
			return 0
		}
		resolved := vars.Resolve(raw)
		sv := ParseSizeValue(resolved)
		return sv.Resolve(ctx)
	}

	// Width/Height
	if es.WidthRaw != "" {
		es.Style.Width = resolveSize(es.WidthRaw)
	}
	if es.HeightRaw != "" {
		es.Style.Height = resolveSize(es.HeightRaw)
	}
	if es.MinWidthRaw != "" {
		es.Style.MinWidth = resolveSize(es.MinWidthRaw)
	}
	if es.MinHeightRaw != "" {
		es.Style.MinHeight = resolveSize(es.MinHeightRaw)
	}
	if es.MaxWidthRaw != "" {
		es.Style.MaxWidth = resolveSize(es.MaxWidthRaw)
	}
	if es.MaxHeightRaw != "" {
		es.Style.MaxHeight = resolveSize(es.MaxHeightRaw)
	}
	if es.GapRaw != "" {
		es.Style.Gap = resolveSize(es.GapRaw)
	}
	if es.FontSizeRaw != "" {
		es.Style.FontSize = resolveSize(es.FontSizeRaw)
	}

	// Padding
	if es.PaddingTopRaw != "" {
		es.Style.Padding.Top = resolveSize(es.PaddingTopRaw)
	}
	if es.PaddingRightRaw != "" {
		es.Style.Padding.Right = resolveSize(es.PaddingRightRaw)
	}
	if es.PaddingBottomRaw != "" {
		es.Style.Padding.Bottom = resolveSize(es.PaddingBottomRaw)
	}
	if es.PaddingLeftRaw != "" {
		es.Style.Padding.Left = resolveSize(es.PaddingLeftRaw)
	}

	// Margin
	if es.MarginTopRaw != "" {
		es.Style.Margin.Top = resolveSize(es.MarginTopRaw)
	}
	if es.MarginRightRaw != "" {
		es.Style.Margin.Right = resolveSize(es.MarginRightRaw)
	}
	if es.MarginBottomRaw != "" {
		es.Style.Margin.Bottom = resolveSize(es.MarginBottomRaw)
	}
	if es.MarginLeftRaw != "" {
		es.Style.Margin.Left = resolveSize(es.MarginLeftRaw)
	}

	// Colors with var() support
	if es.Style.Background != "" {
		es.Style.Background = vars.Resolve(es.Style.Background)
	}
	if es.Style.Color != "" {
		es.Style.Color = vars.Resolve(es.Style.Color)
	}
	if es.Style.Border != "" {
		es.Style.Border = vars.Resolve(es.Style.Border)
	}
}

// ============================================================================
// calc() Function Support
// ============================================================================

// CalcExpression represents a CSS calc() expression
type CalcExpression struct {
	original string
	tokens   []calcToken
}

type calcToken struct {
	value    float64
	unit     SizeUnit
	operator rune // +, -, *, /
}

// ParseCalc parses a calc() expression
func ParseCalc(s string) *CalcExpression {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "calc(") {
		return nil
	}

	s = strings.TrimPrefix(s, "calc(")
	s = strings.TrimSuffix(s, ")")

	return &CalcExpression{
		original: s,
		tokens:   tokenizeCalc(s),
	}
}

func tokenizeCalc(s string) []calcToken {
	tokens := make([]calcToken, 0)
	re := regexp.MustCompile(`([+-]?\s*[\d.]+)(px|%|vw|vh|em|rem)?|([+\-*/])`)

	matches := re.FindAllStringSubmatch(s, -1)
	for _, match := range matches {
		if match[3] != "" {
			// Operator
			tokens = append(tokens, calcToken{operator: rune(match[3][0])})
		} else if match[1] != "" {
			// Value with optional unit
			sv := ParseSizeValue(match[1] + match[2])
			tokens = append(tokens, calcToken{value: sv.Value, unit: sv.Unit})
		}
	}

	return tokens
}

// Resolve evaluates the calc expression
func (ce *CalcExpression) Resolve(ctx SizeContext) float64 {
	if ce == nil || len(ce.tokens) == 0 {
		return 0
	}

	// Simple evaluation (handles + and - for now)
	result := float64(0)
	operator := '+'

	for _, token := range ce.tokens {
		if token.operator != 0 {
			operator = token.operator
			continue
		}

		sv := SizeValue{Value: token.value, Unit: token.unit}
		val := sv.Resolve(ctx)

		switch operator {
		case '+':
			result += val
		case '-':
			result -= val
		case '*':
			result *= val
		case '/':
			if val != 0 {
				result /= val
			}
		}
	}

	return result
}
