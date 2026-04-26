package ui

import (
	"strings"
	"sync"
	"testing"
	"time"
)

// ============================================================================
// Observable Tests
// ============================================================================

func TestObservable(t *testing.T) {
	t.Run("get initial value", func(t *testing.T) {
		obs := NewObservable(42)
		if got := obs.Get(); got != 42 {
			t.Errorf("Get() = %v, want 42", got)
		}
	})

	t.Run("set and get", func(t *testing.T) {
		obs := NewObservable(10)
		obs.Set(20)
		if got := obs.Get(); got != 20 {
			t.Errorf("Get() = %v, want 20", got)
		}
	})

	t.Run("notify single listener", func(t *testing.T) {
		obs := NewObservable(0)
		called := false
		var value int

		obs.Subscribe(func(v int) {
			called = true
			value = v
		})

		obs.Set(42)

		if !called {
			t.Error("Listener was not called")
		}
		if value != 42 {
			t.Errorf("Listener received %v, want 42", value)
		}
	})

	t.Run("notify multiple listeners", func(t *testing.T) {
		obs := NewObservable(0)
		count := 0

		obs.Subscribe(func(v int) { count++ })
		obs.Subscribe(func(v int) { count++ })
		obs.Subscribe(func(v int) { count++ })

		obs.Set(100)

		if count != 3 {
			t.Errorf("Expected 3 calls, got %d", count)
		}
	})

	t.Run("unsubscribe", func(t *testing.T) {
		obs := NewObservable(0)
		count := 0

		unsub := obs.Subscribe(func(v int) { count++ })

		obs.Set(1) // Should call
		unsub()    // Unsubscribe
		obs.Set(2) // Should not call

		if count != 1 {
			t.Errorf("Expected 1 call, got %d", count)
		}
	})

	t.Run("unsubscribe all", func(t *testing.T) {
		obs := NewObservable(0)

		unsub1 := obs.Subscribe(func(v int) {})
		unsub2 := obs.Subscribe(func(v int) {})

		unsub1()
		unsub2()

		// Should not panic
		obs.Set(999)
	})

	t.Run("set string value", func(t *testing.T) {
		obs := NewObservable("hello")
		if got := obs.Get(); got != "hello" {
			t.Errorf("Get() = %v, want hello", got)
		}

		obs.Set("world")
		if got := obs.Get(); got != "world" {
			t.Errorf("Get() = %v, want world", got)
		}
	})

	t.Run("set bool value", func(t *testing.T) {
		obs := NewObservable(true)
		if !obs.Get() {
			t.Error("Get() = false, want true")
		}

		obs.Set(false)
		if obs.Get() {
			t.Error("Get() = true, want false")
		}
	})
}

func TestObservableThreadSafety(t *testing.T) {
	t.Run("concurrent reads", func(t *testing.T) {
		obs := NewObservable(0)
		var wg sync.WaitGroup

		// Start multiple goroutines reading concurrently
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 100; j++ {
					_ = obs.Get()
				}
			}()
		}

		wg.Wait()
		// Should not panic or deadlock
	})

	t.Run("concurrent writes", func(t *testing.T) {
		obs := NewObservable(0)
		var wg sync.WaitGroup

		// Start multiple goroutines writing concurrently
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(val int) {
				defer wg.Done()
				for j := 0; j < 100; j++ {
					obs.Set(val)
				}
			}(i)
		}

		wg.Wait()
		// Should not panic or deadlock
	})

	t.Run("concurrent read and write", func(t *testing.T) {
		obs := NewObservable(0)
		var wg sync.WaitGroup
		done := make(chan bool)

		// Writer goroutine
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				obs.Set(i)
				select {
				case <-done:
					return
				default:
				}
			}
		}()

		// Reader goroutine
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				_ = obs.Get()
				select {
				case <-done:
					return
				default:
				}
			}
		}()

		// Let them run for a bit
		time.Sleep(100 * time.Millisecond)
		close(done)
		wg.Wait()
		// Should not panic or deadlock
	})

	t.Run("concurrent subscribe and set", func(t *testing.T) {
		obs := NewObservable(0)
		var wg sync.WaitGroup
		done := make(chan bool)

		// Subscriber goroutine
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				unsub := obs.Subscribe(func(v int) {})
				if i%10 == 0 {
					unsub()
				}
				select {
				case <-done:
					return
				default:
				}
			}
		}()

		// Writer goroutine
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				obs.Set(i)
				select {
				case <-done:
					return
				default:
				}
			}
		}()

		time.Sleep(100 * time.Millisecond)
		close(done)
		wg.Wait()
		// Should not panic or deadlock
	})
}

// ============================================================================
// BindingContext Tests
// ============================================================================

func TestBindingContext(t *testing.T) {
	t.Run("set and get", func(t *testing.T) {
		bc := NewBindingContext()
		bc.Set("key", "value")

		if got := bc.Get("key"); got != "value" {
			t.Errorf("Get() = %v, want value", got)
		}
	})

	t.Run("get nonexistent key", func(t *testing.T) {
		bc := NewBindingContext()
		if got := bc.Get("nonexistent"); got != nil {
			t.Errorf("Get() = %v, want nil", got)
		}
	})

	t.Run("GetString", func(t *testing.T) {
		bc := NewBindingContext()
		bc.Set("name", "John")

		if got := bc.GetString("name"); got != "John" {
			t.Errorf("GetString() = %v, want John", got)
		}
	})

	t.Run("GetString conversion", func(t *testing.T) {
		bc := NewBindingContext()
		bc.Set("number", 42)

		// Should convert to string
		got := bc.GetString("number")
		if got != "42" {
			t.Errorf("GetString() = %v, want 42", got)
		}
	})

	t.Run("GetInt", func(t *testing.T) {
		bc := NewBindingContext()
		bc.Set("count", 100)

		if got := bc.GetInt("count"); got != 100 {
			t.Errorf("GetInt() = %v, want 100", got)
		}
	})

	t.Run("GetInt from float", func(t *testing.T) {
		bc := NewBindingContext()
		bc.Set("value", 3.14)

		if got := bc.GetInt("value"); got != 3 {
			t.Errorf("GetInt() = %v, want 3", got)
		}
	})

	t.Run("GetFloat", func(t *testing.T) {
		bc := NewBindingContext()
		bc.Set("pi", 3.14159)

		got := bc.GetFloat("pi")
		if got < 3.14 || got > 3.15 {
			t.Errorf("GetFloat() = %v, want ~3.14", got)
		}
	})

	t.Run("GetBool", func(t *testing.T) {
		bc := NewBindingContext()
		bc.Set("active", true)

		if !bc.GetBool("active") {
			t.Error("GetBool() = false, want true")
		}
	})

	t.Run("GetBool false", func(t *testing.T) {
		bc := NewBindingContext()
		bc.Set("active", false)

		if bc.GetBool("active") {
			t.Error("GetBool() = true, want false")
		}
	})

	t.Run("Bind and update", func(t *testing.T) {
		bc := NewBindingContext()
		var lastValue interface{}

		widget := NewPanel("test")
		bc.Bind("value", widget, func(v interface{}) {
			lastValue = v
		})

		bc.Set("value", "updated")

		if lastValue != "updated" {
			t.Errorf("Binding not updated, got %v", lastValue)
		}
	})

	t.Run("Bind initial value", func(t *testing.T) {
		bc := NewBindingContext()
		bc.Set("initial", "start")

		var received interface{}
		widget := NewPanel("test")
		bc.Bind("initial", widget, func(v interface{}) {
			received = v
		})

		if received != "start" {
			t.Errorf("Initial binding value = %v, want start", received)
		}
	})

	t.Run("BindText", func(t *testing.T) {
		bc := NewBindingContext()
		textWidget := NewText("text", "original")

		bc.BindText("message", textWidget)
		bc.Set("message", "updated")

		if textWidget.Content != "updated" {
			t.Errorf("Text content = %v, want updated", textWidget.Content)
		}
	})

	t.Run("BindProgress", func(t *testing.T) {
		bc := NewBindingContext()
		progressBar := NewProgressBar("progress")

		bc.BindProgress("progress", progressBar)
		bc.Set("progress", 0.75)

		if progressBar.Value != 0.75 {
			t.Errorf("Progress value = %v, want 0.75", progressBar.Value)
		}
	})

	t.Run("BindVisible", func(t *testing.T) {
		bc := NewBindingContext()
		widget := NewPanel("panel")
		widget.SetVisible(true)

		bc.BindVisible("visible", widget)
		bc.Set("visible", false)

		if widget.Visible() {
			t.Error("Widget should be invisible")
		}
	})

	t.Run("BindEnabled", func(t *testing.T) {
		bc := NewBindingContext()
		widget := NewPanel("panel")
		widget.SetEnabled(true)

		bc.BindEnabled("enabled", widget)
		bc.Set("enabled", false)

		if widget.Enabled() {
			t.Error("Widget should be disabled")
		}
	})

	t.Run("BindCheckbox", func(t *testing.T) {
		bc := NewBindingContext()
		checkbox := NewCheckbox("check", "Label")

		bc.BindCheckbox("checked", checkbox)

		// Initial set
		bc.Set("checked", true)
		if !checkbox.Checked {
			t.Error("Checkbox should be checked")
		}

		// Widget -> Data
		checkbox.HandleClick()
		if got := bc.GetBool("checked"); got {
			t.Error("Checkbox data should be false after toggle")
		}
	})

	t.Run("BindSlider", func(t *testing.T) {
		bc := NewBindingContext()
		slider := NewSlider("slider")
		slider.Min = 0
		slider.Max = 100

		bc.BindSlider("value", slider)

		// Initial set
		bc.Set("value", 75.0)
		if slider.Value != 75.0 {
			t.Errorf("Slider value = %v, want 75.0", slider.Value)
		}

		// Widget -> Data
		slider.SetValue(50.0)
		if got := bc.GetFloat("value"); got != 50.0 {
			t.Errorf("Slider data = %v, want 50.0", got)
		}
	})
}

func TestDeclarativeXMLBindings(t *testing.T) {
	t.Run("bind text visible and enabled from XML", func(t *testing.T) {
		ui := New(320, 200)
		err := ui.LoadLayout(`
			<panel id="root">
				<text id="name" bind-text="player.name" bind-visible="player.visible" bind-enabled="player.enabled">fallback</text>
			</panel>
		`)
		if err != nil {
			t.Fatalf("LoadLayout() error = %v", err)
		}

		ui.Bind("player.name", "Ada")
		ui.Bind("player.visible", false)
		ui.Bind("player.enabled", false)

		name := ui.GetText("name")
		if name == nil {
			t.Fatal("expected text widget")
		}
		if name.Content != "Ada" {
			t.Errorf("Content = %q, want Ada", name.Content)
		}
		if name.Visible() {
			t.Error("text should be invisible")
		}
		if name.Enabled() {
			t.Error("text should be disabled")
		}
	})

	t.Run("bind input value two-way from XML", func(t *testing.T) {
		ui := New(320, 200)
		err := ui.LoadLayout(`<input id="username" bind-value="user.name" />`)
		if err != nil {
			t.Fatalf("LoadLayout() error = %v", err)
		}

		ui.Bind("user.name", "Ada")

		input := ui.GetTextInput("username")
		if input == nil {
			t.Fatal("expected text input")
		}
		if input.Text != "Ada" {
			t.Errorf("Text = %q, want Ada", input.Text)
		}

		input.SetText("")
		input.insertString("Grace")
		if got := ui.Bindings().GetString("user.name"); got != "Grace" {
			t.Errorf("binding value = %q, want Grace", got)
		}
	})

	t.Run("bind checkbox checked two-way from XML alias", func(t *testing.T) {
		ui := New(320, 200)
		err := ui.LoadLayout(`<checkbox id="music" data-bind-checked="settings.music">Music</checkbox>`)
		if err != nil {
			t.Fatalf("LoadLayout() error = %v", err)
		}

		ui.Bind("settings.music", true)

		checkbox := ui.GetCheckbox("music")
		if checkbox == nil {
			t.Fatal("expected checkbox")
		}
		if !checkbox.Checked {
			t.Error("checkbox should be checked")
		}

		checkbox.HandleClick()
		if got := ui.Bindings().GetBool("settings.music"); got {
			t.Error("binding value should be false after click")
		}
	})

	t.Run("bind text template expressions from XML", func(t *testing.T) {
		ui := New(320, 200)
		err := ui.LoadLayout(`<text id="status">Player {{player.name}} Lv.{{player.level}}</text>`)
		if err != nil {
			t.Fatalf("LoadLayout() error = %v", err)
		}

		ui.Bind("player.name", "Ada")
		ui.Bind("player.level", 7)

		status := ui.GetText("status")
		if status == nil {
			t.Fatal("expected text widget")
		}
		if status.Content != "Player Ada Lv.7" {
			t.Errorf("Content = %q, want Player Ada Lv.7", status.Content)
		}

		ui.Bind("player.level", 8)
		if status.Content != "Player Ada Lv.8" {
			t.Errorf("Content = %q, want Player Ada Lv.8", status.Content)
		}
	})

	t.Run("bind repeat creates children from XML template", func(t *testing.T) {
		type player struct {
			Name  string
			Level int
		}

		ui := New(320, 200)
		err := ui.LoadLayout(`
			<panel id="root">
				<text id="player-{{index}}" bind-repeat="players">{{item.Name}} Lv.{{item.Level}}</text>
			</panel>
		`)
		if err != nil {
			t.Fatalf("LoadLayout() error = %v", err)
		}

		ui.Bind("players", []player{
			{Name: "Ada", Level: 7},
			{Name: "Grace", Level: 9},
		})

		root := ui.GetPanel("root")
		if root == nil {
			t.Fatal("expected root panel")
		}
		if got := len(root.Children()); got != 2 {
			t.Fatalf("len(root.Children()) = %d, want 2", got)
		}
		first := ui.GetText("player-0")
		second := ui.GetText("player-1")
		if first == nil || second == nil {
			t.Fatal("expected repeated text widgets in ID cache")
		}
		if first.Content != "Ada Lv.7" {
			t.Errorf("first.Content = %q, want Ada Lv.7", first.Content)
		}
		if second.Content != "Grace Lv.9" {
			t.Errorf("second.Content = %q, want Grace Lv.9", second.Content)
		}

		ui.Bind("players", []player{{Name: "Linus", Level: 11}})
		if got := len(root.Children()); got != 1 {
			t.Fatalf("len(root.Children()) after update = %d, want 1", got)
		}
		updated := ui.GetText("player-0")
		if updated == nil {
			t.Fatal("expected updated repeated text widget in ID cache")
		}
		if updated.Content != "Linus Lv.11" {
			t.Errorf("updated.Content = %q, want Linus Lv.11", updated.Content)
		}
		if stale := ui.GetText("player-1"); stale != nil {
			t.Fatal("stale repeated widget should be removed from ID cache")
		}
	})

	t.Run("bind repeat supports map items and for-each alias", func(t *testing.T) {
		ui := New(320, 200)
		err := ui.LoadLayout(`
			<panel id="root">
				<button id="quest-{{index}}" for-each="quests" label="{{item.title}}" />
			</panel>
		`)
		if err != nil {
			t.Fatalf("LoadLayout() error = %v", err)
		}

		ui.Bind("quests", []map[string]interface{}{
			{"title": "Find key"},
			{"title": "Open door"},
		})

		first := ui.GetButton("quest-0")
		second := ui.GetButton("quest-1")
		if first == nil || second == nil {
			t.Fatal("expected repeated buttons")
		}
		if first.Label != "Find key" {
			t.Errorf("first.Label = %q, want Find key", first.Label)
		}
		if second.Label != "Open door" {
			t.Errorf("second.Label = %q, want Open door", second.Label)
		}
	})

	t.Run("bind if attaches and detaches XML child", func(t *testing.T) {
		ui := New(320, 200)
		err := ui.LoadLayout(`
			<panel id="root">
				<text id="warning" bind-if="showWarning">Warning</text>
			</panel>
		`)
		if err != nil {
			t.Fatalf("LoadLayout() error = %v", err)
		}

		root := ui.GetPanel("root")
		if root == nil {
			t.Fatal("expected root panel")
		}
		if got := len(root.Children()); got != 0 {
			t.Fatalf("len(root.Children()) before binding = %d, want 0", got)
		}
		if warning := ui.GetText("warning"); warning != nil {
			t.Fatal("conditional widget should not exist before truthy binding")
		}

		ui.Bind("showWarning", true)
		if got := len(root.Children()); got != 1 {
			t.Fatalf("len(root.Children()) after true = %d, want 1", got)
		}
		warning := ui.GetText("warning")
		if warning == nil {
			t.Fatal("conditional widget should exist in ID cache after true")
		}
		if warning.Content != "Warning" {
			t.Errorf("warning.Content = %q, want Warning", warning.Content)
		}

		ui.Bind("showWarning", false)
		if got := len(root.Children()); got != 0 {
			t.Fatalf("len(root.Children()) after false = %d, want 0", got)
		}
		if warning := ui.GetText("warning"); warning != nil {
			t.Fatal("conditional widget should be removed from ID cache after false")
		}
	})

	t.Run("bind if supports data alias and truthy values", func(t *testing.T) {
		ui := New(320, 200)
		err := ui.LoadLayout(`
			<panel id="root">
				<button id="start" data-bind-if="canStart">Start</button>
			</panel>
		`)
		if err != nil {
			t.Fatalf("LoadLayout() error = %v", err)
		}

		ui.Bind("canStart", "yes")
		if start := ui.GetButton("start"); start == nil {
			t.Fatal("conditional button should exist for truthy string")
		}

		ui.Bind("canStart", 0)
		if start := ui.GetButton("start"); start != nil {
			t.Fatal("conditional button should be removed for numeric zero")
		}
	})

	t.Run("rich template expressions update from dependencies", func(t *testing.T) {
		ui := New(320, 200)
		err := ui.LoadLayout(`
			<panel id="root">
				<text id="greeting">Hello {{upper(user.name || "guest")}}</text>
				<text id="summary">{{count + 1}}/{{total}}</text>
			</panel>
		`)
		if err != nil {
			t.Fatalf("LoadLayout() error = %v", err)
		}

		greeting := ui.GetText("greeting")
		summary := ui.GetText("summary")
		if greeting == nil || summary == nil {
			t.Fatal("expected expression-bound text widgets")
		}
		if greeting.Content != "Hello GUEST" {
			t.Errorf("initial greeting = %q, want Hello GUEST", greeting.Content)
		}

		ui.Bind("user.name", "Ada")
		ui.Bind("count", 2)
		ui.Bind("total", 5)
		if greeting.Content != "Hello ADA" {
			t.Errorf("updated greeting = %q, want Hello ADA", greeting.Content)
		}
		if summary.Content != "3/5" {
			t.Errorf("summary = %q, want 3/5", summary.Content)
		}
	})

	t.Run("binding expression helpers format common values", func(t *testing.T) {
		ui := New(320, 200)
		err := ui.LoadLayout(`
			<panel id="root">
				<text id="count">Items {{len(items)}}</text>
				<text id="rounded">{{round(score, 1)}}/{{floor(score)}}/{{ceil(score)}}</text>
				<text id="contains">{{contains(title, "Quest")}} {{contains(items, "bow")}}</text>
				<text id="joined">{{join(items, " | ")}}</text>
				<text id="formatted">{{format("%s:%d", player.name, level)}}</text>
			</panel>
		`)
		if err != nil {
			t.Fatalf("LoadLayout() error = %v", err)
		}

		ui.Bind("items", []string{"sword", "bow", "potion"})
		ui.Bind("score", 12.26)
		ui.Bind("title", "Quest Board")
		ui.Bind("player.name", "Ada")
		ui.Bind("level", 7)

		if got := ui.GetText("count").Content; got != "Items 3" {
			t.Fatalf("count = %q, want Items 3", got)
		}
		if got := ui.GetText("rounded").Content; got != "12.3/12/13" {
			t.Fatalf("rounded = %q, want 12.3/12/13", got)
		}
		if got := ui.GetText("contains").Content; got != "true true" {
			t.Fatalf("contains = %q, want true true", got)
		}
		if got := ui.GetText("joined").Content; got != "sword | bow | potion" {
			t.Fatalf("joined = %q, want sword | bow | potion", got)
		}
		if got := ui.GetText("formatted").Content; got != "Ada:7" {
			t.Fatalf("formatted = %q, want Ada:7", got)
		}

		ui.Bind("items", []string{"wand"})
		if got := ui.GetText("count").Content; got != "Items 1" {
			t.Fatalf("updated count = %q, want Items 1", got)
		}
		if got := ui.GetText("contains").Content; got != "true false" {
			t.Fatalf("updated contains = %q, want true false", got)
		}
	})

	t.Run("attribute and style bindings update widgets", func(t *testing.T) {
		ui := New(320, 200)
		err := ui.LoadLayout(`
			<button id="cta"
				bind-attr-label="label || 'Fallback'"
				bind-attr-width="button.width"
				bind-style-opacity="enabled &amp;&amp; 1 || 0.5"
				bind-style-color="color">Fallback</button>
		`)
		if err != nil {
			t.Fatalf("LoadLayout() error = %v", err)
		}

		ui.Bind("label", "Start")
		ui.Bind("button.width", 120)
		ui.Bind("enabled", false)
		ui.Bind("color", "#ff0000")

		button := ui.GetButton("cta")
		if button == nil {
			t.Fatal("expected button")
		}
		if button.Label != "Start" {
			t.Errorf("Label = %q, want Start", button.Label)
		}
		if button.Style().Width != 120 || !button.Style().WidthSet {
			t.Errorf("Width = %v set=%v, want 120 true", button.Style().Width, button.Style().WidthSet)
		}
		if button.Style().Opacity != 0.5 {
			t.Errorf("Opacity = %v, want 0.5", button.Style().Opacity)
		}
		if button.Style().Color != "#ff0000" {
			t.Errorf("Color = %q, want #ff0000", button.Style().Color)
		}

		ui.Bind("enabled", true)
		if button.Style().Opacity != 1 {
			t.Errorf("Opacity after enabled = %v, want 1", button.Style().Opacity)
		}
	})

	t.Run("xml event commands dispatch registered handlers", func(t *testing.T) {
		ui := New(320, 200)
		err := ui.LoadLayout(`<button id="save" onClick="saveGame">Save</button>`)
		if err != nil {
			t.Fatalf("LoadLayout() error = %v", err)
		}

		var called bool
		var got Widget
		ui.RegisterCommand("saveGame", func(widget Widget) {
			called = true
			got = widget
		})

		button := ui.GetButton("save")
		if button == nil {
			t.Fatal("expected button")
		}
		button.HandleClick()

		if !called {
			t.Fatal("command handler was not called")
		}
		if got != button {
			t.Fatalf("command widget = %v, want button", got)
		}
	})

	t.Run("dropdown options bind from primitive collection", func(t *testing.T) {
		ui := New(320, 200)
		err := ui.LoadLayout(`<dropdown id="choice" bind-options="items" bind-value="selected" />`)
		if err != nil {
			t.Fatalf("LoadLayout() error = %v", err)
		}

		ui.Bind("items", []string{"Easy", "Normal", "Hard"})
		ui.Bind("selected", "Normal")

		dropdown := ui.GetWidget("choice").(*Dropdown)
		if len(dropdown.Options) != 3 {
			t.Fatalf("option count = %d, want 3", len(dropdown.Options))
		}
		if got := dropdown.GetSelectedValue(); got != "Normal" {
			t.Fatalf("selected value = %q, want Normal", got)
		}

		ui.Bind("items", []string{"Normal", "Expert"})
		if len(dropdown.Options) != 2 {
			t.Fatalf("updated option count = %d, want 2", len(dropdown.Options))
		}
		if got := dropdown.GetSelectedValue(); got != "Normal" {
			t.Fatalf("selected value after refresh = %q, want Normal", got)
		}
	})

	t.Run("dropdown options bind from mapped collection", func(t *testing.T) {
		ui := New(320, 200)
		err := ui.LoadLayout(`<select id="hero" bind-options="heroes" option-label="name" option-value="id" />`)
		if err != nil {
			t.Fatalf("LoadLayout() error = %v", err)
		}

		ui.Bind("heroes", []map[string]interface{}{
			{"id": "h1", "name": "Ada"},
			{"id": "h2", "name": "Grace"},
		})

		dropdown := ui.GetWidget("hero").(*Dropdown)
		if len(dropdown.Options) != 2 {
			t.Fatalf("option count = %d, want 2", len(dropdown.Options))
		}
		if dropdown.Options[1].Label != "Grace" || dropdown.Options[1].Value != "h2" {
			t.Fatalf("mapped option = %+v, want Grace/h2", dropdown.Options[1])
		}
	})

	t.Run("radio options bind from mapped collection", func(t *testing.T) {
		ui := New(320, 200)
		err := ui.LoadLayout(`<panel id="difficulty" bind-options="levels" option-type="radio" option-label="label" option-value="id" bind-value="selected" />`)
		if err != nil {
			t.Fatalf("LoadLayout() error = %v", err)
		}

		ui.Bind("selected", "normal")
		ui.Bind("levels", []map[string]interface{}{
			{"id": "easy", "label": "Easy"},
			{"id": "normal", "label": "Normal"},
			{"id": "hard", "label": "Hard"},
		})

		container := ui.GetPanel("difficulty")
		if container == nil {
			t.Fatal("expected radio option container")
		}
		if len(container.Children()) != 3 {
			t.Fatalf("radio option count = %d, want 3", len(container.Children()))
		}
		normal, ok := ui.GetWidget("difficulty-option-normal").(*RadioButton)
		if !ok {
			t.Fatal("expected generated normal radio button")
		}
		if !normal.Selected || normal.Label != "Normal" || normal.Value != "normal" {
			t.Fatalf("normal radio = %+v, want selected Normal/normal", normal)
		}

		hard, ok := ui.GetWidget("difficulty-option-hard").(*RadioButton)
		if !ok {
			t.Fatal("expected generated hard radio button")
		}
		hard.HandleClick()
		if got := ui.Bindings().GetString("selected"); got != "hard" {
			t.Fatalf("selected binding after click = %q, want hard", got)
		}

		ui.Bind("levels", []map[string]interface{}{
			{"id": "normal", "label": "Normal"},
			{"id": "hard", "label": "Hard"},
		})
		if len(container.Children()) != 2 {
			t.Fatalf("updated radio option count = %d, want 2", len(container.Children()))
		}
		refreshed, ok := ui.GetWidget("difficulty-option-hard").(*RadioButton)
		if !ok || !refreshed.Selected {
			t.Fatalf("refreshed hard radio = %+v, want selected", refreshed)
		}
		if ui.GetWidget("difficulty-option-easy") != nil {
			t.Fatal("removed radio option should leave widget cache")
		}
	})

	t.Run("checkbox options bind from primitive collection", func(t *testing.T) {
		ui := New(320, 200)
		err := ui.LoadLayout(`<panel id="filters" bind-options="tags" option-type="checkbox" bind-value="selected" />`)
		if err != nil {
			t.Fatalf("LoadLayout() error = %v", err)
		}

		ui.Bind("selected", []string{"Fast"})
		ui.Bind("tags", []string{"Fast", "Safe", "Cheap"})

		container := ui.GetPanel("filters")
		if container == nil {
			t.Fatal("expected checkbox option container")
		}
		if len(container.Children()) != 3 {
			t.Fatalf("checkbox option count = %d, want 3", len(container.Children()))
		}
		fast, ok := ui.GetWidget("filters-option-fast").(*Checkbox)
		if !ok {
			t.Fatal("expected generated Fast checkbox")
		}
		if !fast.Checked || fast.Label != "Fast" {
			t.Fatalf("fast checkbox = %+v, want checked Fast", fast)
		}

		safe := ui.GetWidget("filters-option-safe").(*Checkbox)
		safe.HandleClick()
		selected, ok := ui.Bindings().Get("selected").([]string)
		if !ok {
			t.Fatalf("selected binding type = %T, want []string", ui.Bindings().Get("selected"))
		}
		if strings.Join(selected, ",") != "Fast,Safe" {
			t.Fatalf("selected binding = %v, want [Fast Safe]", selected)
		}

		ui.Bind("tags", []string{"Safe", "Cheap"})
		if len(container.Children()) != 2 {
			t.Fatalf("updated checkbox option count = %d, want 2", len(container.Children()))
		}
		refreshed, ok := ui.GetWidget("filters-option-safe").(*Checkbox)
		if !ok || !refreshed.Checked {
			t.Fatalf("refreshed safe checkbox = %+v, want checked", refreshed)
		}
		if ui.GetWidget("filters-option-fast") != nil {
			t.Fatal("removed checkbox option should leave widget cache")
		}
	})

	t.Run("checkbox options bind from mapped and struct collections", func(t *testing.T) {
		type tagOption struct {
			ID    string
			Label string
		}

		ui := New(320, 200)
		err := ui.LoadLayout(`<panel id="features" bind-options="items" option-type="checkbox" option-label="label" option-value="id" option-id-prefix="feature" bind-value="selected" />`)
		if err != nil {
			t.Fatalf("LoadLayout() error = %v", err)
		}

		ui.Bind("selected", []interface{}{"a"})
		ui.Bind("items", []map[string]interface{}{
			{"id": "a", "label": "Alpha"},
			{"id": "b", "label": "Beta"},
		})
		alpha, ok := ui.GetWidget("feature-option-a").(*Checkbox)
		if !ok || !alpha.Checked || alpha.Label != "Alpha" {
			t.Fatalf("mapped alpha checkbox = %+v, want checked Alpha", alpha)
		}

		ui.Bind("items", []tagOption{
			{ID: "a", Label: "Alpha"},
			{ID: "c", Label: "Gamma"},
		})
		gamma, ok := ui.GetWidget("feature-option-c").(*Checkbox)
		if !ok || gamma.Checked || gamma.Label != "Gamma" {
			t.Fatalf("struct gamma checkbox = %+v, want unchecked Gamma", gamma)
		}
	})

	t.Run("binding diagnostics include widget and attribute", func(t *testing.T) {
		ui := New(320, 200)
		err := ui.LoadLayout(`<dropdown id="bad" bind-options="missing + " />`)
		if err != nil {
			t.Fatalf("LoadLayout() error = %v", err)
		}

		diagnostics := ui.Bindings().Diagnostics()
		if len(diagnostics) == 0 {
			t.Fatal("expected binding diagnostic")
		}
		if diagnostics[0].WidgetID != "bad" || diagnostics[0].Attribute != "bind-options" {
			t.Fatalf("diagnostic = %+v, want widget bad bind-options", diagnostics[0])
		}
	})
}

func TestBindingContextComputed(t *testing.T) {
	t.Run("computed value", func(t *testing.T) {
		bc := NewBindingContext()
		bc.Set("first", "Hello")
		bc.Set("last", "World")

		var computed interface{}
		bc.Computed("full", []string{"first", "last"}, func() interface{} {
			computed = bc.GetString("first") + " " + bc.GetString("last")
			return computed
		})

		if computed != "Hello World" {
			t.Errorf("Computed value = %v, want 'Hello World'", computed)
		}
	})

	t.Run("computed updates on dependency change", func(t *testing.T) {
		bc := NewBindingContext()
		bc.Set("a", 10)
		bc.Set("b", 20)

		var result int
		bc.Computed("sum", []string{"a", "b"}, func() interface{} {
			result = bc.GetInt("a") + bc.GetInt("b")
			return result
		})

		if result != 30 {
			t.Errorf("Initial sum = %v, want 30", result)
		}

		bc.Set("a", 15)
		if result != 35 {
			t.Errorf("Updated sum = %v, want 35", result)
		}
	})
}

func TestBindingContextFormatBinding(t *testing.T) {
	t.Run("format binding", func(t *testing.T) {
		bc := NewBindingContext()
		bc.Set("name", "Alice")
		bc.Set("age", 30)

		obs := bc.FormatBinding("%s is %d years old", "name", "age")

		if got := obs.Get(); got != "Alice is 30 years old" {
			t.Errorf("FormatBinding() = %v, want 'Alice is 30 years old'", got)
		}
	})

	t.Run("format binding updates", func(t *testing.T) {
		bc := NewBindingContext()
		bc.Set("name", "Bob")

		obs := bc.FormatBinding("Hello, %s!", "name")

		if obs.Get() != "Hello, Bob!" {
			t.Errorf("Initial = %v, want 'Hello, Bob!'", obs.Get())
		}

		bc.Set("name", "Carol")
		if got := obs.Get(); got != "Hello, Carol!" {
			t.Errorf("Updated = %v, want 'Hello, Carol!'", got)
		}
	})
}

// ============================================================================
// ListBinding Tests
// ============================================================================

func TestListBinding(t *testing.T) {
	t.Run("add item", func(t *testing.T) {
		lb := NewListBinding[string]()
		lb.Add("item1")

		items := lb.Get()
		if len(items) != 1 {
			t.Errorf("len(Get()) = %d, want 1", len(items))
		}
		if items[0] != "item1" {
			t.Errorf("Get()[0] = %v, want item1", items[0])
		}
	})

	t.Run("add multiple items", func(t *testing.T) {
		lb := NewListBinding[int]()
		lb.Add(1)
		lb.Add(2)
		lb.Add(3)

		if got := lb.Len(); got != 3 {
			t.Errorf("Len() = %d, want 3", got)
		}
	})

	t.Run("remove item", func(t *testing.T) {
		lb := NewListBinding[string]()
		lb.Add("a")
		lb.Add("b")
		lb.Add("c")

		lb.Remove(1) // Remove "b"

		items := lb.Get()
		if len(items) != 2 {
			t.Errorf("len(Get()) = %d, want 2", len(items))
		}
		if items[0] != "a" || items[1] != "c" {
			t.Errorf("Get() = %v, want [a, c]", items)
		}
	})

	t.Run("remove out of bounds", func(t *testing.T) {
		lb := NewListBinding[int]()
		lb.Add(1)

		// Should not panic
		lb.Remove(100)
		lb.Remove(-1)

		if got := lb.Len(); got != 1 {
			t.Errorf("Len() = %d, want 1", got)
		}
	})

	t.Run("clear", func(t *testing.T) {
		lb := NewListBinding[string]()
		lb.Add("a")
		lb.Add("b")

		lb.Clear()

		if got := lb.Len(); got != 0 {
			t.Errorf("Len() = %d, want 0", got)
		}
		if got := lb.Get(); len(got) != 0 {
			t.Errorf("Get() = %v, want []", got)
		}
	})

	t.Run("subscribe to changes", func(t *testing.T) {
		lb := NewListBinding[int]()
		var callCount int
		var lastItems []int

		lb.Subscribe(func(items []int) {
			callCount++
			lastItems = items
		})

		lb.Add(10)
		if callCount != 1 {
			t.Errorf("Subscribe called %d times, want 1", callCount)
		}
		if len(lastItems) != 1 || lastItems[0] != 10 {
			t.Errorf("Last items = %v, want [10]", lastItems)
		}

		lb.Add(20)
		if callCount != 2 {
			t.Errorf("Subscribe called %d times, want 2", callCount)
		}
	})

	t.Run("unsubscribe", func(t *testing.T) {
		lb := NewListBinding[int]()
		count := 0

		unsub := lb.Subscribe(func(items []int) { count++ })

		lb.Add(1) // Should call
		unsub()   // Unsubscribe
		lb.Add(2) // Should not call

		if count != 1 {
			t.Errorf("Subscribe called %d times, want 1", count)
		}
	})

	t.Run("clear notifies", func(t *testing.T) {
		lb := NewListBinding[string]()
		lb.Add("a")

		notified := false
		lb.Subscribe(func(items []string) {
			notified = true
			if len(items) != 0 {
				t.Errorf("Cleared items = %v, want []", items)
			}
		})

		lb.Clear()

		if !notified {
			t.Error("Subscribe was not notified on Clear")
		}
	})
}

func TestListBindingThreadSafety(t *testing.T) {
	t.Run("concurrent adds", func(t *testing.T) {
		lb := NewListBinding[int]()
		var wg sync.WaitGroup

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(val int) {
				defer wg.Done()
				for j := 0; j < 100; j++ {
					lb.Add(val)
				}
			}(i)
		}

		wg.Wait()
		// Should not panic or deadlock
	})

	t.Run("concurrent read and write", func(t *testing.T) {
		lb := NewListBinding[int]()
		var wg sync.WaitGroup
		done := make(chan bool)

		// Writer
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 500; i++ {
				lb.Add(i)
				select {
				case <-done:
					return
				default:
				}
			}
		}()

		// Reader
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 500; i++ {
				_ = lb.Get()
				_ = lb.Len()
				select {
				case <-done:
					return
				default:
				}
			}
		}()

		time.Sleep(100 * time.Millisecond)
		close(done)
		wg.Wait()
		// Should not panic or deadlock
	})
}

// ============================================================================
// BindModel Tests
// ============================================================================

func TestBindModel(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	t.Run("bind struct", func(t *testing.T) {
		bc := NewBindingContext()
		person := Person{Name: "John", Age: 30}

		err := bc.BindModel("", &person)
		if err != nil {
			t.Fatalf("BindModel() error = %v", err)
		}

		if got := bc.GetString("Name"); got != "John" {
			t.Errorf("Name = %v, want John", got)
		}
		if got := bc.GetInt("Age"); got != 30 {
			t.Errorf("Age = %v, want 30", got)
		}
	})

	t.Run("bind with prefix", func(t *testing.T) {
		bc := NewBindingContext()
		person := Person{Name: "Jane", Age: 25}

		err := bc.BindModel("user", &person)
		if err != nil {
			t.Fatalf("BindModel() error = %v", err)
		}

		if got := bc.GetString("user.Name"); got != "Jane" {
			t.Errorf("user.Name = %v, want Jane", got)
		}
	})

	t.Run("bind with custom tag", func(t *testing.T) {
		type User struct {
			Username string `binding:"login"`
			Email    string
		}

		bc := NewBindingContext()
		user := User{Username: "admin", Email: "admin@example.com"}

		err := bc.BindModel("", &user)
		if err != nil {
			t.Fatalf("BindModel() error = %v", err)
		}

		if got := bc.GetString("login"); got != "admin" {
			t.Errorf("login = %v, want admin", got)
		}
		if got := bc.GetString("Email"); got != "admin@example.com" {
			t.Errorf("Email = %v, want admin@example.com", got)
		}
	})

	t.Run("bind non-struct", func(t *testing.T) {
		bc := NewBindingContext()

		err := bc.BindModel("", "not a struct")
		if err == nil {
			t.Error("BindModel() error = nil, want error")
		}
	})

	t.Run("bind nil model", func(t *testing.T) {
		bc := NewBindingContext()

		var model interface{} = (*Person)(nil)
		err := bc.BindModel("", model)
		// Should handle gracefully
		_ = err
	})
}
