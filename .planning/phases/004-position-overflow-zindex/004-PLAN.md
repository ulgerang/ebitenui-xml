# Phase 4 Plan: Positioning, Overflow, and Z Ordering

## Status

Completed.

## Objective

Add browser-like positioned layout, child clipping, and stable z-order for both
rendering and input hit testing.

## Canonical Inputs

- `ui/layout.go`
- `ui/widget.go`
- `ui/ui.go`
- `ui/types.go`
- `ui/style.go`
- `cmd/css_testloop/testcases.go`

## Tasks

1. [x] Add layout tests for relative parent and absolute child scenarios.
2. [x] Implement `position: absolute`:
   - position against nearest positioned ancestor or parent
   - support `top`, `right`, `bottom`, `left`
   - support explicit width/height and inset-derived size
3. [x] Preserve normal flex layout for non-absolute siblings.
4. [x] Implement `overflow: hidden` clipping:
   - clip child drawing to parent content box
   - keep hit testing consistent with clipped area
5. [x] Implement z-index ordering:
   - sort draw order by `z-index`
   - sort hit testing in reverse visual order
   - preserve source order for equal z-index
6. [x] Add regression coverage for overlap, clipping, and absolute positioning.

## Verification

- `go test ./ui`
- `go test ./...`
- `go build ./...`
- CSS testloop render/reference generation for overlap and clipping cases.

## Done When

- Absolute children do not disturb flex siblings.
- Clipped children cannot draw or receive input outside clipped bounds.
- Z-index affects both visual and input order.

## Evidence

- `go test ./ui`
- `openspec validate add-declarative-data-binding --strict`
