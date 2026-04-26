## Why

`fontFamily` is parsed but the runtime always falls back to the default face or
source. XML/CSS authors need deterministic family fallback selection without
pulling in platform font discovery.

## What Changes

- Add explicit UI font-family registration APIs for fixed faces and scalable
  Go text sources.
- Resolve comma-separated `fontFamily` lists in order, with quoted family names
  supported.
- Keep deterministic default font behavior when no requested family is
  registered.
- Add regression coverage for fallback selection and line-height measurement.

## Impact

- Affects UI font assignment only.
- Does not implement OS/browser font discovery or complex shaping fallback.
