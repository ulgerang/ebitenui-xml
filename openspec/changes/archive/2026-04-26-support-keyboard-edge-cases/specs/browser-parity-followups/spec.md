## MODIFIED Requirements

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
