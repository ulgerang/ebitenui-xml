package ui

import (
	"math"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
)

// Default intrinsic dimensions for widgets without explicit size
const (
	defaultIntrinsicWidth  = 50.0
	defaultIntrinsicHeight = 30.0
)

// LayoutEngine handles the layout calculation for widgets
type LayoutEngine struct{}

// NewLayoutEngine creates a new layout engine
func NewLayoutEngine() *LayoutEngine {
	return &LayoutEngine{}
}

// LayoutWidget is a convenience function to layout a widget tree
// It uses a shared LayoutEngine instance
func LayoutWidget(root Widget) {
	if root == nil {
		return
	}
	le := NewLayoutEngine()
	rect := root.ComputedRect()
	le.Layout(root, rect.W, rect.H)
}

// DrawWidget draws a widget (and its children via widget.Draw's internal
// drawChildren call).  The previous recursive child loop has been removed
// because widget.Draw() already calls drawChildren() internally — the extra
// loop caused every child to be drawn twice, resulting in double-blend
// artefacts (visible in overflow:hidden tests).
func DrawWidget(screen *ebiten.Image, widget Widget) {
	if widget == nil || !widget.Visible() {
		return
	}
	widget.Draw(screen)
}

// Layout calculates positions and sizes for a widget tree
func (le *LayoutEngine) Layout(root Widget, containerWidth, containerHeight float64) {
	// Set root size/position. Root margin should offset the root box in the
	// viewport, matching CSS block flow behavior for top-level test widgets.
	style := root.Style()
	rootRect := Rect{
		X: style.Margin.Left,
		Y: style.Margin.Top,
	}
	if style.Width > 0 {
		rootRect.W = style.Width
	} else {
		rootRect.W = containerWidth - style.Margin.Left - style.Margin.Right
	}
	if style.Height > 0 {
		rootRect.H = style.Height
	} else {
		rootRect.H = containerHeight - style.Margin.Top - style.Margin.Bottom
	}
	if rootRect.W < 0 {
		rootRect.W = 0
	}
	if rootRect.H < 0 {
		rootRect.H = 0
	}
	root.SetComputedRect(rootRect)

	// Layout children
	le.layoutChildren(root)
}

// layoutChildren arranges children within a parent widget
func (le *LayoutEngine) layoutChildren(parent Widget) {
	allChildren := visibleLayoutChildren(parent.Children())
	if len(allChildren) == 0 {
		updateOverflowContentSize(parent)
		return
	}

	style := parent.Style()
	if bw, ok := parent.(*BaseWidget); ok {
		style = bw.getActiveStyle()
	}
	parentRect := parent.ComputedRect()

	// Available space after padding AND border (border-box model)
	bwTop := style.BorderTopWidth
	bwRight := style.BorderRightWidth
	bwBottom := style.BorderBottomWidth
	bwLeft := style.BorderLeftWidth
	if style.BorderWidth > 0 {
		if bwTop == 0 {
			bwTop = style.BorderWidth
		}
		if bwRight == 0 {
			bwRight = style.BorderWidth
		}
		if bwBottom == 0 {
			bwBottom = style.BorderWidth
		}
		if bwLeft == 0 {
			bwLeft = style.BorderWidth
		}
	}

	availX := parentRect.X + style.Padding.Left + bwLeft
	availY := parentRect.Y + style.Padding.Top + bwTop
	availW := parentRect.W - style.Padding.Left - style.Padding.Right - bwLeft - bwRight
	availH := parentRect.H - style.Padding.Top - style.Padding.Bottom - bwTop - bwBottom
	containingRect := Rect{X: availX, Y: availY, W: availW, H: availH}

	children, absoluteChildren := splitPositionedChildren(allChildren)
	if len(children) == 0 {
		for _, child := range absoluteChildren {
			le.layoutAbsoluteChild(child, containingRect)
		}
		return
	}

	direction := style.Direction
	if direction == "" {
		direction = LayoutColumn
	}

	gap := style.Gap
	if style.FlexWrap == FlexWrapNormal || style.FlexWrap == FlexWrapReverse {
		le.layoutWrappedChildren(parent, children, style, Rect{X: availX, Y: availY, W: availW, H: availH})
		updateOverflowContentSize(parent)
		return
	}

	// Calculate total fixed size and flex grow
	var totalFixed float64
	var totalFlexGrow float64
	for _, child := range children {
		childStyle := child.Style()
		if direction == LayoutRow {
			if childStyle.Width > 0 {
				totalFixed += outerWidth(childStyle, childStyle.Width) + childStyle.Margin.Left + childStyle.Margin.Right
			} else if childStyle.FlexGrow > 0 {
				totalFlexGrow += childStyle.FlexGrow
			} else {
				// Use intrinsic width if available
				iw := child.IntrinsicWidth()
				if iw <= 0 {
					iw = defaultIntrinsicWidth
				}
				totalFixed += iw + childStyle.Margin.Left + childStyle.Margin.Right
			}
		} else {
			if childStyle.Height > 0 {
				totalFixed += outerHeight(childStyle, childStyle.Height) + childStyle.Margin.Top + childStyle.Margin.Bottom
			} else if childStyle.FlexGrow > 0 {
				totalFlexGrow += childStyle.FlexGrow
			} else {
				// Use intrinsic height if available
				ih := child.IntrinsicHeight()
				if ih <= 0 {
					ih = defaultIntrinsicHeight
				}
				totalFixed += ih + childStyle.Margin.Top + childStyle.Margin.Bottom
			}
		}
	}

	// Add gaps
	totalGaps := gap * float64(len(children)-1)
	totalFixed += totalGaps

	// Calculate flex space
	var flexSpace float64
	if direction == LayoutRow {
		flexSpace = availW - totalFixed
	} else {
		flexSpace = availH - totalFixed
	}
	if flexSpace < 0 {
		flexSpace = 0
	}
	shrinkDelta := le.flexShrinkDelta(children, direction, availW, availH, gap)

	// Position children
	var offset float64
	switch style.Justify {
	case JustifyCenter:
		if direction == LayoutRow {
			offset = (availW - totalFixed) / 2
		} else {
			offset = (availH - totalFixed) / 2
		}
	case JustifyEnd:
		if direction == LayoutRow {
			offset = availW - totalFixed
		} else {
			offset = availH - totalFixed
		}
	}

	// Calculate spacing for space-between/around/evenly
	var spacingBetween float64
	useSpacing := false

	switch style.Justify {
	case JustifyBetween:
		if len(children) > 1 {
			var remainingSpace float64
			if direction == LayoutRow {
				remainingSpace = availW - totalFixed + totalGaps
			} else {
				remainingSpace = availH - totalFixed + totalGaps
			}
			spacingBetween = remainingSpace / float64(len(children)-1)
			useSpacing = true
		}
	case JustifyAround:
		var remainingSpace float64
		if direction == LayoutRow {
			remainingSpace = availW - totalFixed + totalGaps
		} else {
			remainingSpace = availH - totalFixed + totalGaps
		}
		spacingBetween = remainingSpace / float64(len(children))
		offset = spacingBetween / 2
		useSpacing = true
	case JustifyEvenly:
		var remainingSpace float64
		if direction == LayoutRow {
			remainingSpace = availW - totalFixed + totalGaps
		} else {
			remainingSpace = availH - totalFixed + totalGaps
		}
		spacingBetween = remainingSpace / float64(len(children)+1)
		offset = spacingBetween
		useSpacing = true
	}

	currentX := availX + offset
	currentY := availY + offset

	for i, child := range children {
		childStyle := child.Style()
		var childRect Rect

		if direction == LayoutRow {
			// Horizontal layout
			childRect.X = currentX + childStyle.Margin.Left
			childRect.Y = availY + childStyle.Margin.Top

			// Width (Main Axis)
			if childStyle.Width > 0 {
				childRect.W = outerWidth(childStyle, childStyle.Width)
			} else if childStyle.FlexGrow > 0 && totalFlexGrow > 0 {
				childRect.W = (childStyle.FlexGrow / totalFlexGrow) * flexSpace
			} else {
				iw := child.IntrinsicWidth()
				if iw <= 0 {
					iw = defaultIntrinsicWidth
				}
				childRect.W = iw
			}
			childRect.W = le.applyFlexShrink(childRect.W, childStyle, shrinkDelta, LayoutRow)

			// Height (Cross Axis)
			if childStyle.Height > 0 {
				childRect.H = outerHeight(childStyle, childStyle.Height)
			} else {
				// Default to stretch (fill cross-axis) unless align is center/end/start
				if style.Align == AlignCenter || style.Align == AlignEnd || style.Align == AlignStart {
					ih := child.IntrinsicHeight()
					if ih > 0 {
						childRect.H = ih
					} else {
						childRect.H = availH - childStyle.Margin.Top - childStyle.Margin.Bottom
					}
				} else {
					// Stretch (default)
					childRect.H = availH - childStyle.Margin.Top - childStyle.Margin.Bottom
				}
			}

			if childRect.H < 0 {
				childRect.H = 0
			}

			// Apply alignment (offset within Cross Axis)
			switch style.Align {
			case AlignCenter:
				childRect.Y = availY + (availH-childRect.H)/2
			case AlignEnd:
				childRect.Y = availY + availH - childRect.H - childStyle.Margin.Bottom
			}

			currentX += childRect.W + childStyle.Margin.Left + childStyle.Margin.Right
			if i < len(children)-1 {
				if useSpacing {
					currentX += spacingBetween
				} else {
					currentX += gap
				}
			}
		} else {
			// Vertical layout
			childRect.X = availX + childStyle.Margin.Left
			childRect.Y = currentY + childStyle.Margin.Top

			// Width
			if childStyle.Width > 0 {
				childRect.W = outerWidth(childStyle, childStyle.Width)
			} else {
				// Default to stretch (fill cross-axis) unless align is center/end/start
				if style.Align == AlignCenter || style.Align == AlignEnd || style.Align == AlignStart {
					childRect.W = child.IntrinsicWidth()
				} else {
					// Stretch
					childRect.W = availW - childStyle.Margin.Left - childStyle.Margin.Right
				}
			}

			if childRect.W < 0 {
				childRect.W = 0
			}

			// Height
			if childStyle.Height > 0 {
				childRect.H = outerHeight(childStyle, childStyle.Height)
			} else if childStyle.FlexGrow > 0 && totalFlexGrow > 0 {
				childRect.H = (childStyle.FlexGrow / totalFlexGrow) * flexSpace
			} else {
				ih := child.IntrinsicHeight()
				if ih <= 0 {
					ih = defaultIntrinsicHeight
				}
				childRect.H = ih
			}
			childRect.H = le.applyFlexShrink(childRect.H, childStyle, shrinkDelta, LayoutColumn)

			// Apply alignment
			switch style.Align {
			case AlignCenter:
				childRect.X = availX + (availW-childRect.W)/2
			case AlignEnd:
				childRect.X = availX + availW - childRect.W - childStyle.Margin.Right
			}

			currentY += childRect.H + childStyle.Margin.Top + childStyle.Margin.Bottom
			if i < len(children)-1 {
				if useSpacing {
					currentY += spacingBetween
				} else {
					currentY += gap
				}
			}
		}

		childRect.X = math.Round(childRect.X)
		childRect.Y = math.Round(childRect.Y)
		childRect.W = math.Round(childRect.W)
		childRect.H = math.Round(childRect.H)
		child.SetComputedRect(childRect)

		// Apply min/max constraints
		if childStyle.MinWidth > 0 && childRect.W < childStyle.MinWidth {
			childRect.W = childStyle.MinWidth
		}
		if childStyle.MaxWidth > 0 && childRect.W > childStyle.MaxWidth {
			childRect.W = childStyle.MaxWidth
		}
		if childStyle.MinHeight > 0 && childRect.H < childStyle.MinHeight {
			childRect.H = childStyle.MinHeight
		}
		if childStyle.MaxHeight > 0 && childRect.H > childStyle.MaxHeight {
			childRect.H = childStyle.MaxHeight
		}
		childRect.W = math.Round(childRect.W)
		childRect.H = math.Round(childRect.H)
		child.SetComputedRect(childRect)

		// Recursively layout grandchildren
		le.layoutChildren(child)
	}
	for _, child := range absoluteChildren {
		le.layoutAbsoluteChild(child, containingRect)
	}
	updateOverflowContentSize(parent)
}

func visibleLayoutChildren(children []Widget) []Widget {
	if len(children) == 0 {
		return nil
	}
	visible := make([]Widget, 0, len(children))
	for _, child := range children {
		if child == nil {
			continue
		}
		if !child.Visible() {
			child.SetComputedRect(Rect{})
			continue
		}
		visible = append(visible, child)
	}
	return visible
}

func updateOverflowContentSize(parent Widget) {
	bw := baseWidgetOf(parent)
	if bw == nil {
		return
	}
	content := bw.ContentRect()
	maxRight := content.X + content.W
	maxBottom := content.Y + content.H
	for _, child := range parent.Children() {
		rect := child.ComputedRect()
		if rect.X+rect.W > maxRight {
			maxRight = rect.X + rect.W
		}
		if rect.Y+rect.H > maxBottom {
			maxBottom = rect.Y + rect.H
		}
	}
	bw.setScrollContentSize(maxRight-content.X, maxBottom-content.Y)
}

func splitPositionedChildren(children []Widget) ([]Widget, []Widget) {
	normal := make([]Widget, 0, len(children))
	absolute := make([]Widget, 0)
	for _, child := range children {
		if child.Style().Position == "absolute" {
			absolute = append(absolute, child)
		} else {
			normal = append(normal, child)
		}
	}
	return normal, absolute
}

func (le *LayoutEngine) layoutAbsoluteChild(child Widget, containing Rect) {
	style := child.Style()
	w, h := preferredOuterSize(child)
	if style.WidthSet || style.Width > 0 {
		w = outerWidth(style, style.Width)
	}
	if style.HeightSet || style.Height > 0 {
		h = outerHeight(style, style.Height)
	}
	if style.LeftSet && style.RightSet && !style.WidthSet && style.Width <= 0 {
		w = containing.W - style.Left - style.Right - style.Margin.Left - style.Margin.Right
	}
	if style.TopSet && style.BottomSet && !style.HeightSet && style.Height <= 0 {
		h = containing.H - style.Top - style.Bottom - style.Margin.Top - style.Margin.Bottom
	}
	rect := Rect{W: w, H: h}
	if style.LeftSet {
		rect.X = containing.X + style.Left + style.Margin.Left
	} else if style.RightSet {
		rect.X = containing.X + containing.W - rect.W - style.Right - style.Margin.Right
	} else {
		rect.X = containing.X + style.Margin.Left
	}
	if style.TopSet {
		rect.Y = containing.Y + style.Top + style.Margin.Top
	} else if style.BottomSet {
		rect.Y = containing.Y + containing.H - rect.H - style.Bottom - style.Margin.Bottom
	} else {
		rect.Y = containing.Y + style.Margin.Top
	}
	rect = constrainedRect(rect, style)
	rect.X = math.Round(rect.X)
	rect.Y = math.Round(rect.Y)
	rect.W = math.Round(rect.W)
	rect.H = math.Round(rect.H)
	child.SetComputedRect(rect)
	le.layoutChildren(child)
}

func sortedChildrenByZ(children []Widget, reverse bool) []Widget {
	sorted := append([]Widget(nil), children...)
	sort.SliceStable(sorted, func(i, j int) bool {
		if reverse {
			return sorted[i].Style().ZIndex > sorted[j].Style().ZIndex
		}
		return sorted[i].Style().ZIndex < sorted[j].Style().ZIndex
	})
	return sorted
}

func (le *LayoutEngine) layoutWrappedChildren(parent Widget, children []Widget, style *Style, avail Rect) {
	direction := style.Direction
	if direction == "" {
		direction = LayoutColumn
	}
	gap := style.Gap

	type layoutItem struct {
		widget Widget
		rect   Rect
		main   float64
		cross  float64
	}
	var lines [][]layoutItem
	var line []layoutItem
	var lineMain float64
	lineLimit := avail.W
	if direction == LayoutColumn {
		lineLimit = avail.H
	}

	for _, child := range children {
		childStyle := child.Style()
		w, h := preferredOuterSize(child)
		mainSize := w + childStyle.Margin.Left + childStyle.Margin.Right
		crossSize := h + childStyle.Margin.Top + childStyle.Margin.Bottom
		if direction == LayoutColumn {
			mainSize = h + childStyle.Margin.Top + childStyle.Margin.Bottom
			crossSize = w + childStyle.Margin.Left + childStyle.Margin.Right
		}
		nextMain := mainSize
		if len(line) > 0 {
			nextMain = lineMain + gap + mainSize
		}
		if len(line) > 0 && nextMain > lineLimit {
			lines = append(lines, line)
			line = nil
			lineMain = 0
		}
		line = append(line, layoutItem{widget: child, rect: Rect{W: w, H: h}, main: mainSize, cross: crossSize})
		if lineMain > 0 {
			lineMain += gap
		}
		lineMain += mainSize
	}
	if len(line) > 0 {
		lines = append(lines, line)
	}
	if style.FlexWrap == FlexWrapReverse {
		for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
			lines[i], lines[j] = lines[j], lines[i]
		}
	}

	crossOffset := 0.0
	for _, line := range lines {
		lineMain = 0
		lineCross := 0.0
		for i, item := range line {
			if i > 0 {
				lineMain += gap
			}
			lineMain += item.main
			if item.cross > lineCross {
				lineCross = item.cross
			}
		}
		mainOffset, between := justifyOffsets(style.Justify, lineLimit, lineMain, len(line), gap)
		cursor := mainOffset
		for i, item := range line {
			child := item.widget
			childStyle := child.Style()
			rect := item.rect
			if direction == LayoutRow {
				rect.X = avail.X + cursor + childStyle.Margin.Left
				rect.Y = avail.Y + crossOffset + childStyle.Margin.Top
				if style.Align == AlignCenter {
					rect.Y = avail.Y + crossOffset + (lineCross-rect.H)/2
				} else if style.Align == AlignEnd {
					rect.Y = avail.Y + crossOffset + lineCross - rect.H - childStyle.Margin.Bottom
				} else if style.Align == AlignStretch && !childStyle.HeightSet && childStyle.Height <= 0 {
					rect.H = lineCross - childStyle.Margin.Top - childStyle.Margin.Bottom
				}
				cursor += item.main
			} else {
				rect.X = avail.X + crossOffset + childStyle.Margin.Left
				rect.Y = avail.Y + cursor + childStyle.Margin.Top
				if style.Align == AlignCenter {
					rect.X = avail.X + crossOffset + (lineCross-rect.W)/2
				} else if style.Align == AlignEnd {
					rect.X = avail.X + crossOffset + lineCross - rect.W - childStyle.Margin.Right
				} else if style.Align == AlignStretch && !childStyle.WidthSet && childStyle.Width <= 0 {
					rect.W = lineCross - childStyle.Margin.Left - childStyle.Margin.Right
				}
				cursor += item.main
			}
			if i < len(line)-1 {
				cursor += between
			}
			rect = constrainedRect(rect, childStyle)
			rect.X = math.Round(rect.X)
			rect.Y = math.Round(rect.Y)
			rect.W = math.Round(rect.W)
			rect.H = math.Round(rect.H)
			child.SetComputedRect(rect)
			le.layoutChildren(child)
		}
		crossOffset += lineCross + gap
	}
	updateOverflowContentSize(parent)
}

func (le *LayoutEngine) flexShrinkDelta(children []Widget, direction LayoutDirection, availW, availH, gap float64) float64 {
	gapCount := len(children) - 1
	if gapCount < 0 {
		gapCount = 0
	}
	total := gap * float64(gapCount)
	shrinkWeight := 0.0
	for _, child := range children {
		childStyle := child.Style()
		w, h := preferredOuterSize(child)
		if direction == LayoutRow {
			total += w + childStyle.Margin.Left + childStyle.Margin.Right
		} else {
			total += h + childStyle.Margin.Top + childStyle.Margin.Bottom
		}
		shrink := childStyle.FlexShrink
		if !childStyle.FlexShrinkSet {
			shrink = 1
		}
		shrinkWeight += shrink
	}
	limit := availW
	if direction == LayoutColumn {
		limit = availH
	}
	if total <= limit || shrinkWeight <= 0 {
		return 0
	}
	return (total - limit) / shrinkWeight
}

func (le *LayoutEngine) applyFlexShrink(size float64, style *Style, shrinkDelta float64, direction LayoutDirection) float64 {
	if shrinkDelta <= 0 {
		return size
	}
	shrink := style.FlexShrink
	if !style.FlexShrinkSet {
		shrink = 1
	}
	if shrink <= 0 {
		return size
	}
	size -= shrinkDelta * shrink
	if direction == LayoutRow && style.MinWidth > 0 && size < style.MinWidth {
		return style.MinWidth
	}
	if direction == LayoutColumn && style.MinHeight > 0 && size < style.MinHeight {
		return style.MinHeight
	}
	if size < 0 {
		return 0
	}
	return size
}

func preferredOuterSize(widget Widget) (float64, float64) {
	style := widget.Style()
	w := style.Width
	if w <= 0 {
		w = widget.IntrinsicWidth()
		if w <= 0 {
			w = defaultIntrinsicWidth
		}
	}
	h := style.Height
	if h <= 0 {
		h = widget.IntrinsicHeight()
		if h <= 0 {
			h = defaultIntrinsicHeight
		}
	}
	return outerWidth(style, w), outerHeight(style, h)
}

func outerWidth(style *Style, width float64) float64 {
	if style.BoxSizing == "border-box" {
		return width
	}
	return width + style.Padding.Left + style.Padding.Right + horizontalBorderWidth(style)
}

func outerHeight(style *Style, height float64) float64 {
	if style.BoxSizing == "border-box" {
		return height
	}
	return height + style.Padding.Top + style.Padding.Bottom + verticalBorderWidth(style)
}

func horizontalBorderWidth(style *Style) float64 {
	left, right := style.BorderLeftWidth, style.BorderRightWidth
	if style.BorderWidth > 0 {
		if left == 0 {
			left = style.BorderWidth
		}
		if right == 0 {
			right = style.BorderWidth
		}
	}
	return left + right
}

func verticalBorderWidth(style *Style) float64 {
	top, bottom := style.BorderTopWidth, style.BorderBottomWidth
	if style.BorderWidth > 0 {
		if top == 0 {
			top = style.BorderWidth
		}
		if bottom == 0 {
			bottom = style.BorderWidth
		}
	}
	return top + bottom
}

func constrainedRect(rect Rect, style *Style) Rect {
	if style.MinWidth > 0 && rect.W < style.MinWidth {
		rect.W = style.MinWidth
	}
	if style.MaxWidth > 0 && rect.W > style.MaxWidth {
		rect.W = style.MaxWidth
	}
	if style.MinHeight > 0 && rect.H < style.MinHeight {
		rect.H = style.MinHeight
	}
	if style.MaxHeight > 0 && rect.H > style.MaxHeight {
		rect.H = style.MaxHeight
	}
	return rect
}

func justifyOffsets(justify Justify, limit, used float64, count int, gap float64) (float64, float64) {
	remaining := limit - used
	if remaining < 0 {
		remaining = 0
	}
	switch justify {
	case JustifyCenter:
		return remaining / 2, gap
	case JustifyEnd:
		return remaining, gap
	case JustifyBetween:
		if count > 1 {
			return 0, gap + remaining/float64(count-1)
		}
	case JustifyAround:
		if count > 0 {
			space := remaining / float64(count)
			return space / 2, gap + space
		}
	case JustifyEvenly:
		if count > 0 {
			space := remaining / float64(count+1)
			return space, gap + space
		}
	}
	return 0, gap
}
