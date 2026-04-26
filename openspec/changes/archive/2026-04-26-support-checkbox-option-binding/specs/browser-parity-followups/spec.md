## MODIFIED Requirements

### Requirement: Dynamic collection option binding

The XML binding system SHALL support declarative option collection binding for
complex selection widgets.

#### Scenario: dropdown options from collection

Given a dropdown declares `bind-options` and label/value field mappings
When the bound collection changes
Then the dropdown options are rebuilt from the collection
And the selected value is preserved when it still exists.

#### Scenario: checkbox options from collection

Given a panel declares `bind-options`, `option-type="checkbox"`, and
label/value field mappings
When the bound collection changes
Then checkbox children are rebuilt from the collection
And the checked value collection is preserved for matching options.
