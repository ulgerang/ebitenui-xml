package ui

import (
	"strings"
	"unicode"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// TextWrapper handles text wrapping and measurement
type TextWrapper struct {
	face text.Face
}

// NewTextWrapper creates a new text wrapper
func NewTextWrapper(face text.Face) *TextWrapper {
	return &TextWrapper{face: face}
}

// WrapText wraps text to fit within maxWidth
func (tw *TextWrapper) WrapText(s string, maxWidth float64) []string {
	if tw.face == nil || maxWidth <= 0 {
		return []string{s}
	}

	var lines []string
	paragraphs := strings.Split(s, "\n")

	for _, para := range paragraphs {
		if para == "" {
			lines = append(lines, "")
			continue
		}

		tokens := tw.splitTokens(para)
		if len(tokens) == 0 {
			lines = append(lines, "")
			continue
		}

		var currentLine strings.Builder
		var currentWidth float64

		for _, tok := range tokens {
			if tok == "" {
				continue
			}
			// Skip leading spaces on a new line.
			if currentWidth == 0 && tok == " " {
				continue
			}

			tokWidth, _ := text.Measure(tok, tw.face, 0)
			if currentWidth+tokWidth <= maxWidth || currentWidth == 0 {
				currentLine.WriteString(tok)
				currentWidth += tokWidth
				continue
			}

			// Token doesn't fit, start new line.
			line := strings.TrimRight(currentLine.String(), " ")
			lines = append(lines, line)
			currentLine.Reset()
			currentWidth = 0

			// Don't start a new line with a space.
			if tok == " " {
				continue
			}
			currentLine.WriteString(tok)
			currentWidth = tokWidth
		}

		// Flush last line for this paragraph.
		lines = append(lines, strings.TrimRight(currentLine.String(), " "))
	}

	return lines
}

// splitTokens splits text into tokens suitable for wrapping.
//
// - Collapses any whitespace run into a single " " token.
// - Emits each CJK rune as its own token so wrapping can occur anywhere.
// - Emits non-CJK sequences (e.g., Latin words) as a token.
//
// This avoids inserting extra spaces between CJK glyphs during wrapping.
func (tw *TextWrapper) splitTokens(s string) []string {
	var tokens []string
	var current strings.Builder

	flushCurrent := func() {
		if current.Len() == 0 {
			return
		}
		tokens = append(tokens, current.String())
		current.Reset()
	}

	appendSpace := func() {
		if len(tokens) == 0 {
			return
		}
		if tokens[len(tokens)-1] == " " {
			return
		}
		tokens = append(tokens, " ")
	}

	for _, r := range s {
		if unicode.IsSpace(r) {
			flushCurrent()
			appendSpace()
			continue
		}
		if isCJK(r) {
			flushCurrent()
			tokens = append(tokens, string(r))
			continue
		}
		current.WriteRune(r)
	}

	flushCurrent()
	return tokens
}

// isCJK checks if a rune is a CJK character
func isCJK(r rune) bool {
	return unicode.Is(unicode.Han, r) ||
		unicode.Is(unicode.Hangul, r) ||
		unicode.Is(unicode.Hiragana, r) ||
		unicode.Is(unicode.Katakana, r)
}

// MeasureText measures text dimensions
func (tw *TextWrapper) MeasureText(s string) (width, height float64) {
	if tw.face == nil {
		return 0, 0
	}
	return text.Measure(s, tw.face, 0)
}

// MeasureLines measures wrapped text dimensions
func (tw *TextWrapper) MeasureLines(lines []string, lineHeight float64) (width, height float64) {
	if tw.face == nil {
		return 0, 0
	}

	for _, line := range lines {
		w, _ := text.Measure(line, tw.face, 0)
		if w > width {
			width = w
		}
	}

	height = float64(len(lines)) * lineHeight
	return width, height
}

// TruncateWithEllipsis truncates text to fit maxWidth, adding ellipsis
func (tw *TextWrapper) TruncateWithEllipsis(s string, maxWidth float64) string {
	if tw.face == nil {
		return s
	}

	fullWidth, _ := text.Measure(s, tw.face, 0)
	if fullWidth <= maxWidth {
		return s
	}

	ellipsis := "..."
	ellipsisWidth, _ := text.Measure(ellipsis, tw.face, 0)
	targetWidth := maxWidth - ellipsisWidth

	if targetWidth <= 0 {
		return ellipsis
	}

	// Binary search for the right length
	runes := []rune(s)
	low, high := 0, len(runes)

	for low < high {
		mid := (low + high + 1) / 2
		substr := string(runes[:mid])
		w, _ := text.Measure(substr, tw.face, 0)
		if w <= targetWidth {
			low = mid
		} else {
			high = mid - 1
		}
	}

	return string(runes[:low]) + ellipsis
}

// LineHeight returns a reasonable line height for the font
func (tw *TextWrapper) LineHeight() float64 {
	if tw.face == nil {
		return 16 // default
	}

	metrics := tw.face.Metrics()
	return metrics.HAscent + metrics.HDescent + metrics.HLineGap
}
