## Why

`clipPath` already supports inset, circle, and polygon. SVG path parsing exists
elsewhere in the renderer, so CSS `path(...)` can be added without introducing
a separate geometry stack.

## What Changes

- Parse `clipPath: path("...")` and `clipPath: path('...')`.
- Reuse SVG path command parsing and existing offscreen clip-mask compositing.
- Treat empty or malformed path strings as a no-op.
- Add parser/runtime smoke coverage and a visual testloop case.

## Impact

- Affects CSS clip-path parsing and visual test fixtures.
- Does not add CSS fill-rule syntax.
