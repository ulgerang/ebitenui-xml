# Support Basic CSS Media Queries

## Why

The CSS rule subset now covers common selector/cascade behavior, but responsive
styles still cannot be scoped to viewport size. A small `@media` evaluator makes
layout rules usable across UI sizes without implementing the full media query
language.

## What Changes

- Evaluate simple `@media` blocks in `UI.LoadCSS`.
- Support `min-width`, `max-width`, `min-height`, and `max-height` conditions.
- Inline matching media blocks before sending CSS to the existing style parser.
- Drop non-matching media blocks.

## Impact

- Affects UI-level CSS loading only because viewport size belongs to `UI`.
- Does not support media types, orientation, comma-separated query lists, or
  dynamic live re-evaluation after resize.
