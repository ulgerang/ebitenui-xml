## Why

The UI runtime already supports core focus traversal and several control keys,
but common browser-style keyboard activation still has gaps for form controls.
Closing the practical cases makes XML-authored forms usable without pointer
input.

## What Changes

- Add Space activation for focused button, checkbox, toggle, and radio widgets.
- Add Home/End slider movement to min/max.
- Submit the nearest form when Enter is pressed from a focused single-line text
  input.
- Include Space/Home/End in runtime keyboard dispatch.

## Impact

- Affects UI keyboard dispatch only.
- Adds regression coverage through simulated key presses.
