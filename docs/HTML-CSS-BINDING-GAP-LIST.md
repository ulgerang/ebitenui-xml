# HTML/CSS/Data Binding Gap List

This project aims to bring familiar HTML structure and CSS-like styling into
Ebiten through XML layouts and JSON styles. The current implementation already
covers core widgets, flex layout, selectors, CSS variables, SVG, input widgets,
and imperative data binding. The gaps below are the implementation backlog.

## Implemented in Current Slice

- XML-declared common data bindings:
  - `bind-text` / `data-bind-text`
  - `bind-value` / `data-bind-value`
  - `bind-checked` / `data-bind-checked`
  - `bind-visible` / `data-bind-visible`
  - `bind-enabled` / `data-bind-enabled`
- Text template interpolation for XML text-like content:
  - `Player {{player.name}} Lv.{{player.level}}`
- XML repeat/list binding:
  - `bind-repeat`, `data-bind-repeat`, and `for-each` repeat an XML template
    for collection items.
  - Repeated templates support `{{item}}`, `{{item.field}}`, and `{{index}}`
    in text and attributes.
  - Repeated children are rebuilt when the bound collection changes, including
    the runtime widget ID cache.
- XML conditional rendering:
  - `bind-if` and `data-bind-if` attach or detach an XML template from the
    widget tree from a bound truthy/falsy value.
  - Conditional children update the runtime widget ID cache when added or
    removed.
- Rich binding expressions:
  - Templates and one-way bindings support fallback (`||`), boolean operators,
    comparisons, simple arithmetic, grouping, and helpers such as `upper()`,
    `lower()`, `default()`, `number()`, `len()`, `round()`, `floor()`,
    `ceil()`, `contains()`, `join()`, and `format()`.
- XML attribute and style bindings:
  - `bind-attr-*` / `data-bind-attr-*` update common widget attributes such as
    label, content, placeholder, value, checked, disabled, visible, and size
    attributes.
  - `bind-style-*` / `data-bind-style-*` update safe style fields such as
    color, background, border, opacity, display, visibility, transform, filter,
    and animation.
- XML event command binding:
  - `onClick`, `onChange`, and `onSubmit` dispatch to handlers registered with
    `UI.RegisterCommand`.
- Dynamic collection binding:
  - `bind-options` / `data-bind-options` rebuild dropdown/select options from
    bound collections.
  - Panel containers with `option-type="radio"` use `bind-options` to generate
    radio button children from bound collections.
  - Panel containers with `option-type="checkbox"` use `bind-options` to
    generate checkbox children from bound collections.
  - `option-label` and `option-value` map struct/map item fields to option
    labels and values.
  - Selected dropdown values are preserved across option refreshes when the
    value still exists.
  - Generated radio groups preserve selection across refreshes and can update
    a `bind-value` target when a generated radio button is clicked.
  - Generated checkbox groups preserve checked values across refreshes and
    update a `bind-value` target with all checked values.
  - Binding diagnostics record non-fatal expression failures with widget and
    attribute context.
- Form validation rules and messages:
  - XML form fields parse `required`, `min`, `max`, `minlength`, `maxlength`,
    `pattern`, `type`, and `validation-message` attributes.
  - Text inputs and textareas validate length, regex pattern, email, and number
    constraints.
  - Checkbox/toggle/radio required state, dropdown required selection, and
    slider numeric ranges are evaluated.
  - `SubmitForm` validates descendants before command dispatch and blocks
    invalid submit handlers.
  - `ValidateForm`, `GetValidationState`, and `GetValidationMessage` expose
    validation results for XML-rendered message widgets.
- Keyboard and modal focus policies:
  - Dropdown/select widgets open, move highlighted options with arrows, select
    with Enter, and close with Escape.
  - Buttons activate with Enter or Space when focused.
  - Checkboxes, toggles, and focused radio buttons activate with Space.
  - Radio groups move selection with arrow keys from the focused radio.
  - Sliders support keyboard stepping through XML `step` and jump to min/max
    with Home/End.
  - Focused single-line text inputs submit the nearest valid form with Enter.
  - Open modals trap Tab focus inside their descendants and restore the
    previous focus target when closed with Escape.
  - Nested modals use the deepest open dialog as the active focus trap and
    restore focus one level at a time.
- Scroll overflow runtime:
  - `overflow: scroll` and `overflow: auto` widgets keep runtime scroll offsets
    and content extents.
  - `UI.SetWidgetScroll` and `UI.ScrollWidgetBy` expose deterministic scroll
    control for tests and application code.
  - Mouse wheel input scrolls the hovered overflow container.
  - Child drawing and hit testing are mapped through the scroll offset while
    still clipping to the visible content box.
- CSS keyframes syntax:
  - `StyleEngine.LoadCSS` and `UI.LoadCSS` ingest literal `@keyframes` blocks.
  - `LoadFromString` also detects non-JSON strings containing `@keyframes`.
  - `from`, `to`, and percentage selectors map to registered animations.
  - Supported declarations include `opacity`, `transform`, `width`, `height`,
    background color, border color, and simple box-shadow blur/spread fields.
  - Unsupported declarations are ignored; malformed keyframe blocks return
    structured errors.
- CSS rule syntax subset:
  - `StyleEngine.LoadCSS` parses simple selector blocks such as `.card { ... }`
    and `#title { ... }`.
  - `LoadFromString` auto-detects non-JSON CSS rule strings and routes them to
    the CSS parser.
  - `UI.LoadCSS` and `UI.LoadStyles` resolve `var(...)` references from
    `UI.SetVariable` before parsing, including fallback values.
  - Comma-separated selectors are supported.
  - Descendant selectors such as `.card .title` and direct child selectors
    such as `.toolbar > button` apply during UI style reapplication.
  - Sibling selectors such as `.label + .value` and `.label ~ .hint` apply to
    matching following siblings within the same parent.
  - Matching CSS rules are applied by specificity and source order so later
    same-specificity rules win for overlapping fields.
  - Declaration-level `!important` is applied as a later cascade layer over
    normal type/class/ID rules.
  - Terminal state pseudo selectors for `:hover`, `:active`, `:focus`, and
    `:disabled` populate state styles instead of changing the base style.
  - CSS declaration and selector-list splitting preserves separators inside
    quoted strings and function parentheses.
  - `UI.LoadCSS` evaluates top-level `@media` blocks for `screen`/`all`,
    `orientation`, comma-separated query lists, `min-width`, `max-width`,
    `min-height`, and `max-height` against the current UI size.
  - Common layout, spacing, color, border, typography, transform, filter,
    animation, overflow, and positioning declarations map to existing `Style`
    fields.
  - Unknown declarations are ignored.
- HTML/XML structure and form semantics:
  - Semantic aliases for headings, lists, table-like nodes, forms, fieldsets,
    and landmarks parse into existing widgets with semantic metadata/classes.
  - Table groups default to column flow, rows default to row flow, and cells
    default to flexible panel cells for simple table-like layouts.
  - `QueryByClass`, `QueryByType`, and `Query` support simple DOM-like lookup.
  - Form submit/reset helpers dispatch registered commands and reset field
    values to their initial XML state.
  - `fieldset disabled="true"` propagates disabled state to descendants.
  - Radio buttons group by XML `name`.
  - `tabindex` and `FocusNext` provide deterministic focus traversal.
- CSS flex and sizing layout parity:
  - `space-between`, `space-around`, and `space-evenly` distribution are
    covered by layout tests.
  - `boxSizing` supports `content-box` and `border-box` layout sizing.
  - Min/max constraints are enforced during flex layout.
  - `flex-wrap` places overflowing children onto additional lines.
  - `flex-shrink` respects zero-shrink children and min-size constraints.
- CSS positioning and z-ordering:
  - `position: absolute` supports `top`, `right`, `bottom`, and `left` insets.
  - Absolute children no longer consume flex layout space for normal siblings.
  - Opposing insets can derive width or height when no explicit size is set.
  - `overflow: hidden`, `scroll`, and `auto` clip child drawing and hit testing.
  - `zIndex` sorts draw order and reverse hit testing order for overlapping
    siblings.
- CSS effect runtime improvements:
  - `filter: blur(Npx)` now renders through a two-pass Gaussian blur.
  - `transform`, `filter`, `opacity`, and `animation` compose full content for
    text-like widgets (`Text`, `Button`, `TextInput`, `TextArea`).
  - CSS transitions now start on widget state changes.
  - JSON styles can declare registered animations with `animation`, for example
    `"pulse 750ms ease-in-out infinite"`.
  - Top-level JSON `keyframes` blocks can register custom named animations with
    `from` / `to` / percentage frames.
  - Built-in custom-rendered widgets now route transform, filter, opacity,
    animation, and clip-path through whole-widget offscreen compositing.
- CSS frame/shadow improvements:
  - `boxShadow` and `textShadow` accept comma-separated shadow lists while
    preserving commas inside color functions such as `rgba(...)`.
  - Blurred `textShadow` values render through an offscreen Gaussian blur pass.
  - Individual corner radii are used for backgrounds, gradients, and uniform
    borders.
  - `clipPath` supports `inset(...)`, `circle(...)`, `polygon(...)`, and
    quoted `path(...)` values backed by the SVG path parser.
  - Per-corner frame geometry is supported for backgrounds, borders, gradients,
    and box shadows.
  - Text rendering resolves comma-separated `fontFamily` lists through
    explicitly registered UI font faces or font sources before falling back to
    the configured default face/source. Full OS/browser font discovery remains
    out of scope.
- CSS relative unit utilities:
  - Go callers can parse and resolve `%`, `vw`, `vh`, `em`, `rem`, px/unitless
    values, and simple `calc(...)` expressions through `ParseSizeValue`,
    `SizeValue.Resolve`, `ParseCalc`, and `CalcExpression.Resolve` with an
    explicit `SizeContext`.

## Data Binding Gaps

1. Additional complex widget collection binding beyond dropdown/select and
   generated radio/checkbox options.
2. Domain-specific expression helpers beyond the current general-purpose set.

## HTML/XML Structure Gaps

1. Browser-level table layout semantics beyond basic row/cell flow.
2. More browser-default keyboard edge cases beyond the documented controls.

## CSS Visual/Effect Gaps

1. Full OS/browser font discovery and shaping fallback behavior.
2. Full CSS parser semantics beyond the current selector, media, state pseudo,
   and `!important` subset.
3. Production style/layout loading does not yet live-resolve `%`, `vw`, `vh`,
   `em`, `rem`, or `calc()` values from CSS rules, XML inline styles, binding
   style updates, or keyframe layout fields. Current production loaders store
   numeric pixel-style values on `Style`, while relative unit resolution remains
   a lower-level utility.

## Deferred HTML Layout Scope

- Table-like tags (`table`, `thead`, `tbody`, `tr`, `td`, `th`) use semantic
  panel widgets with basic group/row/cell flex defaults. Real browser table
  layout is deferred until project usage requires row/column spanning and
  intrinsic table measurement.

## Recommended Implementation Order

1. Finish high-impact flex parity: justify distribution, min/max constraints,
   box sizing, then flex-wrap/shrink.
2. Finish visual parity: multi-shadow, text-shadow blur, custom keyframes, and
   clip-path.
3. Add collection binding for richer generated widget sets beyond dropdown,
   select, radio, and checkbox options as product needs appear.
