## Why

The current CSS subset handles practical selector blocks, but a few browser
cascade/parser edge cases can produce surprising results: pseudo-class rules can
act like normal rules, `!important` is ignored, and splitting declarations or
selector lists can break on punctuation inside functions or quoted strings.

## What Changes

- Parse selector lists and declarations while respecting quotes and parentheses.
- Support `!important` declarations as a later cascade layer.
- Route terminal `:hover`, `:active`, `:focus`, and `:disabled` rules into the
  corresponding state styles instead of applying them as normal base rules.
- Preserve existing source-order and specificity behavior for normal rules.

## Impact

- Affects literal CSS rule parsing and style reapplication.
- Does not implement full browser selector pseudo-classes such as `:has()`.
