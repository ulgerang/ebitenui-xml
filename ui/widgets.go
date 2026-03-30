package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// Widget-specific constants
const (
	// Slider dimensions
	sliderTrackHeight = 4.0
	sliderThumbSize   = 16.0

	// Checkbox dimensions
	checkboxBoxSize = 18.0

	// Progress bar
	progressBarCornerRadiusFactor = 0.5
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

// IntrinsicWidth returns the button's natural width based on label and padding
func (b *Button) IntrinsicWidth() float64 {
	if b.Label == "" || b.FontFace == nil {
		return 0
	}
	tw, _ := text.Measure(b.Label, b.FontFace, 0)
	bw := b.style.BorderWidth
	return tw + b.style.Padding.Left + b.style.Padding.Right + bw*2
}

// IntrinsicHeight returns the button's natural height based on label and padding
func (b *Button) IntrinsicHeight() float64 {
	if b.Label == "" || b.FontFace == nil {
		return 0
	}
	_, th := text.Measure(b.Label, b.FontFace, 0)
	bw := b.style.BorderWidth
	return th + b.style.Padding.Top + b.style.Padding.Bottom + bw*2
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
		// Use cap-height for more visually balanced vertical centering if available,
		// otherwise fall back to em-height.
		ascent := metrics.HAscent
		descent := metrics.HDescent
		emHeight := ascent + descent

		rContent := b.ContentRect()
		x := rContent.X + (rContent.W-textW)/2
		y := rContent.Y + (rContent.H-emHeight)/2

		// Draw text shadow for button label
		shadow := style.parsedTextShadow
		if shadow == nil && style.TextShadow != "" {
			shadow = ParseTextShadow(style.TextShadow)
			style.parsedTextShadow = shadow
		}
		if shadow != nil {
			shadowOp := &text.DrawOptions{}
			shadowOp.GeoM.Translate(snapToPixel(x+shadow.OffsetX), snapToPixel(y+shadow.OffsetY))
			shadowColor := shadow.Color
			if shadowColor == nil {
				shadowColor = color.RGBA{0, 0, 0, 128}
			}
			shadowOp.ColorScale.ScaleWithColor(shadowColor)
			text.Draw(screen, b.Label, b.FontFace, shadowOp)
		}

		// Original label drawing
		op := &text.DrawOptions{}
		op.GeoM.Translate(snapToPixel(x), snapToPixel(y))
		op.ColorScale.ScaleWithColor(textColor)
		text.Draw(screen, b.Label, b.FontFace, op)
	}
}

// Text is a text display widget with word wrapping support
type Text struct {
	*BaseWidget
	Content  string
	FontFace text.Face

	// Cached layout state
	wrappedLines     []string
	lastWidth        float64
	layoutCache      *textLayout
	layoutText       string
	layoutFace       text.Face
	layoutLineHeight float64
	layoutWrap       bool
	HoveredCluster   int
	onClusterHover   func(TextHit)
	onClusterLeave   func()
}

// NewText creates a new text widget
func NewText(id, content string) *Text {
	return &Text{
		BaseWidget:     NewBaseWidget(id, "text"),
		Content:        content,
		HoveredCluster: -1,
	}
}

// IntrinsicWidth returns the text's natural width
func (t *Text) IntrinsicWidth() float64 {
	if t.Content == "" || t.FontFace == nil {
		return 0
	}
	tw, _ := text.Measure(t.Content, t.FontFace, 0)
	bw := t.style.BorderWidth
	return tw + t.style.Padding.Left + t.style.Padding.Right + bw*2
}

// IntrinsicHeight returns the text's natural height
func (t *Text) IntrinsicHeight() float64 {
	if t.Content == "" || t.FontFace == nil {
		return 0
	}
	_, th := text.Measure(t.Content, t.FontFace, 0)
	bw := t.style.BorderWidth
	return th + t.style.Padding.Top + t.style.Padding.Bottom + bw*2
}

// SetContent sets the text content and invalidates cache
func (t *Text) SetContent(content string) {
	if t.Content != content {
		t.Content = content
		t.invalidateLayout()
	}
}

func (t *Text) invalidateLayout() {
	t.wrappedLines = nil
	t.layoutCache = nil
	t.layoutText = ""
	t.layoutFace = nil
	t.layoutLineHeight = 0
	t.layoutWrap = false
	t.lastWidth = 0
	t.ClearHoveredCluster()
}

// OnClusterHover registers a handler that fires when the hovered grapheme cluster changes.
func (t *Text) OnClusterHover(handler func(TextHit)) {
	t.onClusterHover = handler
}

// OnClusterLeave registers a handler that fires when the hovered grapheme cluster is cleared.
func (t *Text) OnClusterLeave(handler func()) {
	t.onClusterLeave = handler
}

// ClearHoveredCluster clears the current hovered grapheme cluster.
func (t *Text) ClearHoveredCluster() {
	if t.HoveredCluster == -1 {
		return
	}
	t.HoveredCluster = -1
	if t.onClusterLeave != nil {
		t.onClusterLeave()
	}
}

func (t *Text) ensureLayout(maxWidth float64, style *Style) *textLayout {
	if t.FontFace == nil {
		return nil
	}
	if style == nil {
		style = t.getActiveStyle()
	}

	lineHeight := style.LineHeight
	if lineHeight <= 0 {
		lineHeight = measureLineHeight(t.FontFace)
	}
	wrap := style.TextWrap != "nowrap"
	displayText := t.Content
	if !wrap && style.TextOverflow == "ellipsis" && maxWidth > 0 {
		displayText = truncateTextWithEllipsis(displayText, t.FontFace, maxWidth)
	}

	if t.layoutCache != nil &&
		t.layoutText == displayText &&
		t.layoutFace == t.FontFace &&
		t.layoutLineHeight == lineHeight &&
		t.layoutWrap == wrap &&
		t.lastWidth == maxWidth {
		return t.layoutCache
	}

	t.layoutCache = newTextLayout(displayText, t.FontFace, textLayoutOptions{
		MaxWidth:               maxWidth,
		Wrap:                   wrap,
		WhiteSpace:             textWhiteSpaceNormal,
		LineHeight:             lineHeight,
		TrimTrailingWhitespace: true,
	})
	t.layoutText = displayText
	t.layoutFace = t.FontFace
	t.layoutLineHeight = lineHeight
	t.layoutWrap = wrap
	t.lastWidth = maxWidth

	t.wrappedLines = t.wrappedLines[:0]
	for _, line := range t.layoutCache.lines {
		t.wrappedLines = append(t.wrappedLines, line.Text)
	}
	return t.layoutCache
}

func (t *Text) textStartY(r Rect, layout *textLayout, style *Style) float64 {
	startY := r.Y
	if style.VerticalAlign == "center" {
		startY = r.Y + (r.H-layout.height)/2
	} else if style.VerticalAlign == "bottom" {
		startY = r.Y + r.H - layout.height
	}
	return startY
}

func (t *Text) lineOriginX(r Rect, lineWidth float64, style *Style) float64 {
	x := r.X
	if style.TextAlign == "center" {
		x = r.X + (r.W-lineWidth)/2
	} else if style.TextAlign == "right" {
		x = r.X + r.W - lineWidth
	}
	return x
}

// HitTest returns the grapheme cluster under the given absolute coordinates.
func (t *Text) HitTest(x, y float64) (TextHit, bool) {
	if t.FontFace == nil || !t.visible {
		return TextHit{}, false
	}
	style := t.getActiveStyle()
	r := t.ContentRect()
	layout := t.ensureLayout(r.W, style)
	if layout == nil || len(layout.lines) == 0 {
		return TextHit{}, false
	}

	lineHeight := layout.lineHeight
	lineTop := t.textStartY(r, layout, style)
	for lineIndex, line := range layout.lines {
		if y < lineTop || y > lineTop+lineHeight {
			lineTop += lineHeight
			continue
		}
		lineX := t.lineOriginX(r, line.Width, style)
		localX := x - lineX
		if localX < 0 || localX > line.Width {
			return TextHit{}, false
		}
		for clusterIndex := line.StartCluster; clusterIndex < line.EndCluster; clusterIndex++ {
			cluster := layout.clusters[clusterIndex]
			if localX < cluster.X || localX > cluster.X+cluster.Width {
				continue
			}
			return TextHit{
				LineIndex:    lineIndex,
				ClusterIndex: clusterIndex,
				Text:         cluster.Text,
				RuneStart:    cluster.RuneStart,
				RuneEnd:      cluster.RuneEnd,
				Rect: Rect{
					X: lineX + cluster.X,
					Y: lineTop,
					W: cluster.Width,
					H: lineHeight,
				},
			}, true
		}
		return TextHit{}, false
	}
	return TextHit{}, false
}

// HandlePointerMove updates the hovered grapheme cluster for the given absolute coordinates.
func (t *Text) HandlePointerMove(x, y float64) {
	hit, ok := t.HitTest(x, y)
	if !ok {
		t.ClearHoveredCluster()
		return
	}
	if hit.ClusterIndex == t.HoveredCluster {
		return
	}
	t.HoveredCluster = hit.ClusterIndex
	if t.onClusterHover != nil {
		t.onClusterHover(hit)
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

	style := t.getActiveStyle()
	r := t.ContentRect()

	textColor := style.TextColor
	if textColor == nil {
		textColor = color.White
	}

	// Apply opacity if needed
	if style.Opacity > 0 && style.Opacity < 1 {
		textColor = applyOpacity(textColor, style.Opacity)
	}

	layout := t.ensureLayout(r.W, style)
	if layout == nil {
		return
	}

	// Draw each line — use font metrics for precise vertical positioning
	metrics := t.FontFace.Metrics()
	emHeight := metrics.HAscent + metrics.HDescent
	halfLeading := (layout.lineHeight - emHeight) / 2
	startY := t.textStartY(r, layout, style)

	// In Ebitengine v2 text/v2, the origin is the top-left of the glyph's em-box.
	// So we only need to move to the top of the centered em-box.
	y := startY + halfLeading
	for _, line := range layout.lines {
		x := t.lineOriginX(r, line.Width, style)

		// Draw text shadow first (behind the text)
		shadow := style.parsedTextShadow
		if shadow == nil && style.TextShadow != "" {
			shadow = ParseTextShadow(style.TextShadow)
			style.parsedTextShadow = shadow
		}
		if shadow != nil {
			shadowOp := &text.DrawOptions{}
			shadowOp.GeoM.Translate(snapToPixel(x+shadow.OffsetX), snapToPixel(y+shadow.OffsetY))
			shadowColor := shadow.Color
			if shadowColor == nil {
				shadowColor = color.RGBA{0, 0, 0, 128}
			}
			shadowOp.ColorScale.ScaleWithColor(shadowColor)
			text.Draw(screen, line.Text, t.FontFace, shadowOp)
		}

		// Original text drawing
		op := &text.DrawOptions{}
		op.GeoM.Translate(snapToPixel(x), snapToPixel(y))
		op.ColorScale.ScaleWithColor(textColor)
		text.Draw(screen, line.Text, t.FontFace, op)

		y += layout.lineHeight
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
	Value      float64 // Current value in [Min, Max]
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
	norm := s.normalizedValue()

	// Draw track
	trackY := r.Y + (r.H-sliderTrackHeight)/2
	trackRect := Rect{X: r.X, Y: trackY, W: r.W, H: sliderTrackHeight}
	DrawRoundedRectPath(screen, trackRect, 2, s.TrackColor)

	// Draw filled portion
	fillW := r.W * norm
	fillRect := Rect{X: r.X, Y: trackY, W: fillW, H: sliderTrackHeight}
	DrawRoundedRectPath(screen, fillRect, 2, s.ThumbColor)

	// Draw thumb
	thumbX := r.X + fillW - sliderThumbSize/2
	thumbY := r.Y + (r.H-sliderThumbSize)/2
	thumbRect := Rect{X: thumbX, Y: thumbY, W: sliderThumbSize, H: sliderThumbSize}
	DrawRoundedRectPath(screen, thumbRect, sliderThumbSize/2, s.ThumbColor)
}

func (s *Slider) normalizedValue() float64 {
	valueRange := s.Max - s.Min
	if valueRange <= 0 {
		return 0
	}
	return clamp((s.Value-s.Min)/valueRange, 0, 1)
}

// SetValue sets the slider value while clamping to [Min, Max].
func (s *Slider) SetValue(value float64) {
	if s.Max < s.Min {
		s.Min, s.Max = s.Max, s.Min
	}
	value = clamp(value, s.Min, s.Max)
	if s.Value == value {
		return
	}
	s.Value = value
	if s.OnChange != nil {
		s.OnChange(value)
	}
}

func (s *Slider) setValueFromCursor(mouseX float64) {
	if !s.enabled {
		return
	}
	r := s.computedRect
	if r.W <= 0 {
		return
	}
	// FIX: mouseX is now widget-relative (already has r.X subtracted)
	// Bug was: expected absolute coordinates but parameter name suggested relative
	ratio := clamp(mouseX/r.W, 0, 1)
	s.SetValue(s.Min + ratio*(s.Max-s.Min))
}

// HandleClick updates slider value based on cursor position.
func (s *Slider) HandleClick() {
	if !s.enabled {
		return
	}
	mx, _ := ebiten.CursorPosition()
	rect := s.ComputedRect()
	// FIX: Convert absolute screen coordinates to widget-relative coordinates
	// Bug: Previously passed absolute mx directly, causing incorrect value calculation
	// when widget wasn't at screen position (0, 0)
	s.setValueFromCursor(float64(mx) - rect.X)
	if s.onClickHandler != nil {
		s.onClickHandler()
	}
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

	// Draw checkbox box
	boxRect := Rect{X: r.X, Y: r.Y + (r.H-checkboxBoxSize)/2, W: checkboxBoxSize, H: checkboxBoxSize}

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
			W: checkboxBoxSize - 8,
			H: checkboxBoxSize - 8,
		}
		DrawRoundedRectPath(screen, checkRect, 2, c.CheckColor)
	}

	// Draw label
	if c.Label != "" && c.FontFace != nil {
		textColor := c.style.TextColor
		if textColor == nil {
			textColor = color.White
		}

		// In Ebitengine v2 text/v2, the origin is the top-left of the glyph's em-box.
		x := r.X + checkboxBoxSize + 8
		metrics := c.FontFace.Metrics()
		emHeight := metrics.HAscent + metrics.HDescent
		y := r.Y + (r.H-emHeight)/2

		op := &text.DrawOptions{}
		op.GeoM.Translate(snapToPixel(x), snapToPixel(y))
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
