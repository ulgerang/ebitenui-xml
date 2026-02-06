# Ebiten UI - 빠른 참조 치트시트

## 🏷️ XML 태그 - 기본 위젯

```xml
<ui id="root" width="640" height="480">
<panel id="name" class="class1 class2" direction="row|column">
<button id="btn">Label</button>
<text id="txt">Content</text>
<progressbar id="bar" value="0.5"/>
<image id="img" src="path/to/image.png"/>
<svg id="icon" src="path/to/icon.svg" width="32" height="32"/>
```

## 🏷️ XML 태그 - 확장 위젯

```xml
<!-- 입력 위젯 -->
<textinput id="name" placeholder="Enter name..."/>
<textarea id="desc" placeholder="Description..." rows="5"/>
<slider id="vol" min="0" max="100" value="50"/>
<checkbox id="opt" label="Enable Feature"/>

<!-- 토글 & 라디오 -->
<toggle id="sound" label="Sound" checked="true"/>
<radiobutton id="rb1" value="easy" label="Easy"/>
<radiobutton id="rb2" value="hard" label="Hard"/>

<!-- 드롭다운 -->
<dropdown id="res" placeholder="Resolution">
    <option value="720p">1280x720</option>
    <option value="1080p">1920x1080</option>
</dropdown>

<!-- 정보 표시 -->
<badge id="count" text="99+"/>
<spinner id="loading"/>

<!-- 스크롤 -->
<scrollable id="list" direction="vertical">
    <panel>Item 1</panel>
</scrollable>
```

## 🎨 JSON 스타일 셀렉터

```json
"#id-name"     // ID 셀렉터
"button"       // 태그 셀렉터
".class-name"  // 클래스 셀렉터
```

## 📐 레이아웃

```json
{
  "direction": "row | column",
  "gap": 10,
  "justifyContent": "flex-start | center | flex-end | space-between",
  "alignItems": "flex-start | center | flex-end | stretch",
  "flexGrow": 1,
  "width": 200,
  "height": 50
}
```

## 📏 패딩 & 마진

```json
{
  "padding": { "top": 10, "right": 15, "bottom": 10, "left": 15 },
  "margin": { "top": 5, "bottom": 5 }
}
```

## 🎨 배경 스타일

```json
{
  "background": "#1a1a2e",
  "background": "rgb(26,26,46)",
  "background": "rgba(26,26,46,0.8)",
  "background": "linear-gradient(90deg, #color1, #color2)",
  "opacity": 0.9
}
```

## 🔲 테두리 & 그림자

```json
{
  "borderWidth": 2,
  "border": "dodgerblue",
  "borderRadius": 8,
  "boxShadow": "0 4 8 0 rgba(0,0,0,0.3)",
  "outline": "2px solid blue",
  "outlineOffset": 4
}
```

## ✏️ 텍스트

```json
{
  "color": "white",
  "fontSize": 16,
  "textAlign": "left | center | right",
  "textShadow": "2 2 4 rgba(0,0,0,0.5)"
}
```

## 🔄 상태 스타일

```json
{
  "button": {
    "background": "blue",
    "hover": { "background": "lightblue" },
    "active": { "background": "darkblue" },
    "disabled": { "opacity": 0.5 }
  }
}
```

## 🎬 애니메이션 (Go)

```go
// 재생
widget.PlayAnimation("fadeIn")
widget.PlayAnimation("pulse")
widget.PlayAnimation("shake")
widget.PlayAnimation("bounce")
widget.PlayAnimation("wobble")
widget.PlayAnimation("zoomIn")
widget.PlayAnimation("slideInLeft")

// 제어
widget.StopAnimation()
widget.PauseAnimation()
widget.ResumeAnimation()

// 콜백
widget.OnAnimationComplete(func() { ... })
```

## 🔧 Go API

```go
// UI 생성
ui := ui.New(width, height)
ui.LoadStyles(jsonString)
ui.LoadLayout(xmlString)

// 위젯 조회
widget := ui.GetWidget("id")
btn := ui.GetButton("id")
txt := ui.GetText("id")

// 이벤트
btn.OnClick(func() { ... })

// 속성 변경
txt.Content = "New text"
btn.SetEnabled(false)
widget.SetVisible(true)

// 메인 루프
ui.Update()
ui.Draw(screen)
```

## 🌈 색상 표기법

```
#RGB          #f00
#RRGGBB       #ff0000
#RRGGBBAA     #ff0000ff
rgb(r,g,b)    rgb(255,0,0)
rgba(r,g,b,a) rgba(255,0,0,0.5)
이름           red, blue, royalblue
```

## 📦 그라디언트

```
linear-gradient(각도, 색1, 색2, ...)

0deg   = ↑    (아래→위)
90deg  = →    (왼쪽→오른쪽)
180deg = ↓    (위→아래)
270deg = ←    (오른쪽→왼쪽)
```

## ⚡ 이징 함수

```
EaseLinear, EaseInQuad, EaseOutQuad, EaseInOutQuad
EaseInCubic, EaseOutCubic, EaseInOutCubic
EaseOutElastic, EaseOutBounce
```

## 🆕 확장 위젯 Go API

```go
// Toggle
if w := ui.GetWidget("sound"); w != nil {
    if t, ok := w.(*ui.Toggle); ok {
        t.OnChange = func(checked bool) { ... }
    }
}

// RadioGroup (코드에서 생성)
group := ui.NewRadioGroup("difficulty")
group.AddButton(rb1)
group.AddButton(rb2)
group.OnChange = func(value string) { ... }

// Dropdown
if w := ui.GetWidget("res"); w != nil {
    if d, ok := w.(*ui.Dropdown); ok {
        d.OnChange = func(idx int, val string) { ... }
        d.Update()  // 매 프레임 호출 필수
    }
}

// Toast (동적 생성)
toast := ui.NewToast("id", "Message")
toast.ToastType = "success"  // info|success|warning|error
toast.Show()
toast.Update()  // 매 프레임
toast.Draw(screen)  // Draw에서

// Modal (동적 생성)
modal := ui.NewModal("id", "Title")
modal.Content = "Dialog content"
modal.AddButton(confirmBtn)
modal.Open()
modal.Draw(screen)  // Draw에서
```

## 📚 문서 링크

- [REFERENCE.md](./REFERENCE.md) - 전체 레퍼런스
- [WIDGETS_EXTENDED.md](./WIDGETS_EXTENDED.md) - 확장 위젯 가이드
- [SVG_REFERENCE.md](./SVG_REFERENCE.md) - SVG 렌더링 가이드
