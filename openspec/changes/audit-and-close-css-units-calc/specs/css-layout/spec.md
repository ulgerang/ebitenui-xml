## MODIFIED Requirements

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

### Requirement: CSS unit and calc audit boundary

The project SHALL document and test the current support boundary for relative
CSS units and `calc()` before claiming browser-like runtime sizing parity.

#### Scenario: lower-level unit resolver handles relative units

Given Go code parses `%`, `vw`, `vh`, `em`, and `rem` values with
`ParseSizeValue`
When those values are resolved with a `SizeContext`
Then the result uses the supplied parent size, viewport size, current font size,
or root font size.

#### Scenario: lower-level calc resolver handles mixed units

Given Go code parses `calc(50% - 10px)` with `ParseCalc`
When the expression is resolved with a parent size of 200 pixels
Then the result is 90 pixels.

#### Scenario: production style loaders remain pixel-based for layout values

Given a CSS rule declaration or XML inline style uses `%`, `vw`, `vh`, `em`,
`rem`, or `calc()` for width-like layout properties
When the style is loaded through the current production loader
Then the value is not browser-live-resolved against parent, viewport, or font
context
And the support matrix records this as partial support.
