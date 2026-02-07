package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

// TestResult holds comparison result for one test case.
type TestResult struct {
	TC         CSSTestCase
	Index      int
	DiffPixels int
	TotalPix   int
	DiffPct    float64
	AvgDelta   float64
	Pass       bool // DiffPct < threshold
}

const diffThreshold = 10 // color delta threshold per channel

// CompareImages loads both screenshots, crops each cell region, and computes per-cell diffs.
func CompareImages(browserPath, ebitenPath string, cases []CSSTestCase) []TestResult {
	browserImg := loadPNG(browserPath)
	ebitenImg := loadPNG(ebitenPath)

	w, _ := gridSize(len(cases))
	results := make([]TestResult, len(cases))

	for i, tc := range cases {
		region := CellRegion(i, w)
		r := TestResult{TC: tc, Index: i}
		r.DiffPixels, r.TotalPix, r.DiffPct, r.AvgDelta = diffRegion(browserImg, ebitenImg, region)
		r.Pass = r.DiffPct < 5.0 // less than 5% different = pass
		results[i] = r
	}
	return results
}

func diffRegion(img1, img2 image.Image, region image.Rectangle) (diffPx, totalPx int, pct, avgDelta float64) {
	b1 := img1.Bounds()
	b2 := img2.Bounds()

	for y := region.Min.Y; y < region.Max.Y; y++ {
		for x := region.Min.X; x < region.Max.X; x++ {
			totalPx++
			var r1, g1, b1c, r2, g2, b2c uint32
			if x < b1.Max.X && y < b1.Max.Y {
				r1, g1, b1c, _ = img1.At(x, y).RGBA()
			}
			if x < b2.Max.X && y < b2.Max.Y {
				r2, g2, b2c, _ = img2.At(x, y).RGBA()
			}
			dr := absI(int(r1>>8) - int(r2>>8))
			dg := absI(int(g1>>8) - int(g2>>8))
			db := absI(int(b1c>>8) - int(b2c>>8))
			delta := float64(dr+dg+db) / 3.0
			avgDelta += delta
			if delta > diffThreshold {
				diffPx++
			}
		}
	}
	if totalPx > 0 {
		pct = float64(diffPx) / float64(totalPx) * 100
		avgDelta /= float64(totalPx)
	}
	return
}

func absI(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func loadPNG(path string) image.Image {
	f, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot open %s: %v\n", path, err)
		// Return blank image to allow partial comparison
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot decode %s: %v\n", path, err)
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}
	return img
}

// GenerateReport creates an HTML report showing per-test-case results.
func GenerateReport(results []TestResult, outPath string) error {
	passed, failed := 0, 0
	for _, r := range results {
		if r.Pass {
			passed++
		} else {
			failed++
		}
	}

	var rows string
	for _, r := range results {
		statusClass := "pass"
		statusText := "PASS"
		if !r.Pass {
			statusClass = "fail"
			statusText = "FAIL"
		}
		rows += fmt.Sprintf(
			`<tr class="%s"><td>%s</td><td>%s</td><td>%s</td><td>%.1f%%</td><td>%.1f</td><td>%s</td></tr>`+"\n",
			statusClass, r.TC.ID, r.TC.Category, r.TC.Property, r.DiffPct, r.AvgDelta, statusText)
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html><head><meta charset="UTF-8"><title>CSS TestLoop Report</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{background:#0d1117;color:#c9d1d9;font-family:'Segoe UI',system-ui,sans-serif;padding:24px}
h1{color:#58a6ff;margin-bottom:16px}
.summary{display:flex;gap:16px;margin-bottom:24px}
.card{background:#161b22;border:1px solid #30363d;border-radius:8px;padding:16px 24px;text-align:center;flex:1}
.card .val{font-size:28px;font-weight:bold}
.card .lbl{font-size:12px;color:#8b949e;margin-top:4px}
table{width:100%%;border-collapse:collapse;background:#161b22;border:1px solid #30363d;border-radius:8px;overflow:hidden}
th{text-align:left;padding:10px 14px;background:#1c2128;color:#8b949e;font-size:12px;border-bottom:1px solid #30363d}
td{padding:8px 14px;border-bottom:1px solid #21262d;font-size:13px;font-family:monospace}
tr.pass td:last-child{color:#3fb950}
tr.fail td:last-child{color:#f85149;font-weight:bold}
tr:hover{background:#1c2128}
</style></head><body>
<h1>CSS TestLoop Comparison Report</h1>
<div class="summary">
  <div class="card"><div class="val">%d</div><div class="lbl">Total Tests</div></div>
  <div class="card"><div class="val" style="color:#3fb950">%d</div><div class="lbl">Passed (<5%% diff)</div></div>
  <div class="card"><div class="val" style="color:#f85149">%d</div><div class="lbl">Failed (>=5%% diff)</div></div>
</div>
<table>
<thead><tr><th>Test ID</th><th>Category</th><th>Property</th><th>Diff %%</th><th>Avg Delta</th><th>Status</th></tr></thead>
<tbody>
%s
</tbody></table>
</body></html>`, len(results), passed, failed, rows)

	if err := os.WriteFile(outPath, []byte(html), 0644); err != nil {
		return err
	}
	fmt.Printf("Report: %s (%d passed, %d failed)\n", outPath, passed, failed)
	return nil
}

// keep compiler happy
var (
	_ = color.White
	_ = math.Ceil
)
