# Phase 2 Plan: HTML/XML Structure and Form Semantics

Status: Completed

## Objective

Make XML authoring feel closer to practical HTML by adding semantic aliases,
query helpers, form behavior, grouping, and keyboard focus navigation.

## Canonical Inputs

- `docs/HTML-CSS-BINDING-GAP-LIST.md`
- `ui/parser.go`
- `ui/ui.go`
- `ui/widgets.go`
- `ui/widgets_extended.go`
- `ui/input.go`

## Tasks

1. Add semantic aliases in `WidgetFactory.createWidget`:
   - headings: `h1` to `h6`
   - lists: `ul`, `ol`, `li`
   - table-like: `table`, `thead`, `tbody`, `tr`, `td`, `th`
   - landmarks/groups: `form`, `fieldset`, `legend`, `nav`, `section`,
     `article`, `header`, `footer`, `main`
2. Add default classes or type metadata where aliases need distinguishable
   styling without new widget types.
3. Add DOM-like query helpers:
   - `QueryByClass(class string) []Widget`
   - `QueryByType(typeName string) []Widget`
   - `Query(root Widget, selector string) []Widget` for simple `.class`, `#id`,
     and type selectors.
4. Add form container behavior:
   - submit/reset command dispatch
   - validation state API on input widgets
   - disabled propagation from fieldset/form groups
5. Add XML radio/select grouping:
   - radio groups by `name`
   - select/options from XML and dynamic option binding compatibility
6. Add focus traversal:
   - `tabindex`
   - next/previous focusable widget
   - keyboard Tab/Shift+Tab handling in UI update path
7. Add tests for aliases, queries, form reset/submit, grouping, and focus order.

## Verification

- `go test ./ui`
- `go test ./...`
- `go build ./...`
- Optional smoke demo under `cmd/` for form traversal.

## Done When

- Common HTML-like XML structures parse without custom Go setup.
- Form submit/reset and grouped inputs work from XML.
- Keyboard traversal is deterministic and tested.

## Completion Evidence

- Added semantic metadata, tabindex, focusable, reset, and validation metadata
  to `BaseWidget`.
- Added HTML semantic alias parsing, radio grouping, fieldset disabled
  propagation, and form command tracking in `ui/parser.go`.
- Added DOM-like query helpers, form submit/reset helpers, validation helpers,
  and focus traversal in `ui/ui.go`.
- Added coverage in `ui/html_semantics_test.go`.
