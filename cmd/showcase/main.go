package main

import (
	"bytes"
	"log"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/ulgerang/ebitenui-xml/ui"
	"image/png"
)

type Game struct {
	engine *ui.UI
	done   bool
}

func (g *Game) Update() error {
	g.engine.Update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.engine.Draw(screen)
	if !g.done {
		f, _ := os.Create("cmd/showcase/ebiten_showcase.png")
		png.Encode(f, screen)
		f.Close()
		log.Println("Saved cmd/showcase/ebiten_showcase.png")
		g.done = true
		go func() {
			time.Sleep(500 * time.Millisecond)
			os.Exit(0)
		}()
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 800, 600
}

func main() {
	xmlData := `
<panel id="main">
	<panel id="header">
		<text id="title">ANTIGRAVITY OS</text>
	</panel>
	
	<panel id="content">
		<panel id="sidebar">
			<panel class="nav-item active"><text>Dashboard</text></panel>
			<panel class="nav-item"><text>Analytics</text></panel>
			<panel class="nav-item"><text>Security</text></panel>
			<panel class="nav-item"><text>Settings</text></panel>
		</panel>
		
		<panel id="viewport">
			<panel id="card">
				<panel id="card-header">
					<text>System Performance</text>
				</panel>
				<panel id="stats">
					<panel class="stat-box">
						<text class="stat-label">CPU</text>
						<text class="stat-val">12%</text>
					</panel>
					<panel class="stat-box">
						<text class="stat-label">GPU</text>
						<text class="stat-val">45%</text>
					</panel>
					<panel class="stat-box">
						<text class="stat-label">RAM</text>
						<text class="stat-val">2.4GB</text>
					</panel>
				</panel>
				<panel id="progress-container">
					<panel id="progress-bar"></panel>
				</panel>
				<panel id="footer-actions">
					<panel class="btn-primary"><text>OPTIMIZE</text></panel>
					<panel class="btn-secondary"><text>REBOOT</text></panel>
				</panel>
			</panel>
		</panel>
	</panel>
</panel>`

	styleData := `
{
	"#main": {
		"width": 800,
		"height": 600,
		"background": "#0f172a",
		"direction": "column"
	},
	"#header": {
		"height": 60,
		"background": "linear-gradient(90deg, #1e293b, #334155)",
		"padding": {"left": 20, "right": 20},
		"justify": "center",
		"borderBottomWidth": 2,
		"borderBottom": "#38bdf8",
		"verticalAlign": "center"
	},
	"#title": {
		"color": "#f8fafc",
		"fontSize": 20,
		"fontWeight": "bold"
	},
	"#content": {
		"flexGrow": 1,
		"direction": "row"
	},
	"#sidebar": {
		"width": 200,
		"background": "#1e293b",
		"padding": {"all": 10},
		"gap": 8,
		"borderRightWidth": 1,
		"borderRight": "#334155"
	},
	".nav-item": {
		"height": 44,
		"padding": {"left": 20, "right": 20},
		"justify": "center",
		"borderRadius": 8,
		"color": "#94a3b8",
		"fontSize": 14,
		"verticalAlign": "center"
	},
	".nav-item.active": {
		"background": "#38bdf8",
		"color": "#ffffff"
	},
	"#viewport": {
		"flexGrow": 1,
		"background": "#0f172a",
		"padding": {"all": 30},
		"justify": "center",
		"align": "center"
	},
	"#card": {
		"width": 480,
		"height": 380,
		"background": "#1e293b",
		"borderRadius": 16,
		"borderWidth": 1,
		"border": "#334155",
		"padding": {"all": 28},
		"direction": "column",
		"gap": 24,
		"boxShadow": "0 12 35 0 rgba(0,0,0,0.4)"
	},
	"#card-header": {
		"color": "#f8fafc",
		"fontSize": 20,
		"fontWeight": "bold"
	},
	"#stats": {
		"direction": "row",
		"align": "stretch",
		"gap": 15
	},
	".stat-box": {
		"flexGrow": 1,
		"background": "#0f172a",
		"borderRadius": 10,
		"padding": {"all": 16},
		"direction": "column",
		"align": "center",
		"justify": "center",
		"gap": 8
	},
	".stat-label": { "color": "#64748b", "fontSize": 14, "textAlign": "center" },
	".stat-val": { "color": "#38bdf8", "fontSize": 22, "fontWeight": "bold", "textAlign": "center" },
	"#progress-container": {
		"height": 12,
		"background": "#0f172a",
		"borderRadius": 6,
		"overflow": "hidden",
		"align": "stretch"
	},
	"#progress-bar": {
		"width": 266,
		"height": 12,
		"background": "linear-gradient(90deg, #38bdf8, #818cf8)"
	},
	"#footer-actions": {
		"height": 40,
		"direction": "row",
		"justify": "end",
		"align": "center",
		"gap": 10
	},
	".btn-primary": {
		"width": 120,
		"height": 40,
		"background": "#38bdf8",
		"borderRadius": 8,
		"justify": "center",
		"align": "center",
		"color": "#ffffff",
		"fontSize": 14,
		"fontWeight": "bold",
		"verticalAlign": "center"
	},
	".btn-secondary": {
		"width": 120,
		"height": 40,
		"borderWidth": 1,
		"border": "#64748b",
		"borderRadius": 8,
		"justify": "center",
		"align": "center",
		"color": "#94a3b8",
		"fontSize": 14,
		"verticalAlign": "center"
	}
}
`

	engine := ui.New(800, 600)

	// Load a real TTF font from Windows
	fontData, err := os.ReadFile("C:/Windows/Fonts/segoeui.ttf")
	if err != nil {
		fontData, err = os.ReadFile("C:/Windows/Fonts/arial.ttf")
	}
	if err != nil {
		fontData, _ = os.ReadFile("C:/Windows/Fonts/malgun.ttf")
	}

	if fontData != nil {
		source, _ := text.NewGoTextFaceSource(bytes.NewReader(fontData))
		engine.DefaultFont = source
	}

	// Load bold font for fontWeight: "bold"
	boldFontData, err := os.ReadFile("C:/Windows/Fonts/segoeuib.ttf")
	if err != nil {
		boldFontData, err = os.ReadFile("C:/Windows/Fonts/arialbd.ttf")
	}
	if err != nil {
		boldFontData, _ = os.ReadFile("C:/Windows/Fonts/malgunbd.ttf")
	}
	if boldFontData != nil {
		boldSource, _ := text.NewGoTextFaceSource(bytes.NewReader(boldFontData))
		engine.DefaultBoldFont = boldSource
	}

	if err := engine.LoadLayout(xmlData); err != nil {
		log.Fatal(err)
	}
	if err := engine.LoadStyles(styleData); err != nil {
		log.Fatal(err)
	}

	game := &Game{engine: engine}

	ebiten.SetWindowSize(800, 600)
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
