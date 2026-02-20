package ui

import (
	"encoding/json"
	"testing"
)

// TestStyleMergeExplicitZero tests that explicitly-set zero values are respected
func TestStyleMergeExplicitZero(t *testing.T) {
	// Create base style with non-zero values
	base := &Style{
		Width:        100,
		Height:       50,
		FlexGrow:     1,
		Opacity:      1,
		FontSize:     16,
		BorderRadius: 5,
	}

	// Create override style with explicit zeros (from JSON)
	override := &Style{
		Width:           0,
		WidthSet:        true, // Explicitly set to 0
		Height:          0,
		HeightSet:       true, // Explicitly set to 0
		FlexGrow:        0,
		FlexGrowSet:     true, // Explicitly set to 0
		Opacity:         0,
		OpacitySet:      true, // Explicitly set to 0 (invisible)
		FontSize:        0,
		FontSizeSet:     true, // Explicitly set to 0
		BorderRadius:    0,
		BorderRadiusSet: true, // Explicitly set to 0
	}

	// Merge should apply explicit zeros
	base.Merge(override)

	// Verify zeros were applied
	if base.Width != 0 {
		t.Errorf("Width should be 0 (explicit), got %v", base.Width)
	}
	if base.Height != 0 {
		t.Errorf("Height should be 0 (explicit), got %v", base.Height)
	}
	if base.FlexGrow != 0 {
		t.Errorf("FlexGrow should be 0 (explicit), got %v", base.FlexGrow)
	}
	if base.Opacity != 0 {
		t.Errorf("Opacity should be 0 (explicit), got %v", base.Opacity)
	}
	if base.FontSize != 0 {
		t.Errorf("FontSize should be 0 (explicit), got %v", base.FontSize)
	}
	if base.BorderRadius != 0 {
		t.Errorf("BorderRadius should be 0 (explicit), got %v", base.BorderRadius)
	}
}

// TestStyleMergeOmittedValues tests that omitted values don't override
func TestStyleMergeOmittedValues(t *testing.T) {
	base := &Style{
		Width:    100,
		Height:   50,
		FlexGrow: 1,
	}

	// Create override style with only Width changed (Height not set)
	override := &Style{
		Width:    200,
		WidthSet: true,
		// Height not set - should remain 50
	}

	base.Merge(override)

	if base.Width != 200 {
		t.Errorf("Width should be 200, got %v", base.Width)
	}
	if base.Height != 50 {
		t.Errorf("Height should remain 50 (not overridden), got %v", base.Height)
	}
	if base.FlexGrow != 1 {
		t.Errorf("FlexGrow should remain 1 (not overridden), got %v", base.FlexGrow)
	}
}

// TestStyleEngineDetectExplicitFields tests JSON detection of explicit fields
func TestStyleEngineDetectExplicitFields(t *testing.T) {
	se := NewStyleEngine()

	// JSON with explicit zero values
	jsonData := []byte(`{
		"width": 0,
		"height": 0,
		"flexGrow": 0,
		"opacity": 0,
		"fontSize": 0,
		"borderRadius": 0,
		"gap": 0,
		"top": 0,
		"right": 0,
		"bottom": 0,
		"left": 0,
		"zIndex": 0,
		"minWidth": 0,
		"maxWidth": 0,
		"flexShrink": 0,
		"lineHeight": 0,
		"letterSpacing": 0,
		"outlineOffset": 0,
		"borderTopWidth": 0,
		"borderRightWidth": 0,
		"borderBottomWidth": 0,
		"borderLeftWidth": 0,
		"borderTopLeftRadius": 0,
		"borderTopRightRadius": 0,
		"borderBottomLeftRadius": 0,
		"borderBottomRightRadius": 0
	}`)

	var style Style
	if err := json.Unmarshal(jsonData, &style); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Detect explicit fields
	se.detectExplicitFields(&style, json.RawMessage(jsonData))

	// Verify all Set flags are true
	checks := []struct {
		name string
		set  bool
	}{
		{"WidthSet", style.WidthSet},
		{"HeightSet", style.HeightSet},
		{"FlexGrowSet", style.FlexGrowSet},
		{"OpacitySet", style.OpacitySet},
		{"FontSizeSet", style.FontSizeSet},
		{"BorderRadiusSet", style.BorderRadiusSet},
		{"GapSet", style.GapSet},
		{"TopSet", style.TopSet},
		{"RightSet", style.RightSet},
		{"BottomSet", style.BottomSet},
		{"LeftSet", style.LeftSet},
		{"ZIndexSet", style.ZIndexSet},
		{"MinWidthSet", style.MinWidthSet},
		{"MaxWidthSet", style.MaxWidthSet},
		{"FlexShrinkSet", style.FlexShrinkSet},
		{"LineHeightSet", style.LineHeightSet},
		{"LetterSpacingSet", style.LetterSpacingSet},
		{"OutlineOffsetSet", style.OutlineOffsetSet},
		{"BorderTopWidthSet", style.BorderTopWidthSet},
		{"BorderRightWidthSet", style.BorderRightWidthSet},
		{"BorderBottomWidthSet", style.BorderBottomWidthSet},
		{"BorderLeftWidthSet", style.BorderLeftWidthSet},
		{"BorderTopLeftRadiusSet", style.BorderTopLeftRadiusSet},
		{"BorderTopRightRadiusSet", style.BorderTopRightRadiusSet},
		{"BorderBottomLeftRadiusSet", style.BorderBottomLeftRadiusSet},
		{"BorderBottomRightRadiusSet", style.BorderBottomRightRadiusSet},
	}

	for _, check := range checks {
		if !check.set {
			t.Errorf("%s should be true (field present in JSON)", check.name)
		}
	}

	// Verify zero values are set
	if style.Width != 0 {
		t.Errorf("Width should be 0, got %v", style.Width)
	}
	if style.Height != 0 {
		t.Errorf("Height should be 0, got %v", style.Height)
	}
}

// TestStyleMergePositionProperties tests position property merging
func TestStyleMergePositionProperties(t *testing.T) {
	base := &Style{
		Top:    10,
		Right:  20,
		Bottom: 30,
		Left:   40,
		ZIndex: 5,
	}

	// Override with explicit zeros
	override := &Style{
		Top:       0,
		TopSet:    true,
		Right:     0,
		RightSet:  true,
		ZIndex:    0,
		ZIndexSet: true,
		// Bottom and Left not set - should remain unchanged
	}

	base.Merge(override)

	if base.Top != 0 {
		t.Errorf("Top should be 0 (explicit), got %v", base.Top)
	}
	if base.Right != 0 {
		t.Errorf("Right should be 0 (explicit), got %v", base.Right)
	}
	if base.ZIndex != 0 {
		t.Errorf("ZIndex should be 0 (explicit), got %v", base.ZIndex)
	}
	if base.Bottom != 30 {
		t.Errorf("Bottom should remain 30, got %v", base.Bottom)
	}
	if base.Left != 40 {
		t.Errorf("Left should remain 40, got %v", base.Left)
	}
}

// TestStyleClonePreservesSetFlags tests that Clone() preserves Set flags
func TestStyleClonePreservesSetFlags(t *testing.T) {
	original := &Style{
		Width:       100,
		WidthSet:    true,
		Height:      0,
		HeightSet:   true,
		FlexGrow:    0,
		FlexGrowSet: true,
		Opacity:     0.5,
		OpacitySet:  true,
	}

	cloned := original.Clone()

	if cloned.Width != original.Width || cloned.WidthSet != original.WidthSet {
		t.Errorf("Width/WidthSet not preserved: got %v/%v, want %v/%v",
			cloned.Width, cloned.WidthSet, original.Width, original.WidthSet)
	}
	if cloned.Height != original.Height || cloned.HeightSet != original.HeightSet {
		t.Errorf("Height/HeightSet not preserved: got %v/%v, want %v/%v",
			cloned.Height, cloned.HeightSet, original.Height, original.HeightSet)
	}
	if cloned.FlexGrow != original.FlexGrow || cloned.FlexGrowSet != original.FlexGrowSet {
		t.Errorf("FlexGrow/FlexGrowSet not preserved: got %v/%v, want %v/%v",
			cloned.FlexGrow, cloned.FlexGrowSet, original.FlexGrow, original.FlexGrowSet)
	}
	if cloned.Opacity != original.Opacity || cloned.OpacitySet != original.OpacitySet {
		t.Errorf("Opacity/OpacitySet not preserved: got %v/%v, want %v/%v",
			cloned.Opacity, cloned.OpacitySet, original.Opacity, original.OpacitySet)
	}
}
