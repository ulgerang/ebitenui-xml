# Phase 9 Plan: Form Validation and Messages

## Status

Completed.

## Objective

Extend the existing form state helpers into browser-like validation rules,
invalid-submit behavior, and renderable validation messages.

## Canonical Inputs

- `ui/types.go`
- `ui/widget.go`
- `ui/parser.go`
- `ui/ui.go`
- `ui/input.go`
- `ui/widgets.go`
- `ui/html_semantics_test.go`
- `docs/HTML-CSS-BINDING-GAP-LIST.md`
- `openspec/changes/complete-browser-parity-followups`

## Tasks

1. [x] Add validation metadata to `BaseWidget`:
   - `required`
   - `min` / `max`
   - `minLength` / `maxLength`
   - `pattern`
   - `type` where applicable
2. [x] Parse validation attributes from XML aliases and form widgets.
3. [x] Implement validation evaluation:
   - text input and textarea length/pattern rules
   - checkbox required semantics
   - slider numeric min/max
   - dropdown required selection
4. [x] Add submit behavior:
   - `SubmitForm` validates descendants before dispatch
   - invalid forms can skip command dispatch
   - validation messages are stored per widget
5. [x] Add render hooks/classes or state flags for validation message widgets.
6. [x] Update docs and OpenSpec.

## Evidence

- Added `ValidationRules` and per-widget validation messages on `BaseWidget`.
- Parsed XML validation attributes in `WidgetFactory.applyWidgetMetadata`.
- Added `ValidateForm`, submit blocking, and `GetValidationMessage`.
- Added tests for required/email, minlength/pattern, dropdown required, and
  slider min/max constraints.

## Verification

- `go test ./ui`
- `go test ./...`
- `go build ./...`
- `openspec validate complete-browser-parity-followups --strict`

## Done When

- Common validation rules work without custom Go code.
- Submit/reset behavior remains backward compatible.
- Validation messages are testable and renderable from XML.
