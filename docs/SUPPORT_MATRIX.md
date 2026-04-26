# HTML/CSS/Data Binding Support Matrix

This matrix records the browser-parity surface implemented by the current GSD
milestone.

## Supported

| Area | Support |
| --- | --- |
| XML structure | HTML-like aliases for headings, lists, landmarks, forms, fieldsets, and table-like tags with basic row/cell flow |
| Data binding | `bind-text`, `bind-value`, `bind-checked`, `bind-visible`, `bind-enabled`, `bind-repeat`, `bind-if`, template interpolation, attribute/style bindings, command events, and general expression helpers |
| Collection binding | `bind-options` for dropdown/select plus panel `option-type="radio"` and `option-type="checkbox"` groups with `option-label` and `option-value` mappings |
| Forms | Submit/reset helpers, validation state, validation messages, required/min/max/minlength/maxlength/pattern/type rules |
| Keyboard | Tab traversal, button Enter/Space, checkbox/toggle/radio Space, dropdown arrows/Enter/Escape, radio arrows, slider arrows/Home/End, text-input Enter form submit, modal focus trap/restore including nested modals |
| Layout | Flex direction, gap, justify distribution, align, box sizing, min/max, wrap, shrink, absolute positioning, z-index |
| Overflow | `hidden`, `scroll`, and `auto` clipping; runtime scroll offsets; wheel scrolling; scrolled hit testing |
| Visuals | Backgrounds, gradients, borders, per-corner radius backgrounds/borders/box shadows, multi box shadows, blurred text shadows |
| Effects | Opacity, transform, filter blur, backdrop filter, transitions, JSON keyframes, literal CSS `@keyframes`, and simple CSS rule blocks |
| CSS variables | `UI.SetVariable` with `var(...)` resolution during `UI.LoadCSS` and `UI.LoadStyles`, including fallbacks |
| CSS unit utilities | Go-level `ParseSizeValue`, `SizeValue.Resolve`, `ParseCalc`, and `CalcExpression.Resolve` support `%`, `vw`, `vh`, `em`, `rem`, px/unitless values, and simple `calc(...)` expressions when supplied with an explicit `SizeContext` |
| Clip path | `inset(...)`, `circle(...)`, `polygon(...)`, quoted `path(...)` |
| Font family | Explicit `UI.RegisterFontFace` / `UI.RegisterFontSource` family lookup with comma-list fallback to configured defaults |

## Partial

| Area | Current boundary |
| --- | --- |
| Table layout | Basic table/section column flow, row flow, and flexible cells; no intrinsic browser table algorithm |
| Text metrics | Uses Ebiten `text/v2` metrics and configured font caches, not OS/browser shaping fallback |
| CSS syntax | Simple selector declaration blocks are accepted through `LoadCSS` and non-JSON `LoadFromString`; descendant, child, adjacent sibling, and general sibling selectors are supported; matching rules apply by specificity/source order; declaration-level `!important` is supported; terminal `:hover`, `:active`, `:focus`, and `:disabled` map to state styles; viewport `@media` supports `screen`/`all`, orientation, comma lists, and min/max width/height |
| CSS relative units in production style loading | Lower-level unit/calc resolvers exist, but `LoadCSS`, XML inline styles, binding style updates, keyframe width/height parsing, and `LayoutEngine` currently consume numeric `Style` fields through pixel-based parsing; `%`, `vw`, `vh`, `em`, `rem`, and `calc()` are not live-resolved against parent, viewport, or font context in layout |

## Intentionally Unsupported For This Milestone

| Feature | Reason |
| --- | --- |
| Full browser table algorithm | Requires intrinsic column measurement, row/column spans, and border-collapse semantics |
| Browser font fallback stack | OS font discovery and shaping fallback remain outside the current dependency-light renderer |

## Future Work

- Expand CSS syntax only when product usage needs selectors or at-rules beyond
  the current selector, state pseudo, important, and media subset.
- Promote basic table flow into a true table layout engine when product usage
  requires `rowspan`, `colspan`, and intrinsic column sizing.
- Add richer visual compare automation for real browser screenshot capture.
- Wire relative units and `calc()` through production CSS/XML style loading and
  layout resolution before claiming browser-like runtime unit parity.
