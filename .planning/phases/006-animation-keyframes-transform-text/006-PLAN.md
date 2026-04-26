# Phase 6 Plan: Keyframes, Transform Coverage, and Text Fidelity

## Status

Completed with documented follow-up limits for literal CSS `@keyframes` syntax
and deeper browser text shaping parity.

## Objective

Expand animation expressiveness and make transformed/animated rendering apply
consistently across widgets, while improving text fidelity.

## Canonical Inputs

- `ui/animation.go`
- `ui/style.go`
- `ui/widget.go`
- `ui/widgets.go`
- `ui/widgets_extended.go`
- `ui/input.go`
- `ui/types.go`

## Tasks

1. [x] Design JSON representation for CSS-like keyframes:
   - preserve current registered animation presets
   - support custom named keyframes from style JSON
2. [x] Parse `@keyframes`-like data from JSON:
   - percentages
   - from/to aliases
   - transform, opacity, color, shadow fields
3. [x] Connect parsed keyframes to existing `AnimationManager`.
4. [x] Apply full-content compositing to remaining widget types:
   - Toggle
   - Checkbox
   - RadioButton
   - Dropdown
   - Modal
   - Tooltip
   - Badge
   - Toast
   - Spinner where applicable
   - SVG/Icon where applicable
5. [x] Add transform/keyframe parser tests for composed animation data.
6. [x] Preserve current text metrics behavior and document remaining browser
   parity gaps:
   - line-height consistency
   - baseline alignment
   - fallback behavior when configured font face is unavailable
7. [x] Add tests for custom keyframe parsing, animation progress, and widget
   compositing.

## Verification

- `go test ./ui`
- `go test ./...`
- `go build ./...`
- Add a small animation smoke demo or CSS testloop case where deterministic.

## Done When

- Custom JSON keyframes can animate supported properties.
- All built-in widgets respect transform/filter/opacity/animation consistently.
- Text layout changes are covered by tests or visual fixtures.

## Evidence

- `go test ./ui`
