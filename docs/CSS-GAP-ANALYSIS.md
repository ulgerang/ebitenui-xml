# CSS Gap Analysis: 미구현/불완전 CSS 기능 분석

> **목적**: ebitenui-xml의 CSS 렌더링을 브라우저와 비교했을 때 차이가 나는 항목을 정리하고, 수정 우선순위를 결정한다.

---

## 요약 표

| # | 기능 | 현재 상태 | 영향도 | 수정 파일 |
|---|------|----------|--------|----------|
| 1 | **Gradient 방향 (angle)** | 파싱 O, 항상 수평으로만 렌더링 | 높음 | `effects.go` |
| 2 | **justify: space-between/around/evenly** | 파싱 O, 레이아웃 미반영 | 높음 | `layout.go` |
| 3 | **box-sizing: border-box** | 미구현 (border가 크기에 추가됨) | 높음 | `layout.go`, `widget.go` |
| 4 | **Text Shadow 렌더링** | 파싱 O (`ParseTextShadow`), Draw 미호출 | 중간 | `widgets.go` |
| 5 | **min/max Width/Height** | 파싱 O, 레이아웃 제약 미적용 | 중간 | `layout.go` |
| 6 | **flex-wrap** | 파싱 O, 줄바꿈 로직 없음 | 중간 | `layout.go` |
| 7 | **개별 코너 border-radius** | 파싱 O, 렌더링은 단일값만 | 낮음 | `effects.go` |
| 8 | **overflow: hidden** | 파싱 O, 클리핑 없음 | 중간 | `widget.go` |
| 9 | **flex-shrink** | 파싱 O, 축소 로직 없음 | 낮음 | `layout.go` |
| 10 | **transition 보간** | 파싱 O, 상태 변경 시 즉시 전환 | 낮음 | `widget.go` |
| 11 | **position: absolute** | 파싱 O, 모든 위젯 flow 기반 | 낮음 | `layout.go` |
| 12 | **filter / backdrop-filter** | 구조체만 존재, 렌더링 없음 | 낮음 | `effects.go` |
| 13 | **per-side border width** | 파싱 O, 렌더링은 단일값만 | 낮음 | `effects.go` |

---

## 상세 분석

### 1. Gradient 방향 (angle) - 🔴 높음

**현재**: `effects.go:DrawGradient()` - 항상 왼쪽→오른쪽(수평)으로만 렌더링
```go
// effects.go:208 - 항상 x축 기준
for x := 0.0; x < r.W; x++ {
    t := x / r.W
    clr := interpolateGradient(g.ColorStops, t)
    // ...세로 스트라이프로 그림
}
```

**문제**: `linear-gradient(180deg, ...)` 같은 세로 그라데이션이 수평으로 나옴.

**수정 방안**: angle을 라디안으로 변환, 각 픽셀의 gradient 진행값 t를 각도 기반으로 계산
```go
// t = (x*cos(angle) + y*sin(angle)) / (w*cos(angle) + h*sin(angle))
```

### 2. justify: space-between/around/evenly - 🔴 높음

**현재**: `layout.go:116~129` - start/center/end만 구현
```go
switch style.Justify {
case JustifyCenter:
    offset = (availW - totalFixed) / 2
case JustifyEnd:
    offset = availW - totalFixed
}
```

**문제**: `space-between`, `space-around`, `space-evenly` 무시됨

**수정 방안**:
```go
case JustifyBetween:
    spaceBetween = remainingSpace / float64(len(children)-1)
    // 각 자식 사이에 spaceBetween 배치
case JustifyAround:
    spaceAround = remainingSpace / float64(len(children))
    // 양쪽에 spaceAround/2, 사이에 spaceAround
case JustifyEvenly:
    spaceEvenly = remainingSpace / float64(len(children)+1)
    // 모든 간격 동일
```

### 3. box-sizing: border-box - 🔴 높음

**현재**: border width가 위젯 크기에 추가됨 (content-box 동작)

**문제**: HTML에서는 `box-sizing: border-box`가 기본이므로 border가 크기 안에 포함. ebitenui-xml에서는 border가 크기 밖에 그려져 레이아웃 차이 발생.

**수정 방안**: `layoutChildren()`에서 자식 크기 계산 시 borderWidth를 padding처럼 내부로 처리

### 4. Text Shadow 렌더링 - 🟡 중간

**현재**: `ParseTextShadow()` 파싱 OK, `style.parsedTextShadow` 필드 존재. 하지만 `Text.Draw()`와 `Button.Draw()`에서 호출하지 않음.

**수정 방안**: 텍스트 Draw 직전에 shadow 색상으로 offset만큼 이동하여 먼저 그리기
```go
if style.parsedTextShadow != nil {
    shadowOp := &text.DrawOptions{}
    shadowOp.GeoM.Translate(x + shadow.OffsetX, y + shadow.OffsetY)
    shadowOp.ColorScale.ScaleWithColor(shadow.Color)
    text.Draw(screen, content, fontFace, shadowOp)
}
// 원본 텍스트 그리기
text.Draw(screen, content, fontFace, op)
```

### 5. min/max Width/Height - 🟡 중간

**현재**: `types.go`에 `MinWidth`, `MaxWidth`, `MinHeight`, `MaxHeight` 필드 존재. `layout.go`에서 완전히 무시.

**수정 방안**: `layoutChildren()`에서 자식 크기 결정 후 clamp 적용
```go
if childStyle.MinWidth > 0 && childRect.W < childStyle.MinWidth {
    childRect.W = childStyle.MinWidth
}
if childStyle.MaxWidth > 0 && childRect.W > childStyle.MaxWidth {
    childRect.W = childStyle.MaxWidth
}
```

### 6. flex-wrap - 🟡 중간

**현재**: `FlexWrap` 필드가 `types.go`에 정의, `Style.Merge()`에서 전파됨. `layout.go`에서 완전히 무시.

**수정 방안**: `layoutChildren()`에서 현재 줄의 자식들이 가용 공간을 초과하면 다음 줄로 이동하는 로직 추가. 복잡도 높음.

### 7. 개별 코너 border-radius - 🟢 낮음

**현재**: `BorderTopLeftRadius` 등 4개 필드 존재. `DrawRoundedRectPath()`에서 단일 `radius` 값만 사용.

**수정 방안**: Path 구성 시 각 코너에 다른 radius 사용
```go
path.QuadTo(x+w, y, x+w, y+radTopRight)  // 각 코너마다 다른 radius
```

### 8. overflow: hidden - 🟡 중간

**현재**: `Overflow` 필드 파싱만. 자식이 부모 밖으로 나가도 클리핑 없음.

**수정 방안**: Ebiten의 `SubImage`를 활용한 클리핑 또는 stencil buffer 기법
```go
if style.Overflow == "hidden" {
    // 임시 이미지에 자식 렌더링 후 부모 영역만 복사
    tmpImg := ebiten.NewImage(int(r.W), int(r.H))
    // ... 자식을 tmpImg에 그린 후 부모에 블리트
}
```

### 9. flex-shrink - 🟢 낮음

**현재**: 자식들이 부모보다 클 때 축소하지 않음.

**수정 방안**: `totalFixed > availSpace`일 때 flexShrink 비율에 따라 각 자식 크기 축소.

### 10. transition 보간 - 🟢 낮음

**현재**: hover/active 상태 전환 시 스타일이 즉시 변경됨. `parsedTransitions`는 있지만 사용되지 않음.

**수정 방안**: 상태 변경 시 이전 스타일 값을 저장하고 프레임마다 보간값을 계산. `animation.go`의 easing 함수 재활용 가능.

---

## 10-Round 개선 계획

| Round | 작업 | 예상 영향 |
|-------|------|----------|
| 1 | Gradient 방향 (angle) | DIFF -5~10% |
| 2 | justify space-between/around/evenly | DIFF -3~8% |
| 3 | box-sizing border-box | DIFF -5~15% |
| 4 | Text Shadow 렌더링 | DIFF -1~3% |
| 5 | min/max Width/Height | DIFF -2~5% |
| 6 | flex-wrap | DIFF -2~5% |
| 7 | 개별 코너 border-radius | DIFF -1~2% |
| 8 | overflow: hidden | DIFF -1~3% |
| 9 | flex-shrink | DIFF -1~3% |
| 10 | 종합 테스트 페이지 + 미세조정 | 최종 검증 |

---

## 현재 데모(layout.xml)에서 발생하는 구체적 차이

1. **#header 그라데이션**: `linear-gradient(90deg, ...)` → 수평이므로 현재도 맞을 수 있으나, angle 지원이 없어 다른 각도 사용 시 문제
2. **#footer 그라데이션**: 동일
3. **#content border**: `borderWidth: 2` → border가 위젯 크기 밖에 추가되어 전체 크기가 미세하게 다름
4. **#title textShadow**: `"2 2 4 rgba(0,0,0,0.5)"` → 파싱은 되지만 렌더링 안 됨
5. **sidebar 버튼 간격**: `gap: 12` 동작하지만, 레이아웃 기본값(30px 높이) 때문에 브라우저와 미세 차이
6. **텍스트 렌더링**: 비트맵 폰트 사용으로 브라우저 monospace 폰트와 글리프 차이 (이것은 본질적 한계)
