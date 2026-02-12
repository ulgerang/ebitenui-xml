package ui

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// ============================================================================
// Scrollable Container Widget
// ============================================================================

// Scrollable is a container that can scroll its content
type Scrollable struct {
	*BaseWidget

	// Scroll state
	ScrollX float64
	ScrollY float64

	// Content size (calculated from children)
	ContentWidth  float64
	ContentHeight float64

	// Scroll behavior
	ScrollbarWidth    float64
	ShowHorizontal    bool
	ShowVertical      bool
	AutoHideScrollbar bool
	ScrollSpeed       float64

	// Visual
	ScrollbarColor      color.Color
	ScrollbarTrackColor color.Color
	ScrollbarRadius     float64

	// State
	draggingVertical   bool
	draggingHorizontal bool
	dragStartY         float64
	dragStartX         float64
	dragStartScrollY   float64
	dragStartScrollX   float64
	hoverScrollbarV    bool
	hoverScrollbarH    bool
	scrollbarOpacity   float64

	// Momentum scrolling
	velocityX float64
	velocityY float64
}

// NewScrollable creates a new scrollable container
func NewScrollable(id string) *Scrollable {
	return &Scrollable{
		BaseWidget:          NewBaseWidget(id, "scrollable"),
		ScrollbarWidth:      8,
		ShowVertical:        true,
		AutoHideScrollbar:   true,
		ScrollSpeed:         40,
		ScrollbarColor:      color.RGBA{100, 100, 100, 200},
		ScrollbarTrackColor: color.RGBA{40, 40, 40, 100},
		ScrollbarRadius:     4,
		scrollbarOpacity:    0,
	}
}

// MaxScrollX returns maximum horizontal scroll
func (s *Scrollable) MaxScrollX() float64 {
	r := s.ContentRect()
	return max(0, s.ContentWidth-r.W)
}

// MaxScrollY returns maximum vertical scroll
func (s *Scrollable) MaxScrollY() float64 {
	r := s.ContentRect()
	return max(0, s.ContentHeight-r.H)
}

// ScrollToTop scrolls to the top
func (s *Scrollable) ScrollToTop() {
	s.ScrollY = 0
}

// ScrollToBottom scrolls to the bottom
func (s *Scrollable) ScrollToBottom() {
	s.ScrollY = s.MaxScrollY()
}

// ScrollTo scrolls to a specific position
func (s *Scrollable) ScrollTo(x, y float64) {
	s.ScrollX = clamp(x, 0, s.MaxScrollX())
	s.ScrollY = clamp(y, 0, s.MaxScrollY())
}

// ScrollBy scrolls by a delta
func (s *Scrollable) ScrollBy(dx, dy float64) {
	s.ScrollTo(s.ScrollX+dx, s.ScrollY+dy)
}

// Update handles scroll input
func (s *Scrollable) Update() {
	if !s.visible {
		return
	}

	mx, my := ebiten.CursorPosition()
	mouseX, mouseY := float64(mx), float64(my)
	r := s.computedRect

	// Check if mouse is over this widget
	if r.Contains(mouseX, mouseY) {
		// Handle mouse wheel
		_, wheelY := ebiten.Wheel()
		if wheelY != 0 {
			s.ScrollY -= wheelY * s.ScrollSpeed
			s.ScrollY = clamp(s.ScrollY, 0, s.MaxScrollY())
		}

		// Show scrollbar
		s.scrollbarOpacity = 1
	} else if s.AutoHideScrollbar && !s.draggingVertical && !s.draggingHorizontal {
		// Fade out scrollbar
		s.scrollbarOpacity = max(0, s.scrollbarOpacity-0.05)
	}

	// Handle scrollbar dragging
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		if s.draggingVertical {
			s.handleVerticalDrag(mouseY)
		} else if s.draggingHorizontal {
			s.handleHorizontalDrag(mouseX)
		} else {
			// Check if clicking on scrollbar
			vThumbRect := s.getVerticalThumbRect()
			if vThumbRect.Contains(mouseX, mouseY) {
				s.draggingVertical = true
				s.dragStartY = mouseY
				s.dragStartScrollY = s.ScrollY
			}

			hThumbRect := s.getHorizontalThumbRect()
			if hThumbRect.Contains(mouseX, mouseY) {
				s.draggingHorizontal = true
				s.dragStartX = mouseX
				s.dragStartScrollX = s.ScrollX
			}
		}
	} else {
		s.draggingVertical = false
		s.draggingHorizontal = false
	}

	// Apply momentum
	if math.Abs(s.velocityY) > 0.1 {
		s.ScrollY += s.velocityY
		s.ScrollY = clamp(s.ScrollY, 0, s.MaxScrollY())
		s.velocityY *= 0.95 // Friction
	} else {
		s.velocityY = 0
	}

	if math.Abs(s.velocityX) > 0.1 {
		s.ScrollX += s.velocityX
		s.ScrollX = clamp(s.ScrollX, 0, s.MaxScrollX())
		s.velocityX *= 0.95
	} else {
		s.velocityX = 0
	}

	// Calculate content size from children
	s.calculateContentSize()
}

func (s *Scrollable) calculateContentSize() {
	maxW, maxH := float64(0), float64(0)
	for _, child := range s.children {
		cr := child.ComputedRect()
		right := cr.X - s.computedRect.X + cr.W
		bottom := cr.Y - s.computedRect.Y + cr.H
		if right > maxW {
			maxW = right
		}
		if bottom > maxH {
			maxH = bottom
		}
	}
	s.ContentWidth = maxW
	s.ContentHeight = maxH
}

func (s *Scrollable) handleVerticalDrag(mouseY float64) {
	trackRect := s.getVerticalTrackRect()
	if trackRect.H <= 0 {
		return
	}
	thumbHeight := s.getVerticalThumbHeight()
	maxThumbY := trackRect.H - thumbHeight
	if maxThumbY <= 0 {
		return
	}

	maxScroll := s.MaxScrollY()
	if maxScroll <= 0 {
		s.ScrollY = 0
		return
	}
	deltaY := mouseY - s.dragStartY
	scrollRatio := deltaY / maxThumbY
	s.ScrollY = s.dragStartScrollY + scrollRatio*maxScroll
	s.ScrollY = clamp(s.ScrollY, 0, s.MaxScrollY())
}

func (s *Scrollable) handleHorizontalDrag(mouseX float64) {
	trackRect := s.getHorizontalTrackRect()
	if trackRect.W <= 0 {
		return
	}
	thumbWidth := s.getHorizontalThumbWidth()
	maxThumbX := trackRect.W - thumbWidth
	if maxThumbX <= 0 {
		return
	}

	maxScroll := s.MaxScrollX()
	if maxScroll <= 0 {
		s.ScrollX = 0
		return
	}
	deltaX := mouseX - s.dragStartX
	scrollRatio := deltaX / maxThumbX
	s.ScrollX = s.dragStartScrollX + scrollRatio*maxScroll
	s.ScrollX = clamp(s.ScrollX, 0, s.MaxScrollX())
}

func (s *Scrollable) getVerticalThumbHeight() float64 {
	trackRect := s.getVerticalTrackRect()
	if trackRect.H <= 0 {
		return 0
	}
	if s.ContentHeight <= 0 {
		return 0
	}
	ratio := trackRect.H / s.ContentHeight
	if ratio >= 1 {
		return 0
	}
	return max(20, trackRect.H*ratio)
}

func (s *Scrollable) getHorizontalThumbWidth() float64 {
	trackRect := s.getHorizontalTrackRect()
	if trackRect.W <= 0 {
		return 0
	}
	if s.ContentWidth <= 0 {
		return 0
	}
	ratio := trackRect.W / s.ContentWidth
	if ratio >= 1 {
		return 0
	}
	return max(20, trackRect.W*ratio)
}

func (s *Scrollable) getVerticalThumbRect() Rect {
	trackRect := s.getVerticalTrackRect()
	if trackRect.H <= 0 {
		return Rect{}
	}

	thumbHeight := s.getVerticalThumbHeight()
	maxScroll := s.MaxScrollY()
	scrollRatio := float64(0)
	if maxScroll > 0 {
		scrollRatio = s.ScrollY / maxScroll
	}

	maxThumbY := trackRect.H - thumbHeight
	if maxThumbY < 0 {
		maxThumbY = 0
	}
	thumbY := trackRect.Y + scrollRatio*maxThumbY

	return Rect{
		X: trackRect.X,
		Y: thumbY,
		W: trackRect.W,
		H: thumbHeight,
	}
}

func (s *Scrollable) getHorizontalThumbRect() Rect {
	trackRect := s.getHorizontalTrackRect()
	if trackRect.W <= 0 {
		return Rect{}
	}
	thumbWidth := s.getHorizontalThumbWidth()
	maxScroll := s.MaxScrollX()
	scrollRatio := float64(0)
	if maxScroll > 0 {
		scrollRatio = s.ScrollX / maxScroll
	}

	maxThumbX := trackRect.W - thumbWidth
	if maxThumbX < 0 {
		maxThumbX = 0
	}
	thumbX := trackRect.X + scrollRatio*maxThumbX

	return Rect{
		X: thumbX,
		Y: trackRect.Y,
		W: thumbWidth,
		H: trackRect.H,
	}
}

func (s *Scrollable) getVerticalTrackRect() Rect {
	r := s.ContentRect()
	if !s.ShowVertical || s.ContentHeight <= r.H {
		return Rect{}
	}

	trackHeight := r.H
	if s.ShowHorizontal && s.ContentWidth > r.W {
		trackHeight -= s.ScrollbarWidth
	}
	if trackHeight <= 0 {
		return Rect{}
	}

	return Rect{
		X: r.X + r.W - s.ScrollbarWidth,
		Y: r.Y,
		W: s.ScrollbarWidth,
		H: trackHeight,
	}
}

func (s *Scrollable) getHorizontalTrackRect() Rect {
	r := s.ContentRect()
	if !s.ShowHorizontal || s.ContentWidth <= r.W {
		return Rect{}
	}

	trackWidth := r.W
	if s.ShowVertical && s.ContentHeight > r.H {
		trackWidth -= s.ScrollbarWidth
	}
	if trackWidth <= 0 {
		return Rect{}
	}

	return Rect{
		X: r.X,
		Y: r.Y + r.H - s.ScrollbarWidth,
		W: trackWidth,
		H: s.ScrollbarWidth,
	}
}

// Draw renders the scrollable container and its children
func (s *Scrollable) Draw(screen *ebiten.Image) {
	if !s.visible {
		return
	}

	// Draw only base visuals; children are rendered once below with scroll offset.
	children := s.children
	s.children = nil
	s.BaseWidget.Draw(screen)
	s.children = children

	r := s.ContentRect()

	// Create a clipping region by drawing to a sub-image
	// For now, we'll use simple bounds checking in children

	// Offset children by scroll amount
	for _, child := range s.children {
		cr := child.ComputedRect()

		// Apply scroll offset
		offsetRect := Rect{
			X: cr.X - s.ScrollX,
			Y: cr.Y - s.ScrollY,
			W: cr.W,
			H: cr.H,
		}

		// Skip if completely outside visible area
		if offsetRect.Y+offsetRect.H < r.Y || offsetRect.Y > r.Y+r.H {
			continue
		}
		if offsetRect.X+offsetRect.W < r.X || offsetRect.X > r.X+r.W {
			continue
		}

		// Temporarily adjust child rect for drawing
		child.SetComputedRect(offsetRect)
		child.Draw(screen)
		child.SetComputedRect(cr) // Restore
	}

	// Draw scrollbars
	if s.scrollbarOpacity > 0 {
		s.drawScrollbars(screen)
	}
}

func (s *Scrollable) drawScrollbars(screen *ebiten.Image) {
	// Vertical scrollbar
	if trackRect := s.getVerticalTrackRect(); trackRect.H > 0 {
		// Track
		trackColor := applyOpacity(s.ScrollbarTrackColor, s.scrollbarOpacity)
		DrawRoundedRectPath(screen, trackRect, s.ScrollbarRadius, trackColor)

		// Thumb
		thumbRect := s.getVerticalThumbRect()
		thumbColor := applyOpacity(s.ScrollbarColor, s.scrollbarOpacity)
		if s.draggingVertical || s.hoverScrollbarV {
			// Highlight on hover/drag
			thumbColor = applyOpacity(color.RGBA{150, 150, 150, 255}, s.scrollbarOpacity)
		}
		DrawRoundedRectPath(screen, thumbRect, s.ScrollbarRadius, thumbColor)
	}

	// Horizontal scrollbar
	if trackRect := s.getHorizontalTrackRect(); trackRect.W > 0 {
		// Track
		trackColor := applyOpacity(s.ScrollbarTrackColor, s.scrollbarOpacity)
		DrawRoundedRectPath(screen, trackRect, s.ScrollbarRadius, trackColor)

		// Thumb
		thumbRect := s.getHorizontalThumbRect()
		thumbColor := applyOpacity(s.ScrollbarColor, s.scrollbarOpacity)
		if s.draggingHorizontal || s.hoverScrollbarH {
			thumbColor = applyOpacity(color.RGBA{150, 150, 150, 255}, s.scrollbarOpacity)
		}
		DrawRoundedRectPath(screen, thumbRect, s.ScrollbarRadius, thumbColor)
	}
}

func clamp(v, minVal, maxVal float64) float64 {
	if v < minVal {
		return minVal
	}
	if v > maxVal {
		return maxVal
	}
	return v
}
