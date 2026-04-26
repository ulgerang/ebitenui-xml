# CSS 비주얼 비교 시스템

> **버전**: 1.0.0  
> **최종 수정**: 2026-02-07  
> **의존성**: ebitenui-xml + ebiten-ertp + Chrome/Edge

EbitenUI-XML의 CSS 구현이 실제 브라우저 렌더링과 얼마나 일치하는지 **자동으로 비교하는 도구**입니다.
XML 레이아웃과 JSON 스타일을 HTML/CSS로 변환한 뒤, 브라우저 렌더링과 Ebiten 렌더링을 픽셀 단위로 비교합니다.

---

## 📋 개요

### 문제

EbitenUI-XML은 CSS와 유사한 스타일 시스템을 구현하지만, 실제 브라우저의 CSS 렌더링과 차이가 존재할 수 있습니다.
수동으로 비교하는 것은 비효율적이고, 새로운 CSS 속성을 추가할 때마다 회귀 테스트가 필요합니다.

### 해결책

자동화된 비교 파이프라인:

```
Layout XML + Styles JSON
     │
     ├─→ [converter] ─→ Reference HTML/CSS ─→ [Chrome headless] ─→ browser.png
     │
     └─→ [css_compare] ─→ Ebiten + ERTP ─→ [/screenshot API] ─→ ebiten.png
                                                                     │
                                           [pixeldiff] ←────────────┘
                                                │
                                           diff.png + 통계
                                                │
                                           report.html (비교 리포트)
```

---

## 📁 프로젝트 구조

```
ebitenui-xml/
├── cmd/
│   └── css_compare/
│       └── main.go                  # ERTP 테스트 하네스
├── tools/
│   └── css_compare/
│       ├── cmd/
│       │   ├── converter/
│       │   │   └── main.go          # XML+JSON → HTML/CSS 변환기
│       │   └── pixeldiff/
│       │       └── main.go          # 픽셀 비교 도구
│       ├── Run-CSSCompare.ps1       # 오케스트레이션 스크립트
│       ├── converter.exe            # 빌드된 변환기
│       ├── pixeldiff.exe            # 빌드된 비교 도구
│       └── css_compare_output/      # 생성된 출력 파일
│           ├── reference_*.html     # HTML/CSS 레퍼런스 페이지
│           ├── browser_*.png        # 브라우저 스크린샷
│           ├── ebiten_*.png         # Ebiten 스크린샷
│           ├── diff_*.png           # 픽셀 차이 시각화
│           └── report_*.html        # 비교 리포트
```

---

## 🔧 구성 요소

### 1. Converter (변환기)

XML 레이아웃과 JSON 스타일을 표준 HTML/CSS 파일로 변환합니다.

**위치**: `tools/css_compare/cmd/converter/main.go`

**사용법**:
```bash
converter.exe -layout <layout.xml> -styles <styles.json> -out <output.html> -width 640 -height 480
```

**플래그**:

| 플래그 | 기본값 | 설명 |
|--------|--------|------|
| `-layout` | `assets/layout.xml` | 레이아웃 XML 경로 |
| `-styles` | `assets/styles.json` | 스타일 JSON 경로 |
| `-out` | `reference.html` | 출력 HTML 경로 |
| `-width` | `640` | 캔버스 너비 (px) |
| `-height` | `480` | 캔버스 높이 (px) |

**변환 매핑**:

| ebitenui-xml 태그 | HTML 태그 |
|--------------------|-----------|
| `<ui>` | `<div>` |
| `<panel>` | `<div>` |
| `<button>` | `<button>` |
| `<text>` | `<span>` |
| `<progressbar>` | `<div>` (+ `.progress-fill` 자식) |
| `<image>` | `<img>` |
| `<textinput>` | `<input>` |

**스타일 변환 예시**:

```
ebitenui-xml JSON                  →    CSS
─────────────────────────────────────────────────
"direction": "column"              →    flex-direction: column
"gap": 10                          →    gap: 10px
"background": "#1a1a2e"            →    background: #1a1a2e
"borderRadius": 8                  →    border-radius: 8px
"boxShadow": "0 4 8 0 rgba(...)"   →    box-shadow: 0px 4px 8px 0px rgba(...)
"padding": {"top":10,"right":15}   →    padding: 10px 15px ...
"hover": {"background": "blue"}    →    :hover { background: blue }
```

### 2. CSS Compare 하네스

EbitenUI-XML 앱을 실행하면서 ERTP 디버그 서버를 내장합니다. 
외부에서 HTTP를 통해 스크린샷을 캡처할 수 있습니다.

**위치**: `cmd/css_compare/main.go`

**사용법**:
```bash
css_compare.exe -layout <layout.xml> -styles <styles.json> -port :9222 -width 640 -height 480
```

**플래그**:

| 플래그 | 기본값 | 설명 |
|--------|--------|------|
| `-layout` | `assets/layout.xml` | 레이아웃 XML 경로 |
| `-styles` | `assets/styles.json` | 스타일 JSON 경로 |
| `-port` | `:9222` | ERTP 서버 포트 |
| `-width` | `640` | 윈도우 너비 |
| `-height` | `480` | 윈도우 높이 |

**ERTP 엔드포인트**:

| 엔드포인트 | 메서드 | 설명 |
|------------|--------|------|
| `/screenshot` | GET | 현재 프레임 PNG 캡처 |
| `/state` | GET | 게임 상태 (tick 등) JSON |

**의존성**: [ebiten-ertp](https://github.com/ulgerang/ebiten-ertp) 프로젝트가 `../ebiten-ertp`에 있어야 합니다.

### 3. Pixel Diff (픽셀 비교)

두 PNG 이미지를 픽셀 단위로 비교하여 차이 이미지와 통계를 출력합니다.

**위치**: `tools/css_compare/cmd/pixeldiff/main.go`

**사용법**:
```bash
pixeldiff.exe <image1.png> <image2.png> <diff_output.png>
```

**출력 (stdout)**:
```
DIFF_PIXELS=217660
TOTAL_PIXELS=307200
DIFF_PCT=70.85
AVG_DELTA=22.85
```

**diff 이미지 해석**:
- **마젠타 (밝은 보라)**: 차이가 큰 영역. 밝을수록 차이가 큼
- **어두운 영역**: 일치하는 부분 (원본의 50% 밝기로 표시)
- **임계값**: 색상 델타 > 10 이상이면 "다름"으로 판정

### 4. Run-CSSCompare.ps1 (오케스트레이션)

전체 파이프라인을 하나의 명령으로 실행하는 PowerShell 스크립트입니다.

**사용법**:
```powershell
.\Run-CSSCompare.ps1 [-LayoutPath <path>] [-StylesPath <path>] [-Width <int>] [-Height <int>] [-Port <int>] [-OutputDir <path>] [-SkipBuild]
```

**파라미터**:

| 파라미터 | 기본값 | 설명 |
|----------|--------|------|
| `-LayoutPath` | `../../assets/layout.xml` | 레이아웃 XML 경로 |
| `-StylesPath` | `../../assets/styles.json` | 스타일 JSON 경로 |
| `-Width` | `640` | 캔버스 너비 |
| `-Height` | `480` | 캔버스 높이 |
| `-Port` | `9222` | ERTP 서버 포트 |
| `-OutputDir` | `./css_compare_output` | 출력 디렉토리 |
| `-SkipBuild` | `$false` | Go 바이너리 빌드 건너뛰기 |

**실행 단계**:

1. **Phase 1**: Go 도구 빌드 → HTML 레퍼런스 생성
2. **Phase 2**: Chrome Headless로 브라우저 스크린샷 캡처
3. **Phase 3**: ERTP 하네스 실행 → Ebiten 스크린샷 캡처
4. **Phase 4**: 픽셀 비교 → 비교 리포트 HTML 생성

---

## 🚀 사용법

### 사전 요구 사항

1. **Go 툴체인** 설치
2. **Chrome 또는 Edge** 브라우저 설치 (headless 스크린샷용)
3. **ebiten-ertp** 프로젝트가 `e:\works\ebiten-ertp`에 있어야 함
4. `go.mod`에 로컬 replace 설정:
   ```
   replace github.com/ulgerang/ebiten-ertp => ../ebiten-ertp
   ```

### 빌드

프로젝트 루트(`e:\works\ebitenui-xml`)에서:

```powershell
# 변환기 빌드
go build -o tools/css_compare/converter.exe ./tools/css_compare/cmd/converter

# 픽셀 비교 도구 빌드
go build -o tools/css_compare/pixeldiff.exe ./tools/css_compare/cmd/pixeldiff

# ERTP 하네스 빌드
go build -o tools/css_compare/css_compare_output/css_compare.exe ./cmd/css_compare
```

### 전체 자동 실행

```powershell
cd tools/css_compare
.\Run-CSSCompare.ps1
```

### 개별 단계별 실행

```powershell
# 1. HTML 레퍼런스 생성
.\converter.exe -layout ../../assets/layout.xml -styles ../../assets/styles.json -out ./css_compare_output/reference.html

# 2. 브라우저 스크린샷
& "C:\Program Files\Google\Chrome\Application\chrome.exe" --headless=new --disable-gpu --no-sandbox --hide-scrollbars --window-size=640,480 --screenshot=./css_compare_output/browser.png "file:///e:/works/ebitenui-xml/tools/css_compare/css_compare_output/reference.html"

# 3. Ebiten 하네스 실행 (별도 터미널)
.\css_compare_output\css_compare.exe -layout ../../assets/layout.xml -styles ../../assets/styles.json

# 4. ERTP 스크린샷 캡처
Invoke-WebRequest -Uri "http://localhost:9222/screenshot" -OutFile ./css_compare_output/ebiten.png

# 5. 픽셀 비교
.\pixeldiff.exe ./css_compare_output/browser.png ./css_compare_output/ebiten.png ./css_compare_output/diff.png
```

---

## 📊 비교 리포트 이해하기

생성된 `report_*.html`에는 다음이 포함됩니다:

### 통계 카드

| 메트릭 | 설명 | 좋은 수치 |
|--------|------|-----------|
| **Pixel Difference** | 전체 픽셀 중 다른 픽셀의 비율 | < 5% |
| **Different Pixels** | 절대 차이 픽셀 수 | - |
| **Total Pixels** | 전체 픽셀 수 (width × height) | - |
| **Avg Color Delta** | 평균 색상 차이 값 | < 5.0 |

### 색상 코드

| 색상 | Diff % | 의미 |
|------|--------|------|
| 🟢 초록 | < 5% | 우수한 일치 |
| 🟠 주황 | 5-20% | 보통 차이 |
| 🔴 빨강 | > 20% | 큰 차이 |

### CSS 속성 구현 감사

리포트 하단에 CSS 속성별 구현 상태 테이블이 있습니다:

| 태그 | 의미 |
|------|------|
| **YES** (초록) | 완전 구현됨 |
| **PARTIAL** (주황) | 부분 구현됨 |
| **NO** (빨강) | 미구현 |

---

## 📋 CSS 속성 구현 현황

### ✅ 완전 구현

| CSS 속성 | 비고 |
|----------|------|
| `display: flex` | 핵심 레이아웃 엔진 |
| `flex-direction` | row / column |
| `justify-content` | start, center, end, space-between, space-around, space-evenly |
| `align-items` | start, center, end, stretch |
| `flex-grow` | 남은 공간 분배 |
| `flex-wrap` | nowrap, wrap, wrap-reverse |
| `gap` | Flex 자식 간격 |
| `padding` / `margin` | 4방향 개별 지정 |
| `width` / `height` | 고정 크기 |
| `min/max-width/height` | 크기 제약 |
| `background` (단색) | hex, rgb, rgba, 이름 |
| `background` (그라디언트) | linear-gradient, radial-gradient |
| `color` | 텍스트 색상 |
| `border` | 너비 + 색상 |
| `border-radius` | 둥근 모서리 |
| `box-shadow` | offset, blur, spread, color, inset |
| `font-size` | 픽셀 기반 |
| `text-align` | left, center, right |
| `line-height` | 픽셀 단위 |
| `opacity` | 0-1 float |
| `:hover` / `:active` / `:disabled` / `:focus` | 상태 스타일 |
| `overflow` (scroll) | 스크롤 컨테이너 |
| CSS Variables | `--var-name` / `var(--var-name)` |
| `z-index` | 레이어 순서 |

### ⚠️ 부분 구현

| CSS 속성 | 제한 사항 |
|----------|-----------|
| `text-shadow` | 기본 지원만 |
| `transform` | translate, scale, rotate, skew |
| `transition` | 속성 애니메이션 |
| `outline` | 기본 아웃라인 |
| `position: absolute` | 제한적 위치 지정 |

### ❌ 미구현

| CSS 속성 | 이유 |
|----------|------|
| `font-family` | 비트맵 폰트만 사용 |
| `font-weight` | 비트맵 폰트 한계 |
| `text-decoration` | 미구현 |
| `backdrop-filter` | GPU 블러 미지원 |
| `cursor` | Ebiten 커서 API 없음 |
| `overflow-x` / `overflow-y` | 결합된 overflow만 |

---

## 🛠️ 트러블슈팅

### 빌드 오류: ebiten-ertp 의존성

```
cannot find module github.com/ulgerang/ebiten-ertp
```

**해결**: `go.mod`에 로컬 replace 추가:
```bash
go mod edit -require "github.com/ulgerang/ebiten-ertp@v0.0.0" -replace "github.com/ulgerang/ebiten-ertp=../ebiten-ertp"
go mod tidy
```

### Chrome headless 스크린샷 실패

Chrome이 설치되지 않았거나 경로가 다를 수 있습니다. 스크립트는 다음 경로를 순서대로 검색합니다:

1. `C:\Program Files\Google\Chrome\Application\chrome.exe`
2. `C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`
3. `C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`
4. `C:\Program Files\Microsoft\Edge\Application\msedge.exe`

### ERTP 서버 연결 실패

```
ERTP server did not respond within timeout
```

- 하네스가 정상적으로 시작되었는지 확인
- 포트 충돌 여부 확인 (`netstat -an | findstr 9222`)
- GPU 드라이버 문제 시 `EBITEN_GRAPHICS_LIBRARY=opengl` 환경변수 설정

### 높은 Pixel Difference (> 50%)

대부분 **레이아웃 차이**로 인한 것입니다. 주요 원인:

1. **Flexbox 구현 차이**: Ebiten의 커스텀 Flexbox와 브라우저의 CSS Flexbox 알고리즘이 다를 수 있음
2. **폰트 렌더링**: 비트맵 폰트 vs 브라우저 TrueType 폰트
3. **서브픽셀 렌더링**: 브라우저의 안티앨리어싱 vs Ebiten의 픽셀 렌더링

---

## 📚 관련 문서

- [REFERENCE.md](./REFERENCE.md) - 전체 스타일 레퍼런스
- [CHEATSHEET.md](./CHEATSHEET.md) - 빠른 참조
- [WIDGETS_EXTENDED.md](./WIDGETS_EXTENDED.md) - 확장 위젯 가이드
- [ebiten-ertp](../../ebiten-ertp/) - ERTP 프로토콜 문서
# CSS Visual Compare Workflow

Use this repeatable Windows-friendly workflow for the built-in CSS visual
fixtures in `cmd/css_testloop`.

```powershell
$out = "$env:TEMP\ebitenui-css"
go run ./cmd/css_testloop -mode render  -out "$out-ebiten.png"
go run ./cmd/css_testloop -mode html    -out "$out-reference.html"
```

Open or capture `$out-reference.html` in a browser at the generated viewport
size, then compare:

```powershell
go run ./cmd/css_testloop -mode compare -browser "$out-browser.png" -ebiten "$out-ebiten.png" -out "$out-report.html"
```

Smoke check without manual code edits:

```powershell
go run ./cmd/css_testloop -mode compare -browser "$out-ebiten.png" -ebiten "$out-ebiten.png" -out "$out-smoke-report.html"
```

The fixture grid now includes follow-up coverage for overflow scroll and
`clip-path: polygon(...)`. The compare report is a per-cell diff summary; it is
evidence for regressions, not a substitute for Go unit tests.

---
