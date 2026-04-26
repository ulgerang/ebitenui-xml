# html-structure Specification

## Purpose
TBD - created by archiving change add-declarative-data-binding. Update Purpose after archive.
## Requirements
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

### Requirement: DOM-like query helpers

Go callers SHALL be able to query widget subtrees with simple class, ID, and type
selectors.

#### Scenario: query by class type and ID

Given a loaded widget tree contains widgets with class `top`, semantic tag `li`,
and ID `item-b`
When the caller queries `.top`, `li`, or `#item-b`
Then the matching widgets are returned in tree order.

### Requirement: Form semantics

HTML-like form tags SHALL map to XML widgets with reset, submit, and validation
semantics.

#### Scenario: form submit command

Given a form declares `onSubmit`
When the form is submitted explicitly or by pressing Enter in a focused
single-line text input descendant
Then validation runs
And the registered submit command receives the form widget when valid.

### Requirement: XML radio grouping and focus traversal

XML authors SHALL be able to group radio buttons by `name` and define keyboard
focus order with `tabindex`.

#### Scenario: radio group by name

Given two radio buttons share the same `name`
When one radio is clicked
Then it is selected and the other radio in the group is unselected.

#### Scenario: focus traversal by tabindex

Given focusable widgets have tabindex values
When focus moves forward or backward
Then widgets are focused in tabindex order, preserving tree order for ties.

