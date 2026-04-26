## Why

The public docs mention relative CSS units and `calc()`, and `ui/variables.go`
contains lower-level parsers for `%`, `vw`, `vh`, `em`, `rem`, and `calc()`.
The current production style loading paths still parse common layout values
through pixel-only helpers, so the actual runtime support is ambiguous.

This change audits and locks down the current truth before expanding table/grid
layout work that depends on predictable sizing semantics.

## What Changes

- Document the supported lower-level unit parsing utilities.
- Document the current CSS/XML style loader boundary: layout style properties
  are resolved as numeric pixels, not live browser-relative units.
- Add regression tests that prove the current audit result.
- Update the support matrix and gap list with exact supported/partial status.

## Impact

- Affects OpenSpec docs, project support docs, and audit tests.
- Does not introduce broad runtime layout changes in the audit slice.
- Leaves a small, explicit follow-up path for wiring relative units into
  `LoadCSS`, XML inline styles, binding style updates, and layout resolution.
