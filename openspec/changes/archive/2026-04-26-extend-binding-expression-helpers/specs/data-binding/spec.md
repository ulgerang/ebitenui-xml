# data-binding Specification Delta

## MODIFIED Requirements

### Requirement: Rich binding expressions

Template and one-way XML bindings SHALL support safe expressions for common UI
conditions and formatting without requiring custom Go glue code.

#### Scenario: fallback and formatting helpers update from dependencies

Given a layout contains `<text id="greeting">Hello {{upper(user.name || "guest")}}</text>`
When the binding context has no `user.name`
Then the text content is `Hello GUEST`
When the binding context sets `user.name` to `Ada`
Then the text content is `Hello ADA`.

#### Scenario: boolean and arithmetic expressions update from dependencies

Given a layout contains `<text id="summary">{{count + 1}}/{{total}}</text>`
When the binding context sets `count` to `2` and `total` to `5`
Then the text content is `3/5`.

#### Scenario: helper expressions format common UI values

Given a layout contains helper expressions using `len`, `round`, `contains`, `join`, and `format`
When the binding context updates the referenced values
Then the text content updates using the helper results.
