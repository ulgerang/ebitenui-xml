package ui

import (
	"encoding/xml"
	"fmt"
	"image/color"
	"io"
	"os"
	"reflect"
	"regexp"
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

// GetFirstAttr returns the first non-empty attribute value for the given names.
func (n *XMLNode) GetFirstAttr(names ...string) string {
	for _, name := range names {
		if val := n.GetAttr(name); val != "" {
			return val
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
	styleEngine     *StyleEngine
	bindings        *BindingContext
	commands        map[string]func(Widget)
	radioGroups     map[string]*RadioGroup
	formSubmit      map[Widget]string
	formReset       map[Widget]string
	onTreeChanged   func()
	onLayoutChanged func()
}

// NewWidgetFactory creates a new widget factory
func NewWidgetFactory(styleEngine *StyleEngine, bindings ...*BindingContext) *WidgetFactory {
	var bindingContext *BindingContext
	if len(bindings) > 0 {
		bindingContext = bindings[0]
	}
	return &WidgetFactory{
		styleEngine: styleEngine,
		bindings:    bindingContext,
		commands:    make(map[string]func(Widget)),
		radioGroups: make(map[string]*RadioGroup),
		formSubmit:  make(map[Widget]string),
		formReset:   make(map[Widget]string),
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

	// Transfer XML class attribute to widget's classes field
	// This is essential for class-based style selectors (e.g., ".danger")
	if node.Class != "" {
		for _, class := range strings.Fields(node.Class) {
			widget.AddClass(class)
		}
	}
	f.applySemanticMetadata(widget, node)

	// NOTE: Style application (type/id/class selectors) is handled by
	// reapplyStyles() in LoadLayout/LoadStyles. Applying here would cause
	// double-application, where type defaults overwrite class overrides.

	// Apply inline style attributes (XML attributes like width="100")
	f.applyInlineStyles(widget, node)
	f.applyWidgetMetadata(widget, node)
	f.applyBindingAttributes(widget, node)
	f.applyTemplateBindings(widget, node)
	f.applyAttributeBindings(widget, node)
	f.applyStyleBindings(widget, node)
	f.applyCommandBindings(widget, node)

	// Create children
	for _, childNode := range node.Children {
		// Skip text-only nodes
		if childNode.XMLName.Local == "" {
			continue
		}
		if f.applyRepeatBinding(widget, &childNode) {
			continue
		}
		if f.applyConditionalBinding(widget, &childNode) {
			continue
		}
		child := f.CreateFromXML(&childNode)
		if child != nil {
			widget.AddChild(child)
		}
	}
	f.applyContainerSemantics(widget, node)

	return widget
}

// applyRepeatBinding expands a child XML template for each item in a bound
// collection. The template node itself is not attached to the widget tree.
func (f *WidgetFactory) applyRepeatBinding(parent Widget, template *XMLNode) bool {
	if f.bindings == nil || parent == nil || template == nil {
		return false
	}

	key := template.GetFirstAttr("bind-repeat", "data-bind-repeat", "for-each")
	if key == "" {
		return false
	}

	var rendered []Widget
	render := func(value interface{}) {
		for _, child := range rendered {
			parent.RemoveChild(child)
		}
		rendered = rendered[:0]

		items := bindingItems(value)
		for i, item := range items {
			node := renderRepeatTemplate(template, item, i)
			child := f.CreateFromXML(&node)
			if child == nil {
				continue
			}
			parent.AddChild(child)
			rendered = append(rendered, child)
		}

		if f.onTreeChanged != nil {
			f.onTreeChanged()
		}
	}

	f.bindings.Bind(key, parent, render)
	return true
}

// applyConditionalBinding attaches or detaches a child XML template based on a
// boolean binding value. The template node itself is not attached immediately.
func (f *WidgetFactory) applyConditionalBinding(parent Widget, template *XMLNode) bool {
	if f.bindings == nil || parent == nil || template == nil {
		return false
	}

	key := template.GetFirstAttr("bind-if", "data-bind-if")
	if key == "" {
		return false
	}

	var rendered Widget
	render := func(value interface{}) {
		shouldRender, ok := bindingBool(value)
		if !ok {
			shouldRender = false
		}

		if rendered != nil && !shouldRender {
			parent.RemoveChild(rendered)
			rendered = nil
			if f.onTreeChanged != nil {
				f.onTreeChanged()
			}
			return
		}
		if rendered != nil || !shouldRender {
			return
		}

		node := cloneNodeWithoutAttrs(template, "bind-if", "data-bind-if")
		rendered = f.CreateFromXML(&node)
		if rendered != nil {
			parent.AddChild(rendered)
			if f.onTreeChanged != nil {
				f.onTreeChanged()
			}
		}
	}

	f.bindings.Bind(key, parent, render)
	return true
}

func (f *WidgetFactory) applySemanticMetadata(widget Widget, node *XMLNode) {
	tag := strings.ToLower(node.XMLName.Local)
	if bw := baseWidgetOf(widget); bw != nil {
		bw.SetSemanticType(tag)
	}
	if tag == "" {
		return
	}
	widget.AddClass(tag)
	if tag != widget.Type() {
		widget.AddClass("as-" + tag)
	}
	applySemanticLayoutDefaults(widget, tag)
}

func applySemanticLayoutDefaults(widget Widget, tag string) {
	style := widget.Style()
	switch tag {
	case "table", "thead", "tbody", "tfoot":
		if style.Direction == "" {
			style.Direction = LayoutColumn
		}
		if style.Align == "" {
			style.Align = AlignStretch
		}
	case "tr":
		if style.Direction == "" {
			style.Direction = LayoutRow
		}
		if style.Align == "" {
			style.Align = AlignStretch
		}
	case "td", "th":
		if !style.FlexGrowSet && style.FlexGrow == 0 {
			style.FlexGrow = 1
			style.FlexGrowSet = true
		}
		if style.BoxSizing == "" {
			style.BoxSizing = "border-box"
		}
		if tag == "th" && style.FontWeight == "" {
			style.FontWeight = "bold"
		}
	}
}

func (f *WidgetFactory) applyWidgetMetadata(widget Widget, node *XMLNode) {
	if bw := baseWidgetOf(widget); bw != nil {
		if tabindex := node.GetAttr("tabindex"); tabindex != "" {
			if value, err := strconv.Atoi(tabindex); err == nil {
				bw.SetTabIndex(value)
			}
		}
		if isDefaultFocusable(widget) {
			bw.SetFocusable(true)
		}
		bw.SetValidationRules(validationRulesFromNode(node))
		switch w := widget.(type) {
		case *TextInput:
			bw.SetFormInitialValue(w.Text)
			if rules := bw.ValidationRules(); rules.MaxLengthSet {
				w.MaxLength = rules.MaxLength
			}
		case *TextArea:
			bw.SetFormInitialValue(w.Text)
		case *Checkbox:
			bw.SetFormInitialValue(strconv.FormatBool(w.Checked))
		case *Toggle:
			bw.SetFormInitialValue(strconv.FormatBool(w.Checked))
		case *RadioButton:
			bw.SetFormInitialValue(strconv.FormatBool(w.Selected))
		case *Dropdown:
			bw.SetFormInitialValue(w.GetSelectedValue())
		case *Slider:
			bw.SetFormInitialValue(strconv.FormatFloat(w.Value, 'f', -1, 64))
		}
	}
}

func validationRulesFromNode(node *XMLNode) ValidationRules {
	rules := ValidationRules{
		Required: node.GetAttrBool("required"),
		Pattern:  node.GetAttr("pattern"),
		Type:     strings.ToLower(node.GetAttr("type")),
		Message:  firstNonEmpty(node.GetAttr("validation-message"), node.GetAttr("data-validation-message")),
	}
	if min := node.GetAttr("min"); min != "" {
		if value, err := strconv.ParseFloat(min, 64); err == nil {
			rules.Min = value
			rules.MinSet = true
		}
	}
	if max := node.GetAttr("max"); max != "" {
		if value, err := strconv.ParseFloat(max, 64); err == nil {
			rules.Max = value
			rules.MaxSet = true
		}
	}
	if minLength := firstNonEmpty(node.GetAttr("minlength"), node.GetAttr("minLength")); minLength != "" {
		if value, err := strconv.Atoi(minLength); err == nil {
			rules.MinLength = value
			rules.MinLengthSet = true
		}
	}
	if maxLength := firstNonEmpty(node.GetAttr("maxlength"), node.GetAttr("maxLength")); maxLength != "" {
		if value, err := strconv.Atoi(maxLength); err == nil {
			rules.MaxLength = value
			rules.MaxLengthSet = true
		}
	}
	return rules
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func (f *WidgetFactory) applyContainerSemantics(widget Widget, node *XMLNode) {
	tag := strings.ToLower(node.XMLName.Local)
	if tag == "fieldset" && node.GetAttrBool("disabled") {
		setSubtreeEnabled(widget, false)
	}
}

func baseWidgetOf(widget Widget) *BaseWidget {
	switch w := widget.(type) {
	case *BaseWidget:
		return w
	case *Panel:
		return w.BaseWidget
	case *Button:
		return w.BaseWidget
	case *Text:
		return w.BaseWidget
	case *Image:
		return w.BaseWidget
	case *ProgressBar:
		return w.BaseWidget
	case *Slider:
		return w.BaseWidget
	case *Checkbox:
		return w.BaseWidget
	case *SVGIcon:
		return w.BaseWidget
	case *TextInput:
		return w.BaseWidget
	case *TextArea:
		return w.BaseWidget
	case *Scrollable:
		return w.BaseWidget
	case *Toggle:
		return w.BaseWidget
	case *RadioButton:
		return w.BaseWidget
	case *Dropdown:
		return w.BaseWidget
	case *Modal:
		return w.BaseWidget
	case *Tooltip:
		return w.BaseWidget
	case *Badge:
		return w.BaseWidget
	case *Spinner:
		return w.BaseWidget
	case *Toast:
		return w.BaseWidget
	default:
		return nil
	}
}

func isDefaultFocusable(widget Widget) bool {
	switch widget.(type) {
	case *Button, *TextInput, *TextArea, *Checkbox, *Toggle, *RadioButton, *Dropdown, *Slider:
		return true
	default:
		return false
	}
}

func setSubtreeEnabled(widget Widget, enabled bool) {
	if widget == nil {
		return
	}
	widget.SetEnabled(enabled)
	for _, child := range widget.Children() {
		setSubtreeEnabled(child, enabled)
	}
}

// applyBindingAttributes wires declarative XML binding attributes to the
// BindingContext owned by the UI manager.
func (f *WidgetFactory) applyBindingAttributes(widget Widget, node *XMLNode) {
	if f.bindings == nil || widget == nil || node == nil {
		return
	}

	if key := node.GetFirstAttr("bind-text", "data-bind-text"); key != "" {
		f.bindTextLike(key, widget)
	}
	if key := node.GetFirstAttr("bind-value", "data-bind-value"); key != "" {
		f.bindValue(key, widget)
	}
	if key := node.GetFirstAttr("bind-checked", "data-bind-checked"); key != "" {
		f.bindChecked(key, widget)
	}
	if key := node.GetFirstAttr("bind-visible", "data-bind-visible"); key != "" {
		f.bindExpression(key, widget, func(value interface{}) {
			widget.SetVisible(bindingTruthy(value))
		})
	}
	if key := node.GetFirstAttr("bind-enabled", "data-bind-enabled"); key != "" {
		f.bindExpression(key, widget, func(value interface{}) {
			widget.SetEnabled(bindingTruthy(value))
		})
	}
	if key := node.GetFirstAttr("bind-options", "data-bind-options"); key != "" {
		f.bindOptions(key, widget, node)
	}
}

func (f *WidgetFactory) bindTextLike(key string, widget Widget) {
	f.bindExpression(key, widget, func(value interface{}) {
		textValue := bindingString(value)
		switch w := widget.(type) {
		case *Text:
			w.SetContent(textValue)
		case *Button:
			w.Label = textValue
		case *Checkbox:
			w.Label = textValue
		case *Toggle:
			w.Label = textValue
		case *RadioButton:
			w.Label = textValue
		case *Badge:
			w.Text = textValue
		case *Tooltip:
			w.Text = textValue
		case *Toast:
			w.Message = textValue
		case *Modal:
			w.Content = textValue
		}
	})
}

func (f *WidgetFactory) bindExpression(expr string, widget Widget, updater func(value interface{})) {
	f.bindExpressionAttr(expr, widget, "", updater)
}

func (f *WidgetFactory) bindExpressionAttr(expr string, widget Widget, attr string, updater func(value interface{})) {
	deps := bindingExpressionDependencies(expr)
	if len(deps) == 0 {
		if value, ok := evalBindingExpression(expr, f.bindings); ok {
			updater(value)
		} else if f.bindings != nil {
			f.bindings.ReportError(widget, attr, expr, "failed to evaluate binding expression")
		}
		return
	}

	update := func() {
		if value, ok := evalBindingExpression(expr, f.bindings); ok {
			updater(value)
		} else if f.bindings != nil {
			f.bindings.ReportError(widget, attr, expr, "failed to evaluate binding expression")
		}
	}
	for _, dep := range deps {
		f.bindings.Bind(dep, widget, func(interface{}) {
			update()
		})
	}
	update()
}

func (f *WidgetFactory) bindOptions(expr string, widget Widget, node *XMLNode) {
	switch strings.ToLower(node.GetFirstAttr("option-type", "data-option-type")) {
	case "radio":
		f.bindRadioOptions(expr, widget, node)
		return
	case "checkbox":
		f.bindCheckboxOptions(expr, widget, node)
		return
	}

	dropdown, ok := widget.(*Dropdown)
	if !ok {
		f.bindings.ReportError(widget, "bind-options", expr, "bind-options is only supported on dropdown/select widgets or option-type=radio containers")
		return
	}
	labelPath := node.GetFirstAttr("option-label", "data-option-label")
	valuePath := node.GetFirstAttr("option-value", "data-option-value")
	f.bindExpressionAttr(expr, widget, "bind-options", func(value interface{}) {
		options, ok := dropdownOptionsFromBinding(value, labelPath, valuePath)
		if !ok {
			f.bindings.ReportError(widget, "bind-options", expr, "bound value is not a collection")
			return
		}
		selectedValue := dropdown.GetSelectedValue()
		dropdown.SetOptions(options)
		if selectedValue != "" {
			dropdown.SetValue(selectedValue)
		}
		if dropdown.SelectedIndex < 0 && len(options) > 0 {
			valueAttr := node.GetFirstAttr("value", "selected")
			if valueAttr != "" {
				dropdown.SetValue(valueAttr)
			}
		}
		if f.onLayoutChanged != nil {
			f.onLayoutChanged()
		}
	})
}

func (f *WidgetFactory) bindRadioOptions(expr string, widget Widget, node *XMLNode) {
	container, ok := widget.(*Panel)
	if !ok {
		f.bindings.ReportError(widget, "bind-options", expr, "option-type=radio is only supported on panel-like containers")
		return
	}

	labelPath := node.GetFirstAttr("option-label", "data-option-label")
	valuePath := node.GetFirstAttr("option-value", "data-option-value")
	valueBinding := node.GetFirstAttr("bind-value", "data-bind-value")
	groupName := node.GetFirstAttr("option-name", "data-option-name", "name")
	idPrefix := node.GetFirstAttr("option-id-prefix", "data-option-id-prefix")
	if groupName == "" {
		groupName = widget.ID()
	}
	if groupName == "" {
		groupName = "radio-options"
	}

	group := f.radioGroups[groupName]
	if group == nil {
		group = NewRadioGroup(groupName)
		f.radioGroups[groupName] = group
	}
	syncingValue := false
	if valueBinding != "" {
		originalOnChange := group.OnChange
		group.OnChange = func(value string) {
			if syncingValue {
				if originalOnChange != nil {
					originalOnChange(value)
				}
				return
			}
			f.bindings.Set(valueBinding, value)
			if originalOnChange != nil {
				originalOnChange(value)
			}
		}
		f.bindings.Bind(valueBinding, widget, func(value interface{}) {
			syncingValue = true
			defer func() { syncingValue = false }()
			group.SetValue(bindingString(value))
		})
	}

	var rendered []*RadioButton
	f.bindExpressionAttr(expr, widget, "bind-options", func(value interface{}) {
		options, ok := dropdownOptionsFromBinding(value, labelPath, valuePath)
		if !ok {
			f.bindings.ReportError(widget, "bind-options", expr, "bound value is not a collection")
			return
		}

		for _, child := range rendered {
			container.RemoveChild(child)
		}
		rendered = rendered[:0]
		group.Buttons = group.Buttons[:0]

		selectedValue := group.Value
		if selectedValue == "" && valueBinding != "" {
			selectedValue = bindingString(f.bindings.Get(valueBinding))
		}
		valueStillExists := false
		for i, option := range options {
			id := optionBindingID(container.ID(), groupName, option.Value, i, idPrefix)
			radio := NewRadioButton(id, option.Label, option.Value)
			radio.SetFocusable(true)
			group.AddButton(radio)
			if option.Value == selectedValue {
				valueStillExists = true
			}
			container.AddChild(radio)
			rendered = append(rendered, radio)
		}
		if valueStillExists {
			group.SetValue(selectedValue)
		} else if selectedValue != "" {
			group.Value = ""
		}
		if f.onTreeChanged != nil {
			f.onTreeChanged()
		}
	})
}

func (f *WidgetFactory) bindCheckboxOptions(expr string, widget Widget, node *XMLNode) {
	container, ok := widget.(*Panel)
	if !ok {
		f.bindings.ReportError(widget, "bind-options", expr, "option-type=checkbox is only supported on panel-like containers")
		return
	}

	labelPath := node.GetFirstAttr("option-label", "data-option-label")
	valuePath := node.GetFirstAttr("option-value", "data-option-value")
	valueBinding := node.GetFirstAttr("bind-value", "data-bind-value")
	groupName := node.GetFirstAttr("option-name", "data-option-name", "name")
	idPrefix := node.GetFirstAttr("option-id-prefix", "data-option-id-prefix")
	if groupName == "" {
		groupName = widget.ID()
	}
	if groupName == "" {
		groupName = "checkbox-options"
	}

	syncingValue := false
	var rendered []*Checkbox
	optionValues := make(map[*Checkbox]string)

	updateBindingFromRendered := func() {
		if valueBinding == "" || syncingValue {
			return
		}
		values := make([]string, 0, len(rendered))
		for _, checkbox := range rendered {
			if checkbox.Checked {
				values = append(values, optionValues[checkbox])
			}
		}
		syncingValue = true
		f.bindings.Set(valueBinding, values)
		syncingValue = false
	}

	applySelectedValues := func(value interface{}) {
		if syncingValue {
			return
		}
		selected := bindingStringSet(value)
		for _, checkbox := range rendered {
			checkbox.Checked = selected[optionValues[checkbox]]
		}
	}

	if valueBinding != "" {
		f.bindings.Bind(valueBinding, widget, applySelectedValues)
	}

	f.bindExpressionAttr(expr, widget, "bind-options", func(value interface{}) {
		options, ok := dropdownOptionsFromBinding(value, labelPath, valuePath)
		if !ok {
			f.bindings.ReportError(widget, "bind-options", expr, "bound value is not a collection")
			return
		}

		for _, child := range rendered {
			container.RemoveChild(child)
			delete(optionValues, child)
		}
		rendered = rendered[:0]

		selected := bindingStringSet(f.bindings.Get(valueBinding))
		for i, option := range options {
			id := optionBindingID(container.ID(), groupName, option.Value, i, idPrefix)
			checkbox := NewCheckbox(id, option.Label)
			checkbox.SetFocusable(true)
			checkbox.Checked = selected[option.Value]
			optionValues[checkbox] = option.Value
			checkbox.OnChange = func(bool) {
				updateBindingFromRendered()
			}
			container.AddChild(checkbox)
			rendered = append(rendered, checkbox)
		}
		if f.onTreeChanged != nil {
			f.onTreeChanged()
		}
	})
}

func optionBindingID(containerID, groupName, value string, index int, prefix string) string {
	base := containerID
	if prefix != "" {
		base = prefix
	}
	if base == "" {
		base = groupName
	}
	if base == "" {
		base = "radio"
	}
	slug := strings.ToLower(value)
	slug = regexp.MustCompile(`[^a-z0-9_-]+`).ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = strconv.Itoa(index)
	}
	return fmt.Sprintf("%s-option-%s", base, slug)
}

func bindingStringSet(value interface{}) map[string]bool {
	set := make(map[string]bool)
	for _, item := range bindingItems(value) {
		set[bindingString(item)] = true
	}
	return set
}

func (f *WidgetFactory) bindValue(key string, widget Widget) {
	switch w := widget.(type) {
	case *Text:
		f.bindings.BindText(key, w)
	case *TextInput:
		f.bindings.Bind(key, w, func(value interface{}) {
			w.SetText(fmt.Sprintf("%v", value))
		})
		originalOnChange := w.OnChange
		w.OnChange = func(text string) {
			f.bindings.Set(key, text)
			if originalOnChange != nil {
				originalOnChange(text)
			}
		}
	case *TextArea:
		f.bindings.Bind(key, w, func(value interface{}) {
			w.SetText(fmt.Sprintf("%v", value))
		})
		originalOnChange := w.OnChange
		w.OnChange = func(text string) {
			f.bindings.Set(key, text)
			if originalOnChange != nil {
				originalOnChange(text)
			}
		}
	case *ProgressBar:
		f.bindings.BindProgress(key, w)
	case *Slider:
		f.bindings.BindSlider(key, w)
	case *Dropdown:
		f.bindings.Bind(key, w, func(value interface{}) {
			w.SetValue(fmt.Sprintf("%v", value))
		})
		originalOnChange := w.OnChange
		w.OnChange = func(index int, value string) {
			f.bindings.Set(key, value)
			if originalOnChange != nil {
				originalOnChange(index, value)
			}
		}
	}
}

func (f *WidgetFactory) bindChecked(key string, widget Widget) {
	switch w := widget.(type) {
	case *Checkbox:
		f.bindings.BindCheckbox(key, w)
	case *Toggle:
		f.bindings.Bind(key, w, func(value interface{}) {
			if checked, ok := value.(bool); ok {
				w.Checked = checked
			}
		})
		originalOnChange := w.OnChange
		w.OnChange = func(checked bool) {
			f.bindings.Set(key, checked)
			if originalOnChange != nil {
				originalOnChange(checked)
			}
		}
	case *RadioButton:
		f.bindings.Bind(key, w, func(value interface{}) {
			switch v := value.(type) {
			case bool:
				w.Selected = v
			case string:
				w.Selected = v == w.Value
			}
		})
	}
}

func (f *WidgetFactory) applyTemplateBindings(widget Widget, node *XMLNode) {
	if f.bindings == nil || widget == nil || node == nil {
		return
	}
	if node.GetFirstAttr("bind-text", "data-bind-text") != "" {
		return
	}

	template := f.templateSource(widget, node)
	if !strings.Contains(template, "{{") {
		return
	}

	keys := extractBindingExpressionDeps(template)
	if len(keys) == 0 {
		return
	}

	update := func() {
		textValue := renderBindingExpressionTemplate(template, f.bindings)
		switch w := widget.(type) {
		case *Text:
			w.SetContent(textValue)
		case *Button:
			w.Label = textValue
		case *Checkbox:
			w.Label = textValue
		case *Toggle:
			w.Label = textValue
		case *RadioButton:
			w.Label = textValue
		case *Badge:
			w.Text = textValue
		case *Tooltip:
			w.Text = textValue
		case *Toast:
			w.Message = textValue
		case *Modal:
			w.Content = textValue
		}
	}

	for _, key := range keys {
		f.bindings.Bind(key, widget, func(interface{}) {
			update()
		})
	}
	update()
}

func (f *WidgetFactory) templateSource(widget Widget, node *XMLNode) string {
	textValue := strings.TrimSpace(node.Text)
	if textValue == "" {
		textValue = node.GetFirstAttr("content", "label", "text", "message")
	}
	return textValue
}

func (f *WidgetFactory) applyAttributeBindings(widget Widget, node *XMLNode) {
	if f.bindings == nil || widget == nil || node == nil {
		return
	}
	for _, attr := range node.Attrs {
		name := bindingSuffix(attr.Name.Local, "bind-attr-", "data-bind-attr-")
		if name == "" {
			continue
		}
		attrName := normalizeBindingName(name)
		f.bindExpression(attr.Value, widget, func(value interface{}) {
			f.applyBoundAttribute(widget, attrName, value)
		})
	}
}

func (f *WidgetFactory) applyBoundAttribute(widget Widget, name string, value interface{}) {
	switch name {
	case "class":
		for _, class := range widget.Classes() {
			widget.RemoveClass(class)
		}
		for _, class := range strings.Fields(bindingString(value)) {
			widget.AddClass(class)
		}
		if f.onTreeChanged != nil {
			f.onTreeChanged()
		}
		return
	case "label", "text", "content":
		f.setTextLikeAttribute(widget, bindingString(value))
	case "placeholder":
		switch w := widget.(type) {
		case *TextInput:
			w.Placeholder = bindingString(value)
		case *TextArea:
			w.Placeholder = bindingString(value)
		case *Dropdown:
			w.Placeholder = bindingString(value)
		}
	case "value":
		f.applyBoundValue(widget, value)
	case "checked", "selected":
		switch w := widget.(type) {
		case *Checkbox:
			w.Checked = bindingTruthy(value)
		case *Toggle:
			w.Checked = bindingTruthy(value)
		case *RadioButton:
			w.Selected = bindingTruthy(value)
		}
	case "disabled":
		widget.SetEnabled(!bindingTruthy(value))
	case "enabled":
		widget.SetEnabled(bindingTruthy(value))
	case "visible":
		widget.SetVisible(bindingTruthy(value))
	case "width":
		widget.Style().Width = parseSize(bindingString(value))
		widget.Style().WidthSet = true
	case "height":
		widget.Style().Height = parseSize(bindingString(value))
		widget.Style().HeightSet = true
	case "minwidth":
		widget.Style().MinWidth = parseSize(bindingString(value))
		widget.Style().MinWidthSet = true
	case "maxwidth":
		widget.Style().MaxWidth = parseSize(bindingString(value))
		widget.Style().MaxWidthSet = true
	case "minheight":
		widget.Style().MinHeight = parseSize(bindingString(value))
		widget.Style().MinHeightSet = true
	case "maxheight":
		widget.Style().MaxHeight = parseSize(bindingString(value))
		widget.Style().MaxHeightSet = true
	}
	if f.onLayoutChanged != nil {
		f.onLayoutChanged()
	}
}

func (f *WidgetFactory) setTextLikeAttribute(widget Widget, value string) {
	switch w := widget.(type) {
	case *Text:
		w.SetContent(value)
	case *Button:
		w.Label = value
	case *Checkbox:
		w.Label = value
	case *Toggle:
		w.Label = value
	case *RadioButton:
		w.Label = value
	case *Badge:
		w.Text = value
	case *Tooltip:
		w.Text = value
	case *Toast:
		w.Message = value
	case *Modal:
		w.Content = value
	}
}

func (f *WidgetFactory) applyBoundValue(widget Widget, value interface{}) {
	switch w := widget.(type) {
	case *Text:
		w.SetContent(bindingString(value))
	case *TextInput:
		w.SetText(bindingString(value))
	case *TextArea:
		w.SetText(bindingString(value))
	case *ProgressBar:
		if f, ok := bindingFloat(value); ok {
			w.Value = f
		}
	case *Slider:
		if f, ok := bindingFloat(value); ok {
			w.Value = f
		}
	case *Dropdown:
		w.SetValue(bindingString(value))
	}
}

func (f *WidgetFactory) applyStyleBindings(widget Widget, node *XMLNode) {
	if f.bindings == nil || widget == nil || node == nil {
		return
	}
	for _, attr := range node.Attrs {
		name := bindingSuffix(attr.Name.Local, "bind-style-", "data-bind-style-")
		if name == "" {
			continue
		}
		styleName := normalizeBindingName(name)
		f.bindExpression(attr.Value, widget, func(value interface{}) {
			f.applyBoundStyle(widget, styleName, value)
		})
	}
}

func (f *WidgetFactory) applyBoundStyle(widget Widget, name string, value interface{}) {
	style := widget.Style()
	text := bindingString(value)
	switch name {
	case "color":
		style.Color = text
		style.TextColor = parseColor(text)
	case "background":
		style.Background = text
		style.BackgroundColor = parseColor(text)
		style.parsedGradient = ParseGradient(text)
	case "border":
		style.Border = text
		style.BorderColor = parseColor(text)
	case "opacity":
		if f, ok := bindingFloat(value); ok {
			style.Opacity = f
			style.OpacitySet = true
		}
	case "display":
		style.Display = text
	case "visibility":
		style.Visibility = text
	case "transform":
		style.Transform = text
	case "filter":
		style.Filter = text
		style.parsedFilter = ParseFilter(text)
	case "animation":
		style.Animation = text
		style.parsedAnimation = ParseAnimationDeclaration(text)
	}
	if f.onLayoutChanged != nil {
		f.onLayoutChanged()
	}
}

func (f *WidgetFactory) applyCommandBindings(widget Widget, node *XMLNode) {
	if widget == nil || node == nil {
		return
	}
	for _, attr := range node.Attrs {
		switch normalizeBindingName(attr.Name.Local) {
		case "onclick":
			widget.OnClick(f.widgetCommandHandler(attr.Value, widget))
		case "onchange":
			f.bindChangeCommand(widget, attr.Value)
		case "onsubmit":
			if bw := baseWidgetOf(widget); bw != nil && bw.SemanticType() == "form" {
				f.formSubmit[widget] = attr.Value
			}
			f.bindSubmitCommand(widget, attr.Value)
		case "onreset":
			if bw := baseWidgetOf(widget); bw != nil && bw.SemanticType() == "form" {
				f.formReset[widget] = attr.Value
			}
		}
	}
}

func (f *WidgetFactory) widgetCommandHandler(name string, widget Widget) func() {
	return func() {
		f.runCommand(name, widget)
	}
}

func (f *WidgetFactory) bindChangeCommand(widget Widget, name string) {
	switch w := widget.(type) {
	case *TextInput:
		original := w.OnChange
		w.OnChange = func(text string) {
			if original != nil {
				original(text)
			}
			f.runCommand(name, widget)
		}
	case *TextArea:
		original := w.OnChange
		w.OnChange = func(text string) {
			if original != nil {
				original(text)
			}
			f.runCommand(name, widget)
		}
	case *Checkbox:
		original := w.OnChange
		w.OnChange = func(checked bool) {
			if original != nil {
				original(checked)
			}
			f.runCommand(name, widget)
		}
	case *Toggle:
		original := w.OnChange
		w.OnChange = func(checked bool) {
			if original != nil {
				original(checked)
			}
			f.runCommand(name, widget)
		}
	case *Slider:
		original := w.OnChange
		w.OnChange = func(value float64) {
			if original != nil {
				original(value)
			}
			f.runCommand(name, widget)
		}
	case *Dropdown:
		original := w.OnChange
		w.OnChange = func(index int, value string) {
			if original != nil {
				original(index, value)
			}
			f.runCommand(name, widget)
		}
	}
}

func (f *WidgetFactory) bindSubmitCommand(widget Widget, name string) {
	switch w := widget.(type) {
	case *TextInput:
		original := w.OnSubmit
		w.OnSubmit = func(text string) {
			if original != nil {
				original(text)
			}
			f.runCommand(name, widget)
		}
	}
}

func (f *WidgetFactory) runCommand(name string, widget Widget) {
	if f.commands == nil {
		return
	}
	if handler := f.commands[name]; handler != nil {
		handler(widget)
	}
}

func bindingSuffix(name string, prefixes ...string) string {
	for _, prefix := range prefixes {
		if strings.HasPrefix(name, prefix) {
			return strings.TrimPrefix(name, prefix)
		}
	}
	return ""
}

func normalizeBindingName(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, "-", ""))
}

func extractTemplateKeys(template string) []string {
	seen := make(map[string]bool)
	var keys []string
	remaining := template
	for {
		start := strings.Index(remaining, "{{")
		if start < 0 {
			break
		}
		remaining = remaining[start+2:]
		end := strings.Index(remaining, "}}")
		if end < 0 {
			break
		}
		key := strings.TrimSpace(remaining[:end])
		if key != "" && !seen[key] {
			seen[key] = true
			keys = append(keys, key)
		}
		remaining = remaining[end+2:]
	}
	return keys
}

func renderBindingTemplate(template string, bindings *BindingContext) string {
	var rendered strings.Builder
	remaining := template
	for {
		start := strings.Index(remaining, "{{")
		if start < 0 {
			rendered.WriteString(remaining)
			break
		}
		rendered.WriteString(remaining[:start])
		remaining = remaining[start+2:]
		end := strings.Index(remaining, "}}")
		if end < 0 {
			rendered.WriteString("{{")
			rendered.WriteString(remaining)
			break
		}
		key := strings.TrimSpace(remaining[:end])
		if value := bindings.Get(key); value != nil {
			rendered.WriteString(fmt.Sprintf("%v", value))
		}
		remaining = remaining[end+2:]
	}
	return rendered.String()
}

func renderRepeatTemplate(template *XMLNode, item interface{}, index int) XMLNode {
	node := *template
	node.ID = renderRepeatString(node.ID, item, index)
	node.Class = renderRepeatString(node.Class, item, index)
	node.Text = renderRepeatString(node.Text, item, index)
	node.Attrs = renderRepeatAttrs(node.Attrs, item, index)

	if len(template.Children) > 0 {
		node.Children = make([]XMLNode, len(template.Children))
		for i := range template.Children {
			node.Children[i] = renderRepeatTemplate(&template.Children[i], item, index)
		}
	}
	return node
}

func renderRepeatAttrs(attrs []xml.Attr, item interface{}, index int) []xml.Attr {
	rendered := make([]xml.Attr, 0, len(attrs))
	for _, attr := range attrs {
		switch attr.Name.Local {
		case "bind-repeat", "data-bind-repeat", "for-each":
			continue
		}
		attr.Value = renderRepeatString(attr.Value, item, index)
		rendered = append(rendered, attr)
	}
	return rendered
}

func renderRepeatString(template string, item interface{}, index int) string {
	if !strings.Contains(template, "{{") {
		return template
	}

	var rendered strings.Builder
	remaining := template
	for {
		start := strings.Index(remaining, "{{")
		if start < 0 {
			rendered.WriteString(remaining)
			break
		}
		rendered.WriteString(remaining[:start])
		remaining = remaining[start+2:]
		end := strings.Index(remaining, "}}")
		if end < 0 {
			rendered.WriteString("{{")
			rendered.WriteString(remaining)
			break
		}
		expr := strings.TrimSpace(remaining[:end])
		if value, ok := repeatExpressionValue(expr, item, index); ok {
			rendered.WriteString(fmt.Sprintf("%v", value))
		}
		remaining = remaining[end+2:]
	}
	return rendered.String()
}

func repeatExpressionValue(expr string, item interface{}, index int) (interface{}, bool) {
	switch expr {
	case "index":
		return index, true
	case "item":
		return item, true
	}
	if strings.HasPrefix(expr, "item.") {
		return lookupBindingPath(item, strings.TrimPrefix(expr, "item."))
	}
	return nil, false
}

func bindingItems(value interface{}) []interface{} {
	if value == nil {
		return nil
	}
	if items, ok := value.([]interface{}); ok {
		return items
	}

	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return nil
	}

	items := make([]interface{}, 0, v.Len())
	for i := 0; i < v.Len(); i++ {
		items = append(items, v.Index(i).Interface())
	}
	return items
}

func dropdownOptionsFromBinding(value interface{}, labelPath, valuePath string) ([]DropdownOption, bool) {
	items := bindingItems(value)
	if items == nil && value != nil {
		return nil, false
	}
	options := make([]DropdownOption, 0, len(items))
	for _, item := range items {
		label := bindingString(item)
		optionValue := label
		if labelPath != "" {
			if v, ok := lookupBindingPath(item, labelPath); ok {
				label = bindingString(v)
			} else {
				label = ""
			}
		}
		if valuePath != "" {
			if v, ok := lookupBindingPath(item, valuePath); ok {
				optionValue = bindingString(v)
			} else {
				optionValue = ""
			}
		} else if labelPath != "" {
			optionValue = label
		}
		options = append(options, DropdownOption{Label: label, Value: optionValue})
	}
	return options, true
}

func bindingBool(value interface{}) (bool, bool) {
	switch v := value.(type) {
	case bool:
		return v, true
	case string:
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "true", "1", "yes", "on":
			return true, true
		case "false", "0", "no", "off", "":
			return false, true
		}
	case int:
		return v != 0, true
	case int64:
		return v != 0, true
	case float64:
		return v != 0, true
	}
	return false, false
}

func cloneNodeWithoutAttrs(node *XMLNode, names ...string) XMLNode {
	clone := *node
	clone.Attrs = make([]xml.Attr, 0, len(node.Attrs))
	for _, attr := range node.Attrs {
		remove := false
		for _, name := range names {
			if attr.Name.Local == name {
				remove = true
				break
			}
		}
		if !remove {
			clone.Attrs = append(clone.Attrs, attr)
		}
	}
	if len(node.Children) > 0 {
		clone.Children = make([]XMLNode, len(node.Children))
		for i := range node.Children {
			clone.Children[i] = cloneNodeWithoutAttrs(&node.Children[i])
		}
	}
	return clone
}

func lookupBindingPath(value interface{}, path string) (interface{}, bool) {
	current := reflect.ValueOf(value)
	for _, part := range strings.Split(path, ".") {
		if part == "" {
			return nil, false
		}
		if current.Kind() == reflect.Interface || current.Kind() == reflect.Ptr {
			if current.IsNil() {
				return nil, false
			}
			current = current.Elem()
		}

		switch current.Kind() {
		case reflect.Map:
			key := reflect.ValueOf(part)
			if !key.Type().AssignableTo(current.Type().Key()) {
				if key.Type().ConvertibleTo(current.Type().Key()) {
					key = key.Convert(current.Type().Key())
				} else {
					return nil, false
				}
			}
			current = current.MapIndex(key)
			if !current.IsValid() {
				return nil, false
			}
		case reflect.Struct:
			field := current.FieldByName(part)
			if !field.IsValid() {
				field = current.FieldByNameFunc(func(name string) bool {
					return strings.EqualFold(name, part)
				})
			}
			if !field.IsValid() || !field.CanInterface() {
				return nil, false
			}
			current = field
		default:
			return nil, false
		}
	}
	if current.Kind() == reflect.Interface || current.Kind() == reflect.Ptr {
		if current.IsNil() {
			return nil, false
		}
		current = current.Elem()
	}
	if !current.IsValid() || !current.CanInterface() {
		return nil, false
	}
	return current.Interface(), true
}

// createWidget creates a specific widget type
func (f *WidgetFactory) createWidget(node *XMLNode) Widget {
	switch strings.ToLower(node.XMLName.Local) {
	case "panel", "div", "container", "form", "fieldset", "nav", "section", "article", "header", "footer", "main",
		"ul", "ol", "li", "table", "thead", "tbody", "tr", "td", "th":
		return NewPanel(node.ID)

	case "button", "btn":
		label := strings.TrimSpace(node.Text)
		if label == "" {
			label = node.GetAttr("label")
		}
		return NewButton(node.ID, label)

	case "text", "label", "span", "p", "legend", "h1", "h2", "h3", "h4", "h5", "h6":
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
		if step := node.GetAttrFloat("step"); step > 0 {
			sl.Step = step
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
		if name := node.GetAttr("name"); name != "" {
			group := f.radioGroups[name]
			if group == nil {
				group = NewRadioGroup(name)
				f.radioGroups[name] = group
			}
			group.AddButton(rb)
			if rb.Selected {
				group.SetValue(rb.Value)
			}
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
			style.WidthSet = true
		case "height":
			style.Height = parseSize(attr.Value)
			style.HeightSet = true
		case "direction", "layout":
			style.Direction = LayoutDirection(attr.Value)
		case "align":
			style.Align = Alignment(attr.Value)
		case "justify":
			style.Justify = Justify(attr.Value)
		case "gap":
			style.Gap = parseSize(attr.Value)
			style.GapSet = true
		case "padding":
			p := parseSize(attr.Value)
			style.Padding = Padding{p, p, p, p}
			style.PaddingSet = true
		case "margin":
			m := parseSize(attr.Value)
			style.Margin = Margin{m, m, m, m}
			style.MarginSet = true
		case "background", "bg":
			style.Background = attr.Value
			style.BackgroundColor = parseColor(attr.Value)
		case "border-color":
			style.Border = attr.Value
			style.BorderColor = parseColor(attr.Value)
		case "border-width":
			style.BorderWidth = parseSize(attr.Value)
			style.BorderWidthSet = true
		case "border-radius":
			style.BorderRadius = parseSize(attr.Value)
			style.BorderRadiusSet = true
		case "color":
			style.Color = attr.Value
			style.TextColor = parseColor(attr.Value)
		case "font-size":
			style.FontSize = parseSize(attr.Value)
			style.FontSizeSet = true
		case "line-height":
			style.LineHeight = parseSize(attr.Value)
			style.LineHeightSet = true
		case "letter-spacing":
			style.LetterSpacing = parseSize(attr.Value)
			style.LetterSpacingSet = true
		case "flex-grow", "grow":
			val, _ := strconv.ParseFloat(attr.Value, 64)
			style.FlexGrow = val
			style.FlexGrowSet = true
		case "flex-shrink", "shrink":
			val, _ := strconv.ParseFloat(attr.Value, 64)
			style.FlexShrink = val
			style.FlexShrinkSet = true
		case "opacity":
			val, _ := strconv.ParseFloat(attr.Value, 64)
			style.Opacity = val
			style.OpacitySet = true
		case "top":
			style.Top = parseSize(attr.Value)
			style.TopSet = true
		case "right":
			style.Right = parseSize(attr.Value)
			style.RightSet = true
		case "bottom":
			style.Bottom = parseSize(attr.Value)
			style.BottomSet = true
		case "left":
			style.Left = parseSize(attr.Value)
			style.LeftSet = true
		case "z-index", "zindex":
			val, _ := strconv.Atoi(attr.Value)
			style.ZIndex = val
			style.ZIndexSet = true
		case "min-width":
			style.MinWidth = parseSize(attr.Value)
			style.MinWidthSet = true
		case "min-height":
			style.MinHeight = parseSize(attr.Value)
			style.MinHeightSet = true
		case "max-width":
			style.MaxWidth = parseSize(attr.Value)
			style.MaxWidthSet = true
		case "max-height":
			style.MaxHeight = parseSize(attr.Value)
			style.MaxHeightSet = true
		case "box-sizing":
			style.BoxSizing = attr.Value
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
