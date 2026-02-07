# AGENTS.md — Coding Agent Guidelines for ebitenui-xml

## Project Overview

A **data-driven UI framework** for [Ebitengine](https://ebitengine.org/) in Go.
XML defines layout structure, CSS-like JSON defines styling. The core library lives
in the `ui/` package (flat, ~20 files). Demo apps and visual test tools live under `cmd/`.

**Module**: `github.com/ulgerang/ebitenui-xml`
**Go version**: 1.25.5
**Key dependency**: `github.com/hajimehoshi/ebiten/v2` v2.9.8
**Local sibling dep**: `../ebiten-ertp` (via `replace` directive in go.mod)

## Build / Run / Test Commands

```bash
# Build the project (verify compilation)
go build ./...

# Run the main demo
go run .

# Run a specific demo/tool
go run ./cmd/demo_extended/
go run ./cmd/demo_svg/
go run ./cmd/game1984/

# Build a specific tool binary
go build -o css_testloop.exe ./cmd/css_testloop
go build -o converter.exe    ./tools/css_compare/cmd/converter
go build -o pixeldiff.exe    ./tools/css_compare/cmd/pixeldiff

# Vet (no linter config exists — use go vet)
go vet ./...
```

### Testing

**There are NO unit tests (`*_test.go`) in this project.** The project relies on
visual pixel-comparison regression testing instead:

```bash
# CSS visual regression test loop
go run ./cmd/css_testloop -mode render  -out ebiten.png       # render Ebiten screenshot
go run ./cmd/css_testloop -mode html    -out reference.html   # generate HTML reference
go run ./cmd/css_testloop -mode compare -browser browser.png -ebiten ebiten.png -out report.html

# SVG visual regression test loop (same pattern)
go run ./cmd/svg_testloop -mode render  -out ebiten_svg.png
go run ./cmd/svg_testloop -mode html    -out svg_reference.html
go run ./cmd/svg_testloop -mode compare -browser browser.png -ebiten ebiten_svg.png -out report.html

# Pixel diff tool
go run ./tools/css_compare/cmd/pixeldiff/ expected.png actual.png diff.png
```

If you add new functionality, verify it compiles with `go build ./...` and consider
adding a test case in `cmd/css_testloop/testcases.go` or `cmd/svg_testloop/testcases.go`.

### Running a Single Visual Test Case

There is no built-in single-test runner. To test a specific feature visually,
create a minimal `main.go` under `cmd/` that loads a small XML+JSON snippet
and calls `ebiten.RunGame(...)`. See `cmd/demo_extended/main.go` for an example.

## Code Style Guidelines

### Formatting

- **Use `gofmt`** — tabs for indentation, standard Go formatting.
- **No hard line-length limit**, but keep lines under ~120 characters where practical.
- Align struct field tags and comments in columns when grouped:
  ```go
  Opacity    float64 `json:"opacity"`    // 0-1
  BoxShadow  string  `json:"boxShadow"`  // "offsetX offsetY blur spread color"
  ```

### Imports

Two groups separated by a blank line: **stdlib first, then third-party/internal**.
No import aliases. Alphabetical within groups.

```go
import (
    "fmt"
    "image/color"
    "strings"

    "github.com/ulgerang/ebitenui-xml/ui"
    "github.com/hajimehoshi/ebiten/v2"
)
```

Use `_ "embed"` only when `//go:embed` directives are present in the file.

### Naming Conventions

| Category              | Convention        | Examples                                      |
|-----------------------|-------------------|-----------------------------------------------|
| Package names         | lowercase, short  | `ui`, `main`                                  |
| File names            | `snake_case.go`   | `svg_path.go`, `widgets_extended.go`           |
| Exported types        | `PascalCase`      | `BaseWidget`, `LayoutEngine`, `StyleEngine`    |
| Unexported types      | `camelCase`       | `paddingRaw`, `bindingEntry`, `calcToken`      |
| Exported functions    | `PascalCase`      | `NewPanel()`, `ParseSelector()`, `DrawGradient()` |
| Unexported functions  | `camelCase`       | `parseColor()`, `mergeStyles()`, `hueToRGB()`  |
| Constants (exported)  | `PascalCase`      | `LayoutRow`, `StateNormal`                     |
| Enum-style constants  | Type-prefixed     | `AnimationNormal`, `SelectorTypeClass`, `UnitPx` |
| Constructors          | `New<TypeName>()` | `NewBaseWidget()`, `NewLayoutEngine()`         |

### Error Handling

1. **In library code (`ui/`)**: Return wrapped errors with `fmt.Errorf("context: %w", err)`.
   Never `panic()`. Never `log.Fatal` inside the library.
2. **In app entry points (`cmd/*/main.go`)**: Use `log.Fatal(err)` or `log.Fatalf(...)`.
3. **Silent zero-value defaults for optional parsing**: When parsing optional CSS
   values where a zero default is acceptable, `_` is intentional:
   `f, _ := strconv.ParseFloat(val, 64)  // 0.0 on failure is fine`

### Type & Interface Patterns

- **Central `Widget` interface** in `types.go` — all widgets implement it.
- **Embed `*BaseWidget`** in concrete widget structs. Override only `Draw` and
  widget-specific methods (e.g., `type Button struct { *BaseWidget; Label string }`).
- **String-typed enums** with `const` blocks for CSS-mapped values (e.g., `LayoutRow = "row"`).
- **Int-typed enums** with `iota` for internal state (e.g., `StateNormal`, `StateHover`).
- **Go generics** for reusable data structures: `Observable[T]`, `ListBinding[T]`.

### Comments & Documentation

- **Godoc-style** comments on all exported types and functions.
- **Section separators** using `// ===...===` lines for major blocks within a file.
- **Inline comments** on struct fields for units, formats, or ranges.

### File Organization

Within each `.go` file, order: package → imports → types/structs → constants →
constructors (`New...()`) → methods (grouped by receiver) → unexported helpers.

Files are organized by **domain** (not one-type-per-file). Key files in `ui/`:
`types.go` (interfaces, `Style`), `widget.go` (`BaseWidget`), `widgets.go` (Panel,
Button, Text), `widgets_extended.go` (Toggle, Dropdown, Modal), `layout.go` (flexbox),
`style.go` (style engine), `effects.go` (gradients, shadows), `parser.go` (XML),
`animation.go`, `binding.go`, `variables.go`, `selector.go`, `svg.go`, `ui.go` (facade).

### Architecture Notes

- **Widget creation** uses a factory pattern: XML tag name → `WidgetFactory.createWidget()` switch.
- **Style cascade**: type selectors → class selectors (`.`) → ID selectors (`#`),
  with state-specific overrides (`:hover`, `:active`).
- **Rendering pipeline** (per widget): box-shadow → background/gradient → border →
  content → outline → children (recursive).
- **Concurrency**: `sync.RWMutex` is used in `Observable[T]` and binding types.
  The rendering and update loops are single-threaded (Ebiten game loop).

## Project Structure

```
├── ui/                    # Core library (all rendering, layout, styling)
├── cmd/                   # Executables: demos, test loops, tools
│   ├── css_testloop/      # CSS visual regression tests
│   ├── svg_testloop/      # SVG visual regression tests
│   ├── demo_extended/     # Extended widgets demo
│   ├── demo_svg/          # SVG demo
│   └── game1984/          # Full game demo
├── tools/css_compare/     # HTML converter + pixel diff tool
├── assets/                # Demo XML layouts + JSON style files
├── docs/                  # Documentation (Korean)
└── main.go                # Default demo entry point
```
