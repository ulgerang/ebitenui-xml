# 텍스트 렌더링 및 정밀 레이아웃 최적화 기록

이 문서는 EbitenUI-XML 엔진에서 브라우저 수준의 텍스트 정렬 및 레이아웃 정밀도를 달성하기 위해 적용된 공학적 해결 방안을 기록합니다.

## 1. 해결된 핵심 문제 (Identified Issues)

### 1.1 "Sinking Text" (글자 쏠림 현상)
*   **현상**: 버튼이나 사이드바 메뉴 내의 텍스트가 시각적으로 1~2픽셀 아래로 처지는 현상.
*   **원인**: 단순히 `Ascent` 값만 사용하여 폰트의 베이스라인을 잡을 경우, 폰트 내부의 상단 여백(Internal Leading)과 대문자 높이(Cap Height)의 차이로 인해 시각적 불균형이 발생함.
*   **해결**: **CSS Half-Leading 모델** 도입.

### 1.2 레이아웃 압착 (Layout Compression)
*   **현상**: 요소가 밀집된 공간(예: 통계 박스)에서 글자가 서로 겹치거나 간격(`gap`)이 무시되는 현상.
*   **원인**: 레이아웃 엔진이 텍스트의 실제 물리적 부피(Line-box)가 아닌 글자 픽셀 끝단만을 크기로 인식하여 계산함.
*   **해결**: `ContentSizer` 인터페이스 고도화 및 `shrinkFactor` 보호 로직 적용.

---

## 2. 적용된 공학적 솔루션 (Engineering Solutions)

### 2.1 CSS Half-Leading 모델 (`ui/widgets.go`)
브라우저의 표준 텍스트 렌더링 방식을 재현하기 위해 다음과 같은 공식을 적용했습니다.
*   **Line-Height 계산**: 기본적으로 폰트 크기의 **1.2배**를 `line-box` 높이로 설정.
*   **수직 중앙 정렬 공식**:
    ```go
    emHeight := metrics.HAscent + metrics.HDescent
    halfLeading := (lineHeight - emHeight) / 2
    y := startY + halfLeading + metrics.HAscent
    ```
    이 방식은 글자 위아래에 동일한 여백을 배분하여, 어떤 폰트 크기에서도 시각적 중앙을 보장합니다.

### 2.2 레이아웃 무결성 보호 (`ui/layout.go`)
좁은 공간에서도 텍스트가 찌그러지지 않도록 레이아웃 엔진을 수정했습니다.
*   **Gap 불변의 법칙**: 전체 공간이 부족하여 요소 크기를 줄여야 할 때(`shrinkFactor` 적용 시), 요소 사이의 `gap`은 축소 대상에서 제외하여 가독성을 확보함.
*   **Intrinsic Size 보호**: 텍스트와 같이 자기 크기를 가진 위젯(`ContentSizer`)은 설정된 `min-content` 이하로 줄어들지 않도록 보호 로직 추가.

### 2.3 True Bold 지원 (`ui/ui.go`)
*   단순히 픽셀을 두껍게 그리는 방식이 아니라, `arialbd.ttf`와 같이 실제 **Bold 전용 폰트 데이터**를 로드하여 사용하도록 구조 개선.
*   스타일의 `fontWeight: bold` 속성을 감지하여 엔진이 자동으로 적절한 폰트 소스(`DefaultBoldFont`)를 선택함.

### 2.4 Letter Spacing 및 메트릭 정밀도
*   **Letter Spacing**: 글자 사이의 미세한 간격을 조정하는 `letter-spacing` 속성 구현.
*   **HLineGap 반영**: 텍스트의 높이 보고 시 폰트 자체의 `LineGap` 메트릭을 포함시켜 레이아웃 엔진이 텍스트 간의 물리적 충돌을 방지함.

---

## 3. 향후 유지보수 가이드

*   **폰트 교체 시**: 새로운 `.ttf` 폰트를 적용할 때, 반드시 `HAscent`와 `CapHeight`가 조화로운지 확인해야 합니다.
*   **레이아웃 디버깅**: 글자가 잘리거나 정렬이 어긋난다면 `ui/widgets.go`의 `textLineHeight` 함수와 `halfLeading` 계산 부분을 가장 먼저 점검하십시오.
*   **표준 준수**: 모든 텍스트 정렬은 수치상의 정렬보다 브라우저(Chrome/Edge)의 렌더링 결과와 시각적으로 대조하여 검증하는 것을 원칙으로 합니다.
