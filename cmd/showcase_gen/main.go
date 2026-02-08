package main

import (
	"fmt"
	"os"
)

func main() {
	xml := `
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

	style := `
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
		"padding": {"left": 20},
		"justify": "center",
		"borderBottomWidth": 2,
		"borderBottom": "#38bdf8"
	},
	"#title": {
		"color": "#f8fafc",
		"fontSize": 20
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
		"height": 40,
		"padding": {"left": 15},
		"justify": "center",
		"borderRadius": 6,
		"color": "#94a3b8"
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
		"width": 450,
		"height": 320,
		"background": "#1e293b",
		"borderRadius": 12,
		"borderWidth": 1,
		"border": "#334155",
		"padding": {"all": 20},
		"direction": "column",
		"gap": 20,
		"boxShadow": "0 10 25 0 rgba(0,0,0,0.5)"
	},
	"#card-header": {
		"color": "#f8fafc",
		"fontSize": 18
	},
	"#stats": {
		"direction": "row",
		"gap": 15
	},
	".stat-box": {
		"flexGrow": 1,
		"background": "#0f172a",
		"borderRadius": 8,
		"padding": {"all": 12},
		"direction": "column",
		"align": "center",
		"gap": 5
	},
	".stat-label": { "color": "#64748b", "fontSize": 12 },
	".stat-val": { "color": "#38bdf8", "fontSize": 18 },
	"#progress-container": {
		"height": 10,
		"background": "#0f172a",
		"borderRadius": 5,
		"overflow": "hidden"
	},
	"#progress-bar": {
		"width": 280,
		"height": 10,
		"background": "linear-gradient(90deg, #38bdf8, #818cf8)"
	},
	"#footer-actions": {
		"direction": "row",
		"justify": "end",
		"gap": 10
	},
	".btn-primary": {
		"width": 100,
		"height": 36,
		"background": "#38bdf8",
		"borderRadius": 6,
		"justify": "center",
		"align": "center",
		"color": "#ffffff"
	},
	".btn-secondary": {
		"width": 100,
		"height": 36,
		"borderWidth": 1,
		"border": "#64748b",
		"borderRadius": 6,
		"justify": "center",
		"align": "center",
		"color": "#94a3b8"
	}
}
`

	html := fmt.Sprintf(`<!DOCTYPE html>
<html><head><meta charset="UTF-8">
<style>
	body { margin: 0; background: #000; display: flex; justify-content: center; align-items: center; height: 100vh; font-family: sans-serif; }
	#main { width: 800px; height: 600px; background: #0f172a; display: flex; flex-direction: column; overflow: hidden; }
	#header { height: 60px; background: linear-gradient(90deg, #1e293b, #334155); padding: 0 20px; display: flex; align-items: center; border-bottom: 2px solid #38bdf8; box-sizing: border-box; }
	#title { color: #f8fafc; font-size: 20px; font-weight: bold; }
	#content { flex: 1; display: flex; flex-direction: row; }
	#sidebar { width: 200px; background: #1e293b; padding: 10px; display: flex; flex-direction: column; gap: 8px; border-right: 1px solid #334155; box-sizing: border-box; }
	.nav-item { height: 40px; padding: 0 15px; display: flex; align-items: center; border-radius: 6px; color: #94a3b8; font-size: 14px; }
	.nav-item.active { background: #38bdf8; color: #ffffff; }
	#viewport { flex: 1; background: #0f172a; padding: 30px; display: flex; justify-content: center; align-items: center; }
	#card { width: 450px; height: 320px; background: #1e293b; border-radius: 12px; border: 1px solid #334155; padding: 20px; display: flex; flex-direction: column; gap: 20px; box-shadow: 0 10px 25px 0 rgba(0,0,0,0.5); box-sizing: border-box; }
	#card-header { color: #f8fafc; font-size: 18px; font-weight: bold; }
	#stats { display: flex; flex-direction: row; gap: 15px; }
	.stat-box { flex: 1; background: #0f172a; border-radius: 8px; padding: 12px; display: flex; flex-direction: column; align-items: center; gap: 5px; }
	.stat-label { color: #64748b; font-size: 12px; }
	.stat-val { color: #38bdf8; font-size: 18px; font-weight: bold; }
	#progress-container { height: 10px; background: #0f172a; border-radius: 5px; overflow: hidden; }
	#progress-bar { width: 65%%; height: 100%%; background: linear-gradient(90deg, #38bdf8, #818cf8); }
	#footer-actions { display: flex; flex-direction: row; justify-content: flex-end; gap: 10px; }
	.btn-primary { width: 100px; height: 36px; background: #38bdf8; border-radius: 6px; display: flex; justify-content: center; align-items: center; color: #ffffff; font-size: 13px; font-weight: bold; }
	.btn-secondary { width: 100px; height: 36px; border: 1px solid #64748b; border-radius: 6px; display: flex; justify-content: center; align-items: center; color: #94a3b8; font-size: 13px; box-sizing: border-box; }
</style>
</head>
<body>
	<div id="main">
		<div id="header"><div id="title">ANTIGRAVITY OS</div></div>
		<div id="content">
			<div id="sidebar">
				<div class="nav-item active">Dashboard</div>
				<div class="nav-item">Analytics</div>
				<div class="nav-item">Security</div>
				<div class="nav-item">Settings</div>
			</div>
			<div id="viewport">
				<div id="card">
					<div id="card-header">System Performance</div>
					<div id="stats">
						<div class="stat-box">
							<div class="stat-label">CPU</div>
							<div class="stat-val">12%%</div>
						</div>
						<div class="stat-box">
							<div class="stat-label">GPU</div>
							<div class="stat-val">45%%</div>
						</div>
						<div class="stat-box">
							<div class="stat-label">RAM</div>
							<div class="stat-val">2.4GB</div>
						</div>
					</div>
					<div id="progress-container"><div id="progress-bar"></div></div>
					<div id="footer-actions">
						<div class="btn-primary">OPTIMIZE</div>
						<div class="btn-secondary">REBOOT</div>
					</div>
				</div>
			</div>
		</div>
	</div>
</body></html>`)

	os.WriteFile("cmd/showcase/index.html", []byte(html), 0644)

	ebitenCode := fmt.Sprintf(`package main
import (
	"log"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/ulgerang/ebitenui-xml/ui"
	"image/png"
	"os"
	"time"
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
		f, _ := os.Create("ebiten_showcase.png")
		png.Encode(f, screen)
		f.Close()
		log.Println("Saved ebiten_showcase.png")
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
	xmlData := %%q
	styleData := %%q
	
	engine := ui.New(800, 600)
	if err := engine.LoadLayout(xmlData); err != nil { log.Fatal(err) }
	if err := engine.LoadStyles(styleData); err != nil { log.Fatal(err) }
	
	game := &Game{engine: engine}
	
	ebiten.SetWindowSize(800, 600)
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
`, xml, style)

	os.WriteFile("cmd/showcase/main.go", []byte(ebitenCode), 0644)
}
