## Why

The existing CSS loader supports only simple viewport `min/max-width` and
`min/max-height` media blocks. Authoring responsive Ebiten UI skins still needs
the common browser patterns that select by media type, orientation, or a comma
query list.

## What Changes

- Extend `UI.LoadCSS` media evaluation to accept `screen` and `all` media types.
- Add `orientation: landscape|portrait` query matching from the UI viewport.
- Add comma-separated media query lists where any matching query applies the
  nested rules.
- Keep unsupported media types and unknown features as non-matching no-ops.

## Impact

- Affects `ui/ui.go` media query preprocessing.
- Adds focused unit coverage for the expanded media query subset.
- Does not implement the full CSS media grammar.
