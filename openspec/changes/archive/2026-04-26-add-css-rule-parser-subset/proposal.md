# Add CSS Rule Parser Subset

## Why

The renderer already supports many CSS-like properties, but style authoring is
JSON-first except for literal `@keyframes`. A small CSS rule parser lets users
write familiar `.class { ... }` and `#id { ... }` rules for common properties
without adopting a full browser CSS engine.

## What Changes

- Extend `StyleEngine.LoadCSS` to parse simple selector blocks.
- Support comma-separated selectors and common declaration names.
- Keep JSON loading backward compatible.
- Continue treating unknown declarations as no-ops.

## Impact

- Affected spec: `css-effects`, `css-layout`
- Affected code: `ui/style.go`, tests/docs
