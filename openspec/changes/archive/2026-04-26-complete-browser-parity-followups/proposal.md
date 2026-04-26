# Complete Browser Parity Follow-ups

## Why

The first HTML/CSS/data-binding milestone delivered practical XML binding,
semantic aliases, flex/position layout, visual effects, animation, and regression
coverage. Remaining work is advanced browser parity and runtime polish, not the
basic authoring foundation.

## What Changes

- Add dynamic collection binding for complex widgets such as dropdown/select.
- Add richer form validation rules and renderable validation messages.
- Strengthen keyboard policies for dropdowns, radio groups, sliders, forms, and
  modals.
- Implement true scroll offsets for `overflow: scroll` and `overflow: auto`.
- Support literal CSS-like `@keyframes` syntax in addition to JSON `keyframes`.
- Improve advanced visual fidelity for complex clip paths, shadows, text, and
  optionally table layout.
- Harden visual comparison workflow and publish a feature support matrix.

## Non-Goals

- Full browser engine compatibility.
- General CSS cascade/parser replacement beyond the narrow syntax needed for
  keyframes and documented style features.
- Implementing table layout unless actual project usage justifies the cost.

## Validation

- Go tests for binding, form, keyboard, scroll, parser, and visual helper logic.
- `go test ./...`
- `go build ./...`
- `go vet ./...`
- `openspec validate complete-browser-parity-followups --strict`
- CSS testloop render/html/compare coverage where visual parity matters.
