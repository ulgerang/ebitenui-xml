# Phase 5 Plan: Borders, Shadows, and Clip Paths

## Status

Completed.

## Objective

Improve CSS visual fidelity for frame primitives and shadows.

## Canonical Inputs

- `ui/effects.go`
- `ui/widget.go`
- `ui/widgets.go`
- `ui/style.go`
- `ui/types.go`
- `ui/shaders/`
- `cmd/css_testloop/testcases.go`

## Tasks

1. [x] Add parser support for comma-separated multi-shadow lists:
   - `boxShadow`
   - `textShadow`
   - preserve color functions with commas such as `rgba(...)`
2. [x] Render box shadow lists in order behind widgets.
3. [x] Render text shadow lists in order behind glyphs.
4. [x] Implement actual text-shadow blur:
   - render shadow text to offscreen
   - apply existing Gaussian blur pass
   - composite behind main text
5. [x] Implement per-side border widths:
   - top/right/bottom/left widths and colors
   - preserve existing uniform fast path
6. [x] Implement per-corner border radii:
   - top-left, top-right, bottom-right, bottom-left
   - apply to background, border, shadow where feasible
7. [x] Implement initial `clip-path` support:
   - `inset()`
   - `circle()`
   - defer complex polygon/path if shader or mask complexity is too high
8. [x] Add regression coverage for new parsers and clip-path shape support.

## Verification

- `go test ./ui`
- `go test ./...`
- `go build ./...`
- CSS visual regression render/reference/compare where feasible.

## Done When

- Common multi-shadow and border cases render without panics.
- Text-shadow blur is visibly blurred, not just offset.
- Unsupported clip paths fail as no-op with documented limits.

## Evidence

- `go test ./ui`
