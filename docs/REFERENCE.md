# Ebiten XML UI Framework - 완전 레퍼런스

> **버전**: 1.0.0  
> **최종 수정**: 2026-02-05  
> **언어**: Go + XML + JSON

Ebiten 게임 엔진을 위한 **데이터 기반 UI 프레임워크**입니다. XML로 레이아웃을 정의하고, JSON으로 CSS와 유사한 스타일링을 적용합니다.

---

## 📁 프로젝트 구조

```
ebitenui-xml/
├── main.go              # 애플리케이션 진입점
├── assets/
│   ├── layout.xml       # UI 레이아웃 정의
│   └── styles.json      # 스타일 시트
└── ui/
    ├── animation.go     # 애니메이션 시스템
    ├── effects.go       # 시각 효과 (그라디언트, 그림자 등)
    ├── layout.go        # Flexbox 레이아웃 엔진
    ├── nineslice.go     # 9-slice 이미지 스케일링
    ├── parser.go        # XML 파서
    ├── style_parser.go  # JSON 스타일 파서
    ├── types.go         # 공통 타입 정의
    ├── ui.go            # UI 매니저
    ├── widget.go        # 기본 위젯 클래스
    └── widgets.go       # 구체적 위젯들 (Button, Text 등)
```

---

## 🏷️ XML 레이아웃 문법

### 기본 구조

```xml
<ui id="root" width="640" height="480">
    <!-- 위젯들을 여기에 배치 -->
</ui>
```

### 지원 위젯 태그

| 태그 | 설명 | 주요 속성 |
|------|------|-----------|
| `<ui>` | 루트 컨테이너 | `id`, `width`, `height` |
| `<panel>` | 컨테이너/박스 | `id`, `class`, `direction` |
| `<button>` | 클릭 가능한 버튼 | `id`, `class` |
| `<text>` | 텍스트 레이블 | `id`, `class` |
| `<progressbar>` | 진행 바 | `id`, `class`, `value` |
| `<image>` | 이미지 (9-slice 지원) | `id`, `class`, `src` |

### 공통 속성

```xml
<!-- ID와 클래스 -->
<panel id="unique-id" class="class1 class2">

<!-- 방향 지정 -->
<panel direction="row">      <!-- 가로 배치 -->
<panel direction="column">   <!-- 세로 배치 (기본값) -->

<!-- 텍스트 내용 -->
<text>표시할 텍스트</text>
<button>버튼 레이블</button>

<!-- 프로그레스바 값 (0.0 ~ 1.0) -->
<progressbar value="0.75"/>
```

### 레이아웃 예시

```xml
<ui id="root" width="640" height="480">
    <!-- 헤더 -->
    <panel id="header">
        <text id="title">앱 제목</text>
    </panel>

    <!-- 메인 영역 (가로 배치) -->
    <panel id="main" direction="row">
        <!-- 사이드바 (세로 배치) -->
        <panel id="sidebar" direction="column">
            <button id="btn-1">버튼 1</button>
            <button id="btn-2">버튼 2</button>
        </panel>

        <!-- 콘텐츠 영역 -->
        <panel id="content">
            <text>메인 콘텐츠</text>
        </panel>
    </panel>

    <!-- 푸터 -->
    <panel id="footer">
        <text id="status">상태 표시</text>
    </panel>
</ui>
```

---

## 🎨 JSON 스타일 문법

### 기본 구조

```json
{
    "styles": {
        "selector": {
            "property": "value"
        }
    }
}
```

### 셀렉터 종류

| 셀렉터 타입 | 예시 | 설명 |
|-------------|------|------|
| ID | `"#header"` | 특정 ID를 가진 요소 |
| 태그 | `"button"` | 해당 타입의 모든 요소 |
| 클래스 | `".menu-btn"` | 해당 클래스를 가진 요소 |

### 우선순위
```
ID (#id) > 클래스 (.class) > 태그 (button)
```

---

## 📐 레이아웃 속성

### 크기 (Sizing)

```json
{
    "width": 200,           // 고정 너비 (px)
    "height": 50,           // 고정 높이 (px)
    "minWidth": 100,        // 최소 너비
    "maxWidth": 400,        // 최대 너비
    "minHeight": 30,        // 최소 높이
    "maxHeight": 200        // 최대 높이
}
```

### Flexbox 레이아웃

```json
{
    "direction": "row",           // "row" | "column"
    "justifyContent": "center",   // 주축 정렬
    "alignItems": "center",       // 교차축 정렬
    "gap": 10,                    // 자식 요소 간격 (px)
    "flexGrow": 1,                // 남은 공간 비율
    "flexShrink": 0               // 축소 비율
}
```

**justifyContent 값:**
- `flex-start` (기본값)
- `flex-end`
- `center`
- `space-between`
- `space-around`
- `space-evenly`

**alignItems 값:**
- `flex-start`
- `center`
- `flex-end`
- `stretch` (기본값)

### 패딩 & 마진

```json
{
    "padding": {
        "top": 10,
        "right": 15,
        "bottom": 10,
        "left": 15
    },
    "margin": {
        "top": 5,
        "right": 10,
        "bottom": 5,
        "left": 10
    }
}
```

**간편 표기법:**
```json
{
    "padding": 10,           // 모든 방향 동일
    "paddingTop": 10,        // 개별 지정
    "marginLeft": 20
}
```

---

## 🎨 시각 스타일 속성

### 배경 (Background)

```json
{
    // 단색 배경
    "background": "#1a1a2e",
    "background": "royalblue",
    "background": "rgba(100,149,237,0.5)",

    // 선형 그라디언트
    "background": "linear-gradient(90deg, #16213e, #1a1a2e, #0f3460)",
    
    // 수직 그라디언트
    "background": "linear-gradient(180deg, #ff6b6b, #4ecdc4)"
}
```

**그라디언트 문법:**
```
linear-gradient(각도, 색상1, 색상2, ...)

각도 예시:
- 0deg   = 아래→위
- 90deg  = 왼쪽→오른쪽
- 180deg = 위→아래
- 270deg = 오른쪽→왼쪽
```

### 테두리 (Border)

```json
{
    "borderWidth": 2,              // 테두리 두께
    "border": "dodgerblue",        // 테두리 색상
    "borderColor": "#ffffff",      // 대체 문법
    "borderRadius": 8              // 둥근 모서리 (px)
}
```

### 박스 그림자 (Box Shadow)

```json
{
    "boxShadow": "offsetX offsetY blur spread color"
}
```

**예시:**
```json
{
    // 기본 그림자
    "boxShadow": "0 4 8 0 rgba(0,0,0,0.3)",
    
    // 더 큰 그림자
    "boxShadow": "0 8 16 4 rgba(100,149,237,0.4)",
    
    // 내부 그림자 (inset)
    "boxShadow": "inset 0 2 4 0 rgba(0,0,0,0.2)"
}
```

### 아웃라인 (Outline)

```json
{
    "outline": "2px solid rgba(100,149,237,0.5)",
    "outlineOffset": 4
}
```

### 텍스트 그림자 (Text Shadow)

```json
{
    "textShadow": "offsetX offsetY blur color"
}
```

**예시:**
```json
{
    "textShadow": "2 2 4 rgba(0,0,0,0.5)"
}
```

### 투명도 (Opacity)

```json
{
    "opacity": 0.8    // 0.0 (투명) ~ 1.0 (불투명)
}
```

---

## ✏️ 텍스트 스타일

```json
{
    "color": "white",              // 텍스트 색상
    "fontSize": 16,                // 폰트 크기
    "fontWeight": "bold",          // "normal" | "bold"
    "textAlign": "center",         // "left" | "center" | "right"
    "lineHeight": 1.5,             // 줄 간격 배율
    
    // 텍스트 오버플로우
    "textOverflow": "ellipsis",    // "clip" | "ellipsis"
    "whiteSpace": "nowrap"         // "normal" | "nowrap"
}
```

---

## 🔄 상태 기반 스타일

### Hover 상태

```json
{
    "button": {
        "background": "royalblue",
        "hover": {
            "background": "dodgerblue",
            "boxShadow": "0 6 12 0 rgba(0,0,0,0.4)"
        }
    }
}
```

### Active 상태 (클릭 중)

```json
{
    "button": {
        "background": "royalblue",
        "active": {
            "background": "darkblue",
            "transform": "scale(0.95)"
        }
    }
}
```

### Disabled 상태

```json
{
    "button": {
        "disabled": {
            "opacity": 0.5,
            "background": "gray"
        }
    }
}
```

### Focus 상태

```json
{
    "input": {
        "focus": {
            "borderColor": "cornflowerblue",
            "outline": "2px solid rgba(100,149,237,0.5)"
        }
    }
}
```

---

## 🎬 애니메이션 시스템

### 내장 애니메이션

| 이름 | 설명 | 효과 |
|------|------|------|
| `fadeIn` | 페이드 인 | 투명→불투명 |
| `fadeOut` | 페이드 아웃 | 불투명→투명 |
| `pulse` | 맥박 | 3회 커졌다 작아짐 |
| `bounce` | 바운스 | 통통 튀는 효과 |
| `shake` | 흔들림 | 좌우로 흔들림 |
| `wobble` | 워블 | 불규칙 흔들림 |
| `slideInLeft` | 슬라이드 인 | 왼쪽에서 등장 |
| `slideInRight` | 슬라이드 인 | 오른쪽에서 등장 |
| `slideInUp` | 슬라이드 인 | 아래에서 등장 |
| `slideInDown` | 슬라이드 인 | 위에서 등장 |
| `zoomIn` | 줌 인 | 작게→크게 |
| `zoomOut` | 줌 아웃 | 크게→작게 |
| `rotateIn` | 회전 등장 | 회전하며 등장 |
| `heartbeat` | 심장박동 | 두 번 펌핑 (반복) |
| `glow` | 빛남 | 그림자 크기 변화 (반복) |

### Go 코드에서 애니메이션 사용

```go
// 이름으로 애니메이션 재생
btn.PlayAnimation("pulse")

// 애니메이션 제어
btn.PauseAnimation()
btn.ResumeAnimation()
btn.StopAnimation()

// 상태 확인
if btn.IsAnimating() {
    // ...
}

// 완료 콜백
btn.OnAnimationComplete(func() {
    fmt.Println("애니메이션 완료!")
})
```

### 커스텀 애니메이션 정의

```go
import "time"

customAnim := &ui.Animation{
    Name:           "myBounce",
    Duration:       500 * time.Millisecond,
    IterationCount: 1,  // -1 = 무한 반복
    Direction:      ui.AnimationNormal,
    TimingFunc:     ui.EaseOutCubic,
    Keyframes: []ui.Keyframe{
        {Percent: 0, Properties: ui.KeyframeProperties{
            TranslateY: 0, ScaleX: 1, ScaleY: 1,
        }},
        {Percent: 50, Properties: ui.KeyframeProperties{
            TranslateY: -20, ScaleX: 1.1, ScaleY: 0.9,
        }},
        {Percent: 100, Properties: ui.KeyframeProperties{
            TranslateY: 0, ScaleX: 1, ScaleY: 1,
        }},
    },
}

// 등록
ui.RegisterAnimation("myBounce", customAnim)

// 사용
widget.PlayAnimation("myBounce")
```

### 키프레임 속성

| 속성 | 타입 | 설명 |
|------|------|------|
| `TranslateX` | float64 | X축 이동 (px) |
| `TranslateY` | float64 | Y축 이동 (px) |
| `ScaleX` | float64 | X축 크기 배율 |
| `ScaleY` | float64 | Y축 크기 배율 |
| `Rotate` | float64 | 회전 각도 (도) |
| `Opacity` | float64 | 투명도 (0~1) |
| `BoxShadowBlur` | float64 | 그림자 블러 |
| `BoxShadowSpread` | float64 | 그림자 확산 |

### 이징 함수

| 함수 | 설명 |
|------|------|
| `EaseLinear` | 일정한 속도 |
| `EaseInQuad` | 가속 (제곱) |
| `EaseOutQuad` | 감속 (제곱) |
| `EaseInOutQuad` | 가속 후 감속 |
| `EaseInCubic` | 강한 가속 |
| `EaseOutCubic` | 강한 감속 |
| `EaseInOutCubic` | 강한 가속 후 감속 |
| `EaseOutElastic` | 탄성 효과 |
| `EaseOutBounce` | 바운스 효과 |

---

## 🔧 Go API 레퍼런스

### UI 매니저

```go
// UI 생성
ui := ui.New(screenWidth, screenHeight)

// 폰트 설정
ui.DefaultFontFace = fontData

// 스타일 로드
err := ui.LoadStyles(stylesJSON)

// 레이아웃 로드
err := ui.LoadLayout(layoutXML)

// 위젯 조회
widget := ui.GetWidget("widget-id")
btn := ui.GetButton("button-id")
txt := ui.GetText("text-id")

// 업데이트 & 렌더링
ui.Update()
ui.Draw(screen)
```

### 위젯 공통 메서드

```go
// 기본 정보
widget.ID() string
widget.Type() string
widget.Classes() []string

// 가시성
widget.SetVisible(true)
widget.Visible() bool

// 활성화
widget.SetEnabled(true)
widget.Enabled() bool

// 스타일
widget.SetStyle(style)
widget.Style() *Style

// 자식 관리
widget.AddChild(child)
widget.Children() []Widget

// 이벤트
widget.OnClick(func() { ... })
widget.OnHover(func() { ... })

// 애니메이션
widget.PlayAnimation("name")
widget.StopAnimation()
widget.IsAnimating() bool
```

### Button 위젯

```go
btn := ui.GetButton("my-button")

// 클릭 이벤트
btn.OnClick(func() {
    fmt.Println("클릭됨!")
})

// 텍스트 변경
btn.Label = "새 레이블"

// 상태 변경
btn.SetEnabled(false)  // 비활성화
```

### Text 위젯

```go
txt := ui.GetText("my-text")

// 내용 변경
txt.Content = "새로운 텍스트"

// 동적 업데이트
txt.Content = fmt.Sprintf("점수: %d", score)
```

### ProgressBar 위젯

```go
bar := ui.GetProgressBar("hp-bar")

// 값 설정 (0.0 ~ 1.0)
bar.Value = 0.75

// 색상 설정
bar.FillColor = color.RGBA{0, 255, 0, 255}      // 채워진 부분
bar.BackgroundColor = color.RGBA{50, 50, 50, 255}  // 배경
```

---

## 💡 완전한 예제

### styles.json

```json
{
    "styles": {
        "#root": {
            "direction": "column",
            "background": "#0a0a18",
            "padding": { "top": 15, "right": 15, "bottom": 15, "left": 15 }
        },
        "#header": {
            "height": 50,
            "background": "linear-gradient(90deg, #16213e, #1a1a2e, #0f3460)",
            "padding": { "top": 10, "right": 20, "bottom": 10, "left": 20 },
            "margin": { "bottom": 10 },
            "borderRadius": 8,
            "boxShadow": "0 4 12 2 rgba(0,0,0,0.4)"
        },
        "#title": {
            "color": "cornflowerblue",
            "fontSize": 18,
            "textShadow": "2 2 4 rgba(0,0,0,0.5)"
        },
        "button": {
            "height": 40,
            "background": "royalblue",
            "borderRadius": 8,
            "borderWidth": 2,
            "border": "dodgerblue",
            "color": "white",
            "boxShadow": "0 4 8 0 rgba(0,0,0,0.3)",
            "hover": {
                "background": "dodgerblue",
                "boxShadow": "0 6 12 0 rgba(0,0,0,0.4)"
            },
            "active": {
                "background": "darkslateblue"
            }
        },
        ".danger": {
            "background": "crimson",
            "border": "darkred",
            "hover": {
                "background": "red"
            }
        }
    }
}
```

### layout.xml

```xml
<ui id="root" width="640" height="480">
    <panel id="header">
        <text id="title">My Game</text>
    </panel>

    <panel id="main" direction="row">
        <panel id="sidebar" direction="column">
            <button id="btn-play">Play</button>
            <button id="btn-quit" class="danger">Quit</button>
        </panel>
        <panel id="content">
            <text>Welcome!</text>
        </panel>
    </panel>

    <panel id="footer">
        <text id="status">Ready</text>
    </panel>
</ui>
```

### main.go

```go
package main

import (
    _ "embed"
    "log"
    
    "github.com/example/ebitenui-xml/ui"
    "github.com/hajimehoshi/ebiten/v2"
)

//go:embed assets/layout.xml
var layoutXML string

//go:embed assets/styles.json
var stylesJSON string

type Game struct {
    ui *ui.UI
}

func NewGame() (*Game, error) {
    g := &Game{}
    g.ui = ui.New(640, 480)
    
    if err := g.ui.LoadStyles(stylesJSON); err != nil {
        return nil, err
    }
    if err := g.ui.LoadLayout(layoutXML); err != nil {
        return nil, err
    }
    
    // 이벤트 핸들러 설정
    if btn := g.ui.GetButton("btn-play"); btn != nil {
        btn.OnClick(func() {
            btn.PlayAnimation("pulse")
            log.Println("Play clicked!")
        })
    }
    
    return g, nil
}

func (g *Game) Update() error {
    g.ui.Update()
    return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
    g.ui.Draw(screen)
}

func (g *Game) Layout(w, h int) (int, int) {
    return 640, 480
}

func main() {
    game, err := NewGame()
    if err != nil {
        log.Fatal(err)
    }
    
    ebiten.SetWindowSize(640, 480)
    ebiten.SetWindowTitle("My Game")
    
    if err := ebiten.RunGame(game); err != nil {
        log.Fatal(err)
    }
}
```

---

## 📝 색상 값 참조

### 지원 형식

```
#RGB        → #f00 (빨강)
#RRGGBB     → #ff0000 (빨강)
#RRGGBBAA   → #ff0000ff (빨강, 불투명)
rgb(r,g,b)  → rgb(255,0,0)
rgba(r,g,b,a) → rgba(255,0,0,0.5)
이름        → red, blue, cornflowerblue 등
```

### 주요 색상 이름

| 이름 | 색상 | Hex |
|------|------|-----|
| `white` | 흰색 | #FFFFFF |
| `black` | 검정 | #000000 |
| `red` | 빨강 | #FF0000 |
| `green` | 초록 | #008000 |
| `blue` | 파랑 | #0000FF |
| `royalblue` | 로열블루 | #4169E1 |
| `cornflowerblue` | 콘플라워블루 | #6495ED |
| `dodgerblue` | 다저블루 | #1E90FF |
| `crimson` | 크림슨 | #DC143C |
| `gold` | 골드 | #FFD700 |
| `transparent` | 투명 | #00000000 |

---

## 🚀 빌드 & 실행

```bash
# 의존성 설치
go mod tidy

# 빌드
go build .

# 실행
./ebitenui-xml
```

---

## 🎨 CSS Variables (Custom Properties)

CSS Variables allow you to define reusable values that can be used throughout your styles.

### Defining Variables

Variables are defined with the `--` prefix:

```json
{
    ":root": {
        "--primary-color": "#4169E1",
        "--secondary-color": "#FFD700",
        "--spacing-unit": "8",
        "--border-radius": "8"
    }
}
```

### Using Variables

Use the `var()` function to reference variables:

```json
{
    ".button": {
        "background": "var(--primary-color)",
        "borderRadius": "var(--border-radius)"
    }
}
```

### Fallback Values

Provide fallback values for undefined variables:

```json
{
    ".text": {
        "color": "var(--text-color, #ffffff)"
    }
}
```

### Go API

```go
// Set a variable
ui.SetVariable("--primary-color", "#4169E1")

// Get a variable
color := ui.GetVariable("--primary-color")

// Access the variables container
vars := ui.Variables()
vars.Set("--theme", "dark")
```

---

## 📏 Relative Units

The framework supports multiple size units beyond pixels.

### Supported Units

| Unit | Description | Example |
|------|-------------|---------|
| `px` | Absolute pixels | `100px` |
| `%` | Percentage of parent | `50%` |
| `vw` | Viewport width | `100vw` |
| `vh` | Viewport height | `100vh` |
| `em` | Relative to font size | `2em` |
| `rem` | Relative to root font size | `1.5rem` |
| `auto` | Automatic sizing | `auto` |

### Usage in XML

```xml
<panel width="80%" height="100vh">
    <text fontSize="1.5rem">Hello World</text>
    <panel width="calc(100% - 40px)" height="auto" />
</panel>
```

### calc() Function

Perform calculations with mixed units:

```json
{
    ".content": {
        "width": "calc(100% - 200px)",
        "height": "calc(100vh - 60px)"
    }
}
```

Supported operators: `+`, `-`, `*`, `/`

---

## 🔗 Data Binding

Reactive data binding connects your UI to application state.

### Basic Binding

```go
// Get the binding context
bindings := ui.Bindings()

// Set a value
bindings.Set("playerName", "Hero")
bindings.Set("health", 100)
bindings.Set("isAlive", true)

// Bind to widgets
ui.BindText("playerName", "name-label")
ui.BindProgress("health", "health-bar")
ui.BindVisible("isAlive", "player-panel")
```

### Two-Way Binding

For interactive widgets like checkboxes and sliders:

```go
// Checkbox binding
checkbox := ui.GetCheckbox("settings-audio")
bindings.BindCheckbox("audioEnabled", checkbox)

// Slider binding
slider := ui.GetSlider("volume-slider")
bindings.BindSlider("volume", slider)
```

### Model Binding

Bind struct fields to the data context:

```go
type Player struct {
    Name   string
    Health int
    Level  int
}

player := &Player{Name: "Hero", Health: 100, Level: 1}
bindings.BindModel(player)

// Changes to player fields update bound widgets
player.Health = 80
bindings.Set("Health", 80) // Triggers update
```

### Computed Properties

Create values that depend on other values:

```go
bindings.AddComputed("healthPercent", []string{"health", "maxHealth"}, func(values ...interface{}) interface{} {
    health := values[0].(int)
    maxHealth := values[1].(int)
    return float64(health) / float64(maxHealth) * 100
})
```

### Formatted Bindings

Create formatted strings from multiple values:

```go
formatted := bindings.FormatBinding("Level: {level} | HP: {health}/{maxHealth}", "level", "health", "maxHealth")
formatted.Subscribe(func(s string) {
    // s = "Level: 5 | HP: 80/100"
})
```

### Observable Values

For fine-grained reactivity:

```go
health := ui.NewObservable(100)
health.Subscribe(func(value int) {
    fmt.Printf("Health changed to: %d\n", value)
})

health.Set(80) // Triggers all subscribers
```

---

## ⌨️ Input Widgets

### TextInput

Single-line text input field:

```xml
<input id="username" placeholder="Enter username" maxlength="20" />
<input id="password" placeholder="Password" password="true" />
```

Go API:
```go
input := ui.GetTextInput("username")
input.Text = "DefaultValue"
input.Placeholder = "Enter text..."
input.MaxLength = 50
input.ReadOnly = false
input.Password = false

input.OnChange = func(text string) {
    fmt.Println("Text changed:", text)
}

input.OnSubmit = func(text string) {
    fmt.Println("Enter pressed:", text)
}
```

### TextArea

Multi-line text input:

```xml
<textarea id="bio" placeholder="Tell us about yourself..." />
```

Go API:
```go
textarea := ui.GetTextArea("bio")
textarea.SetText("Initial text\nWith multiple lines")
```

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Backspace` | Delete character before cursor |
| `Delete` | Delete character after cursor |
| `Left/Right` | Move cursor |
| `Home/End` | Move to start/end |
| `Shift+Arrow` | Select text |
| `Ctrl+A` | Select all |
| `Enter` | Submit (TextInput) / New line (TextArea) |

---

## 📜 Scrollable Containers

Create scrollable areas for content overflow:

```xml
<scrollable id="log-view" height="200" vertical="true" horizontal="false">
    <panel id="content">
        <!-- Many children here -->
    </panel>
</scrollable>
```

### Configuration

```go
scroll := ui.GetScrollable("log-view")

// Enable/disable scrollbars
scroll.ShowVertical = true
scroll.ShowHorizontal = false

// Customize appearance
scroll.ScrollbarWidth = 8
scroll.ScrollbarColor = color.RGBA{100, 100, 100, 200}
scroll.ScrollbarRadius = 4
scroll.AutoHideScrollbar = true

// Manual scroll control
scroll.ScrollTo(0, 100)
scroll.ScrollBy(0, -50)
scroll.ScrollToTop()
scroll.ScrollToBottom()
```

### Properties

| Property | Type | Description |
|----------|------|-------------|
| `ScrollX` | float64 | Current horizontal scroll |
| `ScrollY` | float64 | Current vertical scroll |
| `ContentWidth` | float64 | Total content width |
| `ContentHeight` | float64 | Total content height |
| `ScrollSpeed` | float64 | Mouse wheel speed |
| `AutoHideScrollbar` | bool | Fade scrollbar when idle |

---

## 🔍 Advanced Selectors

### Descendant Selector

Select elements nested within parents:

```json
{
    ".sidebar .button": {
        "background": "#335588"
    },
    "#main-menu .item": {
        "padding": "10"
    }
}
```

### Child Selector

Select direct children only:

```json
{
    ".menu > .item": {
        "margin": "5"
    }
}
```

### Compound Selectors

Combine multiple conditions:

```json
{
    "button.primary": {
        "background": "#4169E1"
    },
    "panel.dark#sidebar": {
        "background": "#1a1a2e"
    }
}
```

### Attribute Selectors

Select by attribute values:

```json
{
    "[data-type=primary]": {
        "background": "#4169E1"
    }
}
```

### Pseudo-Classes

State-based styling:

```json
{
    ".button:hover": {
        "transform": "scale(1.05)"
    },
    ".input:focus": {
        "borderColor": "#4169E1"
    }
}
```

### Specificity

Selectors are applied by specificity (CSS-like):
- ID selectors: 100 points
- Class selectors: 10 points each
- Type selectors: 1 point
- Later rules win for equal specificity

---

## 🔄 Focus Management

```go
// Set focus to a widget
ui.Focus("username-input")

// Remove focus
ui.Blur()

// Get focused widget
focused := ui.FocusedWidget()
```

---

*© 2026 Ebiten XML UI Framework*
