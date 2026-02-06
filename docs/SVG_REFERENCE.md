# SVG 렌더링 시스템 레퍼런스

> **버전**: 1.1.0  
> **최종 수정**: 2026-02-06

이 문서는 Ebiten XML UI 프레임워크의 **SVG 렌더링 시스템**을 설명합니다. 벡터 그래픽을 게임 UI에 통합할 수 있습니다.

---

## 📦 지원 SVG 요소

### 기본 도형

| 요소 | 설명 | 예시 |
|------|------|------|
| `<rect>` | 사각형 | `<rect x="10" y="10" width="100" height="50"/>` |
| `<circle>` | 원 | `<circle cx="50" cy="50" r="25"/>` |
| `<ellipse>` | 타원 | `<ellipse cx="50" cy="50" rx="40" ry="20"/>` |
| `<line>` | 선 | `<line x1="0" y1="0" x2="100" y2="100"/>` |
| `<polyline>` | 다중선 | `<polyline points="0,0 50,25 100,0"/>` |
| `<polygon>` | 다각형 | `<polygon points="50,0 100,100 0,100"/>` |
| `<path>` | 경로 | `<path d="M 10 10 L 90 90"/>` |

### 텍스트

```xml
<text x="50" y="50" font-size="16" text-anchor="middle">Hello World</text>
```

### 그룹 및 변환

```xml
<g transform="translate(100, 50) rotate(45)">
    <rect x="0" y="0" width="50" height="50"/>
</g>
```

### 정의 및 재사용

```xml
<defs>
    <linearGradient id="grad1" x1="0%" y1="0%" x2="100%" y2="0%">
        <stop offset="0%" style="stop-color:rgb(255,255,0);stop-opacity:1"/>
        <stop offset="100%" style="stop-color:rgb(255,0,0);stop-opacity:1"/>
    </linearGradient>
</defs>
<rect fill="url(#grad1)" width="100" height="50"/>
```

---

## 🎨 지원 속성

### 채우기 (Fill)

| 속성 | 값 예시 | 설명 |
|------|---------|------|
| `fill` | `#ff0000`, `rgb(255,0,0)`, `red` | 채우기 색상 |
| `fill-opacity` | `0.5` | 채우기 투명도 (0~1) |
| `fill-rule` | `nonzero`, `evenodd` | 채우기 규칙 |

### 선 (Stroke)

| 속성 | 값 예시 | 설명 |
|------|---------|------|
| `stroke` | `#000000`, `black` | 선 색상 |
| `stroke-width` | `2` | 선 두께 |
| `stroke-opacity` | `0.8` | 선 투명도 |
| `stroke-linecap` | `butt`, `round`, `square` | 선 끝 모양 |
| `stroke-linejoin` | `miter`, `round`, `bevel` | 선 연결 모양 |
| `stroke-dasharray` | `5,3` | 점선 패턴 |

### 변환 (Transform)

| 함수 | 예시 | 설명 |
|------|------|------|
| `translate(x, y)` | `translate(50, 100)` | 이동 |
| `rotate(angle)` | `rotate(45)` | 회전 (도) |
| `scale(x, y)` | `scale(2, 1.5)` | 크기 조절 |
| `skewX(angle)` | `skewX(30)` | X축 기울이기 |
| `skewY(angle)` | `skewY(30)` | Y축 기울이기 |
| `matrix(a,b,c,d,e,f)` | `matrix(1,0,0,1,0,0)` | 변환 행렬 |

---

## 📐 Path 명령어

SVG Path는 `d` 속성에 그리기 명령을 포함합니다.

### 기본 명령어

| 명령 | 매개변수 | 설명 |
|------|----------|------|
| `M/m` | x, y | 이동 (Move to) |
| `L/l` | x, y | 직선 (Line to) |
| `H/h` | x | 수평선 (Horizontal) |
| `V/v` | y | 수직선 (Vertical) |
| `Z/z` | - | 경로 닫기 (Close) |

### 곡선 명령어

| 명령 | 매개변수 | 설명 |
|------|----------|------|
| `C/c` | x1,y1 x2,y2 x,y | 베지어 곡선 (Cubic) |
| `S/s` | x2,y2 x,y | 부드러운 베지어 |
| `Q/q` | x1,y1 x,y | 2차 베지어 (Quadratic) |
| `T/t` | x, y | 부드러운 2차 베지어 |
| `A/a` | rx ry rotation large-arc sweep x y | 호 (Arc) |

### 예시

```xml
<!-- 별 모양 -->
<path d="M 50 0 L 61 35 L 98 35 L 68 57 L 79 91 L 50 70 L 21 91 L 32 57 L 2 35 L 39 35 Z"
      fill="gold" stroke="orange" stroke-width="2"/>

<!-- 하트 모양 -->
<path d="M 50 30 
         C 50 25 40 0 20 0
         C 0 0 0 20 0 30
         C 0 50 20 65 50 80
         C 80 65 100 50 100 30
         C 100 20 100 0 80 0
         C 60 0 50 25 50 30"
      fill="#e74c3c"/>
```

---

## 🖼️ XML UI에서 SVG 사용

### SVG 위젯

```xml
<svg id="icon" src="assets/icons/settings.svg" width="32" height="32"/>
```

### 인라인 SVG

```xml
<svg id="custom-icon" width="48" height="48">
    <circle cx="24" cy="24" r="20" fill="#6c5ce7"/>
    <text x="24" y="30" font-size="14" fill="white" text-anchor="middle">!</text>
</svg>
```

---

## 🔧 Go API

### SVG 파일 로드

```go
import "ebitenui-xml/ui"

// 파일에서 로드
svgWidget, err := ui.LoadSVG("assets/icon.svg")
if err != nil {
    log.Fatal(err)
}

// 크기 설정
svgWidget.SetSize(64, 64)

// 패널에 추가
panel.AddChild(svgWidget)
```

### SVG 문자열 파싱

```go
svgContent := `<svg viewBox="0 0 100 100">
    <rect width="100" height="100" fill="blue"/>
    <circle cx="50" cy="50" r="30" fill="white"/>
</svg>`

svgWidget, err := ui.ParseSVGString(svgContent)
if err != nil {
    log.Fatal(err)
}
```

### 동적 색상 변경

```go
// 채우기 색상 변경
svgWidget.SetFillColor("myRect", color.RGBA{255, 0, 0, 255})

// 선 색상 변경
svgWidget.SetStrokeColor("myPath", color.RGBA{0, 0, 0, 255})
```

### SVG 그리기

```go
func (g *Game) Draw(screen *ebiten.Image) {
    // SVG 위젯 그리기
    g.svgWidget.Draw(screen)
    
    // 특정 위치에 그리기
    op := &ebiten.DrawImageOptions{}
    op.GeoM.Translate(100, 50)
    g.svgWidget.DrawWithOptions(screen, op)
}
```

---

## 🎨 그라디언트

### 선형 그라디언트

```xml
<defs>
    <linearGradient id="sunset" x1="0%" y1="0%" x2="0%" y2="100%">
        <stop offset="0%" stop-color="#ff7e5f"/>
        <stop offset="50%" stop-color="#feb47b"/>
        <stop offset="100%" stop-color="#86a8e7"/>
    </linearGradient>
</defs>
<rect x="0" y="0" width="200" height="100" fill="url(#sunset)"/>
```

### 방사형 그라디언트

```xml
<defs>
    <radialGradient id="sphere" cx="30%" cy="30%">
        <stop offset="0%" stop-color="white"/>
        <stop offset="50%" stop-color="#6c5ce7"/>
        <stop offset="100%" stop-color="#341f97"/>
    </radialGradient>
</defs>
<circle cx="50" cy="50" r="40" fill="url(#sphere)"/>
```

---

## 📁 아이콘 팩 구성

권장 프로젝트 구조:

```
assets/
└── icons/
    ├── ui/
    │   ├── close.svg
    │   ├── menu.svg
    │   └── settings.svg
    ├── game/
    │   ├── heart.svg
    │   ├── coin.svg
    │   └── star.svg
    └── social/
        ├── share.svg
        └── link.svg
```

### 아이콘 매니저 패턴

```go
type IconManager struct {
    icons map[string]*ui.SVGWidget
}

func NewIconManager() *IconManager {
    return &IconManager{
        icons: make(map[string]*ui.SVGWidget),
    }
}

func (im *IconManager) Load(name, path string) error {
    svg, err := ui.LoadSVG(path)
    if err != nil {
        return err
    }
    im.icons[name] = svg
    return nil
}

func (im *IconManager) Get(name string) *ui.SVGWidget {
    return im.icons[name]
}

// 사용
icons := NewIconManager()
icons.Load("settings", "assets/icons/ui/settings.svg")
icons.Load("heart", "assets/icons/game/heart.svg")

settingsIcon := icons.Get("settings")
settingsIcon.SetSize(24, 24)
```

---

## ⚠️ 제한사항

### 지원되지 않는 기능

- `<filter>` - 필터 효과
- `<clipPath>` - 클리핑 경로 (부분 지원)
- `<mask>` - 마스킹
- `<pattern>` - 패턴 채우기
- `<use>` - 재사용 (부분 지원)
- `<foreignObject>` - 외부 객체
- CSS 애니메이션
- JavaScript 상호작용

### 성능 고려사항

1. **복잡한 Path 최적화**: 많은 노드를 가진 path는 성능에 영향을 줄 수 있음
2. **캐싱**: 자주 사용되는 SVG는 이미지로 래스터화 고려
3. **ViewBox 사용**: 정확한 크기 조절을 위해 viewBox 속성 권장

```xml
<svg viewBox="0 0 100 100" width="50" height="50">
    <!-- 100x100 좌표계를 50x50 픽셀로 렌더링 -->
</svg>
```

---

## 💡 예제: 게임 HUD 아이콘

```go
// HP 아이콘 (하트)
heartSVG := `<svg viewBox="0 0 24 24">
    <path d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 
             2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09
             C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5
             c0 3.78-3.4 6.86-8.55 11.54L12 21.35z"
          fill="#e74c3c"/>
</svg>`

hp, _ := ui.ParseSVGString(heartSVG)
hp.SetSize(32, 32)

// 코인 아이콘
coinSVG := `<svg viewBox="0 0 24 24">
    <circle cx="12" cy="12" r="10" fill="#f1c40f" stroke="#f39c12" stroke-width="2"/>
    <text x="12" y="16" font-size="12" fill="#d35400" text-anchor="middle">$</text>
</svg>`

coin, _ := ui.ParseSVGString(coinSVG)
coin.SetSize(32, 32)
```

---

## 🔗 관련 문서

- [REFERENCE.md](./REFERENCE.md) - 기본 UI 레퍼런스
- [WIDGETS_EXTENDED.md](./WIDGETS_EXTENDED.md) - 확장 위젯 가이드
