# Resolve CSS Variables On Load

## Why

The UI exposes CSS variables through `SetVariable`, but style strings are parsed
before those variables are resolved. Color and effect declarations containing
`var(--name)` should resolve through the UI loading path before parsing.

## What Changes

- Resolve `var(...)` references in `UI.LoadStyles`.
- Resolve `var(...)` references in `UI.LoadCSS`.
- Preserve fallback values for undefined variables through the existing
  `CSSVariables.Resolve` behavior.

## Impact

- Affects UI-level style loading only.
- Does not add live re-resolution when variables change after styles are loaded.
