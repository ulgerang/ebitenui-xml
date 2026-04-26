# Phase 1 Plan: Complete Data Binding

Status: Completed

## Objective

Finish declarative data binding so XML layouts can express common dynamic UI
behavior without custom widget wiring.

## Canonical Inputs

- `docs/HTML-CSS-BINDING-GAP-LIST.md`
- `openspec/changes/add-declarative-data-binding/specs/data-binding/spec.md`
- `ui/parser.go`
- `ui/binding.go`
- `ui/ui.go`
- `ui/binding_test.go`

## Tasks

1. Extend OpenSpec with expression, attribute/style, and command-binding
   requirements.
2. Implement a small expression evaluator for template bindings:
   - fallback: `{{user.name || "Guest"}}`
   - boolean: `&&`, `||`, `!`
   - comparison: `==`, `!=`, `<`, `<=`, `>`, `>=`
   - arithmetic: `+`, `-`, `*`, `/` for numeric values
   - formatting helpers: `upper()`, `lower()`, `default()`, `number()`
3. Replace current simple template key extraction with dependency extraction
   from parsed expressions.
4. Add `bind-attr-*` and `data-bind-attr-*`:
   - class, label, content, placeholder, value, checked, disabled
   - width, height, min/max, progress value, dropdown options where feasible
5. Add `bind-style-*` and `data-bind-style-*` for safe style fields:
   - color/background/border/opacity/display/visibility
   - transform/filter/animation where parsed support already exists
6. Add command registration API:
   - `UI.RegisterCommand(name string, handler func(Widget))`
   - XML attributes: `onClick`, `onChange`, `onSubmit`
7. Add tests for expression parsing, dependency updates, attr/style updates,
   and command dispatch.
8. Update docs and OpenSpec tasks.

## Verification

- `go test ./ui`
- `go test ./...`
- `go build ./...`
- `openspec validate add-declarative-data-binding --strict`

## Done When

- Attribute/style bindings and commands work from XML.
- Expression bindings update when any dependency changes.
- Unsupported expressions fail safely without panics.

## Completion Evidence

- Added `ui/binding_expr.go`.
- Added tests for rich expressions, attribute/style binding, and command
  dispatch in `ui/binding_test.go`.
- Updated OpenSpec data-binding requirements and task checklist.
