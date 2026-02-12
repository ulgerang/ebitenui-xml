package ui

import (
	"image/color"
	"strings"
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// ============================================================================
// TextInput Widget - Single-line text input
// ============================================================================

// TextInput is a single-line text input widget
type TextInput struct {
	*BaseWidget

	// Content
	Text        string
	Placeholder string

	// State
	Focused     bool
	CursorPos   int
	SelectStart int
	SelectEnd   int

	// Visual
	FontFace         text.Face
	PlaceholderColor color.Color
	CursorColor      color.Color
	SelectionColor   color.Color

	// Behavior
	MaxLength int
	ReadOnly  bool
	Password  bool // Mask characters

	// Events
	OnChange func(text string)
	OnSubmit func(text string) // Enter key
	OnFocus  func()
	OnBlur   func()

	// Internal
	cursorBlink     float64
	cursorVisible   bool
	scrollOffset    float64
	repeatKey       ebiten.Key
	repeatStartTime float64
	repeatNextTime  float64
}

// NewTextInput creates a new text input widget
func NewTextInput(id string) *TextInput {
	return &TextInput{
		BaseWidget:       NewBaseWidget(id, "input"),
		PlaceholderColor: color.RGBA{128, 128, 128, 255},
		CursorColor:      color.White,
		SelectionColor:   color.RGBA{100, 149, 237, 128},
	}
}

// SetText sets the text content
func (ti *TextInput) SetText(s string) {
	if ti.MaxLength > 0 && utf8.RuneCountInString(s) > ti.MaxLength {
		s = string([]rune(s)[:ti.MaxLength])
	}
	ti.Text = s
	ti.clampIndices()
}

func (ti *TextInput) clampIndices() {
	runeLen := utf8.RuneCountInString(ti.Text)

	if ti.CursorPos < 0 {
		ti.CursorPos = 0
	} else if ti.CursorPos > runeLen {
		ti.CursorPos = runeLen
	}

	if ti.SelectStart < 0 {
		ti.SelectStart = 0
	} else if ti.SelectStart > runeLen {
		ti.SelectStart = runeLen
	}

	if ti.SelectEnd < 0 {
		ti.SelectEnd = 0
	} else if ti.SelectEnd > runeLen {
		ti.SelectEnd = runeLen
	}
}

// Focus gives focus to this input
func (ti *TextInput) Focus() {
	ti.Focused = true
	ti.state = StateFocused
	if ti.OnFocus != nil {
		ti.OnFocus()
	}
}

// Blur removes focus from this input
func (ti *TextInput) Blur() {
	ti.Focused = false
	ti.state = StateNormal
	ti.SelectStart = 0
	ti.SelectEnd = 0
	if ti.OnBlur != nil {
		ti.OnBlur()
	}
}

// HandleInput processes keyboard input
func (ti *TextInput) HandleInput() {
	if !ti.Focused || ti.ReadOnly {
		return
	}

	// Handle text input
	inputChars := ebiten.AppendInputChars(nil)
	for _, char := range inputChars {
		ti.insertChar(char)
	}

	// Handle key presses
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		ti.handleBackspace()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDelete) {
		ti.handleDelete()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		ti.moveCursor(-1, ebiten.IsKeyPressed(ebiten.KeyShift))
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		ti.moveCursor(1, ebiten.IsKeyPressed(ebiten.KeyShift))
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyHome) {
		ti.CursorPos = 0
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnd) {
		ti.CursorPos = utf8.RuneCountInString(ti.Text)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter) {
		if ti.OnSubmit != nil {
			ti.OnSubmit(ti.Text)
		}
	}

	// Ctrl+A: Select all
	if ebiten.IsKeyPressed(ebiten.KeyControl) && inpututil.IsKeyJustPressed(ebiten.KeyA) {
		ti.SelectStart = 0
		ti.SelectEnd = utf8.RuneCountInString(ti.Text)
		ti.CursorPos = ti.SelectEnd
	}

	// Ctrl+C: Copy (placeholder - clipboard requires platform-specific code)
	// Ctrl+V: Paste (placeholder)
	// Ctrl+X: Cut (placeholder)

	// Update cursor blink
	ti.cursorBlink += 1.0 / 60.0
	if ti.cursorBlink >= 0.5 {
		ti.cursorBlink = 0
		ti.cursorVisible = !ti.cursorVisible
	}
}

func (ti *TextInput) insertChar(char rune) {
	if ti.MaxLength > 0 && utf8.RuneCountInString(ti.Text) >= ti.MaxLength {
		return
	}

	ti.clampIndices()

	// Delete selection if any
	if ti.SelectStart != ti.SelectEnd {
		ti.deleteSelection()
	}

	runes := []rune(ti.Text)
	newRunes := make([]rune, 0, len(runes)+1)
	newRunes = append(newRunes, runes[:ti.CursorPos]...)
	newRunes = append(newRunes, char)
	newRunes = append(newRunes, runes[ti.CursorPos:]...)
	ti.Text = string(newRunes)
	ti.CursorPos++

	if ti.OnChange != nil {
		ti.OnChange(ti.Text)
	}
}

func (ti *TextInput) handleBackspace() {
	ti.clampIndices()

	if ti.SelectStart != ti.SelectEnd {
		ti.deleteSelection()
		return
	}

	if ti.CursorPos > 0 {
		runes := []rune(ti.Text)
		ti.Text = string(append(runes[:ti.CursorPos-1], runes[ti.CursorPos:]...))
		ti.CursorPos--

		if ti.OnChange != nil {
			ti.OnChange(ti.Text)
		}
	}
}

func (ti *TextInput) handleDelete() {
	ti.clampIndices()

	if ti.SelectStart != ti.SelectEnd {
		ti.deleteSelection()
		return
	}

	runes := []rune(ti.Text)
	if ti.CursorPos < len(runes) {
		ti.Text = string(append(runes[:ti.CursorPos], runes[ti.CursorPos+1:]...))

		if ti.OnChange != nil {
			ti.OnChange(ti.Text)
		}
	}
}

func (ti *TextInput) deleteSelection() {
	ti.clampIndices()

	if ti.SelectStart == ti.SelectEnd {
		return
	}

	start, end := ti.SelectStart, ti.SelectEnd
	if start > end {
		start, end = end, start
	}

	runes := []rune(ti.Text)
	if start > len(runes) {
		start = len(runes)
	}
	if end > len(runes) {
		end = len(runes)
	}
	if start == end {
		return
	}
	ti.Text = string(append(runes[:start], runes[end:]...))
	ti.CursorPos = start
	ti.SelectStart = 0
	ti.SelectEnd = 0

	if ti.OnChange != nil {
		ti.OnChange(ti.Text)
	}
}

func (ti *TextInput) moveCursor(delta int, selecting bool) {
	ti.clampIndices()

	newPos := ti.CursorPos + delta
	if newPos < 0 {
		newPos = 0
	}
	runes := []rune(ti.Text)
	if newPos > len(runes) {
		newPos = len(runes)
	}

	if selecting {
		if ti.SelectStart == ti.SelectEnd {
			ti.SelectStart = ti.CursorPos
		}
		ti.SelectEnd = newPos
	} else {
		ti.SelectStart = 0
		ti.SelectEnd = 0
	}

	ti.CursorPos = newPos
	ti.cursorBlink = 0
	ti.cursorVisible = true
}

// Draw renders the text input
func (ti *TextInput) Draw(screen *ebiten.Image) {
	if !ti.visible {
		return
	}

	// Draw base
	ti.BaseWidget.Draw(screen)

	style := ti.getActiveStyle()
	r := ti.ContentRect()

	if ti.FontFace == nil {
		return
	}

	// Determine display text
	displayText := ti.Text
	if ti.Password {
		displayText = strings.Repeat("●", utf8.RuneCountInString(ti.Text))
	}
	displayRunes := []rune(displayText)
	ti.clampIndices()

	// Draw Placeholder if empty
	if displayText == "" && ti.Placeholder != "" && !ti.Focused {
		metrics := ti.FontFace.Metrics()
		emHeight := metrics.HAscent + metrics.HDescent
		y := r.Y + (r.H-emHeight)/2
		op := &text.DrawOptions{}
		op.GeoM.Translate(r.X, y)
		op.ColorScale.ScaleWithColor(ti.PlaceholderColor)
		text.Draw(screen, ti.Placeholder, ti.FontFace, op)
		return
	}

	// Draw selection highlight
	if ti.Focused && ti.SelectStart != ti.SelectEnd {
		start, end := ti.SelectStart, ti.SelectEnd
		if start > end {
			start, end = end, start
		}
		if start > len(displayRunes) {
			start = len(displayRunes)
		}
		if end > len(displayRunes) {
			end = len(displayRunes)
		}

		startX := ti.measureTextWidth(string(displayRunes[:start]))
		endX := ti.measureTextWidth(string(displayRunes[:end]))

		selRect := Rect{
			X: r.X + startX - ti.scrollOffset,
			Y: r.Y + 2,
			W: endX - startX,
			H: r.H - 4,
		}
		DrawRoundedRectPath(screen, selRect, 2, ti.SelectionColor)
	}

	// Draw text
	textColor := style.TextColor
	if textColor == nil {
		textColor = color.White
	}

	metrics := ti.FontFace.Metrics()
	emHeight := metrics.HAscent + metrics.HDescent
	y := r.Y + (r.H-emHeight)/2

	op := &text.DrawOptions{}
	op.GeoM.Translate(r.X-ti.scrollOffset, y)
	op.ColorScale.ScaleWithColor(textColor)
	text.Draw(screen, displayText, ti.FontFace, op)

	// Draw cursor
	if ti.Focused && ti.cursorVisible {
		cursorPos := ti.CursorPos
		if cursorPos > len(displayRunes) {
			cursorPos = len(displayRunes)
		}
		cursorX := r.X + ti.measureTextWidth(string(displayRunes[:cursorPos])) - ti.scrollOffset

		DrawRoundedRectPath(screen, Rect{
			X: cursorX,
			Y: r.Y + 4,
			W: 2,
			H: r.H - 8,
		}, 1, ti.CursorColor)
	}
}

func (ti *TextInput) measureTextWidth(s string) float64 {
	if ti.FontFace == nil {
		return 0
	}
	w, _ := text.Measure(s, ti.FontFace, 0)
	return w
}

// ============================================================================
// TextArea Widget - Multi-line text input
// ============================================================================

// TextArea is a multi-line text input widget
type TextArea struct {
	*BaseWidget

	Text        string
	Placeholder string

	// State
	Focused    bool
	CursorPos  int
	CursorLine int
	CursorCol  int

	// Visual
	FontFace         text.Face
	PlaceholderColor color.Color
	CursorColor      color.Color

	// Behavior
	MaxLength int
	ReadOnly  bool

	// Events
	OnChange func(text string)

	// Scroll
	ScrollY    float64
	MaxScrollY float64

	// Internal
	lines         []string
	cursorBlink   float64
	cursorVisible bool
}

// NewTextArea creates a new text area widget
func NewTextArea(id string) *TextArea {
	return &TextArea{
		BaseWidget:       NewBaseWidget(id, "textarea"),
		PlaceholderColor: color.RGBA{128, 128, 128, 255},
		CursorColor:      color.White,
		lines:            []string{""},
	}
}

// SetText sets the text content
func (ta *TextArea) SetText(s string) {
	if ta.MaxLength > 0 && utf8.RuneCountInString(s) > ta.MaxLength {
		s = string([]rune(s)[:ta.MaxLength])
	}
	ta.Text = s
	ta.clampCursorPos()
	ta.updateLines()
	ta.updateCursorLineCol()
}

func (ta *TextArea) clampCursorPos() {
	runeLen := utf8.RuneCountInString(ta.Text)
	if ta.CursorPos < 0 {
		ta.CursorPos = 0
	} else if ta.CursorPos > runeLen {
		ta.CursorPos = runeLen
	}
}

func (ta *TextArea) updateLines() {
	ta.lines = splitLines(ta.Text)
	if len(ta.lines) == 0 {
		ta.lines = []string{""}
	}
}

func splitLines(s string) []string {
	if s == "" {
		return []string{""}
	}
	lines := make([]string, 0)
	current := ""
	for _, r := range s {
		if r == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(r)
		}
	}
	lines = append(lines, current)
	return lines
}

// Focus gives focus to this text area
func (ta *TextArea) Focus() {
	ta.Focused = true
	ta.state = StateFocused
}

// Blur removes focus
func (ta *TextArea) Blur() {
	ta.Focused = false
	ta.state = StateNormal
}

// HandleInput processes keyboard input
func (ta *TextArea) HandleInput() {
	if !ta.Focused || ta.ReadOnly {
		return
	}

	// Handle text input
	inputChars := ebiten.AppendInputChars(nil)
	for _, char := range inputChars {
		ta.insertChar(char)
	}

	// Handle key presses
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		ta.handleBackspace()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDelete) {
		ta.handleDelete()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter) {
		ta.insertChar('\n')
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		ta.moveCursorHorizontal(-1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		ta.moveCursorHorizontal(1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		ta.moveCursorVertical(-1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		ta.moveCursorVertical(1)
	}

	// Update cursor blink
	ta.cursorBlink += 1.0 / 60.0
	if ta.cursorBlink >= 0.5 {
		ta.cursorBlink = 0
		ta.cursorVisible = !ta.cursorVisible
	}
}

func (ta *TextArea) insertChar(char rune) {
	if ta.MaxLength > 0 && utf8.RuneCountInString(ta.Text) >= ta.MaxLength {
		return
	}

	ta.clampCursorPos()

	runes := []rune(ta.Text)
	newRunes := make([]rune, 0, len(runes)+1)
	newRunes = append(newRunes, runes[:ta.CursorPos]...)
	newRunes = append(newRunes, char)
	newRunes = append(newRunes, runes[ta.CursorPos:]...)
	ta.Text = string(newRunes)
	ta.CursorPos++
	ta.updateLines()
	ta.updateCursorLineCol()

	if ta.OnChange != nil {
		ta.OnChange(ta.Text)
	}
}

func (ta *TextArea) handleBackspace() {
	ta.clampCursorPos()

	if ta.CursorPos > 0 {
		runes := []rune(ta.Text)
		ta.Text = string(append(runes[:ta.CursorPos-1], runes[ta.CursorPos:]...))
		ta.CursorPos--
		ta.updateLines()
		ta.updateCursorLineCol()

		if ta.OnChange != nil {
			ta.OnChange(ta.Text)
		}
	}
}

func (ta *TextArea) handleDelete() {
	ta.clampCursorPos()

	runes := []rune(ta.Text)
	if ta.CursorPos < len(runes) {
		ta.Text = string(append(runes[:ta.CursorPos], runes[ta.CursorPos+1:]...))
		ta.updateLines()
		ta.updateCursorLineCol()

		if ta.OnChange != nil {
			ta.OnChange(ta.Text)
		}
	}
}

func (ta *TextArea) moveCursorHorizontal(delta int) {
	ta.clampCursorPos()

	newPos := ta.CursorPos + delta
	if newPos < 0 {
		newPos = 0
	}
	runes := []rune(ta.Text)
	if newPos > len(runes) {
		newPos = len(runes)
	}
	ta.CursorPos = newPos
	ta.updateCursorLineCol()
	ta.cursorBlink = 0
	ta.cursorVisible = true
}

func (ta *TextArea) moveCursorVertical(delta int) {
	newLine := ta.CursorLine + delta
	if newLine < 0 {
		newLine = 0
	}
	if newLine >= len(ta.lines) {
		newLine = len(ta.lines) - 1
	}

	ta.CursorLine = newLine
	lineRuneLen := utf8.RuneCountInString(ta.lines[newLine])
	if ta.CursorCol > lineRuneLen {
		ta.CursorCol = lineRuneLen
	}

	ta.updateCursorPosFromLineCol()
	ta.cursorBlink = 0
	ta.cursorVisible = true
}

func (ta *TextArea) updateCursorLineCol() {
	ta.clampCursorPos()

	pos := 0
	for i, line := range ta.lines {
		lineLen := len([]rune(line))
		if ta.CursorPos <= pos+lineLen {
			ta.CursorLine = i
			ta.CursorCol = ta.CursorPos - pos
			return
		}
		pos += lineLen + 1 // +1 for newline
	}
	ta.CursorLine = len(ta.lines) - 1
	ta.CursorCol = len([]rune(ta.lines[ta.CursorLine]))
}

func (ta *TextArea) updateCursorPosFromLineCol() {
	pos := 0
	for i := 0; i < ta.CursorLine && i < len(ta.lines); i++ {
		pos += len([]rune(ta.lines[i])) + 1
	}
	if ta.CursorLine >= 0 && ta.CursorLine < len(ta.lines) {
		maxCol := len([]rune(ta.lines[ta.CursorLine]))
		if ta.CursorCol > maxCol {
			ta.CursorCol = maxCol
		}
	}
	pos += ta.CursorCol
	ta.CursorPos = pos
	ta.clampCursorPos()
}

// Draw renders the text area
func (ta *TextArea) Draw(screen *ebiten.Image) {
	if !ta.visible {
		return
	}

	ta.BaseWidget.Draw(screen)

	style := ta.getActiveStyle()
	r := ta.ContentRect()

	if ta.FontFace == nil {
		return
	}

	// Draw placeholder if empty
	if ta.Text == "" && ta.Placeholder != "" && !ta.Focused {
		op := &text.DrawOptions{}
		op.GeoM.Translate(r.X, r.Y)
		op.ColorScale.ScaleWithColor(ta.PlaceholderColor)
		text.Draw(screen, ta.Placeholder, ta.FontFace, op)
		return
	}

	// Draw lines
	textColor := style.TextColor
	if textColor == nil {
		textColor = color.White
	}

	_, lineH := text.Measure("Ag", ta.FontFace, 0)
	lineHeight := lineH * 1.2

	for i, line := range ta.lines {
		y := r.Y + float64(i)*lineHeight - ta.ScrollY
		if y < r.Y-lineHeight || y > r.Y+r.H+lineHeight {
			continue // Skip lines outside visible area
		}

		op := &text.DrawOptions{}
		op.GeoM.Translate(r.X, y)
		op.ColorScale.ScaleWithColor(textColor)
		text.Draw(screen, line, ta.FontFace, op)
	}

	// Draw cursor
	if ta.Focused && ta.cursorVisible {
		cursorX := r.X
		if ta.CursorLine < len(ta.lines) {
			lineText := ta.lines[ta.CursorLine]
			if ta.CursorCol <= len([]rune(lineText)) {
				w, _ := text.Measure(string([]rune(lineText)[:ta.CursorCol]), ta.FontFace, 0)
				cursorX += w
			}
		}
		cursorY := r.Y + float64(ta.CursorLine)*lineHeight - ta.ScrollY + lineHeight*0.2

		DrawRoundedRectPath(screen, Rect{
			X: cursorX,
			Y: cursorY,
			W: 2,
			H: lineHeight * 0.8,
		}, 1, ta.CursorColor)
	}
}
