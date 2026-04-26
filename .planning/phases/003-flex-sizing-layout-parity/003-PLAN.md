# Phase 3 Plan: Flex and Sizing Layout Parity

Status: Completed

## Objective

Close the most visible CSS layout gaps in flex distribution, sizing constraints,
wrapping, and shrinking.

## Canonical Inputs

- `docs/HTML-CSS-BINDING-GAP-LIST.md`
- `ui/layout.go`
- `ui/types.go`
- `ui/style.go`
- `cmd/css_testloop/testcases.go`

## Tasks

1. Characterize current flex behavior with unit tests before changing layout.
2. Complete justify distribution:
   - `space-between`
   - `space-around`
   - `space-evenly`
3. Implement `box-sizing` style field:
   - `content-box`
   - `border-box`
   - include padding/border effects in layout calculations
4. Enforce min/max size constraints during flex layout:
   - `min-width`, `max-width`
   - `min-height`, `max-height`
5. Implement `flex-wrap` line layout:
   - row and column flows
   - gap between items and lines
   - align behavior per line
6. Implement `flex-shrink`:
   - shrink only along main axis
   - respect min sizes
   - keep zero-shrink children stable
7. Add CSS visual regression test cases for each layout feature.

## Verification

- `go test ./ui`
- `go test ./...`
- `go build ./...`
- `go run ./cmd/css_testloop -mode render -out ebiten.png`
- `go run ./cmd/css_testloop -mode html -out reference.html`

## Done When

- Unit tests cover flex math edge cases.
- CSS testloop has browser-comparison cases for new layout behavior.

## Completion Evidence

- Added `boxSizing` to `Style`.
- Updated `ui/layout.go` with content-box/border-box sizing, wrap lines, and
  shrink behavior that respects zero-shrink and min-size constraints.
- Added Phase 3 layout coverage in `ui/layout_test.go`.
- Added OpenSpec CSS layout requirements.
