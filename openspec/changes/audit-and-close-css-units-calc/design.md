## Audit Findings

Current implementation has two layers:

1. Lower-level utilities in `ui/variables.go`
   - `ParseSizeValue` recognizes numeric px, `%`, `vw`, `vh`, `em`, `rem`, and
     `auto`.
   - `SizeValue.Resolve(SizeContext)` resolves those units against a caller
     supplied parent size, viewport size, font size, and root font size.
   - `ParseCalc` tokenizes `calc(...)` expressions using the same units and
     resolves them with `CalcExpression.Resolve`.
   - The audit fixed tokenizer handling for `+`, `-`, `*`, and `/` so simple
     mixed-unit expressions such as `calc(50% - 10px)` resolve correctly.

2. Production style/layout loading paths
   - CSS rule declarations in `StyleEngine.LoadCSS` use `parseCSSPixels` for
     layout, spacing, typography, transform pixel components, and keyframe
     width/height fields.
   - XML inline style attributes use `parseSize`, which strips `px` and `%`
     suffixes and stores the remaining number.
   - `LayoutEngine` consumes already-resolved `float64` fields on `Style`;
     it does not preserve raw unit strings or re-resolve live relative units.

## Decision

This change records the audit as partial support rather than claiming browser
relative unit parity.

Supported now:

- Direct use of `ParseSizeValue` / `SizeValue.Resolve` by Go callers.
- Direct use of `ParseCalc` / `CalcExpression.Resolve` by Go callers.
- Pixel values and unitless numeric values through JSON, XML inline styles, and
  CSS rule declarations.
- Existing CSS variable substitution before CSS parsing.

Partial / not wired now:

- `%`, `vw`, `vh`, `em`, `rem`, and `calc()` in CSS rule declarations for
  layout properties are not live-resolved into `Style` fields.
- XML inline `%` values are currently stripped to their numeric value, not
  resolved against parent size.
- Relative units are not preserved as raw values through `Style` and therefore
  do not re-resolve when viewport, parent size, or font size changes.

## Follow-up Implementation Boundary

Future runtime closure should add raw unit storage or a dedicated resolved style
layer before changing layout behavior. The implementation should start with
width/height/min/max/gap/padding/margin/font-size/line-height and should avoid
changing existing px behavior.
