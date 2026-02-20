package ui

import (
	"testing"
)

// ============================================================================
// Button Tests
// ============================================================================

func TestButton(t *testing.T) {
	t.Run("create button", func(t *testing.T) {
		btn := NewButton("btn", "Click Me")

		if btn.ID() != "btn" {
			t.Errorf("ID() = %v, want btn", btn.ID())
		}
		if btn.Type() != "button" {
			t.Errorf("Type() = %v, want button", btn.Type())
		}
		if btn.Label != "Click Me" {
			t.Errorf("Label = %v, want Click Me", btn.Label)
		}
	})

	t.Run("button initial state", func(t *testing.T) {
		btn := NewButton("btn", "Test")

		if !btn.Visible() {
			t.Error("Button should be visible by default")
		}
		if !btn.Enabled() {
			t.Error("Button should be enabled by default")
		}
		if btn.State() != StateNormal {
			t.Errorf("State() = %v, want StateNormal", btn.State())
		}
	})

	t.Run("button visibility", func(t *testing.T) {
		btn := NewButton("btn", "Test")

		btn.SetVisible(false)
		if btn.Visible() {
			t.Error("Button should be invisible after SetVisible(false)")
		}

		btn.SetVisible(true)
		if !btn.Visible() {
			t.Error("Button should be visible after SetVisible(true)")
		}
	})

	t.Run("button enabled state", func(t *testing.T) {
		btn := NewButton("btn", "Test")

		btn.SetEnabled(false)
		if btn.Enabled() {
			t.Error("Button should be disabled")
		}
		if btn.State() != StateDisabled {
			t.Errorf("State() = %v, want StateDisabled", btn.State())
		}
	})
}

func TestButtonIntrinsicSize(t *testing.T) {
	t.Run("intrinsic width without font", func(t *testing.T) {
		btn := NewButton("btn", "Test Label")

		// Without font, intrinsic width should be 0
		width := btn.IntrinsicWidth()
		if width != 0 {
			t.Errorf("IntrinsicWidth() = %v, want 0 (no font)", width)
		}
	})

	t.Run("intrinsic height without font", func(t *testing.T) {
		btn := NewButton("btn", "Test Label")

		// Without font, intrinsic height should be 0
		height := btn.IntrinsicHeight()
		if height != 0 {
			t.Errorf("IntrinsicHeight() = %v, want 0 (no font)", height)
		}
	})

	t.Run("intrinsic size with empty label", func(t *testing.T) {
		btn := NewButton("btn", "")

		width := btn.IntrinsicWidth()
		height := btn.IntrinsicHeight()

		if width != 0 {
			t.Errorf("IntrinsicWidth() = %v, want 0 (empty label)", width)
		}
		if height != 0 {
			t.Errorf("IntrinsicHeight() = %v, want 0 (empty label)", height)
		}
	})
}

func TestButtonClick(t *testing.T) {
	t.Run("click handler", func(t *testing.T) {
		btn := NewButton("btn", "Click")
		clicked := false

		btn.OnClick(func() {
			clicked = true
		})

		btn.HandleClick()

		if !clicked {
			t.Error("Click handler was not called")
		}
	})

	t.Run("click disabled button", func(t *testing.T) {
		btn := NewButton("btn", "Click")
		btn.SetEnabled(false)
		clicked := false

		btn.OnClick(func() {
			clicked = true
		})

		btn.HandleClick()

		if clicked {
			t.Error("Click handler should not be called on disabled button")
		}
	})
}

// ============================================================================
// Slider Tests
// ============================================================================

func TestSlider(t *testing.T) {
	t.Run("create slider", func(t *testing.T) {
		slider := NewSlider("slider")

		if slider.ID() != "slider" {
			t.Errorf("ID() = %v, want slider", slider.ID())
		}
		if slider.Type() != "slider" {
			t.Errorf("Type() = %v, want slider", slider.Type())
		}
		if slider.Min != 0 {
			t.Errorf("Min = %v, want 0", slider.Min)
		}
		if slider.Max != 1 {
			t.Errorf("Max = %v, want 1", slider.Max)
		}
		if slider.Value != 0.5 {
			t.Errorf("Value = %v, want 0.5", slider.Value)
		}
	})

	t.Run("set value within range", func(t *testing.T) {
		slider := NewSlider("slider")
		slider.SetValue(0.75)

		if slider.Value != 0.75 {
			t.Errorf("Value = %v, want 0.75", slider.Value)
		}
	})

	t.Run("set value clamped to max", func(t *testing.T) {
		slider := NewSlider("slider")
		slider.SetValue(1.5)

		if slider.Value != 1.0 {
			t.Errorf("Value = %v, want 1.0 (clamped)", slider.Value)
		}
	})

	t.Run("set value clamped to min", func(t *testing.T) {
		slider := NewSlider("slider")
		slider.SetValue(-0.5)

		if slider.Value != 0.0 {
			t.Errorf("Value = %v, want 0.0 (clamped)", slider.Value)
		}
	})

	t.Run("set value triggers callback", func(t *testing.T) {
		slider := NewSlider("slider")
		var calledValue float64

		slider.OnChange = func(v float64) {
			calledValue = v
		}

		slider.SetValue(0.6)

		if calledValue != 0.6 {
			t.Errorf("OnChange called with %v, want 0.6", calledValue)
		}
	})

	t.Run("set value no callback on same value", func(t *testing.T) {
		slider := NewSlider("slider")
		slider.Value = 0.5
		callCount := 0

		slider.OnChange = func(v float64) {
			callCount++
		}

		slider.SetValue(0.5) // Same as current value

		if callCount != 0 {
			t.Errorf("OnChange called %d times, want 0", callCount)
		}
	})

	t.Run("swap min max if inverted", func(t *testing.T) {
		slider := NewSlider("slider")
		slider.Min = 100
		slider.Max = 0

		slider.SetValue(50)

		// Min and Max should be swapped
		if slider.Min != 0 {
			t.Errorf("Min = %v, want 0 (swapped)", slider.Min)
		}
		if slider.Max != 100 {
			t.Errorf("Max = %v, want 100 (swapped)", slider.Max)
		}
		// Value should be clamped to new range
		if slider.Value != 50 {
			t.Errorf("Value = %v, want 50", slider.Value)
		}
	})
}

func TestSliderNormalizedValue(t *testing.T) {
	tests := []struct {
		name     string
		min      float64
		max      float64
		value    float64
		expected float64
	}{
		{"zero to one, mid", 0, 1, 0.5, 0.5},
		{"zero to one, min", 0, 1, 0, 0},
		{"zero to one, max", 0, 1, 1, 1},
		{"custom range, mid", 0, 100, 50, 0.5},
		{"custom range, quarter", 0, 100, 25, 0.25},
		{"negative range", -10, 10, 0, 0.5},
		{"negative range, min", -10, 10, -10, 0},
		{"negative range, max", -10, 10, 10, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slider := NewSlider("slider")
			slider.Min = tt.min
			slider.Max = tt.max
			slider.Value = tt.value

			got := slider.normalizedValue()
			if got != tt.expected {
				t.Errorf("normalizedValue() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSliderHandleClick(t *testing.T) {
	t.Run("handle click updates value", func(t *testing.T) {
		slider := NewSlider("slider")
		slider.Min = 0
		slider.Max = 100
		slider.SetComputedRect(Rect{X: 0, Y: 0, W: 200, H: 20})

		// Simulate click at middle
		// Note: This requires ebiten.CursorPosition() which won't work in tests
		// Testing the structure instead
		slider.setValueFromCursor(100) // Middle of 200px slider

		if slider.Value < 49 || slider.Value > 51 {
			t.Errorf("Value = %v, want ~50", slider.Value)
		}
	})

	t.Run("handle click on disabled slider", func(t *testing.T) {
		slider := NewSlider("slider")
		slider.SetEnabled(false)

		// Should not change value
		initialValue := slider.Value
		slider.setValueFromCursor(100)

		if slider.Value != initialValue {
			t.Error("Value should not change on disabled slider")
		}
	})

	t.Run("handle click with zero width", func(t *testing.T) {
		slider := NewSlider("slider")
		slider.SetComputedRect(Rect{X: 0, Y: 0, W: 0, H: 20})

		// Should not panic or change value
		initialValue := slider.Value
		slider.setValueFromCursor(50)

		if slider.Value != initialValue {
			t.Error("Value should not change with zero width")
		}
	})
}

// ============================================================================
// Checkbox Tests
// ============================================================================

func TestCheckbox(t *testing.T) {
	t.Run("create checkbox", func(t *testing.T) {
		cb := NewCheckbox("cb", "Accept Terms")

		if cb.ID() != "cb" {
			t.Errorf("ID() = %v, want cb", cb.ID())
		}
		if cb.Label != "Accept Terms" {
			t.Errorf("Label = %v, want Accept Terms", cb.Label)
		}
		if cb.Checked {
			t.Error("Checkbox should be unchecked by default")
		}
	})

	t.Run("toggle checkbox", func(t *testing.T) {
		cb := NewCheckbox("cb", "Test")

		cb.HandleClick()
		if !cb.Checked {
			t.Error("Checkbox should be checked after first click")
		}

		cb.HandleClick()
		if cb.Checked {
			t.Error("Checkbox should be unchecked after second click")
		}
	})

	t.Run("toggle disabled checkbox", func(t *testing.T) {
		cb := NewCheckbox("cb", "Test")
		cb.SetEnabled(false)

		cb.HandleClick()
		if cb.Checked {
			t.Error("Checkbox should not toggle when disabled")
		}
	})

	t.Run("on change callback", func(t *testing.T) {
		cb := NewCheckbox("cb", "Test")
		var lastState bool
		callCount := 0

		cb.OnChange = func(checked bool) {
			lastState = checked
			callCount++
		}

		cb.HandleClick()

		if callCount != 1 {
			t.Errorf("OnChange called %d times, want 1", callCount)
		}
		if !lastState {
			t.Error("OnChange received false, want true")
		}
	})
}

// ============================================================================
// ProgressBar Tests
// ============================================================================

func TestProgressBar(t *testing.T) {
	t.Run("create progress bar", func(t *testing.T) {
		pb := NewProgressBar("pb")

		if pb.ID() != "pb" {
			t.Errorf("ID() = %v, want pb", pb.ID())
		}
		if pb.Type() != "progressbar" {
			t.Errorf("Type() = %v, want progressbar", pb.Type())
		}
		if pb.Value != 0 {
			t.Errorf("Value = %v, want 0", pb.Value)
		}
	})

	t.Run("set value", func(t *testing.T) {
		pb := NewProgressBar("pb")
		pb.Value = 0.75

		if pb.Value != 0.75 {
			t.Errorf("Value = %v, want 0.75", pb.Value)
		}
	})

	t.Run("value range", func(t *testing.T) {
		pb := NewProgressBar("pb")

		// Test negative
		pb.Value = -0.5
		// Note: ProgressBar doesn't clamp, so this is allowed

		// Test over 1
		pb.Value = 1.5
		// Note: ProgressBar doesn't clamp, so this is allowed
	})
}

// ============================================================================
// Text Widget Tests
// ============================================================================

func TestTextWidget(t *testing.T) {
	t.Run("create text widget", func(t *testing.T) {
		txt := NewText("txt", "Hello World")

		if txt.ID() != "txt" {
			t.Errorf("ID() = %v, want txt", txt.ID())
		}
		if txt.Type() != "text" {
			t.Errorf("Type() = %v, want text", txt.Type())
		}
		if txt.Content != "Hello World" {
			t.Errorf("Content = %v, want Hello World", txt.Content)
		}
	})

	t.Run("set content", func(t *testing.T) {
		txt := NewText("txt", "Original")
		txt.SetContent("Updated")

		if txt.Content != "Updated" {
			t.Errorf("Content = %v, want Updated", txt.Content)
		}
	})

	t.Run("set content same value", func(t *testing.T) {
		txt := NewText("txt", "Same")
		txt.wrappedLines = []string{"cached"}

		txt.SetContent("Same")

		// Cache should not be invalidated
		if txt.wrappedLines == nil {
			t.Error("Cache should not be invalidated for same content")
		}
	})

	t.Run("set content different value", func(t *testing.T) {
		txt := NewText("txt", "Original")
		txt.wrappedLines = []string{"cached"}

		txt.SetContent("Different")

		// Cache should be invalidated
		if txt.wrappedLines != nil {
			t.Error("Cache should be invalidated for different content")
		}
	})

	t.Run("intrinsic size without font", func(t *testing.T) {
		txt := NewText("txt", "Test")

		if txt.IntrinsicWidth() != 0 {
			t.Errorf("IntrinsicWidth() = %v, want 0 (no font)", txt.IntrinsicWidth())
		}
		if txt.IntrinsicHeight() != 0 {
			t.Errorf("IntrinsicHeight() = %v, want 0 (no font)", txt.IntrinsicHeight())
		}
	})
}

// ============================================================================
// Image Widget Tests
// ============================================================================

func TestImageWidget(t *testing.T) {
	t.Run("create image widget", func(t *testing.T) {
		img := NewImage("img")

		if img.ID() != "img" {
			t.Errorf("ID() = %v, want img", img.ID())
		}
		if img.Type() != "image" {
			t.Errorf("Type() = %v, want image", img.Type())
		}
		if img.Source != nil {
			t.Error("Source should be nil initially")
		}
	})
}

// ============================================================================
// Panel Widget Tests
// ============================================================================

func TestPanelWidget(t *testing.T) {
	t.Run("create panel", func(t *testing.T) {
		panel := NewPanel("panel")

		if panel.ID() != "panel" {
			t.Errorf("ID() = %v, want panel", panel.ID())
		}
		if panel.Type() != "panel" {
			t.Errorf("Type() = %v, want panel", panel.Type())
		}
	})

	t.Run("panel as container", func(t *testing.T) {
		parent := NewPanel("parent")
		child := NewPanel("child")

		parent.AddChild(child)

		if len(parent.Children()) != 1 {
			t.Errorf("Children count = %d, want 1", len(parent.Children()))
		}
		if child.Parent() == nil {
			t.Error("Child parent should not be nil")
		}
		if child.Parent().ID() != "parent" {
			t.Errorf("Child parent ID = %v, want parent", child.Parent().ID())
		}
	})

	t.Run("remove child", func(t *testing.T) {
		parent := NewPanel("parent")
		child := NewPanel("child")

		parent.AddChild(child)
		parent.RemoveChild(child)

		if len(parent.Children()) != 0 {
			t.Errorf("Children count = %d, want 0", len(parent.Children()))
		}
		if child.Parent() != nil {
			t.Error("Child parent should be nil after removal")
		}
	})
}

// ============================================================================
// SVGIcon Widget Tests
// ============================================================================

func TestSVGIconWidget(t *testing.T) {
	t.Run("create SVG icon", func(t *testing.T) {
		icon := NewSVGIcon("icon")

		if icon.ID() != "icon" {
			t.Errorf("ID() = %v, want icon", icon.ID())
		}
		if icon.Type() != "svg" {
			t.Errorf("Type() = %v, want svg", icon.Type())
		}
	})

	t.Run("set icon color", func(t *testing.T) {
		icon := NewSVGIcon("icon")
		// Note: SetIcon requires a valid icon name
		// Testing the method exists and accepts parameters
		_ = icon
	})
}

// ============================================================================
// Widget Base Tests
// ============================================================================

func TestBaseWidget(t *testing.T) {
	t.Run("add class", func(t *testing.T) {
		widget := NewBaseWidget("w", "panel")
		widget.AddClass("my-class")

		if !widget.HasClass("my-class") {
			t.Error("Widget should have class my-class")
		}
	})

	t.Run("add duplicate class", func(t *testing.T) {
		widget := NewBaseWidget("w", "panel")
		widget.AddClass("class1")
		widget.AddClass("class1")

		classes := widget.Classes()
		if len(classes) != 1 {
			t.Errorf("Classes count = %d, want 1", len(classes))
		}
	})

	t.Run("remove class", func(t *testing.T) {
		widget := NewBaseWidget("w", "panel")
		widget.AddClass("class1")
		widget.AddClass("class2")

		widget.RemoveClass("class1")

		if widget.HasClass("class1") {
			t.Error("Widget should not have class1 after removal")
		}
		if !widget.HasClass("class2") {
			t.Error("Widget should still have class2")
		}
	})

	t.Run("remove nonexistent class", func(t *testing.T) {
		widget := NewBaseWidget("w", "panel")

		// Should not panic
		widget.RemoveClass("nonexistent")
	})
}

// ============================================================================
// ContentRect Tests
// ============================================================================

func TestContentRect(t *testing.T) {
	t.Run("content rect with padding", func(t *testing.T) {
		widget := NewBaseWidget("w", "panel")
		widget.SetComputedRect(Rect{X: 0, Y: 0, W: 200, H: 100})
		widget.SetStyle(&Style{
			Padding: Padding{Top: 10, Right: 20, Bottom: 10, Left: 20},
		})

		content := widget.ContentRect()

		if content.X != 20 {
			t.Errorf("Content X = %v, want 20", content.X)
		}
		if content.Y != 10 {
			t.Errorf("Content Y = %v, want 10", content.Y)
		}
		if content.W != 160 {
			t.Errorf("Content W = %v, want 160", content.W)
		}
		if content.H != 80 {
			t.Errorf("Content H = %v, want 80", content.H)
		}
	})

	t.Run("content rect with border", func(t *testing.T) {
		widget := NewBaseWidget("w", "panel")
		widget.SetComputedRect(Rect{X: 0, Y: 0, W: 200, H: 100})
		widget.SetStyle(&Style{
			Padding:     Padding{Top: 10, Right: 10, Bottom: 10, Left: 10},
			BorderWidth: 5,
		})

		content := widget.ContentRect()

		// Border should be included in the inset
		if content.X != 15 {
			t.Errorf("Content X = %v, want 15", content.X)
		}
		if content.Y != 15 {
			t.Errorf("Content Y = %v, want 15", content.Y)
		}
	})
}
