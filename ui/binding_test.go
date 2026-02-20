package ui

import (
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
