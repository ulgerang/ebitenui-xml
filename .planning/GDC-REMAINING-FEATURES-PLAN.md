# GDC Plan: Remaining HTML/CSS/Data Binding Parity

## Current GDC State

- `gdc.exe` is installed.
- This repo does not currently have `.gdc/`, so `gdc list` and `gdc check`
  fail until `gdc init` is run.
- OpenSpec remains the requirements source of truth. GDC is the structural
  touched-scope and drift check once initialized.

## Scope

Implement these remaining requested features:

1. `clip-path: path(...)`
2. Per-corner box-shadow geometry
3. Browser-like text metrics and font fallback
4. Expanded CSS media query support
5. Additional complex collection binding
6. Browser-default keyboard edge cases

Full browser table layout is intentionally excluded from this plan because the
user selected items 1, 2, 4, 5, 6, and 7 from the remaining list.

## Phase 0: GDC Bootstrap

Goal: Initialize GDC without replacing OpenSpec as the source of truth.

Tasks:

- Run `gdc init`.
- Sync existing Go code into initial graph nodes.
- Create or refine high-level nodes for:
  - `StyleEngine`
  - `UI`
  - `WidgetFactory`
  - `BaseWidget`
  - `LayoutEngine`
  - `Text`
  - `TextInput`
  - `TextArea`
  - `RadioGroup`
  - `Dropdown`
  - SVG/path helpers
- Run baseline checks:
  - `gdc sync --direction both --source ui --strategy merge`
  - `gdc check --verify-impl`
  - `openspec validate --all --strict`
  - `go test ./...`
  - `go build ./...`

GDC touched scope:

- `.gdc/**`
- `ui/**/*.go`
- `openspec/specs/**`
- `docs/HTML-CSS-BINDING-GAP-LIST.md`
- `docs/SUPPORT_MATRIX.md`

Exit criteria:

- `.gdc/` exists and `gdc check --verify-impl` has no new blocker errors.
- Any initial unrelated/orphan warnings are recorded but not treated as blockers.

## Phase 15: CSS `clip-path: path(...)`

Goal: Reuse existing SVG path parsing/masking primitives to support CSS
`clip-path: path(...)`.

OpenSpec target:

- `css-effects`
- Requirement: `Per-corner frame rendering and clip paths`

Implementation scope:

- `ui/effects.go`
- `ui/svg.go`
- `ui/widget.go`
- `ui/widgets.go`
- `ui/widgets_extended.go`
- `ui/effects_runtime_test.go`
- `cmd/css_testloop/testcases.go`

Tasks:

- Parse CSS `path("...")` and `path('...')`.
- Convert supported SVG path commands into a clipping mask.
- Apply the mask through the existing whole-widget compositing path.
- Treat invalid/unsupported path syntax as no-op with no panic.
- Add unit tests for parsing and mask generation.
- Add one CSS visual testloop case.

GDC checks:

- `gdc sync --direction both --files ui/effects.go,ui/svg.go,ui/widget.go,ui/widgets.go,ui/widgets_extended.go`
- `gdc check --verify-impl --orphan-filter cmd/`

Validation:

- `go test ./ui`
- `go test ./...`
- `go build ./...`
- `go vet ./...`
- `openspec validate <change-id> --strict`
- CSS testloop render/html smoke for the new clip-path case.

Risk:

- SVG path fill behavior may not match browser clipping exactly. First pass
  should document supported path commands and defer fill-rule complexity if
  needed.

## Phase 16: Per-Corner Box-Shadow Geometry

Goal: Make box shadows follow individual corner radii instead of the current
uniform radius approximation.

OpenSpec target:

- `css-effects`
- Requirement: `Shadow lists and blurred text shadows`
- Requirement: `Per-corner frame rendering and clip paths`

Implementation scope:

- `ui/effects.go`
- `ui/widget.go`
- `ui/types.go`
- `ui/effects_runtime_test.go`
- `cmd/css_testloop/testcases.go`

Tasks:

- Extend shadow rendering to use per-corner radius data.
- Support multi-shadow lists with per-corner geometry.
- Preserve current blur/spread behavior.
- Keep uniform fallback when only `borderRadius` is set.
- Add tests for per-corner shadow path construction.
- Add visual fixture comparing asymmetric radius shadow.

GDC checks:

- `gdc sync --direction both --files ui/effects.go,ui/widget.go,ui/types.go`
- `gdc check --verify-impl --category dependency`

Validation:

- `go test ./ui`
- `go test ./...`
- `go build ./...`
- `go vet ./...`
- CSS visual smoke render/html.

Risk:

- GPU/performance cost can grow if shadows allocate large temporary images.
  Reuse existing image pool paths where possible.

## Phase 17: Text Metrics And Font Fallback

Goal: Improve browser-like text measurement and fallback behavior while keeping
the renderer dependency-light.

OpenSpec target:

- `css-effects`
- Requirement: text fidelity extension under visual properties

Implementation scope:

- `ui/font_cache.go` or equivalent font cache files
- `ui/textwrap.go`
- `ui/widgets.go`
- `ui/input.go`
- `ui/types.go`
- `ui/effects_runtime_test.go`

Tasks:

- Audit current font cache and `text/v2` measurement paths.
- Add font-family fallback list parsing.
- Resolve first available font face from configured cache.
- Keep deterministic default when no fallback font is available.
- Normalize line-height/baseline behavior across `Text`, `Button`,
  `TextInput`, and `TextArea`.
- Add unit tests for fallback selection and line-height measurement.

GDC checks:

- `gdc sync --direction both --files ui/textwrap.go,ui/widgets.go,ui/input.go,ui/types.go`
- `gdc trace Text`
- `gdc check --verify-impl`

Validation:

- `go test ./ui`
- `go test ./...`
- `go build ./...`
- `go vet ./...`

Risk:

- True browser shaping is large. This phase should improve fallback and metric
  consistency, not claim full browser shaping parity unless proven by tests.

## Phase 18: Expanded CSS Media Queries

Goal: Extend existing viewport media query support beyond min/max width/height.

OpenSpec target:

- `css-layout`
- Requirement: `CSS rule subset for layout properties`

Implementation scope:

- `ui/ui.go`
- `ui/style.go`
- `ui/effects_runtime_test.go`

Tasks:

- Support `orientation: portrait|landscape`.
- Support optional media types: `screen`, `all`.
- Support comma-separated media query lists.
- Keep unsupported media types as non-matching no-op.
- Add tests for type, orientation, comma list, and existing min/max behavior.

GDC checks:

- `gdc sync --direction both --files ui/ui.go,ui/style.go`
- `gdc check --verify-impl --category dependency`

Validation:

- `go test ./ui`
- `go test ./...`
- `go build ./...`
- `go vet ./...`
- `openspec validate <change-id> --strict`

Risk:

- Full CSS media grammar is broad. Keep parser intentionally scoped and
  document unsupported expressions.

## Phase 19: Additional Collection Binding

Goal: Extend collection binding beyond dropdown/select and generated radio
options.

OpenSpec target:

- `data-binding`
- Requirement: `XML repeat/list binding`
- `browser-parity-followups` dynamic collection binding

Implementation scope:

- `ui/parser.go`
- `ui/binding.go`
- `ui/binding_expr.go`
- `ui/binding_test.go`
- `ui/widgets.go`
- `ui/widgets_extended.go`

Tasks:

- Add generated checkbox groups with `option-type="checkbox"`.
- Support binding selected values to `[]string` or `[]interface{}`.
- Preserve checked values across option refreshes.
- Add optional `option-id-prefix` to stabilize generated IDs if needed.
- Add diagnostics for unsupported collection item types.
- Add tests for primitive, map, and struct item collections.

GDC checks:

- `gdc sync --direction both --files ui/parser.go,ui/binding.go,ui/binding_expr.go,ui/widgets.go,ui/widgets_extended.go`
- `gdc trace WidgetFactory`
- `gdc check --verify-impl`

Validation:

- `go test ./ui`
- `go test ./...`
- `go build ./...`
- `go vet ./...`

Risk:

- Two-way binding of slices can cause update loops. Use guarded sync flags like
  the radio binding implementation.

## Phase 20: Browser Keyboard Edge Cases

Goal: Close practical browser-default keyboard behavior gaps for form-like
widgets.

OpenSpec target:

- `browser-parity-followups`
- Requirement: `Keyboard and modal focus policies`
- `html-structure`
- Requirement: `Form semantics`

Implementation scope:

- `ui/ui.go`
- `ui/input.go`
- `ui/widgets.go`
- `ui/widgets_extended.go`
- `ui/html_semantics_test.go`
- `ui/e2e_text_test.go`

Tasks:

- Define keyboard policy table before implementation.
- Add Space activation for buttons, checkbox, toggle, and focused radio.
- Add Home/End behavior for sliders and text inputs where appropriate.
- Add Enter submit behavior for form descendants.
- Add Escape behavior for text input/modal/dropdown precedence.
- Add disabled/hidden focus skip edge-case tests.
- Document any browser behaviors intentionally not implemented.

GDC checks:

- `gdc sync --direction both --files ui/ui.go,ui/input.go,ui/widgets.go,ui/widgets_extended.go`
- `gdc trace UI`
- `gdc check --verify-impl`

Validation:

- `go test ./ui`
- `go test ./...`
- `go build ./...`
- `go vet ./...`

Risk:

- Keyboard behavior can conflict between focused text fields and global UI
  shortcuts. Tests should define precedence before code changes.

## Recommended Execution Order

1. Phase 0: GDC Bootstrap
2. Phase 18: Expanded CSS Media Queries
3. Phase 19: Additional Collection Binding
4. Phase 20: Browser Keyboard Edge Cases
5. Phase 15: CSS `clip-path: path(...)`
6. Phase 16: Per-Corner Box-Shadow Geometry
7. Phase 17: Text Metrics And Font Fallback

Reasoning:

- Phases 18-20 are lower-risk and mostly parser/binding/input logic.
- Phases 15-17 touch rendering fidelity and need more visual proof.
- Text/font work is last because its correctness depends on clearer visual
  regression expectations.

## Per-Phase Required Closeout

Each phase must finish with:

- OpenSpec change archived.
- GDC touched scope synced and checked, if `.gdc/` is initialized.
- `go test ./...`
- `go build ./...`
- `go vet ./...`
- `openspec validate --all --strict`
- Docs updated:
  - `docs/HTML-CSS-BINDING-GAP-LIST.md`
  - `docs/SUPPORT_MATRIX.md`
  - `.planning/STATE.md`

For visual phases, also run:

- `go run ./cmd/css_testloop -mode render -out %TEMP%/<case>.png`
- `go run ./cmd/css_testloop -mode html -out %TEMP%/<case>.html`
