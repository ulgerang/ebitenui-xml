# css-layout Specification Delta

## ADDED Requirements

### Requirement: CSS rule subset for layout properties

The style engine SHALL accept simple CSS selector blocks for common layout
properties.

#### Scenario: class rule sets flex layout

Given a stylesheet contains `.card { display: flex; flex-direction: column; gap: 8px; width: 120px; }`
When the stylesheet is loaded
Then the `.card` style sets column direction, gap, and width.
