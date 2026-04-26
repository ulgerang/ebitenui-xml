package ui

import (
	"image"
	"image/color"
	"math"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// BaseWidget provides common functionality for all widgets
type BaseWidget struct {
	id           string
	widgetType   string
	semanticType string
	classes      []string
	parent       Widget
	children     []Widget
	style        *Style
	computedRect Rect
	state        WidgetState
	visible      bool
	enabled      bool
	tabIndex     int
	tabIndexSet  bool
	focusable    bool
	formInitial  string
	validation   ValidationState
	rules        ValidationRules
	message      string
	scrollX      float64
	scrollY      float64
	contentW     float64
	contentH     float64

	// Event handlers
	onClickHandler func()
	onHoverHandler func()

	// 9-slice image for background
	nineSlice *NineSlice

	// Animation state
	animating            bool
	animState            *AnimationState
	animationDeclaration string

	// CSS transition state
	transitionEngine *TransitionEngine
	compositing      bool
}

// NewBaseWidget creates a new base widget
func NewBaseWidget(id, widgetType string) *BaseWidget {
	return &BaseWidget{
		id:         id,
		widgetType: widgetType,
		classes:    make([]string, 0),
		children:   make([]Widget, 0),
		style:      &Style{Opacity: 1},
		visible:    true,
		enabled:    true,
		state:      StateNormal,
		tabIndex:   0,
	}
}

// ID returns the widget's unique identifier
func (w *BaseWidget) ID() string { return w.id }

// Type returns the widget type
func (w *BaseWidget) Type() string { return w.widgetType }

// SemanticType returns the XML semantic tag used to create this widget, if any.
func (w *BaseWidget) SemanticType() string { return w.semanticType }

// SetSemanticType sets the XML semantic tag metadata for this widget.
func (w *BaseWidget) SetSemanticType(tag string) { w.semanticType = tag }

// Classes returns the widget's CSS classes
func (w *BaseWidget) Classes() []string { return w.classes }

// AddClass adds a CSS class
func (w *BaseWidget) AddClass(class string) {
	if !w.HasClass(class) {
		w.classes = append(w.classes, class)
	}
}

// RemoveClass removes a CSS class
func (w *BaseWidget) RemoveClass(class string) {
	for i, c := range w.classes {
		if c == class {
			w.classes = append(w.classes[:i], w.classes[i+1:]...)
			break
		}
	}
}

// HasClass checks if widget has a class
func (w *BaseWidget) HasClass(class string) bool {
	for _, c := range w.classes {
		if c == class {
			return true
		}
	}
	return false
}

// Parent returns the parent widget
func (w *BaseWidget) Parent() Widget { return w.parent }

// SetParent sets the parent widget
func (w *BaseWidget) SetParent(p Widget) { w.parent = p }

// Children returns all child widgets
func (w *BaseWidget) Children() []Widget { return w.children }

// AddChild adds a child widget
func (w *BaseWidget) AddChild(child Widget) {
	child.SetParent(w)
	w.children = append(w.children, child)
}

// RemoveChild removes a child widget
func (w *BaseWidget) RemoveChild(child Widget) {
	for i, c := range w.children {
		if c == child {
			w.children = append(w.children[:i], w.children[i+1:]...)
			child.SetParent(nil)
			break
		}
	}
}

// ComputedRect returns the calculated rectangle
func (w *BaseWidget) ComputedRect() Rect { return w.computedRect }

// SetComputedRect sets the calculated rectangle
func (w *BaseWidget) SetComputedRect(r Rect) { w.computedRect = r }

// Style returns the widget's style
func (w *BaseWidget) Style() *Style { return w.style }

// SetStyle sets the widget's style
func (w *BaseWidget) SetStyle(s *Style) { w.style = s }

// IntrinsicWidth returns the widget's natural width based on content or children
func (w *BaseWidget) IntrinsicWidth() float64 {
	if len(w.children) == 0 {
		return 0
	}

	style := w.getActiveStyle()
	padding := style.Padding
	bw := style.BorderWidth
	gap := style.Gap

	var width float64
	if style.Direction == LayoutRow {
		// Sum of children widths + gaps
		visibleCount := visibleChildCount(w.children)
		visibleIndex := 0
		for _, child := range w.children {
			if !child.Visible() {
				continue
			}
			cw := child.Style().Width
			if cw <= 0 {
				cw = child.IntrinsicWidth()
			}
			width += cw + child.Style().Margin.Left + child.Style().Margin.Right
			if visibleIndex < visibleCount-1 {
				width += gap
			}
			visibleIndex++
		}
	} else {
		// Max of children widths
		for _, child := range w.children {
			if !child.Visible() {
				continue
			}
			cw := child.Style().Width
			if cw <= 0 {
				cw = child.IntrinsicWidth()
			}
			cw += child.Style().Margin.Left + child.Style().Margin.Right
			if cw > width {
				width = cw
			}
		}
	}

	return width + padding.Left + padding.Right + bw*2
}

// IntrinsicHeight returns the widget's natural height based on content or children
func (w *BaseWidget) IntrinsicHeight() float64 {
	if len(w.children) == 0 {
		return 0
	}

	style := w.getActiveStyle()
	padding := style.Padding
	bw := style.BorderWidth
	gap := style.Gap

	var height float64
	if style.Direction == LayoutRow {
		// Max of children heights
		for _, child := range w.children {
			if !child.Visible() {
				continue
			}
			ch := child.Style().Height
			if ch <= 0 {
				ch = child.IntrinsicHeight()
			}
			ch += child.Style().Margin.Top + child.Style().Margin.Bottom
			if ch > height {
				height = ch
			}
		}
	} else {
		// Sum of children heights + gaps
		visibleCount := visibleChildCount(w.children)
		visibleIndex := 0
		for _, child := range w.children {
			if !child.Visible() {
				continue
			}
			ch := child.Style().Height
			if ch <= 0 {
				ch = child.IntrinsicHeight()
			}
			height += ch + child.Style().Margin.Top + child.Style().Margin.Bottom
			if visibleIndex < visibleCount-1 {
				height += gap
			}
			visibleIndex++
		}
	}

	return height + padding.Top + padding.Bottom + bw*2
}

func visibleChildCount(children []Widget) int {
	count := 0
	for _, child := range children {
		if child.Visible() {
			count++
		}
	}
	return count
}

// State returns the widget's state

func (w *BaseWidget) State() WidgetState { return w.state }

// SetState sets the widget's state
func (w *BaseWidget) SetState(s WidgetState) {
	if w.state == s {
		return
	}

	oldStyle := w.currentTransitionBaseStyle()
	w.state = s
	newStyle := w.getActiveStyle()
	w.startStyleTransitions(oldStyle, newStyle)
}

// Visible returns whether the widget is visible
func (w *BaseWidget) Visible() bool { return w.visible }

// SetVisible sets the widget's visibility
func (w *BaseWidget) SetVisible(v bool) { w.visible = v }

// Enabled returns whether the widget is enabled
func (w *BaseWidget) Enabled() bool { return w.enabled }

// SetEnabled sets the widget's enabled state
func (w *BaseWidget) SetEnabled(e bool) {
	w.enabled = e
	if !e {
		w.state = StateDisabled
	} else if w.state == StateDisabled {
		w.state = StateNormal
	}
}

// SetTabIndex sets the widget's keyboard focus order. Negative values skip tab traversal.
func (w *BaseWidget) SetTabIndex(index int) {
	w.tabIndex = index
	w.tabIndexSet = true
}

// TabIndex returns the widget's keyboard focus order.
func (w *BaseWidget) TabIndex() int { return w.tabIndex }

// SetFocusable marks whether this widget participates in default focus traversal.
func (w *BaseWidget) SetFocusable(focusable bool) { w.focusable = focusable }

// Focusable returns whether this widget can receive keyboard focus.
func (w *BaseWidget) Focusable() bool {
	return w.focusable || (w.tabIndexSet && w.tabIndex >= 0)
}

// SetFormInitialValue stores the reset value used by form semantics.
func (w *BaseWidget) SetFormInitialValue(value string) { w.formInitial = value }

// FormInitialValue returns the reset value used by form semantics.
func (w *BaseWidget) FormInitialValue() string { return w.formInitial }

// SetValidationState updates this widget's form validation state.
func (w *BaseWidget) SetValidationState(state ValidationState) {
	w.validation = state
}

// ValidationState returns this widget's form validation state.
func (w *BaseWidget) ValidationState() ValidationState {
	return w.validation
}

// SetValidationRules stores the validation constraints for this widget.
func (w *BaseWidget) SetValidationRules(rules ValidationRules) {
	w.rules = rules
}

// ValidationRules returns the validation constraints for this widget.
func (w *BaseWidget) ValidationRules() ValidationRules {
	return w.rules
}

// SetValidationMessage stores the latest validation message for this widget.
func (w *BaseWidget) SetValidationMessage(message string) {
	w.message = message
}

// ValidationMessage returns the latest validation message for this widget.
func (w *BaseWidget) ValidationMessage() string {
	return w.message
}

// SetScrollOffset sets the runtime scroll offset for overflow scroll/auto widgets.
func (w *BaseWidget) SetScrollOffset(x, y float64) {
	w.scrollX = clamp(x, 0, w.MaxScrollX())
	w.scrollY = clamp(y, 0, w.MaxScrollY())
}

// ScrollOffset returns the runtime scroll offset for overflow scroll/auto widgets.
func (w *BaseWidget) ScrollOffset() (float64, float64) {
	return w.scrollX, w.scrollY
}

// ScrollBy moves the runtime scroll offset by a delta.
func (w *BaseWidget) ScrollBy(dx, dy float64) {
	w.SetScrollOffset(w.scrollX+dx, w.scrollY+dy)
}

// MaxScrollX returns the maximum horizontal scroll offset.
func (w *BaseWidget) MaxScrollX() float64 {
	return max(0, w.contentW-w.ContentRect().W)
}

// MaxScrollY returns the maximum vertical scroll offset.
func (w *BaseWidget) MaxScrollY() float64 {
	return max(0, w.contentH-w.ContentRect().H)
}

func (w *BaseWidget) setScrollContentSize(width, height float64) {
	w.contentW = max(width, w.ContentRect().W)
	w.contentH = max(height, w.ContentRect().H)
	w.SetScrollOffset(w.scrollX, w.scrollY)
}

// Set9Slice sets a 9-slice image for the widget background
func (w *BaseWidget) Set9Slice(ns *NineSlice) {
	w.nineSlice = ns
}

// OnClick sets the click handler
func (w *BaseWidget) OnClick(handler func()) { w.onClickHandler = handler }

// OnHover sets the hover handler
func (w *BaseWidget) OnHover(handler func()) { w.onHoverHandler = handler }

// HandleClick triggers the click handler
func (w *BaseWidget) HandleClick() {
	if w.enabled && w.onClickHandler != nil {
		w.onClickHandler()
	}
}

// HandleHover triggers the hover handler
func (w *BaseWidget) HandleHover() {
	if w.onHoverHandler != nil {
		w.onHoverHandler()
	}
}

// ============================================================================
// Animation Methods
// ============================================================================

// PlayAnimation starts an animation by name
func (w *BaseWidget) PlayAnimation(name string) {
	anim := GetAnimation(name)
	if anim == nil {
		return
	}
	w.PlayAnimationInstance(anim)
}

// PlayAnimationInstance starts a specific animation instance
func (w *BaseWidget) PlayAnimationInstance(anim *Animation) {
	if anim == nil {
		return
	}
	w.animState = &AnimationState{
		Animation: anim,
	}
	w.animState.Start()
	w.animating = true
}

// StopAnimation stops the current animation
func (w *BaseWidget) StopAnimation() {
	if w.animState != nil {
		w.animState.Stop()
	}
	w.animating = false
}

// IsAnimating returns true if the widget is currently animating
func (w *BaseWidget) IsAnimating() bool {
	return w.animating && w.animState != nil && w.animState.IsPlaying
}

// PauseAnimation pauses the current animation
func (w *BaseWidget) PauseAnimation() {
	if w.animState != nil {
		w.animState.Pause()
	}
}

// ResumeAnimation resumes a paused animation
func (w *BaseWidget) ResumeAnimation() {
	if w.animState != nil {
		w.animState.Resume()
	}
}

// OnAnimationComplete sets a callback for when animation completes
func (w *BaseWidget) OnAnimationComplete(callback func()) {
	if w.animState != nil {
		w.animState.OnComplete = callback
	}
}

// getAnimationTransform returns the current animation transform values
func (w *BaseWidget) getAnimationTransform() (translateX, translateY, scaleX, scaleY, rotate, opacity float64) {
	// Default values
	scaleX, scaleY = 1, 1
	opacity = 1

	if w.animState == nil || !w.IsAnimating() {
		return
	}

	props := w.animState.Update()

	// Check if animation finished
	if !w.animState.IsPlaying {
		w.animating = false
	}

	translateX = props.TranslateX
	translateY = props.TranslateY
	if props.ScaleX != 0 {
		scaleX = props.ScaleX
	}
	if props.ScaleY != 0 {
		scaleY = props.ScaleY
	}
	rotate = props.Rotate
	if props.Opacity != 0 {
		opacity = props.Opacity
	}

	return
}

func (w *BaseWidget) ensureDeclarativeAnimation(style *Style) {
	if style == nil {
		return
	}

	if style.Animation == "" {
		if w.animationDeclaration != "" {
			w.StopAnimation()
			w.animationDeclaration = ""
		}
		return
	}

	if w.animationDeclaration == style.Animation && w.animState != nil {
		return
	}

	anim := style.parsedAnimation
	if anim == nil {
		anim = ParseAnimationDeclaration(style.Animation)
		style.parsedAnimation = anim
	}
	if anim == nil {
		return
	}

	w.animationDeclaration = style.Animation
	w.PlayAnimationInstance(anim)
}

func (w *BaseWidget) currentTransitionBaseStyle() *Style {
	if w.transitionEngine != nil && w.transitionEngine.IsActive() {
		return w.transitionEngine.Apply(w.getActiveStyle())
	}
	return w.getActiveStyle()
}

func (w *BaseWidget) startStyleTransitions(oldStyle, newStyle *Style) {
	if oldStyle == nil || newStyle == nil {
		return
	}

	declarations := oldStyle.parsedTransitions
	if len(declarations) == 0 {
		declarations = newStyle.parsedTransitions
	}
	if len(declarations) == 0 {
		return
	}

	if w.transitionEngine == nil {
		w.transitionEngine = NewTransitionEngine()
	}
	w.transitionEngine.StartTransitions(oldStyle, newStyle, declarations)
}

func (w *BaseWidget) renderStyle(style *Style) *Style {
	if w.transitionEngine != nil && w.transitionEngine.IsActive() {
		return w.transitionEngine.Apply(style)
	}
	return style
}

func (w *BaseWidget) animationTransform() (geoM ebiten.GeoM, hasTransform bool, opacity float64) {
	opacity = 1
	if !w.IsAnimating() {
		return geoM, false, opacity
	}

	transX, transY, scaleX, scaleY, rotate, animOpacity := w.getAnimationTransform()
	opacity = animOpacity
	if transX == 0 && transY == 0 && scaleX == 1 && scaleY == 1 && rotate == 0 {
		return geoM, false, opacity
	}

	r := w.computedRect
	originX := r.X + r.W/2
	originY := r.Y + r.H/2
	geoM.Translate(-originX, -originY)
	if rotate != 0 {
		geoM.Rotate(rotate * math.Pi / 180)
	}
	if scaleX != 1 || scaleY != 1 {
		geoM.Scale(scaleX, scaleY)
	}
	geoM.Translate(originX+transX, originY+transY)
	return geoM, true, opacity
}

// Draw renders the widget's background, border, outline, and children.
//
// When opacity < 1, a CSS transform is set, or a CSS filter is active, the
// entire widget (background + border + outline + children) is rendered to an
// offscreen buffer first, then composited with the appropriate alpha / GeoM /
// shader.  This prevents the double-attenuation artefact that occurs when
// per-element opacity is applied to overlapping regions (e.g. background
// bleeding through a semi-transparent border).
func (w *BaseWidget) Draw(screen *ebiten.Image) {
	if !w.visible {
		return
	}

	r := w.computedRect
	style := w.renderStyle(w.getActiveStyle())
	w.ensureDeclarativeAnimation(style)

	// Get opacity from style
	opacity := style.Opacity
	if opacity <= 0 {
		opacity = 1 // default
	}

	animGeoM, hasAnimTransform, animOpacity := w.animationTransform()
	opacity *= animOpacity

	// Determine if we need offscreen compositing
	hasTransform := style.Transform != "" && style.Transform != "none"
	hasFilter := style.parsedFilter != nil
	hasClipPath := style.ClipPath != "" && style.ClipPath != "none"
	needsOffscreen := opacity < 1 || hasTransform || hasFilter || hasAnimTransform || hasClipPath

	if needsOffscreen {
		w.drawWithCompositing(screen, r, style, opacity, hasTransform, hasFilter, hasClipPath, animGeoM, hasAnimTransform)
	} else {
		w.drawDirect(screen, r, style)
	}
}

// drawDirect renders the widget directly to screen (no offscreen compositing needed).
func (w *BaseWidget) drawDirect(screen *ebiten.Image, r Rect, style *Style) {
	// 1. Box shadow (behind everything)
	w.drawBoxShadow(screen, r, style)

	// 2. Backdrop filter (behind background, glassmorphism)
	w.drawBackdropFilter(screen, r, style)

	// 3. Background (9-slice, gradient, or solid)
	w.drawBackground(screen, r, style)

	// 3. Border
	if style.BorderColor != nil && style.BorderWidth > 0 {
		tl, tr, br, bl := w.getCornerRadii(style)
		drawRoundedRectStrokeEx(screen, r, tl, tr, br, bl, style.BorderWidth, style.BorderColor)
	}
	w.drawIndividualBorders(screen, r, style)

	// 4. Outline
	w.drawOutline(screen, r, style)

	// 5. Children
	w.drawChildren(screen, r, style)
}

// drawWithCompositing renders to an offscreen buffer, applies transform/filter/opacity,
// then composites onto the target screen.
func (w *BaseWidget) drawWithCompositing(screen *ebiten.Image, r Rect, style *Style, opacity float64, hasTransform, hasFilter, hasClipPath bool, animGeoM ebiten.GeoM, hasAnimTransform bool) {
	w.drawCustomWithCompositing(screen, r, style, opacity, hasTransform, hasFilter, hasClipPath, animGeoM, hasAnimTransform, w.drawContentOnly)
}

func (w *BaseWidget) drawCustomWithCompositing(screen *ebiten.Image, r Rect, style *Style, opacity float64, hasTransform, hasFilter, hasClipPath bool, animGeoM ebiten.GeoM, hasAnimTransform bool, drawContent func(*ebiten.Image, Rect, *Style)) {
	bounds := screen.Bounds()
	screenW, screenH := bounds.Dx(), bounds.Dy()
	if screenW <= 0 || screenH <= 0 {
		return
	}

	// Box shadow is drawn directly to screen — it is beneath the widget and
	// should not be affected by the widget's own transform or opacity.
	w.drawBoxShadow(screen, r, style)

	offscreen := globalImagePool.Get(screenW, screenH)

	// Draw widget content to offscreen at original coordinates
	drawContent(offscreen, r, style)

	if hasClipPath {
		applyCSSClipPath(offscreen, r, style.ClipPath)
	}

	// Apply CSS filter if present
	if hasFilter {
		filtered := applyCSSFilter(offscreen, style.parsedFilter)
		if filtered != offscreen {
			globalImagePool.Put(offscreen)
			offscreen = filtered
		}
	}

	// Composite to screen with transform and opacity
	op := &ebiten.DrawImageOptions{}

	if hasTransform {
		originX, originY := parseCSSTransformOrigin(style.TransformOrigin, r.W, r.H)
		originX += r.X
		originY += r.Y
		geoM := parseCSSTransform(style.Transform, originX, originY)
		op.GeoM = geoM
	}
	if hasAnimTransform {
		op.GeoM.Concat(animGeoM)
	}

	if opacity < 1 {
		op.ColorScale.ScaleAlpha(float32(opacity))
	}

	// Crop to the actual screen size because the pooled image may be larger
	// (power-of-2 bucketing).
	cropped := offscreen.SubImage(image.Rect(0, 0, screenW, screenH)).(*ebiten.Image)
	screen.DrawImage(cropped, op)
	globalImagePool.Put(offscreen)
}

func (w *BaseWidget) drawFullWidgetWithEffects(screen *ebiten.Image, draw func(*ebiten.Image)) bool {
	if w.compositing || !w.visible {
		return false
	}
	r := w.computedRect
	style := w.renderStyle(w.getActiveStyle())
	w.ensureDeclarativeAnimation(style)
	opacity := style.Opacity
	if opacity <= 0 {
		opacity = 1
	}
	animGeoM, hasAnimTransform, animOpacity := w.animationTransform()
	opacity *= animOpacity
	hasTransform := style.Transform != "" && style.Transform != "none"
	hasFilter := style.parsedFilter != nil
	hasClipPath := style.ClipPath != "" && style.ClipPath != "none"
	if opacity >= 1 && !hasTransform && !hasFilter && !hasAnimTransform && !hasClipPath {
		return false
	}

	bounds := screen.Bounds()
	screenW, screenH := bounds.Dx(), bounds.Dy()
	if screenW <= 0 || screenH <= 0 {
		return true
	}
	offscreen := globalImagePool.Get(screenW, screenH)
	w.compositing = true
	draw(offscreen)
	w.compositing = false

	if hasClipPath {
		applyCSSClipPath(offscreen, r, style.ClipPath)
	}
	if hasFilter {
		filtered := applyCSSFilter(offscreen, style.parsedFilter)
		if filtered != offscreen {
			globalImagePool.Put(offscreen)
			offscreen = filtered
		}
	}

	op := &ebiten.DrawImageOptions{}
	if hasTransform {
		originX, originY := parseCSSTransformOrigin(style.TransformOrigin, r.W, r.H)
		originX += r.X
		originY += r.Y
		op.GeoM = parseCSSTransform(style.Transform, originX, originY)
	}
	if hasAnimTransform {
		op.GeoM.Concat(animGeoM)
	}
	if opacity < 1 {
		op.ColorScale.ScaleAlpha(float32(opacity))
	}
	cropped := offscreen.SubImage(image.Rect(0, 0, screenW, screenH)).(*ebiten.Image)
	screen.DrawImage(cropped, op)
	globalImagePool.Put(offscreen)
	return true
}

// drawContentOnly renders widget content (bg, border, outline, children) without
// shadow or compositing.
func (w *BaseWidget) drawContentOnly(screen *ebiten.Image, r Rect, style *Style) {
	// Backdrop filter (behind background)
	w.drawBackdropFilter(screen, r, style)

	// Background
	w.drawBackground(screen, r, style)

	// Border
	if style.BorderColor != nil && style.BorderWidth > 0 {
		tl, tr, br, bl := w.getCornerRadii(style)
		drawRoundedRectStrokeEx(screen, r, tl, tr, br, bl, style.BorderWidth, style.BorderColor)
	}
	w.drawIndividualBorders(screen, r, style)

	// Outline
	w.drawOutline(screen, r, style)

	// Children
	w.drawChildren(screen, r, style)
}

// drawBoxShadow draws the box shadow effect, parsing on first use.
func (w *BaseWidget) drawBoxShadow(screen *ebiten.Image, r Rect, style *Style) {
	shadows := style.parsedBoxShadows
	if len(shadows) == 0 && style.BoxShadow != "" {
		shadows = ParseBoxShadowList(style.BoxShadow)
		style.parsedBoxShadows = shadows
		if len(shadows) > 0 {
			style.parsedBoxShadow = shadows[0]
		}
	}
	if len(shadows) == 0 && style.parsedBoxShadow != nil {
		shadows = []*BoxShadow{style.parsedBoxShadow}
	}
	radTL, radTR, radBR, radBL := w.getCornerRadii(style)
	for _, shadow := range shadows {
		DrawBoxShadowEx(screen, r, shadow, radTL, radTR, radBR, radBL)
	}
}

// drawBackdropFilter applies the CSS backdrop-filter effect (glassmorphism).
// Captures the screen region behind this widget and applies blur/brightness/saturate.
// Must be called before drawBackground so the captured content is uncontaminated.
func (w *BaseWidget) drawBackdropFilter(screen *ebiten.Image, r Rect, style *Style) {
	if style.parsedBackdropFilter != nil {
		ApplyBackdropFilter(screen, r, style.parsedBackdropFilter)
	} else if style.BackdropFilter != "" {
		style.parsedBackdropFilter = ParseBackdropFilter(style.BackdropFilter)
		if style.parsedBackdropFilter != nil {
			ApplyBackdropFilter(screen, r, style.parsedBackdropFilter)
		}
	}
}

// drawBackground draws the widget background (9-slice, gradient, or solid colour).
func (w *BaseWidget) drawBackground(screen *ebiten.Image, r Rect, style *Style) {
	if w.nineSlice != nil {
		w.nineSlice.Draw(screen, r.X, r.Y, r.W, r.H, nil)
	} else if style.parsedGradient != nil {
		radTL, radTR, radBR, radBL := w.getCornerRadii(style)
		if radTL > 0 || radTR > 0 || radBR > 0 || radBL > 0 {
			drawGradientWithRadius(screen, r, style.parsedGradient, radTL, radTR, radBR, radBL)
		} else if style.parsedGradient.Type == GradientRadial {
			DrawRadialGradient(screen, r, style.parsedGradient)
		} else {
			DrawGradient(screen, r, style.parsedGradient)
		}
	} else if style.BackgroundColor != nil {
		radTL, radTR, radBR, radBL := w.getCornerRadii(style)
		DrawRoundedRectPathEx(screen, r, radTL, radTR, radBR, radBL, style.BackgroundColor)
	}
}

// drawGradientWithRadius draws a gradient clipped to rounded corners.
// Uses clipComposite to render the gradient masked by the rounded rect shape.
func drawGradientWithRadius(screen *ebiten.Image, r Rect, g *Gradient, radTL, radTR, radBR, radBL float64) {
	iw := int(math.Ceil(r.W))
	ih := int(math.Ceil(r.H))

	localRect := Rect{X: 0, Y: 0, W: r.W, H: r.H}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(r.X, r.Y)

	clipComposite(screen, iw, ih,
		func(content *ebiten.Image) {
			if g.Type == GradientRadial {
				DrawRadialGradient(content, localRect, g)
			} else {
				DrawGradient(content, localRect, g)
			}
		},
		func(mask *ebiten.Image) {
			DrawRoundedRectPathEx(mask, localRect, radTL, radTR, radBR, radBL, color.White)
		},
		op,
	)
}

// drawOutline draws the CSS outline, parsing on first use.
func (w *BaseWidget) drawOutline(screen *ebiten.Image, r Rect, style *Style) {
	if style.parsedOutline != nil {
		style.parsedOutline.Offset = style.OutlineOffset
		DrawOutline(screen, r, style.parsedOutline, style.BorderRadius)
	} else if style.Outline != "" {
		style.parsedOutline = ParseOutline(style.Outline)
		if style.parsedOutline != nil {
			style.parsedOutline.Offset = style.OutlineOffset
			DrawOutline(screen, r, style.parsedOutline, style.BorderRadius)
		}
	}
}

// drawChildren renders child widgets with overflow handling.
func (w *BaseWidget) drawChildren(screen *ebiten.Image, r Rect, style *Style) {
	if style.Overflow == "hidden" || style.Overflow == "scroll" || style.Overflow == "auto" {
		content := w.ContentRect()
		clipW := int(content.W)
		clipH := int(content.H)
		if clipW > 0 && clipH > 0 {
			tmpImg := globalImagePool.Get(clipW, clipH)
			scrollX, scrollY := 0.0, 0.0
			if style.Overflow == "scroll" || style.Overflow == "auto" {
				scrollX, scrollY = w.ScrollOffset()
			}
			for _, child := range sortedChildrenByZ(w.children, false) {
				translateWidgetTree(child, -content.X-scrollX, -content.Y-scrollY)
				DrawWidget(tmpImg, child)
				translateWidgetTree(child, content.X+scrollX, content.Y+scrollY)
			}
			drawOp := &ebiten.DrawImageOptions{}
			drawOp.GeoM.Translate(content.X, content.Y)
			// Crop to the actual clip size because the pooled image may be
			// larger (power-of-2 bucketing); without this, transparent padding
			// from the oversized buffer leaks into the composite.
			cropped := tmpImg.SubImage(image.Rect(0, 0, clipW, clipH)).(*ebiten.Image)
			screen.DrawImage(cropped, drawOp)
			globalImagePool.Put(tmpImg)
		}
	} else {
		for _, child := range sortedChildrenByZ(w.children, false) {
			child.Draw(screen)
		}
	}
}

func translateWidgetTree(widget Widget, dx, dy float64) {
	if widget == nil {
		return
	}
	r := widget.ComputedRect()
	r.X += dx
	r.Y += dy
	widget.SetComputedRect(r)
	for _, child := range widget.Children() {
		translateWidgetTree(child, dx, dy)
	}
}

// getCornerRadii returns per-corner radii, falling back to uniform BorderRadius.
func (w *BaseWidget) getCornerRadii(style *Style) (tl, tr, br, bl float64) {
	tl = style.BorderTopLeftRadius
	tr = style.BorderTopRightRadius
	br = style.BorderBottomRightRadius
	bl = style.BorderBottomLeftRadius

	if tl == 0 && tr == 0 && br == 0 && bl == 0 {
		tl = style.BorderRadius
		tr = style.BorderRadius
		br = style.BorderRadius
		bl = style.BorderRadius
	}
	return
}

// ============================================================================
// CSS Filter Application
// ============================================================================

// applyCSSFilter applies CSS colour-matrix filter to an offscreen image.
// Returns the filtered image; the caller must Put() the returned image back
// to globalImagePool when it differs from src.  The original src is NOT
// released here.
func applyCSSFilter(src *ebiten.Image, f *Filter) *ebiten.Image {
	if f == nil {
		return src
	}

	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w <= 0 || h <= 0 {
		return src
	}

	current := src
	if f.Blur > 0 {
		current = applyGaussianBlur(current, f.Blur)
	}

	if f.Brightness == 1 && f.Contrast == 1 && f.Saturate == 1 &&
		f.Grayscale == 0 && f.Sepia == 0 && f.HueRotate == 0 && f.Invert == 0 {
		return current
	}

	shader := getFilterShader()
	dst := globalImagePool.Get(w, h)
	op := &ebiten.DrawRectShaderOptions{}
	op.Images[0] = current
	op.Uniforms = map[string]any{
		"Brightness": float32(f.Brightness),
		"Contrast":   float32(f.Contrast),
		"Saturate":   float32(f.Saturate),
		"Grayscale":  float32(f.Grayscale),
		"Sepia":      float32(f.Sepia),
		"HueRotate":  float32(f.HueRotate * math.Pi / 180),
		"Invert":     float32(f.Invert),
	}
	dst.DrawRectShader(w, h, shader, op)

	if current != src {
		globalImagePool.Put(current)
	}
	return dst
}

func applyGaussianBlur(src *ebiten.Image, sigma float64) *ebiten.Image {
	if sigma <= 0 {
		return src
	}

	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w <= 0 || h <= 0 {
		return src
	}

	shader := getBackdropBlurShader()
	tmp := globalImagePool.Get(w, h)
	out := globalImagePool.Get(w, h)

	horizontal := &ebiten.DrawRectShaderOptions{}
	horizontal.Images[0] = src
	horizontal.Uniforms = map[string]interface{}{
		"Sigma":     float32(sigma),
		"Direction": [2]float32{1, 0},
	}
	tmp.DrawRectShader(w, h, shader, horizontal)

	vertical := &ebiten.DrawRectShaderOptions{}
	vertical.Images[0] = tmp.SubImage(image.Rect(0, 0, w, h)).(*ebiten.Image)
	vertical.Uniforms = map[string]interface{}{
		"Sigma":     float32(sigma),
		"Direction": [2]float32{0, 1},
	}
	out.DrawRectShader(w, h, shader, vertical)

	globalImagePool.Put(tmp)
	return out
}

// ============================================================================
// CSS Clip-Path Support
// ============================================================================

// drawWithClipPath renders a widget with clip-path clipping.
// Uses clipComposite for offscreen compositing with destination-in blending.
func drawWithClipPath(screen *ebiten.Image, drawFunc func(*ebiten.Image), clipPath *vector.Path) {
	if clipPath == nil {
		drawFunc(screen)
		return
	}

	bounds := screen.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	clipComposite(screen, w, h,
		drawFunc,
		func(mask *ebiten.Image) {
			vs, is := clipPath.AppendVerticesAndIndicesForFilling(nil, nil)
			applyColorToVertices(vs, color.White)
			mask.DrawTriangles(vs, is, whiteImage, &ebiten.DrawTrianglesOptions{
				AntiAlias: true,
				FillRule:  ebiten.FillRuleNonZero,
			})
		},
		nil,
	)
}

func applyCSSClipPath(target *ebiten.Image, r Rect, clipPath string) {
	path := parseCSSClipPath(clipPath, r)
	if path == nil {
		return
	}
	bounds := target.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w <= 0 || h <= 0 {
		return
	}
	mask := globalImagePool.Get(w, h)
	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
	applyColorToVertices(vs, color.White)
	mask.DrawTriangles(vs, is, whiteImage, &ebiten.DrawTrianglesOptions{
		AntiAlias: true,
		FillRule:  ebiten.FillRuleNonZero,
	})
	cropped := mask.SubImage(image.Rect(0, 0, w, h)).(*ebiten.Image)
	target.DrawImage(cropped, &ebiten.DrawImageOptions{Blend: ebiten.BlendDestinationIn})
	globalImagePool.Put(mask)
}

func parseCSSClipPath(value string, r Rect) *vector.Path {
	value = strings.TrimSpace(value)
	if value == "" || value == "none" {
		return nil
	}
	switch {
	case strings.HasPrefix(value, "inset(") && strings.HasSuffix(value, ")"):
		return parseInsetClipPath(strings.TrimSuffix(strings.TrimPrefix(value, "inset("), ")"), r)
	case strings.HasPrefix(value, "circle(") && strings.HasSuffix(value, ")"):
		return parseCircleClipPath(strings.TrimSuffix(strings.TrimPrefix(value, "circle("), ")"), r)
	case strings.HasPrefix(value, "polygon(") && strings.HasSuffix(value, ")"):
		return parsePolygonClipPath(strings.TrimSuffix(strings.TrimPrefix(value, "polygon("), ")"), r)
	case strings.HasPrefix(value, "path(") && strings.HasSuffix(value, ")"):
		return parsePathClipPath(strings.TrimSuffix(strings.TrimPrefix(value, "path("), ")"), r)
	default:
		return nil
	}
}

func parseInsetClipPath(args string, r Rect) *vector.Path {
	parts := strings.Fields(args)
	if len(parts) == 0 {
		return nil
	}
	values := make([]float64, len(parts))
	for i, part := range parts {
		values[i] = parseCSSClipLength(part, r.W)
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
	clip := Rect{X: r.X + left, Y: r.Y + top, W: r.W - left - right, H: r.H - top - bottom}
	if clip.W <= 0 || clip.H <= 0 {
		return nil
	}
	path := &vector.Path{}
	path.MoveTo(float32(clip.X), float32(clip.Y))
	path.LineTo(float32(clip.X+clip.W), float32(clip.Y))
	path.LineTo(float32(clip.X+clip.W), float32(clip.Y+clip.H))
	path.LineTo(float32(clip.X), float32(clip.Y+clip.H))
	path.Close()
	return path
}

func parseCircleClipPath(args string, r Rect) *vector.Path {
	args = strings.TrimSpace(args)
	radiusPart := args
	centerX, centerY := r.X+r.W/2, r.Y+r.H/2
	if before, after, ok := strings.Cut(args, " at "); ok {
		radiusPart = strings.TrimSpace(before)
		centerParts := strings.Fields(after)
		if len(centerParts) >= 2 {
			centerX = r.X + parseCSSClipLength(centerParts[0], r.W)
			centerY = r.Y + parseCSSClipLength(centerParts[1], r.H)
		}
	}
	radius := min(r.W, r.H) / 2
	if radiusPart != "" && radiusPart != "closest-side" {
		radius = parseCSSClipLength(radiusPart, min(r.W, r.H))
	}
	if radius <= 0 {
		return nil
	}
	path := &vector.Path{}
	path.Arc(float32(centerX), float32(centerY), float32(radius), 0, 2*math.Pi, vector.Clockwise)
	path.Close()
	return path
}

func parsePolygonClipPath(args string, r Rect) *vector.Path {
	points := strings.Split(args, ",")
	if len(points) < 3 {
		return nil
	}
	path := &vector.Path{}
	for i, point := range points {
		coords := strings.Fields(strings.TrimSpace(point))
		if len(coords) < 2 {
			return nil
		}
		x := r.X + parseCSSClipLength(coords[0], r.W)
		y := r.Y + parseCSSClipLength(coords[1], r.H)
		if i == 0 {
			path.MoveTo(float32(x), float32(y))
		} else {
			path.LineTo(float32(x), float32(y))
		}
	}
	path.Close()
	return path
}

func parsePathClipPath(args string, r Rect) *vector.Path {
	pathData := strings.TrimSpace(args)
	if len(pathData) >= 2 {
		quote := pathData[0]
		if (quote == '\'' || quote == '"') && pathData[len(pathData)-1] == quote {
			pathData = pathData[1 : len(pathData)-1]
		}
	}
	pathData = strings.TrimSpace(pathData)
	if pathData == "" {
		return nil
	}
	return ParsePathDataScaled(pathData, r.X, r.Y, 1, 1)
}

func parseCSSClipLength(value string, reference float64) float64 {
	value = strings.TrimSpace(value)
	if strings.HasSuffix(value, "%") {
		pct, _ := strconv.ParseFloat(strings.TrimSuffix(value, "%"), 64)
		return reference * pct / 100
	}
	return parseCSSPixels(value)
}

// applyOpacity applies opacity to a color, returning a non-premultiplied
// (straight-alpha) color.NRGBA.  The previous implementation returned
// color.RGBA (premultiplied type) but only scaled A — producing invalid
// premultiplied values (R>A) that caused over-bright rendering in
// vector.DrawFilledRect / ScaleWithColor.
func applyOpacity(c color.Color, opacity float64) color.Color {
	r, g, b, a := c.RGBA() // premultiplied, [0..0xffff]
	if a == 0 {
		return color.NRGBA{}
	}
	// Un-premultiply to recover straight RGB, then scale alpha by opacity.
	return color.NRGBA{
		R: uint8(float64(r) * 0xffff / float64(a)),
		G: uint8(float64(g) * 0xffff / float64(a)),
		B: uint8(float64(b) * 0xffff / float64(a)),
		A: uint8(float64(a>>8) * opacity),
	}
}

func snapToPixel(v float64) float64 {
	return math.Round(v)
}

// getActiveStyle returns the style based on current state
func (w *BaseWidget) getActiveStyle() *Style {
	switch w.state {
	case StateHover:
		if w.style.HoverStyle != nil {
			return mergeStyles(w.style, w.style.HoverStyle)
		}
	case StateActive:
		if w.style.ActiveStyle != nil {
			return mergeStyles(w.style, w.style.ActiveStyle)
		}
	case StateDisabled:
		if w.style.DisabledStyle != nil {
			return mergeStyles(w.style, w.style.DisabledStyle)
		}
	case StateFocused:
		if w.style.FocusStyle != nil {
			return mergeStyles(w.style, w.style.FocusStyle)
		}
	}
	return w.style
}

// mergeStyles merges base style with override style
func mergeStyles(base, override *Style) *Style {
	merged := base.Clone()
	merged.Merge(override)
	return merged
}

// drawRoundedRect draws a filled rounded rectangle
func drawRoundedRect(screen *ebiten.Image, r Rect, radius float64, clr color.Color) {
	DrawRoundedRectPath(screen, r, radius, clr)
}

// Note: drawRoundedRectStroke is defined in effects.go

// ContentRect returns the content area (rect minus padding and border)
func (w *BaseWidget) ContentRect() Rect {
	bw := w.style.BorderWidth
	return w.computedRect.Inset(
		w.style.Padding.Top+bw,
		w.style.Padding.Right+bw,
		w.style.Padding.Bottom+bw,
		w.style.Padding.Left+bw,
	)
}

// ============================================================================
// CSS Transform Parsing
// ============================================================================

// parseCSSTransform parses a CSS transform string and returns an ebiten.GeoM.
// Supported functions: rotate(), scale(), translate(), skewX(), skewY().
// The transform is applied around the given origin point (originX, originY)
// in screen-space coordinates.
func parseCSSTransform(transform string, originX, originY float64) ebiten.GeoM {
	var geoM ebiten.GeoM

	transform = strings.TrimSpace(transform)
	if transform == "" || transform == "none" {
		return geoM
	}

	// Translate origin to (0,0) so transforms are applied relative to the origin.
	geoM.Translate(-originX, -originY)

	// Parse and apply each transform function in order.
	remaining := transform
	for remaining != "" {
		remaining = strings.TrimSpace(remaining)
		if remaining == "" {
			break
		}

		// Find function name and opening paren
		parenIdx := strings.Index(remaining, "(")
		if parenIdx < 0 {
			break
		}
		funcName := strings.TrimSpace(remaining[:parenIdx])

		// Find closing paren
		closeIdx := strings.Index(remaining, ")")
		if closeIdx < 0 {
			break
		}
		args := remaining[parenIdx+1 : closeIdx]
		remaining = remaining[closeIdx+1:]

		switch strings.ToLower(funcName) {
		case "rotate":
			angle := parseCSSAngle(args)
			geoM.Rotate(angle)
		case "scale":
			parts := strings.Split(args, ",")
			sx := parseCSSScaleValue(strings.TrimSpace(parts[0]))
			sy := sx
			if len(parts) > 1 {
				sy = parseCSSScaleValue(strings.TrimSpace(parts[1]))
			}
			geoM.Scale(sx, sy)
		case "translate":
			parts := strings.Split(args, ",")
			tx := parseCSSPixels(strings.TrimSpace(parts[0]))
			ty := 0.0
			if len(parts) > 1 {
				ty = parseCSSPixels(strings.TrimSpace(parts[1]))
			}
			geoM.Translate(tx, ty)
		case "skewx":
			angle := parseCSSAngle(args)
			var skew ebiten.GeoM
			skew.SetElement(0, 1, math.Tan(angle))
			geoM.Concat(skew)
		case "skewy":
			angle := parseCSSAngle(args)
			var skew ebiten.GeoM
			skew.SetElement(1, 0, math.Tan(angle))
			geoM.Concat(skew)
		}
	}

	// Translate origin back.
	geoM.Translate(originX, originY)

	return geoM
}

// parseCSSAngle parses "45deg" or "0.785rad" to radians.
func parseCSSAngle(s string) float64 {
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, "rad") {
		s = strings.TrimSuffix(s, "rad")
		f, _ := strconv.ParseFloat(s, 64)
		return f
	}
	s = strings.TrimSuffix(s, "deg")
	f, _ := strconv.ParseFloat(s, 64)
	return f * math.Pi / 180
}

// parseCSSScaleValue parses a scale factor.  Returns 1.0 for empty strings.
func parseCSSScaleValue(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 1
	}
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// parseCSSPixels parses "10px" or "10" to float64.
func parseCSSPixels(s string) float64 {
	s = strings.TrimSuffix(strings.TrimSpace(s), "px")
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// parseCSSTransformOrigin parses CSS transform-origin values like "center",
// "top left", "50% 50%", or "10px 20px" into screen-space coordinates within
// the given widget dimensions (w × h).
func parseCSSTransformOrigin(origin string, w, h float64) (float64, float64) {
	if origin == "" || origin == "center" {
		return w / 2, h / 2
	}

	parts := strings.Fields(origin)
	ox, oy := w/2, h/2

	if len(parts) >= 1 {
		switch parts[0] {
		case "left":
			ox = 0
		case "center":
			ox = w / 2
		case "right":
			ox = w
		default:
			if strings.HasSuffix(parts[0], "%") {
				pct, _ := strconv.ParseFloat(strings.TrimSuffix(parts[0], "%"), 64)
				ox = w * pct / 100
			} else {
				ox = parseCSSPixels(parts[0])
			}
		}
	}
	if len(parts) >= 2 {
		switch parts[1] {
		case "top":
			oy = 0
		case "center":
			oy = h / 2
		case "bottom":
			oy = h
		default:
			if strings.HasSuffix(parts[1], "%") {
				pct, _ := strconv.ParseFloat(strings.TrimSuffix(parts[1], "%"), 64)
				oy = h * pct / 100
			} else {
				oy = parseCSSPixels(parts[1])
			}
		}
	}

	return ox, oy
}

// drawIndividualBorders draws borders for each side separately.
func (w *BaseWidget) drawIndividualBorders(screen *ebiten.Image, r Rect, style *Style) {
	// Top
	if style.BorderTopWidth > 0 {
		c := style.BorderTopColor
		if c == nil {
			c = style.BorderColor
		}
		if c != nil {
			vector.DrawFilledRect(screen, float32(r.X), float32(r.Y), float32(r.W), float32(style.BorderTopWidth), c, false)
		}
	}
	// Bottom
	if style.BorderBottomWidth > 0 {
		c := style.BorderBottomColor
		if c == nil {
			c = style.BorderColor
		}
		if c != nil {
			vector.DrawFilledRect(screen, float32(r.X), float32(r.Y+r.H-style.BorderBottomWidth), float32(r.W), float32(style.BorderBottomWidth), c, false)
		}
	}
	// Left
	if style.BorderLeftWidth > 0 {
		c := style.BorderLeftColor
		if c == nil {
			c = style.BorderColor
		}
		if c != nil {
			vector.DrawFilledRect(screen, float32(r.X), float32(r.Y), float32(style.BorderLeftWidth), float32(r.H), c, false)
		}
	}
	// Right
	if style.BorderRightWidth > 0 {
		c := style.BorderRightColor
		if c == nil {
			c = style.BorderColor
		}
		if c != nil {
			vector.DrawFilledRect(screen, float32(r.X+r.W-style.BorderRightWidth), float32(r.Y), float32(style.BorderRightWidth), float32(r.H), c, false)
		}
	}
}
