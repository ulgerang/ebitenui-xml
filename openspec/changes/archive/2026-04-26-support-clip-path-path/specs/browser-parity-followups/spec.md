## MODIFIED Requirements

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
