# Detect CSS Rule Strings

## Why

`LoadCSS` can parse rule blocks, but `LoadFromString` only routes non-JSON input
to CSS parsing when `@keyframes` is present. A stylesheet containing only
`.card { ... }` should be accepted through the generic string loading path.

## What Changes

- Detect non-JSON strings that look like CSS declaration blocks.
- Route those strings through `LoadCSS`.
- Keep JSON object strings on the existing JSON loader path.

## Impact

- Affects `StyleEngine.LoadFromString`.
- Adds regression coverage for pure CSS rule strings.
