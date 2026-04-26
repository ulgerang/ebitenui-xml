# ebitenui-xml GSD Project

## Goal

Bring practical HTML/CSS-style authoring into Ebitengine through XML layout,
CSS-like JSON styling, data binding, visual effects, and visual regression
coverage.

## Intent Baseline

- OpenSpec is the product intent authority.
- Current active change: `openspec/changes/add-declarative-data-binding`
- Follow-up change: `openspec/changes/complete-browser-parity-followups`
- Gap inventory: `docs/HTML-CSS-BINDING-GAP-LIST.md`
- Implementation evidence must come from `go test ./...`, `go build ./...`,
  OpenSpec validation, and where relevant CSS/SVG visual regression tools.

## Execution Rules

- Keep each phase independently buildable and testable.
- Update OpenSpec delta specs/tasks when behavior changes.
- Add Go unit tests for parser, binding, layout, and rendering helpers.
- Add or update `cmd/css_testloop/testcases.go` for browser-comparison features.
- GDC is currently unavailable because `.gdc/` does not exist.
- Graphify is currently unavailable because `graphify-out/` does not exist.
