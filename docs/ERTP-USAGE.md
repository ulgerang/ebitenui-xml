# ERTP 사용 가이드 (CSS 비교 테스트용)

> **ebiten-ertp**: Ebitengine 게임을 HTTP로 원격 제어/스크린샷 캡처하는 프로토콜.
> CSS 비교 테스트에서는 주로 **스크린샷 캡처** 기능만 사용.

---

## 1. 위치 & 의존성

```
E:\works\ebiten-ertp\        # 상위 폴더에 존재
E:\works\ebitenui-xml\go.mod  # replace github.com/ulgerang/ebiten-ertp => ../ebiten-ertp
```

---

## 2. 게임에 ERTP 통합하기 (이미 구현됨)

`cmd/css_compare/main.go`에 이미 통합되어 있다:

```go
import debug "github.com/ulgerang/ebiten-ertp/debug"

// 초기화
g.debugServer = debug.New()
g.debugServer.Start(":9222")

// Update()에서
g.debugServer.UpdateTick()
g.debugServer.Input.Update()

// Draw()에서
g.debugServer.CaptureFrame(screen)  // 매 프레임 캡처
```

---

## 3. HTTP API (테스트에 필요한 것만)

| 메서드 | 엔드포인트 | 용도 |
|--------|-----------|------|
| GET | `/screenshot` | 현재 프레임 PNG 캡처 |
| GET | `/state` | 게임 상태 (tick, FPS, 화면 크기) |
| GET | `/` | 브라우저 라이브 프리뷰 대시보드 |

### 스크린샷 캡처 (가장 많이 쓰는 기능)

```bash
# curl로 캡처
curl http://localhost:9222/screenshot -o screenshot.png

# PowerShell로 캡처
Invoke-WebRequest http://localhost:9222/screenshot -OutFile screenshot.png
```

### 상태 확인

```bash
curl http://localhost:9222/state
# {"tick":42,"tps":60,"fps":60,"screen":{"width":640,"height":480}}
```

---

## 4. CSS 비교 워크플로우

### Step 1: ebitenui-xml 스크린샷 (ERTP 활용)

```bash
# 1. CSS Compare 하네스 실행
go run ./cmd/css_compare/ -layout assets/layout.xml -styles assets/styles.json -width 640 -height 480

# 2. 서버 준비 대기 (약 1~2초)
# 3. 스크린샷 캡처
curl http://localhost:9222/screenshot -o actual.png
```

### Step 2: HTML 레퍼런스 생성 + 브라우저 스크린샷

```bash
# 1. HTML 생성
go run ./tools/css_compare/cmd/converter/ \
  -layout assets/layout.xml -styles assets/styles.json \
  -out reference.html -width 640 -height 480

# 2. 브라우저에서 열어 스크린샷 (수동 또는 Playwright/Puppeteer)
```

### Step 3: 픽셀 비교

```bash
go run ./tools/css_compare/cmd/pixeldiff/ expected.png actual.png diff.png
# DIFF_PIXELS=1234
# DIFF_PCT=0.40
```

---

## 5. PowerShell 클라이언트 (선택사항)

```powershell
# 클라이언트 로드
. "E:\works\ebiten-ertp\scripts\ERTP-Client.ps1"

# 클라이언트 생성
$client = New-ERTPClient -Port 9222

# 서버 준비 대기
Wait-ERTPReady $client -TimeoutSeconds 30

# 스크린샷
Get-ERTPScreenshot $client -OutputPath "./actual.png"

# 상태 확인
$state = Get-ERTPState $client
Write-Host "Screen: $($state.screen.width)x$($state.screen.height)"
```

---

## 6. 프레임 녹화 (애니메이션 비교 시)

```bash
# 녹화 시작 (50ms 간격 = 20FPS)
curl -X POST "http://localhost:9222/recording/start?interval=50&dir=./frames"

# ... 대기 ...

# 녹화 중지
curl -X POST http://localhost:9222/recording/stop

# 결과: frames/frame_00000.png, frame_00001.png, ...
```

---

## 7. ERTP 서버 구조 (참고용)

```
debug/
├── server.go            # HTTP 서버 + /screenshot, /state, /recording 등
├── input.go             # VirtualInput (원격 입력 주입)
└── widget_inspector.go  # 위젯 트리 JSON 직렬화 타입
```

- `CaptureFrame(screen)`: Draw()에서 호출, 현재 프레임을 내부 버퍼에 복사
- `ReadPixels()` → PNG 인코딩 → HTTP 응답
- CORS 헤더 자동 설정 (브라우저에서도 접근 가능)
