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
		child.SetComputedRect(childRect)

		// Recursively layout grandchildren
		le.layoutChildren(child)
	}
}
