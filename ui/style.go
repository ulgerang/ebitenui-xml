package ui

import (
	"encoding/json"
	"fmt"
	"image/color"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ParseColor is the exported color parsing function for external use
func ParseColor(s string) color.Color {
	return parseColor(s)
}

// StyleSheet holds all style definitions
type StyleSheet struct {
	Styles    map[string]*Style                   `json:"styles"`
	Keyframes map[string]map[string]KeyframeStyle `json:"keyframes"`
}

// KeyframeStyle describes animatable properties in a JSON keyframe block.
type KeyframeStyle struct {
	Transform       string  `json:"transform"`
	Opacity         float64 `json:"opacity"`
	Width           float64 `json:"width"`
	Height          float64 `json:"height"`
	Background      string  `json:"background"`
	Border          string  `json:"border"`
	BoxShadowBlur   float64 `json:"boxShadowBlur"`
	BoxShadowSpread float64 `json:"boxShadowSpread"`
}

// StyleEngine manages and applies styles
type StyleEngine struct {
	styles map[string]*Style
	rules  []styleRuleRecord
}

type styleRuleRecord struct {
	Selector    string
	Style       *Style
	Specificity int
	Order       int
	Important   bool
}

type cssParsedRule struct {
	Selector  string
	Style     *Style
	Important bool
}

// NewStyleEngine creates a new style engine
func NewStyleEngine() *StyleEngine {
	return &StyleEngine{
		styles: make(map[string]*Style),
		rules:  make([]styleRuleRecord, 0),
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

	if rawKeyframes, ok := rawMap["keyframes"]; ok {
		if err := se.loadKeyframes(rawKeyframes); err != nil {
			return err
		}
		delete(rawMap, "keyframes")
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

		// Handle shorthand padding/margin: {"all": N}
		var rawFields map[string]interface{}
		if err := json.Unmarshal(rawStyle, &rawFields); err == nil {
			if p, ok := rawFields["padding"].(map[string]interface{}); ok {
				if all, ok := p["all"].(float64); ok {
					style.Padding = PaddingAll(all)
				}
			}
			if m, ok := rawFields["margin"].(map[string]interface{}); ok {
				if all, ok := m["all"].(float64); ok {
					style.Margin = MarginAll(all)
				}
			}
		}

		// Detect explicitly-set fields from raw JSON
		se.detectExplicitFields(&style, rawStyle)

		se.AddStyle(selector, &style)
	}

	return nil
}

func (se *StyleEngine) loadKeyframes(raw json.RawMessage) error {
	var rawAnimations map[string]map[string]KeyframeStyle
	if err := json.Unmarshal(raw, &rawAnimations); err != nil {
		return fmt.Errorf("failed to parse keyframes: %w", err)
	}
	for name, frames := range rawAnimations {
		anim := animationFromKeyframeStyles(name, frames)
		if anim != nil {
			RegisterAnimation(name, anim)
		}
	}
	return nil
}

// LoadCSS loads a small CSS subset: simple selector blocks plus literal @keyframes blocks.
func (se *StyleEngine) LoadCSS(css string) error {
	animations, err := parseCSSKeyframes(css)
	if err != nil {
		return err
	}
	for name, frames := range animations {
		anim := animationFromKeyframeStyles(name, frames)
		if anim != nil {
			RegisterAnimation(name, anim)
		}
	}
	rules, err := parseCSSStyleRules(css)
	if err != nil {
		return err
	}
	for _, rule := range rules {
		se.addStyle(rule.Selector, rule.Style, rule.Important)
	}
	return nil
}

func parseCSSKeyframes(css string) (map[string]map[string]KeyframeStyle, error) {
	result := make(map[string]map[string]KeyframeStyle)
	remaining := stripCSSComments(css)
	for {
		idx := strings.Index(remaining, "@keyframes")
		if idx < 0 {
			break
		}
		remaining = remaining[idx+len("@keyframes"):]
		remaining = strings.TrimSpace(remaining)
		nameEnd := strings.IndexAny(remaining, " \t\r\n{")
		if nameEnd <= 0 {
			return nil, fmt.Errorf("failed to parse @keyframes: missing animation name")
		}
		name := strings.TrimSpace(remaining[:nameEnd])
		remaining = strings.TrimSpace(remaining[nameEnd:])
		if !strings.HasPrefix(remaining, "{") {
			return nil, fmt.Errorf("failed to parse @keyframes %q: missing block", name)
		}
		body, rest, ok := consumeCSSBlock(remaining)
		if !ok {
			return nil, fmt.Errorf("failed to parse @keyframes %q: unclosed block", name)
		}
		frames, err := parseCSSKeyframeBody(name, body)
		if err != nil {
			return nil, err
		}
		result[name] = frames
		remaining = rest
	}
	if len(result) == 0 && strings.Contains(css, "@keyframes") {
		return nil, fmt.Errorf("failed to parse @keyframes")
	}
	return result, nil
}

func parseCSSStyleRules(css string) ([]cssParsedRule, error) {
	remaining := stripCSSAtKeyframes(stripCSSComments(css))
	rules := make([]cssParsedRule, 0)
	for strings.TrimSpace(remaining) != "" {
		open := strings.Index(remaining, "{")
		if open < 0 {
			if strings.TrimSpace(remaining) == "" {
				break
			}
			return nil, fmt.Errorf("failed to parse CSS rule: missing block")
		}
		selectorText := strings.TrimSpace(remaining[:open])
		if selectorText == "" {
			return nil, fmt.Errorf("failed to parse CSS rule: empty selector")
		}
		block, rest, ok := consumeCSSBlock(remaining[open:])
		if !ok {
			return nil, fmt.Errorf("failed to parse CSS rule %q: unclosed block", selectorText)
		}
		normalStyle, importantStyle := stylesFromCSSDeclarations(block)
		for _, selector := range splitCSSSelectorList(selectorText) {
			selector = strings.TrimSpace(selector)
			if selector != "" {
				if normalStyle != nil {
					rules = append(rules, cssParsedRule{Selector: selector, Style: normalStyle.Clone()})
				}
				if importantStyle != nil {
					rules = append(rules, cssParsedRule{Selector: selector, Style: importantStyle.Clone(), Important: true})
				}
			}
		}
		remaining = strings.TrimSpace(rest)
	}
	return rules, nil
}

func styleFromCSSDeclarations(block string) *Style {
	style, important := stylesFromCSSDeclarations(block)
	if style != nil {
		return style
	}
	if important != nil {
		return important
	}
	return &Style{}
}

func stylesFromCSSDeclarations(block string) (*Style, *Style) {
	normalStyle := &Style{}
	importantStyle := &Style{}
	hasNormal := false
	hasImportant := false
	for _, declaration := range splitCSSDeclarations(block) {
		name, value, ok := strings.Cut(declaration, ":")
		if !ok {
			continue
		}
		cleanValue, important := parseCSSImportantValue(value)
		if important {
			applyCSSDeclaration(importantStyle, strings.ToLower(strings.TrimSpace(name)), cleanValue)
			hasImportant = true
		} else {
			applyCSSDeclaration(normalStyle, strings.ToLower(strings.TrimSpace(name)), cleanValue)
			hasNormal = true
		}
	}
	if !hasNormal {
		normalStyle = nil
	}
	if !hasImportant {
		importantStyle = nil
	}
	return normalStyle, importantStyle
}

func applyCSSDeclaration(style *Style, prop, value string) {
	switch prop {
	case "display":
		style.Display = value
	case "flex-direction":
		switch value {
		case "row":
			style.Direction = LayoutRow
		case "column":
			style.Direction = LayoutColumn
		}
	case "justify-content":
		style.Justify = cssJustify(value)
	case "align-items":
		style.Align = cssAlign(value)
	case "gap":
		style.Gap = parseCSSPixels(value)
		style.GapSet = true
	case "width":
		style.Width = parseCSSPixels(value)
		style.WidthSet = true
	case "height":
		style.Height = parseCSSPixels(value)
		style.HeightSet = true
	case "min-width":
		style.MinWidth = parseCSSPixels(value)
		style.MinWidthSet = true
	case "min-height":
		style.MinHeight = parseCSSPixels(value)
		style.MinHeightSet = true
	case "max-width":
		style.MaxWidth = parseCSSPixels(value)
		style.MaxWidthSet = true
	case "max-height":
		style.MaxHeight = parseCSSPixels(value)
		style.MaxHeightSet = true
	case "box-sizing":
		style.BoxSizing = value
	case "flex-grow":
		style.FlexGrow, _ = strconv.ParseFloat(value, 64)
		style.FlexGrowSet = true
	case "flex-shrink":
		style.FlexShrink, _ = strconv.ParseFloat(value, 64)
		style.FlexShrinkSet = true
	case "flex-wrap":
		style.FlexWrap = FlexWrap(value)
	case "padding":
		spacing := cssBoxSpacing(value)
		style.Padding = Padding(spacing)
		style.PaddingSet = true
	case "margin":
		spacing := cssBoxSpacing(value)
		style.Margin = Margin(spacing)
		style.MarginSet = true
	case "background", "background-color":
		style.Background = value
	case "color":
		style.Color = value
	case "border":
		applyCSSBorderDeclaration(style, value)
	case "border-color":
		style.Border = value
	case "border-width":
		style.BorderWidth = parseCSSPixels(value)
		style.BorderWidthSet = true
	case "border-radius":
		style.BorderRadius = parseCSSPixels(value)
		style.BorderRadiusSet = true
	case "font-size":
		style.FontSize = parseCSSPixels(value)
		style.FontSizeSet = true
	case "font-weight":
		style.FontWeight = value
	case "line-height":
		style.LineHeight = parseCSSPixels(value)
		style.LineHeightSet = true
	case "letter-spacing":
		style.LetterSpacing = parseCSSPixels(value)
		style.LetterSpacingSet = true
	case "opacity":
		style.Opacity, _ = strconv.ParseFloat(value, 64)
		style.OpacitySet = true
	case "box-shadow":
		style.BoxShadow = value
	case "text-shadow":
		style.TextShadow = value
	case "transition":
		style.Transition = value
	case "animation":
		style.Animation = value
	case "filter":
		style.Filter = value
	case "backdrop-filter":
		style.BackdropFilter = value
	case "transform":
		style.Transform = value
	case "transform-origin":
		style.TransformOrigin = value
	case "clip-path":
		style.ClipPath = value
	case "overflow":
		style.Overflow = value
	case "overflow-x":
		style.OverflowX = value
	case "overflow-y":
		style.OverflowY = value
	case "position":
		style.Position = value
	case "top":
		style.Top = parseCSSPixels(value)
		style.TopSet = true
	case "right":
		style.Right = parseCSSPixels(value)
		style.RightSet = true
	case "bottom":
		style.Bottom = parseCSSPixels(value)
		style.BottomSet = true
	case "left":
		style.Left = parseCSSPixels(value)
		style.LeftSet = true
	case "z-index":
		style.ZIndex, _ = strconv.Atoi(strings.TrimSpace(value))
		style.ZIndexSet = true
	case "visibility":
		style.Visibility = value
	}
}

func cssJustify(value string) Justify {
	switch value {
	case "center":
		return JustifyCenter
	case "flex-end", "end":
		return JustifyEnd
	case "space-between":
		return JustifyBetween
	case "space-around":
		return JustifyAround
	case "space-evenly":
		return JustifyEvenly
	default:
		return JustifyStart
	}
}

func cssAlign(value string) Alignment {
	switch value {
	case "center":
		return AlignCenter
	case "flex-end", "end":
		return AlignEnd
	case "stretch":
		return AlignStretch
	default:
		return AlignStart
	}
}

type cssSpacing struct {
	Top, Right, Bottom, Left float64
}

func cssBoxSpacing(value string) cssSpacing {
	parts := strings.Fields(value)
	if len(parts) == 0 {
		return cssSpacing{}
	}
	values := make([]float64, len(parts))
	for i, part := range parts {
		values[i] = parseCSSPixels(part)
	}
	top, right, bottom, left := values[0], values[0], values[0], values[0]
	if len(values) == 2 {
		top, bottom = values[0], values[0]
		right, left = values[1], values[1]
	} else if len(values) == 3 {
		top = values[0]
		right, left = values[1], values[1]
		bottom = values[2]
	} else if len(values) >= 4 {
		top, right, bottom, left = values[0], values[1], values[2], values[3]
	}
	return cssSpacing{Top: top, Right: right, Bottom: bottom, Left: left}
}

func applyCSSBorderDeclaration(style *Style, value string) {
	parts := strings.Fields(value)
	for _, part := range parts {
		if strings.HasSuffix(part, "px") || isCSSNumeric(part) {
			style.BorderWidth = parseCSSPixels(part)
			style.BorderWidthSet = true
			continue
		}
		if strings.HasPrefix(part, "#") || strings.HasPrefix(part, "rgb") || isNamedColor(part) {
			style.Border = part
		}
	}
}

func isCSSNumeric(value string) bool {
	_, err := strconv.ParseFloat(value, 64)
	return err == nil
}

func isNamedColor(value string) bool {
	switch strings.ToLower(value) {
	case "white", "black", "red", "green", "blue", "transparent":
		return true
	default:
		return false
	}
}

func parseCSSKeyframeBody(name, body string) (map[string]KeyframeStyle, error) {
	frames := make(map[string]KeyframeStyle)
	remaining := strings.TrimSpace(body)
	for remaining != "" {
		open := strings.Index(remaining, "{")
		if open < 0 {
			if strings.TrimSpace(remaining) == "" {
				break
			}
			return nil, fmt.Errorf("failed to parse @keyframes %q: malformed selector", name)
		}
		selector := strings.TrimSpace(remaining[:open])
		if selector == "" {
			return nil, fmt.Errorf("failed to parse @keyframes %q: empty selector", name)
		}
		block, rest, ok := consumeCSSBlock(remaining[open:])
		if !ok {
			return nil, fmt.Errorf("failed to parse @keyframes %q: unclosed keyframe block", name)
		}
		for _, part := range strings.Split(selector, ",") {
			label := strings.TrimSpace(part)
			if _, ok := parseKeyframePercent(label); !ok {
				return nil, fmt.Errorf("failed to parse @keyframes %q: invalid selector %q", name, label)
			}
			frames[label] = keyframeStyleFromCSSDeclarations(block)
		}
		remaining = strings.TrimSpace(rest)
	}
	if len(frames) == 0 {
		return nil, fmt.Errorf("failed to parse @keyframes %q: no frames", name)
	}
	return frames, nil
}

func keyframeStyleFromCSSDeclarations(block string) KeyframeStyle {
	var frame KeyframeStyle
	for _, declaration := range splitCSSDeclarations(block) {
		name, value, ok := strings.Cut(declaration, ":")
		if !ok {
			continue
		}
		prop := strings.ToLower(strings.TrimSpace(name))
		text, _ := parseCSSImportantValue(value)
		switch prop {
		case "opacity":
			frame.Opacity, _ = strconv.ParseFloat(text, 64)
		case "transform":
			frame.Transform = text
		case "width":
			frame.Width = parseCSSPixels(text)
		case "height":
			frame.Height = parseCSSPixels(text)
		case "background", "background-color":
			frame.Background = text
		case "border", "border-color":
			frame.Border = text
		case "box-shadow-blur":
			frame.BoxShadowBlur = parseCSSPixels(text)
		case "box-shadow-spread":
			frame.BoxShadowSpread = parseCSSPixels(text)
		}
	}
	return frame
}

func splitCSSDeclarations(block string) []string {
	return splitCSSListLike(block, ';')
}

func splitCSSSelectorList(selectorText string) []string {
	return splitCSSListLike(selectorText, ',')
}

func splitCSSListLike(s string, separator rune) []string {
	parts := make([]string, 0)
	var current strings.Builder
	depth := 0
	var quote rune
	escaped := false
	for _, r := range s {
		if quote != 0 {
			current.WriteRune(r)
			if escaped {
				escaped = false
				continue
			}
			if r == '\\' {
				escaped = true
				continue
			}
			if r == quote {
				quote = 0
			}
			continue
		}
		switch r {
		case '\'', '"':
			quote = r
			current.WriteRune(r)
		case '(':
			depth++
			current.WriteRune(r)
		case ')':
			if depth > 0 {
				depth--
			}
			current.WriteRune(r)
		default:
			if r == separator && depth == 0 {
				if part := strings.TrimSpace(current.String()); part != "" {
					parts = append(parts, part)
				}
				current.Reset()
				continue
			}
			current.WriteRune(r)
		}
	}
	if part := strings.TrimSpace(current.String()); part != "" {
		parts = append(parts, part)
	}
	return parts
}

func parseCSSImportantValue(value string) (string, bool) {
	value = strings.TrimSpace(value)
	lower := strings.ToLower(value)
	if !strings.HasSuffix(lower, "!important") {
		return value, false
	}
	before := strings.TrimSpace(value[:len(value)-len("!important")])
	return before, true
}

func stripCSSComments(css string) string {
	for {
		start := strings.Index(css, "/*")
		if start < 0 {
			return css
		}
		end := strings.Index(css[start+2:], "*/")
		if end < 0 {
			return css[:start]
		}
		css = css[:start] + css[start+2+end+2:]
	}
}

func stripCSSAtKeyframes(css string) string {
	var out strings.Builder
	remaining := css
	for {
		idx := strings.Index(remaining, "@keyframes")
		if idx < 0 {
			out.WriteString(remaining)
			return out.String()
		}
		out.WriteString(remaining[:idx])
		after := strings.TrimSpace(remaining[idx+len("@keyframes"):])
		nameEnd := strings.IndexAny(after, " \t\r\n{")
		if nameEnd < 0 {
			return out.String()
		}
		after = strings.TrimSpace(after[nameEnd:])
		if !strings.HasPrefix(after, "{") {
			return out.String()
		}
		_, rest, ok := consumeCSSBlock(after)
		if !ok {
			return out.String()
		}
		remaining = rest
	}
}

func consumeCSSBlock(s string) (body, rest string, ok bool) {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "{") {
		return "", s, false
	}
	depth := 0
	for i, r := range s {
		switch r {
		case '{':
			depth++
			if depth == 1 {
				continue
			}
		case '}':
			depth--
			if depth == 0 {
				return s[1:i], s[i+1:], true
			}
		}
	}
	return "", s, false
}

func animationFromKeyframeStyles(name string, frames map[string]KeyframeStyle) *Animation {
	if name == "" || len(frames) == 0 {
		return nil
	}
	keyframes := make([]Keyframe, 0, len(frames))
	for label, frame := range frames {
		percent, ok := parseKeyframePercent(label)
		if !ok {
			continue
		}
		keyframes = append(keyframes, Keyframe{
			Percent:    percent,
			Properties: keyframePropertiesFromStyle(frame),
		})
	}
	if len(keyframes) == 0 {
		return nil
	}
	sort.Slice(keyframes, func(i, j int) bool {
		return keyframes[i].Percent < keyframes[j].Percent
	})
	return &Animation{
		Name:           name,
		Duration:       300 * time.Millisecond,
		IterationCount: 1,
		TimingFunc:     EaseLinear,
		Keyframes:      keyframes,
	}
}

func parseKeyframePercent(label string) (float64, bool) {
	label = strings.TrimSpace(strings.ToLower(label))
	switch label {
	case "from":
		return 0, true
	case "to":
		return 100, true
	}
	label = strings.TrimSuffix(label, "%")
	value, err := strconv.ParseFloat(label, 64)
	if err != nil {
		return 0, false
	}
	if value < 0 {
		value = 0
	}
	if value > 100 {
		value = 100
	}
	return value, true
}

func keyframePropertiesFromStyle(frame KeyframeStyle) KeyframeProperties {
	props := KeyframeProperties{
		Opacity:         frame.Opacity,
		Width:           frame.Width,
		Height:          frame.Height,
		BackgroundColor: parseColor(frame.Background),
		BorderColor:     parseColor(frame.Border),
		BoxShadowBlur:   frame.BoxShadowBlur,
		BoxShadowSpread: frame.BoxShadowSpread,
	}
	if frame.Transform != "" {
		transform := parseKeyframeTransform(frame.Transform)
		props.TranslateX = transform.TranslateX
		props.TranslateY = transform.TranslateY
		props.ScaleX = transform.ScaleX
		props.ScaleY = transform.ScaleY
		props.Rotate = transform.Rotate
		props.SkewX = transform.SkewX
		props.SkewY = transform.SkewY
	}
	return props
}

func parseKeyframeTransform(value string) KeyframeProperties {
	props := KeyframeProperties{ScaleX: 1, ScaleY: 1}
	for _, fn := range parseTransformFunctions(value) {
		name, args, ok := strings.Cut(fn, "(")
		if !ok {
			continue
		}
		args = strings.TrimSuffix(args, ")")
		parts := strings.Split(args, ",")
		if len(parts) == 1 {
			parts = strings.Fields(args)
		}
		switch strings.ToLower(strings.TrimSpace(name)) {
		case "translate", "translate3d":
			if len(parts) > 0 {
				props.TranslateX = parseCSSPixels(parts[0])
			}
			if len(parts) > 1 {
				props.TranslateY = parseCSSPixels(parts[1])
			}
		case "translatex":
			if len(parts) > 0 {
				props.TranslateX = parseCSSPixels(parts[0])
			}
		case "translatey":
			if len(parts) > 0 {
				props.TranslateY = parseCSSPixels(parts[0])
			}
		case "scale":
			if len(parts) > 0 {
				props.ScaleX = parseCSSScaleValue(parts[0])
				props.ScaleY = props.ScaleX
			}
			if len(parts) > 1 {
				props.ScaleY = parseCSSScaleValue(parts[1])
			}
		case "scalex":
			if len(parts) > 0 {
				props.ScaleX = parseCSSScaleValue(parts[0])
			}
		case "scaley":
			if len(parts) > 0 {
				props.ScaleY = parseCSSScaleValue(parts[0])
			}
		case "rotate":
			if len(parts) > 0 {
				props.Rotate = parseCSSAngle(parts[0]) * 180 / math.Pi
			}
		case "skewx":
			if len(parts) > 0 {
				props.SkewX = parseCSSAngle(parts[0]) * 180 / math.Pi
			}
		case "skewy":
			if len(parts) > 0 {
				props.SkewY = parseCSSAngle(parts[0]) * 180 / math.Pi
			}
		}
	}
	return props
}

func parseTransformFunctions(value string) []string {
	var funcs []string
	start := -1
	depth := 0
	for i, r := range value {
		if start == -1 && r != ' ' && r != '\t' {
			start = i
		}
		switch r {
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
			if depth == 0 && start >= 0 {
				funcs = append(funcs, strings.TrimSpace(value[start:i+1]))
				start = -1
			}
		}
	}
	return funcs
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

	// Sizing properties
	if _, ok := rawFields["width"]; ok {
		style.WidthSet = true
	}
	if _, ok := rawFields["height"]; ok {
		style.HeightSet = true
	}
	if _, ok := rawFields["minWidth"]; ok {
		style.MinWidthSet = true
	}
	if _, ok := rawFields["minHeight"]; ok {
		style.MinHeightSet = true
	}
	if _, ok := rawFields["maxWidth"]; ok {
		style.MaxWidthSet = true
	}
	if _, ok := rawFields["maxHeight"]; ok {
		style.MaxHeightSet = true
	}
	if _, ok := rawFields["flexGrow"]; ok {
		style.FlexGrowSet = true
	}
	if _, ok := rawFields["flexShrink"]; ok {
		style.FlexShrinkSet = true
	}

	// Layout properties
	if _, ok := rawFields["gap"]; ok {
		style.GapSet = true
	}

	// Border radius properties
	if _, ok := rawFields["borderRadius"]; ok {
		style.BorderRadiusSet = true
	}
	if _, ok := rawFields["borderTopWidth"]; ok {
		style.BorderTopWidthSet = true
	}
	if _, ok := rawFields["borderRightWidth"]; ok {
		style.BorderRightWidthSet = true
	}
	if _, ok := rawFields["borderBottomWidth"]; ok {
		style.BorderBottomWidthSet = true
	}
	if _, ok := rawFields["borderLeftWidth"]; ok {
		style.BorderLeftWidthSet = true
	}
	if _, ok := rawFields["borderTopLeftRadius"]; ok {
		style.BorderTopLeftRadiusSet = true
	}
	if _, ok := rawFields["borderTopRightRadius"]; ok {
		style.BorderTopRightRadiusSet = true
	}
	if _, ok := rawFields["borderBottomLeftRadius"]; ok {
		style.BorderBottomLeftRadiusSet = true
	}
	if _, ok := rawFields["borderBottomRightRadius"]; ok {
		style.BorderBottomRightRadiusSet = true
	}

	// Text properties
	if _, ok := rawFields["fontSize"]; ok {
		style.FontSizeSet = true
	}
	if _, ok := rawFields["lineHeight"]; ok {
		style.LineHeightSet = true
	}
	if _, ok := rawFields["letterSpacing"]; ok {
		style.LetterSpacingSet = true
	}

	// Visual effects
	if _, ok := rawFields["opacity"]; ok {
		style.OpacitySet = true
	}
	if _, ok := rawFields["outlineOffset"]; ok {
		style.OutlineOffsetSet = true
	}

	// Position properties
	if _, ok := rawFields["top"]; ok {
		style.TopSet = true
	}
	if _, ok := rawFields["right"]; ok {
		style.RightSet = true
	}
	if _, ok := rawFields["bottom"]; ok {
		style.BottomSet = true
	}
	if _, ok := rawFields["left"]; ok {
		style.LeftSet = true
	}
	if _, ok := rawFields["zIndex"]; ok {
		style.ZIndexSet = true
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
	if style.BorderTop != "" {
		style.BorderTopColor = parseColor(style.BorderTop)
	}
	if style.BorderRight != "" {
		style.BorderRightColor = parseColor(style.BorderRight)
	}
	if style.BorderBottom != "" {
		style.BorderBottomColor = parseColor(style.BorderBottom)
	}
	if style.BorderLeft != "" {
		style.BorderLeftColor = parseColor(style.BorderLeft)
	}
	if style.Color != "" {
		style.TextColor = parseColor(style.Color)
	}

	// Parse box shadow
	if style.BoxShadow != "" {
		style.parsedBoxShadow = parseBoxShadow(style.BoxShadow)
		style.parsedBoxShadows = ParseBoxShadowList(style.BoxShadow)
	}

	// Parse text shadow
	if style.TextShadow != "" {
		style.parsedTextShadow = ParseTextShadow(style.TextShadow)
		style.parsedTextShadows = ParseTextShadowList(style.TextShadow)
	}

	// Parse transitions
	if style.Transition != "" {
		style.parsedTransitions = parseTransitions(style.Transition)
	}

	// Parse declarative animation
	if style.Animation != "" {
		style.parsedAnimation = ParseAnimationDeclaration(style.Animation)
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
	trimmed := strings.TrimSpace(s)
	if !strings.HasPrefix(trimmed, "{") && (strings.Contains(trimmed, "@keyframes") || looksLikeCSSRuleString(trimmed)) {
		return se.LoadCSS(s)
	}
	return se.LoadFromJSON([]byte(s))
}

func looksLikeCSSRuleString(s string) bool {
	open := strings.Index(s, "{")
	close := strings.LastIndex(s, "}")
	if open <= 0 || close <= open {
		return false
	}
	selector := strings.TrimSpace(s[:open])
	block := s[open+1 : close]
	return selector != "" && strings.Contains(block, ":") && strings.Contains(block, ";")
}

// AddStyle adds a style for a selector
func (se *StyleEngine) AddStyle(selector string, style *Style) {
	se.addStyle(selector, style, false)
}

func (se *StyleEngine) addStyle(selector string, style *Style, important bool) {
	se.parseStyleColors(style)
	specificity := complexSelectorSpecificity(selector)
	selector, style = styleForTerminalPseudoSelector(selector, style)
	if !important {
		se.styles[selector] = style
	}
	se.rules = append(se.rules, styleRuleRecord{
		Selector:    selector,
		Style:       style.Clone(),
		Specificity: specificity,
		Order:       len(se.rules),
		Important:   important,
	})
}

func styleForTerminalPseudoSelector(selector string, style *Style) (string, *Style) {
	base, pseudo, ok := splitTerminalStatePseudo(selector)
	if !ok {
		return selector, style
	}
	stateStyle := &Style{}
	switch pseudo {
	case "hover":
		stateStyle.HoverStyle = style.Clone()
	case "active":
		stateStyle.ActiveStyle = style.Clone()
	case "focus", "focused":
		stateStyle.FocusStyle = style.Clone()
	case "disabled":
		stateStyle.DisabledStyle = style.Clone()
	default:
		return selector, style
	}
	return base, stateStyle
}

func splitTerminalStatePseudo(selector string) (string, string, bool) {
	depth := 0
	var quote rune
	escaped := false
	for i := len(selector) - 1; i >= 0; i-- {
		r := rune(selector[i])
		if quote != 0 {
			if escaped {
				escaped = false
				continue
			}
			if r == '\\' {
				escaped = true
				continue
			}
			if r == quote {
				quote = 0
			}
			continue
		}
		switch r {
		case '\'', '"':
			quote = r
		case ')':
			depth++
		case '(':
			if depth > 0 {
				depth--
			}
		case ':':
			if depth == 0 {
				pseudo := strings.TrimSpace(selector[i+1:])
				if strings.ContainsAny(pseudo, " .#[:>+~") {
					return selector, "", false
				}
				base := strings.TrimSpace(selector[:i])
				return base, strings.ToLower(pseudo), base != ""
			}
		}
	}
	return selector, "", false
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

// ParseAnimationDeclaration parses a compact CSS-like animation declaration.
// Supported form: "name [duration] [easing] [iteration-count]".
func ParseAnimationDeclaration(s string) *Animation {
	s = strings.TrimSpace(s)
	if s == "" || s == "none" {
		return nil
	}

	parts := strings.Fields(s)
	if len(parts) == 0 {
		return nil
	}

	anim := GetAnimation(parts[0])
	if anim == nil {
		return nil
	}

	for _, part := range parts[1:] {
		lower := strings.ToLower(part)
		switch {
		case strings.HasSuffix(lower, "ms") || strings.HasSuffix(lower, "s"):
			anim.Duration = timeDuration(parseDuration(lower))
		case lower == "infinite":
			anim.IterationCount = -1
		case isNumeric(lower):
			count, err := strconv.Atoi(lower)
			if err == nil {
				anim.IterationCount = count
			}
		default:
			anim.TimingFunc = ParseEasing(lower)
		}
	}

	return anim
}

func timeDuration(seconds float64) time.Duration {
	return time.Duration(seconds * float64(time.Second))
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
	colors := map[string]color.NRGBA{
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

	return color.NRGBA{r, g, b, a}
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

	return color.NRGBA{r, g, b, a}
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
	return color.NRGBA{r, g, b, uint8(a * 255)}
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
