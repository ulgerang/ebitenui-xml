## MODIFIED Requirements

### Requirement: CSS rule subset for visual properties

The style engine SHALL accept simple CSS selector blocks for common visual and
effect properties.

#### Scenario: class rule sets effects

Given a stylesheet contains `.badge { transform: scale(1.2); filter: blur(2px); opacity: 0.5; }`
When the stylesheet is loaded
Then the `.badge` style records transform, filter, and opacity declarations.

#### Scenario: descendant selector sets visual style

Given a stylesheet contains `.card .title { color: #ffffff; }`
When the stylesheet is loaded and the layout is applied
Then only title descendants of `.card` receive the declared color.

#### Scenario: UI style loading resolves CSS variables

Given the UI variable `--primary` is set to `#ffffff`
When `UI.LoadCSS` or `UI.LoadStyles` receives a declaration containing `var(--primary)`
Then the resulting widget style uses the resolved value.
