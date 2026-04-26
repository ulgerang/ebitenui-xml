# css-effects Specification Delta

## ADDED Requirements

### Requirement: CSS rule subset for visual properties

The style engine SHALL accept simple CSS selector blocks for common visual
properties.

#### Scenario: id rule sets visual effects

Given a stylesheet contains `#title { color: white; opacity: 0.5; transform: scale(1.2); }`
When the stylesheet is loaded
Then the `#title` style records color, opacity, and transform values.
