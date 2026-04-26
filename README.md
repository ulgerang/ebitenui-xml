# EbitenUI-XML

A **data-driven UI framework** for [Ebitengine](https://ebitengine.org/) (Ebiten) in Go, featuring **XML for layout structure** and **CSS-like JSON for styling**.

## ✨ Features

- **XML Layouts** - Declarative UI structure with familiar HTML-like syntax
- **CSS-like JSON Styling** - Flexible styling with selectors, classes, and cascading
- **Flexbox Layout** - Row/column direction, justify, align, gap, and wrap
- **HTML-like XML Semantics** - Semantic aliases, forms, fieldsets, radio
  groups, focus traversal, and DOM-like query helpers
- **SVG Rendering** - Built-in vector graphics with native SVG parser
- **Icon Library** - 20+ built-in icons (arrow, check, heart, star, etc.)
- **9-Slice Scaling** - Scalable UI backgrounds
- **Data Binding** - Reactive state management
- **XML Data Binding** - `bind-text`, `bind-value`, `bind-checked`,
  `bind-visible`, `bind-enabled`, `bind-repeat`, `bind-if`, template
  interpolation, attribute/style bindings, and command events
- **CSS Effects** - Blur filters, transforms, transitions, animations,
  multi-shadows, text-shadow blur, z-index, runtime overflow scrolling, and
  `inset`/`circle`/`polygon` clip-path support
- **Animation System** - Built-in presets plus JSON custom keyframes and
  literal CSS `@keyframes`
- **Extended Widgets** - Toggle, Dropdown, Modal, Toast, Spinner, and more

## 🚀 Quick Start

```go
package main

import (
    "log"
    "github.com/ulgerang/ebitenui-xml/ui"
    "github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
    uiManager *ui.UI
}

func NewGame() *Game {
    g := &Game{}
    g.uiManager = ui.New(800, 600)
    
    // Load styles
    g.uiManager.LoadStyles(`{
        "#root": {
            "background": "#1E1E28",
            "padding": {"all": 20},
            "direction": "column",
            "gap": 10
        },
        ".btn": {
            "background": "#4A90D9",
            "padding": {"all": 12},
            "borderRadius": 8
        }
    }`)
    
    // Load layout
    g.uiManager.LoadLayout(`
        <panel id="root">
            <button class="btn" onClick="handleClick">Click Me</button>
            <text>Hello, EbitenUI-XML!</text>
        </panel>
    `)
    
    return g
}

func (g *Game) Update() error {
    g.uiManager.Update()
    return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
    g.uiManager.Draw(screen)
}

func (g *Game) Layout(w, h int) (int, int) {
    return 800, 600
}

func main() {
    ebiten.SetWindowSize(800, 600)
    ebiten.SetWindowTitle("EbitenUI-XML Demo")
    if err := ebiten.RunGame(NewGame()); err != nil {
        log.Fatal(err)
    }
}
```

## 📦 Installation

```bash
go get github.com/ulgerang/ebitenui-xml
```

## 🎨 XML Elements

| Element | Description |
|---------|-------------|
| `<panel>` | Container with flexbox layout |
| `<text>` | Text display |
| `<button>` | Clickable button |
| `<image>` | Image display |
| `<input>` | Text input field |
| `<textarea>` | Multi-line text input |
| `<checkbox>` | Checkbox toggle |
| `<slider>` | Value slider |
| `<icon>` | SVG icon |
| `<svg>` | Custom SVG content |
| `<toggle>` | Toggle switch |
| `<dropdown>` | Dropdown select |
| `<modal>` | Modal dialog |
| `<toast>` | Toast notification |
| `<spinner>` | Loading spinner |

## 🎯 Styling Properties

```json
{
    "#elementId": {
        "width": 200,
        "height": 100,
        "background": "#FF5722",
        "color": "#FFFFFF",
        "fontSize": 16,
        "padding": {"top": 10, "right": 20, "bottom": 10, "left": 20},
        "margin": {"all": 5},
        "borderRadius": 8,
        "direction": "row",
        "justify": "center",
        "align": "center",
        "gap": 10
    }
}
```

Top-level JSON `keyframes` can register custom animations:

```json
{
    "keyframes": {
        "popIn": {
            "from": {"opacity": 0, "transform": "translateY(16px) scale(0.9)"},
            "to": {"opacity": 1, "transform": "translateY(0) scale(1)"}
        }
    },
    "styles": {
        ".card": {"animation": "popIn 350ms ease-out 1"}
    }
}
```

## 🔤 Built-in Icons

```xml
<icon icon="arrow-left" stroke="#4A90D9" stroke-width="2"/>
<icon icon="check" stroke="#4CAF50"/>
<icon icon="heart" stroke="#E91E63"/>
<icon icon="star" stroke="#FFC107"/>
```

Available icons: `arrow-left`, `arrow-right`, `arrow-up`, `arrow-down`, `check`, `x`, `plus`, `minus`, `heart`, `star`, `home`, `settings`, `user`, `search`, `menu`, `bell`, `mail`, `calendar`, `clock`, `trash`

## 📖 Documentation

See the [docs](./docs/) folder for detailed documentation:
- [Cheatsheet](./docs/CHEATSHEET.md)
- [Extended Widgets](./docs/WIDGETS_EXTENDED.md)
- [Support Matrix](./docs/SUPPORT_MATRIX.md)
- [CSS Visual Compare](./docs/CSS_COMPARE.md)

## 📝 License

MIT License
