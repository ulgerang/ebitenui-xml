// Package main provides a tool that converts ebitenui-xml layout (XML) and
// styles (JSON) into a standard HTML+CSS page. The generated page can be
// opened in a browser and screenshot-captured for pixel-level comparison
// with the Ebiten rendering output.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

// ------------------------------------------------------------------
// Data model – mirrors the subset of ebitenui-xml types we need.
// ------------------------------------------------------------------

type Padding struct {
	Top    float64 `json:"top"`
	Right  float64 `json:"right"`
	Bottom float64 `json:"bottom"`
	Left   float64 `json:"left"`
}

type Style struct {
	// Layout
	Direction string  `json:"direction"`
	Align     string  `json:"align"`
	Justify   string  `json:"justify"`
	Gap       float64 `json:"gap"`
	FlexWrap  string  `json:"flexWrap"`
	FlexGrow  float64 `json:"flexGrow"`

	// Sizing
	Width     float64 `json:"width"`
	Height    float64 `json:"height"`
	MinWidth  float64 `json:"minWidth"`
	MinHeight float64 `json:"minHeight"`
	MaxWidth  float64 `json:"maxWidth"`
	MaxHeight float64 `json:"maxHeight"`

	// Spacing
	Padding Padding `json:"padding"`
	Margin  Padding `json:"margin"` // reuses Padding shape

	// Colors
	Background string `json:"background"`
	Border     string `json:"border"`
	Color      string `json:"color"`

	// Border
	BorderWidth  float64 `json:"borderWidth"`
	BorderRadius float64 `json:"borderRadius"`

	// Text
	FontSize   float64 `json:"fontSize"`
	TextAlign  string  `json:"textAlign"`
	LineHeight float64 `json:"lineHeight"`
	TextWrap   string  `json:"textWrap"`

	// Effects
	Opacity    float64 `json:"opacity"`
	BoxShadow  string  `json:"boxShadow"`
	TextShadow string  `json:"textShadow"`
	Outline    string  `json:"outline"`

	// States
	Hover  *Style `json:"hover"`
	Active *Style `json:"active"`
}

type StyleSheet struct {
	Styles map[string]*Style `json:"styles"`
}

// ------------------------------------------------------------------
// XML layout structs (lightweight)
// ------------------------------------------------------------------

// We do a simplified manual XML parse because the layout is small/flat.
// For robustness, a real implementation would use encoding/xml.

type XMLNode struct {
	Tag      string
	ID       string
	Classes  []string
	Attrs    map[string]string
	Text     string
	Children []*XMLNode
}

// ------------------------------------------------------------------
// Style → CSS conversion
// ------------------------------------------------------------------

func styleToCSS(s *Style) string {
	var b strings.Builder

	// Display: all ebitenui-xml elements are flex containers.
	b.WriteString("display: flex;\n")

	if s.Direction == "column" {
		b.WriteString("  flex-direction: column;\n")
	} else if s.Direction == "row" || s.Direction == "" {
		b.WriteString("  flex-direction: row;\n")
	}

	if s.Align != "" {
		b.WriteString(fmt.Sprintf("  align-items: %s;\n", mapAlign(s.Align)))
	}
	if s.Justify != "" {
		b.WriteString(fmt.Sprintf("  justify-content: %s;\n", s.Justify))
	}
	if s.Gap > 0 {
		b.WriteString(fmt.Sprintf("  gap: %.0fpx;\n", s.Gap))
	}
	if s.FlexWrap != "" {
		b.WriteString(fmt.Sprintf("  flex-wrap: %s;\n", s.FlexWrap))
	}
	if s.FlexGrow > 0 {
		b.WriteString(fmt.Sprintf("  flex-grow: %g;\n", s.FlexGrow))
	}

	// Sizing
	if s.Width > 0 {
		b.WriteString(fmt.Sprintf("  width: %.0fpx;\n", s.Width))
	}
	if s.Height > 0 {
		b.WriteString(fmt.Sprintf("  height: %.0fpx;\n", s.Height))
	}
	if s.MinWidth > 0 {
		b.WriteString(fmt.Sprintf("  min-width: %.0fpx;\n", s.MinWidth))
	}
	if s.MinHeight > 0 {
		b.WriteString(fmt.Sprintf("  min-height: %.0fpx;\n", s.MinHeight))
	}
	if s.MaxWidth > 0 {
		b.WriteString(fmt.Sprintf("  max-width: %.0fpx;\n", s.MaxWidth))
	}
	if s.MaxHeight > 0 {
		b.WriteString(fmt.Sprintf("  max-height: %.0fpx;\n", s.MaxHeight))
	}

	// Spacing
	if s.Padding.Top > 0 || s.Padding.Right > 0 || s.Padding.Bottom > 0 || s.Padding.Left > 0 {
		b.WriteString(fmt.Sprintf("  padding: %.0fpx %.0fpx %.0fpx %.0fpx;\n",
			s.Padding.Top, s.Padding.Right, s.Padding.Bottom, s.Padding.Left))
	}
	if s.Margin.Top > 0 || s.Margin.Right > 0 || s.Margin.Bottom > 0 || s.Margin.Left > 0 {
		b.WriteString(fmt.Sprintf("  margin: %.0fpx %.0fpx %.0fpx %.0fpx;\n",
			s.Margin.Top, s.Margin.Right, s.Margin.Bottom, s.Margin.Left))
	}

	// Colors
	if s.Background != "" {
		b.WriteString(fmt.Sprintf("  background: %s;\n", s.Background))
	}
	if s.Color != "" {
		b.WriteString(fmt.Sprintf("  color: %s;\n", s.Color))
	}

	// Border
	if s.BorderWidth > 0 && s.Border != "" {
		b.WriteString(fmt.Sprintf("  border: %.0fpx solid %s;\n", s.BorderWidth, s.Border))
	} else if s.BorderWidth > 0 {
		b.WriteString(fmt.Sprintf("  border-width: %.0fpx;\n  border-style: solid;\n", s.BorderWidth))
	}
	if s.BorderRadius > 0 {
		b.WriteString(fmt.Sprintf("  border-radius: %.0fpx;\n", s.BorderRadius))
	}

	// Text
	if s.FontSize > 0 {
		b.WriteString(fmt.Sprintf("  font-size: %.0fpx;\n", s.FontSize))
	}
	if s.TextAlign != "" {
		b.WriteString(fmt.Sprintf("  text-align: %s;\n", s.TextAlign))
	}
	if s.LineHeight > 0 {
		b.WriteString(fmt.Sprintf("  line-height: %.0fpx;\n", s.LineHeight))
	}
	if s.TextWrap == "normal" {
		b.WriteString("  white-space: normal;\n  word-wrap: break-word;\n")
	} else if s.TextWrap == "nowrap" {
		b.WriteString("  white-space: nowrap;\n")
	}

	// Effects
	if s.Opacity > 0 && s.Opacity < 1 {
		b.WriteString(fmt.Sprintf("  opacity: %g;\n", s.Opacity))
	}
	if s.BoxShadow != "" {
		b.WriteString(fmt.Sprintf("  box-shadow: %s;\n", convertBoxShadow(s.BoxShadow)))
	}
	if s.TextShadow != "" {
		b.WriteString(fmt.Sprintf("  text-shadow: %s;\n", convertBoxShadow(s.TextShadow)))
	}
	if s.Outline != "" {
		b.WriteString(fmt.Sprintf("  outline: %s;\n", s.Outline))
	}

	// box-sizing
	b.WriteString("  box-sizing: border-box;\n")

	return b.String()
}

func mapAlign(a string) string {
	switch a {
	case "start":
		return "flex-start"
	case "end":
		return "flex-end"
	case "center":
		return "center"
	case "stretch":
		return "stretch"
	default:
		return a
	}
}

// convertBoxShadow converts ebitenui-xml shadow format to CSS.
// ebitenui-xml: "offsetX offsetY blur spread color"
// CSS:          "offsetXpx offsetYpx blurpx spreadpx color"
func convertBoxShadow(s string) string {
	parts := strings.Fields(s)
	if len(parts) < 4 {
		return s
	}
	// First 4 are numeric, add "px" suffix; rest is color.
	var result []string
	for i, p := range parts {
		if i < 4 && isNumericStart(p) {
			result = append(result, p+"px")
		} else {
			result = append(result, p)
		}
	}
	return strings.Join(result, " ")
}

func isNumericStart(s string) bool {
	if len(s) == 0 {
		return false
	}
	c := s[0]
	return (c >= '0' && c <= '9') || c == '-' || c == '.'
}

// ------------------------------------------------------------------
// Generate hover/active CSS
// ------------------------------------------------------------------

func stateCSS(selector string, state string, s *Style) string {
	if s == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s:%s {\n", selector, state))
	if s.Background != "" {
		b.WriteString(fmt.Sprintf("  background: %s;\n", s.Background))
	}
	if s.Color != "" {
		b.WriteString(fmt.Sprintf("  color: %s;\n", s.Color))
	}
	if s.BoxShadow != "" {
		b.WriteString(fmt.Sprintf("  box-shadow: %s;\n", convertBoxShadow(s.BoxShadow)))
	}
	b.WriteString("}\n")
	return b.String()
}

// ------------------------------------------------------------------
// Parse simplified XML (good enough for ebitenui-xml layouts)
// ------------------------------------------------------------------

func parseXML(data string) (*XMLNode, error) {
	data = strings.TrimSpace(data)
	root, _, err := parseElement(data, 0)
	return root, err
}

func parseElement(data string, pos int) (*XMLNode, int, error) {
	// Skip whitespace and comments
	pos = skipWhitespaceAndComments(data, pos)
	if pos >= len(data) || data[pos] != '<' {
		return nil, pos, fmt.Errorf("expected '<' at pos %d", pos)
	}

	// Read tag opening
	pos++ // skip '<'
	tagEnd := strings.IndexAny(data[pos:], " \t\n\r/>")
	if tagEnd < 0 {
		return nil, pos, fmt.Errorf("unclosed tag at pos %d", pos)
	}
	node := &XMLNode{Tag: data[pos : pos+tagEnd], Attrs: make(map[string]string)}
	pos += tagEnd

	// Read attributes
	for {
		pos = skipWhitespace(data, pos)
		if pos >= len(data) {
			break
		}
		if data[pos] == '/' && pos+1 < len(data) && data[pos+1] == '>' {
			pos += 2 // self-closing
			break
		}
		if data[pos] == '>' {
			pos++ // end of opening tag
			break
		}
		// Read attribute name=value
		eqIdx := strings.IndexByte(data[pos:], '=')
		if eqIdx < 0 {
			break
		}
		name := strings.TrimSpace(data[pos : pos+eqIdx])
		pos += eqIdx + 1
		pos = skipWhitespace(data, pos)
		if pos >= len(data) || data[pos] != '"' {
			break
		}
		pos++
		valEnd := strings.IndexByte(data[pos:], '"')
		if valEnd < 0 {
			break
		}
		value := data[pos : pos+valEnd]
		pos += valEnd + 1

		switch name {
		case "id":
			node.ID = value
		case "class":
			node.Classes = strings.Fields(value)
		default:
			node.Attrs[name] = value
		}
	}

	// Check for self-closing (already handled above) or read children/text
	if pos > 2 && data[pos-2] == '/' && data[pos-1] == '>' {
		return node, pos, nil
	}

	// Read children and text until closing tag
	closingTag := fmt.Sprintf("</%s>", node.Tag)
	var textParts []string

	for {
		pos = skipWhitespaceAndComments(data, pos)
		if pos >= len(data) {
			break
		}

		// Check for closing tag
		if strings.HasPrefix(data[pos:], closingTag) {
			pos += len(closingTag)
			break
		}

		// Check for child element
		if data[pos] == '<' && pos+1 < len(data) && data[pos+1] != '/' {
			child, newPos, err := parseElement(data, pos)
			if err != nil {
				// Maybe it's text with angle brackets?
				break
			}
			node.Children = append(node.Children, child)
			pos = newPos
			continue
		}

		// Read text
		nextTag := strings.IndexByte(data[pos:], '<')
		if nextTag < 0 {
			textParts = append(textParts, strings.TrimSpace(data[pos:]))
			break
		}
		text := strings.TrimSpace(data[pos : pos+nextTag])
		if text != "" {
			textParts = append(textParts, text)
		}
		pos += nextTag
	}

	node.Text = strings.Join(textParts, " ")
	return node, pos, nil
}

func skipWhitespace(data string, pos int) int {
	for pos < len(data) && (data[pos] == ' ' || data[pos] == '\t' || data[pos] == '\n' || data[pos] == '\r') {
		pos++
	}
	return pos
}

func skipWhitespaceAndComments(data string, pos int) int {
	for {
		pos = skipWhitespace(data, pos)
		if pos+4 < len(data) && data[pos:pos+4] == "<!--" {
			end := strings.Index(data[pos:], "-->")
			if end >= 0 {
				pos += end + 3
				continue
			}
		}
		break
	}
	return pos
}

// ------------------------------------------------------------------
// Map ebitenui-xml tags → HTML tags
// ------------------------------------------------------------------

func tagToHTML(tag string) string {
	switch tag {
	case "ui":
		return "div"
	case "panel":
		return "div"
	case "button":
		return "button"
	case "text":
		return "span"
	case "progressbar":
		return "div"
	case "textinput":
		return "input"
	default:
		return "div"
	}
}

// ------------------------------------------------------------------
// Generate HTML recursively
// ------------------------------------------------------------------

func nodeToHTML(node *XMLNode, indent int) string {
	var b strings.Builder
	prefix := strings.Repeat("  ", indent)
	htmlTag := tagToHTML(node.Tag)

	// Build class list
	var classes []string
	classes = append(classes, fmt.Sprintf("eui-%s", node.Tag)) // type class for CSS matching
	classes = append(classes, node.Classes...)

	b.WriteString(fmt.Sprintf("%s<%s", prefix, htmlTag))
	if node.ID != "" {
		b.WriteString(fmt.Sprintf(` id="%s"`, node.ID))
	}
	if len(classes) > 0 {
		b.WriteString(fmt.Sprintf(` class="%s"`, strings.Join(classes, " ")))
	}

	// Special: progressbar
	if node.Tag == "progressbar" {
		val := node.Attrs["value"]
		if val == "" {
			val = "0"
		}
		b.WriteString(fmt.Sprintf(` data-value="%s"`, val))
	}

	b.WriteString(">\n")

	// Text content
	if node.Text != "" {
		b.WriteString(fmt.Sprintf("%s  %s\n", prefix, node.Text))
	}

	// Progress bar inner fill
	if node.Tag == "progressbar" {
		val := node.Attrs["value"]
		if val == "" {
			val = "0"
		}
		b.WriteString(fmt.Sprintf(`%s  <div class="progress-fill" style="width: calc(%s * 100%%); height: 100%%;"></div>`+"\n", prefix, val))
	}

	// Children
	for _, child := range node.Children {
		b.WriteString(nodeToHTML(child, indent+1))
	}

	b.WriteString(fmt.Sprintf("%s</%s>\n", prefix, htmlTag))
	return b.String()
}

// ------------------------------------------------------------------
// Main: read inputs → generate HTML page
// ------------------------------------------------------------------

const htmlTemplate = `<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>EbitenUI-XML CSS Reference</title>
  <style>
    /* Reset */
    *, *::before, *::after { margin: 0; padding: 0; box-sizing: border-box; }
    body {
      width: {{.Width}}px;
      height: {{.Height}}px;
      overflow: hidden;
      font-family: monospace;
      background: #0f0f1a;
    }

    /* Progress bar fill defaults */
    .progress-fill {
      border-radius: inherit;
      transition: width 0.3s;
    }

    /* === Generated Styles === */
    {{.CSS}}
  </style>
</head>
<body>
{{.HTML}}
</body>
</html>
`

func main() {
	layoutPath := flag.String("layout", "", "Path to layout XML file")
	stylesPath := flag.String("styles", "", "Path to styles JSON file")
	outPath := flag.String("out", "reference.html", "Output HTML file path")
	width := flag.Int("width", 640, "Canvas width")
	height := flag.Int("height", 480, "Canvas height")
	flag.Parse()

	if *layoutPath == "" || *stylesPath == "" {
		fmt.Println("Usage: converter -layout <xml> -styles <json> [-out <html>] [-width 640] [-height 480]")
		fmt.Println("\nConverts ebitenui-xml layout + styles into an HTML/CSS reference page")
		fmt.Println("for visual comparison testing.")
		os.Exit(1)
	}

	// Read inputs
	layoutData, err := os.ReadFile(*layoutPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading layout: %v\n", err)
		os.Exit(1)
	}
	stylesData, err := os.ReadFile(*stylesPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading styles: %v\n", err)
		os.Exit(1)
	}

	// Parse styles
	var sheet StyleSheet
	// Try nested format first
	if err := json.Unmarshal(stylesData, &sheet); err != nil || sheet.Styles == nil {
		// Try flat format: { "#root": {...} }
		sheet.Styles = make(map[string]*Style)
		if err := json.Unmarshal(stylesData, &sheet.Styles); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing styles: %v\n", err)
			os.Exit(1)
		}
	}

	// Parse layout
	root, err := parseXML(string(layoutData))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing layout: %v\n", err)
		os.Exit(1)
	}

	// Generate CSS
	var cssBuilder strings.Builder
	for selector, style := range sheet.Styles {
		cssSelector := selectorToCSS(selector)
		cssBuilder.WriteString(fmt.Sprintf("%s {\n  %s}\n\n", cssSelector, styleToCSS(style)))

		// Hover state
		if style.Hover != nil {
			cssBuilder.WriteString(stateCSS(cssSelector, "hover", style.Hover))
		}
		// Active state
		if style.Active != nil {
			cssBuilder.WriteString(stateCSS(cssSelector, "active", style.Active))
		}
	}

	// Generate HTML
	htmlContent := nodeToHTML(root, 1)

	// Render template
	tmpl, err := template.New("page").Parse(htmlTemplate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Template error: %v\n", err)
		os.Exit(1)
	}

	// Ensure output directory exists
	outDir := filepath.Dir(*outPath)
	if outDir != "" && outDir != "." {
		os.MkdirAll(outDir, 0755)
	}

	f, err := os.Create(*outPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	tmpl.Execute(f, map[string]interface{}{
		"Width":  *width,
		"Height": *height,
		"CSS":    template.CSS(cssBuilder.String()),
		"HTML":   template.HTML(htmlContent),
	})

	fmt.Printf("✅ Generated reference HTML: %s (%dx%d)\n", *outPath, *width, *height)
}

// selectorToCSS converts ebitenui-xml selectors to CSS selectors.
// "#id" → "#id", ".class" → ".class", "tagName" → ".eui-tagName"
func selectorToCSS(sel string) string {
	if strings.HasPrefix(sel, "#") || strings.HasPrefix(sel, ".") {
		return sel
	}
	// Type selector: "button" → ".eui-button"
	return ".eui-" + sel
}
