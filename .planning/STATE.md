# GSD State

## Status

Follow-up phases 8-14 completed.

## Current Baseline

- XML common bindings implemented.
- XML repeat/list binding implemented.
- XML conditional rendering implemented.
- Rich binding expressions implemented.
- XML attribute/style bindings implemented.
- XML event command binding implemented.
- HTML semantic aliases implemented.
- DOM-like query helpers implemented.
- Form submit/reset, validation state, disabled fieldset propagation,
  XML radio grouping, and focus traversal implemented.
- Flex distribution, box sizing, min/max constraints, flex-wrap, and
  flex-shrink parity implemented.
- Absolute positioning, inset-derived sizing, overflow hit clipping, and
  z-index draw/input ordering implemented.
- Multi-shadow lists, blurred text shadows, per-corner frame rendering, and
  basic `clipPath` masking implemented.
- JSON custom keyframes and whole-widget compositing coverage for built-in
  custom-rendered widgets implemented.
- CSS visual testloop cases and documentation closeout completed.

## Follow-up Roadmap

- Phase 8: Dynamic collection binding.
- Phase 9: Form validation and messages.
- Phase 10: Keyboard and modal focus policies.
- Phase 11: Scroll overflow runtime.
- Phase 12: CSS keyframes syntax parser.
- Phase 13: Advanced visual fidelity.
- Phase 14: Visual compare hardening and release closeout.

## Completed Follow-ups

- Phase 8: Dynamic dropdown/select option binding with label/value mapping and
  binding diagnostics.
- Phase 9: XML form validation rules, invalid-submit blocking, and exposed
  validation messages.
- Phase 10: Dropdown/radio/slider keyboard policies and modal focus
  trap/restore.
- Phase 11: Runtime scroll offsets for overflow scroll/auto, wheel scrolling,
  and scrolled hit testing.
- Phase 12: Literal CSS `@keyframes` ingestion with structured parse errors.
- Phase 13: CSS `clip-path: polygon(...)` support and explicit deferrals for
  path clips, uniform-radius shadows, browser font fallback, and real table
  layout.
- Phase 14: Visual compare workflow hardening, support matrix publication, and
  OpenSpec archive reconciliation.
- Additional binding expression helpers implemented: `len`, `round`, `floor`,
  `ceil`, `contains`, `join`, and `format`.
- CSS simple selector declaration blocks implemented in `LoadCSS`, including
  common layout, spacing, visual, effect, overflow, and positioning properties.
- Nested modal focus stacks implemented so the deepest open dialog traps focus
  and Escape restores one modal level at a time.
- Basic table semantic flow implemented for table groups, rows, and flexible
  cells; full browser table algorithms remain deferred.
- CSS descendant and direct child selector rules now apply during style
  reapplication for CSS-loaded rule blocks.
- `LoadFromString` now auto-detects pure CSS rule strings while preserving JSON
  style loading.
- `UI.LoadCSS` and `UI.LoadStyles` now resolve `var(...)` references through
  the UI variable container before parsing.
- CSS adjacent sibling (`+`) and general sibling (`~`) selector rules now apply
  during style reapplication.
- CSS-loaded matching rules now preserve specificity/source-order cascade for
  same-specificity overlaps while retaining ID precedence.
- `UI.LoadCSS` now evaluates viewport `@media` blocks for `screen`/`all`,
  orientation, comma-separated query lists, and min/max width/height before CSS
  rule parsing.
- Keyboard handling now covers button Enter/Space activation, checkbox/toggle
  Space activation, radio Space selection, slider Home/End, and text-input
  Enter submit for the nearest form.
- Panel containers can now use `bind-options` with `option-type="radio"` to
  generate two-way bound radio option groups.
- Panel containers can now use `bind-options` with `option-type="checkbox"` to
  generate two-way bound checkbox option groups.
- CSS `filter: blur(...)` implemented.
- Declarative registered animation support implemented.
- Full-content compositing for text-like widgets implemented.
- CSS `clipPath: path(...)` implemented through the SVG path parser and the
  existing clip-mask compositing path.
- Box shadows now use per-corner radii in the GPU SDF shader while preserving
  blur, spread, and multi-shadow behavior.
- Text font resolution now supports registered font-family fallback lists via
  `UI.RegisterFontFace`, `UI.RegisterFontSource`, and configured defaults.
- CSS parsing now preserves quoted/function punctuation while splitting
  selector lists and declarations, supports declaration-level `!important`, and
  routes terminal state pseudo selectors into state styles.

## Validation Baseline

Last verified commands:

- `go test ./...`
- `go build ./...`
- `openspec validate add-declarative-data-binding --strict`
- `git diff --check`
- `go run ./cmd/css_testloop -mode render -out %TEMP%/ebitenui-css-phase7.png`
- `go run ./cmd/css_testloop -mode html -out %TEMP%/ebitenui-css-phase7.html`

Most recent phase-local verification:

- `go test ./...`
- `go build ./...`
- `go vet ./...`
- `openspec validate --all --strict`
- `go run ./cmd/css_testloop -mode render -out %TEMP%/ebitenui-css-phase14.png`
- `go run ./cmd/css_testloop -mode html -out %TEMP%/ebitenui-css-phase14.html`
- `go run ./cmd/css_testloop -mode compare -browser %TEMP%/ebitenui-css-phase14.png -ebiten %TEMP%/ebitenui-css-phase14.png -out %TEMP%/ebitenui-css-phase14-smoke-report.html`

## Notes

- `.omx/` is pre-existing untracked workspace output and is not part of this
  plan.
- This repo did not have `.planning/` before this GSD planning pass.
- Phase 23 started from `.planning/HTML-CSS-BINDING-REMAINING-PLAN.md`:
  OpenSpec change `audit-and-close-css-units-calc` records that lower-level
  `%`/`vw`/`vh`/`em`/`rem`/`calc()` utilities exist, while production CSS/XML
  style loading remains pixel-field based. Audit tests also fixed simple
  `calc(...)` operator tokenization in `ui/variables.go`.
