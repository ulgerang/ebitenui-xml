package ui

import (
	"encoding/json"
	"fmt"
	"image/color"
	"math"
	"os"
	"strconv"
	"strings"
)

// ParseColor is the exported color parsing function for external use
func ParseColor(s string) color.Color {
	return parseColor(s)
}

// StyleSheet holds all style definitions
type StyleSheet struct {
	Styles map[string]*Style `json:"styles"`
}

// StyleEngine manages and applies styles
type StyleEngine struct {
	styles map[string]*Style
}

// NewStyleEngine creates a new style engine
func NewStyleEngine() *StyleEngine {
	return &StyleEngine{
		styles: make(map[string]*Style),
	}
}

// LoadFromFile loads styles from a JSON file
func (se *StyleEngine) LoadFromFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read style file: %w", err)
	}
	return se.LoadFromJSON(data)
}

// LoadFromJSON loads styles from JSON data
// Supports both flat format: { "#root": {...} } and nested format: { "styles": { "#root": {...} } }
func (se *StyleEngine) LoadFromJSON(data []byte) error {
	// Parse raw JSON to detect explicitly-set fields
	var rawMap map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMap); err != nil {
		return fmt.Errorf("failed to parse styles: %w", err)
	}

	// Determine the style map source
	var styleRawMap map[string]json.RawMessage

	// Check if it's nested format with "styles" key
	if stylesRaw, ok := rawMap["styles"]; ok {
		if err := json.Unmarshal(stylesRaw, &styleRawMap); err == nil && len(styleRawMap) > 0 {
			// nested format
		} else {
			styleRawMap = rawMap // fallback to flat
		}
	} else {
		styleRawMap = rawMap // flat format
	}

	for selector, rawStyle := range styleRawMap {
		var style Style
		if err := json.Unmarshal(rawStyle, &style); err != nil {
			continue
		}

		// Detect explicitly-set fields from raw JSON
		se.detectExplicitFields(&style, rawStyle)

		se.parseStyleColors(&style)
		se.styles[selector] = &style
	}

	return nil
}

// detectExplicitFields scans raw JSON to set flags for explicitly-specified fields
func (se *StyleEngine) detectExplicitFields(style *Style, rawJSON json.RawMessage) {
	var rawFields map[string]json.RawMessage
	if err := json.Unmarshal(rawJSON, &rawFields); err != nil {
		return
	}

	if _, ok := rawFields["padding"]; ok {
		style.PaddingSet = true
	}
	if _, ok := rawFields["margin"]; ok {
		style.MarginSet = true
	}
	if _, ok := rawFields["borderWidth"]; ok {
		style.BorderWidthSet = true
	}

	// Recursively handle state styles
	if hoverRaw, ok := rawFields["hover"]; ok && style.HoverStyle != nil {
		se.detectExplicitFields(style.HoverStyle, hoverRaw)
	}
	if activeRaw, ok := rawFields["active"]; ok && style.ActiveStyle != nil {
		se.detectExplicitFields(style.ActiveStyle, activeRaw)
	}
	if disabledRaw, ok := rawFields["disabled"]; ok && style.DisabledStyle != nil {
		se.detectExplicitFields(style.DisabledStyle, disabledRaw)
	}
	if focusRaw, ok := rawFields["focus"]; ok && style.FocusStyle != nil {
		se.detectExplicitFields(style.FocusStyle, focusRaw)
	}
}

// parseStyleColors recursively parses color strings in a style
func (se *StyleEngine) parseStyleColors(style *Style) {
	if style == nil {
		return
	}

	// Parse main colors
	if style.Background != "" {
		// Check if it's a gradient
		if strings.HasPrefix(style.Background, "linear-gradient") {
			style.parsedGradient = ParseGradient(style.Background)
		} else if strings.HasPrefix(style.Background, "radial-gradient") {
			style.parsedGradient = parseRadialGradientCSS(style.Background)
		} else {
			style.BackgroundColor = parseColor(style.Background)
		}
	}
	if style.Border != "" {
		style.BorderColor = parseColor(style.Border)
	}
	if style.Color != "" {
		style.TextColor = parseColor(style.Color)
	}

	// Parse box shadow
	if style.BoxShadow != "" {
		style.parsedBoxShadow = parseBoxShadow(style.BoxShadow)
	}

	// Parse transitions
	if style.Transition != "" {
		style.parsedTransitions = parseTransitions(style.Transition)
	}

	// Parse filter
	if style.Filter != "" {
		style.parsedFilter = ParseFilter(style.Filter)
	}

	// Parse backdrop-filter
	if style.BackdropFilter != "" {
		style.parsedBackdropFilter = ParseBackdropFilter(style.BackdropFilter)
	}

	// Parse state styles
	if style.HoverStyle != nil {
		se.parseStyleColors(style.HoverStyle)
	}
	if style.ActiveStyle != nil {
		se.parseStyleColors(style.ActiveStyle)
	}
	if style.DisabledStyle != nil {
		se.parseStyleColors(style.DisabledStyle)
	}
	if style.FocusStyle != nil {
		se.parseStyleColors(style.FocusStyle)
	}
}

// LoadFromString loads styles from a JSON string
func (se *StyleEngine) LoadFromString(s string) error {
	return se.LoadFromJSON([]byte(s))
}

// AddStyle adds a style for a selector
func (se *StyleEngine) AddStyle(selector string, style *Style) {
	se.parseStyleColors(style)
	se.styles[selector] = style
}

// GetStyle gets a style by selector
func (se *StyleEngine) GetStyle(selector string) *Style {
	return se.styles[selector]
}

// ApplyStyle applies a style to a widget by selector
func (se *StyleEngine) ApplyStyle(widget Widget, selector string) {
	style := se.styles[selector]
	if style == nil {
		return
	}

	// Merge with existing style
	existing := widget.Style()
	merged := mergeStylesFully(existing, style)
	widget.SetStyle(merged)
}

// mergeStylesFully merges all properties from src into dst
func mergeStylesFully(dst, src *Style) *Style {
	result := dst.Clone()
	result.Merge(src)
	return result
}

// parseBoxShadow parses a CSS box-shadow string
// Format: "offsetX offsetY blur spread color [inset]"
func parseBoxShadow(s string) *BoxShadow {
	s = strings.TrimSpace(s)
	if s == "" || s == "none" {
		return nil
	}

	bs := &BoxShadow{}
	parts := strings.Fields(s)

	if len(parts) < 3 {
		return nil
	}

	// Check for inset
	if parts[0] == "inset" {
		bs.Inset = true
		parts = parts[1:]
	}

	// Parse numeric values
	idx := 0
	if idx < len(parts) {
		bs.OffsetX = parsePixelValue(parts[idx])
		idx++
	}
	if idx < len(parts) {
		bs.OffsetY = parsePixelValue(parts[idx])
		idx++
	}
	if idx < len(parts) {
		bs.Blur = parsePixelValue(parts[idx])
		idx++
	}
	if idx < len(parts) {
		// This could be spread or color
		if isNumeric(parts[idx]) {
			bs.Spread = parsePixelValue(parts[idx])
			idx++
		}
	}

	// Rest is color
	if idx < len(parts) {
		colorStr := strings.Join(parts[idx:], " ")
		bs.Color = parseColor(colorStr)
	}

	if bs.Color == nil {
		bs.Color = color.RGBA{0, 0, 0, 128}
	}

	return bs
}

// parseTransitions parses CSS transition string
// Format: "property duration [easing] [delay], ..."
func parseTransitions(s string) []Transition {
	if s == "" || s == "none" {
		return nil
	}

	var transitions []Transition
	parts := strings.Split(s, ",")

	for _, part := range parts {
		t := parseTransition(strings.TrimSpace(part))
		if t.Property != "" {
			transitions = append(transitions, t)
		}
	}

	return transitions
}

func parseTransition(s string) Transition {
	t := Transition{Easing: EaseLinear}
	parts := strings.Fields(s)

	if len(parts) >= 1 {
		t.Property = parts[0]
	}
	if len(parts) >= 2 {
		t.Duration = parseDuration(parts[1])
	}
	if len(parts) >= 3 {
		t.Easing = ParseEasing(parts[2])
	}
	if len(parts) >= 4 {
		t.Delay = parseDuration(parts[3])
	}

	return t
}

// parseDuration parses a CSS duration (e.g., "0.3s", "300ms")
func parseDuration(s string) float64 {
	s = strings.ToLower(strings.TrimSpace(s))

	if strings.HasSuffix(s, "ms") {
		s = strings.TrimSuffix(s, "ms")
		f, _ := strconv.ParseFloat(s, 64)
		return f / 1000
	}

	if strings.HasSuffix(s, "s") {
		s = strings.TrimSuffix(s, "s")
		f, _ := strconv.ParseFloat(s, 64)
		return f
	}

	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// parsePixelValue parses a pixel value (e.g., "10px", "10")
func parsePixelValue(s string) float64 {
	s = strings.TrimSuffix(s, "px")
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// isNumeric checks if a string starts with a digit or minus
func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	c := s[0]
	return (c >= '0' && c <= '9') || c == '-' || c == '.'
}

// parseColor parses a color string (hex, rgb, rgba, hsl, hsla, or named)
func parseColor(s string) color.Color {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)

	// Named colors (extended)
	if clr := getNamedColor(s); clr != nil {
		return clr
	}

	// Hex color
	if strings.HasPrefix(s, "#") {
		return parseHexColor(s[1:])
	}

	// RGB/RGBA
	if strings.HasPrefix(s, "rgb") {
		return parseRGBColor(s)
	}

	// HSL/HSLA
	if strings.HasPrefix(s, "hsl") {
		return parseHSLColor(s)
	}

	return nil
}

// getNamedColor returns a color by CSS name
func getNamedColor(name string) color.Color {
	colors := map[string]color.RGBA{
		// Basic colors
		"black":       {0, 0, 0, 255},
		"white":       {255, 255, 255, 255},
		"red":         {255, 0, 0, 255},
		"green":       {0, 128, 0, 255},
		"blue":        {0, 0, 255, 255},
		"yellow":      {255, 255, 0, 255},
		"cyan":        {0, 255, 255, 255},
		"magenta":     {255, 0, 255, 255},
		"gray":        {128, 128, 128, 255},
		"grey":        {128, 128, 128, 255},
		"transparent": {0, 0, 0, 0},

		// Extended CSS colors
		"aliceblue":            {240, 248, 255, 255},
		"antiquewhite":         {250, 235, 215, 255},
		"aqua":                 {0, 255, 255, 255},
		"aquamarine":           {127, 255, 212, 255},
		"azure":                {240, 255, 255, 255},
		"beige":                {245, 245, 220, 255},
		"bisque":               {255, 228, 196, 255},
		"blanchedalmond":       {255, 235, 205, 255},
		"blueviolet":           {138, 43, 226, 255},
		"brown":                {165, 42, 42, 255},
		"burlywood":            {222, 184, 135, 255},
		"cadetblue":            {95, 158, 160, 255},
		"chartreuse":           {127, 255, 0, 255},
		"chocolate":            {210, 105, 30, 255},
		"coral":                {255, 127, 80, 255},
		"cornflowerblue":       {100, 149, 237, 255},
		"cornsilk":             {255, 248, 220, 255},
		"crimson":              {220, 20, 60, 255},
		"darkblue":             {0, 0, 139, 255},
		"darkcyan":             {0, 139, 139, 255},
		"darkgoldenrod":        {184, 134, 11, 255},
		"darkgray":             {169, 169, 169, 255},
		"darkgreen":            {0, 100, 0, 255},
		"darkkhaki":            {189, 183, 107, 255},
		"darkmagenta":          {139, 0, 139, 255},
		"darkolivegreen":       {85, 107, 47, 255},
		"darkorange":           {255, 140, 0, 255},
		"darkorchid":           {153, 50, 204, 255},
		"darkred":              {139, 0, 0, 255},
		"darksalmon":           {233, 150, 122, 255},
		"darkseagreen":         {143, 188, 143, 255},
		"darkslateblue":        {72, 61, 139, 255},
		"darkslategray":        {47, 79, 79, 255},
		"darkturquoise":        {0, 206, 209, 255},
		"darkviolet":           {148, 0, 211, 255},
		"deeppink":             {255, 20, 147, 255},
		"deepskyblue":          {0, 191, 255, 255},
		"dimgray":              {105, 105, 105, 255},
		"dodgerblue":           {30, 144, 255, 255},
		"firebrick":            {178, 34, 34, 255},
		"floralwhite":          {255, 250, 240, 255},
		"forestgreen":          {34, 139, 34, 255},
		"fuchsia":              {255, 0, 255, 255},
		"gainsboro":            {220, 220, 220, 255},
		"ghostwhite":           {248, 248, 255, 255},
		"gold":                 {255, 215, 0, 255},
		"goldenrod":            {218, 165, 32, 255},
		"greenyellow":          {173, 255, 47, 255},
		"honeydew":             {240, 255, 240, 255},
		"hotpink":              {255, 105, 180, 255},
		"indianred":            {205, 92, 92, 255},
		"indigo":               {75, 0, 130, 255},
		"ivory":                {255, 255, 240, 255},
		"khaki":                {240, 230, 140, 255},
		"lavender":             {230, 230, 250, 255},
		"lavenderblush":        {255, 240, 245, 255},
		"lawngreen":            {124, 252, 0, 255},
		"lemonchiffon":         {255, 250, 205, 255},
		"lightblue":            {173, 216, 230, 255},
		"lightcoral":           {240, 128, 128, 255},
		"lightcyan":            {224, 255, 255, 255},
		"lightgoldenrodyellow": {250, 250, 210, 255},
		"lightgray":            {211, 211, 211, 255},
		"lightgreen":           {144, 238, 144, 255},
		"lightpink":            {255, 182, 193, 255},
		"lightsalmon":          {255, 160, 122, 255},
		"lightseagreen":        {32, 178, 170, 255},
		"lightskyblue":         {135, 206, 250, 255},
		"lightslategray":       {119, 136, 153, 255},
		"lightsteelblue":       {176, 196, 222, 255},
		"lightyellow":          {255, 255, 224, 255},
		"lime":                 {0, 255, 0, 255},
		"limegreen":            {50, 205, 50, 255},
		"linen":                {250, 240, 230, 255},
		"maroon":               {128, 0, 0, 255},
		"mediumaquamarine":     {102, 205, 170, 255},
		"mediumblue":           {0, 0, 205, 255},
		"mediumorchid":         {186, 85, 211, 255},
		"mediumpurple":         {147, 112, 219, 255},
		"mediumseagreen":       {60, 179, 113, 255},
		"mediumslateblue":      {123, 104, 238, 255},
		"mediumspringgreen":    {0, 250, 154, 255},
		"mediumturquoise":      {72, 209, 204, 255},
		"mediumvioletred":      {199, 21, 133, 255},
		"midnightblue":         {25, 25, 112, 255},
		"mintcream":            {245, 255, 250, 255},
		"mistyrose":            {255, 228, 225, 255},
		"moccasin":             {255, 228, 181, 255},
		"navajowhite":          {255, 222, 173, 255},
		"navy":                 {0, 0, 128, 255},
		"oldlace":              {253, 245, 230, 255},
		"olive":                {128, 128, 0, 255},
		"olivedrab":            {107, 142, 35, 255},
		"orange":               {255, 165, 0, 255},
		"orangered":            {255, 69, 0, 255},
		"orchid":               {218, 112, 214, 255},
		"palegoldenrod":        {238, 232, 170, 255},
		"palegreen":            {152, 251, 152, 255},
		"paleturquoise":        {175, 238, 238, 255},
		"palevioletred":        {219, 112, 147, 255},
		"papayawhip":           {255, 239, 213, 255},
		"peachpuff":            {255, 218, 185, 255},
		"peru":                 {205, 133, 63, 255},
		"pink":                 {255, 192, 203, 255},
		"plum":                 {221, 160, 221, 255},
		"powderblue":           {176, 224, 230, 255},
		"purple":               {128, 0, 128, 255},
		"rebeccapurple":        {102, 51, 153, 255},
		"rosybrown":            {188, 143, 143, 255},
		"royalblue":            {65, 105, 225, 255},
		"saddlebrown":          {139, 69, 19, 255},
		"salmon":               {250, 128, 114, 255},
		"sandybrown":           {244, 164, 96, 255},
		"seagreen":             {46, 139, 87, 255},
		"seashell":             {255, 245, 238, 255},
		"sienna":               {160, 82, 45, 255},
		"silver":               {192, 192, 192, 255},
		"skyblue":              {135, 206, 235, 255},
		"slateblue":            {106, 90, 205, 255},
		"slategray":            {112, 128, 144, 255},
		"snow":                 {255, 250, 250, 255},
		"springgreen":          {0, 255, 127, 255},
		"steelblue":            {70, 130, 180, 255},
		"tan":                  {210, 180, 140, 255},
		"teal":                 {0, 128, 128, 255},
		"thistle":              {216, 191, 216, 255},
		"tomato":               {255, 99, 71, 255},
		"turquoise":            {64, 224, 208, 255},
		"violet":               {238, 130, 238, 255},
		"wheat":                {245, 222, 179, 255},
		"whitesmoke":           {245, 245, 245, 255},
		"yellowgreen":          {154, 205, 50, 255},
	}

	if clr, ok := colors[name]; ok {
		return clr
	}
	return nil
}

// parseHexColor parses a hex color string (without #)
func parseHexColor(s string) color.Color {
	var r, g, b, a uint8 = 0, 0, 0, 255

	switch len(s) {
	case 3: // #RGB
		r = parseHexByte(s[0:1]) * 17
		g = parseHexByte(s[1:2]) * 17
		b = parseHexByte(s[2:3]) * 17
	case 4: // #RGBA
		r = parseHexByte(s[0:1]) * 17
		g = parseHexByte(s[1:2]) * 17
		b = parseHexByte(s[2:3]) * 17
		a = parseHexByte(s[3:4]) * 17
	case 6: // #RRGGBB
		r = parseHexByte(s[0:2])
		g = parseHexByte(s[2:4])
		b = parseHexByte(s[4:6])
	case 8: // #RRGGBBAA
		r = parseHexByte(s[0:2])
		g = parseHexByte(s[2:4])
		b = parseHexByte(s[4:6])
		a = parseHexByte(s[6:8])
	}

	return color.RGBA{r, g, b, a}
}

// parseHexByte parses a hex string to uint8
func parseHexByte(s string) uint8 {
	val, _ := strconv.ParseUint(s, 16, 8)
	return uint8(val)
}

// parseRGBColor parses rgb() or rgba() color strings
func parseRGBColor(s string) color.Color {
	s = strings.TrimPrefix(s, "rgba(")
	s = strings.TrimPrefix(s, "rgb(")
	s = strings.TrimSuffix(s, ")")

	parts := strings.Split(s, ",")
	if len(parts) < 3 {
		return nil
	}

	r := parseColorValue(strings.TrimSpace(parts[0]))
	g := parseColorValue(strings.TrimSpace(parts[1]))
	b := parseColorValue(strings.TrimSpace(parts[2]))
	a := uint8(255)

	if len(parts) >= 4 {
		alpha := strings.TrimSpace(parts[3])
		if strings.Contains(alpha, ".") {
			f, _ := strconv.ParseFloat(alpha, 64)
			a = uint8(f * 255)
		} else {
			a = parseColorValue(alpha)
		}
	}

	return color.RGBA{r, g, b, a}
}

// parseHSLColor parses hsl() or hsla() color strings
func parseHSLColor(s string) color.Color {
	s = strings.TrimPrefix(s, "hsla(")
	s = strings.TrimPrefix(s, "hsl(")
	s = strings.TrimSuffix(s, ")")

	parts := strings.Split(s, ",")
	if len(parts) < 3 {
		return nil
	}

	h := parseHueValue(strings.TrimSpace(parts[0]))
	sat := parsePercentValue(strings.TrimSpace(parts[1]))
	l := parsePercentValue(strings.TrimSpace(parts[2]))
	var a float64 = 1

	if len(parts) >= 4 {
		alpha := strings.TrimSpace(parts[3])
		if strings.Contains(alpha, "%") {
			a = parsePercentValue(alpha)
		} else {
			a, _ = strconv.ParseFloat(alpha, 64)
		}
	}

	r, g, b := hslToRGB(h, sat, l)
	return color.RGBA{r, g, b, uint8(a * 255)}
}

// parseHueValue parses a hue value (0-360 degrees)
func parseHueValue(s string) float64 {
	s = strings.TrimSuffix(s, "deg")
	f, _ := strconv.ParseFloat(s, 64)
	return f / 360.0
}

// parsePercentValue parses a percentage value (0-100%)
func parsePercentValue(s string) float64 {
	s = strings.TrimSuffix(s, "%")
	f, _ := strconv.ParseFloat(s, 64)
	return f / 100.0
}

// hslToRGB converts HSL to RGB
func hslToRGB(h, s, l float64) (uint8, uint8, uint8) {
	var r, g, b float64

	if s == 0 {
		r, g, b = l, l, l
	} else {
		var q float64
		if l < 0.5 {
			q = l * (1 + s)
		} else {
			q = l + s - l*s
		}
		p := 2*l - q
		r = hueToRGB(p, q, h+1.0/3.0)
		g = hueToRGB(p, q, h)
		b = hueToRGB(p, q, h-1.0/3.0)
	}

	return uint8(math.Round(r * 255)), uint8(math.Round(g * 255)), uint8(math.Round(b * 255))
}

func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6*t
	}
	if t < 1.0/2.0 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6
	}
	return p
}

// parseColorValue parses a single color component
func parseColorValue(s string) uint8 {
	if strings.HasSuffix(s, "%") {
		s = strings.TrimSuffix(s, "%")
		f, _ := strconv.ParseFloat(s, 64)
		return uint8((f / 100) * 255)
	}
	i, _ := strconv.Atoi(s)
	if i > 255 {
		i = 255
	}
	if i < 0 {
		i = 0
	}
	return uint8(i)
}

// ============================================================================
// Filter Parsing
// ============================================================================

// ParseFilter parses a CSS filter string.
// Supported functions: blur(), brightness(), contrast(), grayscale(), sepia(),
// saturate(), hue-rotate(), invert()
// Example: "blur(5px) brightness(1.2) contrast(0.8)"
func ParseFilter(s string) *Filter {
	s = strings.TrimSpace(s)
	if s == "" || s == "none" {
		return nil
	}

	f := NewFilter()
	remaining := s

	for remaining != "" {
		remaining = strings.TrimSpace(remaining)
		if remaining == "" {
			break
		}

		parenIdx := strings.Index(remaining, "(")
		if parenIdx < 0 {
			break
		}
		funcName := strings.TrimSpace(remaining[:parenIdx])

		closeIdx := strings.Index(remaining, ")")
		if closeIdx < 0 {
			break
		}
		arg := strings.TrimSpace(remaining[parenIdx+1 : closeIdx])
		remaining = remaining[closeIdx+1:]

		switch strings.ToLower(funcName) {
		case "blur":
			f.Blur = parsePixelValue(arg)
		case "brightness":
			f.Brightness = parseFilterAmount(arg)
		case "contrast":
			f.Contrast = parseFilterAmount(arg)
		case "grayscale":
			f.Grayscale = parseFilterAmount(arg)
		case "sepia":
			f.Sepia = parseFilterAmount(arg)
		case "saturate":
			f.Saturate = parseFilterAmount(arg)
		case "hue-rotate":
			f.HueRotate = parseFilterAngle(arg)
		case "invert":
			f.Invert = parseFilterAmount(arg)
		}
	}

	return f
}

// ParseBackdropFilter parses a CSS backdrop-filter string.
// Currently supports: blur(Npx), brightness(), saturate()
func ParseBackdropFilter(s string) *BackdropFilter {
	s = strings.TrimSpace(s)
	if s == "" || s == "none" {
		return nil
	}

	bf := &BackdropFilter{Brightness: 1, Saturate: 1}
	remaining := s

	for remaining != "" {
		remaining = strings.TrimSpace(remaining)
		if remaining == "" {
			break
		}

		parenIdx := strings.Index(remaining, "(")
		if parenIdx < 0 {
			break
		}
		funcName := strings.TrimSpace(remaining[:parenIdx])

		closeIdx := strings.Index(remaining, ")")
		if closeIdx < 0 {
			break
		}
		arg := strings.TrimSpace(remaining[parenIdx+1 : closeIdx])
		remaining = remaining[closeIdx+1:]

		switch strings.ToLower(funcName) {
		case "blur":
			bf.Blur = parsePixelValue(arg)
		case "brightness":
			bf.Brightness = parseFilterAmount(arg)
		case "saturate":
			bf.Saturate = parseFilterAmount(arg)
		}
	}

	return bf
}

// parseFilterAmount parses a filter function amount.
// Accepts: "50%", "0.5", "1.2"
func parseFilterAmount(s string) float64 {
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, "%") {
		s = strings.TrimSuffix(s, "%")
		f, _ := strconv.ParseFloat(s, 64)
		return f / 100
	}
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// parseFilterAngle parses a filter angle argument.
// Accepts: "90deg", "1.57rad"
func parseFilterAngle(s string) float64 {
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, "rad") {
		s = strings.TrimSuffix(s, "rad")
		f, _ := strconv.ParseFloat(s, 64)
		return f
	}
	s = strings.TrimSuffix(s, "deg")
	f, _ := strconv.ParseFloat(s, 64)
	return f // degrees (will be converted to radians by the shader caller)
}

// ============================================================================
// Radial Gradient CSS Parsing
// ============================================================================

// parseRadialGradientCSS parses a CSS radial-gradient string.
// Example: "radial-gradient(circle, #ff0000, #0000ff)"
// Example: "radial-gradient(ellipse, red 0%, blue 100%)"
func parseRadialGradientCSS(s string) *Gradient {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "radial-gradient(")
	s = strings.TrimSuffix(s, ")")

	parts := strings.Split(s, ",")
	if len(parts) < 2 {
		return nil
	}

	g := &Gradient{
		Type:       GradientRadial,
		ColorStops: make([]ColorStop, 0),
	}

	startIdx := 0
	first := strings.TrimSpace(parts[0])
	firstLower := strings.ToLower(first)

	// Check if first part is shape keyword (circle, ellipse, etc.)
	if firstLower == "circle" || firstLower == "ellipse" ||
		strings.HasPrefix(firstLower, "circle ") || strings.HasPrefix(firstLower, "ellipse ") ||
		strings.Contains(firstLower, "at ") {
		startIdx = 1
	}

	// Parse color stops
	numColors := len(parts) - startIdx
	if numColors < 2 {
		return nil
	}

	for i := startIdx; i < len(parts); i++ {
		colorStr := strings.TrimSpace(parts[i])
		// Check for explicit position like "red 50%"
		colorAndPos := strings.Fields(colorStr)
		var clr color.Color
		stopPos := float64(i-startIdx) / float64(numColors-1)

		if len(colorAndPos) >= 2 {
			// Last field might be a percentage
			lastField := colorAndPos[len(colorAndPos)-1]
			if strings.HasSuffix(lastField, "%") {
				pct := strings.TrimSuffix(lastField, "%")
				f, err := strconv.ParseFloat(pct, 64)
				if err == nil {
					stopPos = f / 100
				}
				clr = parseColor(strings.Join(colorAndPos[:len(colorAndPos)-1], " "))
			} else {
				clr = parseColor(colorStr)
			}
		} else {
			clr = parseColor(colorStr)
		}

		if clr != nil {
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
