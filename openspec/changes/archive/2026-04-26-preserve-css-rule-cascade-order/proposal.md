# Preserve CSS Rule Cascade Order

## Why

CSS rule loading currently stores styles in a selector map. That is fine for
direct selector lookup, but it does not preserve source order when multiple
same-specificity selectors match the same widget. Browser CSS expects later
rules with equal specificity to win.

## What Changes

- Track loaded style rules in source order.
- Apply matching style rules by specificity and source order during UI style
  reapplication.
- Keep JSON map ordering unspecified; this change targets CSS/loader rule order.

## Impact

- Affects style cascade in `UI.reapplyStyles`.
- Adds regression coverage for same-specificity class selector ordering.
