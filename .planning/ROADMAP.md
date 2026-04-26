# Roadmap

## Phase 1: Complete Data Binding

Goal: Finish the remaining high-value declarative binding features: rich
expressions, attribute/style binding, and event command binding.

Requirements:
- DB-EXPR: Template expressions support fallback, boolean operators, simple
  comparisons, arithmetic, and formatting helpers.
- DB-ATTR: XML can bind attributes and style properties without custom Go code.
- DB-EVENT: XML can dispatch event command names into a registered command map.

## Phase 2: HTML/XML Structure and Form Semantics

Goal: Add browser-familiar XML aliases, query helpers, form behavior, select/radio
grouping, and keyboard focus traversal.

Requirements:
- HTML-ALIASES: Add headings, list, table-like, form, nav, and section aliases.
- DOM-QUERY: Add class/type/subtree query helpers.
- FORM-SEM: Add submit/reset, validation state, fieldset disabled propagation.
- FORM-GROUPS: Support native-like select/radio grouping from XML.
- FOCUS-NAV: Add deterministic tab order and keyboard navigation policies.

## Phase 3: Flex and Sizing Layout Parity

Goal: Close flexbox and sizing gaps that affect most practical UI layouts.

Requirements:
- LAYOUT-JUSTIFY: Complete `space-between`, `space-around`, `space-evenly`.
- LAYOUT-SIZING: Implement `box-sizing: border-box` and min/max constraints.
- LAYOUT-WRAP: Implement `flex-wrap`.
- LAYOUT-SHRINK: Implement `flex-shrink` overflow reduction.

## Phase 4: Positioning, Overflow, and Z Ordering

Goal: Add browser-like absolute positioning, clipping, scrolling boundaries, and
stable z-order rendering/input.

Requirements:
- POS-ABS: Support `position: absolute` with `top/right/bottom/left`.
- OVERFLOW-HIDDEN: Clip children for `overflow: hidden`.
- Z-INDEX: Sort draw/input order by `z-index`.

## Phase 5: Borders, Shadows, and Clip Paths

Goal: Improve visual fidelity for common CSS frame/effect primitives.

Requirements:
- BORDER-SIDES: Render per-side border widths.
- RADIUS-CORNERS: Render per-corner border radii.
- SHADOW-LISTS: Support multi-shadow lists for `boxShadow` and `textShadow`.
- TEXT-SHADOW-BLUR: Render actual blurred text shadows.
- CLIP-PATH: Implement initial `clip-path` support.

## Phase 6: Keyframes, Transform Coverage, and Text Fidelity

Goal: Expand animation and typography fidelity after the layout/rendering
foundation is stable.

Requirements:
- KEYFRAMES: Parse and run CSS-like `@keyframes` declarations.
- TRANSFORM-COVERAGE: Apply transform/filter/opacity/animation to remaining
  custom widget types.
- TEXT-METRICS: Improve browser-like text metrics and font fallback behavior.

## Phase 7: Visual Regression and Documentation Hardening

Goal: Make the new HTML/CSS parity stable, documented, and visually checked.

Requirements:
- VISUAL-CASES: Add CSS testloop cases for each new visual/layout feature.
- DOCS: Update README/docs with supported XML/CSS/binding feature matrix.
- SPEC-CLOSEOUT: Reconcile and archive OpenSpec changes when implementation is
  complete.

## Phase 8: Dynamic Collection Binding

Goal: Bind complex widget collections, especially dropdown/select options, from
runtime data without imperative widget setup.

Requirements:
- DB-OPTIONS: Dropdown/select widgets can bind option collections from data.
- DB-OPTION-LABEL: Option label/value fields can be configured from item paths.
- DB-ERRORS: Binding expression failures are surfaced with actionable context.
- DB-HELPERS: Add high-value helpers only where implementation tests prove need.

## Phase 9: Form Validation and Messages

Goal: Move from validation state storage to browser-like form rule evaluation and
message rendering hooks.

Requirements:
- FORM-RULES: Support required, min/max, minLength/maxLength, pattern, and type
  validation for form-capable widgets.
- FORM-MESSAGES: Expose validation messages and allow XML/CSS-driven rendering.
- FORM-SUBMIT-BLOCK: Invalid forms can block submit command dispatch.
- FORM-TESTS: Cover validation edge cases with unit tests.

## Phase 10: Keyboard and Modal Focus Policies

Goal: Make keyboard interaction feel native for forms, menus, dropdowns, and
modals.

Requirements:
- KEY-ARROWS: Arrow keys navigate dropdown/radio/slider where appropriate.
- KEY-ENTER-ESC: Enter and Escape policies are consistent for forms, dropdowns,
  and modal dismissal.
- FOCUS-TRAP: Modal focus is trapped while open and restored on close.
- FOCUS-VISUAL: Focus state remains visible and testable.

## Phase 11: Scroll Overflow Runtime

Goal: Turn `overflow: scroll` / `auto` from clipping into real scrollable
containers.

Requirements:
- SCROLL-OFFSET: Widgets with scroll overflow maintain scroll offsets.
- SCROLL-INPUT: Mouse wheel and drag interactions update scroll offset.
- SCROLL-HIT: Hit testing maps through scroll offsets and clipped bounds.
- SCROLL-VISUAL: Scrollbar/thumb rendering is available or explicitly optional.

## Phase 12: CSS Keyframes Syntax Parser

Goal: Support literal CSS-like `@keyframes` syntax in addition to JSON
`keyframes`.

Requirements:
- CSS-KEYFRAMES: Parse `@keyframes name { from {...} 50% {...} to {...} }`.
- CSS-PROP-MAP: Map supported CSS declarations to existing keyframe properties.
- CSS-CASCADE: Decide whether literal CSS syntax enters through a new loader or
  preprocessor without breaking current JSON styles.
- CSS-ERRORS: Unsupported declarations fail as warnings/no-ops, not panics.

## Phase 13: Advanced Visual Fidelity

Goal: Improve browser parity for remaining complex visual details.

Requirements:
- CLIP-POLYGON: Implement `clip-path: polygon(...)`.
- CLIP-PATH: Evaluate `path(...)` support or document a deliberate no-op.
- SHADOW-RADIUS: Improve box-shadow geometry for per-corner radii.
- TEXT-FIDELITY: Improve baseline metrics and font fallback behavior.
- TABLE-LAYOUT: Add real table layout only if product usage justifies it.

## Phase 14: Visual Compare Hardening and Release Closeout

Goal: Turn the expanded feature set into reliable regression evidence and a
release-ready support matrix.

Requirements:
- COMPARE-RUNNER: Make CSS testloop compare easy to run end-to-end on Windows.
- SNAPSHOTS: Add deterministic cases for scroll, validation, dropdown binding,
  keyframes, and advanced visual effects.
- SUPPORT-MATRIX: Publish supported/partial/unsupported feature matrix.
- SPEC-ARCHIVE: Archive or split OpenSpec changes so completed and future work
  are no longer mixed.

## Phase 15: CSS Path Clip Fidelity

Goal: Implement `clip-path: path(...)` using existing SVG path primitives where
possible.

Requirements:
- CLIP-PATH-PATH: Parse CSS `path("...")` / `path('...')`.
- CLIP-PATH-MASK: Generate a clip mask from supported SVG path commands.
- CLIP-PATH-SAFE: Invalid path clips are no-op, not panics.
- GDC-SCOPE: Sync/check touched rendering and SVG helper nodes.

## Phase 16: Per-Corner Box Shadow Geometry

Goal: Make box shadows follow individual corner radii instead of uniform-radius
approximation.

Requirements:
- SHADOW-CORNERS: Use per-corner radius geometry for box-shadow paths.
- SHADOW-LISTS: Preserve multi-shadow list behavior.
- SHADOW-PERF: Keep temporary image allocation under control.
- GDC-SCOPE: Sync/check touched effects, widget, and style type nodes.

## Phase 17: Text Metrics And Font Fallback

Goal: Improve browser-like text metrics and fallback behavior within the
dependency-light renderer.

Requirements:
- FONT-FALLBACK: Parse and resolve font-family fallback lists.
- TEXT-METRICS: Normalize line-height and baseline behavior across text-like
  widgets.
- TEXT-DETERMINISM: Preserve deterministic fallback when fonts are unavailable.
- GDC-SCOPE: Sync/check touched text, input, and font cache nodes.

## Phase 18: Expanded CSS Media Queries

Goal: Extend current viewport media query support.

Requirements:
- MEDIA-ORIENTATION: Support `orientation: portrait|landscape`.
- MEDIA-TYPES: Support `screen` and `all`.
- MEDIA-LISTS: Support comma-separated query lists.
- GDC-SCOPE: Sync/check touched UI/CSS parser nodes.

## Phase 19: Additional Collection Binding

Goal: Extend collection binding beyond dropdown/select and generated radio
options.

Requirements:
- DB-CHECKBOX-OPTIONS: Generate checkbox groups from bound collections.
- DB-CHECKED-VALUES: Bind checked values to a selected value collection.
- DB-COLLECTION-DIAGNOSTICS: Report unsupported item shapes clearly.
- GDC-SCOPE: Sync/check touched binding and widget factory nodes.

## Phase 20: Browser Keyboard Edge Cases

Goal: Close practical browser-default keyboard behavior gaps.

Requirements:
- KEY-SPACE: Space activates focused button/checkbox/toggle/radio controls.
- KEY-HOME-END: Home/End behavior is defined for sliders and text controls.
- KEY-FORM-SUBMIT: Enter submit behavior works for form descendants.
- KEY-PRECEDENCE: Escape precedence across text input, dropdown, and modal is
  documented and tested.
- GDC-SCOPE: Sync/check touched UI/input/widget nodes.
