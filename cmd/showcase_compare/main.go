package main

import (
	"fmt"
	"html/template"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	ebitenPath  = "cmd/showcase/ebiten_showcase.png"
	browserPath = "cmd/showcase/browser_showcase.png"
	diffPath    = "cmd/showcase/diff_showcase.png"
	reportPath  = "cmd/showcase/report_showcase.html"

	threshold = 10
)

type reportData struct {
	DiffPct  float64
	AvgDelta float64
	Status   string
	Rule     string
}

func main() {
	fmt.Println("=== Showcase Compare ===")

	fmt.Println("1) Render ebiten showcase")
	if err := run("go", "run", "./cmd/showcase"); err != nil {
		fmt.Printf("render failed: %v\n", err)
		return
	}

	fmt.Println("2) Capture browser showcase")
	if err := captureBrowser(); err != nil {
		fmt.Printf("browser capture failed: %v\n", err)
		return
	}

	fmt.Println("3) Compare images")
	ebitenImg, err := loadPNG(ebitenPath)
	if err != nil {
		fmt.Printf("read ebiten png failed: %v\n", err)
		return
	}
	browserImg, err := loadPNG(browserPath)
	if err != nil {
		fmt.Printf("read browser png failed: %v\n", err)
		return
	}

	diffPct, avgDelta, diffImg := compare(ebitenImg, browserImg)
	if err := savePNG(diffPath, diffImg); err != nil {
		fmt.Printf("write diff png failed: %v\n", err)
		return
	}
	if err := writeReport(reportPath, diffPct, avgDelta); err != nil {
		fmt.Printf("write report failed: %v\n", err)
		return
	}

	status := verdict(diffPct, avgDelta)
	fmt.Printf("Done: %s (diff=%.2f%%, avgDelta=%.2f)\n", status, diffPct, avgDelta)
	fmt.Printf("Report: %s\n", reportPath)
	fmt.Printf("Diff  : %s\n", diffPath)
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runQuiet(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}
	if len(out) > 0 {
		s := string(out)
		if len(s) > 160 {
			s = s[:160] + "..."
		}
		return fmt.Errorf("%v (%s)", err, s)
	}
	return err
}

func captureBrowser() error {
	// Preferred: existing Puppeteer script
	if err := runQuiet("node", "cmd/showcase/capture_browser.mjs"); err == nil {
		return nil
	} else {
		fmt.Printf("  node capture unavailable, fallback to Chrome CLI: %v\n", err)
	}

	// Fallback: Chrome CLI headless screenshot
	chromeCandidates := []string{
		"C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe",
		"C:\\Program Files (x86)\\Google\\Chrome\\Application\\chrome.exe",
		"C:\\Program Files\\Microsoft\\Edge\\Application\\msedge.exe",
		"C:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe",
		"chrome.exe",
		"msedge.exe",
	}

	var chrome string
	for _, c := range chromeCandidates {
		if filepath.IsAbs(c) {
			if _, err := os.Stat(c); err == nil {
				chrome = c
				break
			}
		} else {
			if _, err := exec.LookPath(c); err == nil {
				chrome = c
				break
			}
		}
	}
	if chrome == "" {
		return fmt.Errorf("no node+puppeteer and no Chrome/Edge found")
	}

	absHTML, err := filepath.Abs("cmd/showcase/index.html")
	if err != nil {
		return err
	}
	absShot, err := filepath.Abs(browserPath)
	if err != nil {
		return err
	}
	url := "file:///" + filepath.ToSlash(absHTML)

	return run(chrome,
		"--headless=new",
		"--disable-gpu",
		"--no-sandbox",
		"--hide-scrollbars",
		"--window-size=960,640",
		"--screenshot="+absShot,
		url,
	)
}

func loadPNG(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func savePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func compare(a, b image.Image) (diffPct float64, avgDelta float64, out image.Image) {
	ba := a.Bounds()
	bb := b.Bounds()
	w := ba.Dx()
	h := ba.Dy()
	if bb.Dx() < w {
		w = bb.Dx()
	}
	if bb.Dy() < h {
		h = bb.Dy()
	}

	diff := 0
	total := w * h
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			ar, ag, ab, _ := a.At(x+ba.Min.X, y+ba.Min.Y).RGBA()
			br, bg, bb2, _ := b.At(x+bb.Min.X, y+bb.Min.Y).RGBA()

			dr := abs(int(ar>>8) - int(br>>8))
			dg := abs(int(ag>>8) - int(bg>>8))
			db := abs(int(ab>>8) - int(bb2>>8))
			d := float64(dr+dg+db) / 3.0
			avgDelta += d

			if dr > threshold || dg > threshold || db > threshold {
				diff++
				img.Set(x, y, color.RGBA{255, 59, 48, 255})
			} else {
				img.Set(x, y, color.RGBA{uint8(ar >> 8), uint8(ag >> 8), uint8(ab >> 8), 100})
			}
		}
	}

	if total > 0 {
		diffPct = float64(diff) * 100 / float64(total)
		avgDelta = avgDelta / float64(total)
	}
	return diffPct, avgDelta, img
}

func writeReport(path string, diffPct, avgDelta float64) error {
	status := verdict(diffPct, avgDelta)

	tpl := `<!DOCTYPE html>
<html><head><meta charset="UTF-8"><title>Showcase Compare Report</title>
<style>
*{box-sizing:border-box}body{margin:0;background:#0b1118;color:#d3dfeb;font-family:'Segoe UI',system-ui,sans-serif}
.wrap{max-width:1160px;margin:24px auto;padding:0 16px}
.cards{display:flex;gap:12px;margin:14px 0 20px}.card{flex:1;background:#121b26;border:1px solid #22313f;border-radius:10px;padding:12px}
.val{font-size:28px;font-weight:700}.lbl{font-size:12px;color:#8fa3b7}
.grid{display:grid;grid-template-columns:1fr 1fr;gap:12px}
.panel{background:#121b26;border:1px solid #22313f;border-radius:10px;padding:10px}
.panel h3{margin:0 0 8px;font-size:13px;color:#9ec7df}
img{width:100%;display:block;border-radius:6px;background:#070c12}
.rule{margin:0 0 10px;color:#8fa3b7;font-size:12px}
</style></head><body>
<div class="wrap">
  <h1 style="margin:0 0 8px">Showcase Compare Report</h1>
  <p class="rule">{{.Rule}}</p>
  <div class="cards">
    <div class="card"><div class="val">{{printf "%.2f%%" .DiffPct}}</div><div class="lbl">Diff Pixels</div></div>
    <div class="card"><div class="val">{{printf "%.2f" .AvgDelta}}</div><div class="lbl">Average Delta</div></div>
    <div class="card"><div class="val">{{.Status}}</div><div class="lbl">Visual Verdict</div></div>
  </div>
  <div class="grid">
    <div class="panel"><h3>Browser Reference</h3><img src="browser_showcase.png" /></div>
    <div class="panel"><h3>Ebiten Render</h3><img src="ebiten_showcase.png" /></div>
    <div class="panel" style="grid-column:1 / span 2"><h3>Diff</h3><img src="diff_showcase.png" /></div>
  </div>
</div>
</body></html>`

	t, err := template.New("report").Parse(tpl)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return t.Execute(f, reportData{
		DiffPct:  diffPct,
		AvgDelta: avgDelta,
		Status:   status,
		Rule:     "Rule: PASS if Diff<5% OR AvgDelta<10. WARN if AvgDelta<14. Otherwise FAIL.",
	})
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func verdict(diffPct, avgDelta float64) string {
	if diffPct < 5 || avgDelta < 10 {
		return "PASS"
	}
	if avgDelta < 14 {
		return "WARN"
	}
	return "FAIL"
}
