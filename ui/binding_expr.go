package ui

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type bindingExprTokenType int

const (
	bindingTokenEOF bindingExprTokenType = iota
	bindingTokenIdent
	bindingTokenNumber
	bindingTokenString
	bindingTokenOperator
	bindingTokenLParen
	bindingTokenRParen
	bindingTokenComma
)

type bindingExprToken struct {
	typ bindingExprTokenType
	val string
}

type bindingExprParser struct {
	tokens []bindingExprToken
	pos    int
	deps   map[string]bool
	ctx    *BindingContext
}

func renderBindingExpressionTemplate(template string, bindings *BindingContext) string {
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
		if value, ok := evalBindingExpression(expr, bindings); ok && value != nil {
			rendered.WriteString(bindingString(value))
		}
		remaining = remaining[end+2:]
	}
	return rendered.String()
}

func extractBindingExpressionDeps(template string) []string {
	seen := make(map[string]bool)
	var deps []string
	for _, expr := range extractTemplateExpressions(template) {
		for _, dep := range bindingExpressionDependencies(expr) {
			if !seen[dep] {
				seen[dep] = true
				deps = append(deps, dep)
			}
		}
	}
	return deps
}

func extractTemplateExpressions(template string) []string {
	var expressions []string
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
		expr := strings.TrimSpace(remaining[:end])
		if expr != "" {
			expressions = append(expressions, expr)
		}
		remaining = remaining[end+2:]
	}
	return expressions
}

func bindingExpressionDependencies(expr string) []string {
	parser := newBindingExprParser(expr, nil)
	if parser == nil {
		return nil
	}
	_, _ = parser.parseExpression()
	deps := make([]string, 0, len(parser.deps))
	for dep := range parser.deps {
		deps = append(deps, dep)
	}
	return deps
}

func evalBindingExpression(expr string, bindings *BindingContext) (interface{}, bool) {
	parser := newBindingExprParser(expr, bindings)
	if parser == nil {
		return nil, false
	}
	value, ok := parser.parseExpression()
	if !ok {
		return nil, false
	}
	return value, true
}

func newBindingExprParser(expr string, bindings *BindingContext) *bindingExprParser {
	tokens, ok := tokenizeBindingExpression(expr)
	if !ok {
		return nil
	}
	return &bindingExprParser{
		tokens: tokens,
		deps:   make(map[string]bool),
		ctx:    bindings,
	}
}

func tokenizeBindingExpression(expr string) ([]bindingExprToken, bool) {
	var tokens []bindingExprToken
	for i := 0; i < len(expr); {
		r := rune(expr[i])
		if unicode.IsSpace(r) {
			i++
			continue
		}
		switch {
		case isBindingIdentStart(r):
			start := i
			i++
			for i < len(expr) && isBindingIdentPart(rune(expr[i])) {
				i++
			}
			tokens = append(tokens, bindingExprToken{typ: bindingTokenIdent, val: expr[start:i]})
		case unicode.IsDigit(r) || r == '.':
			start := i
			i++
			for i < len(expr) && (unicode.IsDigit(rune(expr[i])) || expr[i] == '.') {
				i++
			}
			tokens = append(tokens, bindingExprToken{typ: bindingTokenNumber, val: expr[start:i]})
		case r == '"' || r == '\'':
			quote := byte(r)
			i++
			var b strings.Builder
			for i < len(expr) && expr[i] != quote {
				if expr[i] == '\\' && i+1 < len(expr) {
					i++
				}
				b.WriteByte(expr[i])
				i++
			}
			if i >= len(expr) {
				return nil, false
			}
			i++
			tokens = append(tokens, bindingExprToken{typ: bindingTokenString, val: b.String()})
		case r == '(':
			tokens = append(tokens, bindingExprToken{typ: bindingTokenLParen, val: "("})
			i++
		case r == ')':
			tokens = append(tokens, bindingExprToken{typ: bindingTokenRParen, val: ")"})
			i++
		case r == ',':
			tokens = append(tokens, bindingExprToken{typ: bindingTokenComma, val: ","})
			i++
		default:
			if i+1 < len(expr) {
				two := expr[i : i+2]
				switch two {
				case "&&", "||", "==", "!=", "<=", ">=":
					tokens = append(tokens, bindingExprToken{typ: bindingTokenOperator, val: two})
					i += 2
					continue
				}
			}
			if strings.ContainsRune("!+-*/<>", r) {
				tokens = append(tokens, bindingExprToken{typ: bindingTokenOperator, val: string(r)})
				i++
				continue
			}
			return nil, false
		}
	}
	tokens = append(tokens, bindingExprToken{typ: bindingTokenEOF})
	return tokens, true
}

func isBindingIdentStart(r rune) bool {
	return unicode.IsLetter(r) || r == '_'
}

func isBindingIdentPart(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '.' || r == '-'
}

func (p *bindingExprParser) parseExpression() (interface{}, bool) {
	return p.parseOr()
}

func (p *bindingExprParser) parseOr() (interface{}, bool) {
	left, ok := p.parseAnd()
	if !ok {
		return nil, false
	}
	for p.matchOperator("||") {
		right, ok := p.parseAnd()
		if !ok {
			return nil, false
		}
		if bindingTruthy(left) {
			continue
		}
		left = right
	}
	return left, true
}

func (p *bindingExprParser) parseAnd() (interface{}, bool) {
	left, ok := p.parseComparison()
	if !ok {
		return nil, false
	}
	for p.matchOperator("&&") {
		right, ok := p.parseComparison()
		if !ok {
			return nil, false
		}
		if !bindingTruthy(left) {
			left = false
			continue
		}
		left = right
	}
	return left, true
}

func (p *bindingExprParser) parseComparison() (interface{}, bool) {
	left, ok := p.parseAdditive()
	if !ok {
		return nil, false
	}
	for {
		op := p.peek().val
		if p.peek().typ != bindingTokenOperator || !isComparisonOperator(op) {
			return left, true
		}
		p.pos++
		right, ok := p.parseAdditive()
		if !ok {
			return nil, false
		}
		left = compareBindingValues(left, right, op)
	}
}

func (p *bindingExprParser) parseAdditive() (interface{}, bool) {
	left, ok := p.parseMultiplicative()
	if !ok {
		return nil, false
	}
	for {
		op := p.peek().val
		if op != "+" && op != "-" {
			return left, true
		}
		p.pos++
		right, ok := p.parseMultiplicative()
		if !ok {
			return nil, false
		}
		if op == "+" {
			if ln, lok := bindingFloat(left); lok {
				if rn, rok := bindingFloat(right); rok {
					left = normalizeBindingNumber(ln + rn)
					continue
				}
			}
			left = bindingString(left) + bindingString(right)
		} else {
			left = numericBindingOperation(left, right, op)
		}
	}
}

func (p *bindingExprParser) parseMultiplicative() (interface{}, bool) {
	left, ok := p.parseUnary()
	if !ok {
		return nil, false
	}
	for {
		op := p.peek().val
		if op != "*" && op != "/" {
			return left, true
		}
		p.pos++
		right, ok := p.parseUnary()
		if !ok {
			return nil, false
		}
		left = numericBindingOperation(left, right, op)
	}
}

func (p *bindingExprParser) parseUnary() (interface{}, bool) {
	if p.matchOperator("!") {
		value, ok := p.parseUnary()
		return !bindingTruthy(value), ok
	}
	if p.matchOperator("-") {
		value, ok := p.parseUnary()
		if !ok {
			return nil, false
		}
		f, _ := bindingFloat(value)
		return normalizeBindingNumber(-f), true
	}
	return p.parsePrimary()
}

func (p *bindingExprParser) parsePrimary() (interface{}, bool) {
	token := p.peek()
	switch token.typ {
	case bindingTokenNumber:
		p.pos++
		f, err := strconv.ParseFloat(token.val, 64)
		return normalizeBindingNumber(f), err == nil
	case bindingTokenString:
		p.pos++
		return token.val, true
	case bindingTokenIdent:
		p.pos++
		if p.match(bindingTokenLParen) {
			return p.parseFunctionCall(token.val)
		}
		return p.resolveIdentifier(token.val), true
	case bindingTokenLParen:
		p.pos++
		value, ok := p.parseExpression()
		if !ok || !p.match(bindingTokenRParen) {
			return nil, false
		}
		return value, true
	default:
		return nil, false
	}
}

func (p *bindingExprParser) parseFunctionCall(name string) (interface{}, bool) {
	var args []interface{}
	if !p.match(bindingTokenRParen) {
		for {
			value, ok := p.parseExpression()
			if !ok {
				return nil, false
			}
			args = append(args, value)
			if p.match(bindingTokenRParen) {
				break
			}
			if !p.match(bindingTokenComma) {
				return nil, false
			}
		}
	}
	switch strings.ToLower(name) {
	case "upper":
		if len(args) != 1 {
			return nil, false
		}
		return strings.ToUpper(bindingString(args[0])), true
	case "lower":
		if len(args) != 1 {
			return nil, false
		}
		return strings.ToLower(bindingString(args[0])), true
	case "default":
		if len(args) != 2 {
			return nil, false
		}
		if bindingTruthy(args[0]) {
			return args[0], true
		}
		return args[1], true
	case "number":
		if len(args) == 0 || len(args) > 2 {
			return nil, false
		}
		f, ok := bindingFloat(args[0])
		if !ok {
			return 0, true
		}
		if len(args) == 2 {
			digits, ok := bindingFloat(args[1])
			if ok {
				return strconv.FormatFloat(f, 'f', int(digits), 64), true
			}
		}
		return normalizeBindingNumber(f), true
	case "len":
		if len(args) != 1 {
			return nil, false
		}
		return bindingLength(args[0]), true
	case "round":
		if len(args) == 0 || len(args) > 2 {
			return nil, false
		}
		value, ok := bindingFloat(args[0])
		if !ok {
			return 0, true
		}
		digits := 0
		if len(args) == 2 {
			if parsed, ok := bindingFloat(args[1]); ok {
				digits = int(parsed)
			}
		}
		return normalizeBindingNumber(roundBindingNumber(value, digits)), true
	case "floor":
		if len(args) != 1 {
			return nil, false
		}
		value, ok := bindingFloat(args[0])
		if !ok {
			return 0, true
		}
		return normalizeBindingNumber(math.Floor(value)), true
	case "ceil":
		if len(args) != 1 {
			return nil, false
		}
		value, ok := bindingFloat(args[0])
		if !ok {
			return 0, true
		}
		return normalizeBindingNumber(math.Ceil(value)), true
	case "contains":
		if len(args) != 2 {
			return nil, false
		}
		return bindingContains(args[0], args[1]), true
	case "join":
		if len(args) == 0 || len(args) > 2 {
			return nil, false
		}
		sep := ", "
		if len(args) == 2 {
			sep = bindingString(args[1])
		}
		return bindingJoin(args[0], sep), true
	case "format":
		if len(args) == 0 {
			return nil, false
		}
		template := bindingString(args[0])
		return fmt.Sprintf(template, args[1:]...), true
	default:
		return nil, false
	}
}

func (p *bindingExprParser) resolveIdentifier(name string) interface{} {
	switch strings.ToLower(name) {
	case "true":
		return true
	case "false":
		return false
	case "nil", "null":
		return nil
	}
	p.deps[name] = true
	if p.ctx == nil {
		return nil
	}
	return p.ctx.Get(name)
}

func (p *bindingExprParser) peek() bindingExprToken {
	if p.pos >= len(p.tokens) {
		return bindingExprToken{typ: bindingTokenEOF}
	}
	return p.tokens[p.pos]
}

func (p *bindingExprParser) match(typ bindingExprTokenType) bool {
	if p.peek().typ != typ {
		return false
	}
	p.pos++
	return true
}

func (p *bindingExprParser) matchOperator(op string) bool {
	if p.peek().typ != bindingTokenOperator || p.peek().val != op {
		return false
	}
	p.pos++
	return true
}

func isComparisonOperator(op string) bool {
	switch op {
	case "==", "!=", "<", "<=", ">", ">=":
		return true
	default:
		return false
	}
}

func compareBindingValues(left, right interface{}, op string) bool {
	if lf, lok := bindingFloat(left); lok {
		if rf, rok := bindingFloat(right); rok {
			switch op {
			case "==":
				return lf == rf
			case "!=":
				return lf != rf
			case "<":
				return lf < rf
			case "<=":
				return lf <= rf
			case ">":
				return lf > rf
			case ">=":
				return lf >= rf
			}
		}
	}
	ls := bindingString(left)
	rs := bindingString(right)
	switch op {
	case "==":
		return ls == rs
	case "!=":
		return ls != rs
	case "<":
		return ls < rs
	case "<=":
		return ls <= rs
	case ">":
		return ls > rs
	case ">=":
		return ls >= rs
	default:
		return false
	}
}

func numericBindingOperation(left, right interface{}, op string) interface{} {
	lf, lok := bindingFloat(left)
	rf, rok := bindingFloat(right)
	if !lok || !rok {
		return 0
	}
	switch op {
	case "-":
		return normalizeBindingNumber(lf - rf)
	case "*":
		return normalizeBindingNumber(lf * rf)
	case "/":
		if rf == 0 {
			return 0
		}
		return normalizeBindingNumber(lf / rf)
	default:
		return 0
	}
}

func bindingFloat(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	case string:
		f, err := strconv.ParseFloat(v, 64)
		return f, err == nil
	}
	return 0, false
}

func bindingLength(value interface{}) int {
	if value == nil {
		return 0
	}
	switch v := value.(type) {
	case string:
		return utf8.RuneCountInString(v)
	}
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
		return rv.Len()
	default:
		text := bindingString(value)
		if text == "" {
			return 0
		}
		return utf8.RuneCountInString(text)
	}
}

func roundBindingNumber(value float64, digits int) float64 {
	if digits <= 0 {
		return math.Round(value)
	}
	scale := math.Pow10(digits)
	return math.Round(value*scale) / scale
}

func bindingContains(value, needle interface{}) bool {
	if value == nil {
		return false
	}
	if text, ok := value.(string); ok {
		return strings.Contains(text, bindingString(needle))
	}
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < rv.Len(); i++ {
			if bindingValuesEqual(rv.Index(i).Interface(), needle) {
				return true
			}
		}
	case reflect.Map:
		needleText := bindingString(needle)
		for _, key := range rv.MapKeys() {
			if bindingString(key.Interface()) == needleText || bindingValuesEqual(rv.MapIndex(key).Interface(), needle) {
				return true
			}
		}
	}
	return false
}

func bindingJoin(value interface{}, sep string) string {
	if value == nil {
		return ""
	}
	if text, ok := value.(string); ok {
		return text
	}
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Array && rv.Kind() != reflect.Slice {
		return bindingString(value)
	}
	parts := make([]string, 0, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		parts = append(parts, bindingString(rv.Index(i).Interface()))
	}
	return strings.Join(parts, sep)
}

func bindingValuesEqual(left, right interface{}) bool {
	if lf, lok := bindingFloat(left); lok {
		if rf, rok := bindingFloat(right); rok {
			return lf == rf
		}
	}
	return bindingString(left) == bindingString(right)
}

func normalizeBindingNumber(value float64) interface{} {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0
	}
	if math.Trunc(value) == value {
		return int(value)
	}
	return value
}

func bindingString(value interface{}) string {
	if value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func bindingTruthy(value interface{}) bool {
	switch v := value.(type) {
	case nil:
		return false
	case bool:
		return v
	case string:
		return v != "" && strings.ToLower(v) != "false" && v != "0"
	case int:
		return v != 0
	case int64:
		return v != 0
	case float64:
		return v != 0
	default:
		return true
	}
}
