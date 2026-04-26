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
