# browser-parity-followups Specification

## Purpose
TBD - created by archiving change complete-browser-parity-followups. Update Purpose after archive.
## Requirements
### Requirement: Dynamic collection option binding

The XML binding system SHALL support declarative option collection binding for
complex selection widgets.

#### Scenario: dropdown options from collection

Given a dropdown declares `bind-options` and label/value field mappings
When the bound collection changes
Then the dropdown options are rebuilt from the collection
And the selected value is preserved when it still exists.

#### Scenario: checkbox options from collection

Given a panel declares `bind-options`, `option-type="checkbox"`, and
label/value field mappings
When the bound collection changes
Then checkbox children are rebuilt from the collection
And the checked value collection is preserved for matching options.

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
focus behavior comparable to common browser controls.

#### Scenario: modal focus trap

Given a modal is open
When the user tabs through focusable elements
Then focus remains within the modal
And focus is restored to the previous widget when the modal closes.

#### Scenario: nested modal focus restore

Given an outer modal is open and focus is inside it
And an inner modal is opened from the outer modal
When the user presses Escape
Then only the inner modal closes
And focus is restored to the previous outer modal widget
When the user presses Escape again
Then the outer modal closes
And focus is restored to the widget that was focused before the outer modal opened.

#### Scenario: keyboard activation and form submit

Given a focused button, checkbox, toggle, radio button, slider, or form text input
When the user presses the browser-standard activation key for that control
Then the control applies the corresponding click, check, selection, min/max, or
form submit behavior.

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

#### Scenario: path clip path

Given a widget has `clipPath: "path('M0 0 L100 0 L50 100 Z')"`
When the widget is rendered
Then the widget content is masked to the parsed path.

### Requirement: Visual comparison hardening

The project SHALL provide repeatable visual comparison evidence for browser
parity features.

#### Scenario: compare workflow produces report

Given the CSS testloop render and HTML reference outputs exist
When the compare workflow is run with a browser capture
Then a report is produced without manual code edits.

