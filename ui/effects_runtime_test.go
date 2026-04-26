package ui

import (
	"image/color"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestCSSFilterBlurRuntime(t *testing.T) {
	src := ebiten.NewImage(24, 24)
	filter := ParseFilter("blur(4px)")
	if filter == nil || filter.Blur != 4 {
		t.Fatalf("ParseFilter blur = %#v, want blur 4", filter)
	}

	filtered := applyCSSFilter(src, filter)
	if filtered == src {
		t.Fatal("applyCSSFilter should return a filtered image for blur")
	}
	globalImagePool.Put(filtered)
}

func TestParseAnimationDeclaration(t *testing.T) {
	anim := ParseAnimationDeclaration("pulse 750ms ease-in-out infinite")
	if anim == nil {
		t.Fatal("expected animation")
	}
	if anim.Name != "pulse" {
		t.Errorf("Name = %q, want pulse", anim.Name)
	}
	if anim.Duration != 750*time.Millisecond {
		t.Errorf("Duration = %v, want 750ms", anim.Duration)
	}
	if anim.IterationCount != -1 {
		t.Errorf("IterationCount = %d, want infinite", anim.IterationCount)
	}
}

func TestBaseWidgetStateTransitionStartsFromCSS(t *testing.T) {
	w := NewPanel("panel")
	w.SetStyle(&Style{
		Opacity:           1,
		OpacitySet:        true,
		Transition:        "opacity 1s linear",
		parsedTransitions: parseTransitions("opacity 1s linear"),
		HoverStyle: &Style{
			Opacity:    0.25,
			OpacitySet: true,
		},
	})

	w.SetState(StateHover)

	if w.transitionEngine == nil || !w.transitionEngine.IsActive() {
		t.Fatal("expected active transition after state change")
	}
	rendered := w.renderStyle(w.getActiveStyle())
	if rendered.Opacity < 0.25 || rendered.Opacity > 1 {
		t.Errorf("transition opacity = %v, want between 0.25 and 1", rendered.Opacity)
	}
}

func TestDeclarativeAnimationStartsOnDraw(t *testing.T) {
	w := NewPanel("panel")
	w.SetComputedRect(Rect{X: 0, Y: 0, W: 20, H: 20})
	w.SetStyle(&Style{
		Animation:       "pulse 100ms 1",
		parsedAnimation: ParseAnimationDeclaration("pulse 100ms 1"),
	})

	screen := ebiten.NewImage(32, 32)
	w.Draw(screen)

	if !w.IsAnimating() {
		t.Fatal("expected declarative animation to start during draw")
	}
}

func TestPhase5MultiShadowAndClipParsing(t *testing.T) {
	boxShadows := ParseBoxShadowList("0 1px 2px rgba(0, 0, 0, 0.4), inset 1px 2px 3px 4px #fff")
	if len(boxShadows) != 2 {
		t.Fatalf("box shadow count = %d, want 2", len(boxShadows))
	}
	if boxShadows[0].OffsetY != 1 || boxShadows[0].Blur != 2 {
		t.Fatalf("first box shadow = %+v, want offsetY=1 blur=2", boxShadows[0])
	}
	if !boxShadows[1].Inset || boxShadows[1].Spread != 4 {
		t.Fatalf("second box shadow = %+v, want inset spread=4", boxShadows[1])
	}

	textShadows := ParseTextShadowList("1px 2px 3px rgba(0, 0, 0, 0.5), -1px 0 #fff")
	if len(textShadows) != 2 {
		t.Fatalf("text shadow count = %d, want 2", len(textShadows))
	}
	if textShadows[0].Blur != 3 {
		t.Fatalf("first text shadow blur = %v, want 3", textShadows[0].Blur)
	}
	if textShadows[1].OffsetX != -1 {
		t.Fatalf("second text shadow offsetX = %v, want -1", textShadows[1].OffsetX)
	}

	if parseCSSClipPath("inset(10px 20px)", Rect{X: 0, Y: 0, W: 100, H: 80}) == nil {
		t.Fatal("expected inset clip path")
	}
	if parseCSSClipPath("circle(40% at 25% 50%)", Rect{X: 0, Y: 0, W: 100, H: 80}) == nil {
		t.Fatal("expected circle clip path")
	}
	if parseCSSClipPath("polygon(0 0, 100% 0, 0 100%)", Rect{W: 100, H: 80}) == nil {
		t.Fatal("expected polygon clip path")
	}
	if parseCSSClipPath("path('M0 0 L100 0 L50 80 Z')", Rect{X: 10, Y: 20, W: 100, H: 80}) == nil {
		t.Fatal("expected path clip path")
	}
	if parseCSSClipPath(`path("M0 0 L100 0 L50 80 Z")`, Rect{W: 100, H: 80}) == nil {
		t.Fatal("expected double-quoted path clip path")
	}
	if parseCSSClipPath("path('')", Rect{W: 100, H: 80}) != nil {
		t.Fatal("empty clip-path path() should be ignored")
	}
}

func TestShadowCornerRadiusClamp(t *testing.T) {
	if got := shadowCornerRadius(12, 4, 20); got != 16 {
		t.Fatalf("shadowCornerRadius spread = %v, want 16", got)
	}
	if got := shadowCornerRadius(3, -8, 20); got != 0 {
		t.Fatalf("shadowCornerRadius negative = %v, want 0", got)
	}
	if got := shadowCornerRadius(30, 8, 24); got != 24 {
		t.Fatalf("shadowCornerRadius clamp = %v, want 24", got)
	}
}

func TestFontFamilyFallbackSelection(t *testing.T) {
	defaultFace := testTextFace()
	registeredFace := testTextFace()
	ui := New(200, 120)
	ui.DefaultFontFace = defaultFace
	ui.RegisterFontFace("Registered UI", registeredFace)

	textWidget := NewText("label", "Hello")
	textWidget.SetStyle(&Style{FontFamily: `Missing, "Registered UI", sans-serif`})
	ui.SetRoot(textWidget)

	if textWidget.FontFace != registeredFace {
		t.Fatal("expected registered font face to win after missing family")
	}

	fallbackWidget := NewText("fallback", "Hello")
	fallbackWidget.SetStyle(&Style{FontFamily: "Missing, sans-serif"})
	ui.SetRoot(fallbackWidget)
	if fallbackWidget.FontFace != defaultFace {
		t.Fatal("expected default font face when no registered family matches")
	}

	families := parseFontFamilyList(`"A Font", B, 'C Font'`)
	if len(families) != 3 || families[0] != "A Font" || families[2] != "C Font" {
		t.Fatalf("parseFontFamilyList = %#v", families)
	}
	if measureLineHeight(registeredFace) <= 0 {
		t.Fatal("expected positive line height for registered face")
	}
}

func TestPhase6JSONKeyframesRegisterAnimation(t *testing.T) {
	engine := NewStyleEngine()
	err := engine.LoadFromString(`{
		"keyframes": {
			"popInCustom": {
				"from": {"opacity": 0, "transform": "translateX(-20px) scale(0.5)"},
				"50%": {"opacity": 0.5, "transform": "rotate(10deg)"},
				"to": {"opacity": 1, "transform": "translate(0px, 4px) scale(1, 1)"}
			}
		},
		"styles": {
			"#target": {"animation": "popInCustom 600ms ease-out 1"}
		}
	}`)
	if err != nil {
		t.Fatalf("LoadFromString keyframes: %v", err)
	}

	anim := GetAnimation("popInCustom")
	if anim == nil {
		t.Fatal("expected registered custom animation")
	}
	if len(anim.Keyframes) != 3 {
		t.Fatalf("keyframe count = %d, want 3", len(anim.Keyframes))
	}
	if anim.Keyframes[0].Percent != 0 || anim.Keyframes[2].Percent != 100 {
		t.Fatalf("keyframes not sorted: %+v", anim.Keyframes)
	}
	if anim.Keyframes[0].Properties.TranslateX != -20 {
		t.Fatalf("from translateX = %v, want -20", anim.Keyframes[0].Properties.TranslateX)
	}
	if anim.Keyframes[1].Properties.Rotate != 10 {
		t.Fatalf("middle rotate = %v, want 10", anim.Keyframes[1].Properties.Rotate)
	}

	style := engine.GetStyle("#target")
	if style == nil || style.parsedAnimation == nil {
		t.Fatal("expected style animation declaration to resolve custom keyframes")
	}
	if style.parsedAnimation.Duration != 600*time.Millisecond {
		t.Fatalf("duration = %v, want 600ms", style.parsedAnimation.Duration)
	}
}

func TestCSSKeyframesRegisterAnimation(t *testing.T) {
	engine := NewStyleEngine()
	err := engine.LoadCSS(`
		@keyframes cssPop {
			from { opacity: 0; transform: translateX(-12px) scale(0.5); width: 20px; }
			50% { opacity: 0.5; transform: rotate(15deg); unsupported: ignored; }
			to { opacity: 1; transform: translate(4px, 8px) scale(1, 1); background-color: #ff0000; }
		}
	`)
	if err != nil {
		t.Fatalf("LoadCSS keyframes: %v", err)
	}

	anim := GetAnimation("cssPop")
	if anim == nil {
		t.Fatal("expected registered CSS animation")
	}
	if len(anim.Keyframes) != 3 {
		t.Fatalf("keyframe count = %d, want 3", len(anim.Keyframes))
	}
	if anim.Keyframes[0].Percent != 0 || anim.Keyframes[2].Percent != 100 {
		t.Fatalf("keyframes not sorted: %+v", anim.Keyframes)
	}
	if anim.Keyframes[0].Properties.TranslateX != -12 {
		t.Fatalf("from translateX = %v, want -12", anim.Keyframes[0].Properties.TranslateX)
	}
	if math.Abs(anim.Keyframes[1].Properties.Rotate-15) > 0.0001 {
		t.Fatalf("middle rotate = %v, want 15", anim.Keyframes[1].Properties.Rotate)
	}
	if anim.Keyframes[2].Properties.TranslateY != 8 {
		t.Fatalf("to translateY = %v, want 8", anim.Keyframes[2].Properties.TranslateY)
	}
}

func TestCSSKeyframesMalformedInput(t *testing.T) {
	engine := NewStyleEngine()
	if err := engine.LoadCSS(`@keyframes broken { from { opacity: 0; }`); err == nil {
		t.Fatal("expected malformed keyframes to return an error")
	}
}

func TestCSSRuleSubsetLoadsStyles(t *testing.T) {
	engine := NewStyleEngine()
	err := engine.LoadCSS(`
		.card, #panel {
			display: flex;
			flex-direction: column;
			justify-content: space-between;
			align-items: center;
			gap: 8px;
			width: 120px;
			padding: 4px 6px;
			background: #112233;
			border: 2px solid #ffffff;
			border-radius: 7px;
			opacity: 0.5;
			transform: scale(1.2);
			unknown-prop: ignored;
		}
	`)
	if err != nil {
		t.Fatalf("LoadCSS rules: %v", err)
	}

	style := engine.GetStyle(".card")
	if style == nil {
		t.Fatal("expected .card style")
	}
	if style.Direction != LayoutColumn {
		t.Fatalf("Direction = %q, want column", style.Direction)
	}
	if style.Justify != JustifyBetween || style.Align != AlignCenter {
		t.Fatalf("justify/align = %q/%q, want space-between/center", style.Justify, style.Align)
	}
	if style.Gap != 8 || !style.GapSet {
		t.Fatalf("Gap = %v set=%v, want 8 true", style.Gap, style.GapSet)
	}
	if style.Width != 120 || !style.WidthSet {
		t.Fatalf("Width = %v set=%v, want 120 true", style.Width, style.WidthSet)
	}
	if style.Padding.Top != 4 || style.Padding.Right != 6 || style.Padding.Bottom != 4 || style.Padding.Left != 6 {
		t.Fatalf("Padding = %+v, want 4 6 4 6", style.Padding)
	}
	if style.BackgroundColor == nil || style.BorderColor == nil {
		t.Fatal("expected parsed background and border colors")
	}
	if style.BorderWidth != 2 || style.BorderRadius != 7 {
		t.Fatalf("border width/radius = %v/%v, want 2/7", style.BorderWidth, style.BorderRadius)
	}
	if style.Opacity != 0.5 || style.Transform != "scale(1.2)" {
		t.Fatalf("opacity/transform = %v/%q, want 0.5/scale(1.2)", style.Opacity, style.Transform)
	}
	if engine.GetStyle("#panel") == nil {
		t.Fatal("expected comma-separated #panel style")
	}
}

func TestLoadFromStringDetectsPureCSSRules(t *testing.T) {
	engine := NewStyleEngine()
	if err := engine.LoadFromString(`.card { width: 120px; gap: 8px; }`); err != nil {
		t.Fatalf("LoadFromString CSS rule: %v", err)
	}
	style := engine.GetStyle(".card")
	if style == nil {
		t.Fatal("expected .card style from CSS string")
	}
	if style.Width != 120 || !style.WidthSet || style.Gap != 8 || !style.GapSet {
		t.Fatalf("style width/gap = %v/%v set=%v/%v, want 120/8 set", style.Width, style.Gap, style.WidthSet, style.GapSet)
	}

	jsonEngine := NewStyleEngine()
	if err := jsonEngine.LoadFromString(`{"#panel": {"width": 88}}`); err != nil {
		t.Fatalf("LoadFromString JSON: %v", err)
	}
	if style := jsonEngine.GetStyle("#panel"); style == nil || style.Width != 88 {
		t.Fatalf("JSON style = %#v, want width 88", style)
	}
}

func TestUILoadingResolvesCSSVariables(t *testing.T) {
	t.Run("css rules", func(t *testing.T) {
		ui := New(320, 240)
		ui.SetVariable("--primary", "#ffffff")
		if err := ui.LoadCSS(`.title { color: var(--primary); background: var(--missing, #112233); }`); err != nil {
			t.Fatalf("LoadCSS with variables: %v", err)
		}
		if err := ui.LoadLayout(`<panel id="root"><text id="title" class="title">Title</text></panel>`); err != nil {
			t.Fatalf("LoadLayout() error = %v", err)
		}
		style := ui.GetText("title").Style()
		if style.TextColor == nil || style.BackgroundColor == nil {
			t.Fatal("expected resolved text and fallback background colors")
		}
	})

	t.Run("json styles", func(t *testing.T) {
		ui := New(320, 240)
		ui.SetVariable("primary", "#ffffff")
		if err := ui.LoadStyles(`{"#title": {"color": "var(--primary)", "background": "var(--missing, #112233)"}}`); err != nil {
			t.Fatalf("LoadStyles with variables: %v", err)
		}
		if err := ui.LoadLayout(`<panel id="root"><text id="title">Title</text></panel>`); err != nil {
			t.Fatalf("LoadLayout() error = %v", err)
		}
		style := ui.GetText("title").Style()
		if style.TextColor == nil || style.BackgroundColor == nil {
			t.Fatal("expected resolved JSON text and fallback background colors")
		}
	})
}

func TestCSSComplexSelectorRulesApplyToWidgets(t *testing.T) {
	ui := New(320, 240)
	if err := ui.LoadCSS(`
		.card .title { color: #ffffff; }
		.toolbar > button { width: 48px; }
		.toolbar button { height: 24px; }
	`); err != nil {
		t.Fatalf("LoadCSS complex selectors: %v", err)
	}
	if err := ui.LoadLayout(`
		<panel id="root">
			<panel id="card" class="card">
				<text id="title" class="title">Title</text>
			</panel>
			<panel id="outside">
				<text id="outside-title" class="title">Title</text>
			</panel>
			<panel id="toolbar" class="toolbar">
				<button id="direct">Direct</button>
				<panel id="nested">
					<button id="nested-button">Nested</button>
				</panel>
			</panel>
		</panel>
	`); err != nil {
		t.Fatalf("LoadLayout() error = %v", err)
	}

	title := ui.GetText("title")
	outside := ui.GetText("outside-title")
	if title.Style().TextColor == nil {
		t.Fatal("expected descendant selector to color nested .title")
	}
	if outside.Style().TextColor != nil {
		t.Fatal("descendant selector should not color .title outside .card")
	}

	direct := ui.GetButton("direct")
	nested := ui.GetButton("nested-button")
	if direct.Style().Width != 48 || !direct.Style().WidthSet {
		t.Fatalf("direct child width = %v set=%v, want 48 true", direct.Style().Width, direct.Style().WidthSet)
	}
	if direct.Style().Height != 24 || nested.Style().Height != 24 {
		t.Fatalf("descendant height = %v/%v, want 24/24", direct.Style().Height, nested.Style().Height)
	}
	if nested.Style().WidthSet {
		t.Fatal("child selector should not set width on nested button")
	}
}

func TestCSSSiblingSelectorRulesApplyToWidgets(t *testing.T) {
	ui := New(320, 240)
	if err := ui.LoadCSS(`
		.label + .value { color: #ffffff; }
		.label ~ .hint { background: #112233; }
	`); err != nil {
		t.Fatalf("LoadCSS sibling selectors: %v", err)
	}
	if err := ui.LoadLayout(`
		<panel id="root">
			<text id="label" class="label">Name</text>
			<text id="value" class="value">Ada</text>
			<text id="hint" class="hint">Required</text>
			<panel id="nested">
				<text id="nested-value" class="value">Nested</text>
			</panel>
		</panel>
	`); err != nil {
		t.Fatalf("LoadLayout() error = %v", err)
	}

	value := ui.GetText("value")
	hint := ui.GetText("hint")
	nested := ui.GetText("nested-value")
	if value.Style().TextColor == nil {
		t.Fatal("expected adjacent sibling selector to color .value")
	}
	if hint.Style().BackgroundColor == nil {
		t.Fatal("expected general sibling selector to color .hint background")
	}
	if nested.Style().TextColor != nil {
		t.Fatal("sibling selector should not cross into nested descendants")
	}
}

func TestCSSSourceOrderWinsForSameSpecificity(t *testing.T) {
	ui := New(320, 240)
	if err := ui.LoadCSS(`
		.primary { color: #111111; }
		.accent { color: #ffffff; }
		#title { color: #222222; }
	`); err != nil {
		t.Fatalf("LoadCSS source order: %v", err)
	}
	if err := ui.LoadLayout(`<panel id="root"><text id="title" class="accent primary">Title</text></panel>`); err != nil {
		t.Fatalf("LoadLayout() error = %v", err)
	}

	title := ui.GetText("title")
	if got := color.RGBAModel.Convert(title.Style().TextColor).(color.RGBA); got.R != 0x22 || got.G != 0x22 || got.B != 0x22 {
		t.Fatalf("ID color = %#v, want #222222", got)
	}

	uiNoID := New(320, 240)
	if err := uiNoID.LoadCSS(`
		.primary { color: #111111; }
		.accent { color: #ffffff; }
	`); err != nil {
		t.Fatalf("LoadCSS source order without ID: %v", err)
	}
	if err := uiNoID.LoadLayout(`<panel id="root"><text id="title" class="accent primary">Title</text></panel>`); err != nil {
		t.Fatalf("LoadLayout() error = %v", err)
	}
	if got := color.RGBAModel.Convert(uiNoID.GetText("title").Style().TextColor).(color.RGBA); got.R != 0xff || got.G != 0xff || got.B != 0xff {
		t.Fatalf("same-specificity color = %#v, want later .accent #ffffff", got)
	}
}

func TestUILoadCSSAppliesBasicMediaQueries(t *testing.T) {
	wide := New(320, 240)
	if err := wide.LoadCSS(`
		.card { height: 20px; }
		@media (min-width: 300px) {
			.card { width: 120px; }
		}
		@media (max-width: 200px) {
			.card { gap: 9px; }
		}
	`); err != nil {
		t.Fatalf("wide LoadCSS media: %v", err)
	}
	if err := wide.LoadLayout(`<panel id="root"><panel id="card" class="card"></panel></panel>`); err != nil {
		t.Fatalf("wide LoadLayout() error = %v", err)
	}
	wideStyle := wide.GetPanel("card").Style()
	if wideStyle.Width != 120 || !wideStyle.WidthSet {
		t.Fatalf("wide media width = %v set=%v, want 120 true", wideStyle.Width, wideStyle.WidthSet)
	}
	if wideStyle.GapSet {
		t.Fatal("non-matching max-width media rule should not apply")
	}

	narrow := New(180, 240)
	if err := narrow.LoadCSS(`
		@media (min-width: 300px) {
			.card { width: 120px; }
		}
		@media (max-width: 200px) and (min-height: 200px) {
			.card { gap: 9px; }
		}
	`); err != nil {
		t.Fatalf("narrow LoadCSS media: %v", err)
	}
	if err := narrow.LoadLayout(`<panel id="root"><panel id="card" class="card"></panel></panel>`); err != nil {
		t.Fatalf("narrow LoadLayout() error = %v", err)
	}
	narrowStyle := narrow.GetPanel("card").Style()
	if narrowStyle.WidthSet {
		t.Fatal("non-matching min-width media rule should not apply")
	}
	if narrowStyle.Gap != 9 || !narrowStyle.GapSet {
		t.Fatalf("narrow media gap = %v set=%v, want 9 true", narrowStyle.Gap, narrowStyle.GapSet)
	}
}

func TestUILoadCSSAppliesExpandedMediaQueries(t *testing.T) {
	landscape := New(360, 240)
	if !landscape.matchesCSSMediaCondition("screen and (orientation: landscape), print") {
		t.Fatal("expected landscape media condition to match")
	}
	css := `
		@media screen and (orientation: landscape), print {
			.card { gap: 12px; }
		}
		@media all and (min-height: 200px) and (max-width: 400px) {
			.size { width: 144px; }
		}
		@media speech, print {
			.bad { height: 88px; }
		}
	`
	if err := landscape.LoadCSS(css); err != nil {
		t.Fatalf("landscape LoadCSS media: %v", err)
	}
	if err := landscape.LoadLayout(`<panel id="root"><panel id="card" class="card"></panel><panel id="size" class="size"></panel><panel id="bad" class="bad"></panel></panel>`); err != nil {
		t.Fatalf("landscape LoadLayout() error = %v", err)
	}
	landscapeStyle := landscape.GetPanel("card").Style()
	if landscapeStyle.Gap != 12 || !landscapeStyle.GapSet {
		t.Fatalf("landscape media gap = %v set=%v, want 12 true", landscapeStyle.Gap, landscapeStyle.GapSet)
	}
	sizeStyle := landscape.GetPanel("size").Style()
	if sizeStyle.Width != 144 || !sizeStyle.WidthSet {
		t.Fatalf("all media width = %v set=%v, want 144 true", sizeStyle.Width, sizeStyle.WidthSet)
	}
	if landscape.GetPanel("bad").Style().HeightSet {
		t.Fatal("unsupported media types should not apply")
	}

	portrait := New(240, 360)
	if err := portrait.LoadCSS(`
		@media screen and (orientation: landscape), print {
			.card { gap: 12px; }
		}
		@media (orientation: portrait) {
			.card { height: 64px; }
		}
	`); err != nil {
		t.Fatalf("portrait LoadCSS media: %v", err)
	}
	if err := portrait.LoadLayout(`<panel id="root"><panel id="card" class="card"></panel></panel>`); err != nil {
		t.Fatalf("portrait LoadLayout() error = %v", err)
	}
	portraitStyle := portrait.GetPanel("card").Style()
	if portraitStyle.GapSet {
		t.Fatal("non-matching landscape orientation should not apply")
	}
	if portraitStyle.Height != 64 || !portraitStyle.HeightSet {
		t.Fatalf("portrait media height = %v set=%v, want 64 true", portraitStyle.Height, portraitStyle.HeightSet)
	}
}

func TestCSSCascadeParserEdgeCases(t *testing.T) {
	ui := New(320, 240)
	if err := ui.LoadCSS(`
		.card { width: 80px !important; color: #111111; }
		#card { width: 120px; color: #222222; }
		.card:hover { color: #ffffff; }
		#clip { clip-path: path('M0 0 L10 0; L10 10 Z'); }
	`); err != nil {
		t.Fatalf("LoadCSS edge cases: %v", err)
	}
	if err := ui.LoadLayout(`<panel id="root"><text id="card" class="card">Card</text><panel id="clip"></panel></panel>`); err != nil {
		t.Fatalf("LoadLayout edge cases: %v", err)
	}

	card := ui.GetText("card")
	style := card.Style()
	if style.Width != 80 || !style.WidthSet {
		t.Fatalf("important width = %v set=%v, want 80 true", style.Width, style.WidthSet)
	}
	if got := color.RGBAModel.Convert(style.TextColor).(color.RGBA); got.R != 0x22 {
		t.Fatalf("normal ID color = %#v, want #222222", got)
	}
	if style.HoverStyle == nil {
		t.Fatal("expected hover style from pseudo selector")
	}
	if got := color.RGBAModel.Convert(style.HoverStyle.TextColor).(color.RGBA); got.R != 0xff {
		t.Fatalf("hover color = %#v, want #ffffff", got)
	}

	clip := ui.GetPanel("clip").Style()
	if !strings.Contains(clip.ClipPath, "; L10 10") {
		t.Fatalf("clip path value = %q, want semicolon preserved inside quoted path", clip.ClipPath)
	}
}

func TestCSSRulesAndKeyframesLoadTogether(t *testing.T) {
	engine := NewStyleEngine()
	err := engine.LoadCSS(`
		@keyframes cssRulePop {
			from { opacity: 0; }
			to { opacity: 1; }
		}
		.badge { animation: cssRulePop 200ms 1; color: white; }
	`)
	if err != nil {
		t.Fatalf("LoadCSS mixed stylesheet: %v", err)
	}
	if GetAnimation("cssRulePop") == nil {
		t.Fatal("expected keyframes animation")
	}
	style := engine.GetStyle(".badge")
	if style == nil || style.parsedAnimation == nil {
		t.Fatal("expected CSS rule style with parsed animation")
	}
}
