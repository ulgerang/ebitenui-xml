package ui

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

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
	fontCache       *FontCache // regular weight
	boldFontCache   *FontCache // bold weight
	fontFaces       map[string]text.Face
	fontSources     map[string]*FontCache
	boldFontSources map[string]*FontCache

	// Dimensions
	width, height float64

	// State tracking
	hoveredWidget Widget
	activeWidget  Widget
	focusedWidget Widget
	activeModal   *Modal
	modalRestore  Widget
	modalStack    []modalFocusState

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

type modalFocusState struct {
	modal   *Modal
	restore Widget
}

// New creates a new UI manager
func New(width, height float64) *UI {
	styleEngine := NewStyleEngine()
	bindings := NewBindingContext()
	manager := &UI{
		styleEngine:     styleEngine,
		layoutEngine:    NewLayoutEngine(),
		factory:         NewWidgetFactory(styleEngine, bindings),
		width:           width,
		height:          height,
		widgetByID:      make(map[string]Widget),
		variables:       NewCSSVariables(),
		bindings:        bindings,
		rootFontSize:    16, // Default browser root font size
		fontFaces:       make(map[string]text.Face),
		fontSources:     make(map[string]*FontCache),
		boldFontSources: make(map[string]*FontCache),
	}
	manager.factory.onTreeChanged = manager.refreshDynamicTree
	manager.factory.onLayoutChanged = manager.refreshDynamicLayout
	return manager
}

// Root returns the root widget
func (ui *UI) Root() Widget { return ui.root }

// RegisterFontFace registers a fixed-size font face for a CSS font-family name.
func (ui *UI) RegisterFontFace(family string, face text.Face) {
	name := normalizeFontFamilyName(family)
	if name == "" || face == nil {
		return
	}
	ui.fontFaces[name] = face
}

// RegisterFontSource registers a scalable Go text font source for a CSS font-family name.
func (ui *UI) RegisterFontSource(family string, source *text.GoTextFaceSource) {
	name := normalizeFontFamilyName(family)
	if name == "" || source == nil {
		return
	}
	ui.fontSources[name] = NewFontCache(source)
}

// RegisterBoldFontSource registers a scalable bold source for a CSS font-family name.
func (ui *UI) RegisterBoldFontSource(family string, source *text.GoTextFaceSource) {
	name := normalizeFontFamilyName(family)
	if name == "" || source == nil {
		return
	}
	ui.boldFontSources[name] = NewFontCache(source)
}

// SetRoot sets the root widget directly and runs the normal style/font/layout pipeline.
func (ui *UI) SetRoot(widget Widget) {
	ui.setRoot(widget)
}

// LoadLayout loads a UI layout from XML
func (ui *UI) LoadLayout(xmlContent string) error {
	parser := NewXMLParser()
	node, err := parser.ParseString(xmlContent)
	if err != nil {
		return err
	}

	ui.setRoot(ui.factory.CreateFromXML(node))

	// Pipeline: styles ??inherit ??fonts ??layout

	// Initial layout

	return nil
}

// LoadLayoutFile loads a UI layout from an XML file
func (ui *UI) LoadLayoutFile(filename string) error {
	parser := NewXMLParser()
	node, err := parser.ParseFile(filename)
	if err != nil {
		return err
	}

	ui.setRoot(ui.factory.CreateFromXML(node))

	// Pipeline: styles ??inherit ??fonts ??layout

	return nil
}

// LoadStyles loads styles from JSON
func (ui *UI) LoadStyles(jsonContent string) error {
	if err := ui.styleEngine.LoadFromString(ui.variables.Resolve(jsonContent)); err != nil {
		return err
	}

	// Pipeline: styles ??inherit ??fonts ??layout
	if ui.root != nil {
		ui.reapplyStyles(ui.root)
		ui.inheritCSSProperties(ui.root, nil)
		ui.setFonts(ui.root)
		ui.Layout()
	}

	return nil
}

// LoadCSS loads literal CSS subset content such as @keyframes blocks.
func (ui *UI) LoadCSS(cssContent string) error {
	css := ui.variables.Resolve(cssContent)
	css = ui.expandCSSMediaQueries(css)
	if err := ui.styleEngine.LoadCSS(css); err != nil {
		return err
	}
	if ui.root != nil {
		ui.reapplyStyles(ui.root)
		ui.inheritCSSProperties(ui.root, nil)
		ui.setFonts(ui.root)
		ui.Layout()
	}
	return nil
}

func (ui *UI) expandCSSMediaQueries(css string) string {
	var out strings.Builder
	remaining := css
	for {
		idx := strings.Index(remaining, "@media")
		if idx < 0 {
			out.WriteString(remaining)
			break
		}
		out.WriteString(remaining[:idx])
		media := remaining[idx:]
		open := strings.Index(media, "{")
		if open < 0 {
			out.WriteString(media)
			break
		}
		condition := strings.TrimSpace(strings.TrimPrefix(media[:open], "@media"))
		closeIdx := matchingCSSBlockEnd(media, open)
		if closeIdx < 0 {
			out.WriteString(media)
			break
		}
		body := media[open+1 : closeIdx]
		if ui.matchesCSSMediaCondition(condition) {
			out.WriteString(body)
		}
		remaining = media[closeIdx+1:]
	}
	return out.String()
}

func (ui *UI) matchesCSSMediaCondition(condition string) bool {
	for _, query := range splitCSSMediaQueryList(condition) {
		if ui.matchesCSSMediaQuery(query) {
			return true
		}
	}
	return false
}

func (ui *UI) matchesCSSMediaQuery(query string) bool {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return false
	}

	parts := regexp.MustCompile(`\s+and\s+`).Split(query, -1)
	hasFeature := false
	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return false
		}
		if i == 0 && !strings.HasPrefix(part, "(") {
			switch part {
			case "all", "screen":
				continue
			default:
				return false
			}
		}
		hasFeature = true
		if !ui.matchesCSSMediaFeature(part) {
			return false
		}
	}
	return hasFeature || parts[0] == "all" || parts[0] == "screen"
}

func (ui *UI) matchesCSSMediaFeature(feature string) bool {
	if match := regexp.MustCompile(`^\(\s*(min|max)-(width|height)\s*:\s*([0-9.]+)px\s*\)$`).FindStringSubmatch(feature); match != nil {
		value, err := strconv.ParseFloat(match[3], 64)
		if err != nil {
			return false
		}
		actual := ui.width
		if match[2] == "height" {
			actual = ui.height
		}
		switch match[1] {
		case "min":
			return actual >= value
		case "max":
			return actual <= value
		}
	}

	if match := regexp.MustCompile(`^\(\s*orientation\s*:\s*(landscape|portrait)\s*\)$`).FindStringSubmatch(feature); match != nil {
		orientation := "portrait"
		if ui.width >= ui.height {
			orientation = "landscape"
		}
		return orientation == match[1]
	}

	return false
}

func splitCSSMediaQueryList(condition string) []string {
	var queries []string
	start := 0
	depth := 0
	for i := 0; i < len(condition); i++ {
		switch condition[i] {
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		case ',':
			if depth == 0 {
				queries = append(queries, condition[start:i])
				start = i + 1
			}
		}
	}
	return append(queries, condition[start:])
}

func matchingCSSBlockEnd(s string, open int) int {
	depth := 0
	for i := open; i < len(s); i++ {
		switch s[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

// LoadStylesFile loads styles from a JSON file
func (ui *UI) LoadStylesFile(filename string) error {
	if err := ui.styleEngine.LoadFromFile(filename); err != nil {
		return err
	}

	// Pipeline: styles ??inherit ??fonts ??layout
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
	ui.syncModalFocusState()

	// FIX: Reset input state at start of each frame to allow fresh input consumption
	// This is required for consumeFrameInput() in input.go to work correctly
	resetInputForFrame()

	// Get mouse position
	mx, my := ebiten.CursorPosition()
	mouseX, mouseY := float64(mx), float64(my)

	// Find widget under cursor
	hoveredWidget := ui.findWidgetAt(ui.root, mouseX, mouseY)
	if _, wheelY := ebiten.Wheel(); wheelY != 0 {
		ui.scrollHoveredWidget(hoveredWidget, 0, -wheelY*40)
	}

	// Handle hover state changes
	if hoveredWidget != ui.hoveredWidget {
		// Leave previous
		if ui.hoveredWidget != nil {
			if txt, ok := ui.hoveredWidget.(*Text); ok {
				txt.ClearHoveredCluster()
			}
			if ui.hoveredWidget == ui.focusedWidget {
				ui.hoveredWidget.SetState(StateFocused)
			} else {
				ui.hoveredWidget.SetState(StateNormal)
			}
		}
		// Enter new
		if hoveredWidget != nil {
			if hoveredWidget != ui.focusedWidget {
				hoveredWidget.SetState(StateHover)
				if bw, ok := hoveredWidget.(*Button); ok {
					bw.HandleHover()
				}
			}
		}
		ui.hoveredWidget = hoveredWidget
	}
	if txt, ok := hoveredWidget.(*Text); ok {
		txt.HandlePointerMove(mouseX, mouseY)
	}

	// Handle clicks
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if hoveredWidget != nil {
			switch w := hoveredWidget.(type) {
			case *TextInput:
				ui.setFocusedWidget(hoveredWidget)
				w.HandlePointerDown(mouseX, mouseY)
			case *TextArea:
				ui.setFocusedWidget(hoveredWidget)
				w.HandlePointerDown(mouseX, mouseY)
			default:
				if ui.focusedWidget != nil && ui.focusedWidget != hoveredWidget {
					ui.Blur()
				}
			}
			hoveredWidget.SetState(StateActive)
			ui.activeWidget = hoveredWidget
		} else {
			ui.Blur()
		}
	}

	if ui.activeWidget != nil && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		if slider, ok := ui.activeWidget.(*Slider); ok {
			mx, _ := ebiten.CursorPosition()
			rect := slider.ComputedRect()
			// FIX: Convert absolute screen coordinates to widget-relative coordinates
			// Same bug as HandleClick - must subtract widget X position
			slider.setValueFromCursor(float64(mx) - rect.X)
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
			if ui.activeWidget == ui.focusedWidget {
				ui.activeWidget.SetState(StateFocused)
			} else {
				ui.activeWidget.SetState(StateNormal)
			}
			if hoveredWidget != nil {
				if hoveredWidget == ui.focusedWidget {
					hoveredWidget.SetState(StateFocused)
				} else {
					hoveredWidget.SetState(StateHover)
				}
			}
			ui.activeWidget = nil
		}
	}

	ui.handleRuntimeKeyboard()
}

// SimulatePointerMove updates hover state as if the pointer moved.
func (ui *UI) SimulatePointerMove(x, y float64) Widget {
	return ui.handlePointerMove(x, y)
}

// SimulatePointerDown presses a pointer button at the given coordinates.
func (ui *UI) SimulatePointerDown(x, y float64, button ebiten.MouseButton) Widget {
	return ui.handlePointerDown(x, y, button)
}

// SimulatePointerUp releases a pointer button at the given coordinates.
func (ui *UI) SimulatePointerUp(x, y float64, button ebiten.MouseButton) Widget {
	hovered := ui.handlePointerMove(x, y)
	ui.handlePointerUp(x, y, button, hovered)
	return hovered
}

// SimulateClick performs a full left-click gesture at the given coordinates.
func (ui *UI) SimulateClick(x, y float64) Widget {
	hovered := ui.SimulatePointerDown(x, y, ebiten.MouseButtonLeft)
	ui.handlePointerUp(x, y, ebiten.MouseButtonLeft, hovered)
	return hovered
}

// SimulateTypeText inserts text into the currently focused text input or text area.
func (ui *UI) SimulateTypeText(s string) {
	switch w := ui.focusedWidget.(type) {
	case *TextInput:
		for _, r := range s {
			w.insertChar(r)
		}
	case *TextArea:
		for _, r := range s {
			w.insertChar(r)
		}
	}
}

// SimulateKeyPress dispatches a keyboard event to the focused widget.
func (ui *UI) SimulateKeyPress(key ebiten.Key, shift, control bool) {
	ui.handleKeyPress(key, shift, control)
}

func (ui *UI) handleRuntimeKeyboard() {
	shift := ebiten.IsKeyPressed(ebiten.KeyShift)
	control := ebiten.IsKeyPressed(ebiten.KeyControl)
	for _, key := range []ebiten.Key{
		ebiten.KeyTab,
		ebiten.KeyUp,
		ebiten.KeyDown,
		ebiten.KeyLeft,
		ebiten.KeyRight,
		ebiten.KeyHome,
		ebiten.KeyEnd,
		ebiten.KeySpace,
		ebiten.KeyEnter,
		ebiten.KeyNumpadEnter,
		ebiten.KeyEscape,
	} {
		if inpututil.IsKeyJustPressed(key) {
			ui.handleKeyPress(key, shift, control)
		}
	}
}

func (ui *UI) handleKeyPress(key ebiten.Key, shift, control bool) {
	ui.syncModalFocusState()
	if key == ebiten.KeyTab {
		ui.FocusNext(shift)
		return
	}
	if key == ebiten.KeyEscape {
		if modal := ui.openModal(); modal != nil {
			modal.Close()
			ui.restoreModalFocus()
			return
		}
		if dropdown, ok := ui.focusedWidget.(*Dropdown); ok && dropdown.IsOpen {
			dropdown.Close()
			return
		}
	}
	switch w := ui.focusedWidget.(type) {
	case *TextInput:
		simulateTextInputKeyPress(w, key, shift, control)
		if key == ebiten.KeyEnter || key == ebiten.KeyNumpadEnter {
			ui.submitNearestForm(w)
		}
	case *TextArea:
		simulateTextAreaKeyPress(w, key, shift, control)
	case *Button:
		if key == ebiten.KeySpace || key == ebiten.KeyEnter || key == ebiten.KeyNumpadEnter {
			w.HandleClick()
		}
	case *Checkbox:
		if key == ebiten.KeySpace {
			w.HandleClick()
		}
	case *Toggle:
		if key == ebiten.KeySpace {
			w.HandleClick()
		}
	case *Dropdown:
		switch key {
		case ebiten.KeyDown, ebiten.KeyRight:
			w.MoveHighlight(1)
		case ebiten.KeyUp, ebiten.KeyLeft:
			w.MoveHighlight(-1)
		case ebiten.KeyEnter, ebiten.KeyNumpadEnter:
			w.SelectHighlighted()
		}
	case *RadioButton:
		if w.Group == nil {
			return
		}
		switch key {
		case ebiten.KeySpace:
			w.HandleClick()
		case ebiten.KeyDown, ebiten.KeyRight:
			if selected := w.Group.MoveSelectionFrom(w, 1); selected != nil {
				ui.setFocusedWidget(selected)
			}
		case ebiten.KeyUp, ebiten.KeyLeft:
			if selected := w.Group.MoveSelectionFrom(w, -1); selected != nil {
				ui.setFocusedWidget(selected)
			}
		}
	case *Slider:
		switch key {
		case ebiten.KeyRight, ebiten.KeyUp:
			w.Increment(1)
		case ebiten.KeyLeft, ebiten.KeyDown:
			w.Increment(-1)
		case ebiten.KeyHome:
			w.SetValue(w.Min)
		case ebiten.KeyEnd:
			w.SetValue(w.Max)
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

// QueryByClass returns widgets that have the given CSS class.
func (ui *UI) QueryByClass(class string) []Widget {
	return ui.Query(ui.root, "."+class)
}

// QueryByType returns widgets matching a concrete widget type or semantic XML tag.
func (ui *UI) QueryByType(typeName string) []Widget {
	return ui.Query(ui.root, typeName)
}

// Query returns widgets below root matching simple #id, .class, or type selectors.
func (ui *UI) Query(root Widget, selector string) []Widget {
	selector = strings.TrimSpace(selector)
	if root == nil || selector == "" {
		return nil
	}
	var matches []Widget
	var walk func(Widget)
	walk = func(widget Widget) {
		if widget == nil {
			return
		}
		if matchesSelector(widget, selector) {
			matches = append(matches, widget)
		}
		for _, child := range widget.Children() {
			walk(child)
		}
	}
	walk(root)
	return matches
}

func matchesSelector(widget Widget, selector string) bool {
	switch {
	case strings.HasPrefix(selector, "#"):
		return widget.ID() == strings.TrimPrefix(selector, "#")
	case strings.HasPrefix(selector, "."):
		return widget.HasClass(strings.TrimPrefix(selector, "."))
	default:
		if widget.Type() == selector {
			return true
		}
		if bw := baseWidgetOf(widget); bw != nil && bw.SemanticType() == selector {
			return true
		}
		return widget.HasClass(selector)
	}
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

// SubmitForm dispatches a form submit command for the form widget with the given ID.
func (ui *UI) SubmitForm(id string) {
	form := ui.GetWidget(id)
	if form == nil || ui.factory == nil {
		return
	}
	if !ui.validateFormWidget(form) {
		return
	}
	if command := ui.factory.formSubmit[form]; command != "" {
		ui.factory.runCommand(command, form)
	}
}

// ValidateForm validates all descendants of a form widget by ID.
func (ui *UI) ValidateForm(id string) bool {
	return ui.validateFormWidget(ui.GetWidget(id))
}

func (ui *UI) validateFormWidget(form Widget) bool {
	if form == nil {
		return false
	}
	valid := true
	var walk func(Widget)
	walk = func(widget Widget) {
		if widget == nil {
			return
		}
		if !validateSingleWidget(widget) {
			valid = false
		}
		for _, child := range widget.Children() {
			walk(child)
		}
	}
	walk(form)
	return valid
}

func (ui *UI) submitNearestForm(widget Widget) {
	for current := widget; current != nil; current = current.Parent() {
		if bw := baseWidgetOf(current); bw != nil && bw.SemanticType() == "form" {
			ui.SubmitForm(current.ID())
			return
		}
	}
}

// ResetForm resets form descendants to their initial XML values and dispatches onReset.
func (ui *UI) ResetForm(id string) {
	form := ui.GetWidget(id)
	if form == nil {
		return
	}
	resetFormSubtree(form)
	if ui.factory != nil {
		if command := ui.factory.formReset[form]; command != "" {
			ui.factory.runCommand(command, form)
		}
	}
	ui.refreshDynamicLayout()
}

// SetValidationState updates a widget's form validation state by ID.
func (ui *UI) SetValidationState(id string, state ValidationState) {
	if bw := baseWidgetOf(ui.GetWidget(id)); bw != nil {
		bw.SetValidationState(state)
	}
}

// GetValidationState returns a widget's form validation state by ID.
func (ui *UI) GetValidationState(id string) ValidationState {
	if bw := baseWidgetOf(ui.GetWidget(id)); bw != nil {
		return bw.ValidationState()
	}
	return ValidationNone
}

// GetValidationMessage returns the latest validation message for a widget by ID.
func (ui *UI) GetValidationMessage(id string) string {
	if bw := baseWidgetOf(ui.GetWidget(id)); bw != nil {
		return bw.ValidationMessage()
	}
	return ""
}

// ScrollWidgetBy scrolls an overflow scroll/auto widget or Scrollable by ID.
func (ui *UI) ScrollWidgetBy(id string, dx, dy float64) {
	switch w := ui.GetWidget(id).(type) {
	case *Scrollable:
		w.ScrollBy(dx, dy)
	default:
		if bw := baseWidgetOf(w); bw != nil {
			bw.ScrollBy(dx, dy)
		}
	}
}

// SetWidgetScroll sets an overflow scroll/auto widget or Scrollable scroll offset by ID.
func (ui *UI) SetWidgetScroll(id string, x, y float64) {
	switch w := ui.GetWidget(id).(type) {
	case *Scrollable:
		w.ScrollTo(x, y)
	default:
		if bw := baseWidgetOf(w); bw != nil {
			bw.SetScrollOffset(x, y)
		}
	}
}

// RegisterCommand registers a named command handler for XML event attributes
// such as onClick, onChange, and onSubmit.
func (ui *UI) RegisterCommand(name string, handler func(Widget)) {
	if ui.factory == nil || name == "" {
		return
	}
	if ui.factory.commands == nil {
		ui.factory.commands = make(map[string]func(Widget))
	}
	ui.factory.commands[name] = handler
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
	ui.setFocusedWidget(w)
}

// FocusNext moves focus to the next or previous focusable widget.
func (ui *UI) FocusNext(reverse bool) Widget {
	ui.syncModalFocusState()
	focusables := ui.focusableWidgets()
	if len(focusables) == 0 {
		ui.Blur()
		return nil
	}
	current := -1
	for i, widget := range focusables {
		if widget == ui.focusedWidget {
			current = i
			break
		}
	}
	next := 0
	if reverse {
		next = len(focusables) - 1
		if current >= 0 {
			next = (current - 1 + len(focusables)) % len(focusables)
		}
	} else if current >= 0 {
		next = (current + 1) % len(focusables)
	}
	ui.setFocusedWidget(focusables[next])
	return focusables[next]
}

// Blur removes focus from the current focused widget
func (ui *UI) Blur() {
	ui.setFocusedWidget(nil)
}

// FocusedWidget returns the currently focused widget
func (ui *UI) FocusedWidget() Widget {
	return ui.focusedWidget
}

func (ui *UI) focusableWidgets() []Widget {
	var widgets []Widget
	root := ui.root
	if modal := ui.openModal(); modal != nil {
		root = modal
	}
	var walk func(Widget)
	walk = func(widget Widget) {
		if widget == nil || !widget.Visible() || !widget.Enabled() {
			return
		}
		if bw := baseWidgetOf(widget); bw != nil && bw.Focusable() && bw.TabIndex() != -1 {
			widgets = append(widgets, widget)
		}
		for _, child := range widget.Children() {
			walk(child)
		}
	}
	walk(root)
	sort.SliceStable(widgets, func(i, j int) bool {
		ai := tabIndexOf(widgets[i])
		aj := tabIndexOf(widgets[j])
		if ai < 0 && aj < 0 {
			return false
		}
		if ai < 0 {
			return false
		}
		if aj < 0 {
			return true
		}
		return ai < aj
	})
	return widgets
}

func (ui *UI) syncModalFocusState() {
	modal := ui.openModal()
	if modal == nil {
		if len(ui.modalStack) > 0 {
			ui.restoreAllModalFocus()
		}
		return
	}

	if len(ui.modalStack) == 0 || ui.modalStack[len(ui.modalStack)-1].modal != modal {
		if idx := ui.modalStackIndex(modal); idx >= 0 {
			for len(ui.modalStack)-1 > idx {
				ui.restoreModalFocus()
			}
		} else {
			restore := Widget(nil)
			if ui.focusedWidget != nil && !isDescendantOf(ui.focusedWidget, modal) {
				restore = ui.focusedWidget
			}
			ui.modalStack = append(ui.modalStack, modalFocusState{modal: modal, restore: restore})
			ui.activeModal = modal
			ui.modalRestore = restore
		}
	}

	if ui.focusedWidget != nil && !isDescendantOf(ui.focusedWidget, modal) {
		ui.setFocusedWidget(nil)
	}
}

func (ui *UI) restoreModalFocus() {
	if len(ui.modalStack) == 0 {
		ui.activeModal = nil
		ui.modalRestore = nil
		ui.Blur()
		return
	}

	state := ui.modalStack[len(ui.modalStack)-1]
	ui.modalStack = ui.modalStack[:len(ui.modalStack)-1]

	ui.activeModal = nil
	ui.modalRestore = nil
	if len(ui.modalStack) > 0 {
		previous := ui.modalStack[len(ui.modalStack)-1]
		ui.activeModal = previous.modal
		ui.modalRestore = previous.restore
	}

	restore := state.restore
	if restore != nil && restore.Visible() && restore.Enabled() {
		ui.setFocusedWidget(restore)
		return
	}
	ui.Blur()
}

func (ui *UI) restoreAllModalFocus() {
	var restore Widget
	for len(ui.modalStack) > 0 {
		state := ui.modalStack[len(ui.modalStack)-1]
		ui.modalStack = ui.modalStack[:len(ui.modalStack)-1]
		if state.restore != nil && state.restore.Visible() && state.restore.Enabled() {
			restore = state.restore
		}
	}
	ui.activeModal = nil
	ui.modalRestore = nil
	if restore != nil {
		ui.setFocusedWidget(restore)
		return
	}
	ui.Blur()
}

func (ui *UI) modalStackIndex(modal *Modal) int {
	for i := len(ui.modalStack) - 1; i >= 0; i-- {
		if ui.modalStack[i].modal == modal {
			return i
		}
	}
	return -1
}

func (ui *UI) openModal() *Modal {
	var found *Modal
	var walk func(Widget)
	walk = func(widget Widget) {
		if widget == nil || found != nil {
			return
		}
		for _, child := range sortedChildrenByZ(widget.Children(), true) {
			walk(child)
		}
		if found != nil {
			return
		}
		if modal, ok := widget.(*Modal); ok && modal.IsOpen && modal.Visible() && modal.Enabled() {
			found = modal
			return
		}
	}
	walk(ui.root)
	return found
}

func isDescendantOf(widget, ancestor Widget) bool {
	for current := widget; current != nil; current = current.Parent() {
		if sameWidgetIdentity(current, ancestor) {
			return true
		}
	}
	return false
}

func sameWidgetIdentity(a, b Widget) bool {
	if a == b {
		return true
	}
	ab := baseWidgetOf(a)
	bb := baseWidgetOf(b)
	return ab != nil && ab == bb
}

func tabIndexOf(widget Widget) int {
	if bw := baseWidgetOf(widget); bw != nil {
		return bw.TabIndex()
	}
	return -1
}

func resetFormSubtree(widget Widget) {
	for _, child := range widget.Children() {
		resetFormSubtree(child)
	}
	bw := baseWidgetOf(widget)
	if bw == nil {
		return
	}
	initial := bw.FormInitialValue()
	switch w := widget.(type) {
	case *TextInput:
		w.SetText(initial)
	case *TextArea:
		w.SetText(initial)
	case *Checkbox:
		w.Checked = initial == "true"
	case *Toggle:
		w.Checked = initial == "true"
	case *RadioButton:
		w.Selected = initial == "true"
		if w.Group != nil && w.Selected {
			w.Group.SetValue(w.Value)
		}
	case *Dropdown:
		w.SetValue(initial)
	case *Slider:
		if value, ok := bindingFloat(initial); ok {
			w.Value = value
		}
	}
	bw.SetValidationState(ValidationNone)
	bw.SetValidationMessage("")
}

func validateSingleWidget(widget Widget) bool {
	bw := baseWidgetOf(widget)
	if bw == nil {
		return true
	}
	rules := bw.ValidationRules()
	if !rules.HasConstraints() {
		bw.SetValidationState(ValidationNone)
		bw.SetValidationMessage("")
		return true
	}
	valid, message := checkWidgetValidity(widget, rules)
	if valid {
		bw.SetValidationState(ValidationValid)
		bw.SetValidationMessage("")
		return true
	}
	if rules.Message != "" {
		message = rules.Message
	}
	bw.SetValidationState(ValidationInvalid)
	bw.SetValidationMessage(message)
	return false
}

func checkWidgetValidity(widget Widget, rules ValidationRules) (bool, string) {
	if rules.Required && isWidgetEmpty(widget) {
		return false, "This field is required."
	}

	text, hasText := widgetTextValue(widget)
	if hasText {
		length := utf8.RuneCountInString(text)
		if rules.MinLengthSet && length < rules.MinLength {
			return false, "Value is too short."
		}
		if rules.MaxLengthSet && length > rules.MaxLength {
			return false, "Value is too long."
		}
		if rules.Type == "email" && strings.TrimSpace(text) != "" && !isEmailLike(text) {
			return false, "Enter a valid email address."
		}
		if rules.Pattern != "" && strings.TrimSpace(text) != "" {
			matched, err := regexp.MatchString(rules.Pattern, text)
			if err != nil || !matched {
				return false, "Value does not match the required pattern."
			}
		}
	}

	if rules.Type == "number" || rules.MinSet || rules.MaxSet {
		if value, ok := widgetNumericValue(widget); ok {
			if rules.MinSet && value < rules.Min {
				return false, "Value must be at least " + formatValidationNumber(rules.Min) + "."
			}
			if rules.MaxSet && value > rules.Max {
				return false, "Value must be at most " + formatValidationNumber(rules.Max) + "."
			}
		} else if rules.Type == "number" && hasText && strings.TrimSpace(text) != "" {
			return false, "Enter a valid number."
		}
	}

	return true, ""
}

func isWidgetEmpty(widget Widget) bool {
	switch w := widget.(type) {
	case *TextInput:
		return strings.TrimSpace(w.Text) == ""
	case *TextArea:
		return strings.TrimSpace(w.Text) == ""
	case *Checkbox:
		return !w.Checked
	case *Toggle:
		return !w.Checked
	case *RadioButton:
		return !w.Selected
	case *Dropdown:
		return w.SelectedIndex < 0 || w.GetSelectedValue() == ""
	default:
		return false
	}
}

func widgetTextValue(widget Widget) (string, bool) {
	switch w := widget.(type) {
	case *TextInput:
		return w.Text, true
	case *TextArea:
		return w.Text, true
	default:
		return "", false
	}
}

func widgetNumericValue(widget Widget) (float64, bool) {
	switch w := widget.(type) {
	case *Slider:
		return w.Value, true
	case *TextInput:
		value, err := strconv.ParseFloat(strings.TrimSpace(w.Text), 64)
		return value, err == nil
	case *TextArea:
		value, err := strconv.ParseFloat(strings.TrimSpace(w.Text), 64)
		return value, err == nil
	default:
		return 0, false
	}
}

func isEmailLike(value string) bool {
	trimmed := strings.TrimSpace(value)
	at := strings.Index(trimmed, "@")
	dot := strings.LastIndex(trimmed, ".")
	return at > 0 && dot > at+1 && dot < len(trimmed)-1
}

func formatValidationNumber(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func (ui *UI) setFocusedWidget(w Widget) {
	if ui.focusedWidget == w {
		return
	}

	if ui.focusedWidget != nil {
		if ti, ok := ui.focusedWidget.(*TextInput); ok {
			ti.Blur()
		}
		if ta, ok := ui.focusedWidget.(*TextArea); ok {
			ta.Blur()
		}
		if ui.focusedWidget == ui.hoveredWidget {
			ui.focusedWidget.SetState(StateHover)
		} else {
			ui.focusedWidget.SetState(StateNormal)
		}
	}

	ui.focusedWidget = w
	if w == nil {
		return
	}

	switch fw := w.(type) {
	case *TextInput:
		fw.Focus()
	case *TextArea:
		fw.Focus()
	default:
		w.SetState(StateFocused)
	}
}

func (ui *UI) setRoot(widget Widget) {
	ui.root = widget
	ui.refreshDynamicTree()
}

func (ui *UI) refreshDynamicTree() {
	if ui.root == nil {
		return
	}
	ui.widgetByID = make(map[string]Widget)
	ui.buildWidgetCache(ui.root)
	ui.reapplyStyles(ui.root)
	ui.inheritCSSProperties(ui.root, nil)
	ui.setFonts(ui.root)
	ui.Layout()
}

func (ui *UI) refreshDynamicLayout() {
	if ui.root == nil {
		return
	}
	ui.inheritCSSProperties(ui.root, nil)
	ui.setFonts(ui.root)
	ui.Layout()
}

func (ui *UI) handlePointerMove(x, y float64) Widget {
	if ui.root == nil {
		return nil
	}

	hoveredWidget := ui.findWidgetAt(ui.root, x, y)
	if hoveredWidget != ui.hoveredWidget {
		if ui.hoveredWidget != nil {
			if txt, ok := ui.hoveredWidget.(*Text); ok {
				txt.ClearHoveredCluster()
			}
			if ui.hoveredWidget == ui.focusedWidget {
				ui.hoveredWidget.SetState(StateFocused)
			} else {
				ui.hoveredWidget.SetState(StateNormal)
			}
		}
		if hoveredWidget != nil {
			if hoveredWidget != ui.focusedWidget {
				hoveredWidget.SetState(StateHover)
				if bw, ok := hoveredWidget.(*Button); ok {
					bw.HandleHover()
				}
			}
		}
		ui.hoveredWidget = hoveredWidget
	}
	if txt, ok := hoveredWidget.(*Text); ok {
		txt.HandlePointerMove(x, y)
	}
	return hoveredWidget
}

func (ui *UI) handlePointerDown(x, y float64, button ebiten.MouseButton) Widget {
	hoveredWidget := ui.handlePointerMove(x, y)
	if button != ebiten.MouseButtonLeft {
		return hoveredWidget
	}

	if hoveredWidget != nil {
		switch w := hoveredWidget.(type) {
		case *TextInput:
			ui.setFocusedWidget(hoveredWidget)
			w.HandlePointerDown(x, y)
		case *TextArea:
			ui.setFocusedWidget(hoveredWidget)
			w.HandlePointerDown(x, y)
		default:
			if ui.focusedWidget != nil && ui.focusedWidget != hoveredWidget {
				ui.Blur()
			}
		}
		hoveredWidget.SetState(StateActive)
		ui.activeWidget = hoveredWidget
		return hoveredWidget
	}

	ui.Blur()
	return nil
}

func (ui *UI) handlePointerUp(x, y float64, button ebiten.MouseButton, hoveredWidget Widget) {
	if button != ebiten.MouseButtonLeft || ui.activeWidget == nil {
		return
	}
	if hoveredWidget == nil {
		hoveredWidget = ui.handlePointerMove(x, y)
	}

	if ui.activeWidget == hoveredWidget {
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

	if ui.activeWidget == ui.focusedWidget {
		ui.activeWidget.SetState(StateFocused)
	} else {
		ui.activeWidget.SetState(StateNormal)
	}
	if hoveredWidget != nil {
		if hoveredWidget == ui.focusedWidget {
			hoveredWidget.SetState(StateFocused)
		} else {
			hoveredWidget.SetState(StateHover)
		}
	}
	ui.activeWidget = nil
}

func simulateTextInputKeyPress(ti *TextInput, key ebiten.Key, shift, control bool) {
	switch {
	case control && key == ebiten.KeyA:
		ti.SelectStart = 0
		ti.SelectEnd = utf8.RuneCountInString(ti.Text)
		ti.CursorPos = ti.SelectEnd
		ti.clampIndices()
	case key == ebiten.KeyBackspace:
		ti.handleBackspace()
	case key == ebiten.KeyDelete:
		ti.handleDelete()
	case key == ebiten.KeyLeft:
		ti.moveCursor(-1, shift)
	case key == ebiten.KeyRight:
		ti.moveCursor(1, shift)
	case key == ebiten.KeyHome:
		ti.CursorPos = 0
		ti.clampIndices()
	case key == ebiten.KeyEnd:
		ti.CursorPos = utf8.RuneCountInString(ti.Text)
		ti.clampIndices()
	case key == ebiten.KeyEnter || key == ebiten.KeyNumpadEnter:
		if ti.OnSubmit != nil {
			ti.OnSubmit(ti.Text)
		}
	}
}

func simulateTextAreaKeyPress(ta *TextArea, key ebiten.Key, _ bool, control bool) {
	switch {
	case control && key == ebiten.KeyA:
		ta.SelectStart = 0
		ta.SelectEnd = utf8.RuneCountInString(ta.Text)
		ta.CursorPos = ta.SelectEnd
		ta.updateCursorLineCol()
	case key == ebiten.KeyBackspace:
		ta.handleBackspace()
	case key == ebiten.KeyDelete:
		ta.handleDelete()
	case key == ebiten.KeyEnter || key == ebiten.KeyNumpadEnter:
		ta.insertChar('\n')
	case key == ebiten.KeyLeft:
		ta.moveCursorHorizontal(-1)
	case key == ebiten.KeyRight:
		ta.moveCursorHorizontal(1)
	case key == ebiten.KeyUp:
		ta.moveCursorVertical(-1)
	case key == ebiten.KeyDown:
		ta.moveCursorVertical(1)
	case key == ebiten.KeyHome:
		ta.CursorCol = 0
		ta.updateCursorPosFromLineCol()
	case key == ebiten.KeyEnd:
		if ta.CursorLine >= 0 && ta.CursorLine < len(ta.lines) {
			ta.CursorCol = utf8.RuneCountInString(ta.lines[ta.CursorLine])
			ta.updateCursorPosFromLineCol()
		}
	}
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
	style := widget.Style()
	if bw := baseWidgetOf(widget); bw != nil {
		style = bw.getActiveStyle()
		if style.Overflow == "hidden" || style.Overflow == "scroll" || style.Overflow == "auto" {
			if !bw.ContentRect().Contains(x, y) {
				return nil
			}
		}
	}

	// Check children first in reverse visual z-order.
	childX, childY := x, y
	if bw := baseWidgetOf(widget); bw != nil && (style.Overflow == "scroll" || style.Overflow == "auto") {
		scrollX, scrollY := bw.ScrollOffset()
		childX += scrollX
		childY += scrollY
	}
	for _, child := range sortedChildrenByZ(widget.Children(), true) {
		found := ui.findWidgetAt(child, childX, childY)
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

func (ui *UI) scrollHoveredWidget(widget Widget, dx, dy float64) {
	for current := widget; current != nil; current = current.Parent() {
		switch w := current.(type) {
		case *Scrollable:
			w.ScrollBy(dx, dy)
			return
		}
		bw := baseWidgetOf(current)
		if bw == nil {
			continue
		}
		style := bw.getActiveStyle()
		if style.Overflow == "scroll" || style.Overflow == "auto" {
			bw.ScrollBy(dx, dy)
			return
		}
	}
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
	fontFace := ui.resolveFontFace(widget.Style())

	// Apply to text-bearing widgets
	if fontFace != nil {
		switch w := widget.(type) {
		case *Button:
			w.FontFace = fontFace
		case *Text:
			w.FontFace = fontFace
		case *TextInput:
			w.FontFace = fontFace
		case *TextArea:
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

func (ui *UI) resolveFontFace(style *Style) text.Face {
	if style == nil {
		return ui.defaultFontFace(14, false)
	}
	fontSize := style.FontSize
	if fontSize <= 0 {
		fontSize = 14
	}
	isBold := style.FontWeight == "bold" || style.FontWeight == "700" ||
		style.FontWeight == "800" || style.FontWeight == "900"

	for _, family := range parseFontFamilyList(style.FontFamily) {
		name := normalizeFontFamilyName(family)
		if name == "" {
			continue
		}
		if face := ui.fontFaces[name]; face != nil {
			return face
		}
		if isBold {
			if cache := ui.boldFontSources[name]; cache != nil {
				return cache.GetFace(fontSize)
			}
		}
		if cache := ui.fontSources[name]; cache != nil {
			return cache.GetFace(fontSize)
		}
	}

	return ui.defaultFontFace(fontSize, isBold)
}

func (ui *UI) defaultFontFace(fontSize float64, isBold bool) text.Face {
	if ui.DefaultFontFace != nil {
		return ui.DefaultFontFace
	}
	if isBold && ui.boldFontCache != nil {
		return ui.boldFontCache.GetFace(fontSize)
	}
	if ui.fontCache != nil {
		return ui.fontCache.GetFace(fontSize)
	}
	return nil
}

func parseFontFamilyList(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	var families []string
	var current strings.Builder
	var quote rune
	for _, r := range value {
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
			} else {
				current.WriteRune(r)
			}
		case r == '\'' || r == '"':
			quote = r
		case r == ',':
			if family := strings.TrimSpace(current.String()); family != "" {
				families = append(families, family)
			}
			current.Reset()
		default:
			current.WriteRune(r)
		}
	}
	if family := strings.TrimSpace(current.String()); family != "" {
		families = append(families, family)
	}
	return families
}

func normalizeFontFamilyName(family string) string {
	family = strings.TrimSpace(family)
	family = strings.Trim(family, `"'`)
	return strings.ToLower(family)
}

// reapplyStyles reapplies styles from the style engine
func (ui *UI) reapplyStyles(widget Widget) {
	ui.reapplyStylesWithAncestors(widget, nil)
}

func (ui *UI) reapplyStylesWithAncestors(widget Widget, ancestors []Widget) {
	if widget == nil {
		return
	}

	// 1. Apply by type (e.g., "panel", "svg", "button")
	ui.styleEngine.ApplyStyle(widget, widget.Type())

	// 2. Apply by individual classes
	for _, class := range widget.Classes() {
		ui.styleEngine.ApplyStyle(widget, "."+class)
	}

	// 3. Apply by compound classes (e.g., ".nav-item.active")
	// This is a simple implementation: try all pairs of classes.
	// For better support, we'd need a real CSS selector engine.
	classes := widget.Classes()
	if len(classes) >= 2 {
		// Try combinations (only pairs for now as it's most common)
		for i := 0; i < len(classes); i++ {
			for j := i + 1; j < len(classes); j++ {
				ui.styleEngine.ApplyStyle(widget, "."+classes[i]+"."+classes[j])
				ui.styleEngine.ApplyStyle(widget, "."+classes[j]+"."+classes[i])
			}
		}
	}

	// 4. Apply source-ordered CSS rules. This pass preserves later wins for
	// same-specificity rules while the surrounding direct passes keep legacy
	// type/class/id behavior stable.
	ui.applyOrderedRuleStyles(widget, ancestors, false)

	// 5. Apply by ID (e.g., "#root", "#header")
	if widget.ID() != "" {
		ui.styleEngine.ApplyStyle(widget, "#"+widget.ID())
	}

	// 6. Apply !important CSS declarations after normal direct styles.
	ui.applyOrderedRuleStyles(widget, ancestors, true)

	// Recursively apply to children
	childAncestors := append(append([]Widget(nil), ancestors...), widget)
	for _, child := range widget.Children() {
		ui.reapplyStylesWithAncestors(child, childAncestors)
	}
}

func (ui *UI) applyOrderedRuleStyles(widget Widget, ancestors []Widget, important bool) {
	matches := ui.styleEngine.matchingRules(widget, ancestors, important)
	for _, rule := range matches {
		existing := widget.Style()
		widget.SetStyle(mergeStylesFully(existing, rule.Style))
	}
}

func (se *StyleEngine) matchingRules(widget Widget, ancestors []Widget, important bool) []styleRuleRecord {
	matches := make([]styleRuleRecord, 0)
	for _, rule := range se.rules {
		if rule.Important != important {
			continue
		}
		if selectorMatchesWidget(widget, ancestors, rule.Selector) {
			matches = append(matches, rule)
		}
	}
	sort.SliceStable(matches, func(i, j int) bool {
		specI := matches[i].Specificity
		specJ := matches[j].Specificity
		if specI == specJ {
			return matches[i].Order < matches[j].Order
		}
		return specI < specJ
	})
	return matches
}

func selectorMatchesWidget(widget Widget, ancestors []Widget, selector string) bool {
	parts, _, ok := splitComplexSelector(selector)
	if !ok || len(parts) == 1 {
		return simpleSelectorPartMatches(widget, strings.TrimSpace(selector))
	}
	return complexSelectorMatchesWidget(widget, ancestors, selector)
}

func complexSelectorMatchesWidget(widget Widget, ancestors []Widget, selector string) bool {
	parts, combinators, ok := splitComplexSelector(selector)
	if !ok {
		return false
	}
	if len(parts) < 2 || len(combinators) != len(parts)-1 {
		return false
	}
	if !simpleSelectorPartMatches(widget, parts[len(parts)-1]) {
		return false
	}

	ancestorIndex := len(ancestors) - 1
	for partIndex := len(parts) - 2; partIndex >= 0; partIndex-- {
		combinator := combinators[partIndex]
		switch combinator {
		case ">":
			if ancestorIndex < 0 || !simpleSelectorPartMatches(ancestors[ancestorIndex], parts[partIndex]) {
				return false
			}
			ancestorIndex--
		case "+":
			if !previousSiblingMatches(widget, ancestors, parts[partIndex], false) {
				return false
			}
		case "~":
			if !previousSiblingMatches(widget, ancestors, parts[partIndex], true) {
				return false
			}
		default:
			found := -1
			for i := ancestorIndex; i >= 0; i-- {
				if simpleSelectorPartMatches(ancestors[i], parts[partIndex]) {
					found = i
					break
				}
			}
			if found < 0 {
				return false
			}
			ancestorIndex = found - 1
		}
	}
	return true
}

func splitComplexSelector(selector string) ([]string, []string, bool) {
	selector = strings.NewReplacer(">", " > ", "+", " + ", "~", " ~ ").Replace(selector)
	fields := strings.Fields(selector)
	if len(fields) == 0 {
		return nil, nil, false
	}

	var parts []string
	var combinators []string
	nextCombinator := " "
	for _, field := range fields {
		if field == ">" || field == "+" || field == "~" {
			if len(parts) == 0 {
				return nil, nil, false
			}
			nextCombinator = field
			continue
		}
		if len(parts) > 0 {
			combinators = append(combinators, nextCombinator)
		}
		parts = append(parts, field)
		nextCombinator = " "
	}
	return parts, combinators, true
}

func complexSelectorSpecificity(selector string) int {
	parts, _, ok := splitComplexSelector(selector)
	if !ok {
		return 0
	}
	specificity := 0
	for _, part := range parts {
		parsed := ParseSelector(part)
		if parsed != nil {
			specificity += parsed.CalculateSpecificity()
		}
	}
	return specificity
}

func previousSiblingMatches(widget Widget, ancestors []Widget, selector string, anyPrevious bool) bool {
	if len(ancestors) == 0 {
		return false
	}
	parent := ancestors[len(ancestors)-1]
	siblings := parent.Children()
	widgetIndex := -1
	for i, sibling := range siblings {
		if sibling == widget {
			widgetIndex = i
			break
		}
	}
	if widgetIndex <= 0 {
		return false
	}
	if !anyPrevious {
		return simpleSelectorPartMatches(siblings[widgetIndex-1], selector)
	}
	for i := widgetIndex - 1; i >= 0; i-- {
		if simpleSelectorPartMatches(siblings[i], selector) {
			return true
		}
	}
	return false
}

func simpleSelectorPartMatches(widget Widget, selector string) bool {
	parsed := ParseSelector(selector)
	return parsed != nil && parsed.Next == nil && parsed.PseudoClass == "" && parsed.matchesSingle(widget)
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
