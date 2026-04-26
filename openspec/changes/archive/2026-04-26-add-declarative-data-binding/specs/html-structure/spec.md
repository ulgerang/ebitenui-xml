# HTML/XML Structure Spec

## ADDED Requirements

### Requirement: Semantic XML aliases

XML authors SHALL be able to use common HTML-like semantic tags as aliases for
existing widgets without custom Go setup.

#### Scenario: headings and landmarks parse into usable widgets

Given a layout contains `<main><nav><h1 id="title">Title</h1></nav></main>`
When the layout is loaded
Then `main` and `nav` are container widgets
And `h1` is a text widget with semantic metadata and a matching CSS class.

### Requirement: DOM-like query helpers

Go callers SHALL be able to query widget subtrees with simple class, ID, and type
selectors.

#### Scenario: query by class type and ID

Given a loaded widget tree contains widgets with class `top`, semantic tag `li`,
and ID `item-b`
When the caller queries `.top`, `li`, or `#item-b`
Then the matching widgets are returned in tree order.

### Requirement: Form semantics

Form-like XML containers SHALL support submit/reset commands, validation state,
and disabled fieldset propagation.

#### Scenario: submit and reset form

Given a form has `onSubmit` and `onReset` commands
When Go calls `SubmitForm` or `ResetForm`
Then the registered command receives the form widget
And reset restores descendant form fields to their initial XML values.

#### Scenario: disabled fieldset disables descendants

Given a layout contains `<fieldset disabled="true">` with input descendants
When the layout is loaded
Then the fieldset descendants are disabled.

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

