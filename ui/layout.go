package ui

import (
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
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

// DrawWidget recursively draws a widget and its children.
// It respects z-index ordering, display:none, and visibility:hidden.
func DrawWidget(screen *ebiten.Image, widget Widget) {
	if widget == nil || !widget.Visible() {
		return
	}
	s := widget.Style()
	// display:none → skip entirely (no space, no rendering)
	if s.Display == "none" {
		return
	}
	// visibility:hidden → skip rendering but children may still be visible
	if s.Visibility != "hidden" {
		widget.Draw(screen)
	}
	children := sortByZIndex(widget.Children())
	for _, child := range children {
		DrawWidget(screen, child)
	}
}

// sortByZIndex returns children sorted by z-index (ascending, higher draws on top).
// If no child has a non-zero z-index, returns the original slice unchanged.
func sortByZIndex(children []Widget) []Widget {
	if len(children) <= 1 {
		return children
	}
	hasZIndex := false
	for _, c := range children {
		if c.Style().ZIndex != 0 {
			hasZIndex = true
			break
		}
	}
	if !hasZIndex {
		return children
	}
	sorted := make([]Widget, len(children))
	copy(sorted, children)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Style().ZIndex < sorted[j].Style().ZIndex
	})
	return sorted
}

// sortByZIndexReverse returns children sorted by z-index descending (highest first).
// Used for hit-testing where higher z-index should be checked first.
func sortByZIndexReverse(children []Widget) []Widget {
	if len(children) <= 1 {
		return children
	}
	hasZIndex := false
	for _, c := range children {
		if c.Style().ZIndex != 0 {
			hasZIndex = true
			break
		}
	}
	if !hasZIndex {
		return children
	}
	sorted := make([]Widget, len(children))
	copy(sorted, children)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Style().ZIndex > sorted[j].Style().ZIndex
	})
	return sorted
}

// Layout calculates positions and sizes for a widget tree
func (le *LayoutEngine) Layout(root Widget, containerWidth, containerHeight float64) {
	// Set root size
	rootStyle := root.Style()
	rootRect := Rect{X: 0, Y: 0, W: containerWidth, H: containerHeight}

	// Apply root margin (like CSS block-level margin)
	marginLeft := rootStyle.Margin.Left
	marginTop := rootStyle.Margin.Top
	rootRect.X += marginLeft
	rootRect.Y += marginTop

	if rootStyle.Width > 0 {
		rootRect.W = rootStyle.Width
	}
	if rootStyle.Height > 0 {
		rootRect.H = rootStyle.Height
	}
	root.SetComputedRect(rootRect)

	// Layout children
	le.layoutChildren(root)
}

// layoutChildren arranges children within a parent widget.
// Supports display:none (skip), position:absolute (out of flow),
// and flex-wrap (multi-line layouts).
func (le *LayoutEngine) layoutChildren(parent Widget) {
	children := parent.Children()
	if len(children) == 0 {
		return
	}

	style := parent.Style()
	parentRect := parent.ComputedRect()

	// Available space after padding AND border (border-box model)
	borderW := style.BorderWidth
	availX := parentRect.X + style.Padding.Left + borderW
	availY := parentRect.Y + style.Padding.Top + borderW
	availW := parentRect.W - style.Padding.Left - style.Padding.Right - borderW*2
	availH := parentRect.H - style.Padding.Top - style.Padding.Bottom - borderW*2

	direction := style.Direction
	if direction == "" {
		direction = LayoutColumn
	}

	gap := style.Gap

	// Separate normal-flow children from absolutely positioned children
	var flowChildren []Widget
	var absChildren []Widget
	for _, child := range children {
		cs := child.Style()
		// display:none → skip entirely
		if cs.Display == "none" {
			continue
		}
		if cs.Position == "absolute" || cs.Position == "fixed" {
			absChildren = append(absChildren, child)
		} else {
			flowChildren = append(flowChildren, child)
		}
	}

	// Layout absolutely positioned children
	for _, child := range absChildren {
		le.layoutAbsolute(child, availX, availY, availW, availH)
	}

	if len(flowChildren) == 0 {
		return
	}

	// Check if flex-wrap is enabled
	if style.FlexWrap == FlexWrapNormal || style.FlexWrap == FlexWrapReverse {
		le.layoutFlexWrap(flowChildren, style, direction, gap, availX, availY, availW, availH)
		return
	}

	// Normal (single-line) flex layout
	le.layoutFlexLine(flowChildren, style, direction, gap, availX, availY, availW, availH)
}

// layoutAbsolute positions an absolutely positioned child relative to its parent's padding box.
func (le *LayoutEngine) layoutAbsolute(child Widget, availX, availY, availW, availH float64) {
	cs := child.Style()
	var childRect Rect

	// Width
	if cs.Width > 0 {
		childRect.W = cs.Width
	} else {
		// If both left and right are set, width = available - left - right
		if cs.Left != 0 && cs.Right != 0 {
			childRect.W = availW - cs.Left - cs.Right
		} else {
			childRect.W = 50 // default
		}
	}

	// Height
	if cs.Height > 0 {
		childRect.H = cs.Height
	} else {
		if cs.Top != 0 && cs.Bottom != 0 {
			childRect.H = availH - cs.Top - cs.Bottom
		} else {
			childRect.H = 30 // default
		}
	}

	// Horizontal position
	if cs.Left != 0 || cs.Right == 0 {
		childRect.X = availX + cs.Left
	} else {
		childRect.X = availX + availW - childRect.W - cs.Right
	}

	// Vertical position
	if cs.Top != 0 || cs.Bottom == 0 {
		childRect.Y = availY + cs.Top
	} else {
		childRect.Y = availY + availH - childRect.H - cs.Bottom
	}

	// Apply min/max
	le.applyMinMax(cs, &childRect)
	child.SetComputedRect(childRect)
	le.layoutChildren(child)
}

// layoutFlexWrap handles multi-line flex layout.
func (le *LayoutEngine) layoutFlexWrap(children []Widget, style *Style, direction LayoutDirection, gap, availX, availY, availW, availH float64) {
	// Break children into lines
	type flexLine struct {
		widgets  []Widget
		mainSize float64
	}

	var lines []flexLine
	var currentLine flexLine
	var lineMainSize float64
	mainAvail := availW
	if direction == LayoutColumn {
		mainAvail = availH
	}

	for _, child := range children {
		cs := child.Style()
		var itemMain float64
		if direction == LayoutRow {
			if cs.Width > 0 {
				itemMain = cs.Width + cs.Margin.Left + cs.Margin.Right
			} else {
				itemMain = 50 + cs.Margin.Left + cs.Margin.Right
			}
		} else {
			if cs.Height > 0 {
				itemMain = cs.Height + cs.Margin.Top + cs.Margin.Bottom
			} else {
				itemMain = 30 + cs.Margin.Top + cs.Margin.Bottom
			}
		}

		// Add gap if not first item on the line
		gapAdd := 0.0
		if len(currentLine.widgets) > 0 {
			gapAdd = gap
		}

		if len(currentLine.widgets) > 0 && lineMainSize+gapAdd+itemMain > mainAvail {
			// Start new line
			currentLine.mainSize = lineMainSize
			lines = append(lines, currentLine)
			currentLine = flexLine{}
			lineMainSize = 0
			gapAdd = 0
		}

		currentLine.widgets = append(currentLine.widgets, child)
		lineMainSize += gapAdd + itemMain
	}
	if len(currentLine.widgets) > 0 {
		currentLine.mainSize = lineMainSize
		lines = append(lines, currentLine)
	}

	// Reverse lines if wrap-reverse
	if style.FlexWrap == FlexWrapReverse {
		for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
			lines[i], lines[j] = lines[j], lines[i]
		}
	}

	// Layout each line
	crossOffset := 0.0
	for _, line := range lines {
		if direction == LayoutRow {
			lineAvailH := availH / float64(len(lines))
			le.layoutFlexLine(line.widgets, style, direction, gap, availX, availY+crossOffset, availW, lineAvailH)
			// Calculate max height of this line
			maxH := 0.0
			for _, w := range line.widgets {
				r := w.ComputedRect()
				h := r.H + w.Style().Margin.Top + w.Style().Margin.Bottom
				if h > maxH {
					maxH = h
				}
			}
			crossOffset += maxH + gap
		} else {
			lineAvailW := availW / float64(len(lines))
			le.layoutFlexLine(line.widgets, style, direction, gap, availX+crossOffset, availY, lineAvailW, availH)
			// Calculate max width of this line
			maxW := 0.0
			for _, w := range line.widgets {
				r := w.ComputedRect()
				ww := r.W + w.Style().Margin.Left + w.Style().Margin.Right
				if ww > maxW {
					maxW = ww
				}
			}
			crossOffset += maxW + gap
		}
	}
}

// layoutFlexLine lays out a single line of flex children (the original layout logic).
func (le *LayoutEngine) layoutFlexLine(children []Widget, style *Style, direction LayoutDirection, gap, availX, availY, availW, availH float64) {
	// Calculate total fixed size and flex grow
	var totalFixed float64
	var totalFlexGrow float64
	for _, child := range children {
		childStyle := child.Style()
		if direction == LayoutRow {
			if childStyle.Width > 0 {
				totalFixed += childStyle.Width + childStyle.Margin.Left + childStyle.Margin.Right
			} else if childStyle.FlexGrow > 0 {
				totalFlexGrow += childStyle.FlexGrow
			} else {
				// Default minimum size
				totalFixed += 50 + childStyle.Margin.Left + childStyle.Margin.Right
			}
		} else {
			if childStyle.Height > 0 {
				totalFixed += childStyle.Height + childStyle.Margin.Top + childStyle.Margin.Bottom
			} else if childStyle.FlexGrow > 0 {
				totalFlexGrow += childStyle.FlexGrow
			} else {
				totalFixed += 30 + childStyle.Margin.Top + childStyle.Margin.Bottom
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
	var shrinkFactor float64 = 1.0
	if flexSpace < 0 {
		// Calculate how much to shrink
		if direction == LayoutRow {
			if totalFixed > 0 {
				shrinkFactor = availW / totalFixed
			}
		} else {
			if totalFixed > 0 {
				shrinkFactor = availH / totalFixed
			}
		}
		if shrinkFactor < 0 {
			shrinkFactor = 0
		}
		if shrinkFactor > 1 {
			shrinkFactor = 1
		}
		flexSpace = 0
	}

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

			// Width
			if childStyle.Width > 0 {
				childRect.W = childStyle.Width * shrinkFactor
			} else if childStyle.FlexGrow > 0 && totalFlexGrow > 0 {
				childRect.W = (childStyle.FlexGrow / totalFlexGrow) * flexSpace
			} else {
				childRect.W = 50 * shrinkFactor // Default
			}

			// Height
			if childStyle.Height > 0 {
				childRect.H = childStyle.Height
			} else {
				childRect.H = availH - childStyle.Margin.Top - childStyle.Margin.Bottom
			}

			// Apply alignment
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

			// Width - use parent's computed available width
			if childStyle.Width > 0 {
				childRect.W = childStyle.Width
			} else {
				// Constrain to available width minus margins
				childRect.W = availW - childStyle.Margin.Left - childStyle.Margin.Right
				if childRect.W < 0 {
					childRect.W = 0
				}
			}

			// Height
			if childStyle.Height > 0 {
				childRect.H = childStyle.Height * shrinkFactor
			} else if childStyle.FlexGrow > 0 && totalFlexGrow > 0 {
				childRect.H = (childStyle.FlexGrow / totalFlexGrow) * flexSpace
			} else {
				childRect.H = 30 * shrinkFactor // Default
			}

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

		child.SetComputedRect(childRect)

		// Apply min/max constraints
		le.applyMinMax(childStyle, &childRect)
		child.SetComputedRect(childRect)

		// Recursively layout grandchildren
		le.layoutChildren(child)
	}
}

// applyMinMax applies min/max width/height constraints to a rect.
func (le *LayoutEngine) applyMinMax(s *Style, r *Rect) {
	if s.MinWidth > 0 && r.W < s.MinWidth {
		r.W = s.MinWidth
	}
	if s.MaxWidth > 0 && r.W > s.MaxWidth {
		r.W = s.MaxWidth
	}
	if s.MinHeight > 0 && r.H < s.MinHeight {
		r.H = s.MinHeight
	}
	if s.MaxHeight > 0 && r.H > s.MaxHeight {
		r.H = s.MaxHeight
	}
}
