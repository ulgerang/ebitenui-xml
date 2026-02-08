package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// UI is the main UI manager
type UI struct {
	root         Widget
	styleEngine  *StyleEngine
	layoutEngine *LayoutEngine
	factory      *WidgetFactory

	// Font for text rendering (use text.Face interface)
	DefaultFont     *text.GoTextFaceSource // For GoTextFace (TTF fonts)
	DefaultBoldFont *text.GoTextFaceSource // For bold weight font
	DefaultFontFace text.Face              // For GoXFace (bitmap fonts)

	// Font caches (lazily initialised)
	fontCache     *FontCache // regular weight
	boldFontCache *FontCache // bold weight

	// Dimensions
	width, height float64

	// State tracking
	hoveredWidget Widget
	activeWidget  Widget
	focusedWidget Widget

	// Widget lookup cache
	widgetByID map[string]Widget

	// CSS Variables
	variables *CSSVariables

	// Data binding
	bindings *BindingContext

	// Viewport for relative units
	viewportWidth  float64
	viewportHeight float64
	rootFontSize   float64
}

// New creates a new UI manager
func New(width, height float64) *UI {
	styleEngine := NewStyleEngine()
	return &UI{
		styleEngine:    styleEngine,
		layoutEngine:   NewLayoutEngine(),
		factory:        NewWidgetFactory(styleEngine),
		width:          width,
		height:         height,
		widgetByID:     make(map[string]Widget),
		variables:      NewCSSVariables(),
		bindings:       NewBindingContext(),
		viewportWidth:  width,
		viewportHeight: height,
		rootFontSize:   16,
	}
}

// LoadLayout loads a UI layout from XML
func (ui *UI) LoadLayout(xmlContent string) error {
	parser := NewXMLParser()
	node, err := parser.ParseString(xmlContent)
	if err != nil {
		return err
	}

	ui.root = ui.factory.CreateFromXML(node)
	ui.buildWidgetCache(ui.root)

	// Pipeline: styles → inherit → fonts → layout
	ui.reapplyStyles(ui.root)
	ui.inheritCSSProperties(ui.root, nil)
	ui.setFonts(ui.root)

	// Initial layout
	ui.Layout()

	return nil
}

// LoadLayoutFile loads a UI layout from an XML file
func (ui *UI) LoadLayoutFile(filename string) error {
	parser := NewXMLParser()
	node, err := parser.ParseFile(filename)
	if err != nil {
		return err
	}

	ui.root = ui.factory.CreateFromXML(node)
	ui.buildWidgetCache(ui.root)

	// Pipeline: styles → inherit → fonts → layout
	ui.reapplyStyles(ui.root)
	ui.inheritCSSProperties(ui.root, nil)
	ui.setFonts(ui.root)
	ui.Layout()

	return nil
}

// LoadStyles loads styles from JSON
func (ui *UI) LoadStyles(jsonContent string) error {
	if err := ui.styleEngine.LoadFromString(jsonContent); err != nil {
		return err
	}

	// Pipeline: styles → inherit → fonts → layout
	if ui.root != nil {
		ui.reapplyStyles(ui.root)
		ui.inheritCSSProperties(ui.root, nil)
		ui.setFonts(ui.root)
		ui.Layout()
	}

	return nil
}

// LoadStylesFile loads styles from a JSON file
func (ui *UI) LoadStylesFile(filename string) error {
	if err := ui.styleEngine.LoadFromFile(filename); err != nil {
		return err
	}

	// Pipeline: styles → inherit → fonts → layout
	if ui.root != nil {
		ui.reapplyStyles(ui.root)
		ui.inheritCSSProperties(ui.root, nil)
		ui.setFonts(ui.root)
		ui.Layout()
	}

	return nil
}

// Layout recalculates the layout
func (ui *UI) Layout() {
	if ui.root != nil {
		ui.layoutEngine.Layout(ui.root, ui.width, ui.height)
	}
}

// Resize updates the UI dimensions
func (ui *UI) Resize(width, height float64) {
	ui.width = width
	ui.height = height
	ui.Layout()
}

// Update handles input and updates UI state
func (ui *UI) Update() {
	if ui.root == nil {
		return
	}

	// Get mouse position
	mx, my := ebiten.CursorPosition()
	mouseX, mouseY := float64(mx), float64(my)

	// Find widget under cursor
	hoveredWidget := ui.findWidgetAt(ui.root, mouseX, mouseY)

	// Handle hover state changes
	if hoveredWidget != ui.hoveredWidget {
		// Leave previous
		if ui.hoveredWidget != nil {
			ui.hoveredWidget.SetState(StateNormal)
		}
		// Enter new
		if hoveredWidget != nil {
			hoveredWidget.SetState(StateHover)
			if bw, ok := hoveredWidget.(*Button); ok {
				bw.HandleHover()
			}
		}
		ui.hoveredWidget = hoveredWidget
	}

	// Handle clicks
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if hoveredWidget != nil {
			hoveredWidget.SetState(StateActive)
			ui.activeWidget = hoveredWidget
		}
	}

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		if ui.activeWidget != nil {
			if ui.activeWidget == hoveredWidget {
				// Click confirmed - handle all widget types
				switch w := ui.activeWidget.(type) {
				case *Button:
					w.HandleClick()
				case *Panel:
					w.HandleClick()
				case *Toggle:
					w.HandleClick()
				case *RadioButton:
					w.HandleClick()
				case *Dropdown:
					w.HandleClick()
				case *Checkbox:
					w.HandleClick()
				case *Slider:
					w.HandleClick()
				}
			}
			ui.activeWidget.SetState(StateNormal)
			if hoveredWidget != nil {
				hoveredWidget.SetState(StateHover)
			}
			ui.activeWidget = nil
		}
	}
}

// Draw renders the UI
func (ui *UI) Draw(screen *ebiten.Image) {
	if ui.root != nil {
		ui.root.Draw(screen)
	}
}

// GetWidget returns a widget by ID
func (ui *UI) GetWidget(id string) Widget {
	return ui.widgetByID[id]
}

// GetButton returns a button by ID
func (ui *UI) GetButton(id string) *Button {
	w := ui.widgetByID[id]
	if btn, ok := w.(*Button); ok {
		return btn
	}
	return nil
}

// GetPanel returns a panel by ID
func (ui *UI) GetPanel(id string) *Panel {
	w := ui.widgetByID[id]
	if p, ok := w.(*Panel); ok {
		return p
	}
	return nil
}

// GetProgressBar returns a progress bar by ID
func (ui *UI) GetProgressBar(id string) *ProgressBar {
	w := ui.widgetByID[id]
	if pb, ok := w.(*ProgressBar); ok {
		return pb
	}
	return nil
}

// GetTextInput returns a text input by ID
func (ui *UI) GetTextInput(id string) *TextInput {
	w := ui.widgetByID[id]
	if ti, ok := w.(*TextInput); ok {
		return ti
	}
	return nil
}

// GetTextArea returns a text area by ID
func (ui *UI) GetTextArea(id string) *TextArea {
	w := ui.widgetByID[id]
	if ta, ok := w.(*TextArea); ok {
		return ta
	}
	return nil
}

// GetScrollable returns a scrollable container by ID
func (ui *UI) GetScrollable(id string) *Scrollable {
	w := ui.widgetByID[id]
	if s, ok := w.(*Scrollable); ok {
		return s
	}
	return nil
}

// GetCheckbox returns a checkbox by ID
func (ui *UI) GetCheckbox(id string) *Checkbox {
	w := ui.widgetByID[id]
	if cb, ok := w.(*Checkbox); ok {
		return cb
	}
	return nil
}

// GetSlider returns a slider by ID
func (ui *UI) GetSlider(id string) *Slider {
	w := ui.widgetByID[id]
	if s, ok := w.(*Slider); ok {
		return s
	}
	return nil
}

// GetText returns a text widget by ID
func (ui *UI) GetText(id string) *Text {
	w := ui.widgetByID[id]
	if t, ok := w.(*Text); ok {
		return t
	}
	return nil
}

// ============================================================================
// CSS Variables
// ============================================================================

// SetVariable sets a CSS variable
func (ui *UI) SetVariable(name, value string) {
	ui.variables.Set(name, value)
}

// GetVariable gets a CSS variable
func (ui *UI) GetVariable(name string) string {
	return ui.variables.Get(name)
}

// Variables returns the CSS variables container
func (ui *UI) Variables() *CSSVariables {
	return ui.variables
}

// ============================================================================
// Data Binding
// ============================================================================

// Bindings returns the binding context
func (ui *UI) Bindings() *BindingContext {
	return ui.bindings
}

// Bind sets a binding value
func (ui *UI) Bind(key string, value interface{}) {
	ui.bindings.Set(key, value)
}

// BindText binds a value to a text widget
func (ui *UI) BindText(key, widgetID string) {
	if tw := ui.GetText(widgetID); tw != nil {
		ui.bindings.BindText(key, tw)
	}
}

// BindProgress binds a value to a progress bar
func (ui *UI) BindProgress(key, widgetID string) {
	if pb := ui.GetProgressBar(widgetID); pb != nil {
		ui.bindings.BindProgress(key, pb)
	}
}

// BindVisible binds visibility to a widget
func (ui *UI) BindVisible(key, widgetID string) {
	if w := ui.GetWidget(widgetID); w != nil {
		ui.bindings.BindVisible(key, w)
	}
}

// ============================================================================
// Focus Management
// ============================================================================

// Focus sets focus to a widget
func (ui *UI) Focus(id string) {
	w := ui.widgetByID[id]
	if w == nil {
		return
	}

	// Blur previous
	if ui.focusedWidget != nil {
		ui.focusedWidget.SetState(StateNormal)
		if ti, ok := ui.focusedWidget.(*TextInput); ok {
			ti.Blur()
		}
		if ta, ok := ui.focusedWidget.(*TextArea); ok {
			ta.Blur()
		}
	}

	// Focus new
	ui.focusedWidget = w
	w.SetState(StateFocused)
	if ti, ok := w.(*TextInput); ok {
		ti.Focus()
	}
	if ta, ok := w.(*TextArea); ok {
		ta.Focus()
	}
}

// Blur removes focus from the current focused widget
func (ui *UI) Blur() {
	if ui.focusedWidget != nil {
		ui.focusedWidget.SetState(StateNormal)
		if ti, ok := ui.focusedWidget.(*TextInput); ok {
			ti.Blur()
		}
		if ta, ok := ui.focusedWidget.(*TextArea); ok {
			ta.Blur()
		}
		ui.focusedWidget = nil
	}
}

// FocusedWidget returns the currently focused widget
func (ui *UI) FocusedWidget() Widget {
	return ui.focusedWidget
}

// buildWidgetCache builds a map of widgets by ID
func (ui *UI) buildWidgetCache(widget Widget) {
	if widget == nil {
		return
	}
	if widget.ID() != "" {
		ui.widgetByID[widget.ID()] = widget
	}
	for _, child := range widget.Children() {
		ui.buildWidgetCache(child)
	}
}

// findWidgetAt finds the deepest widget at a given position
func (ui *UI) findWidgetAt(widget Widget, x, y float64) Widget {
	if widget == nil || !widget.Visible() {
		return nil
	}

	// Check children first (in reverse order for proper z-order)
	children := widget.Children()
	for i := len(children) - 1; i >= 0; i-- {
		found := ui.findWidgetAt(children[i], x, y)
		if found != nil {
			return found
		}
	}

	// Check this widget
	if widget.ComputedRect().Contains(x, y) {
		return widget
	}

	return nil
}

// setFonts sets font faces on text-bearing widgets based on their style properties.
// Respects fontWeight (bold) and fontSize via FontCache for efficient face reuse.
func (ui *UI) setFonts(widget Widget) {
	if widget == nil {
		return
	}

	// Lazily create font caches
	if ui.fontCache == nil && ui.DefaultFont != nil {
		ui.fontCache = NewFontCache(ui.DefaultFont)
	}
	if ui.boldFontCache == nil && ui.DefaultBoldFont != nil {
		ui.boldFontCache = NewFontCache(ui.DefaultBoldFont)
	}

	// Determine font face for this widget
	var fontFace text.Face
	if ui.DefaultFontFace != nil {
		// Bitmap font — can't vary size, use as-is
		fontFace = ui.DefaultFontFace
	} else {
		style := widget.Style()
		fontSize := style.FontSize
		if fontSize <= 0 {
			fontSize = 14
		}

		// Select bold or regular font source
		isBold := style.FontWeight == "bold" || style.FontWeight == "700" ||
			style.FontWeight == "800" || style.FontWeight == "900"
		if isBold && ui.boldFontCache != nil {
			fontFace = ui.boldFontCache.GetFace(fontSize)
		} else if ui.fontCache != nil {
			fontFace = ui.fontCache.GetFace(fontSize)
		}
	}

	// Apply to text-bearing widgets
	if fontFace != nil {
		switch w := widget.(type) {
		case *Button:
			w.FontFace = fontFace
		case *Text:
			w.FontFace = fontFace
		case *Toggle:
			w.FontFace = fontFace
		case *RadioButton:
			w.FontFace = fontFace
		case *Dropdown:
			w.FontFace = fontFace
		case *Badge:
			w.FontFace = fontFace
		case *Spinner:
			// Spinner doesn't need text
		case *Toast:
			w.FontFace = fontFace
		case *Modal:
			w.FontFace = fontFace
			w.TitleFontFace = fontFace
		case *Tooltip:
			w.FontFace = fontFace
		}
	}

	for _, child := range widget.Children() {
		ui.setFonts(child)
	}
}

// reapplyStyles reapplies styles from the style engine
func (ui *UI) reapplyStyles(widget Widget) {
	if widget == nil {
		return
	}

	// Apply by type (e.g., "panel", "svg", "button")
	ui.styleEngine.ApplyStyle(widget, widget.Type())

	// Apply by class (e.g., ".icon-box", ".icon-row")
	for _, class := range widget.Classes() {
		ui.styleEngine.ApplyStyle(widget, "."+class)
	}

	// Apply by ID (e.g., "#root", "#header")
	if widget.ID() != "" {
		ui.styleEngine.ApplyStyle(widget, "#"+widget.ID())
	}

	// Recursively apply to children
	for _, child := range widget.Children() {
		ui.reapplyStyles(child)
	}
}

// inheritCSSProperties inherits CSS-inheritable properties from parent to child.
// Properties like color, font-size, font-weight, text-align, vertical-align, etc.
// cascade down the widget tree following CSS inheritance rules: if the child has
// no explicit value, it inherits from its nearest ancestor that does.
func (ui *UI) inheritCSSProperties(widget Widget, parentStyle *Style) {
	if widget == nil {
		return
	}

	style := widget.Style()
	if parentStyle != nil {
		// Color / TextColor
		if style.Color == "" && parentStyle.Color != "" {
			style.Color = parentStyle.Color
		}
		if style.TextColor == nil && parentStyle.TextColor != nil {
			style.TextColor = parentStyle.TextColor
		}

		// Font properties
		if style.FontSize == 0 && parentStyle.FontSize != 0 {
			style.FontSize = parentStyle.FontSize
		}
		if style.FontWeight == "" && parentStyle.FontWeight != "" {
			style.FontWeight = parentStyle.FontWeight
		}
		if style.FontStyle == "" && parentStyle.FontStyle != "" {
			style.FontStyle = parentStyle.FontStyle
		}

		// Text layout
		if style.TextAlign == "" && parentStyle.TextAlign != "" {
			style.TextAlign = parentStyle.TextAlign
		}
		if style.VerticalAlign == "" && parentStyle.VerticalAlign != "" {
			style.VerticalAlign = parentStyle.VerticalAlign
		}
		if style.LineHeight == 0 && parentStyle.LineHeight != 0 {
			style.LineHeight = parentStyle.LineHeight
		}
		if style.LetterSpacing == 0 && parentStyle.LetterSpacing != 0 {
			style.LetterSpacing = parentStyle.LetterSpacing
		}
	}

	for _, child := range widget.Children() {
		ui.inheritCSSProperties(child, style)
	}
}
