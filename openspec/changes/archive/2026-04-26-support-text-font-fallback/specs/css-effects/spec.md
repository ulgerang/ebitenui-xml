## ADDED Requirements

### Requirement: Text font fallback

The UI runtime SHALL resolve declared font families through explicit registered
font faces or font sources before falling back to the configured default font.

#### Scenario: registered font family fallback

Given a text widget style declares `fontFamily: "Missing, Registered, sans-serif"`
And the UI has a font face registered as `Registered`
When fonts are applied
Then the text widget uses the registered face
And missing families do not prevent fallback to later entries.

#### Scenario: deterministic default font fallback

Given a text widget style declares only unregistered font families
When fonts are applied
Then the text widget uses the configured default font face or source.
