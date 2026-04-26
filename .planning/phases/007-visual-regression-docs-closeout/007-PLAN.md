# Phase 7 Plan: Visual Regression and Documentation Closeout

## Status

Completed.

## Objective

Stabilize the expanded HTML/CSS feature set with visual tests, docs, and
OpenSpec closeout.

## Canonical Inputs

- `cmd/css_testloop/testcases.go`
- `cmd/svg_testloop/testcases.go`
- `tools/css_compare/`
- `docs/HTML-CSS-BINDING-GAP-LIST.md`
- `README.md`
- `openspec/changes/add-declarative-data-binding`

## Tasks

1. [x] Add or update CSS testloop cases for:
   - data-bound dynamic layouts where renderable
   - flex distribution and wrap
   - absolute positioning and clipping
   - z-index overlap
   - border sides/corners
   - multi-shadow and text-shadow blur
   - keyframe/transform static snapshots
2. [x] Add deterministic render fixtures where animation timing would otherwise make
   pixel comparison unstable.
3. [x] Update documentation:
   - supported XML tags/aliases
   - supported binding attributes
   - supported CSS properties
   - known limitations versus browser CSS
4. [x] Reconcile OpenSpec:
   - ensure all completed requirements have tests or visual cases
   - split any remaining future work into a new OpenSpec change if needed
   - archive completed change if project workflow supports it
5. [x] Run final validation commands and record results in docs.

## Verification

- `go test ./...`
- `go build ./...`
- `go vet ./...`
- `openspec validate add-declarative-data-binding --strict`
- CSS testloop render/html/compare path for new visual cases.

## Done When

- Docs match implementation.
- Visual regression coverage exists for high-risk layout/effect changes.
- OpenSpec no longer claims unfinished work as complete.

## Evidence

- CSS testloop cases added for absolute positioning, z-index overlap,
  multi-shadow, blurred text shadow, and inset/circle clip-path.
- `go test ./ui`
- `go test ./...`
- `go build ./...`
- `openspec validate add-declarative-data-binding --strict`
- `go run ./cmd/css_testloop -mode render -out %TEMP%/ebitenui-css-phase7.png`
- `go run ./cmd/css_testloop -mode html -out %TEMP%/ebitenui-css-phase7.html`
