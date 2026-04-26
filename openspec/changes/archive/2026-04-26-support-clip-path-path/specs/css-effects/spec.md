## MODIFIED Requirements

### Requirement: Per-corner frame rendering and clip paths

The renderer SHALL preserve per-corner frame geometry and selected CSS clip path
basic shapes when compositing built-in widgets.

#### Scenario: clip path basic shapes

Given a widget style contains `clipPath` with `inset(...)` or `circle(...)`
When the widget is rendered
Then the widget content is masked to the requested basic shape
And unsupported clip-path functions are treated as a no-op.

#### Scenario: clip path path shape

Given a widget style contains `clipPath` with `path("M0 0 L100 0 L50 100 Z")`
When the widget is rendered
Then the widget content is masked using the parsed SVG path commands.
