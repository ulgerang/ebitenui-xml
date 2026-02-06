package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

const (
	screenWidth  = 480
	screenHeight = 640
	playerSpeed  = 4.0
	bulletSpeed  = 8.0
	enemySpeed   = 2.0
)

// Game represents the main game state
type Game struct {
	player      *Player
	bullets     []*Bullet
	enemies     []*Enemy
	particles   []*Particle
	score       int
	lives       int
	gameOver    bool
	spawnTimer  int
	level       int
	screenShake float64
	fontFace    text.Face
}

// Player represents the player ship
type Player struct {
	x, y       float64
	width      float64
	height     float64
	shootTimer int
	invincible int
}

// Bullet represents a player bullet
type Bullet struct {
	x, y    float64
	vx, vy  float64
	active  bool
	isEnemy bool
}

// Enemy represents an enemy ship
type Enemy struct {
	x, y       float64
	vx, vy     float64
	width      float64
	height     float64
	health     int
	shootTimer int
	enemyType  int // 0: basic, 1: zigzag, 2: boss
	phase      float64
	active     bool
}

// Particle for visual effects
type Particle struct {
	x, y    float64
	vx, vy  float64
	life    int
	maxLife int
	size    float64
	color   color.RGBA
}

func NewGame() *Game {
	g := &Game{
		player: &Player{
			x:      screenWidth / 2,
			y:      screenHeight - 80,
			width:  32,
			height: 40,
		},
		lives: 3,
		level: 1,
	}

	// Create basic font face for score
	g.fontFace = text.NewGoXFace(basicfont.Face7x13)

	return g
}

func (g *Game) Update() error {
	if g.gameOver {
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			// Restart game
			*g = *NewGame()
		}
		return nil
	}

	g.updatePlayer()
	g.updateBullets()
	g.updateEnemies()
	g.updateParticles()
	g.checkCollisions()
	g.spawnEnemies()

	// Reduce screen shake
	g.screenShake *= 0.9

	return nil
}

func (g *Game) updatePlayer() {
	p := g.player

	// Movement
	if ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		p.x -= playerSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		p.x += playerSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		p.y -= playerSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		p.y += playerSpeed
	}

	// Clamp position
	p.x = clamp(p.x, p.width/2, screenWidth-p.width/2)
	p.y = clamp(p.y, p.height/2, screenHeight-p.height/2)

	// Shooting
	p.shootTimer--
	if p.shootTimer <= 0 && (ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsKeyPressed(ebiten.KeyZ)) {
		g.bullets = append(g.bullets, &Bullet{
			x:      p.x,
			y:      p.y - p.height/2,
			vx:     0,
			vy:     -bulletSpeed,
			active: true,
		})
		p.shootTimer = 8
	}

	// Invincibility timer
	if p.invincible > 0 {
		p.invincible--
	}
}

func (g *Game) updateBullets() {
	activeBullets := g.bullets[:0]
	for _, b := range g.bullets {
		if !b.active {
			continue
		}
		b.x += b.vx
		b.y += b.vy

		// Remove off-screen bullets
		if b.y < -10 || b.y > screenHeight+10 || b.x < -10 || b.x > screenWidth+10 {
			continue
		}
		activeBullets = append(activeBullets, b)
	}
	g.bullets = activeBullets
}

func (g *Game) updateEnemies() {
	activeEnemies := g.enemies[:0]
	for _, e := range g.enemies {
		if !e.active {
			continue
		}

		e.phase += 0.05

		switch e.enemyType {
		case 0: // Basic - straight down
			e.y += e.vy
		case 1: // Zigzag
			e.y += e.vy
			e.x += math.Sin(e.phase*2) * 3
		case 2: // Heavy - slow but shoots
			e.y += e.vy * 0.5
			e.shootTimer--
			if e.shootTimer <= 0 {
				// Shoot at player
				dx := g.player.x - e.x
				dy := g.player.y - e.y
				dist := math.Sqrt(dx*dx + dy*dy)
				if dist > 0 {
					g.bullets = append(g.bullets, &Bullet{
						x:       e.x,
						y:       e.y + e.height/2,
						vx:      dx / dist * 4,
						vy:      dy / dist * 4,
						active:  true,
						isEnemy: true,
					})
				}
				e.shootTimer = 60
			}
		}

		// Clamp X position
		e.x = clamp(e.x, e.width/2, screenWidth-e.width/2)

		// Remove off-screen enemies
		if e.y > screenHeight+50 {
			continue
		}

		activeEnemies = append(activeEnemies, e)
	}
	g.enemies = activeEnemies
}

func (g *Game) updateParticles() {
	activeParticles := g.particles[:0]
	for _, p := range g.particles {
		p.x += p.vx
		p.y += p.vy
		p.vy += 0.1 // gravity
		p.life--
		if p.life > 0 {
			activeParticles = append(activeParticles, p)
		}
	}
	g.particles = activeParticles
}

func (g *Game) checkCollisions() {
	// Player bullets vs enemies
	for _, b := range g.bullets {
		if !b.active || b.isEnemy {
			continue
		}
		for _, e := range g.enemies {
			if !e.active {
				continue
			}
			if isColliding(b.x, b.y, 4, 4, e.x-e.width/2, e.y-e.height/2, e.width, e.height) {
				b.active = false
				e.health--
				if e.health <= 0 {
					e.active = false
					g.score += 100 * (e.enemyType + 1)
					g.spawnExplosion(e.x, e.y, 15)
					g.screenShake = 5
				} else {
					g.spawnExplosion(e.x, e.y, 5)
				}
			}
		}
	}

	// Enemy bullets vs player
	if g.player.invincible <= 0 {
		for _, b := range g.bullets {
			if !b.active || !b.isEnemy {
				continue
			}
			p := g.player
			if isColliding(b.x, b.y, 6, 6, p.x-p.width/2, p.y-p.height/2, p.width, p.height) {
				b.active = false
				g.lives--
				g.player.invincible = 120
				g.spawnExplosion(p.x, p.y, 20)
				g.screenShake = 10
				if g.lives <= 0 {
					g.gameOver = true
				}
			}
		}

		// Enemies vs player
		for _, e := range g.enemies {
			if !e.active {
				continue
			}
			p := g.player
			if isColliding(e.x-e.width/2, e.y-e.height/2, e.width, e.height,
				p.x-p.width/2, p.y-p.height/2, p.width, p.height) {
				e.active = false
				g.lives--
				g.player.invincible = 120
				g.spawnExplosion(p.x, p.y, 25)
				g.screenShake = 15
				if g.lives <= 0 {
					g.gameOver = true
				}
			}
		}
	}
}

func (g *Game) spawnEnemies() {
	g.spawnTimer--
	if g.spawnTimer <= 0 {
		enemyType := rand.Intn(3)
		health := 1
		width, height := 30.0, 30.0

		switch enemyType {
		case 1:
			width, height = 25, 25
		case 2:
			health = 3
			width, height = 40, 40
		}

		g.enemies = append(g.enemies, &Enemy{
			x:          rand.Float64()*(screenWidth-60) + 30,
			y:          -40,
			vy:         enemySpeed + rand.Float64()*0.5,
			width:      width,
			height:     height,
			health:     health,
			shootTimer: 60 + rand.Intn(60),
			enemyType:  enemyType,
			active:     true,
		})

		// Spawn rate increases with score
		g.spawnTimer = 60 - min(g.score/500, 40)
	}
}

func (g *Game) spawnExplosion(x, y float64, count int) {
	for i := 0; i < count; i++ {
		angle := rand.Float64() * math.Pi * 2
		speed := rand.Float64()*3 + 1
		g.particles = append(g.particles, &Particle{
			x:       x,
			y:       y,
			vx:      math.Cos(angle) * speed,
			vy:      math.Sin(angle) * speed,
			life:    30 + rand.Intn(20),
			maxLife: 50,
			size:    rand.Float64()*4 + 2,
			color: color.RGBA{
				R: uint8(200 + rand.Intn(55)),
				G: uint8(100 + rand.Intn(100)),
				B: uint8(rand.Intn(50)),
				A: 255,
			},
		})
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Apply screen shake
	shakeX := (rand.Float64() - 0.5) * g.screenShake
	shakeY := (rand.Float64() - 0.5) * g.screenShake

	// Dark background with gradient effect
	screen.Fill(color.RGBA{10, 10, 20, 255})

	// Draw starfield background
	drawStarfield(screen)

	// Draw game elements with shake offset
	g.drawParticles(screen, shakeX, shakeY)
	g.drawBullets(screen, shakeX, shakeY)
	g.drawEnemies(screen, shakeX, shakeY)
	g.drawPlayer(screen, shakeX, shakeY)

	// Draw UI
	g.drawUI(screen)
}

func drawStarfield(screen *ebiten.Image) {
	// Simple static starfield
	starColor := color.RGBA{100, 100, 120, 255}
	for i := 0; i < 50; i++ {
		x := float32((i * 37) % screenWidth)
		y := float32((i * 73) % screenHeight)
		size := float32(1 + (i % 3))
		vector.DrawFilledRect(screen, x, y, size, size, starColor, false)
	}
}

func (g *Game) drawPlayer(screen *ebiten.Image, ox, oy float64) {
	p := g.player

	// Skip drawing during invincibility blink
	if p.invincible > 0 && (p.invincible/4)%2 == 0 {
		return
	}

	x := float32(p.x + ox)
	y := float32(p.y + oy)

	// Draw player ship using vector graphics (SVG-like)
	// Main body - triangle
	bodyColor := color.RGBA{80, 180, 255, 255}
	glowColor := color.RGBA{120, 200, 255, 200}

	// Glow effect
	drawTriangle(screen, x, y-20, x-18, y+18, x+18, y+18, glowColor)

	// Main body
	drawTriangle(screen, x, y-16, x-14, y+14, x+14, y+14, bodyColor)

	// Cockpit
	cockpitColor := color.RGBA{200, 220, 255, 255}
	drawTriangle(screen, x, y-8, x-6, y+4, x+6, y+4, cockpitColor)

	// Engine flames
	flameColor := color.RGBA{255, 150, 50, 255}
	flameHeight := float32(8 + rand.Float32()*4)
	drawTriangle(screen, x-8, y+14, x-4, y+14+flameHeight, x, y+14, flameColor)
	drawTriangle(screen, x, y+14, x+4, y+14+flameHeight, x+8, y+14, flameColor)

	// Wing details
	wingColor := color.RGBA{60, 140, 220, 255}
	vector.DrawFilledRect(screen, x-16, y+5, 8, 3, wingColor, false)
	vector.DrawFilledRect(screen, x+8, y+5, 8, 3, wingColor, false)
}

func (g *Game) drawBullets(screen *ebiten.Image, ox, oy float64) {
	for _, b := range g.bullets {
		if !b.active {
			continue
		}

		x := float32(b.x + ox)
		y := float32(b.y + oy)

		if b.isEnemy {
			// Enemy bullet - red/orange
			vector.DrawFilledCircle(screen, x, y, 5, color.RGBA{255, 100, 50, 200}, false)
			vector.DrawFilledCircle(screen, x, y, 3, color.RGBA{255, 200, 100, 255}, false)
		} else {
			// Player bullet - cyan laser
			vector.DrawFilledRect(screen, x-2, y-8, 4, 16, color.RGBA{100, 200, 255, 200}, false)
			vector.DrawFilledRect(screen, x-1, y-6, 2, 12, color.RGBA{200, 255, 255, 255}, false)
		}
	}
}

func (g *Game) drawEnemies(screen *ebiten.Image, ox, oy float64) {
	for _, e := range g.enemies {
		if !e.active {
			continue
		}

		x := float32(e.x + ox)
		y := float32(e.y + oy)
		w := float32(e.width)
		h := float32(e.height)

		switch e.enemyType {
		case 0: // Basic enemy - diamond shape
			basicColor := color.RGBA{255, 80, 80, 255}
			drawDiamond(screen, x, y, w/2, h/2, basicColor)
			// Core
			vector.DrawFilledCircle(screen, x, y, w/6, color.RGBA{255, 200, 200, 255}, false)

		case 1: // Zigzag enemy - triangle pointing down
			zigzagColor := color.RGBA{255, 180, 50, 255}
			drawTriangle(screen, x, y+h/2, x-w/2, y-h/2, x+w/2, y-h/2, zigzagColor)
			// Eye
			vector.DrawFilledCircle(screen, x, y-h/4, 4, color.RGBA{50, 50, 50, 255}, false)

		case 2: // Heavy enemy - hexagon-like
			heavyColor := color.RGBA{180, 80, 200, 255}
			drawHexagon(screen, x, y, w/2, heavyColor)
			// Health indicator
			for i := 0; i < e.health; i++ {
				vector.DrawFilledCircle(screen, x+float32(i-1)*10, y, 4, color.RGBA{255, 255, 100, 255}, false)
			}
		}
	}
}

func (g *Game) drawParticles(screen *ebiten.Image, ox, oy float64) {
	for _, p := range g.particles {
		alpha := uint8(float64(p.life) / float64(p.maxLife) * float64(p.color.A))
		c := color.RGBA{p.color.R, p.color.G, p.color.B, alpha}
		x := float32(p.x + ox)
		y := float32(p.y + oy)
		size := float32(p.size * float64(p.life) / float64(p.maxLife))
		vector.DrawFilledCircle(screen, x, y, size, c, false)
	}
}

func (g *Game) drawUI(screen *ebiten.Image) {
	// Score
	scoreText := fmt.Sprintf("SCORE: %d", g.score)
	op := &text.DrawOptions{}
	op.GeoM.Translate(10, 10)
	op.ColorScale.ScaleWithColor(color.RGBA{255, 255, 255, 255})
	text.Draw(screen, scoreText, g.fontFace, op)

	// Lives
	for i := 0; i < g.lives; i++ {
		x := float32(screenWidth - 30 - i*25)
		y := float32(20)
		// Mini player ship as life icon
		drawTriangle(screen, x, y-8, x-8, y+8, x+8, y+8, color.RGBA{80, 180, 255, 255})
	}

	// Game Over screen
	if g.gameOver {
		// Dim overlay
		vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{0, 0, 0, 180}, false)

		// Game Over text
		gameOverOp := &text.DrawOptions{}
		gameOverOp.GeoM.Translate(screenWidth/2-50, screenHeight/2-20)
		gameOverOp.ColorScale.ScaleWithColor(color.RGBA{255, 50, 50, 255})
		text.Draw(screen, "GAME OVER", g.fontFace, gameOverOp)

		// Final score
		finalScoreOp := &text.DrawOptions{}
		finalScoreOp.GeoM.Translate(screenWidth/2-60, screenHeight/2+10)
		finalScoreOp.ColorScale.ScaleWithColor(color.RGBA{255, 255, 255, 255})
		text.Draw(screen, fmt.Sprintf("Final Score: %d", g.score), g.fontFace, finalScoreOp)

		// Restart instruction
		restartOp := &text.DrawOptions{}
		restartOp.GeoM.Translate(screenWidth/2-80, screenHeight/2+40)
		restartOp.ColorScale.ScaleWithColor(color.RGBA{200, 200, 200, 255})
		text.Draw(screen, "Press SPACE to restart", g.fontFace, restartOp)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

// Helper drawing functions (SVG-like vector primitives)

func drawTriangle(screen *ebiten.Image, x1, y1, x2, y2, x3, y3 float32, c color.Color) {
	var path vector.Path
	path.MoveTo(x1, y1)
	path.LineTo(x2, y2)
	path.LineTo(x3, y3)
	path.Close()

	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
	r, g, b, a := c.RGBA()
	for i := range vs {
		vs[i].SrcX = 1
		vs[i].SrcY = 1
		vs[i].ColorR = float32(r) / 0xffff
		vs[i].ColorG = float32(g) / 0xffff
		vs[i].ColorB = float32(b) / 0xffff
		vs[i].ColorA = float32(a) / 0xffff
	}
	screen.DrawTriangles(vs, is, emptyImage, &ebiten.DrawTrianglesOptions{
		AntiAlias: true,
	})
}

func drawDiamond(screen *ebiten.Image, cx, cy, rx, ry float32, c color.Color) {
	var path vector.Path
	path.MoveTo(cx, cy-ry) // Top
	path.LineTo(cx+rx, cy) // Right
	path.LineTo(cx, cy+ry) // Bottom
	path.LineTo(cx-rx, cy) // Left
	path.Close()

	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
	r, g, b, a := c.RGBA()
	for i := range vs {
		vs[i].SrcX = 1
		vs[i].SrcY = 1
		vs[i].ColorR = float32(r) / 0xffff
		vs[i].ColorG = float32(g) / 0xffff
		vs[i].ColorB = float32(b) / 0xffff
		vs[i].ColorA = float32(a) / 0xffff
	}
	screen.DrawTriangles(vs, is, emptyImage, &ebiten.DrawTrianglesOptions{
		AntiAlias: true,
	})
}

func drawHexagon(screen *ebiten.Image, cx, cy, r float32, c color.Color) {
	var path vector.Path
	for i := 0; i < 6; i++ {
		angle := float64(i)*math.Pi/3 - math.Pi/2
		x := cx + r*float32(math.Cos(angle))
		y := cy + r*float32(math.Sin(angle))
		if i == 0 {
			path.MoveTo(x, y)
		} else {
			path.LineTo(x, y)
		}
	}
	path.Close()

	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
	cr, cg, cb, ca := c.RGBA()
	for i := range vs {
		vs[i].SrcX = 1
		vs[i].SrcY = 1
		vs[i].ColorR = float32(cr) / 0xffff
		vs[i].ColorG = float32(cg) / 0xffff
		vs[i].ColorB = float32(cb) / 0xffff
		vs[i].ColorA = float32(ca) / 0xffff
	}
	screen.DrawTriangles(vs, is, emptyImage, &ebiten.DrawTrianglesOptions{
		AntiAlias: true,
	})
}

// Empty image for DrawTriangles
var emptyImage = func() *ebiten.Image {
	img := ebiten.NewImage(3, 3)
	img.Fill(color.White)
	return img
}()

// Utility functions

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func isColliding(x1, y1, w1, h1, x2, y2, w2, h2 float64) bool {
	return x1 < x2+w2 && x1+w1 > x2 && y1 < y2+h2 && y1+h1 > y2
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("1984 - Vector Shooter")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
