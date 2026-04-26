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
	GapSet    bool            `json:"-"` // true if gap was explicitly set (allows zero override)
	FlexWrap  FlexWrap        `json:"flexWrap"`

	// Sizing
	Width         float64 `json:"width"`
	WidthSet      bool    `json:"-"` // true if width was explicitly set (allows zero override)
	Height        float64 `json:"height"`
	HeightSet     bool    `json:"-"` // true if height was explicitly set (allows zero override)
	MinWidth      float64 `json:"minWidth"`
	MinWidthSet   bool    `json:"-"` // true if minWidth was explicitly set (allows zero override)
	MinHeight     float64 `json:"minHeight"`
	MinHeightSet  bool    `json:"-"` // true if minHeight was explicitly set (allows zero override)
	MaxWidth      float64 `json:"maxWidth"`
	MaxWidthSet   bool    `json:"-"` // true if maxWidth was explicitly set (allows zero override)
	MaxHeight     float64 `json:"maxHeight"`
	MaxHeightSet  bool    `json:"-"`         // true if maxHeight was explicitly set (allows zero override)
	BoxSizing     string  `json:"boxSizing"` // content-box, border-box
	FlexGrow      float64 `json:"flexGrow"`
	FlexGrowSet   bool    `json:"-"` // true if flexGrow was explicitly set (allows zero override)
	FlexShrink    float64 `json:"flexShrink"`
	FlexShrinkSet bool    `json:"-"` // true if flexShrink was explicitly set (allows zero override)

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
	BorderWidth                float64     `json:"borderWidth"`
	BorderRadius               float64     `json:"borderRadius"`
	BorderRadiusSet            bool        `json:"-"` // true if borderRadius was explicitly set (allows zero override)
	BorderWidthSet             bool        `json:"-"` // true if borderWidth was explicitly set
	BorderTopWidth             float64     `json:"borderTopWidth"`
	BorderTopWidthSet          bool        `json:"-"` // true if borderTopWidth was explicitly set (allows zero override)
	BorderRightWidth           float64     `json:"borderRightWidth"`
	BorderRightWidthSet        bool        `json:"-"` // true if borderRightWidth was explicitly set (allows zero override)
	BorderBottomWidth          float64     `json:"borderBottomWidth"`
	BorderBottomWidthSet       bool        `json:"-"` // true if borderBottomWidth was explicitly set (allows zero override)
	BorderLeftWidth            float64     `json:"borderLeftWidth"`
	BorderLeftWidthSet         bool        `json:"-"` // true if borderLeftWidth was explicitly set (allows zero override)
	BorderTop                  string      `json:"borderTop"`
	BorderRight                string      `json:"borderRight"`
	BorderBottom               string      `json:"borderBottom"`
	BorderLeft                 string      `json:"borderLeft"`
	BorderTopColor             color.Color `json:"-"`
	BorderRightColor           color.Color `json:"-"`
	BorderBottomColor          color.Color `json:"-"`
	BorderLeftColor            color.Color `json:"-"`
	BorderTopLeftRadius        float64     `json:"borderTopLeftRadius"`
	BorderTopLeftRadiusSet     bool        `json:"-"` // true if borderTopLeftRadius was explicitly set (allows zero override)
	BorderTopRightRadius       float64     `json:"borderTopRightRadius"`
	BorderTopRightRadiusSet    bool        `json:"-"` // true if borderTopRightRadius was explicitly set (allows zero override)
	BorderBottomLeftRadius     float64     `json:"borderBottomLeftRadius"`
	BorderBottomLeftRadiusSet  bool        `json:"-"` // true if borderBottomLeftRadius was explicitly set (allows zero override)
	BorderBottomRightRadius    float64     `json:"borderBottomRightRadius"`
	BorderBottomRightRadiusSet bool        `json:"-"` // true if borderBottomRightRadius was explicitly set (allows zero override)

	// Text
	FontSize         float64 `json:"fontSize"`
	FontSizeSet      bool    `json:"-"` // true if fontSize was explicitly set (allows zero override)
	FontFamily       string  `json:"fontFamily"`
	FontWeight       string  `json:"fontWeight"`    // normal, bold, 100-900
	FontStyle        string  `json:"fontStyle"`     // normal, italic
	TextAlign        string  `json:"textAlign"`     // left, center, right
	VerticalAlign    string  `json:"verticalAlign"` // top, center, bottom
	LineHeight       float64 `json:"lineHeight"`
	LineHeightSet    bool    `json:"-"` // true if lineHeight was explicitly set (allows zero override)
	LetterSpacing    float64 `json:"letterSpacing"`
	LetterSpacingSet bool    `json:"-"`            // true if letterSpacing was explicitly set (allows zero override)
	TextWrap         string  `json:"textWrap"`     // normal, nowrap
	TextOverflow     string  `json:"textOverflow"` // clip, ellipsis

	// Visual Effects
	Opacity    float64 `json:"opacity"`    // 0-1
	OpacitySet bool    `json:"-"`          // true if opacity was explicitly set (allows zero override)
	BoxShadow  string  `json:"boxShadow"`  // "offsetX offsetY blur spread color"
	TextShadow string  `json:"textShadow"` // "offsetX offsetY blur color"
	Cursor     string  `json:"cursor"`     // pointer, default, etc.
	Transition string  `json:"transition"` // "property duration easing"
	Animation  string  `json:"animation"`  // "name duration easing iteration-count"

	// Outline (separate from border)
	Outline          string  `json:"outline"`       // "width style color"
	OutlineOffset    float64 `json:"outlineOffset"` // distance from border
	OutlineOffsetSet bool    `json:"-"`             // true if outlineOffset was explicitly set (allows zero override)

	// Filters
	Filter         string `json:"filter"`         // blur(), brightness(), etc.
	BackdropFilter string `json:"backdropFilter"` // for glassmorphism

	// Transform
	Transform       string `json:"transform"`       // rotate(), scale(), translate()
	TransformOrigin string `json:"transformOrigin"` // center, top left, etc.

	// Clip Path
	ClipPath string `json:"clipPath"` // circle(), polygon(), inset(), path()

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
	Position  string  `json:"position"` // relative, absolute
	Top       float64 `json:"top"`
	TopSet    bool    `json:"-"` // true if top was explicitly set (allows zero override)
	Right     float64 `json:"right"`
	RightSet  bool    `json:"-"` // true if right was explicitly set (allows zero override)
	Bottom    float64 `json:"bottom"`
	BottomSet bool    `json:"-"` // true if bottom was explicitly set (allows zero override)
	Left      float64 `json:"left"`
	LeftSet   bool    `json:"-"` // true if left was explicitly set (allows zero override)
	ZIndex    int     `json:"zIndex"`
	ZIndexSet bool    `json:"-"` // true if zIndex was explicitly set (allows zero override)

	// Display
	Display    string `json:"display"`    // block, flex, none
	Visibility string `json:"visibility"` // visible, hidden

	// States
	HoverStyle    *Style `json:"hover"`
	ActiveStyle   *Style `json:"active"`
	DisabledStyle *Style `json:"disabled"`
	FocusStyle    *Style `json:"focus"`

	// Parsed values (internal)
	parsedBoxShadow      *BoxShadow      `json:"-"`
	parsedBoxShadows     []*BoxShadow    `json:"-"`
	parsedTextShadow     *TextShadow     `json:"-"`
	parsedTextShadows    []*TextShadow   `json:"-"`
	parsedOutline        *Outline        `json:"-"`
	parsedTransitions    []Transition    `json:"-"`
	parsed9Slice         *NineSlice      `json:"-"`
	parsedGradient       *Gradient       `json:"-"`
	parsedFilter         *Filter         `json:"-"`
	parsedBackdropFilter *BackdropFilter `json:"-"`
	parsedTransform      *Transform      `json:"-"`
	parsedAnimation      *Animation      `json:"-"`
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
	if other.GapSet || other.Gap != 0 {
		s.Gap = other.Gap
		s.GapSet = other.GapSet
	}
	if other.FlexWrap != "" {
		s.FlexWrap = other.FlexWrap
	}

	// Sizing
	if other.WidthSet || other.Width != 0 {
		s.Width = other.Width
		s.WidthSet = other.WidthSet
	}
	if other.HeightSet || other.Height != 0 {
		s.Height = other.Height
		s.HeightSet = other.HeightSet
	}
	if other.MinWidthSet || other.MinWidth != 0 {
		s.MinWidth = other.MinWidth
		s.MinWidthSet = other.MinWidthSet
	}
	if other.MinHeightSet || other.MinHeight != 0 {
		s.MinHeight = other.MinHeight
		s.MinHeightSet = other.MinHeightSet
	}
	if other.MaxWidthSet || other.MaxWidth != 0 {
		s.MaxWidth = other.MaxWidth
		s.MaxWidthSet = other.MaxWidthSet
	}
	if other.MaxHeightSet || other.MaxHeight != 0 {
		s.MaxHeight = other.MaxHeight
		s.MaxHeightSet = other.MaxHeightSet
	}
	if other.BoxSizing != "" {
		s.BoxSizing = other.BoxSizing
	}
	if other.FlexGrowSet || other.FlexGrow != 0 {
		s.FlexGrow = other.FlexGrow
		s.FlexGrowSet = other.FlexGrowSet
	}
	if other.FlexShrinkSet || other.FlexShrink != 0 {
		s.FlexShrink = other.FlexShrink
		s.FlexShrinkSet = other.FlexShrinkSet
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
	if other.BorderBottom != "" {
		s.BorderBottom = other.BorderBottom
	}
	if other.BorderBottomColor != nil {
		s.BorderBottomColor = other.BorderBottomColor
	}
	if other.BorderLeft != "" {
		s.BorderLeft = other.BorderLeft
	}
	if other.BorderLeftColor != nil {
		s.BorderLeftColor = other.BorderLeftColor
	}
	if other.BorderTop != "" {
		s.BorderTop = other.BorderTop
	}
	if other.BorderTopColor != nil {
		s.BorderTopColor = other.BorderTopColor
	}
	if other.BorderRight != "" {
		s.BorderRight = other.BorderRight
	}
	if other.BorderRightColor != nil {
		s.BorderRightColor = other.BorderRightColor
	}
	if other.Color != "" {
		s.Color = other.Color
	}

	// Border
	if other.BorderWidthSet || other.BorderWidth != 0 {
		s.BorderWidth = other.BorderWidth
		s.BorderWidthSet = true
	}
	if other.BorderTopWidthSet || other.BorderTopWidth != 0 {
		s.BorderTopWidth = other.BorderTopWidth
		s.BorderTopWidthSet = other.BorderTopWidthSet
	}
	if other.BorderRightWidthSet || other.BorderRightWidth != 0 {
		s.BorderRightWidth = other.BorderRightWidth
		s.BorderRightWidthSet = other.BorderRightWidthSet
	}
	if other.BorderBottomWidthSet || other.BorderBottomWidth != 0 {
		s.BorderBottomWidth = other.BorderBottomWidth
		s.BorderBottomWidthSet = other.BorderBottomWidthSet
	}
	if other.BorderLeftWidthSet || other.BorderLeftWidth != 0 {
		s.BorderLeftWidth = other.BorderLeftWidth
		s.BorderLeftWidthSet = other.BorderLeftWidthSet
	}
	if other.BorderRadiusSet || other.BorderRadius != 0 {
		s.BorderRadius = other.BorderRadius
		s.BorderRadiusSet = other.BorderRadiusSet
	}
	if other.BorderTopLeftRadiusSet || other.BorderTopLeftRadius != 0 {
		s.BorderTopLeftRadius = other.BorderTopLeftRadius
		s.BorderTopLeftRadiusSet = other.BorderTopLeftRadiusSet
	}
	if other.BorderTopRightRadiusSet || other.BorderTopRightRadius != 0 {
		s.BorderTopRightRadius = other.BorderTopRightRadius
		s.BorderTopRightRadiusSet = other.BorderTopRightRadiusSet
	}
	if other.BorderBottomLeftRadiusSet || other.BorderBottomLeftRadius != 0 {
		s.BorderBottomLeftRadius = other.BorderBottomLeftRadius
		s.BorderBottomLeftRadiusSet = other.BorderBottomLeftRadiusSet
	}
	if other.BorderBottomRightRadiusSet || other.BorderBottomRightRadius != 0 {
		s.BorderBottomRightRadius = other.BorderBottomRightRadius
		s.BorderBottomRightRadiusSet = other.BorderBottomRightRadiusSet
	}

	// Text
	if other.FontSizeSet || other.FontSize != 0 {
		s.FontSize = other.FontSize
		s.FontSizeSet = other.FontSizeSet
	}
	if other.FontFamily != "" {
		s.FontFamily = other.FontFamily
	}
	if other.TextAlign != "" {
		s.TextAlign = other.TextAlign
	}
	if other.LineHeightSet || other.LineHeight != 0 {
		s.LineHeight = other.LineHeight
		s.LineHeightSet = other.LineHeightSet
	}
	if other.TextWrap != "" {
		s.TextWrap = other.TextWrap
	}
	if other.TextOverflow != "" {
		s.TextOverflow = other.TextOverflow
	}
	if other.FontWeight != "" {
		s.FontWeight = other.FontWeight
	}
	if other.FontStyle != "" {
		s.FontStyle = other.FontStyle
	}
	if other.VerticalAlign != "" {
		s.VerticalAlign = other.VerticalAlign
	}
	if other.LetterSpacingSet || other.LetterSpacing != 0 {
		s.LetterSpacing = other.LetterSpacing
		s.LetterSpacingSet = other.LetterSpacingSet
	}

	// Visual Effects
	if other.OpacitySet || other.Opacity != 0 {
		s.Opacity = other.Opacity
		s.OpacitySet = other.OpacitySet
	}
	if other.BoxShadow != "" {
		s.BoxShadow = other.BoxShadow
		s.parsedBoxShadow = other.parsedBoxShadow
		s.parsedBoxShadows = other.parsedBoxShadows
	}
	if other.TextShadow != "" {
		s.TextShadow = other.TextShadow
		s.parsedTextShadow = other.parsedTextShadow
		s.parsedTextShadows = other.parsedTextShadows
	}
	if other.Outline != "" {
		s.Outline = other.Outline
		s.parsedOutline = other.parsedOutline
	}
	if other.OutlineOffsetSet || other.OutlineOffset != 0 {
		s.OutlineOffset = other.OutlineOffset
		s.OutlineOffsetSet = other.OutlineOffsetSet
	}
	if other.Transition != "" {
		s.Transition = other.Transition
		s.parsedTransitions = other.parsedTransitions
	}
	if other.Animation != "" {
		s.Animation = other.Animation
		s.parsedAnimation = other.parsedAnimation
	}

	// Filter
	if other.Filter != "" {
		s.Filter = other.Filter
		s.parsedFilter = other.parsedFilter
	}
	if other.BackdropFilter != "" {
		s.BackdropFilter = other.BackdropFilter
		s.parsedBackdropFilter = other.parsedBackdropFilter
	}

	// Transform
	if other.Transform != "" {
		s.Transform = other.Transform
		s.parsedTransform = other.parsedTransform
	}
	if other.TransformOrigin != "" {
		s.TransformOrigin = other.TransformOrigin
	}

	// Clip Path
	if other.ClipPath != "" {
		s.ClipPath = other.ClipPath
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
	if other.BackgroundSize != "" {
		s.BackgroundSize = other.BackgroundSize
	}
	if other.BackgroundPosition != "" {
		s.BackgroundPosition = other.BackgroundPosition
	}
	if other.BackgroundRepeat != "" {
		s.BackgroundRepeat = other.BackgroundRepeat
	}

	// Overflow
	if other.Overflow != "" {
		s.Overflow = other.Overflow
	}

	// Position
	if other.Position != "" {
		s.Position = other.Position
	}
	if other.TopSet || other.Top != 0 {
		s.Top = other.Top
		s.TopSet = other.TopSet
	}
	if other.RightSet || other.Right != 0 {
		s.Right = other.Right
		s.RightSet = other.RightSet
	}
	if other.BottomSet || other.Bottom != 0 {
		s.Bottom = other.Bottom
		s.BottomSet = other.BottomSet
	}
	if other.LeftSet || other.Left != 0 {
		s.Left = other.Left
		s.LeftSet = other.LeftSet
	}
	if other.ZIndexSet || other.ZIndex != 0 {
		s.ZIndex = other.ZIndex
		s.ZIndexSet = other.ZIndexSet
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

// ValidationState represents form validation status for form-capable widgets.
type ValidationState string

const (
	ValidationNone    ValidationState = ""
	ValidationValid   ValidationState = "valid"
	ValidationInvalid ValidationState = "invalid"
)

// ValidationRules stores HTML-like form validation constraints for a widget.
type ValidationRules struct {
	Required     bool
	Min          float64
	MinSet       bool
	Max          float64
	MaxSet       bool
	MinLength    int
	MinLengthSet bool
	MaxLength    int
	MaxLengthSet bool
	Pattern      string
	Type         string
	Message      string
}

// HasConstraints returns whether any validation rule is active.
func (r ValidationRules) HasConstraints() bool {
	return r.Required || r.MinSet || r.MaxSet || r.MinLengthSet || r.MaxLengthSet || r.Pattern != "" || r.Type != ""
}

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
	IntrinsicWidth() float64
	IntrinsicHeight() float64

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
