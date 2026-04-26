# css-effects Specification

## Purpose
TBD - created by archiving change add-declarative-data-binding. Update Purpose after archive.
## Requirements
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

The renderer SHALL support common CSS shadow forms for box and text rendering.

#### Scenario: multi-shadow parsing

Given a style contains comma-separated `boxShadow` or `textShadow` values
When the style is parsed
Then each shadow entry is preserved in order.

#### Scenario: blurred text shadow

Given a text widget has a `textShadow` with a blur radius
When the widget is rendered
Then the shadow is drawn through a blurred offscreen pass before the text.

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

#### Scenario: per-corner box shadow

Given a widget has different border radius values per corner and a box shadow
When the widget is rendered
Then the box shadow follows the same per-corner radius geometry.

### Requirement: CSS rule subset for visual properties

The style engine SHALL accept simple CSS selector blocks for common visual and
effect properties.

#### Scenario: class rule sets effects

Given a stylesheet contains `.badge { transform: scale(1.2); filter: blur(2px); opacity: 0.5; }`
When the stylesheet is loaded
Then the `.badge` style records transform, filter, and opacity declarations.

#### Scenario: descendant selector sets visual style

Given a stylesheet contains `.card .title { color: #ffffff; }`
When the stylesheet is loaded and the layout is applied
Then only title descendants of `.card` receive the declared color.

#### Scenario: sibling selector sets visual style

Given a stylesheet contains `.label + .value { color: #ffffff; }`
When the stylesheet is loaded and the layout is applied
Then only matching following sibling widgets receive the declared color.

#### Scenario: same-specificity CSS source order wins

Given a stylesheet contains `.primary { color: #111111; }` before `.accent { color: #ffffff; }`
And a widget has both `primary` and `accent` classes
When styles are applied
Then the later same-specificity rule wins for overlapping properties.

#### Scenario: pseudo state rule applies only in state

Given a stylesheet contains `.button:hover { color: #ffffff; }`
When the stylesheet is loaded and the layout is applied
Then the hover color is stored on the matching style's hover state
And the base color is not changed until the widget enters hover state.

#### Scenario: parser keeps function punctuation inside declarations

Given a stylesheet contains `clip-path: path('M0 0 L10 0; L10 10 Z')`
When the stylesheet is loaded
Then the declaration is parsed as one value despite the semicolon inside quotes.

#### Scenario: UI style loading resolves CSS variables

Given the UI variable `--primary` is set to `#ffffff`
When `UI.LoadCSS` or `UI.LoadStyles` receives a declaration containing `var(--primary)`
Then the resulting widget style uses the resolved value.

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

