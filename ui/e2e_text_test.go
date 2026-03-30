package ui

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func newTextE2EUI() *UI {
	ui := New(420, 240)
	ui.DefaultFontFace = testTextFace()

	root := NewPanel("root")
	root.SetStyle(&Style{
		Direction:  LayoutColumn,
		Width:      420,
		Height:     240,
		Padding:    PaddingAll(10),
		PaddingSet: true,
		Gap:        8,
		GapSet:     true,
	})

	txt := NewText("label", "A👨‍👩‍👧‍👦B")
	txt.SetStyle(&Style{Height: 32})

	input := NewTextInput("input")
	input.SetText("A👨‍👩‍👧‍👦B")
	input.SetStyle(&Style{Height: 32})

	area := NewTextArea("area")
	area.SetText("first\nA👨‍👩‍👧‍👦B")
	area.SetStyle(&Style{Height: 96})

	root.AddChild(txt)
	root.AddChild(input)
	root.AddChild(area)
	ui.SetRoot(root)
	return ui
}

func TestUISimulatePointerMoveTriggersClusterHover(t *testing.T) {
	ui := newTextE2EUI()
	txt := ui.GetText("label")
	if txt == nil {
		t.Fatal("expected text widget")
	}

	var hovered TextHit
	var hoverCalls int
	txt.OnClusterHover(func(hit TextHit) {
		hovered = hit
		hoverCalls++
	})

	layout := txt.ensureLayout(txt.ContentRect().W, txt.Style())
	if layout == nil || len(layout.clusters) < 2 {
		t.Fatal("expected grapheme-aware layout for text widget")
	}
	emoji := layout.clusters[1]
	line := layout.lines[0]
	x := txt.ContentRect().X + emoji.X + emoji.Width/2
	y := txt.textStartY(txt.ContentRect(), layout, txt.Style()) + layout.lineHeight/2

	ui.SimulatePointerMove(x, y)

	if hoverCalls != 1 {
		t.Fatalf("hover callback count = %d, want 1", hoverCalls)
	}
	if hovered.Text != "👨‍👩‍👧‍👦" {
		t.Fatalf("hovered text = %q, want emoji cluster", hovered.Text)
	}
	if txt.HoveredCluster != 1 {
		t.Fatalf("HoveredCluster = %d, want 1", txt.HoveredCluster)
	}
	if line.Text != "A👨‍👩‍👧‍👦B" {
		t.Fatalf("line text = %q, want full text", line.Text)
	}
}

func TestUISimulateClickAndDeleteUsesClusterBoundaries(t *testing.T) {
	ui := newTextE2EUI()
	input := ui.GetTextInput("input")
	if input == nil {
		t.Fatal("expected text input")
	}

	layout := newTextLayout(input.Text, input.FontFace, textLayoutOptions{
		Wrap:       false,
		WhiteSpace: textWhiteSpacePreWrap,
		LineHeight: measureLineHeight(input.FontFace),
	})
	emoji := layout.clusters[1]
	rect := input.ContentRect()

	ui.SimulateClick(rect.X+emoji.X+emoji.Width*0.25, rect.Y+rect.H/2)

	if ui.FocusedWidget() != input {
		t.Fatal("text input should be focused after click")
	}
	if input.CursorPos != emoji.RuneStart {
		t.Fatalf("CursorPos after click = %d, want %d", input.CursorPos, emoji.RuneStart)
	}

	ui.SimulateKeyPress(ebiten.KeyDelete, false, false)

	if input.Text != "AB" {
		t.Fatalf("input text after delete = %q, want %q", input.Text, "AB")
	}
	if input.CursorPos != 1 {
		t.Fatalf("CursorPos after delete = %d, want 1", input.CursorPos)
	}
}

func TestUISimulateTextAreaTypingAndBackspace(t *testing.T) {
	ui := newTextE2EUI()
	area := ui.GetTextArea("area")
	if area == nil {
		t.Fatal("expected text area")
	}

	rect := area.ContentRect()
	ui.SimulateClick(rect.X+4, rect.Y+4)

	if ui.FocusedWidget() != area {
		t.Fatal("text area should be focused after click")
	}

	ui.SimulateKeyPress(ebiten.KeyHome, false, false)
	ui.SimulateTypeText("🙂")
	if area.Text != "🙂first\nA👨‍👩‍👧‍👦B" {
		t.Fatalf("text area after typing = %q", area.Text)
	}

	ui.SimulateKeyPress(ebiten.KeyBackspace, false, false)
	if area.Text != "first\nA👨‍👩‍👧‍👦B" {
		t.Fatalf("text area after backspace = %q", area.Text)
	}
}
