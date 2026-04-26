package ui

import "testing"

func TestRelativeSizeValueResolution(t *testing.T) {
	ctx := SizeContext{
		ParentSize:     200,
		ViewportWidth:  800,
		ViewportHeight: 600,
		FontSize:       20,
		RootFontSize:   16,
	}

	tests := []struct {
		name string
		raw  string
		want float64
		unit SizeUnit
	}{
		{name: "percent", raw: "50%", want: 100, unit: UnitPercent},
		{name: "viewport width", raw: "25vw", want: 200, unit: UnitVw},
		{name: "viewport height", raw: "10vh", want: 60, unit: UnitVh},
		{name: "em", raw: "1.5em", want: 30, unit: UnitEm},
		{name: "rem", raw: "2rem", want: 32, unit: UnitRem},
		{name: "px", raw: "12px", want: 12, unit: UnitPx},
		{name: "unitless", raw: "7", want: 7, unit: UnitPx},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseSizeValue(tt.raw)
			if got.Unit != tt.unit {
				t.Fatalf("unit = %v, want %v", got.Unit, tt.unit)
			}
			if resolved := got.Resolve(ctx); resolved != tt.want {
				t.Fatalf("Resolve(%q) = %v, want %v", tt.raw, resolved, tt.want)
			}
		})
	}
}

func TestCalcExpressionResolution(t *testing.T) {
	ctx := SizeContext{
		ParentSize:     200,
		ViewportWidth:  800,
		ViewportHeight: 600,
		FontSize:       20,
		RootFontSize:   16,
	}

	tests := []struct {
		name string
		raw  string
		want float64
	}{
		{name: "mixed percent px", raw: "calc(50% - 10px)", want: 90},
		{name: "viewport plus rem", raw: "calc(10vw + 2rem)", want: 112},
		{name: "em arithmetic", raw: "calc(2em + 4px)", want: 44},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := ParseCalc(tt.raw)
			if expr == nil {
				t.Fatalf("ParseCalc(%q) returned nil", tt.raw)
			}
			if got := expr.Resolve(ctx); got != tt.want {
				t.Fatalf("Resolve(%q) = %v, want %v", tt.raw, got, tt.want)
			}
		})
	}
}

func TestProductionStyleLoadersRelativeUnitAuditBoundary(t *testing.T) {
	t.Run("css declarations remain pixel parser based", func(t *testing.T) {
		se := NewStyleEngine()
		if err := se.LoadCSS(`.card { width: 50%; height: 25vw; gap: calc(10px + 5px); font-size: 2rem; }`); err != nil {
			t.Fatalf("LoadCSS failed: %v", err)
		}

		style := se.GetStyle(".card")
		if style == nil {
			t.Fatal("missing .card style")
		}
		if style.Width != 0 {
			t.Fatalf("CSS percent width = %v, want 0 because LoadCSS does not resolve relative units", style.Width)
		}
		if style.Height != 0 {
			t.Fatalf("CSS viewport height = %v, want 0 because LoadCSS does not resolve relative units", style.Height)
		}
		if style.Gap != 0 {
			t.Fatalf("CSS calc gap = %v, want 0 because LoadCSS does not resolve calc()", style.Gap)
		}
		if style.FontSize != 0 {
			t.Fatalf("CSS rem font-size = %v, want 0 because LoadCSS does not resolve relative units", style.FontSize)
		}
	})

	t.Run("xml inline percent is numeric suffix stripping not parent resolution", func(t *testing.T) {
		u := New(400, 300)
		if err := u.LoadLayout(`<panel id="root"><panel id="child" width="50%" height="25%" /></panel>`); err != nil {
			t.Fatalf("LoadLayout failed: %v", err)
		}

		child := u.GetWidget("child")
		if child == nil {
			t.Fatal("missing child widget")
		}
		style := child.Style()
		if style.Width != 50 || style.Height != 25 {
			t.Fatalf("inline percent style = %vx%v, want raw numeric 50x25 audit boundary", style.Width, style.Height)
		}
	})
}
