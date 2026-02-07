package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// BaseWidget provides common functionality for all widgets
type BaseWidget struct {
	id           string
	widgetType   string
	classes      []string
	parent       Widget
	children     []Widget
	style        *Style
	computedRect Rect
	state        WidgetState
	visible      bool
	enabled      bool

	// Event handlers
	onClickHandler func()
	onHoverHandler func()

	// 9-slice image for background
	nineSlice *NineSlice

	// Animation state
	animating bool
	animState *AnimationState

	// CSS transition engine
	transitionEngine *TransitionEngine
}

// NewBaseWidget creates a new base widget
func NewBaseWidget(id, widgetType string) *BaseWidget {
	return &BaseWidget{
		id:               id,
		widgetType:       widgetType,
		classes:          make([]string, 0),
		children:         make([]Widget, 0),
		style:            &Style{Opacity: 1},
		visible:          true,
		enabled:          true,
		state:            StateNormal,
		transitionEngine: NewTransitionEngine(),
	}
}

// ID returns the widget's unique identifier
func (w *BaseWidget) ID() string { return w.id }

// Type returns the widget type
func (w *BaseWidget) Type() string { return w.widgetType }

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

// State returns the widget's state
func (w *BaseWidget) State() WidgetState { return w.state }

// SetState sets the widget's state and starts CSS transitions for changed properties.
func (w *BaseWidget) SetState(s WidgetState) {
	if w.state == s {
		return // no change
	}

	// Snapshot the current active style BEFORE state change
	oldActive := w.getActiveStyle()

	// Change state
	w.state = s

	// Compute new active style AFTER state change
	newActive := w.getActiveStyle()

	// Start transitions using the BASE style's transition declarations
	// (CSS spec: transitions are declared on the base style, not the state style)
	if w.style != nil && len(w.style.parsedTransitions) > 0 {
		w.transitionEngine.StartTransitions(oldActive, newActive, w.style.parsedTransitions)
	}
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

// Draw renders the widget's background and border
func (w *BaseWidget) Draw(screen *ebiten.Image) {
	if !w.visible {
		return
	}

	r := w.computedRect
	style := w.getActiveStyle()

	// Get opacity from style
	opacity := style.Opacity
	if opacity <= 0 {
		opacity = 1 // default
	}

	// Apply animation transforms
	if w.IsAnimating() {
		transX, transY, scaleX, scaleY, _, animOpacity := w.getAnimationTransform()

		// Apply translation
		r.X += transX
		r.Y += transY

		// Apply scale (from center)
		if scaleX != 1 || scaleY != 1 {
			centerX := r.X + r.W/2
			centerY := r.Y + r.H/2
			newW := r.W * scaleX
			newH := r.H * scaleY
			r.X = centerX - newW/2
			r.Y = centerY - newH/2
			r.W = newW
			r.H = newH
		}

		// Apply animation opacity
		opacity *= animOpacity
	}

	// 1. Draw box shadow FIRST (behind everything)
	if style.parsedBoxShadow != nil {
		DrawBoxShadow(screen, r, style.parsedBoxShadow, style.BorderRadius)
	} else if style.BoxShadow != "" {
		// Parse on first use
		style.parsedBoxShadow = ParseBoxShadow(style.BoxShadow)
		if style.parsedBoxShadow != nil {
			DrawBoxShadow(screen, r, style.parsedBoxShadow, style.BorderRadius)
		}
	}

	// Determine rendering target for CSS filter.
	// When a filter is active, widget content (background, border, outline,
	// children) is rendered into an offscreen buffer, then the filter shader
	// composites the result onto screen.  Box shadow is always drawn directly
	// to screen because it extends outside the widget rect.
	useFilter := style.parsedFilter != nil && !filterIsDefault(style.parsedFilter)
	target := screen
	drawRect := r
	var filterOffscreen *ebiten.Image

	if useFilter {
		fw, fh := int(r.W), int(r.H)
		if fw > 0 && fh > 0 {
			filterOffscreen = ebiten.NewImage(fw, fh)
			target = filterOffscreen
			drawRect = Rect{X: 0, Y: 0, W: r.W, H: r.H}
		} else {
			useFilter = false
		}
	}

	// 2. Draw background (9-slice, gradient, or solid)
	if w.nineSlice != nil {
		var colorScale *ebiten.ColorScale
		if opacity < 1 {
			cs := ebiten.ColorScale{}
			cs.SetA(float32(opacity))
			colorScale = &cs
		}
		w.nineSlice.Draw(target, drawRect.X, drawRect.Y, drawRect.W, drawRect.H, colorScale)
	} else if style.parsedGradient != nil {
		// Draw gradient background (linear or radial)
		if style.parsedGradient.Type == GradientRadial {
			DrawRadialGradient(target, drawRect, style.parsedGradient)
		} else {
			DrawGradient(target, drawRect, style.parsedGradient)
		}
	} else if style.BackgroundColor != nil {
		// Draw solid background
		bgColor := style.BackgroundColor
		if opacity < 1 {
			bgColor = applyOpacity(bgColor, opacity)
		}
		DrawRoundedRectPath(target, drawRect, style.BorderRadius, bgColor)
	}

	// 3. Draw border
	if style.BorderColor != nil && style.BorderWidth > 0 {
		borderColor := style.BorderColor
		if opacity < 1 {
			borderColor = applyOpacity(borderColor, opacity)
		}
		drawRoundedRectStroke(target, drawRect, style.BorderRadius, style.BorderWidth, borderColor)
	}

	// 4. Draw outline (outside the border)
	if style.parsedOutline != nil {
		style.parsedOutline.Offset = style.OutlineOffset
		DrawOutline(target, drawRect, style.parsedOutline, style.BorderRadius)
	} else if style.Outline != "" {
		// Parse on first use
		style.parsedOutline = ParseOutline(style.Outline)
		if style.parsedOutline != nil {
			style.parsedOutline.Offset = style.OutlineOffset
			DrawOutline(target, drawRect, style.parsedOutline, style.BorderRadius)
		}
	}

	// 5. Draw children - with overflow clipping
	if style.Overflow == "hidden" || style.Overflow == "scroll" || style.Overflow == "auto" {
		clipW := int(drawRect.W)
		clipH := int(drawRect.H)
		if clipW > 0 && clipH > 0 {
			tmpImg := ebiten.NewImage(clipW, clipH)
			// Translate so parent's origin is at (0,0) in temp image
			sortedChildren := sortByZIndex(w.children)
			for _, child := range sortedChildren {
				// Save and adjust position — always shift by original screen-space r
				origRect := child.ComputedRect()
				shifted := Rect{
					X: origRect.X - r.X,
					Y: origRect.Y - r.Y,
					W: origRect.W,
					H: origRect.H,
				}
				child.SetComputedRect(shifted)
				DrawWidget(tmpImg, child)
				child.SetComputedRect(origRect)
			}
			drawOp := &ebiten.DrawImageOptions{}
			drawOp.GeoM.Translate(drawRect.X, drawRect.Y)
			if opacity < 1 {
				drawOp.ColorScale.SetA(float32(opacity))
			}
			target.DrawImage(tmpImg, drawOp)
			tmpImg.Deallocate()
		}
	} else if useFilter {
		// When filter is active, shift children into offscreen coordinates
		sortedChildren := sortByZIndex(w.children)
		for _, child := range sortedChildren {
			origRect := child.ComputedRect()
			shifted := Rect{
				X: origRect.X - r.X,
				Y: origRect.Y - r.Y,
				W: origRect.W,
				H: origRect.H,
			}
			child.SetComputedRect(shifted)
			DrawWidget(target, child)
			child.SetComputedRect(origRect)
		}
	} else {
		sortedChildren := sortByZIndex(w.children)
		for _, child := range sortedChildren {
			child.Draw(target)
		}
	}

	// 6. Apply CSS filter and composite onto screen
	if useFilter && filterOffscreen != nil {
		ApplyFilter(screen, filterOffscreen, r, style.parsedFilter)
		filterOffscreen.Deallocate()
	}
}

// applyOpacity applies opacity to a color.
// Because Go's color.RGBA uses premultiplied alpha, ALL channels (R,G,B,A)
// must be scaled by opacity to maintain the invariant R,G,B <= A.
func applyOpacity(c color.Color, opacity float64) color.Color {
	r, g, b, a := c.RGBA()
	return color.RGBA{
		R: uint8(float64(r>>8) * opacity),
		G: uint8(float64(g>>8) * opacity),
		B: uint8(float64(b>>8) * opacity),
		A: uint8(float64(a>>8) * opacity),
	}
}

// getActiveStyle returns the style based on current state, with CSS transitions applied.
func (w *BaseWidget) getActiveStyle() *Style {
	var active *Style
	switch w.state {
	case StateHover:
		if w.style.HoverStyle != nil {
			active = mergeStyles(w.style, w.style.HoverStyle)
		}
	case StateActive:
		if w.style.ActiveStyle != nil {
			active = mergeStyles(w.style, w.style.ActiveStyle)
		}
	case StateDisabled:
		if w.style.DisabledStyle != nil {
			active = mergeStyles(w.style, w.style.DisabledStyle)
		}
	case StateFocused:
		if w.style.FocusStyle != nil {
			active = mergeStyles(w.style, w.style.FocusStyle)
		}
	}
	if active == nil {
		active = w.style
	}

	// Apply CSS transitions (interpolate values for in-progress transitions)
	if w.transitionEngine != nil {
		active = w.transitionEngine.Apply(active)
	}

	return active
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
