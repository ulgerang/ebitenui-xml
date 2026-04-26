# Extend Binding Expression Helpers

## Why

XML binding expressions already support fallback, arithmetic, boolean logic, and
basic formatting helpers. Common UI text still needs small helper functions for
collection counts, numeric rounding, substring checks, joins, and simple string
formatting without custom Go glue code.

## What Changes

- Add safe binding helpers:
  - `len(value)`
  - `round(value[, digits])`
  - `floor(value)`
  - `ceil(value)`
  - `contains(value, needle)`
  - `join(value, separator)`
  - `format(template, args...)`
- Keep unsupported or invalid helper calls as failed expression evaluations.
- Add unit coverage for helper behavior inside XML template expressions.

## Impact

- Affected spec: `data-binding`
- Affected code: `ui/binding_expr.go`, `ui/binding_test.go`
