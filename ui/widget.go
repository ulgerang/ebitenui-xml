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
		for i, child := range w.children {
			cw := child.Style().Width
			if cw <= 0 {
				cw = child.IntrinsicWidth()
			}
			width += cw + child.Style().Margin.Left + child.Style().Margin.Right
			if i < len(w.children)-1 {
				width += gap
			}
		}
	} else {
		// Max of children widths
		for _, child := range w.children {
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
		for i, child := range w.children {
			ch := child.Style().Height
			if ch <= 0 {
				ch = child.IntrinsicHeight()
			}
			height += ch + child.Style().Margin.Top + child.Style().Margin.Bottom
			if i < len(w.children)-1 {
				height += gap
			}
		}
	}

	return height + padding.Top + padding.Bottom + bw*2
}

// State returns the widget's state

func (w *BaseWidget) State() WidgetState { return w.state }

// SetState sets the widget's state
func (w *BaseWidget) SetState(s WidgetState) { w.state = s }

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

	// Determine if we need offscreen compositing
	hasTransform := style.Transform != "" && style.Transform != "none"
	hasFilter := style.parsedFilter != nil
	needsOffscreen := opacity < 1 || hasTransform || hasFilter

	if needsOffscreen {
		w.drawWithCompositing(screen, r, style, opacity, hasTransform, hasFilter)
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
		drawRoundedRectStroke(screen, r, style.BorderRadius, style.BorderWidth, style.BorderColor)
	}

	// 4. Outline
	w.drawOutline(screen, r, style)

	// 5. Children
	w.drawChildren(screen, r, style)
}

// drawWithCompositing renders to an offscreen buffer, applies transform/filter/opacity,
// then composites onto the target screen.
func (w *BaseWidget) drawWithCompositing(screen *ebiten.Image, r Rect, style *Style, opacity float64, hasTransform, hasFilter bool) {
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
	w.drawContentOnly(offscreen, r, style)

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

	if opacity < 1 {
		op.ColorScale.ScaleAlpha(float32(opacity))
	}

	// Crop to the actual screen size because the pooled image may be larger
	// (power-of-2 bucketing).
	cropped := offscreen.SubImage(image.Rect(0, 0, screenW, screenH)).(*ebiten.Image)
	screen.DrawImage(cropped, op)
	globalImagePool.Put(offscreen)
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
		drawRoundedRectStroke(screen, r, style.BorderRadius, style.BorderWidth, style.BorderColor)
	}

	// Outline
	w.drawOutline(screen, r, style)

	// Children
	w.drawChildren(screen, r, style)
}

// drawBoxShadow draws the box shadow effect, parsing on first use.
func (w *BaseWidget) drawBoxShadow(screen *ebiten.Image, r Rect, style *Style) {
	if style.parsedBoxShadow != nil {
		DrawBoxShadow(screen, r, style.parsedBoxShadow, style.BorderRadius)
	} else if style.BoxShadow != "" {
		style.parsedBoxShadow = ParseBoxShadow(style.BoxShadow)
		if style.parsedBoxShadow != nil {
			DrawBoxShadow(screen, r, style.parsedBoxShadow, style.BorderRadius)
		}
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
		DrawRoundedRectPath(screen, r, style.BorderRadius, style.BackgroundColor)
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
		clipW := int(r.W)
		clipH := int(r.H)
		if clipW > 0 && clipH > 0 {
			tmpImg := globalImagePool.Get(clipW, clipH)
			for _, child := range w.children {
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
			drawOp.GeoM.Translate(r.X, r.Y)
			// Crop to the actual clip size because the pooled image may be
			// larger (power-of-2 bucketing); without this, transparent padding
			// from the oversized buffer leaks into the composite.
			cropped := tmpImg.SubImage(image.Rect(0, 0, clipW, clipH)).(*ebiten.Image)
			screen.DrawImage(cropped, drawOp)
			globalImagePool.Put(tmpImg)
		}
	} else {
		for _, child := range w.children {
			child.Draw(screen)
		}
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

	shader := getFilterShader()
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w <= 0 || h <= 0 {
		return src
	}

	dst := globalImagePool.Get(w, h)
	op := &ebiten.DrawRectShaderOptions{}
	op.Images[0] = src
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

	return dst
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
