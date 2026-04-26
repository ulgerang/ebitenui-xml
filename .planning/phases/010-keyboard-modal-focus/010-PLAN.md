# Phase 10 Plan: Keyboard and Modal Focus Policies

## Status

Completed.

## Objective

Make keyboard behavior predictable for forms, dropdowns, sliders, radio groups,
and modal dialogs.

## Canonical Inputs

- `ui/ui.go`
- `ui/input.go`
- `ui/widgets.go`
- `ui/widgets_extended.go`
- `ui/types.go`
- `ui/html_semantics_test.go`
- `docs/HTML-CSS-BINDING-GAP-LIST.md`
- `openspec/changes/complete-browser-parity-followups`

## Tasks

1. [x] Define a keyboard event policy table:
   - Tab / Shift+Tab
   - Arrow keys
   - Enter
   - Escape
2. [x] Implement dropdown keyboard navigation:
   - open/close
   - highlighted option
   - select with Enter
   - close with Escape
3. [x] Implement radio group arrow navigation.
4. [x] Improve slider keyboard changes with step support.
5. [x] Implement modal focus trap:
   - focus cycles inside modal while open
   - focus restores to previous widget on close
6. [x] Add tests using simulation helpers where possible.
7. [x] Update docs and OpenSpec.

## Evidence

- Added generic keyboard dispatch through `SimulateKeyPress` and runtime key
  handling.
- Added dropdown open/highlight/select/close keyboard behavior.
- Added radio group relative arrow movement and slider step increments.
- Added modal focus trap/restore logic scoped to open modal descendants.
- Added tests for dropdown, radio, slider, and modal focus behavior.

## Verification

- `go test ./ui`
- `go test ./...`
- `go build ./...`
- `openspec validate complete-browser-parity-followups --strict`

## Done When

- Keyboard interaction is deterministic and covered by unit tests.
- Modal focus cannot escape while modal is open.
- Existing pointer behavior remains unchanged.
