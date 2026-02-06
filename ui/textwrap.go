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

		words := tw.splitWords(para)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}

		var currentLine strings.Builder
		var currentWidth float64

		for i, word := range words {
			wordWidth, _ := text.Measure(word, tw.face, 0)
			spaceWidth, _ := text.Measure(" ", tw.face, 0)

			if currentWidth == 0 {
				// First word on line
				currentLine.WriteString(word)
				currentWidth = wordWidth
			} else if currentWidth+spaceWidth+wordWidth <= maxWidth {
				// Word fits on current line
				currentLine.WriteString(" ")
				currentLine.WriteString(word)
				currentWidth += spaceWidth + wordWidth
			} else {
				// Word doesn't fit, start new line
				lines = append(lines, currentLine.String())
				currentLine.Reset()
				currentLine.WriteString(word)
				currentWidth = wordWidth
			}

			// If this is the last word, add the line
			if i == len(words)-1 {
				lines = append(lines, currentLine.String())
			}
		}
	}

	return lines
}

// splitWords splits text into words, preserving spaces for CJK characters
func (tw *TextWrapper) splitWords(s string) []string {
	var words []string
	var current strings.Builder

	for _, r := range s {
		if unicode.IsSpace(r) {
			if current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
			}
		} else if isCJK(r) {
			// CJK characters can be broken anywhere
			if current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
			}
			words = append(words, string(r))
		} else {
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		words = append(words, current.String())
	}

	return words
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
