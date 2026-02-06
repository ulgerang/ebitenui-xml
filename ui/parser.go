package ui

import (
	"encoding/xml"
	"fmt"
	"image/color"
	"io"
	"os"
	"strconv"
	"strings"
)

// XMLNode represents a parsed XML element
type XMLNode struct {
	XMLName  xml.Name
	ID       string     `xml:"id,attr"`
	Class    string     `xml:"class,attr"`
	Text     string     `xml:",chardata"`
	Children []XMLNode  `xml:",any"`
	Attrs    []xml.Attr `xml:",any,attr"`
}

// XMLParser parses XML layout files
type XMLParser struct{}

// NewXMLParser creates a new XML parser
func NewXMLParser() *XMLParser {
	return &XMLParser{}
}

// ParseFile parses an XML layout file
func (p *XMLParser) ParseFile(filename string) (*XMLNode, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return p.Parse(file)
}

// Parse parses XML from a reader
func (p *XMLParser) Parse(r io.Reader) (*XMLNode, error) {
	var root XMLNode
	decoder := xml.NewDecoder(r)
	if err := decoder.Decode(&root); err != nil {
		return nil, fmt.Errorf("failed to decode XML: %w", err)
	}
	return &root, nil
}

// ParseString parses XML from a string
func (p *XMLParser) ParseString(s string) (*XMLNode, error) {
	return p.Parse(strings.NewReader(s))
}

// GetAttr gets an attribute value by name
func (n *XMLNode) GetAttr(name string) string {
	for _, attr := range n.Attrs {
		if attr.Name.Local == name {
			return attr.Value
		}
	}
	return ""
}

// GetAttrFloat gets an attribute as float64
func (n *XMLNode) GetAttrFloat(name string) float64 {
	val := n.GetAttr(name)
	if val == "" {
		return 0
	}
	f, _ := strconv.ParseFloat(val, 64)
	return f
}

// GetAttrInt gets an attribute as int
func (n *XMLNode) GetAttrInt(name string) int {
	val := n.GetAttr(name)
	if val == "" {
		return 0
	}
	i, _ := strconv.Atoi(val)
	return i
}

// GetAttrBool gets an attribute as bool
func (n *XMLNode) GetAttrBool(name string) bool {
	val := strings.ToLower(n.GetAttr(name))
	return val == "true" || val == "1" || val == "yes"
}

// WidgetFactory creates widgets from XML nodes
type WidgetFactory struct {
	styleEngine *StyleEngine
}

// NewWidgetFactory creates a new widget factory
func NewWidgetFactory(styleEngine *StyleEngine) *WidgetFactory {
	return &WidgetFactory{
		styleEngine: styleEngine,
	}
}

// CreateFromXML creates a widget tree from XML
func (f *WidgetFactory) CreateFromXML(node *XMLNode) Widget {
	if node == nil {
		return nil
	}

	widget := f.createWidget(node)
	if widget == nil {
		return nil
	}

	// Apply styles
	if f.styleEngine != nil {
		// Apply by element type
		f.styleEngine.ApplyStyle(widget, node.XMLName.Local)
		// Apply by ID
		if node.ID != "" {
			f.styleEngine.ApplyStyle(widget, "#"+node.ID)
		}
		// Apply by class
		if node.Class != "" {
			classes := strings.Fields(node.Class)
			for _, class := range classes {
				f.styleEngine.ApplyStyle(widget, "."+class)
			}
		}
	}

	// Apply inline style attributes
	f.applyInlineStyles(widget, node)

	// Create children
	for _, childNode := range node.Children {
		// Skip text-only nodes
		if childNode.XMLName.Local == "" {
			continue
		}
		child := f.CreateFromXML(&childNode)
		if child != nil {
			widget.AddChild(child)
		}
	}

	return widget
}

// createWidget creates a specific widget type
func (f *WidgetFactory) createWidget(node *XMLNode) Widget {
	switch strings.ToLower(node.XMLName.Local) {
	case "panel", "div", "container":
		return NewPanel(node.ID)

	case "button", "btn":
		label := strings.TrimSpace(node.Text)
		if label == "" {
			label = node.GetAttr("label")
		}
		return NewButton(node.ID, label)

	case "text", "label", "span", "p":
		content := strings.TrimSpace(node.Text)
		if content == "" {
			content = node.GetAttr("content")
		}
		return NewText(node.ID, content)

	case "image", "img":
		return NewImage(node.ID)

	case "progressbar", "progress":
		pb := NewProgressBar(node.ID)
		if val := node.GetAttrFloat("value"); val > 0 {
			pb.Value = val
		}
		return pb

	case "input", "textinput":
		input := NewTextInput(node.ID)
		if placeholder := node.GetAttr("placeholder"); placeholder != "" {
			input.Placeholder = placeholder
		}
		if value := node.GetAttr("value"); value != "" {
			input.Text = value
		}
		if node.GetAttrBool("password") {
			input.Password = true
		}
		if node.GetAttrBool("readonly") {
			input.ReadOnly = true
		}
		if maxLen := node.GetAttrInt("maxlength"); maxLen > 0 {
			input.MaxLength = maxLen
		}
		return input

	case "textarea":
		ta := NewTextArea(node.ID)
		if placeholder := node.GetAttr("placeholder"); placeholder != "" {
			ta.Placeholder = placeholder
		}
		if value := node.GetAttr("value"); value != "" {
			ta.SetText(value)
		}
		if node.GetAttrBool("readonly") {
			ta.ReadOnly = true
		}
		return ta

	case "scrollable", "scroll", "scrollview":
		sc := NewScrollable(node.ID)
		if !node.GetAttrBool("vertical") && node.GetAttr("vertical") != "" {
			sc.ShowVertical = false
		}
		if node.GetAttrBool("horizontal") {
			sc.ShowHorizontal = true
		}
		return sc

	case "checkbox", "check":
		label := strings.TrimSpace(node.Text)
		if label == "" {
			label = node.GetAttr("label")
		}
		cb := NewCheckbox(node.ID, label)
		if node.GetAttrBool("checked") {
			cb.Checked = true
		}
		return cb

	case "slider", "range":
		sl := NewSlider(node.ID)
		if val := node.GetAttrFloat("value"); val > 0 {
			sl.Value = val
		}
		if min := node.GetAttrFloat("min"); min != 0 {
			sl.Min = min
		}
		if max := node.GetAttrFloat("max"); max != 0 {
			sl.Max = max
		}
		return sl

	case "toggle", "switch":
		label := strings.TrimSpace(node.Text)
		if label == "" {
			label = node.GetAttr("label")
		}
		tg := NewToggle(node.ID, label)
		if node.GetAttrBool("checked") {
			tg.Checked = true
		}
		return tg

	case "radiobutton", "radio":
		label := strings.TrimSpace(node.Text)
		if label == "" {
			label = node.GetAttr("label")
		}
		value := node.GetAttr("value")
		if value == "" {
			value = label
		}
		rb := NewRadioButton(node.ID, label, value)
		if node.GetAttrBool("selected") {
			rb.Selected = true
		}
		return rb

	case "dropdown", "select":
		dd := NewDropdown(node.ID)
		if placeholder := node.GetAttr("placeholder"); placeholder != "" {
			dd.Placeholder = placeholder
		}
		// Options can be added via children <option value="v">Label</option>
		for _, child := range node.Children {
			if strings.ToLower(child.XMLName.Local) == "option" {
				label := strings.TrimSpace(child.Text)
				value := child.GetAttr("value")
				if value == "" {
					value = label
				}
				dd.AddOption(label, value)
				if child.GetAttrBool("selected") {
					dd.SetValue(value)
				}
			}
		}
		return dd

	case "modal", "dialog":
		title := node.GetAttr("title")
		m := NewModal(node.ID, title)
		if content := node.GetAttr("content"); content != "" {
			m.Content = content
		}
		if node.GetAttrBool("open") {
			m.IsOpen = true
		}
		return m

	case "tooltip":
		text := strings.TrimSpace(node.Text)
		if text == "" {
			text = node.GetAttr("text")
		}
		tt := NewTooltip(node.ID, text)
		if pos := node.GetAttr("position"); pos != "" {
			tt.Position = pos
		}
		return tt

	case "badge":
		text := strings.TrimSpace(node.Text)
		if text == "" {
			text = node.GetAttr("text")
		}
		return NewBadge(node.ID, text)

	case "spinner", "loading":
		sp := NewSpinner(node.ID)
		if !node.GetAttrBool("spinning") && node.GetAttr("spinning") != "" {
			sp.IsSpinning = false
		}
		return sp

	case "toast", "notification":
		message := strings.TrimSpace(node.Text)
		if message == "" {
			message = node.GetAttr("message")
		}
		t := NewToast(node.ID, message)
		if typ := node.GetAttr("type"); typ != "" {
			t.ToastType = typ
		}
		if dur := node.GetAttrFloat("duration"); dur > 0 {
			t.Duration = dur
		}
		return t

	case "svg", "icon":
		svg := NewSVGIcon(node.ID)
		// Check for built-in icon
		if iconName := node.GetAttr("icon"); iconName != "" {
			strokeColor := parseColor(node.GetAttr("stroke"))
			if strokeColor == nil {
				strokeColor = parseColor(node.GetAttr("color"))
			}
			if strokeColor == nil {
				strokeColor = color.White
			}
			strokeWidth := node.GetAttrFloat("stroke-width")
			if strokeWidth <= 0 {
				strokeWidth = 2
			}
			svg.SetIcon(iconName, strokeColor, strokeWidth)
		} else if src := node.GetAttr("src"); src != "" {
			// Load from file
			svg.LoadFromFile(src)
		} else {
			// Check for inline SVG content
			inlineSVG := buildInlineSVG(node)
			if inlineSVG != "" {
				svg.LoadFromString(inlineSVG)
			}
		}
		return svg

	case "ui":
		// Root UI element, treat as panel
		return NewPanel(node.ID)

	default:
		// Unknown element, treat as panel
		return NewPanel(node.ID)
	}
}

// applyInlineStyles applies style attributes directly from XML
func (f *WidgetFactory) applyInlineStyles(widget Widget, node *XMLNode) {
	style := widget.Style()

	for _, attr := range node.Attrs {
		switch attr.Name.Local {
		case "width":
			style.Width = parseSize(attr.Value)
		case "height":
			style.Height = parseSize(attr.Value)
		case "direction", "layout":
			style.Direction = LayoutDirection(attr.Value)
		case "align":
			style.Align = Alignment(attr.Value)
		case "justify":
			style.Justify = Justify(attr.Value)
		case "gap":
			style.Gap = parseSize(attr.Value)
		case "padding":
			p := parseSize(attr.Value)
			style.Padding = Padding{p, p, p, p}
		case "margin":
			m := parseSize(attr.Value)
			style.Margin = Margin{m, m, m, m}
		case "background", "bg":
			style.Background = attr.Value
			style.BackgroundColor = parseColor(attr.Value)
		case "border-color":
			style.Border = attr.Value
			style.BorderColor = parseColor(attr.Value)
		case "border-width":
			style.BorderWidth = parseSize(attr.Value)
		case "color":
			style.Color = attr.Value
			style.TextColor = parseColor(attr.Value)
		case "font-size":
			style.FontSize = parseSize(attr.Value)
		case "flex-grow", "grow":
			style.FlexGrow, _ = strconv.ParseFloat(attr.Value, 64)
		}
	}
}

// parseSize parses a size value like "100" or "100px"
func parseSize(s string) float64 {
	s = strings.TrimSuffix(s, "px")
	s = strings.TrimSuffix(s, "%")
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// Note: parseColor is defined in style.go

// buildInlineSVG builds an SVG string from inline XML children
func buildInlineSVG(node *XMLNode) string {
	if len(node.Children) == 0 {
		return ""
	}

	var sb strings.Builder

	// Get viewBox from parent node or use defaults
	viewBox := node.GetAttr("viewBox")
	if viewBox == "" {
		w := node.GetAttr("width")
		h := node.GetAttr("height")
		if w == "" {
			w = "24"
		}
		if h == "" {
			h = "24"
		}
		viewBox = "0 0 " + w + " " + h
	}

	width := node.GetAttr("width")
	height := node.GetAttr("height")
	if width == "" {
		width = "24"
	}
	if height == "" {
		height = "24"
	}

	sb.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" `)
	sb.WriteString(`width="` + width + `" height="` + height + `" `)
	sb.WriteString(`viewBox="` + viewBox + `">`)

	// Build children
	for _, child := range node.Children {
		buildSVGElement(&sb, &child)
	}

	sb.WriteString("</svg>")
	return sb.String()
}

// buildSVGElement recursively builds SVG element strings
func buildSVGElement(sb *strings.Builder, node *XMLNode) {
	if node.XMLName.Local == "" {
		return
	}

	sb.WriteString("<")
	sb.WriteString(node.XMLName.Local)

	// Write attributes
	for _, attr := range node.Attrs {
		sb.WriteString(" ")
		sb.WriteString(attr.Name.Local)
		sb.WriteString(`="`)
		sb.WriteString(attr.Value)
		sb.WriteString(`"`)
	}

	if len(node.Children) == 0 && strings.TrimSpace(node.Text) == "" {
		sb.WriteString("/>")
	} else {
		sb.WriteString(">")
		sb.WriteString(strings.TrimSpace(node.Text))
		for _, child := range node.Children {
			buildSVGElement(sb, &child)
		}
		sb.WriteString("</")
		sb.WriteString(node.XMLName.Local)
		sb.WriteString(">")
	}
}
