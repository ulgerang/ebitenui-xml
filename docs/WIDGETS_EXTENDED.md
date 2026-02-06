# 확장 위젯 레퍼런스 (Extended Widgets Reference)

> **버전**: 1.1.0  
> **최종 수정**: 2026-02-06  
> **언어**: Go + XML + JSON

이 문서는 Ebiten XML UI 프레임워크의 **확장 위젯**들을 설명합니다. 기본 위젯(Panel, Button, Text, ProgressBar, Image)에 추가로 더 풍부한 인터페이스를 구현할 수 있습니다.

---

## 📦 확장 위젯 목록

| 위젯 | XML 태그 | 설명 |
|------|----------|------|
| Toggle | `<toggle>` | 온/오프 스위치 (iOS 스타일) |
| RadioButton | `<radiobutton>` | 라디오 버튼 (단일 선택) |
| Dropdown | `<dropdown>` | 드롭다운 선택 메뉴 |
| Modal | `<modal>` | 팝업 다이얼로그 |
| Badge | `<badge>` | 알림 배지 |
| Spinner | `<spinner>` | 로딩 인디케이터 |
| Toast | `<toast>` | 임시 알림 메시지 |
| Tooltip | `<tooltip>` | 호버 정보 팝업 |

---

## 🔄 Toggle (토글 스위치)

iOS 스타일의 온/오프 스위치입니다. Checkbox의 시각적 대안으로 사용됩니다.

### XML 예시

```xml
<toggle id="sound-toggle" label="Enable Sound"/>
<toggle id="music-toggle" label="Enable Music" checked="true"/>
```

### 속성

| 속성 | 타입 | 설명 | 기본값 |
|------|------|------|--------|
| `id` | string | 고유 식별자 | - |
| `label` | string | 토글 옆에 표시될 텍스트 | - |
| `checked` | boolean | 초기 체크 상태 | `false` |
| `class` | string | 스타일 클래스 | - |

### Go API

```go
// 위젯 가져오기
if w := ui.GetWidget("sound-toggle"); w != nil {
    if toggle, ok := w.(*ui.Toggle); ok {
        // 상태 변경 콜백
        toggle.OnChange = func(checked bool) {
            if checked {
                fmt.Println("Sound enabled")
            } else {
                fmt.Println("Sound disabled")
            }
        }
        
        // 상태 확인/변경
        toggle.Checked = true
        
        // 색상 커스터마이징
        toggle.OnColor = color.RGBA{76, 175, 80, 255}   // 켜졌을 때
        toggle.OffColor = color.RGBA{100, 100, 100, 255} // 꺼졌을 때
        toggle.ThumbColor = color.White                   // 버튼 색상
    }
}
```

### 스타일 예시

```json
{
    "toggle": {
        "height": 32,
        "margin": { "top": 8, "bottom": 8 }
    }
}
```

---

## 🔘 RadioButton (라디오 버튼)

여러 옵션 중 하나만 선택할 수 있는 라디오 버튼입니다. `RadioGroup`과 함께 사용합니다.

### XML 예시

```xml
<radiobutton id="rb-easy" value="easy" label="Easy Mode"/>
<radiobutton id="rb-normal" value="normal" label="Normal Mode" selected="true"/>
<radiobutton id="rb-hard" value="hard" label="Hard Mode"/>
```

### 속성

| 속성 | 타입 | 설명 | 기본값 |
|------|------|------|--------|
| `id` | string | 고유 식별자 | - |
| `value` | string | 선택 시 반환될 값 | - |
| `label` | string | 라벨 텍스트 | - |
| `selected` | boolean | 초기 선택 상태 | `false` |

### Go API

```go
// RadioGroup 생성 (코드에서)
radioGroup := ui.NewRadioGroup("difficulty")
radioGroup.OnChange = func(value string) {
    fmt.Printf("Difficulty: %s\n", value)
}

// XML에서 로드된 라디오 버튼들을 그룹에 연결
radioIDs := []string{"rb-easy", "rb-normal", "rb-hard"}
for _, id := range radioIDs {
    if w := myUI.GetWidget(id); w != nil {
        if rb, ok := w.(*ui.RadioButton); ok {
            radioGroup.AddButton(rb)
        }
    }
}

// 값 설정
radioGroup.SetValue("normal")

// 현재 값 가져오기
currentValue := radioGroup.Value
```

### 스타일 예시

```json
{
    "radiobutton": {
        "height": 28,
        "margin": { "top": 6, "bottom": 6 },
        "color": "#ffffff"
    }
}
```

---

## 📋 Dropdown (드롭다운)

클릭 시 옵션 목록이 펼쳐지는 선택 메뉴입니다.

### XML 예시

```xml
<dropdown id="resolution" placeholder="Select Resolution...">
    <option value="800x600">800 x 600</option>
    <option value="1024x768">1024 x 768</option>
    <option value="1280x720">1280 x 720 (HD)</option>
    <option value="1920x1080">1920 x 1080 (Full HD)</option>
</dropdown>
```

### 속성

| 속성 | 타입 | 설명 | 기본값 |
|------|------|------|--------|
| `id` | string | 고유 식별자 | - |
| `placeholder` | string | 선택 전 표시될 텍스트 | `"Select..."` |
| `class` | string | 스타일 클래스 | - |

### Option 요소

| 속성 | 타입 | 설명 |
|------|------|------|
| `value` | string | 선택 시 반환될 값 |
| (텍스트) | string | 표시될 레이블 |

### Go API

```go
if w := ui.GetWidget("resolution"); w != nil {
    if dropdown, ok := w.(*ui.Dropdown); ok {
        // 변경 콜백
        dropdown.OnChange = func(index int, value string) {
            fmt.Printf("Selected: %s (index %d)\n", value, index)
        }
        
        // 프로그래밍 방식으로 옵션 추가
        dropdown.AddOption("2560x1440", "2K")
        
        // 값으로 선택
        dropdown.SetValue("1920x1080")
        
        // 현재 선택된 값 가져오기
        selected := dropdown.GetSelectedValue()
        
        // 색상 커스터마이징
        dropdown.DropdownBg = color.RGBA{50, 50, 50, 255}
        dropdown.HoverColor = color.RGBA{70, 70, 70, 255}
    }
}
```

### 스타일 예시

```json
{
    "dropdown": {
        "width": 200,
        "height": 40,
        "background": "#2d3436",
        "borderWidth": 1,
        "border": "#636e72",
        "borderRadius": 6,
        "color": "#ffffff"
    }
}
```

### ⚠️ 주의사항

Dropdown의 `Update()` 메서드를 매 프레임 호출해야 호버 효과와 외부 클릭 닫기가 작동합니다:

```go
func (g *Game) Update() error {
    g.ui.Update()
    
    // Dropdown 업데이트
    if w := g.ui.GetWidget("resolution"); w != nil {
        if d, ok := w.(*ui.Dropdown); ok {
            d.Update()
        }
    }
    return nil
}
```

---

## 🔔 Badge (배지)

알림 카운트나 상태를 표시하는 작은 배지입니다.

### XML 예시

```xml
<badge id="badge-new" text="NEW"/>
<badge id="badge-count" text="99+" class="badge-red"/>
<badge id="badge-pro" text="PRO" class="badge-purple"/>
```

### 속성

| 속성 | 타입 | 설명 | 기본값 |
|------|------|------|--------|
| `id` | string | 고유 식별자 | - |
| `text` | string | 배지에 표시될 텍스트 | - |
| `class` | string | 스타일 클래스 | - |

### Go API

```go
badge := ui.NewBadge("notification-badge", "5")

// 텍스트 업데이트
badge.Text = fmt.Sprintf("%d", notificationCount)

// 색상 변경
badge.BadgeColor = color.RGBA{220, 53, 69, 255}  // 빨간색
```

### 스타일 예시

```json
{
    "badge": {
        "borderRadius": 12,
        "color": "#ffffff"
    },
    ".badge-red": {
        "background": "#e74c3c"
    },
    ".badge-purple": {
        "background": "#9b59b6"
    },
    ".badge-green": {
        "background": "#27ae60"
    }
}
```

---

## ⏳ Spinner (로딩 스피너)

로딩 중임을 나타내는 회전 애니메이션입니다.

### XML 예시

```xml
<spinner id="loading-spinner"/>
```

### 속성

| 속성 | 타입 | 설명 | 기본값 |
|------|------|------|--------|
| `id` | string | 고유 식별자 | - |

### Go API

```go
if w := ui.GetWidget("loading-spinner"); w != nil {
    if spinner, ok := w.(*ui.Spinner); ok {
        // 시작/중지
        spinner.IsSpinning = true
        spinner.IsSpinning = false
        
        // 색상 변경
        spinner.SpinnerColor = color.RGBA{100, 149, 237, 255}
        
        // 반드시 매 프레임 Update 호출
        spinner.Update()
    }
}
```

### 스타일 예시

```json
{
    "spinner": {
        "width": 40,
        "height": 40
    }
}
```

### ⚠️ 주의사항

Spinner의 `Update()` 메서드를 매 프레임 호출해야 애니메이션이 작동합니다:

```go
func (g *Game) Update() error {
    g.ui.Update()
    
    if w := g.ui.GetWidget("loading-spinner"); w != nil {
        if s, ok := w.(*ui.Spinner); ok {
            s.Update()
        }
    }
    return nil
}
```

---

## 💬 Toast (토스트 알림)

화면 하단에 잠시 나타났다 사라지는 알림 메시지입니다.

### XML 지원 안함

Toast는 동적으로 생성되어 표시되므로 XML에서 정의하지 않습니다.

### Go API

```go
// 토스트 생성
toast := ui.NewToast("notification", "저장되었습니다!")
toast.FontFace = fontFace
toast.Duration = 3.0  // 3초 후 자동 숨김

// 토스트 타입 (색상이 달라짐)
toast.ToastType = "success"  // success, warning, error, info

// 표시
toast.Show()

// Update에서 호출
func (g *Game) Update() error {
    g.toast.Update()
    return nil
}

// Draw에서 렌더링 (다른 UI 위에)
func (g *Game) Draw(screen *ebiten.Image) {
    g.ui.Draw(screen)
    
    if g.toast.IsVisible {
        g.toast.Draw(screen)
    }
}
```

### 토스트 타입

| 타입 | 색상 | 용도 |
|------|------|------|
| `info` | 파란색 | 일반 정보 |
| `success` | 녹색 | 성공 메시지 |
| `warning` | 주황색 | 경고 메시지 |
| `error` | 빨간색 | 오류 메시지 |

---

## 🪟 Modal (모달 다이얼로그)

화면 중앙에 표시되는 팝업 다이얼로그입니다.

### XML 지원 안함

Modal은 동적으로 생성되어 표시되므로 XML에서 정의하지 않습니다.

### Go API

```go
// 모달 생성
modal := ui.NewModal("confirm-dialog", "확인")
modal.Content = "정말 삭제하시겠습니까?\n이 작업은 취소할 수 없습니다."
modal.FontFace = fontFace

// 버튼 추가
confirmBtn := ui.NewButton("confirm-btn", "확인")
confirmBtn.FontFace = fontFace
confirmBtn.Style().BackgroundColor = color.RGBA{39, 174, 96, 255}
confirmBtn.OnClick(func() {
    modal.Close()
    // 확인 처리
})
modal.AddButton(confirmBtn)

cancelBtn := ui.NewButton("cancel-btn", "취소")
cancelBtn.FontFace = fontFace
cancelBtn.Style().BackgroundColor = color.RGBA{231, 76, 60, 255}
cancelBtn.OnClick(func() {
    modal.Close()
})
modal.AddButton(cancelBtn)

// 닫기 콜백
modal.OnClose = func() {
    fmt.Println("Modal closed")
}

// 표시/숨김
modal.Open()
modal.Close()

// Draw에서 렌더링 (다른 UI 위에)
func (g *Game) Draw(screen *ebiten.Image) {
    g.ui.Draw(screen)
    
    if g.modal.IsOpen {
        g.modal.Draw(screen)
    }
}
```

### 커스터마이징

```go
// 오버레이 색상 변경
modal.OverlayColor = color.RGBA{0, 0, 0, 200}

// 제목 폰트 별도 지정
modal.TitleFontFace = titleFont
```

---

## 💡 Tooltip (툴팁)

위젯 위에 마우스를 올렸을 때 표시되는 정보 팝업입니다.

### XML 지원 안함

Tooltip은 동적으로 생성됩니다.

### Go API

```go
// 툴팁 생성
tooltip := ui.NewTooltip("help-tooltip", "이 버튼을 클릭하면 게임이 시작됩니다.")
tooltip.FontFace = fontFace
tooltip.Position = "top"  // top, bottom, left, right
tooltip.Offset = 8

// 대상 위젯에 연결
if btn := ui.GetButton("play-btn"); btn != nil {
    btn.OnHover(func() {
        tooltip.Show()
    })
    // OnLeave는 별도 구현 필요
}

// Update에서 위치 추적
tooltip.Update()

// Draw에서 렌더링
if tooltip.IsVisible {
    tooltip.Draw(screen)
}
```

---

## 🔧 사용자 입력 위젯

### TextInput (텍스트 입력)

```xml
<textinput id="username" placeholder="Enter username..."/>
```

### TextArea (다중 라인 텍스트)

```xml
<textarea id="description" placeholder="Enter description..." rows="5"/>
```

### Slider (슬라이더)

```xml
<slider id="volume" min="0" max="100" value="50"/>
```

### Checkbox (체크박스)

```xml
<checkbox id="fullscreen" label="Fullscreen Mode"/>
```

### Scrollable (스크롤 컨테이너)

```xml
<scrollable id="item-list" direction="vertical">
    <panel class="item">Item 1</panel>
    <panel class="item">Item 2</panel>
    <!-- ... -->
</scrollable>
```

---

## 📌 통합 사용 예시

### XML 레이아웃

```xml
<?xml version="1.0" encoding="UTF-8"?>
<panel id="root" class="root-panel">
    <text class="title">Game Settings</text>
    
    <panel class="settings-section">
        <text class="section-title">Audio</text>
        <toggle id="sound-toggle" label="Sound Effects"/>
        <toggle id="music-toggle" label="Background Music" checked="true"/>
        <slider id="volume" min="0" max="100" value="80"/>
    </panel>
    
    <panel class="settings-section">
        <text class="section-title">Graphics</text>
        <dropdown id="resolution" placeholder="Resolution">
            <option value="720p">1280 x 720</option>
            <option value="1080p">1920 x 1080</option>
        </dropdown>
        <checkbox id="vsync" label="V-Sync"/>
    </panel>
    
    <panel class="button-row">
        <button id="save-btn" class="btn-success">Save</button>
        <button id="cancel-btn" class="btn-danger">Cancel</button>
    </panel>
</panel>
```

### Go 코드

```go
package main

import (
    _ "embed"
    "log"
    
    "github.com/example/ebitenui-xml/ui"
    "github.com/hajimehoshi/bitmapfont/v4"
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/text/v2"
)

//go:embed assets/settings.xml
var layoutXML string

//go:embed assets/styles.json
var stylesJSON string

type Game struct {
    ui    *ui.UI
    toast *ui.Toast
    modal *ui.Modal
}

func NewGame() (*Game, error) {
    g := &Game{}
    g.ui = ui.New(800, 600)
    g.ui.DefaultFontFace = text.NewGoXFace(bitmapfont.FaceEA)
    
    if err := g.ui.LoadStyles(stylesJSON); err != nil {
        return nil, err
    }
    if err := g.ui.LoadLayout(layoutXML); err != nil {
        return nil, err
    }
    
    g.setupEventHandlers()
    g.createOverlays()
    
    return g, nil
}

func (g *Game) setupEventHandlers() {
    // Toggle 핸들러
    if w := g.ui.GetWidget("sound-toggle"); w != nil {
        if t, ok := w.(*ui.Toggle); ok {
            t.OnChange = func(checked bool) {
                log.Printf("Sound: %v", checked)
            }
        }
    }
    
    // Save 버튼
    if btn := g.ui.GetButton("save-btn"); btn != nil {
        btn.OnClick(func() {
            g.toast.ToastType = "success"
            g.toast.Message = "Settings saved!"
            g.toast.Show()
        })
    }
}

func (g *Game) createOverlays() {
    fontFace := g.ui.DefaultFontFace
    
    g.toast = ui.NewToast("toast", "")
    g.toast.FontFace = fontFace
}

func (g *Game) Update() error {
    g.ui.Update()
    
    // Spinner, Dropdown 등 업데이트
    if w := g.ui.GetWidget("loading-spinner"); w != nil {
        if s, ok := w.(*ui.Spinner); ok {
            s.Update()
        }
    }
    
    g.toast.Update()
    
    return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
    screen.Fill(color.RGBA{26, 26, 46, 255})
    g.ui.Draw(screen)
    
    // 오버레이 렌더링
    if g.toast.IsVisible {
        g.toast.Draw(screen)
    }
}

func (g *Game) Layout(w, h int) (int, int) {
    return 800, 600
}

func main() {
    game, err := NewGame()
    if err != nil {
        log.Fatal(err)
    }
    
    ebiten.SetWindowSize(800, 600)
    ebiten.SetWindowTitle("Settings")
    
    if err := ebiten.RunGame(game); err != nil {
        log.Fatal(err)
    }
}
```

---

## 🎨 스타일 시트 템플릿

```json
{
    "styles": {
        ".root-panel": {
            "width": 800,
            "height": 600,
            "direction": "column",
            "padding": { "top": 20, "right": 20, "bottom": 20, "left": 20 },
            "gap": 20,
            "background": "#1a1a2e"
        },
        
        ".title": {
            "fontSize": 24,
            "color": "#ffffff",
            "textAlign": "center"
        },
        
        ".settings-section": {
            "direction": "column",
            "gap": 12,
            "padding": { "top": 15, "right": 15, "bottom": 15, "left": 15 },
            "background": "#16213e",
            "borderRadius": 8
        },
        
        ".section-title": {
            "color": "#6c5ce7",
            "fontSize": 16
        },
        
        "toggle": {
            "height": 32
        },
        
        "dropdown": {
            "width": 200,
            "height": 40,
            "background": "#2d3436",
            "borderRadius": 6,
            "color": "#ffffff"
        },
        
        ".button-row": {
            "direction": "row",
            "gap": 10,
            "justify": "center"
        },
        
        ".btn-success": {
            "background": "#27ae60",
            "hover": { "background": "#2ecc71" }
        },
        
        ".btn-danger": {
            "background": "#e74c3c",
            "hover": { "background": "#c0392b" }
        }
    }
}
```

---

## 📋 체크리스트

✅ **확장 위젯 추가 완료**
- [x] Toggle (토글 스위치)
- [x] RadioButton (라디오 버튼)
- [x] Dropdown (드롭다운)
- [x] Badge (배지)
- [x] Spinner (로딩 스피너)
- [x] Toast (토스트 알림)
- [x] Modal (모달 다이얼로그)
- [x] Tooltip (툴팁)

✅ **XML 파서 지원**
- [x] 모든 확장 위젯 태그 파싱
- [x] Dropdown의 `<option>` 중첩 요소 지원

✅ **클릭 핸들링**
- [x] UI.Update()에서 모든 확장 위젯 클릭 처리

✅ **폰트 자동 설정**
- [x] UI.setFonts()에서 모든 확장 위젯에 DefaultFontFace 적용

---

## 🔗 관련 문서

- [REFERENCE.md](./REFERENCE.md) - 기본 위젯 및 스타일링 레퍼런스
- [CHEATSHEET.md](./CHEATSHEET.md) - 빠른 참조용 치트시트
