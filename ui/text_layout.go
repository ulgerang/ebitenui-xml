package ui

import (
	"math"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/rivo/uniseg"
)

type textWhiteSpaceMode string

const (
	textWhiteSpaceNormal  textWhiteSpaceMode = "normal"
	textWhiteSpacePreWrap textWhiteSpaceMode = "pre-wrap"
)

type textLayoutOptions struct {
	MaxWidth               float64
	Wrap                   bool
	WhiteSpace             textWhiteSpaceMode
	LineHeight             float64
	TrimTrailingWhitespace bool
}

type TextHit struct {
	LineIndex    int
	ClusterIndex int
	Text         string
	RuneStart    int
	RuneEnd      int
	Rect         Rect
}

type textLayoutCluster struct {
	Text       string
	ByteStart  int
	ByteEnd    int
	RuneStart  int
	RuneEnd    int
	BreakAfter int
	Advance    float64
	X          float64
	Width      float64
	Line       int
}

type textLayoutLine struct {
	Text         string
	StartCluster int
	EndCluster   int
	RuneStart    int
	RuneEnd      int
	Width        float64
}

type textLayout struct {
	text       string
	face       text.Face
	lineHeight float64
	clusters   []textLayoutCluster
	lines      []textLayoutLine
	width      float64
	height     float64
}

type graphemeBoundary struct {
	Text      string
	ByteStart int
	ByteEnd   int
	RuneStart int
	RuneEnd   int
}

func newTextLayout(content string, face text.Face, opts textLayoutOptions) *textLayout {
	layout := &textLayout{
		text: normalizeLayoutText(content, opts.WhiteSpace),
		face: face,
	}
	if opts.LineHeight > 0 {
		layout.lineHeight = opts.LineHeight
	}
	if face == nil {
		return layout
	}
	if layout.lineHeight <= 0 {
		layout.lineHeight = measureLineHeight(face)
	}
	layout.buildClusters()
	layout.buildLines(opts)
	return layout
}

func (tl *textLayout) buildClusters() {
	if tl.text == "" {
		return
	}

	rest := tl.text
	state := -1
	bytePos := 0
	runePos := 0
	var prefix strings.Builder
	var prevWidth float64

	for len(rest) > 0 {
		cluster, next, boundaries, nextState := uniseg.StepString(rest, state)
		runeCount := utf8.RuneCountInString(cluster)

		advance := 0.0
		if cluster != "\n" {
			prefix.WriteString(cluster)
			width, _ := text.Measure(prefix.String(), tl.face, 0)
			advance = width - prevWidth
			prevWidth = width
		} else {
			prefix.Reset()
			prevWidth = 0
		}

		tl.clusters = append(tl.clusters, textLayoutCluster{
			Text:       cluster,
			ByteStart:  bytePos,
			ByteEnd:    bytePos + len(cluster),
			RuneStart:  runePos,
			RuneEnd:    runePos + runeCount,
			BreakAfter: boundaries & uniseg.MaskLine,
			Advance:    advance,
			Line:       -1,
		})

		bytePos += len(cluster)
		runePos += runeCount
		rest = next
		state = nextState
	}
}

func (tl *textLayout) buildLines(opts textLayoutOptions) {
	if len(tl.clusters) == 0 {
		return
	}

	maxWidth := opts.MaxWidth
	if !opts.Wrap || maxWidth <= 0 {
		maxWidth = math.Inf(1)
	}

	lineStart := 0
	lineRuneStart := 0
	currentWidth := 0.0
	lastBreak := -1
	endedWithHardBreak := false

	emitLine := func(start, end int, runeStart int) {
		if start < 0 {
			start = 0
		}
		if end < start {
			end = start
		}

		visibleEnd := end
		if opts.TrimTrailingWhitespace {
			for visibleEnd > start && isWhitespaceCluster(tl.clusters[visibleEnd-1].Text) {
				visibleEnd--
			}
		}

		line := textLayoutLine{
			StartCluster: start,
			EndCluster:   visibleEnd,
			RuneStart:    runeStart,
			RuneEnd:      runeStart,
		}

		if visibleEnd > start {
			var builder strings.Builder
			var prefix strings.Builder
			prevWidth := 0.0
			for i := start; i < visibleEnd; i++ {
				cluster := &tl.clusters[i]
				builder.WriteString(cluster.Text)
				prefix.WriteString(cluster.Text)
				width, _ := text.Measure(prefix.String(), tl.face, 0)
				cluster.X = prevWidth
				cluster.Width = width - prevWidth
				cluster.Line = len(tl.lines)
				prevWidth = width
			}
			line.Text = builder.String()
			line.Width = prevWidth
			line.RuneEnd = tl.clusters[visibleEnd-1].RuneEnd
		}

		if line.Width > tl.width {
			tl.width = line.Width
		}
		tl.lines = append(tl.lines, line)
	}

	for i := 0; i < len(tl.clusters); {
		cluster := tl.clusters[i]

		if cluster.Text == "\n" {
			emitLine(lineStart, i, lineRuneStart)
			endedWithHardBreak = true
			i++
			lineStart = i
			lineRuneStart = cluster.RuneEnd
			currentWidth = 0
			lastBreak = -1
			continue
		}

		endedWithHardBreak = false
		nextWidth := currentWidth + cluster.Advance
		fits := nextWidth <= maxWidth || math.IsInf(maxWidth, 1) || i == lineStart
		if fits {
			currentWidth = nextWidth
			if cluster.BreakAfter != uniseg.LineDontBreak {
				lastBreak = i
			}
			i++
			continue
		}

		if lastBreak >= lineStart {
			emitLine(lineStart, lastBreak+1, lineRuneStart)
			i = lastBreak + 1
			lineStart = i
			if lineStart < len(tl.clusters) {
				lineRuneStart = tl.clusters[lineStart].RuneStart
			}
			currentWidth = 0
			lastBreak = -1
			continue
		}

		emitLine(lineStart, i, lineRuneStart)
		lineStart = i
		lineRuneStart = tl.clusters[lineStart].RuneStart
		currentWidth = 0
		lastBreak = -1
	}

	if lineStart < len(tl.clusters) {
		emitLine(lineStart, len(tl.clusters), lineRuneStart)
	} else if endedWithHardBreak {
		emitLine(lineStart, lineStart, lineRuneStart)
	}

	tl.height = float64(len(tl.lines)) * tl.lineHeight
}

func (tl *textLayout) HitTest(x, y float64) (TextHit, bool) {
	if len(tl.lines) == 0 || tl.lineHeight <= 0 {
		return TextHit{}, false
	}
	lineIndex := clampLineIndex(int(math.Floor(y/tl.lineHeight)), len(tl.lines))
	if lineIndex < 0 {
		return TextHit{}, false
	}
	line := tl.lines[lineIndex]
	for i := line.StartCluster; i < line.EndCluster; i++ {
		cluster := tl.clusters[i]
		if x < cluster.X || x > cluster.X+cluster.Width {
			continue
		}
		return TextHit{
			LineIndex:    lineIndex,
			ClusterIndex: i,
			Text:         cluster.Text,
			RuneStart:    cluster.RuneStart,
			RuneEnd:      cluster.RuneEnd,
			Rect: Rect{
				X: cluster.X,
				Y: float64(lineIndex) * tl.lineHeight,
				W: cluster.Width,
				H: tl.lineHeight,
			},
		}, true
	}
	return TextHit{}, false
}

func (tl *textLayout) CaretRuneIndexAt(x, y float64) int {
	if len(tl.lines) == 0 {
		return 0
	}
	lineIndex := clampLineIndex(int(math.Floor(y/tl.lineHeight)), len(tl.lines))
	if lineIndex < 0 {
		return 0
	}
	line := tl.lines[lineIndex]
	if x <= 0 {
		return line.RuneStart
	}
	for i := line.StartCluster; i < line.EndCluster; i++ {
		cluster := tl.clusters[i]
		mid := cluster.X + cluster.Width/2
		if x <= mid {
			return cluster.RuneStart
		}
		if x <= cluster.X+cluster.Width {
			return cluster.RuneEnd
		}
	}
	return line.RuneEnd
}

func clampLineIndex(index, lineCount int) int {
	if lineCount == 0 {
		return -1
	}
	if index < 0 {
		return 0
	}
	if index >= lineCount {
		return lineCount - 1
	}
	return index
}

func normalizeLayoutText(s string, whiteSpace textWhiteSpaceMode) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	s = strings.ReplaceAll(s, "\f", "\n")
	if whiteSpace == textWhiteSpacePreWrap {
		return s
	}

	parts := strings.Split(s, "\n")
	for i, part := range parts {
		parts[i] = collapseInlineWhitespace(part)
	}
	return strings.Join(parts, "\n")
}

func collapseInlineWhitespace(s string) string {
	var out []rune
	seenSpace := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			seenSpace = true
			continue
		}
		if seenSpace && len(out) > 0 {
			out = append(out, ' ')
		}
		out = append(out, r)
		seenSpace = false
	}
	return string(out)
}

func measureLineHeight(face text.Face) float64 {
	if face == nil {
		return 16
	}
	metrics := face.Metrics()
	return metrics.HAscent + metrics.HDescent + metrics.HLineGap
}

func truncateTextWithEllipsis(s string, face text.Face, maxWidth float64) string {
	if face == nil {
		return s
	}
	fullWidth, _ := text.Measure(s, face, 0)
	if fullWidth <= maxWidth {
		return s
	}

	const ellipsis = "..."
	ellipsisWidth, _ := text.Measure(ellipsis, face, 0)
	targetWidth := maxWidth - ellipsisWidth
	if targetWidth <= 0 {
		return ellipsis
	}

	boundaries := graphemeBoundaries(s)
	var builder strings.Builder
	for _, boundary := range boundaries {
		builder.WriteString(boundary.Text)
		width, _ := text.Measure(builder.String(), face, 0)
		if width > targetWidth {
			if builder.Len() == len(boundary.Text) {
				return ellipsis
			}
			return builder.String()[:builder.Len()-len(boundary.Text)] + ellipsis
		}
	}
	return s
}

func graphemeBoundaries(s string) []graphemeBoundary {
	if s == "" {
		return nil
	}
	rest := s
	state := -1
	bytePos := 0
	runePos := 0
	boundaries := make([]graphemeBoundary, 0)
	for len(rest) > 0 {
		cluster, next, _, nextState := uniseg.StepString(rest, state)
		runeCount := utf8.RuneCountInString(cluster)
		boundaries = append(boundaries, graphemeBoundary{
			Text:      cluster,
			ByteStart: bytePos,
			ByteEnd:   bytePos + len(cluster),
			RuneStart: runePos,
			RuneEnd:   runePos + runeCount,
		})
		bytePos += len(cluster)
		runePos += runeCount
		rest = next
		state = nextState
	}
	return boundaries
}

func snapRuneIndexToBoundary(s string, index int) int {
	if index <= 0 {
		return 0
	}
	runeCount := utf8.RuneCountInString(s)
	if index >= runeCount {
		return runeCount
	}
	for _, boundary := range graphemeBoundaries(s) {
		if index <= boundary.RuneStart {
			return boundary.RuneStart
		}
		if index < boundary.RuneEnd {
			return boundary.RuneStart
		}
	}
	return runeCount
}

func nextGraphemeBoundary(s string, index int) int {
	if index < 0 {
		index = 0
	}
	for _, boundary := range graphemeBoundaries(s) {
		if index < boundary.RuneEnd {
			return boundary.RuneEnd
		}
	}
	return utf8.RuneCountInString(s)
}

func prevGraphemeBoundary(s string, index int) int {
	if index <= 0 {
		return 0
	}
	prev := 0
	for _, boundary := range graphemeBoundaries(s) {
		if index <= boundary.RuneStart {
			return prev
		}
		prev = boundary.RuneStart
		if index <= boundary.RuneEnd {
			return boundary.RuneStart
		}
	}
	return prev
}

func runeIndexToClusterOffset(s string, runeIndex int) int {
	if runeIndex <= 0 {
		return 0
	}
	offset := 0
	for _, boundary := range graphemeBoundaries(s) {
		if runeIndex <= boundary.RuneStart {
			return offset
		}
		if runeIndex <= boundary.RuneEnd {
			return offset + (runeIndex - boundary.RuneStart)
		}
		offset += boundary.RuneEnd - boundary.RuneStart
	}
	return offset
}

func caretRuneIndexForSingleLine(face text.Face, s string, x float64) int {
	if face == nil || s == "" {
		return 0
	}
	layout := newTextLayout(s, face, textLayoutOptions{
		Wrap:       false,
		WhiteSpace: textWhiteSpacePreWrap,
		LineHeight: measureLineHeight(face),
	})
	return layout.CaretRuneIndexAt(x, 0)
}

func isWhitespaceCluster(s string) bool {
	if s == "" || s == "\n" {
		return false
	}
	for _, r := range s {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}
