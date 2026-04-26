# css-layout Specification

## Purpose
TBD - created by archiving change add-declarative-data-binding. Update Purpose after archive.
## Requirements
### Requirement: Flex distribution and sizing parity

The layout engine SHALL support high-impact flexbox distribution and sizing
features for XML/CSS-style layouts.

#### Scenario: semantic table flow

Given table semantic aliases without explicit author layout styles
When layout is calculated
Then table groups use column flow
And table rows use row flow
And table cells share available row width through flex growth.

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

### Requirement: CSS rule subset for layout properties

The style engine SHALL accept simple CSS selector blocks for common layout
properties.

#### Scenario: class rule sets flex layout

Given a stylesheet contains `.card { display: flex; flex-direction: column; gap: 8px; width: 120px; }`
When the stylesheet is loaded
Then the `.card` style sets column direction, gap, and width.

#### Scenario: CSS rule string is auto-detected

Given `LoadFromString` receives `.card { width: 120px; }`
When the string is not a JSON object
Then the style engine parses it as CSS rules.

#### Scenario: viewport media query gates layout rules

Given `UI.LoadCSS` receives `@media (min-width: 300px) { .card { width: 120px; } }`
When the UI viewport width is at least 300 pixels
Then the nested rule is applied
When the UI viewport width is below 300 pixels
Then the nested rule is not applied.

#### Scenario: expanded media query gates layout rules

Given `UI.LoadCSS` receives `@media screen and (orientation: landscape), print { .card { gap: 12px; } }`
When the UI viewport is wider than it is tall
Then the nested rule is applied
When the UI viewport is taller than it is wide
Then the nested rule is not applied.

#### Scenario: descendant and child selectors affect layout

Given a stylesheet contains `.toolbar > button { width: 48px; }`
When the stylesheet is loaded and the layout is applied
Then only direct button children of `.toolbar` receive the declared width.

#### Scenario: important layout declaration wins

Given a stylesheet contains `.card { width: 80px !important; } #card { width: 120px; }`
When the stylesheet is loaded and the layout is applied
Then the `.card` important width wins over the normal ID width.

