# CSS Effects Spec

## ADDED Requirements

### Requirement: CSS filter blur rendering

Widgets SHALL render `filter: blur(Npx)` using the same compositing path as
other CSS filters.

#### Scenario: blur filter creates filtered output

Given a widget style contains `"filter": "blur(4px)"`
When the widget is rendered through the CSS filter pipeline
Then the renderer applies a two-pass blur and returns filtered output.

### Requirement: Full-content transform and filter compositing

Text-like widgets SHALL apply transform, filter, opacity, and animation effects
to the whole widget content, including labels and editable text.

#### Scenario: button transform includes label

Given a button has a transform style
When the button is drawn
Then its background, border, children, and label are composited as one transformed
unit.

### Requirement: State transitions

Widgets SHALL start CSS transitions when their interactive state changes and a
transition declaration exists on the old or new style.

#### Scenario: hover opacity transition

Given a widget has `"transition": "opacity 1s linear"`
And its hover style changes opacity
When the widget enters hover state
Then an opacity transition starts from the current value toward the hover value.

### Requirement: Declarative animation style

Widgets SHALL support a compact JSON style `animation` declaration that starts a
registered animation preset during rendering.

#### Scenario: pulse animation from style

Given a widget style contains `"animation": "pulse 750ms ease-in-out infinite"`
When the widget is drawn
Then the registered `pulse` animation starts with a 750ms duration and infinite
iterations.

#### Scenario: custom JSON keyframes

Given a style JSON document contains a top-level `keyframes` object
When the style engine loads the JSON
Then each named keyframe block is registered as an animation
And animation declarations can reference the custom keyframe name.

### Requirement: Widget-wide compositing coverage

Built-in widgets SHALL route transform, filter, opacity, animation, and clip-path
effects through a whole-widget compositing path where they render custom content
outside the base widget background.

#### Scenario: custom widget content is composited

Given a built-in widget such as a checkbox, slider, dropdown, modal, tooltip,
badge, toast, spinner, image, progress bar, toggle, radio button, or SVG icon
has an effect style
When the widget is drawn
Then the widget draws its complete visual output to an offscreen layer before
the effect is applied.

### Requirement: Shadow lists and blurred text shadows

Widgets SHALL parse and render CSS-like shadow lists for box and text shadows,
including color functions that contain commas.

#### Scenario: multi-shadow parsing

Given a style contains a comma-separated `boxShadow` or `textShadow` value
When the style is parsed
Then each top-level shadow entry is parsed independently
And commas inside color functions such as `rgba(...)` do not split the shadow.

#### Scenario: blurred text shadow

Given a text-bearing widget has a `textShadow` entry with a blur radius
When the widget renders text
Then the shadow text is rendered to an offscreen layer, blurred, and composited
behind the primary text.

### Requirement: Per-corner frame rendering and clip paths

Widgets SHALL support independent corner radii and initial CSS clip paths for
common frame rendering.

#### Scenario: per-corner background and border

Given a widget style contains individual corner radii
When the widget background or uniform border is drawn
Then the renderer uses the individual corner radii instead of only the uniform
radius.

#### Scenario: basic clip path

Given a widget style contains `clipPath` with `inset(...)` or `circle(...)`
When the widget is rendered
Then the renderer masks the composited widget content to that shape.
And unsupported clip-path functions are treated as a no-op.
