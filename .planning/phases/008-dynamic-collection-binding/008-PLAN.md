# Phase 8 Plan: Dynamic Collection Binding

## Status

Completed.

## Objective

Add declarative collection binding for complex widgets, starting with
dropdown/select option lists, while improving expression error visibility.

## Canonical Inputs

- `ui/binding.go`
- `ui/binding_expr.go`
- `ui/parser.go`
- `ui/widgets_extended.go`
- `ui/ui.go`
- `ui/binding_test.go`
- `docs/HTML-CSS-BINDING-GAP-LIST.md`
- `openspec/changes/complete-browser-parity-followups`

## Tasks

1. [x] Define XML attributes for option binding:
   - `bind-options` / `data-bind-options`
   - `option-label`
   - `option-value`
   - optional placeholder fallback behavior
2. [x] Implement dropdown option collection binding:
   - map slices of primitives to label/value pairs
   - map slices of structs/maps using configured label/value paths
   - rebuild options when the bound collection changes
   - preserve selected value when possible
3. [x] Add binding diagnostics:
   - expression parse/evaluation errors include widget ID and attribute name
   - runtime failures do not panic the render/update loop
4. [x] Add only high-value expression helpers discovered during implementation.
5. [x] Update OpenSpec, gap list, and README support notes.

## Verification

- `go test ./ui`
- `go test ./...`
- `go build ./...`
- `openspec validate complete-browser-parity-followups --strict`

## Done When

- Dropdown/select options can be driven entirely from bound runtime data.
- Selection and option label/value mapping are covered by tests.
- Binding errors are visible enough for developers to fix XML/data issues.

## Evidence

- `go test ./ui`
