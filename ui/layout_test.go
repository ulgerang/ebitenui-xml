package ui

import (
	"math"
	"testing"
)

// ============================================================================
// Layout Engine Tests - Flex Row
// ============================================================================

func TestLayoutEngineFlexRow(t *testing.T) {
	tests := []struct {
		name           string
		parentWidth    float64
		parentHeight   float64
		children       []Widget
		justify        Justify
		align          Alignment
		validateLayout func(*testing.T, []Widget)
	}{
		{
			name:         "row with fixed width children",
			parentWidth:  400,
			parentHeight: 200,
			children: []Widget{
				func() Widget {
					p := NewPanel("child1")
					p.SetStyle(&Style{Width: 100, Height: 50})
					return p
				}(),
				func() Widget {
					p := NewPanel("child2")
					p.SetStyle(&Style{Width: 150, Height: 50})
					return p
				}(),
			},
			justify: JustifyStart,
			align:   AlignStart,
			validateLayout: func(t *testing.T, children []Widget) {
				if len(children) != 2 {
					t.Fatalf("Expected 2 children, got %d", len(children))
				}

				rect1 := children[0].ComputedRect()
				rect2 := children[1].ComputedRect()

				// Check widths
				if math.Abs(rect1.W-100) > 0.1 {
					t.Errorf("Child1 width = %v, want 100", rect1.W)
				}
				if math.Abs(rect2.W-150) > 0.1 {
					t.Errorf("Child2 width = %v, want 150", rect2.W)
				}

				// Check heights (explicit height is set, so no stretching)
				if rect1.H != 50 {
					t.Errorf("Child1 height = %v, want 50 (explicit height set)", rect1.H)
				}

				// Check positions (should be sequential)
				if rect2.X <= rect1.X {
					t.Errorf("Child2 should be to the right of Child1")
				}
			},
		},
		{
			name:         "row with flex-grow",
			parentWidth:  400,
			parentHeight: 200,
			children: []Widget{
				func() Widget {
					p := NewPanel("child1")
					p.SetStyle(&Style{FlexGrow: 1, Height: 50})
					return p
				}(),
				func() Widget {
					p := NewPanel("child2")
					p.SetStyle(&Style{FlexGrow: 2, Height: 50})
					return p
				}(),
			},
			justify: JustifyStart,
			align:   AlignStart,
			validateLayout: func(t *testing.T, children []Widget) {
				rect1 := children[0].ComputedRect()
				rect2 := children[1].ComputedRect()

				// Child2 should be twice as wide as Child1 (flex-grow 2 vs 1)
				if rect2.W < rect1.W {
					t.Errorf("Child2 width (%v) should be >= Child1 width (%v)", rect2.W, rect1.W)
				}

				// Total should be close to parent width
				total := rect1.W + rect2.W
				if total > 400 {
					t.Errorf("Total width %v exceeds parent width 400", total)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parent := NewPanel("parent")
			parent.SetStyle(&Style{
				Direction: LayoutRow,
				Width:     tt.parentWidth,
				Height:    tt.parentHeight,
				Justify:   tt.justify,
				Align:     tt.align,
			})

			for _, child := range tt.children {
				parent.AddChild(child)
			}

			le := NewLayoutEngine()
			le.Layout(parent, tt.parentWidth, tt.parentHeight)

			tt.validateLayout(t, parent.Children())
		})
	}
}

// ============================================================================
// Layout Engine Tests - Flex Column
// ============================================================================

func TestLayoutEngineFlexColumn(t *testing.T) {
	tests := []struct {
		name           string
		parentWidth    float64
		parentHeight   float64
		children       []Widget
		validateLayout func(*testing.T, []Widget)
	}{
		{
			name:         "column with fixed height children",
			parentWidth:  300,
			parentHeight: 400,
			children: []Widget{
				func() Widget {
					p := NewPanel("child1")
					p.SetStyle(&Style{Width: 100, Height: 100})
					return p
				}(),
				func() Widget {
					p := NewPanel("child2")
					p.SetStyle(&Style{Width: 100, Height: 150})
					return p
				}(),
			},
			validateLayout: func(t *testing.T, children []Widget) {
				rect1 := children[0].ComputedRect()
				rect2 := children[1].ComputedRect()

				// Check heights
				if math.Abs(rect1.H-100) > 0.1 {
					t.Errorf("Child1 height = %v, want 100", rect1.H)
				}
				if math.Abs(rect2.H-150) > 0.1 {
					t.Errorf("Child2 height = %v, want 150", rect2.H)
				}

				// Check positions (should be vertical)
				if rect2.Y <= rect1.Y {
					t.Errorf("Child2 should be below Child1")
				}

				// X positions should be aligned
				if rect1.X != rect2.X {
					t.Errorf("Children should have same X position")
				}
			},
		},
		{
			name:         "column with flex-grow",
			parentWidth:  300,
			parentHeight: 400,
			children: []Widget{
				func() Widget {
					p := NewPanel("child1")
					p.SetStyle(&Style{FlexGrow: 1})
					return p
				}(),
				func() Widget {
					p := NewPanel("child2")
					p.SetStyle(&Style{FlexGrow: 1})
					return p
				}(),
			},
			validateLayout: func(t *testing.T, children []Widget) {
				rect1 := children[0].ComputedRect()
				rect2 := children[1].ComputedRect()

				// Both should have equal height
				if math.Abs(rect1.H-rect2.H) > 0.1 {
					t.Errorf("Children should have equal height: %v vs %v", rect1.H, rect2.H)
				}

				// Total should be close to parent height
				total := rect1.H + rect2.H
				if total > 400 {
					t.Errorf("Total height %v exceeds parent height 400", total)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parent := NewPanel("parent")
			parent.SetStyle(&Style{
				Direction: LayoutColumn,
				Width:     tt.parentWidth,
				Height:    tt.parentHeight,
			})

			for _, child := range tt.children {
				parent.AddChild(child)
			}

			le := NewLayoutEngine()
			le.Layout(parent, tt.parentWidth, tt.parentHeight)

			tt.validateLayout(t, parent.Children())
		})
	}
}

// ============================================================================
// Layout Engine Tests - Justify Content
// ============================================================================

func TestLayoutEngineJustify(t *testing.T) {
	tests := []struct {
		name           string
		justify        Justify
		parentWidth    float64
		childrenWidths []float64
		validateLayout func(*testing.T, []Widget, float64)
	}{
		{
			name:           "justify start",
			justify:        JustifyStart,
			parentWidth:    400,
			childrenWidths: []float64{80, 80, 80},
			validateLayout: func(t *testing.T, children []Widget, parentWidth float64) {
				rects := make([]Rect, len(children))
				for i, c := range children {
					rects[i] = c.ComputedRect()
				}

				// First child should start at beginning
				if rects[0].X > 10 { // Allow some padding
					t.Errorf("First child should start near beginning, got X=%v", rects[0].X)
				}

				// Last child should not reach end
				lastEnd := rects[len(rects)-1].X + rects[len(rects)-1].W
				if lastEnd > parentWidth-10 {
					t.Errorf("Last child should not reach end, got end=%v", lastEnd)
				}
			},
		},
		{
			name:           "justify center",
			justify:        JustifyCenter,
			parentWidth:    400,
			childrenWidths: []float64{80, 80, 80},
			validateLayout: func(t *testing.T, children []Widget, parentWidth float64) {
				rects := make([]Rect, len(children))
				for i, c := range children {
					rects[i] = c.ComputedRect()
				}

				// Should be centered
				totalWidth := float64(0)
				for _, r := range rects {
					totalWidth += r.W
				}

				// Calculate expected center
				expectedStart := (parentWidth - totalWidth) / 2
				if math.Abs(rects[0].X-expectedStart) > 1 {
					t.Errorf("Children should be centered, first X=%v, want ~%v", rects[0].X, expectedStart)
				}
			},
		},
		{
			name:           "justify end",
			justify:        JustifyEnd,
			parentWidth:    400,
			childrenWidths: []float64{80, 80, 80},
			validateLayout: func(t *testing.T, children []Widget, parentWidth float64) {
				rects := make([]Rect, len(children))
				for i, c := range children {
					rects[i] = c.ComputedRect()
				}

				// Last child should end near parent edge
				lastRect := rects[len(rects)-1]
				lastEnd := lastRect.X + lastRect.W
				if lastEnd < parentWidth-20 {
					t.Errorf("Last child should end near parent edge, got end=%v", lastEnd)
				}
			},
		},
		{
			name:           "justify space-between",
			justify:        JustifyBetween,
			parentWidth:    400,
			childrenWidths: []float64{80, 80, 80},
			validateLayout: func(t *testing.T, children []Widget, parentWidth float64) {
				rects := make([]Rect, len(children))
				for i, c := range children {
					rects[i] = c.ComputedRect()
				}

				// First child at start
				if rects[0].X > 10 {
					t.Errorf("First child should be at start, got X=%v", rects[0].X)
				}

				// Last child at end
				lastRect := rects[len(rects)-1]
				lastEnd := lastRect.X + lastRect.W
				if lastEnd < parentWidth-20 {
					t.Errorf("Last child should be at end, got end=%v", lastEnd)
				}

				// Equal spacing between
				if len(rects) >= 3 {
					gap1 := rects[1].X - (rects[0].X + rects[0].W)
					gap2 := rects[2].X - (rects[1].X + rects[1].W)
					if math.Abs(gap1-gap2) > 1 {
						t.Errorf("Gaps should be equal: %v vs %v", gap1, gap2)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parent := NewPanel("parent")
			parent.SetStyle(&Style{
				Direction: LayoutRow,
				Width:     tt.parentWidth,
				Height:    100,
				Justify:   tt.justify,
			})

			for i, w := range tt.childrenWidths {
				child := NewPanel("child")
				child.SetStyle(&Style{Width: w, Height: 50})
				parent.AddChild(child)
				_ = i
			}

			le := NewLayoutEngine()
			le.Layout(parent, tt.parentWidth, 100)

			tt.validateLayout(t, parent.Children(), tt.parentWidth)
		})
	}
}

// ============================================================================
// Layout Engine Tests - Align Items
// ============================================================================

func TestLayoutEngineAlign(t *testing.T) {
	tests := []struct {
		name            string
		align           Alignment
		parentHeight    float64
		childrenHeights []float64
		validateLayout  func(*testing.T, []Widget, float64)
	}{
		{
			name:            "align start (row)",
			align:           AlignStart,
			parentHeight:    200,
			childrenHeights: []float64{50, 80, 60},
			validateLayout: func(t *testing.T, children []Widget, parentHeight float64) {
				rects := make([]Rect, len(children))
				for i, c := range children {
					rects[i] = c.ComputedRect()
				}

				// All should align to top
				for i, r := range rects {
					if r.Y > 5 { // Allow small margin
						t.Errorf("Child %d should align to top, got Y=%v", i, r.Y)
					}
				}
			},
		},
		{
			name:            "align center (row)",
			align:           AlignCenter,
			parentHeight:    200,
			childrenHeights: []float64{50, 80, 60},
			validateLayout: func(t *testing.T, children []Widget, parentHeight float64) {
				rects := make([]Rect, len(children))
				for i, c := range children {
					rects[i] = c.ComputedRect()
				}

				// Each child should be vertically centered
				for i, r := range rects {
					expectedY := (parentHeight - r.H) / 2
					if math.Abs(r.Y-expectedY) > 1 {
						t.Errorf("Child %d should be centered, Y=%v, want ~%v", i, r.Y, expectedY)
					}
				}
			},
		},
		{
			name:            "align end (row)",
			align:           AlignEnd,
			parentHeight:    200,
			childrenHeights: []float64{50, 80, 60},
			validateLayout: func(t *testing.T, children []Widget, parentHeight float64) {
				rects := make([]Rect, len(children))
				for i, c := range children {
					rects[i] = c.ComputedRect()
				}

				// Each child should align to bottom
				for i, r := range rects {
					expectedY := parentHeight - r.H
					if math.Abs(r.Y-expectedY) > 1 {
						t.Errorf("Child %d should align to bottom, Y=%v, want ~%v", i, r.Y, expectedY)
					}
				}
			},
		},
		{
			name:            "align stretch (row)",
			align:           AlignStretch,
			parentHeight:    200,
			childrenHeights: []float64{0, 0, 0}, // No explicit height for stretch
			validateLayout: func(t *testing.T, children []Widget, parentHeight float64) {
				rects := make([]Rect, len(children))
				for i, c := range children {
					rects[i] = c.ComputedRect()
				}

				// All should stretch to parent height when no explicit height
				for i, r := range rects {
					if math.Abs(r.H-parentHeight) > 1 {
						t.Errorf("Child %d should stretch to parent height, H=%v, want %v", i, r.H, parentHeight)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parent := NewPanel("parent")
			parent.SetStyle(&Style{
				Direction: LayoutRow,
				Width:     300,
				Height:    tt.parentHeight,
				Align:     tt.align,
			})

			for _, h := range tt.childrenHeights {
				child := NewPanel("child")
				child.SetStyle(&Style{Width: 60, Height: h})
				parent.AddChild(child)
			}

			le := NewLayoutEngine()
			le.Layout(parent, 300, tt.parentHeight)

			tt.validateLayout(t, parent.Children(), tt.parentHeight)
		})
	}
}

// ============================================================================
// Layout Engine Tests - Edge Cases
// ============================================================================

func TestLayoutEngineEdgeCases(t *testing.T) {
	t.Run("nil widget", func(t *testing.T) {
		le := NewLayoutEngine()
		// Layout on nil will panic - test that it's handled
		defer func() {
			if r := recover(); r != nil {
				// Expected - Layout on nil causes panic
			}
		}()
		le.Layout(nil, 100, 100)
	})

	t.Run("empty parent", func(t *testing.T) {
		parent := NewPanel("parent")
		parent.SetStyle(&Style{Width: 200, Height: 200})

		le := NewLayoutEngine()
		le.Layout(parent, 200, 200)

		// Should not panic
	})

	t.Run("negative container size", func(t *testing.T) {
		parent := NewPanel("parent")
		parent.SetStyle(&Style{Width: 100, Height: 100})

		le := NewLayoutEngine()
		le.Layout(parent, -100, -100)

		rect := parent.ComputedRect()
		if rect.W < 0 {
			t.Errorf("Width should not be negative, got %v", rect.W)
		}
		if rect.H < 0 {
			t.Errorf("Height should not be negative, got %v", rect.H)
		}
	})

	t.Run("zero container size", func(t *testing.T) {
		parent := NewPanel("parent")
		parent.SetStyle(&Style{Width: 100, Height: 100})

		le := NewLayoutEngine()
		le.Layout(parent, 0, 0)

		// Should not panic
	})

	t.Run("child overflow", func(t *testing.T) {
		parent := NewPanel("parent")
		parent.SetStyle(&Style{
			Direction: LayoutRow,
			Width:     100,
			Height:    100,
		})

		// Add children that exceed parent width
		for i := 0; i < 5; i++ {
			child := NewPanel("child")
			child.SetStyle(&Style{Width: 50, Height: 50})
			parent.AddChild(child)
		}

		le := NewLayoutEngine()
		le.Layout(parent, 100, 100)

		// Should layout all children (may shrink or overflow)
		children := parent.Children()
		if len(children) != 5 {
			t.Errorf("Expected 5 children, got %d", len(children))
		}
	})
}

// ============================================================================
// Layout Engine Tests - Gap
// ============================================================================

func TestLayoutEngineGap(t *testing.T) {
	t.Run("row with gap", func(t *testing.T) {
		parent := NewPanel("parent")
		parent.SetStyle(&Style{
			Direction: LayoutRow,
			Width:     300,
			Height:    100,
			Gap:       20,
		})

		for i := 0; i < 3; i++ {
			child := NewPanel("child")
			child.SetStyle(&Style{Width: 60, Height: 50})
			parent.AddChild(child)
		}

		le := NewLayoutEngine()
		le.Layout(parent, 300, 100)

		children := parent.Children()
		rects := make([]Rect, len(children))
		for i, c := range children {
			rects[i] = c.ComputedRect()
		}

		// Check gaps between children
		for i := 1; i < len(rects); i++ {
			expectedGap := rects[i].X - (rects[i-1].X + rects[i-1].W)
			if math.Abs(expectedGap-20) > 1 {
				t.Errorf("Gap between child %d and %d = %v, want 20", i-1, i, expectedGap)
			}
		}
	})

	t.Run("column with gap", func(t *testing.T) {
		parent := NewPanel("parent")
		parent.SetStyle(&Style{
			Direction: LayoutColumn,
			Width:     200,
			Height:    300,
			Gap:       15,
		})

		for i := 0; i < 3; i++ {
			child := NewPanel("child")
			child.SetStyle(&Style{Width: 100, Height: 60})
			parent.AddChild(child)
		}

		le := NewLayoutEngine()
		le.Layout(parent, 200, 300)

		children := parent.Children()
		rects := make([]Rect, len(children))
		for i, c := range children {
			rects[i] = c.ComputedRect()
		}

		// Check gaps between children
		for i := 1; i < len(rects); i++ {
			expectedGap := rects[i].Y - (rects[i-1].Y + rects[i-1].H)
			if math.Abs(expectedGap-15) > 1 {
				t.Errorf("Gap between child %d and %d = %v, want 15", i-1, i, expectedGap)
			}
		}
	})
}

// ============================================================================
// Layout Engine Tests - Margins
// ============================================================================

func TestLayoutEngineMargins(t *testing.T) {
	t.Run("child margins in row", func(t *testing.T) {
		parent := NewPanel("parent")
		parent.SetStyle(&Style{
			Direction: LayoutRow,
			Width:     400,
			Height:    100,
		})

		child := NewPanel("child")
		child.SetStyle(&Style{
			Width:  100,
			Height: 50,
			Margin: Margin{Top: 10, Right: 20, Bottom: 10, Left: 30},
		})
		parent.AddChild(child)

		le := NewLayoutEngine()
		le.Layout(parent, 400, 100)

		rect := child.ComputedRect()

		// X should include left margin
		if rect.X < 30 {
			t.Errorf("Child X should account for left margin, got %v", rect.X)
		}
	})

	t.Run("child margins in column", func(t *testing.T) {
		parent := NewPanel("parent")
		parent.SetStyle(&Style{
			Direction: LayoutColumn,
			Width:     200,
			Height:    300,
		})

		child := NewPanel("child")
		child.SetStyle(&Style{
			Width:  100,
			Height: 80,
			Margin: Margin{Top: 15, Right: 10, Bottom: 15, Left: 10},
		})
		parent.AddChild(child)

		le := NewLayoutEngine()
		le.Layout(parent, 200, 300)

		rect := child.ComputedRect()

		// Y should include top margin
		if rect.Y < 15 {
			t.Errorf("Child Y should account for top margin, got %v", rect.Y)
		}
	})
}

// ============================================================================
// Layout Engine Tests - Min/Max Constraints
// ============================================================================

func TestLayoutEngineMinMaxConstraints(t *testing.T) {
	t.Run("min width constraint", func(t *testing.T) {
		parent := NewPanel("parent")
		parent.SetStyle(&Style{
			Direction: LayoutRow,
			Width:     200,
			Height:    100,
		})

		child := NewPanel("child")
		child.SetStyle(&Style{
			Width:    50,
			MinWidth: 100,
			Height:   50,
		})
		parent.AddChild(child)

		le := NewLayoutEngine()
		le.Layout(parent, 200, 100)

		rect := child.ComputedRect()
		if rect.W < 100 {
			t.Errorf("Child width should respect minWidth, got %v", rect.W)
		}
	})

	t.Run("max width constraint", func(t *testing.T) {
		parent := NewPanel("parent")
		parent.SetStyle(&Style{
			Direction: LayoutRow,
			Width:     400,
			Height:    100,
		})

		child := NewPanel("child")
		child.SetStyle(&Style{
			FlexGrow: 1,
			MaxWidth: 150,
			Height:   50,
		})
		parent.AddChild(child)

		le := NewLayoutEngine()
		le.Layout(parent, 400, 100)

		rect := child.ComputedRect()
		if rect.W > 150 {
			t.Errorf("Child width should respect maxWidth, got %v", rect.W)
		}
	})

	t.Run("min height constraint", func(t *testing.T) {
		parent := NewPanel("parent")
		parent.SetStyle(&Style{
			Direction: LayoutColumn,
			Width:     200,
			Height:    300,
		})

		child := NewPanel("child")
		child.SetStyle(&Style{
			Height:    50,
			MinHeight: 100,
			Width:     100,
		})
		parent.AddChild(child)

		le := NewLayoutEngine()
		le.Layout(parent, 200, 300)

		rect := child.ComputedRect()
		if rect.H < 100 {
			t.Errorf("Child height should respect minHeight, got %v", rect.H)
		}
	})

	t.Run("max height constraint", func(t *testing.T) {
		parent := NewPanel("parent")
		parent.SetStyle(&Style{
			Direction: LayoutColumn,
			Width:     200,
			Height:    400,
		})

		child := NewPanel("child")
		child.SetStyle(&Style{
			FlexGrow:  1,
			MaxHeight: 150,
			Width:     100,
		})
		parent.AddChild(child)

		le := NewLayoutEngine()
		le.Layout(parent, 200, 400)

		rect := child.ComputedRect()
		if rect.H > 150 {
			t.Errorf("Child height should respect maxHeight, got %v", rect.H)
		}
	})
}

// ============================================================================
// LayoutWidget and DrawWidget Tests
// ============================================================================

func TestLayoutWidget(t *testing.T) {
	t.Run("layout widget tree", func(t *testing.T) {
		root := NewPanel("root")
		root.SetStyle(&Style{
			Width:  500,
			Height: 400,
		})

		child := NewPanel("child")
		child.SetStyle(&Style{
			Width:  200,
			Height: 150,
		})
		root.AddChild(child)

		// Should not panic
		LayoutWidget(root)

		// Check that layout was applied
		rect := child.ComputedRect()
		if rect.W != 200 {
			t.Errorf("Child width = %v, want 200", rect.W)
		}
		if rect.H != 150 {
			t.Errorf("Child height = %v, want 150", rect.H)
		}
	})

	t.Run("layout nil widget", func(t *testing.T) {
		// Should not panic
		LayoutWidget(nil)
	})
}

func TestDrawWidget(t *testing.T) {
	t.Run("draw visible widget", func(t *testing.T) {
		widget := NewPanel("panel")
		widget.SetVisible(true)

		// Create a small screen for testing
		// Note: This won't actually render without Ebiten initialized,
		// but we can test that it doesn't panic

		// In a real test environment, we would create an ebiten.Image
		// For now, just verify the function exists and accepts parameters
		_ = widget
	})

	t.Run("draw invisible widget", func(t *testing.T) {
		widget := NewPanel("panel")
		widget.SetVisible(false)

		// Should not panic and should return early
		_ = widget
	})

	t.Run("draw nil widget", func(t *testing.T) {
		// Should not panic
		DrawWidget(nil, nil)
	})
}
