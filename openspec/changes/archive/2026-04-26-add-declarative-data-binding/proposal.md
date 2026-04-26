# Add Declarative Data Binding

## Why

The framework advertises XML-driven UI and reactive data binding, but binding
connections currently require imperative Go calls after loading the XML. XML
layouts should be able to declare common bindings directly so the layout remains
the source of UI structure and state wiring.

## What Changes

- Add XML binding attributes for common widget state:
  - `bind-text`
  - `bind-value`
  - `bind-checked`
  - `bind-visible`
  - `bind-enabled`
- Support `data-bind-*` aliases for HTML-style authoring.
- Keep existing Go binding APIs working.

## Non-Goals

- Template expressions like `{{name}}`.
- Repeater/list rendering from XML.
- Full expression language or conditional DOM creation.
