package main

import (
	"fmt"
	"math"
	"os"
	"strings"
)

// GenerateHTML creates an HTML reference page that renders all test cases
// in the same grid layout as the Ebiten renderer.
func GenerateHTML(cases []CSSTestCase, outPath string) error {
	w, h := gridSize(len(cases))

	var cells strings.Builder
	var styles strings.Builder

	for i, tc := range cases {
		region := CellRegion(i, w)
		// Wrap each test in a positioned container
		wrapID := fmt.Sprintf("cell-%d", i)
		cells.WriteString(fmt.Sprintf(
			`<div id="%s" class="cell" style="left:%dpx;top:%dpx;width:%dpx;height:%dpx">`,
			wrapID, region.Min.X, region.Min.Y, CellW, CellH))
		cells.WriteString(tc.HTML)
		cells.WriteString("</div>\n")

		// Label
		cells.WriteString(fmt.Sprintf(
			`<div class="label" style="left:%dpx;top:%dpx">%s</div>`+"\n",
			region.Min.X, region.Min.Y-16, tc.ID))

		// Test-specific CSS scoped under #cell-N
		scoped := scopeCSS(tc.CSS, wrapID)
		styles.WriteString(scoped + "\n")
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html><head><meta charset="UTF-8"><title>CSS TestLoop Reference</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{width:%dpx;height:%dpx;overflow:hidden;background:#12121e;font-family:monospace;position:relative}
.cell{position:absolute;overflow:hidden;display:flex;flex-direction:column}
.label{position:absolute;color:#b4b4c8;font-size:12px;font-family:monospace}
%s
</style></head><body>
%s
</body></html>`, w, h, styles.String(), cells.String())

	if err := os.WriteFile(outPath, []byte(html), 0644); err != nil {
		return err
	}
	fmt.Printf("Generated HTML reference: %s (%dx%d, %d cases)\n", outPath, w, h, len(cases))
	return nil
}

// scopeCSS prefixes CSS selectors with #cellID so they don't conflict.
// Simple approach: split by "}" and prefix each rule.
func scopeCSS(css, cellID string) string {
	rules := strings.Split(css, "}")
	var out strings.Builder
	for _, rule := range rules {
		rule = strings.TrimSpace(rule)
		if rule == "" {
			continue
		}
		braceIdx := strings.Index(rule, "{")
		if braceIdx < 0 {
			continue
		}
		selector := strings.TrimSpace(rule[:braceIdx])
		body := rule[braceIdx:]

		// Split multiple selectors by comma
		sels := strings.Split(selector, ",")
		var scoped []string
		for _, s := range sels {
			s = strings.TrimSpace(s)
			scoped = append(scoped, fmt.Sprintf("#%s %s", cellID, s))
		}
		out.WriteString(strings.Join(scoped, ", ") + " " + body + "}\n")
	}
	return out.String()
}

// Unused but keeps compiler happy about math import
var _ = math.Ceil
