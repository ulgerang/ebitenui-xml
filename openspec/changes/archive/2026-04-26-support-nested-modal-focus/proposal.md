# Support Nested Modal Focus

## Why

The current modal focus policy traps and restores focus for a single open modal.
Nested dialogs need the same deterministic behavior: the deepest open modal is
active, Escape closes only that modal, and focus returns through the modal stack.

## What Changes

- Treat the deepest/topmost open modal as the active modal.
- Track modal focus restore targets as a stack.
- Restore focus one level at a time when nested modals close.

## Impact

- Affects `ui.UI` modal focus state and keyboard handling.
- Adds regression coverage for nested dialog traversal and Escape behavior.
