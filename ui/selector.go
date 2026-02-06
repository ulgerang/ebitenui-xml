package ui

import (
	"regexp"
	"strings"
)

// ============================================================================
// Advanced CSS Selector System
// ============================================================================

// SelectorType represents the type of CSS selector
type SelectorType int

const (
	SelectorTypeUniversal   SelectorType = iota // *
	SelectorTypeTag                             // button
	SelectorTypeClass                           // .class
	SelectorTypeID                              // #id
	SelectorTypeDescendant                      // parent child
	SelectorTypeChild                           // parent > child
	SelectorTypeAttribute                       // [attr=value]
	SelectorTypePseudoClass                     // :hover
	SelectorTypeCompound                        // button.class#id
)

// Selector represents a parsed CSS selector
type Selector struct {
	Type        SelectorType
	Value       string
	Classes     []string
	ID          string
	Tag         string
	Attribute   string
	AttrValue   string
	PseudoClass string
	Combinator  string    // " " or ">" or "+" or "~"
	Next        *Selector // For compound/chained selectors
}

// ParseSelector parses a CSS selector string
func ParseSelector(s string) *Selector {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	// Handle descendant combinator (space)
	if strings.Contains(s, " > ") {
		parts := strings.SplitN(s, " > ", 2)
		parent := ParseSelector(parts[0])
		child := ParseSelector(parts[1])
		if parent != nil && child != nil {
			parent.Combinator = ">"
			parent.Next = child
		}
		return parent
	}

	// Handle descendant selector (space without >)
	if strings.Contains(s, " ") {
		parts := strings.SplitN(s, " ", 2)
		parent := ParseSelector(parts[0])
		child := ParseSelector(parts[1])
		if parent != nil && child != nil {
			parent.Combinator = " "
			parent.Next = child
		}
		return parent
	}

	// Parse single selector
	return parseSingleSelector(s)
}

func parseSingleSelector(s string) *Selector {
	sel := &Selector{}

	// Handle pseudo-class
	if idx := strings.Index(s, ":"); idx != -1 {
		sel.PseudoClass = s[idx+1:]
		s = s[:idx]
	}

	// Handle attribute selector
	attrRe := regexp.MustCompile(`\[([^\]=]+)(?:=["']?([^"'\]]*)["']?)?\]`)
	if matches := attrRe.FindStringSubmatch(s); len(matches) > 0 {
		sel.Attribute = matches[1]
		if len(matches) > 2 {
			sel.AttrValue = matches[2]
		}
		s = attrRe.ReplaceAllString(s, "")
	}

	// Parse compound selector (tag.class1.class2#id)
	// ID
	if idx := strings.Index(s, "#"); idx != -1 {
		parts := strings.SplitN(s, "#", 2)
		s = parts[0]
		idPart := parts[1]
		// ID might be followed by classes
		if dotIdx := strings.Index(idPart, "."); dotIdx != -1 {
			sel.ID = idPart[:dotIdx]
			// Remaining classes
			for _, class := range strings.Split(idPart[dotIdx+1:], ".") {
				if class != "" {
					sel.Classes = append(sel.Classes, class)
				}
			}
		} else {
			sel.ID = idPart
		}
		sel.Type = SelectorTypeID
	}

	// Classes
	if strings.Contains(s, ".") {
		parts := strings.Split(s, ".")
		if parts[0] != "" {
			sel.Tag = parts[0]
		}
		for _, class := range parts[1:] {
			if class != "" {
				sel.Classes = append(sel.Classes, class)
			}
		}
		if len(sel.Classes) > 0 {
			sel.Type = SelectorTypeClass
		}
	} else if s != "" {
		sel.Tag = s
		sel.Type = SelectorTypeTag
	}

	// Universal
	if s == "*" {
		sel.Type = SelectorTypeUniversal
	}

	// Determine compound type
	if sel.ID != "" && (len(sel.Classes) > 0 || sel.Tag != "") {
		sel.Type = SelectorTypeCompound
	} else if sel.Tag != "" && len(sel.Classes) > 0 {
		sel.Type = SelectorTypeCompound
	}

	return sel
}

// Matches checks if a widget matches this selector
func (s *Selector) Matches(widget Widget, parent Widget) bool {
	if s == nil {
		return false
	}

	// Match this selector against widget
	if !s.matchesSingle(widget) {
		return false
	}

	// Handle combinators
	if s.Next != nil {
		switch s.Combinator {
		case " ": // Descendant
			return s.matchesDescendant(widget, s.Next)
		case ">": // Direct child
			return s.matchesChild(widget, s.Next)
		}
	}

	return true
}

func (s *Selector) matchesSingle(widget Widget) bool {
	// Match tag
	if s.Tag != "" && s.Tag != widget.Type() {
		return false
	}

	// Match ID
	if s.ID != "" && s.ID != widget.ID() {
		return false
	}

	// Match classes
	for _, class := range s.Classes {
		if !widget.HasClass(class) {
			return false
		}
	}

	// Match attribute
	if s.Attribute != "" {
		// For now, only support class attribute
		if s.Attribute == "class" {
			if s.AttrValue != "" && !widget.HasClass(s.AttrValue) {
				return false
			}
		}
	}

	return true
}

func (s *Selector) matchesDescendant(widget Widget, childSelector *Selector) bool {
	// Check all descendants
	return checkDescendants(widget, childSelector)
}

func checkDescendants(widget Widget, selector *Selector) bool {
	for _, child := range widget.Children() {
		if selector.Matches(child, widget) {
			return true
		}
		if checkDescendants(child, selector) {
			return true
		}
	}
	return false
}

func (s *Selector) matchesChild(widget Widget, childSelector *Selector) bool {
	// Check direct children only
	for _, child := range widget.Children() {
		if childSelector.Matches(child, widget) {
			return true
		}
	}
	return false
}

// ============================================================================
// Enhanced StyleEngine with Advanced Selectors
// ============================================================================

// StyleRule represents a CSS rule with selector
type StyleRule struct {
	Selector    *Selector
	RawSelector string
	Style       *Style
	Specificity int
}

// CalculateSpecificity calculates CSS specificity
func (s *Selector) CalculateSpecificity() int {
	spec := 0

	if s.ID != "" {
		spec += 100
	}
	spec += len(s.Classes) * 10
	if s.Tag != "" {
		spec += 1
	}
	if s.PseudoClass != "" {
		spec += 10
	}

	if s.Next != nil {
		spec += s.Next.CalculateSpecificity()
	}

	return spec
}

// AdvancedStyleEngine extends StyleEngine with complex selectors
type AdvancedStyleEngine struct {
	*StyleEngine
	rules     []StyleRule
	variables *CSSVariables
}

// NewAdvancedStyleEngine creates a new advanced style engine
func NewAdvancedStyleEngine() *AdvancedStyleEngine {
	return &AdvancedStyleEngine{
		StyleEngine: NewStyleEngine(),
		rules:       make([]StyleRule, 0),
		variables:   NewCSSVariables(),
	}
}

// SetVariable sets a CSS variable
func (ase *AdvancedStyleEngine) SetVariable(name, value string) {
	ase.variables.Set(name, value)
}

// GetVariable gets a CSS variable
func (ase *AdvancedStyleEngine) GetVariable(name string) string {
	return ase.variables.Get(name)
}

// AddRule adds a style rule with selector
func (ase *AdvancedStyleEngine) AddRule(selectorStr string, style *Style) {
	selector := ParseSelector(selectorStr)
	if selector != nil {
		rule := StyleRule{
			Selector:    selector,
			RawSelector: selectorStr,
			Style:       style,
			Specificity: selector.CalculateSpecificity(),
		}
		ase.rules = append(ase.rules, rule)

		// Sort by specificity (ascending, last wins for equal)
		// CSS cascade: later rules with same specificity win
	}

	// Also add to base engine for simple lookups
	ase.StyleEngine.AddStyle(selectorStr, style)
}

// GetMatchingStyles returns all styles that match a widget, sorted by specificity
func (ase *AdvancedStyleEngine) GetMatchingStyles(widget Widget, ancestors []Widget) []*Style {
	matched := make([]*Style, 0)

	for _, rule := range ase.rules {
		if ase.ruleMatches(rule, widget, ancestors) {
			matched = append(matched, rule.Style)
		}
	}

	return matched
}

func (ase *AdvancedStyleEngine) ruleMatches(rule StyleRule, widget Widget, ancestors []Widget) bool {
	selector := rule.Selector

	// Simple selector
	if selector.Next == nil {
		return selector.matchesSingle(widget)
	}

	// Complex selector with combinator
	current := widget
	currentSel := selector

	// Walk backwards through ancestors for descendant selectors
	if currentSel.Combinator == " " {
		// Find matching ancestor
		for _, ancestor := range ancestors {
			if currentSel.matchesSingle(ancestor) {
				// Continue matching the rest
				if currentSel.Next.matchesSingle(widget) {
					return true
				}
			}
		}
		return false
	}

	// Direct parent for child selector
	if currentSel.Combinator == ">" && len(ancestors) > 0 {
		parent := ancestors[len(ancestors)-1]
		if currentSel.matchesSingle(parent) {
			return currentSel.Next.matchesSingle(current)
		}
	}

	return false
}

// ApplyAllStyles applies all matching styles to a widget
func (ase *AdvancedStyleEngine) ApplyAllStyles(widget Widget, ancestors []Widget) {
	styles := ase.GetMatchingStyles(widget, ancestors)

	if len(styles) == 0 {
		// Fall back to basic matching
		ase.StyleEngine.ApplyStyle(widget, widget.Type())
		for _, class := range widget.Classes() {
			ase.StyleEngine.ApplyStyle(widget, "."+class)
		}
		if widget.ID() != "" {
			ase.StyleEngine.ApplyStyle(widget, "#"+widget.ID())
		}
		return
	}

	// Merge all matching styles in order
	merged := &Style{Opacity: 1}
	for _, s := range styles {
		merged.Merge(s)
	}

	widget.SetStyle(merged)
}

// ResolveVariables resolves CSS variables in a style
func (ase *AdvancedStyleEngine) ResolveVariables(style *Style) {
	if style == nil {
		return
	}

	if style.Background != "" {
		style.Background = ase.variables.Resolve(style.Background)
	}
	if style.Color != "" {
		style.Color = ase.variables.Resolve(style.Color)
	}
	if style.Border != "" {
		style.Border = ase.variables.Resolve(style.Border)
	}
	if style.BoxShadow != "" {
		style.BoxShadow = ase.variables.Resolve(style.BoxShadow)
	}
}

// ============================================================================
// :nth-child Selector Support
// ============================================================================

// ParseNthChild parses :nth-child(n) expressions
func ParseNthChild(expr string) (a, b int) {
	expr = strings.TrimSpace(expr)

	// Handle special keywords
	if expr == "odd" {
		return 2, 1
	}
	if expr == "even" {
		return 2, 0
	}

	// Handle simple number
	if n, err := parseInt(expr); err == nil {
		return 0, n
	}

	// Handle An+B format
	re := regexp.MustCompile(`(-?\d*)n\s*([+-]\s*\d+)?`)
	matches := re.FindStringSubmatch(expr)
	if len(matches) > 0 {
		aStr := matches[1]
		if aStr == "" || aStr == "-" {
			if aStr == "-" {
				a = -1
			} else {
				a = 1
			}
		} else {
			a, _ = parseInt(aStr)
		}

		if len(matches) > 2 && matches[2] != "" {
			bStr := strings.ReplaceAll(matches[2], " ", "")
			b, _ = parseInt(bStr)
		}
	}

	return a, b
}

// MatchesNthChild checks if index matches :nth-child(An+B)
func MatchesNthChild(index, a, b int) bool {
	if a == 0 {
		return index == b
	}
	return (index-b)%a == 0 && (index-b)/a >= 0
}

func parseInt(s string) (int, error) {
	var n int
	_, err := strings.NewReader(s).Read([]byte{})
	if err != nil {
		return 0, err
	}
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		} else if c == '-' && n == 0 {
			continue
		} else {
			break
		}
	}
	if strings.HasPrefix(s, "-") {
		n = -n
	}
	return n, nil
}
