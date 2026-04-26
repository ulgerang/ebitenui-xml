# CSS Layout Spec

## ADDED Requirements

### Requirement: Flex distribution and sizing parity

The layout engine SHALL support high-impact flexbox distribution and sizing
features for XML/CSS-style layouts.

#### Scenario: justify distribution

Given a row container uses `justify` values `space-between`, `space-around`, or
`space-evenly`
When the layout engine lays out fixed-size children
Then the remaining space is distributed according to the selected justify mode.

#### Scenario: box sizing

Given a child has width, padding, and border
When `boxSizing` is `content-box`
Then the computed outer size includes content, padding, and border
When `boxSizing` is `border-box`
Then the computed outer size is the declared width or height.

#### Scenario: flex shrink and min/max constraints

Given children overflow the main axis
When children have `flexShrink` and min/max constraints
Then shrinkable children reduce along the main axis while zero-shrink and min
size constraints are preserved.

#### Scenario: flex wrap

Given a flex container has `flexWrap: "wrap"`
When children exceed the available main-axis space
Then children are placed onto additional flex lines.

### Requirement: Positioned layout, clipping, and z ordering

The layout and interaction engine SHALL support CSS-like absolute positioning,
overflow clipping, and z-index ordering for overlapping XML widget trees.

#### Scenario: absolute positioning

Given a child has `position: "absolute"` and inset values such as `top`,
`right`, `bottom`, or `left`
When the layout engine lays out the parent
Then the child is positioned against the parent's available content area
And the absolute child does not consume flex layout space for normal siblings.

#### Scenario: inset-derived size

Given an absolute child has opposing horizontal or vertical insets and no
explicit size for that axis
When the layout engine computes its rectangle
Then the child size is derived from the containing area minus those insets and
margins.

#### Scenario: overflow hit clipping

Given a parent has `overflow: "hidden"`, `overflow: "scroll"`, or
`overflow: "auto"`
When a child extends outside the parent's computed rectangle
Then drawing is clipped to the parent rectangle
And hit testing does not return the clipped child outside the parent bounds.

#### Scenario: z-index visual and hit order

Given overlapping siblings have different `zIndex` values
When widgets are drawn or hit tested
Then children are drawn from lower to higher z-index
And hit testing returns the highest visible z-index first while preserving source
order for equal z-index values.
