package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// Panel is a container widget (like <div>)
type Panel struct {
	*BaseWidget
}

// NewPanel creates a new panel widget
func NewPanel(id string) *Panel {
	return &Panel{
		BaseWidget: NewBaseWidget(id, "panel"),
	}
}

// Button is a clickable button widget
type Button struct {
	*BaseWidget
	Label    string
	FontFace text.Face
}

// NewButton creates a new button widget
func NewButton(id, label string) *Button {
	return &Button{
		BaseWidget: NewBaseWidget(id, "button"),
		Label:      label,
	}
}

// Draw renders the button
func (b *Button) Draw(screen *ebiten.Image) {
	if !b.visible {
		return
	}

	// Draw base (background/border)
	b.BaseWidget.Draw(screen)

	// Draw label text
	if b.Label != "" && b.FontFace != nil {
		style := b.getActiveStyle()
		r := b.computedRect

		textColor := style.TextColor
		if textColor == nil {
			textColor = color.White
		}

		// Apply opacity if needed
		if style.Opacity > 0 && style.Opacity < 1 {
			textColor = applyOpacity(textColor, style.Opacity)
		}

		// Measure text for centering
		textW, _ := text.Measure(b.Label, b.FontFace, 0)
		metrics := b.FontFace.Metrics()
		ascent := metrics.HAscent
		emHeight := ascent + metrics.HDescent
		x := r.X + (r.W-textW)/2
		y := r.Y + (r.H-emHeight)/2 + ascent

		// Draw text shadow for button label
		shadow := style.parsedTextShadow
		if shadow == nil && style.TextShadow != "" {
			shadow = ParseTextShadow(style.TextShadow)
			style.parsedTextShadow = shadow
		}
		if shadow != nil {
			shadowOp := &text.DrawOptions{}
			shadowOp.GeoM.Translate(x+shadow.OffsetX, y+shadow.OffsetY)
			shadowColor := shadow.Color
			if shadowColor == nil {
				shadowColor = color.RGBA{0, 0, 0, 128}
			}
			shadowOp.ColorScale.ScaleWithColor(shadowColor)
			text.Draw(screen, b.Label, b.FontFace, shadowOp)
		}

		// Original label drawing
		op := &text.DrawOptions{}
		op.GeoM.Translate(x, y)
		op.ColorScale.ScaleWithColor(textColor)
		text.Draw(screen, b.Label, b.FontFace, op)
	}
}

// Text is a text display widget with word wrapping support
type Text struct {
	*BaseWidget
	Content  string
	FontFace text.Face

	// Cached wrapped lines
	wrappedLines []string
	lastWidth    float64
}

// NewText creates a new text widget
func NewText(id, content string) *Text {
	return &Text{
		BaseWidget: NewBaseWidget(id, "text"),
		Content:    content,
	}
}

// SetContent sets the text content and invalidates cache
func (t *Text) SetContent(content string) {
	if t.Content != content {
		t.Content = content
		t.wrappedLines = nil // invalidate cache
	}
}

// Draw renders the text with word wrapping
func (t *Text) Draw(screen *ebiten.Image) {
	if !t.visible || t.Content == "" {
		return
	}

	// Draw base
	t.BaseWidget.Draw(screen)

	if t.FontFace == nil {
		return
	}

	style := t.style
	r := t.ContentRect()

	textColor := style.TextColor
	if textColor == nil {
		textColor = color.White
	}

	// Apply opacity if needed
	if style.Opacity > 0 && style.Opacity < 1 {
		textColor = applyOpacity(textColor, style.Opacity)
	}

	// Check if we need to wrap text
	wrapper := NewTextWrapper(t.FontFace)

	// Get line height
	lineHeight := style.LineHeight
	if lineHeight <= 0 {
		lineHeight = wrapper.LineHeight()
	}

	// Check text wrap mode
	shouldWrap := style.TextWrap != "nowrap"

	var lines []string
	if shouldWrap && r.W > 0 {
		// Re-wrap if width changed
		if t.wrappedLines == nil || t.lastWidth != r.W {
			t.wrappedLines = wrapper.WrapText(t.Content, r.W)
			t.lastWidth = r.W
		}
		lines = t.wrappedLines
	} else {
		lines = []string{t.Content}
	}

	// Apply text overflow
	if style.TextOverflow == "ellipsis" && len(lines) == 1 {
		lines[0] = wrapper.TruncateWithEllipsis(lines[0], r.W)
	}

	// Draw each line — use font metrics for precise vertical positioning
	metrics := t.FontFace.Metrics()
	ascent := metrics.HAscent
	emHeight := ascent + metrics.HDescent
	halfLeading := (lineHeight - emHeight) / 2

	// Vertical alignment within content rect
	totalTextHeight := float64(len(lines)) * lineHeight
	startY := r.Y // default: top
	if style.VerticalAlign == "center" {
		startY = r.Y + (r.H-totalTextHeight)/2
	} else if style.VerticalAlign == "bottom" {
		startY = r.Y + r.H - totalTextHeight
	}

	y := startY + halfLeading + ascent
	for _, line := range lines {
		// Calculate x position based on text alignment
		x := r.X
		if style.TextAlign == "center" {
			lineW, _ := text.Measure(line, t.FontFace, 0)
			x = r.X + (r.W-lineW)/2
		} else if style.TextAlign == "right" {
			lineW, _ := text.Measure(line, t.FontFace, 0)
			x = r.X + r.W - lineW
		}

		// Draw text shadow first (behind the text)
		shadow := style.parsedTextShadow
		if shadow == nil && style.TextShadow != "" {
			shadow = ParseTextShadow(style.TextShadow)
			style.parsedTextShadow = shadow
		}
		if shadow != nil {
			shadowOp := &text.DrawOptions{}
			shadowOp.GeoM.Translate(x+shadow.OffsetX, y+shadow.OffsetY)
			shadowColor := shadow.Color
			if shadowColor == nil {
				shadowColor = color.RGBA{0, 0, 0, 128}
			}
			shadowOp.ColorScale.ScaleWithColor(shadowColor)
			text.Draw(screen, line, t.FontFace, shadowOp)
		}

		// Original text drawing
		op := &text.DrawOptions{}
		op.GeoM.Translate(x, y)
		op.ColorScale.ScaleWithColor(textColor)
		text.Draw(screen, line, t.FontFace, op)

		y += lineHeight
	}
}

// Image is an image display widget
type Image struct {
	*BaseWidget
	Source *ebiten.Image
}

// NewImage creates a new image widget
func NewImage(id string) *Image {
	return &Image{
		BaseWidget: NewBaseWidget(id, "image"),
	}
}

// Draw renders the image
func (img *Image) Draw(screen *ebiten.Image) {
	if !img.visible {
		return
	}

	// Draw base
	img.BaseWidget.Draw(screen)

	if img.Source != nil {
		r := img.computedRect
		srcW := float64(img.Source.Bounds().Dx())
		srcH := float64(img.Source.Bounds().Dy())

		op := &ebiten.DrawImageOptions{}

		// Scale based on backgroundSize property
		style := img.style
		switch style.BackgroundSize {
		case "contain":
			// Scale to fit within bounds while maintaining aspect ratio
			scaleX := r.W / srcW
			scaleY := r.H / srcH
			scale := min(scaleX, scaleY)
			op.GeoM.Scale(scale, scale)
			// Center
			w, h := srcW*scale, srcH*scale
			op.GeoM.Translate(r.X+(r.W-w)/2, r.Y+(r.H-h)/2)
		case "cover":
			// Scale to cover bounds while maintaining aspect ratio
			scaleX := r.W / srcW
			scaleY := r.H / srcH
			scale := max(scaleX, scaleY)
			op.GeoM.Scale(scale, scale)
			// Center
			w, h := srcW*scale, srcH*scale
			op.GeoM.Translate(r.X+(r.W-w)/2, r.Y+(r.H-h)/2)
		default:
			// Stretch to fit (default)
			scaleX := r.W / srcW
			scaleY := r.H / srcH
			op.GeoM.Scale(scaleX, scaleY)
			op.GeoM.Translate(r.X, r.Y)
		}

		// Apply opacity
		if img.style.Opacity > 0 && img.style.Opacity < 1 {
			op.ColorScale.SetA(float32(img.style.Opacity))
		}

		screen.DrawImage(img.Source, op)
	}
}

// Note: max function is defined in effects.go

// ProgressBar is a progress indicator widget
type ProgressBar struct {
	*BaseWidget
	Value     float64 // 0.0 - 1.0
	FillColor color.Color
}

// NewProgressBar creates a new progress bar widget
func NewProgressBar(id string) *ProgressBar {
	return &ProgressBar{
		BaseWidget: NewBaseWidget(id, "progressbar"),
		Value:      0,
		FillColor:  color.RGBA{76, 175, 80, 255}, // Green
	}
}

// Draw renders the progress bar
func (p *ProgressBar) Draw(screen *ebiten.Image) {
	if !p.visible {
		return
	}

	// Draw base (background)
	p.BaseWidget.Draw(screen)

	// Draw fill
	r := p.computedRect
	padding := p.style.Padding
	fillW := (r.W - padding.Left - padding.Right) * p.Value

	if fillW > 0 {
		fillRect := Rect{
			X: r.X + padding.Left,
			Y: r.Y + padding.Top,
			W: fillW,
			H: r.H - padding.Top - padding.Bottom,
		}
		DrawRoundedRectPath(screen, fillRect, p.style.BorderRadius*0.5, p.FillColor)
	}
}

// Slider is an interactive slider widget
type Slider struct {
	*BaseWidget
	Value      float64 // 0.0 - 1.0
	Min, Max   float64
	TrackColor color.Color
	ThumbColor color.Color
	OnChange   func(value float64)

	dragging bool
}

// NewSlider creates a new slider widget
func NewSlider(id string) *Slider {
	return &Slider{
		BaseWidget: NewBaseWidget(id, "slider"),
		Min:        0,
		Max:        1,
		Value:      0.5,
		TrackColor: color.RGBA{60, 60, 60, 255},
		ThumbColor: color.RGBA{100, 149, 237, 255},
	}
}

// Draw renders the slider
func (s *Slider) Draw(screen *ebiten.Image) {
	if !s.visible {
		return
	}

	r := s.computedRect

	// Draw track
	trackHeight := 4.0
	trackY := r.Y + (r.H-trackHeight)/2
	trackRect := Rect{X: r.X, Y: trackY, W: r.W, H: trackHeight}
	DrawRoundedRectPath(screen, trackRect, 2, s.TrackColor)

	// Draw filled portion
	fillW := r.W * s.Value
	fillRect := Rect{X: r.X, Y: trackY, W: fillW, H: trackHeight}
	DrawRoundedRectPath(screen, fillRect, 2, s.ThumbColor)

	// Draw thumb
	thumbSize := 16.0
	thumbX := r.X + fillW - thumbSize/2
	thumbY := r.Y + (r.H-thumbSize)/2
	thumbRect := Rect{X: thumbX, Y: thumbY, W: thumbSize, H: thumbSize}
	DrawRoundedRectPath(screen, thumbRect, thumbSize/2, s.ThumbColor)
}

// Checkbox is a toggle widget
type Checkbox struct {
	*BaseWidget
	Checked    bool
	Label      string
	FontFace   text.Face
	OnChange   func(checked bool)
	CheckColor color.Color
}

// NewCheckbox creates a new checkbox widget
func NewCheckbox(id, label string) *Checkbox {
	return &Checkbox{
		BaseWidget: NewBaseWidget(id, "checkbox"),
		Label:      label,
		CheckColor: color.RGBA{100, 149, 237, 255},
	}
}

// Draw renders the checkbox
func (c *Checkbox) Draw(screen *ebiten.Image) {
	if !c.visible {
		return
	}

	r := c.computedRect
	boxSize := 18.0

	// Draw checkbox box
	boxRect := Rect{X: r.X, Y: r.Y + (r.H-boxSize)/2, W: boxSize, H: boxSize}

	bgColor := c.style.BackgroundColor
	if bgColor == nil {
		bgColor = color.RGBA{40, 40, 40, 255}
	}
	DrawRoundedRectPath(screen, boxRect, 3, bgColor)

	// Draw border
	borderColor := c.style.BorderColor
	if borderColor == nil {
		borderColor = color.RGBA{100, 100, 100, 255}
	}
	drawRoundedRectStroke(screen, boxRect, 3, 1, borderColor)

	// Draw check mark if checked
	if c.Checked {
		checkRect := Rect{
			X: boxRect.X + 4,
			Y: boxRect.Y + 4,
			W: boxSize - 8,
			H: boxSize - 8,
		}
		DrawRoundedRectPath(screen, checkRect, 2, c.CheckColor)
	}

	// Draw label
	if c.Label != "" && c.FontFace != nil {
		textColor := c.style.TextColor
		if textColor == nil {
			textColor = color.White
		}

		x := r.X + boxSize + 8
		metrics := c.FontFace.Metrics()
		ascent := metrics.HAscent
		emHeight := ascent + metrics.HDescent
		y := r.Y + (r.H-emHeight)/2 + ascent

		op := &text.DrawOptions{}
		op.GeoM.Translate(x, y)
		op.ColorScale.ScaleWithColor(textColor)
		text.Draw(screen, c.Label, c.FontFace, op)
	}
}

// Toggle click handler
func (c *Checkbox) HandleClick() {
	if c.enabled {
		c.Checked = !c.Checked
		if c.OnChange != nil {
			c.OnChange(c.Checked)
		}
		if c.onClickHandler != nil {
			c.onClickHandler()
		}
	}
}

// SVGIcon is a widget that displays SVG content
type SVGIcon struct {
	*BaseWidget
	Document  *SVGDocument
	IconName  string // For built-in icons
	SourceURL string // Source file path
	IconColor color.Color
}

// NewSVGIcon creates a new SVG icon widget
func NewSVGIcon(id string) *SVGIcon {
	return &SVGIcon{
		BaseWidget: NewBaseWidget(id, "svg"),
	}
}

// SetIcon sets a built-in icon by name
func (s *SVGIcon) SetIcon(name string, clr color.Color, strokeWidth float64) {
	s.IconName = name
	s.IconColor = clr
	s.Document = CreateIconSVG(name, 24, clr, strokeWidth)
}

// LoadFromFile loads SVG from a file
func (s *SVGIcon) LoadFromFile(filename string) error {
	doc, err := LoadSVG(filename)
	if err != nil {
		return err
	}
	s.Document = doc
	s.SourceURL = filename
	return nil
}

// LoadFromString loads SVG from a string
func (s *SVGIcon) LoadFromString(svgContent string) error {
	doc, err := ParseSVGString(svgContent)
	if err != nil {
		return err
	}
	s.Document = doc
	return nil
}

// SetColor updates the icon color (for single-color icons)
func (s *SVGIcon) SetColor(clr color.Color) {
	s.IconColor = clr
	if s.IconName != "" {
		s.Document = CreateIconSVG(s.IconName, 24, clr, 2)
	}
}

// Draw renders the SVG icon
func (s *SVGIcon) Draw(screen *ebiten.Image) {
	if !s.visible || s.Document == nil {
		return
	}

	// Draw base (for background/border if styled)
	s.BaseWidget.Draw(screen)

	r := s.ContentRect()
	s.Document.Draw(screen, r.X, r.Y, r.W, r.H)
}
