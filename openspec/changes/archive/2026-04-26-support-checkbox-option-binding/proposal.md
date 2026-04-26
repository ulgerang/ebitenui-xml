## Why

`bind-options` currently covers dropdown/select and generated radio groups.
Multi-select option lists still require manual XML templates and custom Go glue.
Checkbox option groups close the common collection-binding gap without adding a
new widget abstraction.

## What Changes

- Add `option-type="checkbox"` support on panel-like containers.
- Generate checkbox children from primitive, map, or struct collections using
  `option-label` and `option-value` mappings.
- Bind checked values through `bind-value`, accepting existing `[]string` or
  `[]interface{}` values and writing user changes back as `[]string`.
- Add optional `option-id-prefix` for stable generated IDs.

## Impact

- Affects XML widget factory binding logic.
- Adds regression coverage for checkbox option generation and two-way binding.
