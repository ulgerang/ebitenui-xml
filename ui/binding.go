package ui

import (
	"fmt"
	"reflect"
	"sync"
)

// ============================================================================
// Data Binding System - Reactive UI Updates
// ============================================================================

// Observable represents a value that can be watched for changes
type Observable[T any] struct {
	value     T
	listeners []func(T)
	mu        sync.RWMutex
}

// NewObservable creates a new observable value
func NewObservable[T any](initial T) *Observable[T] {
	return &Observable[T]{
		value:     initial,
		listeners: make([]func(T), 0),
	}
}

// Get returns the current value
func (o *Observable[T]) Get() T {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.value
}

// Set updates the value and notifies listeners
func (o *Observable[T]) Set(value T) {
	o.mu.Lock()
	o.value = value
	listeners := make([]func(T), len(o.listeners))
	copy(listeners, o.listeners)
	o.mu.Unlock()

	// Notify listeners outside the lock
	for _, listener := range listeners {
		listener(value)
	}
}

// Subscribe adds a listener that will be called when the value changes
func (o *Observable[T]) Subscribe(listener func(T)) func() {
	o.mu.Lock()
	o.listeners = append(o.listeners, listener)
	idx := len(o.listeners) - 1
	o.mu.Unlock()

	// Return unsubscribe function
	return func() {
		o.mu.Lock()
		defer o.mu.Unlock()
		if idx < len(o.listeners) {
			o.listeners = append(o.listeners[:idx], o.listeners[idx+1:]...)
		}
	}
}

// ============================================================================
// BindingContext - Holds all bindings for a UI
// ============================================================================

// BindingContext manages data bindings for UI elements
type BindingContext struct {
	data     map[string]interface{}
	bindings map[string][]bindingEntry
	computed map[string]func() interface{}
	mu       sync.RWMutex
}

type bindingEntry struct {
	widget  Widget
	updater func(value interface{})
}

// NewBindingContext creates a new binding context
func NewBindingContext() *BindingContext {
	return &BindingContext{
		data:     make(map[string]interface{}),
		bindings: make(map[string][]bindingEntry),
		computed: make(map[string]func() interface{}),
	}
}

// Set sets a value and triggers bound widget updates
func (bc *BindingContext) Set(key string, value interface{}) {
	bc.mu.Lock()
	bc.data[key] = value
	entries := bc.bindings[key]
	bc.mu.Unlock()

	// Update bound widgets
	for _, entry := range entries {
		entry.updater(value)
	}
}

// Get retrieves a value
func (bc *BindingContext) Get(key string) interface{} {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.data[key]
}

// GetString retrieves a string value
func (bc *BindingContext) GetString(key string) string {
	v := bc.Get(key)
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

// GetInt retrieves an int value
func (bc *BindingContext) GetInt(key string) int {
	v := bc.Get(key)
	switch i := v.(type) {
	case int:
		return i
	case int64:
		return int(i)
	case float64:
		return int(i)
	}
	return 0
}

// GetFloat retrieves a float value
func (bc *BindingContext) GetFloat(key string) float64 {
	v := bc.Get(key)
	switch f := v.(type) {
	case float64:
		return f
	case float32:
		return float64(f)
	case int:
		return float64(f)
	}
	return 0
}

// GetBool retrieves a bool value
func (bc *BindingContext) GetBool(key string) bool {
	v := bc.Get(key)
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}

// Bind creates a one-way binding from data to a widget property
func (bc *BindingContext) Bind(key string, widget Widget, updater func(value interface{})) {
	bc.mu.Lock()
	bc.bindings[key] = append(bc.bindings[key], bindingEntry{
		widget:  widget,
		updater: updater,
	})

	// Initial update
	if value, exists := bc.data[key]; exists {
		bc.mu.Unlock()
		updater(value)
	} else {
		bc.mu.Unlock()
	}
}

// BindText binds a data key to a Text widget's content
func (bc *BindingContext) BindText(key string, textWidget *Text) {
	bc.Bind(key, textWidget, func(value interface{}) {
		textWidget.SetContent(fmt.Sprintf("%v", value))
	})
}

// BindProgress binds a data key to a ProgressBar widget's value
func (bc *BindingContext) BindProgress(key string, progressBar *ProgressBar) {
	bc.Bind(key, progressBar, func(value interface{}) {
		if f, ok := value.(float64); ok {
			progressBar.Value = f
		}
	})
}

// BindVisible binds a data key to a widget's visibility
func (bc *BindingContext) BindVisible(key string, widget Widget) {
	bc.Bind(key, widget, func(value interface{}) {
		if b, ok := value.(bool); ok {
			widget.SetVisible(b)
		}
	})
}

// BindEnabled binds a data key to a widget's enabled state
func (bc *BindingContext) BindEnabled(key string, widget Widget) {
	bc.Bind(key, widget, func(value interface{}) {
		if b, ok := value.(bool); ok {
			widget.SetEnabled(b)
		}
	})
}

// BindCheckbox binds a data key to a Checkbox widget (two-way)
func (bc *BindingContext) BindCheckbox(key string, checkbox *Checkbox) {
	// Data -> Widget
	bc.Bind(key, checkbox, func(value interface{}) {
		if b, ok := value.(bool); ok {
			checkbox.Checked = b
		}
	})

	// Widget -> Data
	originalOnChange := checkbox.OnChange
	checkbox.OnChange = func(checked bool) {
		bc.Set(key, checked)
		if originalOnChange != nil {
			originalOnChange(checked)
		}
	}
}

// BindSlider binds a data key to a Slider widget (two-way)
func (bc *BindingContext) BindSlider(key string, slider *Slider) {
	// Data -> Widget
	bc.Bind(key, slider, func(value interface{}) {
		if f, ok := value.(float64); ok {
			slider.Value = f
		}
	})

	// Widget -> Data
	originalOnChange := slider.OnChange
	slider.OnChange = func(value float64) {
		bc.Set(key, value)
		if originalOnChange != nil {
			originalOnChange(value)
		}
	}
}

// Computed creates a computed value that depends on other values
func (bc *BindingContext) Computed(key string, dependencies []string, compute func() interface{}) {
	bc.mu.Lock()
	bc.computed[key] = compute
	bc.mu.Unlock()

	// Update when any dependency changes
	for _, dep := range dependencies {
		bc.Bind(dep, nil, func(_ interface{}) {
			result := compute()
			bc.Set(key, result)
		})
	}

	// Initial computation
	bc.Set(key, compute())
}

// ============================================================================
// Model Binding - Bind structs directly to UI
// ============================================================================

// BindModel binds a struct to the binding context using struct field names as keys
func (bc *BindingContext) BindModel(prefix string, model interface{}) error {
	v := reflect.ValueOf(model)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return fmt.Errorf("model must be a struct")
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		if !fieldValue.CanInterface() {
			continue
		}

		// Use binding tag if present, otherwise use field name
		key := field.Tag.Get("binding")
		if key == "" {
			key = field.Name
		}

		if prefix != "" {
			key = prefix + "." + key
		}

		bc.Set(key, fieldValue.Interface())
	}

	return nil
}

// ============================================================================
// Expression Binding - Simple template expressions
// ============================================================================

// FormatBinding creates a formatted string from bound values
func (bc *BindingContext) FormatBinding(format string, keys ...string) *Observable[string] {
	result := NewObservable("")

	update := func() {
		values := make([]interface{}, len(keys))
		for i, key := range keys {
			values[i] = bc.Get(key)
		}
		result.Set(fmt.Sprintf(format, values...))
	}

	// Subscribe to all keys
	for _, key := range keys {
		k := key // Capture
		bc.Bind(k, nil, func(_ interface{}) {
			update()
		})
	}

	update()
	return result
}

// ============================================================================
// List Binding - For dynamic lists
// ============================================================================

// ListBinding manages a dynamic list of items
type ListBinding[T any] struct {
	items     []T
	listeners []func([]T)
	mu        sync.RWMutex
}

// NewListBinding creates a new list binding
func NewListBinding[T any]() *ListBinding[T] {
	return &ListBinding[T]{
		items:     make([]T, 0),
		listeners: make([]func([]T), 0),
	}
}

// Add adds an item to the list
func (lb *ListBinding[T]) Add(item T) {
	lb.mu.Lock()
	lb.items = append(lb.items, item)
	items := make([]T, len(lb.items))
	copy(items, lb.items)
	listeners := lb.listeners
	lb.mu.Unlock()

	for _, l := range listeners {
		l(items)
	}
}

// Remove removes an item at index
func (lb *ListBinding[T]) Remove(index int) {
	lb.mu.Lock()
	if index >= 0 && index < len(lb.items) {
		lb.items = append(lb.items[:index], lb.items[index+1:]...)
	}
	items := make([]T, len(lb.items))
	copy(items, lb.items)
	listeners := lb.listeners
	lb.mu.Unlock()

	for _, l := range listeners {
		l(items)
	}
}

// Clear removes all items
func (lb *ListBinding[T]) Clear() {
	lb.mu.Lock()
	lb.items = make([]T, 0)
	listeners := lb.listeners
	lb.mu.Unlock()

	for _, l := range listeners {
		l([]T{})
	}
}

// Get returns all items
func (lb *ListBinding[T]) Get() []T {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	result := make([]T, len(lb.items))
	copy(result, lb.items)
	return result
}

// Len returns the number of items
func (lb *ListBinding[T]) Len() int {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	return len(lb.items)
}

// Subscribe subscribes to list changes
func (lb *ListBinding[T]) Subscribe(listener func([]T)) func() {
	lb.mu.Lock()
	lb.listeners = append(lb.listeners, listener)
	idx := len(lb.listeners) - 1
	lb.mu.Unlock()

	return func() {
		lb.mu.Lock()
		defer lb.mu.Unlock()
		if idx < len(lb.listeners) {
			lb.listeners = append(lb.listeners[:idx], lb.listeners[idx+1:]...)
		}
	}
}
