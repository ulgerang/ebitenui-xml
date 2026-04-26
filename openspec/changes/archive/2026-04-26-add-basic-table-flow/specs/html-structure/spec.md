## MODIFIED Requirements

### Requirement: Semantic XML aliases

Common HTML-like XML tags SHALL map to existing widgets while preserving the
source tag as semantic metadata for querying and styling.

#### Scenario: heading and list aliases

Given XML containing `main`, `nav`, `h1`, `ul`, and `li`
When the layout is loaded
Then structural tags map to panel-like widgets
And text tags map to text-like widgets
And the original tag remains queryable as semantic type metadata.

#### Scenario: basic table aliases

Given XML containing `table`, `tbody`, `tr`, `td`, and `th`
When the layout is loaded and arranged
Then table section tags arrange child rows vertically
And row tags arrange cells horizontally
And cell tags can expand as flexible panel cells.
