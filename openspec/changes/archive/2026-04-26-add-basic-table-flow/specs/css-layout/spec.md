## MODIFIED Requirements

### Requirement: Flex distribution and sizing parity

The layout engine SHALL support high-impact flexbox distribution and sizing
features for XML/CSS-style layouts.

#### Scenario: semantic table flow

Given table semantic aliases without explicit author layout styles
When layout is calculated
Then table groups use column flow
And table rows use row flow
And table cells share available row width through flex growth.
