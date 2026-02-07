package main

import (
	"fmt"
	"math"
	"os"
	"strings"
)

// GenerateHTML creates an HTML reference page that renders all SVG test cases
// in the same grid layout as the Ebiten renderer.
func GenerateHTML(cases []SVGTestCase, outPath string) error {
	w, h := gridSize(len(cases))

	var cells strings.Builder

	for i, tc := range cases {
		region := CellRegion(i, w)
		wrapID := fmt.Sprintf("cell-%d", i)

		// Container div positioned absolutely in grid
		cells.WriteString(fmt.Sprintf(
			`<div id="%s" class="cell" style="left:%dpx;top:%dpx;width:%dpx;height:%dpx">`,
			wrapID, region.Min.X, region.Min.Y, CellW, CellH))

		// The SVG is rendered at CellW x CellH, centered
		// We use the exact same SVG string as Ebiten
		cells.WriteString(tc.SVG)
		cells.WriteString("</div>\n")

		// Label above cell
		cells.WriteString(fmt.Sprintf(
			`<div class="label" style="left:%dpx;top:%dpx">%s</div>`+"\n",
			region.Min.X, region.Min.Y-16, tc.ID))
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html><head><meta charset="UTF-8"><title>SVG TestLoop Reference</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{width:%dpx;height:%dpx;overflow:hidden;background:#12121e;font-family:monospace;position:relative}
.cell{position:absolute;overflow:hidden;display:flex;align-items:center;justify-content:center;background:#12121e}
.cell svg{width:%dpx;height:%dpx}
.label{position:absolute;color:#b4b4c8;font-size:12px;font-family:monospace}
</style></head><body>
%s
</body></html>`, w, h, CellW, CellH, cells.String())

	if err := os.WriteFile(outPath, []byte(html), 0644); err != nil {
		return err
	}
	fmt.Printf("Generated HTML reference: %s (%dx%d, %d cases)\n", outPath, w, h, len(cases))
	return nil
}

// unused but keeps compiler happy
var _ = math.Ceil
