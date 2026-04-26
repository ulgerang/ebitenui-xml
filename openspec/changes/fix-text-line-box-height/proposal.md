## Why

Korean text can overlap in column layouts when a `Text` widget's intrinsic
height is measured from glyph bounds instead of the line box used during draw.
The mismatch is visible in Tower100Y month-start sections such as consecutive
resource status labels.

## What Changes

- Align `Text.IntrinsicHeight` with the same line-height calculation used by
  text layout and drawing.
- Respect explicit `lineHeight` style values during intrinsic height
  calculation.
- Keep button label intrinsic height consistent with the same font metrics.
- Re-run layout after declarative binding updates so text that starts empty and
  later receives Korean content gets a fresh intrinsic height.
- Exclude invisible children from layout sizing and gaps so hidden bound text
  cannot shrink visible text rows.
- Add regression coverage for Korean text in a column layout.

## Impact

- Affects automatic sizing for text-like widgets without explicit heights.
- Existing explicit widget heights remain authoritative.
