# Bind Radio Options

## Why

Collection binding currently rebuilds dropdown/select options, but radio option
groups still require static XML or manual repeat templates. A panel-level radio
option binding gives authors a concise way to render a dynamic single-select
choice group from the same collection shape used by dropdowns.

## What Changes

- Allow `bind-options` on panel-like containers when `option-type="radio"`.
- Generate radio button children from primitive, map, or struct collections.
- Reuse `option-label` and `option-value` mappings.
- Preserve/restore selected value and support two-way `bind-value`.

## Impact

- Affects XML binding in `WidgetFactory`.
- Adds dynamic child creation for radio option containers.
