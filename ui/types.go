// Package ui provides a data-driven UI framework for Ebiten
// that separates structure (XML) from styling (CSS-like JSON)
package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// LayoutDirection defines how children are arranged
type LayoutDirection string

const (
	LayoutRow    LayoutDirection = "row"
	LayoutColumn LayoutDirection = "column"
)

// Alignment defines how items are aligned
type Alignment string

const (
	AlignStart   Alignment = "start"
	AlignCenter  Alignment = "center"
	AlignEnd     Alignment = "end"
	AlignStretch Alignment = "stretch"
)

// Justify defines how items are justified along main axis
type Justify string

const (
	JustifyStart   Justify = "start"
	JustifyCenter  Justify = "center"
	JustifyEnd     Justify = "end"
	JustifyBetween Justify = "space-between"
	JustifyAround  Justify = "space-around"
	JustifyEvenly  Justify = "space-evenly"
)

// FlexWrap defines wrapping behavior
type FlexWrap string

const (
	FlexNoWrap      FlexWrap = "nowrap"
	FlexWrapNormal  FlexWrap = "wrap"
	FlexWrapReverse FlexWrap = "wrap-reverse"
)

// Rect represents a rectangle with position and size
type Rect struct {
	X, Y, W, H float64
}

// Contains checks if a point is inside the rectangle
func (r Rect) Contains(x, y float64) bool {
	return x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H
}

// Inset returns a new rect with padding applied inward
func (r Rect) Inset(top, right, bottom, left float64) Rect {
	return Rect{
		X: r.X + left,
		Y: r.Y + top,
		W: r.W - left - right,
		H: r.H - top - bottom,
	}
}

// Padding represents padding values (like CSS shorthand)
type Padding struct {
	Top, Right, Bottom, Left float64
}

// All returns a Padding with all sides equal
func PaddingAll(v float64) Padding {
	return Padding{v, v, v, v}
}

// Horizontal returns horizontal padding total
func (p Padding) Horizontal() float64 {
	return p.Left + p.Right
}

// Vertical returns vertical padding total
func (p Padding) Vertical() float64 {
	return p.Top + p.Bottom
}

// Margin represents margin values
type Margin struct {
	Top, Right, Bottom, Left float64
}

// All returns a Margin with all sides equal
func MarginAll(v float64) Margin {
	return Margin{v, v, v, v}
}

// Horizontal returns horizontal margin total
func (m Margin) Horizontal() float64 {
	return m.Left + m.Right
}

// Vertical returns vertical margin total
func (m Margin) Vertical() float64 {
	return m.Top + m.Bottom
}

// BorderStyle represents individual border properties
type BorderStyle struct {
	Width  float64     `json:"width"`
	Color  string      `json:"color"`
	Radius float64     `json:"radius"`
	Clr    color.Color `json:"-"`
}

// Style holds all visual properties for a widget
// Designed to mirror CSS properties as closely as possible
type Style struct {
	// Layout (Flexbox-like)
	Direction LayoutDirection `json:"direction"`
	Align     Alignment       `json:"align"`
	Justify   Justify         `json:"justify"`
	Gap       float64         `json:"gap"`
	FlexWrap  FlexWrap        `json:"flexWrap"`

	// Sizing
	Width      float64 `json:"width"`
	Height     float64 `json:"height"`
	MinWidth   float64 `json:"minWidth"`
	MinHeight  float64 `json:"minHeight"`
	MaxWidth   float64 `json:"maxWidth"`
	MaxHeight  float64 `json:"maxHeight"`
	FlexGrow   float64 `json:"flexGrow"`
	FlexShrink float64 `json:"flexShrink"`

	// Spacing
	Padding    Padding `json:"padding"`
	Margin     Margin  `json:"margin"`
	PaddingSet bool    `json:"-"` // true if padding was explicitly set (allows zero override)
	MarginSet  bool    `json:"-"` // true if margin was explicitly set (allows zero override)

	// Colors
	BackgroundColor color.Color `json:"-"`
	BorderColor     color.Color `json:"-"`
	TextColor       color.Color `json:"-"`

	// Color strings (for JSON parsing)
	Background string `json:"background"`
	Border     string `json:"border"`
	Color      string `json:"color"`

	// Border (expanded CSS-like properties)
	BorderWidth             float64 `json:"borderWidth"`
	BorderRadius            float64 `json:"borderRadius"`
	BorderWidthSet          bool    `json:"-"` // true if borderWidth was explicitly set
	BorderTopWidth          float64 `json:"borderTopWidth"`
	BorderRightWidth        float64 `json:"borderRightWidth"`
	BorderBottomWidth       float64 `json:"borderBottomWidth"`
	BorderLeftWidth         float64 `json:"borderLeftWidth"`
	BorderTopLeftRadius     float64 `json:"borderTopLeftRadius"`
	BorderTopRightRadius    float64 `json:"borderTopRightRadius"`
	BorderBottomLeftRadius  float64 `json:"borderBottomLeftRadius"`
	BorderBottomRightRadius float64 `json:"borderBottomRightRadius"`

	// Text
	FontSize      float64 `json:"fontSize"`
	FontFamily    string  `json:"fontFamily"`
	FontWeight    string  `json:"fontWeight"` // normal, bold, 100-900
	FontStyle     string  `json:"fontStyle"`  // normal, italic
	TextAlign     string  `json:"textAlign"`  // left, center, right
	LineHeight    float64 `json:"lineHeight"`
	LetterSpacing float64 `json:"letterSpacing"`
	TextWrap      string  `json:"textWrap"`     // normal, nowrap
	TextOverflow  string  `json:"textOverflow"` // clip, ellipsis

	// Visual Effects
	Opacity    float64 `json:"opacity"`    // 0-1
	BoxShadow  string  `json:"boxShadow"`  // "offsetX offsetY blur spread color"
	TextShadow string  `json:"textShadow"` // "offsetX offsetY blur color"
	Cursor     string  `json:"cursor"`     // pointer, default, etc.
	Transition string  `json:"transition"` // "property duration easing"

	// Outline (separate from border)
	Outline       string  `json:"outline"`       // "width style color"
	OutlineOffset float64 `json:"outlineOffset"` // distance from border

	// Filters
	Filter         string `json:"filter"`         // blur(), brightness(), etc.
	BackdropFilter string `json:"backdropFilter"` // for glassmorphism

	// Transform
	Transform       string `json:"transform"`       // rotate(), scale(), translate()
	TransformOrigin string `json:"transformOrigin"` // center, top left, etc.

	// 9-Slice Image
	BackgroundImage    string `json:"backgroundImage"`    // image path
	BorderImage        string `json:"borderImage"`        // image path for 9-slice
	BorderImageSlice   string `json:"borderImageSlice"`   // "top right bottom left"
	BackgroundSize     string `json:"backgroundSize"`     // cover, contain, or dimensions
	BackgroundPosition string `json:"backgroundPosition"` // center, top left, etc.
	BackgroundRepeat   string `json:"backgroundRepeat"`   // repeat, no-repeat

	// Overflow
	Overflow  string `json:"overflow"` // visible, hidden, scroll
	OverflowX string `json:"overflowX"`
	OverflowY string `json:"overflowY"`

	// Position
	Position string  `json:"position"` // relative, absolute
	Top      float64 `json:"top"`
	Right    float64 `json:"right"`
	Bottom   float64 `json:"bottom"`
	Left     float64 `json:"left"`
	ZIndex   int     `json:"zIndex"`

	// Display
	Display    string `json:"display"`    // block, flex, none
	Visibility string `json:"visibility"` // visible, hidden

	// States
	HoverStyle    *Style `json:"hover"`
	ActiveStyle   *Style `json:"active"`
	DisabledStyle *Style `json:"disabled"`
	FocusStyle    *Style `json:"focus"`

	// Parsed values (internal)
	parsedBoxShadow   *BoxShadow   `json:"-"`
	parsedTextShadow  *TextShadow  `json:"-"`
	parsedOutline     *Outline     `json:"-"`
	parsedTransitions []Transition `json:"-"`
	parsed9Slice      *NineSlice   `json:"-"`
	parsedGradient    *Gradient    `json:"-"`
	parsedFilter      *Filter      `json:"-"`
}

// Clone creates a deep copy of the style
func (s *Style) Clone() *Style {
	if s == nil {
		return nil
	}
	copy := *s
	if s.HoverStyle != nil {
		copy.HoverStyle = s.HoverStyle.Clone()
	}
	if s.ActiveStyle != nil {
		copy.ActiveStyle = s.ActiveStyle.Clone()
	}
	if s.DisabledStyle != nil {
		copy.DisabledStyle = s.DisabledStyle.Clone()
	}
	if s.FocusStyle != nil {
		copy.FocusStyle = s.FocusStyle.Clone()
	}
	return &copy
}

// Merge merges another style into this one (non-zero values override)
func (s *Style) Merge(other *Style) {
	if other == nil {
		return
	}

	// Layout
	if other.Direction != "" {
		s.Direction = other.Direction
	}
	if other.Align != "" {
		s.Align = other.Align
	}
	if other.Justify != "" {
		s.Justify = other.Justify
	}
	if other.Gap != 0 {
		s.Gap = other.Gap
	}
	if other.FlexWrap != "" {
		s.FlexWrap = other.FlexWrap
	}

	// Sizing
	if other.Width != 0 {
		s.Width = other.Width
	}
	if other.Height != 0 {
		s.Height = other.Height
	}
	if other.MinWidth != 0 {
		s.MinWidth = other.MinWidth
	}
	if other.MinHeight != 0 {
		s.MinHeight = other.MinHeight
	}
	if other.MaxWidth != 0 {
		s.MaxWidth = other.MaxWidth
	}
	if other.MaxHeight != 0 {
		s.MaxHeight = other.MaxHeight
	}
	if other.FlexGrow != 0 {
		s.FlexGrow = other.FlexGrow
	}
	if other.FlexShrink != 0 {
		s.FlexShrink = other.FlexShrink
	}

	// Spacing - PaddingSet/MarginSet allow explicit zero-padding overrides
	if other.PaddingSet || other.Padding.Top != 0 || other.Padding.Right != 0 || other.Padding.Bottom != 0 || other.Padding.Left != 0 {
		s.Padding = other.Padding
		s.PaddingSet = true
	}
	if other.MarginSet || other.Margin.Top != 0 || other.Margin.Right != 0 || other.Margin.Bottom != 0 || other.Margin.Left != 0 {
		s.Margin = other.Margin
		s.MarginSet = true
	}

	// Colors
	if other.BackgroundColor != nil {
		s.BackgroundColor = other.BackgroundColor
	}
	if other.BorderColor != nil {
		s.BorderColor = other.BorderColor
	}
	if other.TextColor != nil {
		s.TextColor = other.TextColor
	}
	if other.Background != "" {
		s.Background = other.Background
		// Propagate parsed gradient/color when background string changes
		s.parsedGradient = other.parsedGradient
		if other.parsedGradient != nil {
			s.BackgroundColor = nil // gradient takes priority
		}
	}
	if other.Border != "" {
		s.Border = other.Border
	}
	if other.Color != "" {
		s.Color = other.Color
	}

	// Border
	if other.BorderWidthSet || other.BorderWidth != 0 {
		s.BorderWidth = other.BorderWidth
		s.BorderWidthSet = true
	}
	if other.BorderRadius != 0 {
		s.BorderRadius = other.BorderRadius
	}

	// Text
	if other.FontSize != 0 {
		s.FontSize = other.FontSize
	}
	if other.FontFamily != "" {
		s.FontFamily = other.FontFamily
	}
	if other.TextAlign != "" {
		s.TextAlign = other.TextAlign
	}
	if other.LineHeight != 0 {
		s.LineHeight = other.LineHeight
	}
	if other.TextWrap != "" {
		s.TextWrap = other.TextWrap
	}
	if other.TextOverflow != "" {
		s.TextOverflow = other.TextOverflow
	}

	// Visual Effects
	if other.Opacity != 0 {
		s.Opacity = other.Opacity
	}
	if other.BoxShadow != "" {
		s.BoxShadow = other.BoxShadow
		s.parsedBoxShadow = other.parsedBoxShadow
	}
	if other.TextShadow != "" {
		s.TextShadow = other.TextShadow
		s.parsedTextShadow = other.parsedTextShadow
	}
	if other.Outline != "" {
		s.Outline = other.Outline
		s.parsedOutline = other.parsedOutline
	}
	if other.OutlineOffset != 0 {
		s.OutlineOffset = other.OutlineOffset
	}
	if other.Transition != "" {
		s.Transition = other.Transition
		s.parsedTransitions = other.parsedTransitions
	}

	// 9-Slice
	if other.BackgroundImage != "" {
		s.BackgroundImage = other.BackgroundImage
	}
	if other.BorderImage != "" {
		s.BorderImage = other.BorderImage
	}
	if other.BorderImageSlice != "" {
		s.BorderImageSlice = other.BorderImageSlice
	}

	// Overflow
	if other.Overflow != "" {
		s.Overflow = other.Overflow
	}

	// Position
	if other.Position != "" {
		s.Position = other.Position
	}
	if other.ZIndex != 0 {
		s.ZIndex = other.ZIndex
	}

	// Display
	if other.Display != "" {
		s.Display = other.Display
	}
	if other.Visibility != "" {
		s.Visibility = other.Visibility
	}

	// States
	if other.HoverStyle != nil {
		if s.HoverStyle == nil {
			s.HoverStyle = &Style{}
		}
		s.HoverStyle.Merge(other.HoverStyle)
	}
	if other.ActiveStyle != nil {
		if s.ActiveStyle == nil {
			s.ActiveStyle = &Style{}
		}
		s.ActiveStyle.Merge(other.ActiveStyle)
	}
	if other.DisabledStyle != nil {
		if s.DisabledStyle == nil {
			s.DisabledStyle = &Style{}
		}
		s.DisabledStyle.Merge(other.DisabledStyle)
	}
	if other.FocusStyle != nil {
		if s.FocusStyle == nil {
			s.FocusStyle = &Style{}
		}
		s.FocusStyle.Merge(other.FocusStyle)
	}
}

// WidgetState represents the current interaction state
type WidgetState int

const (
	StateNormal WidgetState = iota
	StateHover
	StateActive
	StateDisabled
	StateFocused
)

// Widget is the base interface for all UI elements
type Widget interface {
	// Identification
	ID() string
	Type() string
	Classes() []string
	AddClass(class string)
	RemoveClass(class string)
	HasClass(class string) bool

	// Tree structure
	Parent() Widget
	SetParent(w Widget)
	Children() []Widget
	AddChild(w Widget)
	RemoveChild(w Widget)

	// Layout
	ComputedRect() Rect
	SetComputedRect(r Rect)
	Style() *Style
	SetStyle(s *Style)

	// State
	State() WidgetState
	SetState(s WidgetState)
	Visible() bool
	SetVisible(v bool)
	Enabled() bool
	SetEnabled(e bool)

	// Rendering
	Draw(screen *ebiten.Image)

	// Events
	OnClick(handler func())
	OnHover(handler func())
}

// EventHandler is a function that handles UI events
type EventHandler func(widget Widget, event Event)

// Event represents a UI event
type Event struct {
	Type    EventType
	X, Y    float64
	DeltaX  float64 // for scroll events
	DeltaY  float64
	Button  ebiten.MouseButton
	Key     ebiten.Key
	Char    rune
	Bubbles bool
}

// EventType defines types of UI events
type EventType int

const (
	EventClick EventType = iota
	EventHover
	EventLeave
	EventKeyPress
	EventKeyRelease
	EventFocus
	EventBlur
	EventScroll
	EventDragStart
	EventDrag
	EventDragEnd
)

// Note: AnimationState is now defined in animation.go

// Theme represents a collection of styles
type Theme struct {
	Name   string
	Styles map[string]*Style
}
