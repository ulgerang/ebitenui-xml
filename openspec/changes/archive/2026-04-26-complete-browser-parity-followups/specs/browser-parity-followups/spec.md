# Browser Parity Follow-ups Spec

## ADDED Requirements

### Requirement: Dynamic collection option binding

The XML binding system SHALL support declarative option collection binding for
complex selection widgets.

#### Scenario: dropdown options from collection

Given a dropdown declares `bind-options` and label/value field mappings
When the bound collection changes
Then the dropdown options are rebuilt from the collection
And the selected value is preserved when it still exists.

### Requirement: Form validation rules and messages

Form-capable widgets SHALL evaluate common validation rules and expose
validation messages.

#### Scenario: invalid submit is blocked

Given a form contains a required field with no value
When the form is submitted
Then validation state is set to invalid
And submit command dispatch is blocked or marked invalid according to the form
policy.

### Requirement: Keyboard and modal focus policies

Interactive widgets SHALL support deterministic keyboard navigation and modal
focus containment.

#### Scenario: modal focus trap

Given a modal is open
When the user advances focus with Tab or Shift+Tab
Then focus remains within the modal
And focus is restored to the previous widget when the modal closes.

### Requirement: Scroll overflow runtime

Widgets with scroll overflow SHALL maintain scroll offsets and map drawing and
hit testing through those offsets.

#### Scenario: scrolled child hit test

Given an overflow scroll container has a child below the visible viewport
When the container scroll offset reveals the child
Then hit testing at the visible child position returns that child.

### Requirement: Literal CSS keyframes parser

The style system SHALL support literal CSS-like `@keyframes` syntax in addition
to JSON keyframes.

#### Scenario: keyframes block registers animation

Given a stylesheet contains `@keyframes pop { from { opacity: 0 } to { opacity: 1 } }`
When the stylesheet is loaded
Then the animation registry contains `pop`
And style animation declarations can reference `pop`.

### Requirement: Advanced visual fidelity

The renderer SHALL support selected advanced visual features or document explicit
no-op behavior when the cost is not justified.

#### Scenario: polygon clip path

Given a widget has `clipPath: "polygon(...)"`
When the widget is rendered
Then the widget content is masked to the polygon.

### Requirement: Visual comparison hardening

The project SHALL provide repeatable visual comparison evidence for browser
parity features.

#### Scenario: compare workflow produces report

Given the CSS testloop render and HTML reference outputs exist
When the compare workflow is run with a browser capture
Then a report is produced without manual code edits.
