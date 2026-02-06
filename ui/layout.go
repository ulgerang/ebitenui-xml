package ui

import "github.com/hajimehoshi/ebiten/v2"

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

// DrawWidget recursively draws a widget and its children
func DrawWidget(screen *ebiten.Image, widget Widget) {
	if widget == nil || !widget.Visible() {
		return
	}
	widget.Draw(screen)
	for _, child := range widget.Children() {
		DrawWidget(screen, child)
	}
}

// Layout calculates positions and sizes for a widget tree
func (le *LayoutEngine) Layout(root Widget, containerWidth, containerHeight float64) {
	// Set root size
	rootRect := Rect{X: 0, Y: 0, W: containerWidth, H: containerHeight}
	if root.Style().Width > 0 {
		rootRect.W = root.Style().Width
	}
	if root.Style().Height > 0 {
		rootRect.H = root.Style().Height
	}
	root.SetComputedRect(rootRect)

	// Layout children
	le.layoutChildren(root)
}

// layoutChildren arranges children within a parent widget
func (le *LayoutEngine) layoutChildren(parent Widget) {
	children := parent.Children()
	if len(children) == 0 {
		return
	}

	style := parent.Style()
	parentRect := parent.ComputedRect()

	// Available space after padding
	availX := parentRect.X + style.Padding.Left
	availY := parentRect.Y + style.Padding.Top
	availW := parentRect.W - style.Padding.Left - style.Padding.Right
	availH := parentRect.H - style.Padding.Top - style.Padding.Bottom

	direction := style.Direction
	if direction == "" {
		direction = LayoutColumn
	}

	gap := style.Gap

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
	if flexSpace < 0 {
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
				childRect.W = childStyle.Width
			} else if childStyle.FlexGrow > 0 && totalFlexGrow > 0 {
				childRect.W = (childStyle.FlexGrow / totalFlexGrow) * flexSpace
			} else {
				childRect.W = 50 // Default
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
				currentX += gap
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
				childRect.H = childStyle.Height
			} else if childStyle.FlexGrow > 0 && totalFlexGrow > 0 {
				childRect.H = (childStyle.FlexGrow / totalFlexGrow) * flexSpace
			} else {
				childRect.H = 30 // Default
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
				currentY += gap
			}
		}

		child.SetComputedRect(childRect)

		// Recursively layout grandchildren
		le.layoutChildren(child)
	}
}
