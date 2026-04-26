## Why

Backgrounds, gradients, and borders already honor individual corner radii, but
box shadows still use one uniform radius. Asymmetric cards therefore cast a
shadow that does not match their visible frame.

## What Changes

- Extend box-shadow rendering to accept per-corner radii.
- Preserve the existing uniform-radius API as a compatibility wrapper.
- Apply spread to each corner radius and keep current blur behavior.
- Add unit and visual fixture coverage for asymmetric-radius shadows.

## Impact

- Affects box-shadow shader uniforms and widget shadow dispatch.
- Keeps existing shadow syntax and multi-shadow behavior.
