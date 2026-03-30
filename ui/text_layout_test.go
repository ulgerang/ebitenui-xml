package ui

import (
	"testing"

	"github.com/hajimehoshi/bitmapfont/v4"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

func testTextFace() text.Face {
	return text.NewGoXFace(bitmapfont.Face)
}

func TestGraphemeBoundaryHelpers(t *testing.T) {
	const sample = "A👨‍👩‍👧‍👦éB"

	if got := nextGraphemeBoundary(sample, 0); got != 1 {
		t.Fatalf("nextGraphemeBoundary(0) = %d, want 1", got)
	}
	if got := nextGraphemeBoundary(sample, 1); got != 8 {
		t.Fatalf("nextGraphemeBoundary(1) = %d, want 8", got)
	}
	if got := prevGraphemeBoundary(sample, 8); got != 1 {
		t.Fatalf("prevGraphemeBoundary(8) = %d, want 1", got)
	}
	if got := snapRuneIndexToBoundary(sample, 7); got != 1 {
		t.Fatalf("snapRuneIndexToBoundary(7) = %d, want 1", got)
	}
	if got := snapRuneIndexToBoundary(sample, 9); got != 8 {
		t.Fatalf("snapRuneIndexToBoundary(9) = %d, want 8", got)
	}
}

func TestTextHitTestUsesGraphemeClusters(t *testing.T) {
	txt := NewText("txt", "A👨‍👩‍👧‍👦B")
	txt.FontFace = testTextFace()
	txt.SetComputedRect(Rect{X: 10, Y: 20, W: 240, H: 40})

	layout := txt.ensureLayout(240, txt.Style())
	if layout == nil || len(layout.lines) != 1 {
		t.Fatalf("expected a single-line layout, got %#v", layout)
	}
	if len(layout.clusters) < 2 {
		t.Fatalf("expected grapheme clusters, got %d", len(layout.clusters))
	}

	emoji := layout.clusters[1]
	hit, ok := txt.HitTest(10+emoji.X+emoji.Width/2, 20+layout.lineHeight/2)
	if !ok {
		t.Fatal("HitTest returned no hit for emoji cluster")
	}
	if hit.Text != "👨‍👩‍👧‍👦" {
		t.Fatalf("HitTest text = %q, want emoji cluster", hit.Text)
	}
	if hit.RuneStart != emoji.RuneStart || hit.RuneEnd != emoji.RuneEnd {
		t.Fatalf("HitTest rune range = [%d:%d], want [%d:%d]", hit.RuneStart, hit.RuneEnd, emoji.RuneStart, emoji.RuneEnd)
	}
}

func TestTextInputPointerDownSnapsToClusterBoundaries(t *testing.T) {
	ti := NewTextInput("input")
	ti.FontFace = testTextFace()
	ti.SetText("A👨‍👩‍👧‍👦B")
	ti.SetComputedRect(Rect{X: 10, Y: 10, W: 300, H: 32})

	layout := newTextLayout(ti.Text, ti.FontFace, textLayoutOptions{
		Wrap:       false,
		WhiteSpace: textWhiteSpacePreWrap,
		LineHeight: measureLineHeight(ti.FontFace),
	})
	emoji := layout.clusters[1]

	ti.HandlePointerDown(10+emoji.X+emoji.Width*0.25, 20)
	if ti.CursorPos != emoji.RuneStart {
		t.Fatalf("CursorPos = %d, want %d at left half of emoji cluster", ti.CursorPos, emoji.RuneStart)
	}

	ti.HandlePointerDown(10+emoji.X+emoji.Width*0.75, 20)
	if ti.CursorPos != emoji.RuneEnd {
		t.Fatalf("CursorPos = %d, want %d at right half of emoji cluster", ti.CursorPos, emoji.RuneEnd)
	}
}

func TestTextAreaHorizontalMovementUsesGraphemeBoundaries(t *testing.T) {
	ta := NewTextArea("ta")
	ta.SetText("A👨‍👩‍👧‍👦B")

	ta.CursorPos = 1
	ta.moveCursorHorizontal(1)
	if ta.CursorPos != 8 {
		t.Fatalf("CursorPos after moving right = %d, want 8", ta.CursorPos)
	}

	ta.moveCursorHorizontal(-1)
	if ta.CursorPos != 1 {
		t.Fatalf("CursorPos after moving left = %d, want 1", ta.CursorPos)
	}
}
