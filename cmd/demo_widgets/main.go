package main

import (
	_ "embed"
	"fmt"
	"image/color"
	"log"

	"github.com/ulgerang/ebitenui-xml/ui"
	"github.com/hajimehoshi/bitmapfont/v4"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

//go:embed assets/layout_extended.xml
var layoutXML string

//go:embed assets/styles_extended.json
var stylesJSON string

const (
	screenWidth  = 800
	screenHeight = 600
)

type Game struct {
	ui         *ui.UI
	radioGroup *ui.RadioGroup
	toast      *ui.Toast
	modal      *ui.Modal
}

func NewGame() (*Game, error) {
	g := &Game{}

	// Create UI manager
	g.ui = ui.New(screenWidth, screenHeight)

	// Load font using Ebiten's built-in bitmap font
	fontData := text.NewGoXFace(bitmapfont.FaceEA)
	g.ui.DefaultFontFace = fontData

	// Load styles first
	if err := g.ui.LoadStyles(stylesJSON); err != nil {
		return nil, fmt.Errorf("failed to load styles: %w", err)
	}

	// Load layout
	if err := g.ui.LoadLayout(layoutXML); err != nil {
		return nil, fmt.Errorf("failed to load layout: %w", err)
	}

	// Create auxiliary widgets not in XML
	g.createAuxiliaryWidgets(fontData)

	// Set up event handlers
	g.setupEventHandlers()

	return g, nil
}

func (g *Game) createAuxiliaryWidgets(fontFace text.Face) {
	// Create toast for notifications
	g.toast = ui.NewToast("notification", "")
	g.toast.FontFace = fontFace

	// Create modal dialog
	g.modal = ui.NewModal("confirm-modal", "Confirmation")
	g.modal.Content = "Are you sure you want to proceed?\nThis action cannot be undone."
	g.modal.FontFace = fontFace

	// Add modal buttons
	confirmBtn := ui.NewButton("confirm-btn", "Confirm")
	confirmBtn.FontFace = fontFace
	confirmBtn.Style().BackgroundColor = color.RGBA{39, 174, 96, 255}
	confirmBtn.OnClick(func() {
		g.modal.Close()
		g.toast.ToastType = "success"
		g.toast.Message = "Action confirmed!"
		g.toast.Show()
	})
	g.modal.AddButton(confirmBtn)

	cancelBtn := ui.NewButton("cancel-btn", "Cancel")
	cancelBtn.FontFace = fontFace
	cancelBtn.Style().BackgroundColor = color.RGBA{231, 76, 60, 255}
	cancelBtn.OnClick(func() {
		g.modal.Close()
	})
	g.modal.AddButton(cancelBtn)

	// Create radio group for difficulty radios
	g.radioGroup = ui.NewRadioGroup("difficulty")
	g.radioGroup.OnChange = func(value string) {
		log.Printf("Difficulty changed to: %s", value)
	}

	// Link radio buttons that were loaded from XML
	radioIDs := []string{"rb-easy", "rb-normal", "rb-hard"}
	for _, id := range radioIDs {
		if w := g.ui.GetWidget(id); w != nil {
			if rb, ok := w.(*ui.RadioButton); ok {
				g.radioGroup.AddButton(rb)
			}
		}
	}
	g.radioGroup.SetValue("normal")
}

func (g *Game) setupEventHandlers() {
	// Toggle handlers
	if w := g.ui.GetWidget("sound-toggle"); w != nil {
		if t, ok := w.(*ui.Toggle); ok {
			t.OnChange = func(checked bool) {
				status := "OFF"
				if checked {
					status = "ON"
				}
				log.Printf("Sound: %s", status)
			}
		}
	}

	// Dropdown handler
	if w := g.ui.GetWidget("resolution"); w != nil {
		if d, ok := w.(*ui.Dropdown); ok {
			d.OnChange = func(idx int, value string) {
				log.Printf("Resolution changed to: %s", value)
			}
		}
	}

	// Modal button
	if btn := g.ui.GetButton("open-modal-btn"); btn != nil {
		btn.OnClick(func() {
			log.Println("Opening modal...")
			g.modal.Open()
		})
	}

	// Toast buttons
	toastTypes := []string{"success", "warning", "error", "info"}
	for _, t := range toastTypes {
		toastType := t
		if btn := g.ui.GetButton("toast-" + t); btn != nil {
			btn.OnClick(func() {
				log.Printf("Showing %s toast", toastType)
				g.toast.ToastType = toastType
				g.toast.Message = fmt.Sprintf("%s notification!", toastType)
				g.toast.Show()
			})
		}
	}
}

func (g *Game) Update() error {
	g.ui.Update()

	// Update spinner animation
	if w := g.ui.GetWidget("loading-spinner"); w != nil {
		if s, ok := w.(*ui.Spinner); ok {
			s.Update()
		}
	}

	// Update toast
	if g.toast != nil {
		g.toast.Update()
	}

	// Update dropdowns
	if w := g.ui.GetWidget("resolution"); w != nil {
		if d, ok := w.(*ui.Dropdown); ok {
			d.Update()
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Clear background
	screen.Fill(color.RGBA{26, 26, 46, 255})

	// Draw UI
	g.ui.Draw(screen)

	// Draw overlays (toast, modal)
	if g.toast != nil && g.toast.IsVisible {
		g.toast.Draw(screen)
	}

	if g.modal != nil && g.modal.IsOpen {
		g.modal.Draw(screen)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Extended UI Widgets Demo")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game, err := NewGame()
	if err != nil {
		log.Fatalf("Failed to create game: %v", err)
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
