## MODIFIED Requirements

### Requirement: XML repeat/list binding

XML layout authors SHALL be able to repeat a template element for each item in a
bound collection using `bind-repeat`, `data-bind-repeat`, or `for-each`.

#### Scenario: repeated text widgets render item fields

Given a layout contains `<text id="player-{{index}}" bind-repeat="players">{{item.Name}}</text>`
When the UI binding context sets `players` to two player structs
Then the parent contains two text widgets
And each repeated widget can use `{{item.<field>}}` and `{{index}}` in text and attributes.

#### Scenario: repeated children update when the collection changes

Given a repeated template is bound to `players`
When the UI binding context replaces `players` with a shorter collection
Then old repeated children are removed
And the widget ID lookup cache reflects the new repeated children only.

#### Scenario: radio options binding

Given a panel declares `bind-options="items" option-type="radio" bind-value="selected"`
When the bound collection is set
Then radio button children are rebuilt from that collection
And selecting a generated radio button updates `selected`
And the selected value is preserved if it still exists after refresh.
