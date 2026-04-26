# Phase 12 Plan: CSS Keyframes Syntax Parser

## Status

Completed.

## Objective

Support literal CSS-like `@keyframes` syntax in addition to the current JSON
`keyframes` object.

## Canonical Inputs

- `ui/style.go`
- `ui/animation.go`
- `ui/selector.go`
- `ui/effects_runtime_test.go`
- `docs/HTML-CSS-BINDING-GAP-LIST.md`
- `openspec/changes/complete-browser-parity-followups`

## Tasks

1. [x] Choose ingestion API:
   - extend `LoadStyles` to detect CSS syntax
   - or add `LoadCSS` / `LoadStyleSheetCSS`
2. [x] Implement a small parser for:
   - `@keyframes name { ... }`
   - `from` / `to`
   - percentage selectors
   - declarations inside each keyframe block
3. [x] Map supported declarations to `KeyframeProperties`:
   - `opacity`
   - `transform`
   - `width` / `height`
   - supported colors and shadow fields where possible
4. [x] Return structured parse errors for malformed blocks.
5. [x] Add tests for valid syntax, unsupported declarations, and malformed input.
6. [x] Update docs and OpenSpec.

## Evidence

- Added `StyleEngine.LoadCSS` and `UI.LoadCSS`.
- Added non-JSON `@keyframes` detection in `LoadFromString`.
- Parsed `from`, `to`, percentages, and comma-separated selectors.
- Mapped supported declarations through existing `KeyframeStyle` conversion.
- Added valid and malformed CSS keyframe tests.

## Verification

- `go test ./ui`
- `go test ./...`
- `go build ./...`
- `openspec validate complete-browser-parity-followups --strict`

## Done When

- Literal `@keyframes` can register animations without JSON conversion.
- Existing JSON style loading remains backward compatible.
- Unsupported CSS declarations are no-op/warnings, not panics.
