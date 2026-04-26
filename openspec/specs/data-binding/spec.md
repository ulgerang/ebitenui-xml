# data-binding Specification

## Purpose
TBD - created by archiving change add-declarative-data-binding. Update Purpose after archive.
## Requirements
### Requirement: XML-declared common bindings

XML layout authors SHALL be able to bind common widget properties to
`BindingContext` keys using declarative attributes.

#### Scenario: text and visibility update from bound data

Given a layout contains `<text id="name" bind-text="player.name" bind-visible="player.visible">`
When the UI binding context sets `player.name` to `Ada` and `player.visible` to `false`
Then the text widget content is `Ada`
And the text widget is not visible.

#### Scenario: text input updates bound data

Given a layout contains `<input id="username" bind-value="user.name">`
When the binding context sets `user.name` to `Ada`
Then the input text is `Ada`
When the input text changes to `Grace`
Then `user.name` is `Grace`.

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

#### Scenario: checkbox options binding

Given a panel declares `bind-options="items" option-type="checkbox" bind-value="selected"`
When the bound collection is set
Then checkbox children are rebuilt from that collection
And checking generated boxes updates `selected` with all checked values
And checked values are preserved when matching options still exist after refresh.

### Requirement: XML conditional rendering

XML layout authors SHALL be able to attach or detach a template element using a
boolean binding with `bind-if` or `data-bind-if`.

#### Scenario: conditional child attaches and detaches

Given a layout contains `<text id="warning" bind-if="showWarning">Warning</text>`
When the UI binding context sets `showWarning` to `true`
Then the text widget exists in its parent and in the widget ID lookup cache
When the UI binding context sets `showWarning` to `false`
Then the text widget is removed from its parent and from the widget ID lookup cache.

### Requirement: Rich binding expressions

Template and one-way XML bindings SHALL support safe expressions for common UI
conditions and formatting without requiring custom Go glue code.

#### Scenario: fallback and formatting helpers update from dependencies

Given a layout contains `<text id="greeting">Hello {{upper(user.name || "guest")}}</text>`
When the binding context has no `user.name`
Then the text content is `Hello GUEST`
When the binding context sets `user.name` to `Ada`
Then the text content is `Hello ADA`.

#### Scenario: boolean and arithmetic expressions update from dependencies

Given a layout contains `<text id="summary">{{count + 1}}/{{total}}</text>`
When the binding context sets `count` to `2` and `total` to `5`
Then the text content is `3/5`.

#### Scenario: helper expressions format common UI values

Given a layout contains helper expressions using `len`, `round`, `contains`, `join`, and `format`
When the binding context updates the referenced values
Then the text content updates using the helper results.

### Requirement: XML attribute and style bindings

XML layout authors SHALL be able to bind selected widget attributes and style
properties using `bind-attr-*` and `bind-style-*` attributes.

#### Scenario: style and attribute bindings update widgets

Given a layout contains `<button id="cta" bind-attr-label="label" bind-style-opacity="enabled &amp;&amp; 1 || 0.5">Fallback</button>`
When the binding context sets `label` to `Start` and `enabled` to `false`
Then the button label is `Start`
And the button opacity is `0.5`.

### Requirement: XML event command binding

XML layout authors SHALL be able to dispatch XML event attributes to registered
Go command handlers.

#### Scenario: click command dispatches to registered handler

Given a layout contains `<button id="save" onClick="saveGame">Save</button>`
And Go registers a `saveGame` command handler
When the button is clicked
Then the registered command handler receives the button widget.

