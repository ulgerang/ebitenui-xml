# Phase 11 Plan: Scroll Overflow Runtime

## Status

Completed.

## Objective

Turn `overflow: scroll` and `overflow: auto` into real scrollable containers
with scroll offsets, input handling, hit-test mapping, and visual regression
coverage.

## Canonical Inputs

- `ui/types.go`
- `ui/widget.go`
- `ui/layout.go`
- `ui/ui.go`
- `ui/input.go`
- `cmd/css_testloop/testcases.go`
- `docs/HTML-CSS-BINDING-GAP-LIST.md`
- `openspec/changes/complete-browser-parity-followups`

## Tasks

1. [x] Add scroll state to `BaseWidget` or a narrow scrollable helper:
   - scroll X/Y offset
   - content extents
   - max scroll clamping
2. [x] Compute scrollable content bounds after layout.
3. [x] Apply scroll offset during overflow child drawing.
4. [x] Map hit testing through scroll offsets while respecting clipping bounds.
5. [x] Add input handling:
   - mouse wheel scroll
   - optional draggable scrollbar/thumb
6. [x] Decide scrollbar visual scope:
   - minimal thumb rendering
   - or documented no-scrollbar MVP
7. [x] Add Go tests and CSS testloop cases.
8. [x] Update docs and OpenSpec.

## Evidence

- Added BaseWidget scroll offsets/content extents and public UI scroll helpers.
- Layout now computes overflow content bounds and clamps scroll offsets.
- Overflow drawing translates whole child subtrees through scroll offsets.
- Hit testing maps pointer coordinates into scrolled child coordinates.
- Added wheel dispatch for hovered overflow containers.
- Added `TestOverflowScrollRuntimeOffsetAndHitTesting` and a CSS testloop
  `overflow-scroll` fixture.

## Verification

- `go test ./ui`
- `go test ./...`
- `go build ./...`
- CSS testloop render/html for scroll fixtures
- `openspec validate complete-browser-parity-followups --strict`

## Done When

- Overflow scroll containers can reveal clipped children via scroll offset.
- Hit testing and pointer interaction target the scrolled visual content.
- Auto/scroll behavior is documented with current limitations.
