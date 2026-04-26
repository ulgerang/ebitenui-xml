package ui

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestHTMLSemanticAliasesAndQueries(t *testing.T) {
	ui := New(320, 240)
	err := ui.LoadLayout(`
		<main id="app">
			<nav id="nav" class="top">
				<h1 id="title">Title</h1>
				<ul id="menu">
					<li id="item-a">A</li>
					<li id="item-b">B</li>
				</ul>
			</nav>
		</main>
	`)
	if err != nil {
		t.Fatalf("LoadLayout() error = %v", err)
	}

	if ui.GetPanel("app") == nil {
		t.Fatal("main should parse as a panel")
	}
	title := ui.GetText("title")
	if title == nil {
		t.Fatal("h1 should parse as text")
	}
	if bw := baseWidgetOf(title); bw == nil || bw.SemanticType() != "h1" {
		t.Fatalf("h1 semantic type = %v, want h1", bw)
	}
	if got := len(ui.QueryByClass("top")); got != 1 {
		t.Fatalf("QueryByClass(top) = %d, want 1", got)
	}
	if got := len(ui.QueryByType("li")); got != 2 {
		t.Fatalf("QueryByType(li) = %d, want 2", got)
	}
	if got := ui.Query(ui.Root(), "#item-b"); len(got) != 1 || got[0].ID() != "item-b" {
		t.Fatalf("Query(#item-b) = %v, want item-b", got)
	}
}

func TestBasicTableSemanticFlow(t *testing.T) {
	ui := New(320, 240)
	err := ui.LoadLayout(`
		<table id="score-table" width="300">
			<tbody id="body">
				<tr id="row-a">
					<th id="head-a">Name</th>
					<th id="head-b">Score</th>
				</tr>
				<tr id="row-b">
					<td id="cell-a">Ada</td>
					<td id="cell-b">10</td>
				</tr>
			</tbody>
		</table>
	`)
	if err != nil {
		t.Fatalf("LoadLayout() error = %v", err)
	}

	table := ui.GetPanel("score-table")
	body := ui.GetPanel("body")
	row := ui.GetPanel("row-b")
	cellA := ui.GetPanel("cell-a")
	cellB := ui.GetPanel("cell-b")
	if table == nil || body == nil || row == nil || cellA == nil || cellB == nil {
		t.Fatal("expected table aliases to parse as panels")
	}
	if table.Style().Direction != LayoutColumn {
		t.Fatalf("table direction = %q, want column", table.Style().Direction)
	}
	if row.Style().Direction != LayoutRow {
		t.Fatalf("row direction = %q, want row", row.Style().Direction)
	}
	if cellA.Style().FlexGrow != 1 || cellB.Style().FlexGrow != 1 {
		t.Fatalf("cell flex grow = %v/%v, want 1/1", cellA.Style().FlexGrow, cellB.Style().FlexGrow)
	}

	a := cellA.ComputedRect()
	b := cellB.ComputedRect()
	if a.W != 150 || b.W != 150 {
		t.Fatalf("cell widths = %v/%v, want 150/150", a.W, b.W)
	}
	if b.X != a.X+a.W {
		t.Fatalf("cell positions = %v then %v, want adjacent row cells", a, b)
	}
}

func TestFormSemantics(t *testing.T) {
	t.Run("submit reset and validation state", func(t *testing.T) {
		ui := New(320, 240)
		err := ui.LoadLayout(`
			<form id="profile" onSubmit="saveProfile" onReset="resetProfile">
				<input id="name" value="Ada" />
				<checkbox id="enabled" checked="true">Enabled</checkbox>
			</form>
		`)
		if err != nil {
			t.Fatalf("LoadLayout() error = %v", err)
		}

		var submitted Widget
		var reset Widget
		ui.RegisterCommand("saveProfile", func(widget Widget) { submitted = widget })
		ui.RegisterCommand("resetProfile", func(widget Widget) { reset = widget })

		ui.SubmitForm("profile")
		if submitted == nil || submitted.ID() != "profile" {
			t.Fatalf("submitted = %v, want profile form", submitted)
		}

		input := ui.GetTextInput("name")
		checkbox := ui.GetCheckbox("enabled")
		if input == nil || checkbox == nil {
			t.Fatal("expected form fields")
		}
		input.SetText("Grace")
		checkbox.Checked = false
		ui.SetValidationState("name", ValidationInvalid)

		ui.ResetForm("profile")
		if reset == nil || reset.ID() != "profile" {
			t.Fatalf("reset = %v, want profile form", reset)
		}
		if input.Text != "Ada" {
			t.Fatalf("input.Text = %q, want Ada", input.Text)
		}
		if !checkbox.Checked {
			t.Fatal("checkbox should reset to checked")
		}
		if got := ui.GetValidationState("name"); got != ValidationNone {
			t.Fatalf("validation state = %q, want none", got)
		}
	})

	t.Run("fieldset disabled propagates to descendants", func(t *testing.T) {
		ui := New(320, 240)
		err := ui.LoadLayout(`
			<form id="profile">
				<fieldset id="locked" disabled="true">
					<input id="name" />
					<button id="save">Save</button>
				</fieldset>
			</form>
		`)
		if err != nil {
			t.Fatalf("LoadLayout() error = %v", err)
		}
		if ui.GetTextInput("name").Enabled() {
			t.Fatal("fieldset input should be disabled")
		}
		if ui.GetButton("save").Enabled() {
			t.Fatal("fieldset button should be disabled")
		}
	})

	t.Run("validation blocks submit and stores messages", func(t *testing.T) {
		ui := New(320, 240)
		err := ui.LoadLayout(`
			<form id="profile" onSubmit="saveProfile">
				<input id="email" required="true" type="email" validation-message="Email required" />
				<textarea id="bio" minlength="4" pattern="^[A-Z].+"></textarea>
			</form>
		`)
		if err != nil {
			t.Fatalf("LoadLayout() error = %v", err)
		}

		submitted := false
		ui.RegisterCommand("saveProfile", func(Widget) { submitted = true })
		ui.SubmitForm("profile")
		if submitted {
			t.Fatal("invalid form should not submit")
		}
		if got := ui.GetValidationState("email"); got != ValidationInvalid {
			t.Fatalf("email validation state = %q, want invalid", got)
		}
		if got := ui.GetValidationMessage("email"); got != "Email required" {
			t.Fatalf("email validation message = %q, want custom message", got)
		}

		ui.GetTextInput("email").SetText("ada@example.com")
		ui.GetTextArea("bio").SetText("Ada")
		if ui.ValidateForm("profile") {
			t.Fatal("short bio should keep form invalid")
		}
		if got := ui.GetValidationState("bio"); got != ValidationInvalid {
			t.Fatalf("bio validation state = %q, want invalid", got)
		}

		ui.GetTextArea("bio").SetText("Ada Lovelace")
		ui.SubmitForm("profile")
		if !submitted {
			t.Fatal("valid form should submit")
		}
		if got := ui.GetValidationState("email"); got != ValidationValid {
			t.Fatalf("email validation state = %q, want valid", got)
		}
		if got := ui.GetValidationMessage("email"); got != "" {
			t.Fatalf("email validation message = %q, want empty", got)
		}
	})

	t.Run("dropdown and slider constraints validate", func(t *testing.T) {
		ui := New(320, 240)
		err := ui.LoadLayout(`
			<form id="settings">
				<select id="mode" required="true">
					<option value="easy">Easy</option>
					<option value="hard">Hard</option>
				</select>
				<slider id="volume" min="10" max="90" value="5" />
			</form>
		`)
		if err != nil {
			t.Fatalf("LoadLayout() error = %v", err)
		}

		if ui.ValidateForm("settings") {
			t.Fatal("empty dropdown and low slider should be invalid")
		}
		if got := ui.GetValidationState("mode"); got != ValidationInvalid {
			t.Fatalf("mode validation state = %q, want invalid", got)
		}
		if got := ui.GetValidationState("volume"); got != ValidationInvalid {
			t.Fatalf("volume validation state = %q, want invalid", got)
		}

		ui.GetWidget("mode").(*Dropdown).SetValue("hard")
		ui.GetSlider("volume").SetValue(40)
		if !ui.ValidateForm("settings") {
			t.Fatal("selected dropdown and in-range slider should be valid")
		}
	})
}

func TestRadioGroupingAndFocusTraversal(t *testing.T) {
	t.Run("radio buttons group by name", func(t *testing.T) {
		ui := New(320, 240)
		err := ui.LoadLayout(`
			<form id="settings">
				<radio id="easy" name="difficulty" value="easy">Easy</radio>
				<radio id="hard" name="difficulty" value="hard">Hard</radio>
			</form>
		`)
		if err != nil {
			t.Fatalf("LoadLayout() error = %v", err)
		}

		easy := ui.GetWidget("easy").(*RadioButton)
		hard := ui.GetWidget("hard").(*RadioButton)
		hard.HandleClick()
		if !hard.Selected || easy.Selected {
			t.Fatalf("radio selection easy=%v hard=%v, want only hard", easy.Selected, hard.Selected)
		}
		easy.HandleClick()
		if !easy.Selected || hard.Selected {
			t.Fatalf("radio selection easy=%v hard=%v, want only easy", easy.Selected, hard.Selected)
		}
	})

	t.Run("focus traversal follows tabindex then source order", func(t *testing.T) {
		ui := New(320, 240)
		err := ui.LoadLayout(`
			<form id="profile">
				<input id="third" tabindex="3" />
				<input id="first" tabindex="1" />
				<button id="second" tabindex="2">Next</button>
			</form>
		`)
		if err != nil {
			t.Fatalf("LoadLayout() error = %v", err)
		}

		if got := ui.FocusNext(false); got == nil || got.ID() != "first" {
			t.Fatalf("first focus = %v, want first", got)
		}
		if got := ui.FocusNext(false); got == nil || got.ID() != "second" {
			t.Fatalf("second focus = %v, want second", got)
		}
		if got := ui.FocusNext(false); got == nil || got.ID() != "third" {
			t.Fatalf("third focus = %v, want third", got)
		}
		if got := ui.FocusNext(true); got == nil || got.ID() != "second" {
			t.Fatalf("reverse focus = %v, want second", got)
		}
	})

	t.Run("dropdown radio and slider keyboard policies", func(t *testing.T) {
		ui := New(320, 240)
		err := ui.LoadLayout(`
			<form id="settings">
				<select id="mode">
					<option value="easy">Easy</option>
					<option value="hard">Hard</option>
				</select>
				<radio id="easy" name="difficulty" value="easy">Easy</radio>
				<radio id="hard" name="difficulty" value="hard">Hard</radio>
				<slider id="volume" min="0" max="10" step="2" value="4" />
			</form>
		`)
		if err != nil {
			t.Fatalf("LoadLayout() error = %v", err)
		}

		ui.Focus("mode")
		ui.SimulateKeyPress(ebiten.KeyEnter, false, false)
		ui.SimulateKeyPress(ebiten.KeyDown, false, false)
		ui.SimulateKeyPress(ebiten.KeyEnter, false, false)
		if got := ui.GetWidget("mode").(*Dropdown).GetSelectedValue(); got != "hard" {
			t.Fatalf("dropdown selected = %q, want hard", got)
		}

		ui.Focus("easy")
		ui.SimulateKeyPress(ebiten.KeyRight, false, false)
		if got := ui.FocusedWidget(); got == nil || got.ID() != "hard" {
			t.Fatalf("focused radio = %v, want hard", got)
		}
		if !ui.GetWidget("hard").(*RadioButton).Selected {
			t.Fatal("right arrow should select hard radio")
		}

		ui.Focus("volume")
		ui.SimulateKeyPress(ebiten.KeyRight, false, false)
		if got := ui.GetSlider("volume").Value; got != 6 {
			t.Fatalf("slider value after right = %v, want 6", got)
		}
		ui.SimulateKeyPress(ebiten.KeyLeft, false, false)
		if got := ui.GetSlider("volume").Value; got != 4 {
			t.Fatalf("slider value after left = %v, want 4", got)
		}
		ui.SimulateKeyPress(ebiten.KeyEnd, false, false)
		if got := ui.GetSlider("volume").Value; got != 10 {
			t.Fatalf("slider value after End = %v, want 10", got)
		}
		ui.SimulateKeyPress(ebiten.KeyHome, false, false)
		if got := ui.GetSlider("volume").Value; got != 0 {
			t.Fatalf("slider value after Home = %v, want 0", got)
		}
	})

	t.Run("space activation and enter form submit", func(t *testing.T) {
		ui := New(320, 240)
		err := ui.LoadLayout(`
			<form id="profile" onSubmit="saveProfile">
				<input id="name" value="Ada" />
				<button id="save" onClick="saveClick">Save</button>
				<checkbox id="music">Music</checkbox>
				<toggle id="sound">Sound</toggle>
				<radio id="easy" name="difficulty" value="easy">Easy</radio>
			</form>
		`)
		if err != nil {
			t.Fatalf("LoadLayout() error = %v", err)
		}

		submitted := false
		clicked := false
		ui.RegisterCommand("saveProfile", func(widget Widget) {
			submitted = widget != nil && widget.ID() == "profile"
		})
		ui.RegisterCommand("saveClick", func(Widget) { clicked = true })

		ui.Focus("save")
		ui.SimulateKeyPress(ebiten.KeySpace, false, false)
		if !clicked {
			t.Fatal("focused button Space should dispatch click")
		}

		ui.Focus("music")
		ui.SimulateKeyPress(ebiten.KeySpace, false, false)
		if !ui.GetCheckbox("music").Checked {
			t.Fatal("focused checkbox Space should toggle checked")
		}

		ui.Focus("sound")
		ui.SimulateKeyPress(ebiten.KeySpace, false, false)
		if !ui.GetWidget("sound").(*Toggle).Checked {
			t.Fatal("focused toggle Space should toggle checked")
		}

		ui.Focus("easy")
		ui.SimulateKeyPress(ebiten.KeySpace, false, false)
		if !ui.GetWidget("easy").(*RadioButton).Selected {
			t.Fatal("focused radio Space should select radio")
		}

		ui.Focus("name")
		ui.SimulateKeyPress(ebiten.KeyEnter, false, false)
		if !submitted {
			t.Fatal("focused text input Enter should submit nearest form")
		}
	})

	t.Run("modal focus traps and restores", func(t *testing.T) {
		ui := New(320, 240)
		err := ui.LoadLayout(`
			<panel id="root">
				<button id="before">Before</button>
				<dialog id="dialog" title="Confirm">
					<button id="ok">OK</button>
					<button id="cancel">Cancel</button>
				</dialog>
				<button id="after">After</button>
			</panel>
		`)
		if err != nil {
			t.Fatalf("LoadLayout() error = %v", err)
		}

		ui.Focus("before")
		ui.GetWidget("dialog").(*Modal).Open()
		if got := ui.FocusNext(false); got == nil || got.ID() != "ok" {
			t.Fatalf("first modal focus = %v, want ok", got)
		}
		if got := ui.FocusNext(false); got == nil || got.ID() != "cancel" {
			t.Fatalf("second modal focus = %v, want cancel", got)
		}
		if got := ui.FocusNext(false); got == nil || got.ID() != "ok" {
			t.Fatalf("wrapped modal focus = %v, want ok", got)
		}

		ui.SimulateKeyPress(ebiten.KeyEscape, false, false)
		if ui.GetWidget("dialog").(*Modal).IsOpen {
			t.Fatal("escape should close modal")
		}
		if got := ui.FocusedWidget(); got == nil || got.ID() != "before" {
			t.Fatalf("restored focus = %v, want before", got)
		}
	})

	t.Run("nested modal focus restores one level at a time", func(t *testing.T) {
		ui := New(320, 240)
		err := ui.LoadLayout(`
			<panel id="root">
				<button id="before">Before</button>
				<dialog id="outer" title="Outer">
					<button id="outer-ok">Outer OK</button>
					<dialog id="inner" title="Inner">
						<button id="inner-ok">Inner OK</button>
					</dialog>
				</dialog>
			</panel>
		`)
		if err != nil {
			t.Fatalf("LoadLayout() error = %v", err)
		}

		outer := ui.GetWidget("outer").(*Modal)
		inner := ui.GetWidget("inner").(*Modal)
		ui.Focus("before")
		outer.Open()
		if got := ui.FocusNext(false); got == nil || got.ID() != "outer-ok" {
			t.Fatalf("outer modal focus = %v, want outer-ok", got)
		}

		inner.Open()
		if got := ui.FocusNext(false); got == nil || got.ID() != "inner-ok" {
			t.Fatalf("inner modal focus = %v, want inner-ok", got)
		}

		ui.SimulateKeyPress(ebiten.KeyEscape, false, false)
		if inner.IsOpen {
			t.Fatal("first escape should close inner modal")
		}
		if !outer.IsOpen {
			t.Fatal("first escape should keep outer modal open")
		}
		if got := ui.FocusedWidget(); got == nil || got.ID() != "outer-ok" {
			t.Fatalf("inner close restored focus = %v, want outer-ok", got)
		}

		ui.SimulateKeyPress(ebiten.KeyEscape, false, false)
		if outer.IsOpen {
			t.Fatal("second escape should close outer modal")
		}
		if got := ui.FocusedWidget(); got == nil || got.ID() != "before" {
			t.Fatalf("outer close restored focus = %v, want before", got)
		}
	})
}
