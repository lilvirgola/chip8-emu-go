package emulator

import (
	"image"
	"image/color"
	"log/slog"

	"chip8/internal/cpu"
	cpuImpl "chip8/internal/cpu"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type CPUTypeSelector struct {
	Rect     image.Rectangle
	Types    []cpu.Quirks
	Names    []string
	current  int
	OnChange func(cpu.Quirks)
}

func NewCPUTypeSelector(types []cpu.Quirks, names []string, onChange func(cpu.Quirks)) *CPUTypeSelector {
	return &CPUTypeSelector{
		Types:    types,
		Names:    names,
		OnChange: onChange,
	}
}

func (s *CPUTypeSelector) Contains(x, y int) bool {
	return image.Pt(x, y).In(s.Rect)
}

func (s *CPUTypeSelector) Update() {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		if s.Contains(mx, my) {
			s.current = (s.current + 1) % len(s.Types)
			s.OnChange(s.Types[s.current])
		}
	}
}

func (s *CPUTypeSelector) Draw(dst *ebiten.Image, face text.Face, hovered bool) {
	clr := color.RGBA{60, 60, 60, 220}
	if hovered {
		clr = color.RGBA{90, 90, 90, 220}
	}
	vector.FillRect(dst,
		float32(s.Rect.Min.X), float32(s.Rect.Min.Y),
		float32(s.Rect.Dx()), float32(s.Rect.Dy()),
		clr, false)
	vector.StrokeRect(dst,
		float32(s.Rect.Min.X), float32(s.Rect.Min.Y),
		float32(s.Rect.Dx()), float32(s.Rect.Dy()),
		1, color.White, false)

	label := s.Names[s.current]
	tw, th := text.Measure(label, face, 0)
	op := &text.DrawOptions{}
	cx := float64(s.Rect.Min.X) + (float64(s.Rect.Dx())-tw)/2
	cy := float64(s.Rect.Min.Y) + (float64(s.Rect.Dy())-th)/2
	op.GeoM.Translate(cx, cy)
	text.Draw(dst, label, face, op)
}

func (g *Game) switchCPUType(t cpuImpl.Quirks) {
	g.rebuildCPU(t)
	slog.Debug("CPU type switched", "quirks", t)
}

func (g *Game) rebuildCPU(t cpuImpl.Quirks) {
	newCPU := cpuImpl.NewCPU(t, g.EmuDisplay)
	if g.currentROM != nil {
		newCPU.LoadROM(g.currentROM)
	}
	g.CPU = newCPU
	g.EmuDisplay.Clear()
	g.CPU.DrawFlag = false
	g.showDebug = false
	slog.Debug("CPU rebuilt", "quirks", t, "romSize", len(g.currentROM))
}
