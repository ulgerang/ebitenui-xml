# Add Basic Table Flow

## Why

Table tags currently parse as semantic panel aliases, so rows and cells can
stack like normal panels unless the author writes explicit flex styles. A small
default layout layer makes common XML/HTML table markup useful without taking on
the full browser table algorithm.

## What Changes

- Apply default table semantic layout styles during XML creation.
- Lay out `table`/`thead`/`tbody`/`tfoot` as vertical groups.
- Lay out `tr` as horizontal rows and `td`/`th` as flexible cells.

## Impact

- Affects semantic HTML aliases only when authors do not provide explicit
  styles.
- Does not implement `rowspan`, `colspan`, intrinsic column sizing, or
  border-collapse.
