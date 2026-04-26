# Phase 13 Plan: Advanced Visual Fidelity

## Status

Completed.

## Objective

Improve the remaining high-complexity browser visual parity items: complex
clip-paths, box-shadow radius precision, text metrics, font fallback, and table
layout only if justified.

## Canonical Inputs

- `ui/effects.go`
- `ui/widget.go`
- `ui/widgets.go`
- `ui/text_layout.go`
- `ui/style.go`
- `ui/layout.go`
- `cmd/css_testloop/testcases.go`
- `docs/HTML-CSS-BINDING-GAP-LIST.md`
- `openspec/changes/complete-browser-parity-followups`

## Tasks

1. [x] Implement `clip-path: polygon(...)`:
   - parse percentage and pixel points
   - mask using existing vector path compositing
   - add visual cases
2. [x] Evaluate `clip-path: path(...)`:
   - reuse SVG path parser if practical
   - otherwise document explicit no-op with tests
3. [x] Improve box-shadow radius precision:
   - evaluate per-corner shader/mask approach
   - keep current uniform approximation if performance or complexity is too high
4. [x] Improve text fidelity:
   - baseline consistency
   - line-height edge cases
   - fallback when configured font face is missing
5. [x] Table layout decision gate:
   - implement real table layout only if project usage requires it
   - otherwise keep semantic panel aliases documented
6. [x] Add targeted tests and visual fixtures.
7. [x] Update docs and OpenSpec.

## Evidence

- Added `clip-path: polygon(...)` parsing and vector mask generation.
- Added unit coverage for polygon support and explicit `path(...)` no-op.
- Added CSS visual fixture `clip-path-polygon`.
- Documented deferrals for CSS `path(...)`, per-corner box-shadow precision,
  full browser font fallback, and real table layout.

## Verification

- `go test ./ui`
- `go test ./...`
- `go build ./...`
- CSS testloop render/html for visual fixtures
- `openspec validate complete-browser-parity-followups --strict`

## Done When

- Polygon clip-path is supported.
- Path clip-path and table layout have explicit implemented-or-deferred status.
- Text and shadow fidelity improvements are backed by tests or visual fixtures.
