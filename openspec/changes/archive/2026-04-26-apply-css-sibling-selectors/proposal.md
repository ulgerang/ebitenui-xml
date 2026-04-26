# Apply CSS Sibling Selectors

## Why

Descendant and child selector rules now apply, but sibling combinators remain a
documented CSS parser gap. Adjacent and general sibling selectors are small to
support because the UI tree already preserves parent/child order.

## What Changes

- Apply adjacent sibling selector rules (`A + B`).
- Apply general sibling selector rules (`A ~ B`).
- Keep the scope limited to already-supported simple selector parts.

## Impact

- Affects `UI` complex selector matching.
- Adds regression coverage for sibling rule matching and non-matching nested
  nodes.
