# Apply Complex CSS Selectors

## Why

The CSS rule loader can ingest selector blocks, but the runtime style pass still
only applies direct type, class, compound-class, and ID selectors. Common CSS
authoring patterns such as `.card .title` and `.toolbar > button` should work
for the implemented CSS rule subset.

## What Changes

- Apply descendant selectors (`A B`) during style reapplication.
- Apply direct child selectors (`A > B`) during style reapplication.
- Keep the implementation intentionally narrow; media queries, sibling
  combinators, and full browser cascade remain out of scope.

## Impact

- Affects style matching in `UI.reapplyStyles`.
- Adds tests proving CSS-loaded descendant and child rules affect matching
  widgets without overriding unrelated nodes.
