# ebitenui-xml Project Reference

> **Purpose**: 이 문서를 읽으면 소스코드를 다시 탐색하지 않고도 바로 CSS 비교/수정 작업에 착수할 수 있다.

---

## 1. 프로젝트 개요

XML로 구조(layout)를, JSON으로 스타일(CSS-like)을 정의하여 Ebitengine 위에 UI를 렌더링하는 프레임워크.

```
ebitenui-xml/
├── main.go                          # 메인 데모 앱 (640x480)
├── go.mod                           # ebiten-ertp를 replace로 참조
├── assets/
│   ├── layout.xml                   # 기본 데모 레이아웃 (header/sidebar/content/footer)
│   ├── styles.json                  # 기본 데모 스타일
│   ├── layout_extended.xml          # 확장 위젯 데모
│   ├── styles_extended.json         # 확장 위젯 스타일
│   ├── demo_extended.xml            # 또 다른 확장 데모
│   └── demo_extended.json           # 그에 대한 스타일
├── ui/                              # 핵심 UI 엔진 (모든 렌더링 코드)
│   ├── types.go                     # Style, Widget 인터페이스, 상수 정의
│   ├── widget.go                    # BaseWidget 구현 (Draw 메서드 포함)
│   ├── widgets.go                   # Panel, Button, Text, ProgressBar, Slider, Checkbox, SVGIcon
│   ├── widgets_extended.go          # Toggle, RadioButton, Dropdown, Modal, Toast, Badge, etc.
│   ├── style.go                     # StyleEngine, JSON 파싱, 색상 파싱
│   ├── layout.go                    # LayoutEngine (flexbox-like)
│   ├── effects.go                   # BoxShadow, Gradient, Outline, Transition, 렌더링 함수
│   ├── parser.go                    # XML 파싱 + WidgetFactory
│   ├── ui.go                        # UI 매니저 (LoadLayout, LoadStyles, Update, Draw)
│   ├── selector.go                  # CSS 셀렉터 매칭
│   ├── textwrap.go                  # 텍스트 줄바꿈
│   ├── scrollable.go                # 스크롤 컨테이너
│   ├── input.go                     # TextInput, TextArea 위젯
│   ├── animation.go                 # 키프레임 애니메이션
│   ├── binding.go                   # 데이터 바인딩
│   ├── variables.go                 # CSS 변수
│   ├── nineslice.go                 # 9-slice 이미지
│   ├── svg.go / svg_path.go         # SVG 파싱/렌더링
│   └── effects.go                   # 시각 효과 (shadow, gradient, outline 등)
├── cmd/
│   ├── css_compare/main.go          # ERTP 디버그 서버 연동 테스트 하네스
│   ├── game1984/                    # 게임 데모
│   └── demo_*/                      # 기타 데모들
└── tools/
    └── css_compare/cmd/
        ├── converter/main.go        # XML+JSON → HTML+CSS 변환기
        └── pixeldiff/main.go        # 두 PNG 이미지 픽셀 비교
```

---

## 2. 렌더링 파이프라인 (핵심 흐름)

```
[XML Layout] + [JSON Styles]
         │
         ▼
┌─────────────────────────────────────┐
│  ui.LoadStyles(json)                │  StyleEngine에 셀렉터별 Style 등록
│  ui.LoadLayout(xml)                 │  XML 파싱 → Widget 트리 생성
│    ├── parser.go: CreateFromXML()   │    XMLNode → Widget 변환
│    ├── applyInlineStyles()          │    XML 속성을 Style에 직접 적용
│    ├── reapplyStyles()              │    타입→클래스→ID 순서로 스타일 머지
│    └── setFonts()                   │    텍스트 위젯에 폰트 설정
└─────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────┐
│  LayoutEngine.Layout(root, w, h)    │  flexbox-like 레이아웃 계산
│    ├── 각 위젯의 ComputedRect 결정   │
│    ├── Direction: row / column      │
│    ├── FlexGrow 기반 공간 분배       │
│    ├── Justify: start/center/end    │
│    └── Align: start/center/end     │
└─────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────┐
│  Widget.Draw(screen)                │  실제 렌더링 (widget.go:265~)
│    1. BoxShadow 그리기              │  effects.go: DrawBoxShadow()
│    2. 배경 그리기                    │  9-slice > gradient > solid
│       - 솔리드: DrawRoundedRectPath │
│       - 그라데이션: DrawGradient    │
│    3. 테두리 그리기                  │  drawRoundedRectStroke()
│    4. 아웃라인 그리기               │  DrawOutline()
│    5. 자식 위젯 재귀 Draw           │
│    6. (위젯별) 텍스트, 아이콘 등    │
└─────────────────────────────────────┘
```

---

## 3. 스타일 시스템 상세 (types.go:120~236)

### 3.1 Style 구조체 주요 필드

| 카테고리 | 필드 | JSON 키 | 상태 |
|---------|------|---------|------|
| **레이아웃** | Direction | `direction` | ✅ row/column |
| | Align | `align` | ✅ start/center/end |
| | Justify | `justify` | ⚠️ start/center/end만 구현 (space-between 미구현) |
| | Gap | `gap` | ✅ |
| | FlexWrap | `flexWrap` | ❌ 파싱만, 레이아웃 미반영 |
| **크기** | Width/Height | `width`/`height` | ✅ |
| | MinWidth/MaxWidth | `minWidth`/`maxWidth` | ❌ 파싱만, 레이아웃 미반영 |
| | FlexGrow | `flexGrow` | ✅ |
| | FlexShrink | `flexShrink` | ❌ 파싱만, 레이아웃 미반영 |
| **간격** | Padding | `padding` | ✅ {top,right,bottom,left} |
| | Margin | `margin` | ✅ {top,right,bottom,left} |
| **색상** | Background | `background` | ✅ 솔리드 + linear-gradient |
| | Border (색상) | `border` | ✅ |
| | Color (텍스트) | `color` | ✅ |
| **테두리** | BorderWidth | `borderWidth` | ✅ |
| | BorderRadius | `borderRadius` | ✅ (단일값만, 개별 코너 미지원) |
| | 개별 코너 Radius | `borderTopLeftRadius` 등 | ❌ 파싱만 |
| **텍스트** | FontSize | `fontSize` | ✅ |
| | TextAlign | `textAlign` | ✅ left/center/right |
| | LineHeight | `lineHeight` | ✅ |
| | TextWrap | `textWrap` | ✅ normal/nowrap |
| | TextOverflow | `textOverflow` | ✅ ellipsis |
| **효과** | Opacity | `opacity` | ✅ |
| | BoxShadow | `boxShadow` | ✅ |
| | TextShadow | `textShadow` | ⚠️ 파싱 OK, 렌더링 미구현 |
| | Outline | `outline` | ✅ |
| | Transition | `transition` | ⚠️ 파싱 OK, 보간 미구현 |
| **위치** | Position | `position` | ❌ 파싱만 |
| | Overflow | `overflow` | ❌ 파싱만, 클리핑 미구현 |
| **상태** | HoverStyle | `hover` | ✅ |
| | ActiveStyle | `active` | ✅ |
| | DisabledStyle | `disabled` | ✅ |
| | FocusStyle | `focus` | ✅ |

### 3.2 스타일 적용 우선순위 (ui.go:497~)

```
1. 타입 셀렉터:  widget.Type()  →  "panel", "button", "text" 등
2. 클래스 셀렉터: ".className"
3. ID 셀렉터:    "#widgetId"
```

각 단계에서 `Style.Merge()`로 비-제로 값만 오버라이드. PaddingSet/MarginSet/BorderWidthSet 플래그로 명시적 zero 구분.

### 3.3 색상 파싱 (style.go)

- Hex: `#RGB`, `#RGBA`, `#RRGGBB`, `#RRGGBBAA`
- 함수형: `rgb(r,g,b)`, `rgba(r,g,b,a)`, `hsl(h,s%,l%)`, `hsla(h,s%,l%,a)`
- Named: 전체 CSS named colors 지원 (약 140개)
- Gradient: `linear-gradient(90deg, #color1, #color2)`

---

## 4. 레이아웃 엔진 상세 (layout.go)

### 현재 구현

```go
func (le *LayoutEngine) layoutChildren(parent Widget) {
    // 1. 자식들의 고정 크기와 flexGrow 합산
    // 2. gap 계산 (children-1 개)
    // 3. flexSpace = 가용공간 - 고정크기 - gaps
    // 4. justify에 따라 시작 offset 결정 (center/end)
    // 5. 각 자식 배치:
    //    - 고정 width/height가 있으면 사용
    //    - flexGrow > 0이면 비례 분배
    //    - 없으면 기본값 (row: 50px, column: 30px)
    //    - align에 따라 교차축 정렬
}
```

### 미구현 사항
- `justify: space-between / space-around / space-evenly`
- `flexWrap`
- `flexShrink`
- `minWidth/maxWidth/minHeight/maxHeight` 제약
- `position: absolute` (모든 위젯이 flow 기반)
- `overflow: hidden` (클리핑 없음)

---

## 5. 위젯 타입 매핑

| XML 태그 | Go 타입 | Widget.Type() | 특이사항 |
|---------|---------|---------------|---------|
| `<ui>` | Panel | "panel" | 루트 전용 |
| `<panel>`, `<div>`, `<container>` | Panel | "panel" | |
| `<button>`, `<btn>` | Button | "button" | Label 텍스트 중앙정렬 |
| `<text>`, `<label>`, `<span>`, `<p>` | Text | "text" | 줄바꿈 지원 |
| `<progressbar>`, `<progress>` | ProgressBar | "progressbar" | FillColor 별도 설정 |
| `<input>`, `<textinput>` | TextInput | "textinput" | placeholder, password |
| `<textarea>` | TextArea | "textarea" | 멀티라인 |
| `<scrollable>`, `<scroll>` | Scrollable | "scrollable" | |
| `<checkbox>`, `<check>` | Checkbox | "checkbox" | |
| `<slider>`, `<range>` | Slider | "slider" | |
| `<toggle>`, `<switch>` | Toggle | "toggle" | |
| `<radiobutton>`, `<radio>` | RadioButton | "radiobutton" | |
| `<dropdown>`, `<select>` | Dropdown | "dropdown" | `<option>` 자식 |
| `<modal>`, `<dialog>` | Modal | "modal" | |
| `<tooltip>` | Tooltip | "tooltip" | |
| `<badge>` | Badge | "badge" | |
| `<spinner>`, `<loading>` | Spinner | "spinner" | |
| `<toast>`, `<notification>` | Toast | "toast" | |
| `<svg>`, `<icon>` | SVGIcon | "svg" | built-in 아이콘 |

---

## 6. 기존 도구 (tools/)

### 6.1 HTML 변환기 (`tools/css_compare/cmd/converter/`)

```bash
go run ./tools/css_compare/cmd/converter/ \
  -layout assets/layout.xml \
  -styles assets/styles.json \
  -out reference.html \
  -width 640 -height 480
```

**하는 일**: XML 레이아웃과 JSON 스타일을 HTML+CSS 페이지로 변환. 브라우저에서 열어 "정답" 스크린샷을 만드는 용도.

**변환 로직**:
- 모든 ebitenui-xml 요소 → `display: flex` 기반 div
- 타입 셀렉터 `button` → CSS 클래스 `.eui-button`
- ID/클래스 셀렉터는 그대로 유지
- `box-sizing: border-box` 자동 적용

### 6.2 픽셀 비교기 (`tools/css_compare/cmd/pixeldiff/`)

```bash
go run ./tools/css_compare/cmd/pixeldiff/ \
  img1.png img2.png diff_output.png
```

**출력**: 차이 통계
```
DIFF_PIXELS=1234
TOTAL_PIXELS=307200
DIFF_PCT=0.40
AVG_DELTA=2.15
```

- threshold > 10인 픽셀을 마젠타로 표시
- 일치 영역은 원본을 어둡게 표시

### 6.3 CSS Compare 하네스 (`cmd/css_compare/`)

```bash
go run ./cmd/css_compare/ \
  -layout assets/layout.xml \
  -styles assets/styles.json \
  -port :9222 \
  -width 640 -height 480
```

**하는 일**: ebitenui-xml 데모를 ERTP 디버그 서버와 함께 실행. 스크린샷을 HTTP로 캡처 가능.

---

## 7. 빌드 & 실행

```bash
# 메인 데모
go run .

# CSS 비교 하네스 (ERTP 포함)
go run ./cmd/css_compare/

# HTML 레퍼런스 생성
go run ./tools/css_compare/cmd/converter/ -layout assets/layout.xml -styles assets/styles.json

# 픽셀 비교
go run ./tools/css_compare/cmd/pixeldiff/ expected.png actual.png diff.png
```

### 의존성

```
go.mod: replace github.com/ulgerang/ebiten-ertp => ../ebiten-ertp
```
`ebiten-ertp`는 상위 폴더에 로컬로 존재해야 함.

---

## 8. 주요 수정 대상 파일 (CSS 구현 개선 시)

| 작업 | 파일 | 함수/위치 |
|------|------|-----------|
| 레이아웃 로직 변경 | `ui/layout.go` | `layoutChildren()` |
| 새 CSS 속성 렌더링 | `ui/effects.go` | 새 Draw 함수 추가 |
| 스타일 파싱 추가 | `ui/style.go` | `parseStyleColors()` 등 |
| Style 구조체 확장 | `ui/types.go` | `Style` struct |
| Style 머지 로직 | `ui/types.go` | `Style.Merge()` |
| 위젯 Draw 변경 | `ui/widget.go` | `BaseWidget.Draw()` |
| 특정 위젯 렌더링 | `ui/widgets.go` | 해당 위젯의 `Draw()` |
| HTML 변환기 동기화 | `tools/.../converter/main.go` | `styleToCSS()` |
