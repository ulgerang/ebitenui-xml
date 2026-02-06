package ui

import (
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// ============================================================================
// Toggle/Switch Widget - On/Off switch (visual alternative to Checkbox)
// ============================================================================

// Toggle is a switch-style toggle widget (iOS-style on/off)
type Toggle struct {
	*BaseWidget
	Checked      bool
	Label        string
	FontFace     text.Face
	OnChange     func(checked bool)
	OnColor      color.Color // Color when on
	OffColor     color.Color // Color when off
	ThumbColor   color.Color
	animProgress float64 // 0=off, 1=on for animation
}

// NewToggle creates a new toggle widget
func NewToggle(id, label string) *Toggle {
	return &Toggle{
		BaseWidget: NewBaseWidget(id, "toggle"),
		Label:      label,
		OnColor:    color.RGBA{76, 175, 80, 255},   // Green
		OffColor:   color.RGBA{100, 100, 100, 255}, // Gray
		ThumbColor: color.White,
	}
}

// Draw renders the toggle switch
func (t *Toggle) Draw(screen *ebiten.Image) {
	if !t.visible {
		return
	}

	r := t.computedRect

	// Toggle track dimensions
	trackW := 44.0
	trackH := 24.0
	trackX := r.X
	trackY := r.Y + (r.H-trackH)/2

	// Animate towards target
	target := 0.0
	if t.Checked {
		target = 1.0
	}
	t.animProgress += (target - t.animProgress) * 0.2

	// Draw track with interpolated color
	trackColor := t.OffColor
	if t.Checked {
		trackColor = t.OnColor
	}
	trackRect := Rect{X: trackX, Y: trackY, W: trackW, H: trackH}
	DrawRoundedRectPath(screen, trackRect, trackH/2, trackColor)

	// Draw thumb (circle)
	thumbSize := trackH - 4
	thumbX := trackX + 2 + t.animProgress*(trackW-thumbSize-4)
	thumbY := trackY + 2
	thumbRect := Rect{X: thumbX, Y: thumbY, W: thumbSize, H: thumbSize}
	DrawRoundedRectPath(screen, thumbRect, thumbSize/2, t.ThumbColor)

	// Draw label
	if t.Label != "" && t.FontFace != nil {
		textColor := t.style.TextColor
		if textColor == nil {
			textColor = color.White
		}

		x := r.X + trackW + 12
		_, textH := text.Measure(t.Label, t.FontFace, 0)
		y := r.Y + (r.H+textH)/2 - textH*0.2

		op := &text.DrawOptions{}
		op.GeoM.Translate(x, y)
		op.ColorScale.ScaleWithColor(textColor)
		text.Draw(screen, t.Label, t.FontFace, op)
	}
}

// HandleClick toggles the switch
func (t *Toggle) HandleClick() {
	if t.enabled {
		t.Checked = !t.Checked
		if t.OnChange != nil {
			t.OnChange(t.Checked)
		}
		if t.onClickHandler != nil {
			t.onClickHandler()
		}
	}
}

// ============================================================================
// RadioButton & RadioGroup - Single selection option group
// ============================================================================

// RadioButton is a single radio option
type RadioButton struct {
	*BaseWidget
	Label    string
	Value    string
	Selected bool
	FontFace text.Face
	Group    *RadioGroup
}

// RadioGroup manages a group of radio buttons
type RadioGroup struct {
	ID       string
	Value    string // Currently selected value
	Buttons  []*RadioButton
	OnChange func(value string)
}

// NewRadioGroup creates a new radio button group
func NewRadioGroup(id string) *RadioGroup {
	return &RadioGroup{
		ID:      id,
		Buttons: make([]*RadioButton, 0),
	}
}

// AddButton adds a radio button to the group
func (g *RadioGroup) AddButton(btn *RadioButton) {
	btn.Group = g
	g.Buttons = append(g.Buttons, btn)
}

// SetValue sets the selected value
func (g *RadioGroup) SetValue(value string) {
	g.Value = value
	for _, btn := range g.Buttons {
		btn.Selected = (btn.Value == value)
	}
	if g.OnChange != nil {
		g.OnChange(value)
	}
}

// NewRadioButton creates a new radio button
func NewRadioButton(id, label, value string) *RadioButton {
	return &RadioButton{
		BaseWidget: NewBaseWidget(id, "radiobutton"),
		Label:      label,
		Value:      value,
	}
}

// Draw renders the radio button
func (rb *RadioButton) Draw(screen *ebiten.Image) {
	if !rb.visible {
		return
	}

	r := rb.computedRect
	circleSize := 18.0

	// Draw outer circle
	circleX := r.X
	circleY := r.Y + (r.H-circleSize)/2
	circleRect := Rect{X: circleX, Y: circleY, W: circleSize, H: circleSize}

	bgColor := rb.style.BackgroundColor
	if bgColor == nil {
		bgColor = color.RGBA{40, 40, 40, 255}
	}
	DrawRoundedRectPath(screen, circleRect, circleSize/2, bgColor)

	// Draw border
	borderColor := rb.style.BorderColor
	if borderColor == nil {
		borderColor = color.RGBA{100, 100, 100, 255}
	}
	drawRoundedRectStroke(screen, circleRect, circleSize/2, 1.5, borderColor)

	// Draw inner circle if selected
	if rb.Selected {
		selectedColor := color.RGBA{100, 149, 237, 255}
		innerSize := circleSize - 8
		innerRect := Rect{
			X: circleX + 4,
			Y: circleY + 4,
			W: innerSize,
			H: innerSize,
		}
		DrawRoundedRectPath(screen, innerRect, innerSize/2, selectedColor)
	}

	// Draw label
	if rb.Label != "" && rb.FontFace != nil {
		textColor := rb.style.TextColor
		if textColor == nil {
			textColor = color.White
		}

		x := r.X + circleSize + 8
		_, textH := text.Measure(rb.Label, rb.FontFace, 0)
		y := r.Y + (r.H+textH)/2 - textH*0.2

		op := &text.DrawOptions{}
		op.GeoM.Translate(x, y)
		op.ColorScale.ScaleWithColor(textColor)
		text.Draw(screen, rb.Label, rb.FontFace, op)
	}
}

// HandleClick selects this radio button
func (rb *RadioButton) HandleClick() {
	if rb.enabled && rb.Group != nil {
		rb.Group.SetValue(rb.Value)
	}
	if rb.onClickHandler != nil {
		rb.onClickHandler()
	}
}

// ============================================================================
// Dropdown/Select Widget - Dropdown selection
// ============================================================================

// DropdownOption represents a selectable option
type DropdownOption struct {
	Label string
	Value string
}

// Dropdown is a dropdown selection widget
type Dropdown struct {
	*BaseWidget
	Options       []DropdownOption
	SelectedIndex int
	Placeholder   string
	IsOpen        bool
	FontFace      text.Face
	OnChange      func(index int, value string)

	// Styling
	DropdownBg      color.Color
	HoverColor      color.Color
	hoveredIndex    int
	maxVisibleItems int
}

// NewDropdown creates a new dropdown widget
func NewDropdown(id string) *Dropdown {
	return &Dropdown{
		BaseWidget:      NewBaseWidget(id, "dropdown"),
		Options:         make([]DropdownOption, 0),
		SelectedIndex:   -1,
		Placeholder:     "Select...",
		DropdownBg:      color.RGBA{50, 50, 50, 255},
		HoverColor:      color.RGBA{70, 70, 70, 255},
		hoveredIndex:    -1,
		maxVisibleItems: 5,
	}
}

// AddOption adds an option to the dropdown
func (d *Dropdown) AddOption(label, value string) {
	d.Options = append(d.Options, DropdownOption{Label: label, Value: value})
}

// SetOptions sets all options at once
func (d *Dropdown) SetOptions(options []DropdownOption) {
	d.Options = options
}

// GetSelectedValue returns the currently selected value
func (d *Dropdown) GetSelectedValue() string {
	if d.SelectedIndex >= 0 && d.SelectedIndex < len(d.Options) {
		return d.Options[d.SelectedIndex].Value
	}
	return ""
}

// SetValue sets the selected value by value string
func (d *Dropdown) SetValue(value string) {
	for i, opt := range d.Options {
		if opt.Value == value {
			d.SelectedIndex = i
			return
		}
	}
}

// Draw renders the dropdown
func (d *Dropdown) Draw(screen *ebiten.Image) {
	if !d.visible {
		return
	}

	r := d.computedRect

	// Draw main button area
	d.BaseWidget.Draw(screen)

	// Draw current selection or placeholder
	displayText := d.Placeholder
	if d.SelectedIndex >= 0 && d.SelectedIndex < len(d.Options) {
		displayText = d.Options[d.SelectedIndex].Label
	}

	if d.FontFace != nil {
		textColor := d.style.TextColor
		if textColor == nil {
			textColor = color.White
		}

		// Text with padding
		textX := r.X + 12
		_, textH := text.Measure(displayText, d.FontFace, 0)
		textY := r.Y + (r.H+textH)/2 - textH*0.2

		op := &text.DrawOptions{}
		op.GeoM.Translate(textX, textY)
		op.ColorScale.ScaleWithColor(textColor)
		text.Draw(screen, displayText, d.FontFace, op)

		// Draw dropdown arrow
		arrowX := r.X + r.W - 24
		arrowY := r.Y + r.H/2
		d.drawArrow(screen, arrowX, arrowY, d.IsOpen, textColor)
	}

	// Draw dropdown list if open
	if d.IsOpen && len(d.Options) > 0 {
		d.drawDropdownList(screen)
	}
}

// drawArrow draws the dropdown arrow indicator
func (d *Dropdown) drawArrow(screen *ebiten.Image, x, y float64, up bool, col color.Color) {
	arrowSize := 6.0
	// Simple triangle approximation using a small rect
	if up {
		arrowRect := Rect{X: x - arrowSize/2, Y: y - arrowSize/4, W: arrowSize, H: arrowSize / 2}
		DrawRoundedRectPath(screen, arrowRect, 1, col)
	} else {
		arrowRect := Rect{X: x - arrowSize/2, Y: y - arrowSize/4, W: arrowSize, H: arrowSize / 2}
		DrawRoundedRectPath(screen, arrowRect, 1, col)
	}
}

// drawDropdownList draws the dropdown options list
func (d *Dropdown) drawDropdownList(screen *ebiten.Image) {
	r := d.computedRect
	itemHeight := 36.0

	visibleItems := len(d.Options)
	if visibleItems > d.maxVisibleItems {
		visibleItems = d.maxVisibleItems
	}

	listHeight := float64(visibleItems) * itemHeight
	listRect := Rect{
		X: r.X,
		Y: r.Y + r.H + 4,
		W: r.W,
		H: listHeight,
	}

	// Draw list background
	DrawRoundedRectPath(screen, listRect, 4, d.DropdownBg)
	drawRoundedRectStroke(screen, listRect, 4, 1, color.RGBA{80, 80, 80, 255})

	// Draw items
	for i := 0; i < visibleItems; i++ {
		opt := d.Options[i]
		itemY := listRect.Y + float64(i)*itemHeight
		itemRect := Rect{X: listRect.X + 2, Y: itemY + 2, W: listRect.W - 4, H: itemHeight - 4}

		// Highlight hovered or selected
		if i == d.hoveredIndex {
			DrawRoundedRectPath(screen, itemRect, 3, d.HoverColor)
		} else if i == d.SelectedIndex {
			DrawRoundedRectPath(screen, itemRect, 3, color.RGBA{60, 60, 60, 255})
		}

		// Draw label
		if d.FontFace != nil {
			textColor := d.style.TextColor
			if textColor == nil {
				textColor = color.White
			}

			textX := itemRect.X + 10
			_, textH := text.Measure(opt.Label, d.FontFace, 0)
			textY := itemY + (itemHeight+textH)/2 - textH*0.2

			op := &text.DrawOptions{}
			op.GeoM.Translate(textX, textY)
			op.ColorScale.ScaleWithColor(textColor)
			text.Draw(screen, opt.Label, d.FontFace, op)
		}
	}
}

// HandleClick handles click on dropdown
func (d *Dropdown) HandleClick() {
	if !d.enabled {
		return
	}

	if d.IsOpen {
		// Check if clicking on an option
		r := d.computedRect
		mx, my := ebiten.CursorPosition()
		itemHeight := 36.0
		listY := r.Y + r.H + 4

		visibleItems := len(d.Options)
		if visibleItems > d.maxVisibleItems {
			visibleItems = d.maxVisibleItems
		}

		for i := 0; i < visibleItems; i++ {
			itemY := listY + float64(i)*itemHeight
			if float64(my) >= itemY && float64(my) < itemY+itemHeight &&
				float64(mx) >= r.X && float64(mx) < r.X+r.W {
				d.SelectedIndex = i
				d.IsOpen = false
				if d.OnChange != nil {
					d.OnChange(i, d.Options[i].Value)
				}
				return
			}
		}
		d.IsOpen = false
	} else {
		d.IsOpen = true
	}

	if d.onClickHandler != nil {
		d.onClickHandler()
	}
}

// Update handles dropdown input (should be called every frame)
func (d *Dropdown) Update() {
	if !d.IsOpen {
		return
	}

	r := d.computedRect
	mx, my := ebiten.CursorPosition()

	// Update hover state
	itemHeight := 36.0
	listY := r.Y + r.H + 4

	d.hoveredIndex = -1
	visibleItems := len(d.Options)
	if visibleItems > d.maxVisibleItems {
		visibleItems = d.maxVisibleItems
	}

	for i := 0; i < visibleItems; i++ {
		itemY := listY + float64(i)*itemHeight
		if float64(my) >= itemY && float64(my) < itemY+itemHeight &&
			float64(mx) >= r.X && float64(mx) < r.X+r.W {
			d.hoveredIndex = i
			break
		}
	}

	// Close on click outside
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		listHeight := float64(visibleItems) * itemHeight
		if float64(mx) < r.X || float64(mx) > r.X+r.W ||
			float64(my) < r.Y || float64(my) > r.Y+r.H+4+listHeight {
			d.IsOpen = false
		}
	}
}

// ============================================================================
// Modal/Dialog Widget - Popup dialog overlay
// ============================================================================

// Modal is a popup dialog overlay
type Modal struct {
	*BaseWidget
	Title         string
	Content       string
	FontFace      text.Face
	TitleFontFace text.Face
	IsOpen        bool
	OnClose       func()

	// Buttons
	Buttons []*Button

	// Overlay
	OverlayColor color.Color
}

// NewModal creates a new modal dialog
func NewModal(id, title string) *Modal {
	return &Modal{
		BaseWidget:   NewBaseWidget(id, "modal"),
		Title:        title,
		IsOpen:       false,
		OverlayColor: color.RGBA{0, 0, 0, 180},
		Buttons:      make([]*Button, 0),
	}
}

// Open shows the modal
func (m *Modal) Open() {
	m.IsOpen = true
}

// Close hides the modal
func (m *Modal) Close() {
	m.IsOpen = false
	if m.OnClose != nil {
		m.OnClose()
	}
}

// AddButton adds a button to the modal
func (m *Modal) AddButton(btn *Button) {
	m.Buttons = append(m.Buttons, btn)
}

// Draw renders the modal
func (m *Modal) Draw(screen *ebiten.Image) {
	if !m.IsOpen || !m.visible {
		return
	}

	screenW := float64(screen.Bounds().Dx())
	screenH := float64(screen.Bounds().Dy())

	// Draw overlay
	overlayRect := Rect{X: 0, Y: 0, W: screenW, H: screenH}
	DrawRoundedRectPath(screen, overlayRect, 0, m.OverlayColor)

	// Modal dimensions
	modalW := 400.0
	modalH := 250.0

	// Content affects height
	if m.Content != "" {
		lines := strings.Count(m.Content, "\n") + 1
		modalH = 120 + float64(lines)*24 + 60
	}

	modalX := (screenW - modalW) / 2
	modalY := (screenH - modalH) / 2

	m.computedRect = Rect{X: modalX, Y: modalY, W: modalW, H: modalH}

	// Draw modal background
	bgColor := m.style.BackgroundColor
	if bgColor == nil {
		bgColor = color.RGBA{45, 45, 45, 255}
	}
	modalRect := m.computedRect
	DrawRoundedRectPath(screen, modalRect, 12, bgColor)
	drawRoundedRectStroke(screen, modalRect, 12, 1, color.RGBA{80, 80, 80, 255})

	// Draw title
	if m.Title != "" {
		titleFace := m.TitleFontFace
		if titleFace == nil {
			titleFace = m.FontFace
		}
		if titleFace != nil {
			textColor := m.style.TextColor
			if textColor == nil {
				textColor = color.White
			}

			titleW, titleH := text.Measure(m.Title, titleFace, 0)
			titleX := modalX + (modalW-titleW)/2
			titleY := modalY + 30 + titleH*0.7

			op := &text.DrawOptions{}
			op.GeoM.Translate(titleX, titleY)
			op.ColorScale.ScaleWithColor(textColor)
			text.Draw(screen, m.Title, titleFace, op)
		}
	}

	// Draw content
	if m.Content != "" && m.FontFace != nil {
		contentColor := m.style.TextColor
		if contentColor == nil {
			contentColor = color.RGBA{200, 200, 200, 255}
		}

		lines := strings.Split(m.Content, "\n")
		y := modalY + 70.0
		lineHeight := 24.0

		for _, line := range lines {
			lineW, _ := text.Measure(line, m.FontFace, 0)
			x := modalX + (modalW-lineW)/2

			op := &text.DrawOptions{}
			op.GeoM.Translate(x, y)
			op.ColorScale.ScaleWithColor(contentColor)
			text.Draw(screen, line, m.FontFace, op)

			y += lineHeight
		}
	}

	// Draw buttons
	buttonY := modalY + modalH - 60
	buttonSpacing := 10.0
	totalButtonW := 0.0

	for _, btn := range m.Buttons {
		totalButtonW += btn.computedRect.W + buttonSpacing
	}
	totalButtonW -= buttonSpacing

	buttonX := modalX + (modalW-totalButtonW)/2
	for _, btn := range m.Buttons {
		btn.computedRect.X = buttonX
		btn.computedRect.Y = buttonY
		btn.computedRect.W = 100
		btn.computedRect.H = 36
		btn.Draw(screen)
		buttonX += btn.computedRect.W + buttonSpacing
	}
}

// ============================================================================
// Tooltip Widget - Hover information popup
// ============================================================================

// Tooltip is a hover information popup
type Tooltip struct {
	*BaseWidget
	Text         string
	FontFace     text.Face
	TargetWidget Widget
	IsVisible    bool

	// Positioning
	Position string // "top", "bottom", "left", "right"
	Offset   float64
}

// NewTooltip creates a new tooltip
func NewTooltip(id, text string) *Tooltip {
	return &Tooltip{
		BaseWidget: NewBaseWidget(id, "tooltip"),
		Text:       text,
		Position:   "top",
		Offset:     8,
	}
}

// SetTarget sets the widget this tooltip is attached to
func (t *Tooltip) SetTarget(target Widget) {
	t.TargetWidget = target
}

// Show shows the tooltip
func (t *Tooltip) Show() {
	t.IsVisible = true
}

// Hide hides the tooltip
func (t *Tooltip) Hide() {
	t.IsVisible = false
}

// Draw renders the tooltip
func (t *Tooltip) Draw(screen *ebiten.Image) {
	if !t.IsVisible || t.Text == "" || t.FontFace == nil {
		return
	}

	// Calculate position based on cursor or target
	mx, my := ebiten.CursorPosition()

	// Measure text
	textW, textH := text.Measure(t.Text, t.FontFace, 0)
	paddingX := 12.0
	paddingY := 8.0

	tooltipW := textW + paddingX*2
	tooltipH := textH + paddingY*2

	// Position tooltip above cursor
	tooltipX := float64(mx) - tooltipW/2
	tooltipY := float64(my) - tooltipH - t.Offset

	// Keep on screen
	screenW := float64(screen.Bounds().Dx())
	if tooltipX < 4 {
		tooltipX = 4
	}
	if tooltipX+tooltipW > screenW-4 {
		tooltipX = screenW - 4 - tooltipW
	}
	if tooltipY < 4 {
		tooltipY = float64(my) + 20 + t.Offset // Show below cursor instead
	}

	// Draw background
	bgColor := t.style.BackgroundColor
	if bgColor == nil {
		bgColor = color.RGBA{30, 30, 30, 240}
	}
	tooltipRect := Rect{X: tooltipX, Y: tooltipY, W: tooltipW, H: tooltipH}
	DrawRoundedRectPath(screen, tooltipRect, 6, bgColor)
	drawRoundedRectStroke(screen, tooltipRect, 6, 1, color.RGBA{80, 80, 80, 255})

	// Draw text
	textColor := t.style.TextColor
	if textColor == nil {
		textColor = color.White
	}

	op := &text.DrawOptions{}
	op.GeoM.Translate(tooltipX+paddingX, tooltipY+paddingY+textH*0.75)
	op.ColorScale.ScaleWithColor(textColor)
	text.Draw(screen, t.Text, t.FontFace, op)
}

// ============================================================================
// Badge Widget - Notification badge
// ============================================================================

// Badge is a notification badge widget
type Badge struct {
	*BaseWidget
	Text       string
	FontFace   text.Face
	BadgeColor color.Color
}

// NewBadge creates a new badge widget
func NewBadge(id, text string) *Badge {
	return &Badge{
		BaseWidget: NewBaseWidget(id, "badge"),
		Text:       text,
		BadgeColor: color.RGBA{220, 53, 69, 255}, // Red
	}
}

// Draw renders the badge
func (b *Badge) Draw(screen *ebiten.Image) {
	if !b.visible || b.Text == "" {
		return
	}

	if b.FontFace == nil {
		return
	}

	// Measure text
	textW, textH := text.Measure(b.Text, b.FontFace, 0)

	// Badge dimensions
	minW := 20.0
	paddingX := 8.0
	badgeW := textW + paddingX*2
	if badgeW < minW {
		badgeW = minW
	}
	badgeH := textH + 8

	r := b.computedRect
	badgeRect := Rect{
		X: r.X,
		Y: r.Y,
		W: badgeW,
		H: badgeH,
	}

	// Draw background
	DrawRoundedRectPath(screen, badgeRect, badgeH/2, b.BadgeColor)

	// Draw text
	textColor := b.style.TextColor
	if textColor == nil {
		textColor = color.White
	}

	textX := r.X + (badgeW-textW)/2
	textY := r.Y + (badgeH+textH)/2 - textH*0.2

	op := &text.DrawOptions{}
	op.GeoM.Translate(textX, textY)
	op.ColorScale.ScaleWithColor(textColor)
	text.Draw(screen, b.Text, b.FontFace, op)
}

// ============================================================================
// Spinner/Loading Widget - Loading indicator
// ============================================================================

// Spinner is a loading indicator widget
type Spinner struct {
	*BaseWidget
	IsSpinning   bool
	SpinnerColor color.Color
	rotation     float64
}

// NewSpinner creates a new spinner widget
func NewSpinner(id string) *Spinner {
	return &Spinner{
		BaseWidget:   NewBaseWidget(id, "spinner"),
		IsSpinning:   true,
		SpinnerColor: color.RGBA{100, 149, 237, 255},
	}
}

// Update updates the spinner animation
func (s *Spinner) Update() {
	if s.IsSpinning {
		s.rotation += 0.1
	}
}

// Draw renders the spinner
func (s *Spinner) Draw(screen *ebiten.Image) {
	if !s.visible || !s.IsSpinning {
		return
	}

	r := s.computedRect
	size := r.W
	if r.H < size {
		size = r.H
	}

	centerX := r.X + r.W/2
	centerY := r.Y + r.H/2

	// Draw spinning dots
	dotCount := 8
	dotSize := size / 6
	radius := size/2 - dotSize

	for i := 0; i < dotCount; i++ {
		angle := s.rotation + float64(i)*6.283185/float64(dotCount)
		dotX := centerX + radius*cos(angle) - dotSize/2
		dotY := centerY + radius*sin(angle) - dotSize/2

		// Fade based on position
		alpha := uint8(255 * (float64(dotCount-i) / float64(dotCount)))
		r, g, b, _ := s.SpinnerColor.RGBA()
		dotColor := color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), alpha}

		dotRect := Rect{X: dotX, Y: dotY, W: dotSize, H: dotSize}
		DrawRoundedRectPath(screen, dotRect, dotSize/2, dotColor)
	}
}

// Simple cos/sin approximation (avoiding math import for basic rotation)
func cos(angle float64) float64 {
	// Taylor series approximation
	angle = normalizeAngle(angle)
	x2 := angle * angle
	return 1 - x2/2 + x2*x2/24
}

func sin(angle float64) float64 {
	// Taylor series approximation
	angle = normalizeAngle(angle)
	x2 := angle * angle
	return angle - angle*x2/6 + angle*x2*x2/120
}

func normalizeAngle(angle float64) float64 {
	twoPi := 6.283185
	for angle > twoPi {
		angle -= twoPi
	}
	for angle < -twoPi {
		angle += twoPi
	}
	return angle
}

// ============================================================================
// Toast/Notification Widget - Temporary notification message
// ============================================================================

// Toast is a temporary notification message
type Toast struct {
	*BaseWidget
	Message   string
	Duration  float64 // seconds
	FontFace  text.Face
	ToastType string // "info", "success", "warning", "error"
	IsVisible bool

	remainingTime float64
	alpha         float64
}

// NewToast creates a new toast notification
func NewToast(id, message string) *Toast {
	return &Toast{
		BaseWidget: NewBaseWidget(id, "toast"),
		Message:    message,
		Duration:   3.0,
		ToastType:  "info",
		IsVisible:  false,
		alpha:      1.0,
	}
}

// Show displays the toast
func (t *Toast) Show() {
	t.IsVisible = true
	t.remainingTime = t.Duration
	t.alpha = 1.0
}

// Update updates the toast animation
func (t *Toast) Update() {
	if !t.IsVisible {
		return
	}

	t.remainingTime -= 1.0 / 60.0 // Assume 60 FPS

	// Fade out in last 0.3 seconds
	if t.remainingTime < 0.3 {
		t.alpha = t.remainingTime / 0.3
	}

	if t.remainingTime <= 0 {
		t.IsVisible = false
	}
}

// Draw renders the toast
func (t *Toast) Draw(screen *ebiten.Image) {
	if !t.IsVisible || t.Message == "" {
		return
	}

	// Get color based on type
	var bgColor color.Color
	switch t.ToastType {
	case "success":
		bgColor = color.RGBA{76, 175, 80, uint8(220 * t.alpha)}
	case "warning":
		bgColor = color.RGBA{255, 193, 7, uint8(220 * t.alpha)}
	case "error":
		bgColor = color.RGBA{220, 53, 69, uint8(220 * t.alpha)}
	default: // info
		bgColor = color.RGBA{33, 150, 243, uint8(220 * t.alpha)}
	}

	// Position at bottom center of screen
	screenW := float64(screen.Bounds().Dx())
	screenH := float64(screen.Bounds().Dy())

	// Measure text
	textW := 200.0
	textH := 20.0
	if t.FontFace != nil {
		textW, textH = text.Measure(t.Message, t.FontFace, 0)
	}

	paddingX := 24.0
	paddingY := 12.0
	toastW := textW + paddingX*2
	toastH := textH + paddingY*2

	toastX := (screenW - toastW) / 2
	toastY := screenH - toastH - 40

	toastRect := Rect{X: toastX, Y: toastY, W: toastW, H: toastH}

	// Draw background
	DrawRoundedRectPath(screen, toastRect, 8, bgColor)

	// Draw text
	if t.FontFace != nil {
		textColor := color.RGBA{255, 255, 255, uint8(255 * t.alpha)}

		op := &text.DrawOptions{}
		op.GeoM.Translate(toastX+paddingX, toastY+paddingY+textH*0.75)
		op.ColorScale.ScaleWithColor(textColor)
		text.Draw(screen, t.Message, t.FontFace, op)
	}
}
