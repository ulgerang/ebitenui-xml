## ADDED Requirements

### Requirement: Text intrinsic line box sizing

Text-like widgets SHALL use the same line box height for intrinsic layout
sizing that they use when rendering text.

#### Scenario: explicit line height sizes text intrinsic height

Given a text widget has Korean content and an explicit `lineHeight`
When the layout engine computes the widget's intrinsic height
Then the height includes at least that line height plus padding and border.

#### Scenario: column text siblings do not overlap

Given multiple Korean text widgets are arranged in a column with a gap
When the widgets do not declare explicit heights
Then each widget receives a computed height at least as tall as its text line box
And consecutive text widgets are separated by the requested gap.

#### Scenario: bound text relayout preserves line boxes

Given a Korean text widget starts with empty bound content in a column layout
When the binding updates the widget content to a non-empty Korean string
Then the layout recomputes the widget height from the rendered line box
And subsequent column siblings keep the requested gap.

#### Scenario: invisible bound children do not consume column space

Given a column contains a text child controlled by `bind-visible`
When the binding makes that child invisible
Then the invisible child receives an empty computed rectangle
And it does not contribute intrinsic size or column gap spacing.
