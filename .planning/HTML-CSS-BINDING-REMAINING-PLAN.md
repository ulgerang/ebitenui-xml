# Remaining HTML/CSS/Data Binding Plan

Date: 2026-04-26
Scope: planning only. No production code changes.

## Baseline

OpenSpec is the source of truth for intent and requirements. `openspec list --json`
currently reports no active changes. Existing specs already cover the completed
HTML structure, CSS layout/effects, data binding, and browser parity follow-up
work.

Current evidence used:

- `openspec/specs/html-structure/spec.md`
- `openspec/specs/css-layout/spec.md`
- `openspec/specs/css-effects/spec.md`
- `openspec/specs/data-binding/spec.md`
- `openspec/specs/browser-parity-followups/spec.md`
- `docs/HTML-CSS-BINDING-GAP-LIST.md`
- `docs/SUPPORT_MATRIX.md`
- `.planning/STATE.md`
- `ui/parser.go`, `ui/binding.go`, `ui/binding_expr.go`, `ui/style.go`,
  `ui/selector.go`, `ui/types.go`, `ui/variables.go`, `ui/layout.go`,
  `ui/widget.go`
- Unit/runtime coverage in `ui/*_test.go`

Do not use `docs/CSS-GAP-ANALYSIS.md` as the primary source. It contains stale
or encoding-corrupted content compared with the support matrix, OpenSpec archive,
current tests, and code.

## Already Implemented, Do Not Re-plan

- Declarative binding: `bind-text`, `bind-value`, `bind-checked`,
  `bind-visible`, `bind-enabled`, `bind-repeat`, `bind-if`, template
  interpolation, `bind-attr-*`, `bind-style-*`, command events, dropdown/select
  options, generated radio options, generated checkbox options, option label and
  value mapping, selected value preservation, binding diagnostics, and helpers
  including `upper`, `lower`, `default`, `number`, `len`, `round`, `floor`,
  `ceil`, `contains`, `join`, and `format`.
- HTML/XML semantics: headings, lists, landmarks, forms, fieldsets, basic
  table-like tags, DOM-like query helpers, submit/reset, validation state and
  messages, disabled fieldset propagation, XML radio grouping, and tabindex
  traversal.
- CSS layout/effects: flex distribution, gap, alignment, wrap, shrink, min/max,
  `boxSizing`, absolute positioning, overflow clipping and scroll offsets,
  z-ordering, selectors through descendant/child/sibling combinators,
  source-order cascade, declaration-level `!important`, terminal hover/active/
  focus/disabled pseudo states, viewport media subset, variables, gradients,
  shadows, per-corner radius/frame/shadow geometry, `clipPath` inset/circle/
  polygon/path, transforms, blur filters, backdrop filters, opacity,
  transitions, JSON keyframes, literal CSS `@keyframes`, and explicit
  registered font fallback.

## Prioritized Missing / Partial Feature List

| Priority | Gap | Evidence basis | Why it matters | Boundary |
| --- | --- | --- | --- | --- |
| P0 | Browser-level table layout | `SUPPORT_MATRIX.md` marks table layout partial; `html_semantics_test.go` covers basic semantic flow only; no evidence of `rowspan`, `colspan`, `border-collapse`, or `table-layout` algorithms | Tables are a visible HTML parity gap and cannot be approximated well by flex once spans/intrinsic columns are needed | Implement practical table layout, not the full HTML table spec |
| P1 | CSS Grid layout | `rg` finds no grid layout support in `ui/layout.go` or specs; current layout spec is flex/positioning focused | Modern HTML/CSS layouts often expect grid for dashboards, forms, and fixed two-dimensional composition | Implement a scoped grid subset before advanced browser features |
| P1 | CSS units/calc runtime audit and closure | Docs mention `%`, `vw`, `vh`, `em`, `rem`, and `calc()`; code/GDC show calc-related support, but support matrix does not explicitly guarantee browser-like runtime behavior | Unit behavior affects nearly every layout feature and should be proven before larger layout expansion | Audit first, then implement only missing or partial behavior |
| P2 | Declarative class/state/metadata binding | Existing bindings cover text/value/checked/visible/enabled/attr/style/options, but no clear `bind-class`, `bind-role`, `bind-aria-*`, or semantic metadata binding surface | State-driven classes and semantic metadata are common in data-driven UI authoring | Metadata is product-level Ebiten metadata, not true browser accessibility |
| P2 | Declarative event surface | Current command bindings cover `onClick`, `onChange`, and `onSubmit`; no clear `onInput`, `onFocus`, `onBlur`, `onKeyDown`, event args, or prevent/stop policy | Forms and interactive widgets need richer declarative behavior without custom Go glue | Keep the event payload small and deterministic |
| P2 | High-value selector/parser subset | Current CSS parser supports useful selectors and media subset; OpenSpec/archive explicitly avoid full selector semantics such as `:has()` | Some selectors such as `:not()` and `:first-child` deliver high value without cloning browsers | Do not implement full CSS cascade origins/layers/nesting/supports/container queries unless separately justified |
| P3 | Visual regression runner automation | `css_testloop` exists, but support matrix still lists richer browser screenshot capture as future work | Larger layout/effect features need repeatable visual proof, not ad hoc screenshot steps | Improve tooling around existing testloop; avoid building a full browser farm |
| P3 | Browser font discovery and shaping fallback | `SUPPORT_MATRIX.md` intentionally excludes OS/browser fallback stack; current code supports explicit registered fallback | Full shaping/fallback is a major renderer concern | Keep out of scope unless a real product case requires it |

## Recommended OpenSpec Change Breakdown

Create separate OpenSpec changes instead of one broad "browser parity" change.
This keeps implementation and validation small enough for GSD execution.

1. `add-table-layout-engine`
   - Specs: `html-structure`, `css-layout`
   - Requirements:
     - table cells can declare `rowspan` and `colspan`
     - table columns are measured from intrinsic/min/max cell content
     - `table-layout: auto` and `table-layout: fixed` are supported as scoped modes
     - `border-collapse: collapse|separate` has documented rendering behavior
   - Explicit non-goal: full HTML table anonymous box construction.

2. `add-css-grid-layout-subset`
   - Specs: `css-layout`
   - Requirements:
     - `display: grid`
     - `grid-template-columns` and `grid-template-rows` with px, percent, `fr`,
       and `auto`
     - `gap`, `row-gap`, `column-gap`
     - `grid-column` and `grid-row` line spans
     - fallback/no-op behavior for unsupported grid syntax
   - Explicit non-goal: dense auto-placement, named areas, subgrid, masonry.

3. `audit-and-close-css-units-calc`
   - Specs: `css-layout`, `css-effects`
   - Requirements:
     - document and test current support for `%`, `vw`, `vh`, `em`, `rem`, and
       `calc()`
     - resolve live unit values against viewport, parent content box, and font
       size where appropriate
     - update `SUPPORT_MATRIX.md` with exact supported/partial cases
   - First task is audit-only; implementation follows only for proven gaps.

4. `add-declarative-state-and-metadata-binding`
   - Specs: `data-binding`, `html-structure`
   - Requirements:
     - `bind-class` can toggle class names from string/list/map/boolean forms
     - `bind-attr-role` and `bind-aria-*` update semantic metadata or attributes
     - style reapplication reacts to class changes
   - Explicit non-goal: native screen-reader integration.

5. `add-declarative-input-events`
   - Specs: `data-binding`, `html-structure`
   - Requirements:
     - `onInput`, `onFocus`, `onBlur`, and `onKeyDown` dispatch registered
       commands
     - command context exposes widget ID, current value, checked state, key name,
       and validation state where applicable
     - define prevent/default and propagation policy, even if first release is
       "no propagation model"
   - Explicit non-goal: browser DOM event bubbling clone.

6. `extend-selector-subset`
   - Specs: `css-layout`, `css-effects`
   - Requirements:
     - support high-value pseudo classes such as `:not(...)`, `:first-child`,
       `:last-child`, and `:nth-child(n|odd|even)` if parser cost is contained
     - preserve current specificity/source-order behavior
     - unsupported pseudo classes remain no-op or diagnostic, not panic
   - Explicit non-goal: `:has()` unless product usage proves need.

7. `harden-visual-regression-runner`
   - Specs: `browser-parity-followups`
   - Requirements:
     - add single-case selection for CSS/SVG testloop
     - add documented browser screenshot capture command on Windows
     - generate compare reports with stable output paths under `%TEMP%` or a
       caller-provided directory
     - include fixtures for new table/grid/unit/binding/event cases

## GSD Phase Plan

### Phase 21: Table Layout Engine

Goal: Promote semantic table tags from flex-like row/cell flow to a practical
browser-inspired table layout engine.

Tasks:

1. Create OpenSpec change `add-table-layout-engine`.
   - Files: `openspec/changes/add-table-layout-engine/**`
   - Action: define table requirements for intrinsic column measurement,
     `rowspan`, `colspan`, `table-layout`, and `border-collapse`; explicitly
     exclude full HTML table anonymous box behavior.
   - Verify: `openspec validate add-table-layout-engine --strict`
   - Done: requirements are specific enough to write failing tests.

2. Add layout contracts and tests before implementation.
   - Files: `ui/html_semantics_test.go`, `ui/layout_test.go`
   - Action: add failing tests for `colspan`, `rowspan`, fixed table layout,
     auto column measurement, and collapsed/separate border policy.
   - Verify: `go test ./ui -run "Table|Layout" -count=1`
   - Done: tests fail only for missing table behavior.

3. Implement the table layout slice and docs.
   - Files: `ui/layout.go`, `ui/parser.go`, `ui/types.go`,
     `docs/SUPPORT_MATRIX.md`, `docs/HTML-CSS-BINDING-GAP-LIST.md`
   - Action: add a table-specific layout path behind semantic table metadata;
     preserve current simple table behavior for no-span cases.
   - Verify: `go test ./...`; `go build ./...`; `openspec validate add-table-layout-engine --strict`
   - Done: table requirements pass and docs distinguish supported table subset
     from unsupported browser edge cases.

Context cost: medium-heavy. Keep as its own phase.

### Phase 22: CSS Grid Subset

Goal: Add a scoped CSS Grid layout path for practical two-dimensional layouts.

Tasks:

1. Create OpenSpec change `add-css-grid-layout-subset`.
   - Files: `openspec/changes/add-css-grid-layout-subset/**`
   - Action: define `display: grid`, template columns/rows, px/%/fr/auto
     tracks, gaps, and simple line spans; exclude named areas/subgrid/dense
     placement.
   - Verify: `openspec validate add-css-grid-layout-subset --strict`
   - Done: parser and layout behavior are testable.

2. Add style parser and layout tests.
   - Files: `ui/style_explicit_test.go`, `ui/layout_test.go`
   - Action: test grid declarations through JSON style and CSS rule loading;
     add layout expectations for fixed, percent, `fr`, `auto`, gaps, and spans.
   - Verify: `go test ./ui -run "Grid|Style" -count=1`
   - Done: tests fail only for missing grid behavior.

3. Implement grid subset and visual fixture.
   - Files: `ui/types.go`, `ui/style.go`, `ui/layout.go`,
     `cmd/css_testloop/testcases.go`, `docs/SUPPORT_MATRIX.md`
   - Action: add grid style fields, parse declarations, and route grid
     containers through a dedicated layout branch.
   - Verify: `go test ./...`; `go build ./...`;
     `go run ./cmd/css_testloop -mode render -out %TEMP%/ebitenui-grid.png`
   - Done: grid subset is documented and visually smoke-tested.

Context cost: heavy. Do after table or split further if table introduces shared
layout abstractions that grid should reuse.

### Phase 23: CSS Units And Calc Audit/Closure

Goal: Replace ambiguous unit documentation with tested support and close proven
runtime gaps.

Tasks:

1. Create OpenSpec change `audit-and-close-css-units-calc`.
   - Files: `openspec/changes/audit-and-close-css-units-calc/**`
   - Action: define expected behavior for `%`, `vw`, `vh`, `em`, `rem`, and
     `calc()` per property group.
   - Verify: `openspec validate audit-and-close-css-units-calc --strict`
   - Done: each unit has owner context: viewport, parent, root font, or current
     font.

2. Add audit tests that describe current truth.
   - Files: `ui/style_explicit_test.go`, `ui/layout_test.go`,
     `ui/effects_runtime_test.go`
   - Action: write tests for current unit/calc behavior before implementation;
     mark missing behavior by failing tests rather than assumptions.
   - Verify: `go test ./ui -run "Unit|Calc|Percent|Viewport|Rem|Em" -count=1`
   - Done: exact implemented/partial/missing matrix is known.

3. Implement only proven gaps and update docs.
   - Files: `ui/types.go`, `ui/style.go`, `ui/layout.go`,
     `ui/variables.go`, `docs/SUPPORT_MATRIX.md`,
     `docs/HTML-CSS-BINDING-GAP-LIST.md`
   - Action: resolve live units through existing style/layout paths without
     changing already-passing px behavior.
   - Verify: `go test ./...`; `go build ./...`;
     `openspec validate audit-and-close-css-units-calc --strict`
   - Done: support matrix states exact unit support and tests lock it down.

Context cost: medium. Run before broadening selector/event surfaces if layout
phases depend on unit semantics.

### Phase 24: Declarative State, Metadata, And Events

Goal: Make XML-driven application logic cover common HTML-like state classes,
semantic metadata, and input/focus/key events without custom widget setup.

Tasks:

1. Create OpenSpec changes `add-declarative-state-and-metadata-binding` and
   `add-declarative-input-events`.
   - Files: `openspec/changes/add-declarative-state-and-metadata-binding/**`,
     `openspec/changes/add-declarative-input-events/**`
   - Action: define `bind-class`, semantic role/ARIA metadata behavior,
     `onInput`, `onFocus`, `onBlur`, `onKeyDown`, event payload fields, and
     propagation/default policy.
   - Verify:
     `openspec validate add-declarative-state-and-metadata-binding --strict`;
     `openspec validate add-declarative-input-events --strict`
   - Done: no browser DOM event bubbling or native accessibility claim is made.

2. Add binding and event tests.
   - Files: `ui/binding_test.go`, `ui/html_semantics_test.go`
   - Action: test class binding from string/list/map/boolean, style
     reapplication on class changes, role/ARIA metadata storage, and event
     command payloads for input/focus/blur/key.
   - Verify: `go test ./ui -run "BindClass|Aria|Role|OnInput|OnFocus|OnBlur|OnKeyDown" -count=1`
   - Done: tests fail only for missing binding/event behavior.

3. Implement binding/event surface.
   - Files: `ui/parser.go`, `ui/binding.go`, `ui/binding_expr.go`,
     `ui/ui.go`, `ui/input.go`, `ui/widget.go`, `ui/types.go`,
     `docs/SUPPORT_MATRIX.md`
   - Action: extend parser and binding registry using existing diagnostic
     patterns; keep event payload deterministic and small.
   - Verify: `go test ./...`; `go build ./...`;
     `openspec validate add-declarative-state-and-metadata-binding --strict`;
     `openspec validate add-declarative-input-events --strict`
   - Done: XML authors can drive state classes and input events declaratively.

Context cost: medium-heavy. Keep class/metadata and event changes in one phase
only if file overlap is high; otherwise split into two phases.

### Phase 25: Selector Subset Extension

Goal: Add only high-value selector semantics that are useful for XML UI authoring.

Tasks:

1. Create OpenSpec change `extend-selector-subset`.
   - Files: `openspec/changes/extend-selector-subset/**`
   - Action: specify supported pseudo classes: `:not(...)`, `:first-child`,
     `:last-child`, and `:nth-child(n|odd|even)`; define specificity and
     unsupported pseudo behavior.
   - Verify: `openspec validate extend-selector-subset --strict`
   - Done: `:has()` and CSS layers remain explicit non-goals.

2. Add selector tests.
   - Files: `ui/style_explicit_test.go`, `ui/effects_runtime_test.go`
   - Action: test selector parsing, specificity/source order, and style
     application for the selected pseudo classes.
   - Verify: `go test ./ui -run "Selector|Pseudo|Nth|Not" -count=1`
   - Done: parser failures are isolated to missing pseudo support.

3. Implement parser/matcher extension.
   - Files: `ui/selector.go`, `ui/style.go`, `ui/ui.go`,
     `docs/SUPPORT_MATRIX.md`
   - Action: extend the existing selector AST/matcher rather than adding ad hoc
     string matching in style application.
   - Verify: `go test ./...`; `go build ./...`;
     `openspec validate extend-selector-subset --strict`
   - Done: supported pseudo classes work with current cascade behavior.

Context cost: medium.

### Phase 26: Visual Regression Runner Hardening

Goal: Make browser-parity evidence repeatable for the new table, grid, unit,
binding, event, and selector work.

Tasks:

1. Create OpenSpec change `harden-visual-regression-runner`.
   - Files: `openspec/changes/harden-visual-regression-runner/**`
   - Action: define single-case runner, browser capture command, deterministic
     output paths, and compare report expectations.
   - Verify: `openspec validate harden-visual-regression-runner --strict`
   - Done: workflow is concrete for Windows.

2. Add runner options and fixtures.
   - Files: `cmd/css_testloop/main.go`, `cmd/css_testloop/testcases.go`,
     `cmd/svg_testloop/main.go`, `docs/CSS_COMPARE.md`
   - Action: add `-case` or equivalent filtering, stable output naming, and
     fixtures for table/grid/unit/selectors; document browser screenshot steps.
   - Verify:
     `go run ./cmd/css_testloop -mode render -case grid-basic -out %TEMP%/grid-basic.png`;
     `go run ./cmd/css_testloop -mode html -case grid-basic -out %TEMP%/grid-basic.html`
   - Done: a single case can render without editing code.

3. Close docs and support matrix.
   - Files: `docs/SUPPORT_MATRIX.md`, `docs/HTML-CSS-BINDING-GAP-LIST.md`,
     `.planning/STATE.md`
   - Action: publish final supported/partial/unsupported rows and remove stale
     remaining items that were implemented.
   - Verify: `go test ./...`; `go build ./...`; `openspec validate --all --strict`
   - Done: docs match tests and OpenSpec archive state.

Context cost: medium.

## Recommended Execution Order

1. Phase 23: CSS Units And Calc Audit/Closure
2. Phase 21: Table Layout Engine
3. Phase 22: CSS Grid Subset
4. Phase 24: Declarative State, Metadata, And Events
5. Phase 25: Selector Subset Extension
6. Phase 26: Visual Regression Runner Hardening

Reasoning:

- Unit/calc behavior should be settled before adding grid/table expectations
  that depend on parent, viewport, and font-relative measurements.
- Table and grid are the largest layout gaps and should be isolated from
  binding/event work.
- Selector/event/binding improvements are high-value but should not block core
  layout parity.
- Visual runner hardening should finish after new layout surfaces add fixtures,
  unless a phase needs the runner earlier for reliable evidence.

## Verification Strategy

Every implementation phase should run:

- `openspec validate <change-id> --strict`
- `openspec validate --all --strict` before archive/closeout
- `go test ./...`
- `go build ./...`
- `go vet ./...` when the touched scope includes non-test production code
- `git diff --check`

Visual/layout/effect phases should also run targeted testloop smoke:

- `go run ./cmd/css_testloop -mode render -out %TEMP%/<case>.png`
- `go run ./cmd/css_testloop -mode html -out %TEMP%/<case>.html`
- `go run ./cmd/css_testloop -mode compare -browser <browser.png> -ebiten <ebiten.png> -out %TEMP%/<case>-report.html`

GDC should be used as a touched-scope structural check when implementation
changes meaningful nodes:

- Use targeted sync/check for touched files or nodes, not full-repo drift as a
  blocker.
- Treat the existing `ui.RadioButton <-> ui.RadioGroup` cycle reports and broad
  orphan info as known baseline unless the phase worsens them.
- Example after a table phase:
  `gdc sync --direction both --files ui/layout.go,ui/parser.go,ui/types.go`
  then a dependency/touched-scope check.

OpenSpec closeout:

- Update related `openspec/changes/<change-id>/tasks.md` as implementation
  lands.
- Update affected specs before archive.
- Update `docs/SUPPORT_MATRIX.md` and
  `docs/HTML-CSS-BINDING-GAP-LIST.md` with exact supported/partial/unsupported
  status.

## Out Of Scope

- Full browser clone behavior.
- Full HTML table anonymous box construction.
- Complete CSS parser grammar, cascade origins/layers, full nesting, `@supports`,
  container queries, and all media features.
- `:has()` unless a concrete product layout needs it.
- Native browser DOM event bubbling/capture semantics.
- Native screen-reader integration from ARIA metadata.
- OS/browser font discovery and complex shaping fallback beyond explicit
  registered font sources/faces.
- Replacing the existing Ebiten-first rendering model with a browser engine.
