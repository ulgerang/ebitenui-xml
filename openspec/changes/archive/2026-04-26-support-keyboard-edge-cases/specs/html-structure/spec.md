## MODIFIED Requirements

### Requirement: Form semantics

HTML-like form tags SHALL map to XML widgets with reset, submit, and validation
semantics.

#### Scenario: form submit command

Given a form declares `onSubmit`
When the form is submitted explicitly or by pressing Enter in a focused
single-line text input descendant
Then validation runs
And the registered submit command receives the form widget when valid.
