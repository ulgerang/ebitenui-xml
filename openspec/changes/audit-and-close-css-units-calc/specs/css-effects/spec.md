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

#### Scenario: sibling selector sets visual style

Given a stylesheet contains `.label + .value { color: #ffffff; }`
When the stylesheet is loaded and the layout is applied
Then only matching following sibling widgets receive the declared color.

#### Scenario: same-specificity CSS source order wins

Given a stylesheet contains `.primary { color: #111111; }` before `.accent { color: #ffffff; }`
And a widget has both `primary` and `accent` classes
When styles are applied
Then the later same-specificity rule wins for overlapping properties.

#### Scenario: pseudo state rule applies only in state

Given a stylesheet contains `.button:hover { color: #ffffff; }`
When the stylesheet is loaded and the layout is applied
Then the hover color is stored on the matching style's hover state
And the base color is not changed until the widget enters hover state.

#### Scenario: parser keeps function punctuation inside declarations

Given a stylesheet contains `clip-path: path('M0 0 L10 0; L10 10 Z')`
When the stylesheet is loaded
Then the declaration is parsed as one value despite the semicolon inside quotes.

#### Scenario: UI style loading resolves CSS variables

Given the UI variable `--primary` is set to `#ffffff`
When `UI.LoadCSS` or `UI.LoadStyles` receives a declaration containing `var(--primary)`
Then the resulting widget style uses the resolved value.

#### Scenario: visual relative unit boundary is documented

Given CSS visual declarations use pixel-valued transform, blur, shadow, or
typography fields
When those declarations are loaded today
Then pixel and unitless numeric values are supported
And `%`, `vw`, `vh`, `em`, `rem`, and `calc()` are not claimed as browser-live
visual unit support unless routed through an explicit resolver.
