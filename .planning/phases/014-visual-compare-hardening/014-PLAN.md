# Phase 14 Plan: Visual Compare Hardening and Release Closeout

## Status

Completed.

## Objective

Make the browser comparison path repeatable on Windows and publish a clear
support matrix for implemented, partial, and unsupported HTML/CSS features.

## Canonical Inputs

- `cmd/css_testloop/`
- `tools/css_compare/`
- `docs/HTML-CSS-BINDING-GAP-LIST.md`
- `README.md`
- `docs/CSS_COMPARE.md`
- `openspec/changes/add-declarative-data-binding`
- `openspec/changes/complete-browser-parity-followups`

## Tasks

1. [x] Make the CSS visual workflow easy to run:
   - render Ebiten output
   - generate HTML reference
   - capture browser reference
   - compare and emit report
2. [x] Add deterministic visual cases for newly implemented follow-up features:
   - dynamic dropdown binding
   - validation messages
   - modal focus state
   - scroll offset
   - CSS keyframes snapshot
   - polygon clip-path
3. [x] Publish a feature support matrix:
   - supported
   - partial
   - intentionally unsupported
   - future work
4. [x] Reconcile OpenSpec:
   - archive completed `add-declarative-data-binding` if workflow supports it
   - archive `complete-browser-parity-followups` after all follow-up tasks are complete
5. [x] Run final validation and record evidence.

## Evidence

- Added `docs/SUPPORT_MATRIX.md`.
- Added a clean CSS visual compare workflow section to `docs/CSS_COMPARE.md`.
- Updated README links and feature summary.
- Archived OpenSpec changes `add-declarative-data-binding` and
  `complete-browser-parity-followups`.
- Follow-up visual fixtures now include overflow scroll and polygon clip-path.

## Verification

- `go test ./...`
- `go build ./...`
- `go vet ./...`
- `openspec validate --strict`
- CSS visual compare path produces report without manual code edits

## Done When

- A future agent or developer can run visual regression from docs alone.
- Docs accurately distinguish supported and partial browser parity.
- OpenSpec no longer mixes completed work with future follow-ups.
