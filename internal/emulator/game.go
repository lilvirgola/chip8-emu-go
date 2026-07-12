package emulator

import (
	"bytes"
	"chip8/internal/cpu"
	"chip8/internal/display"
	"chip8/internal/keyboard"
	"fmt"
	"image"
	"image/color"
	"log"

	myAudio "chip8/internal/audio"

	cpuImpl "chip8/internal/cpu"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/gofont/goregular"
)

type Game struct {
	CPU          *cpu.CPU
	EmuDisplay   *display.Display
	DebugDisplay *display.Display
	beepPlayer   *audio.Player
	beepStream   *myAudio.Beep

	cyclesPerFrame int
	debugFace      text.Face
	lineHeight     float64
	showDebug      bool

	// UI elements
	debugToggleBtn *Button
}

func NewGame(cpu *cpu.CPU, display *display.Display, beepStream *myAudio.Beep, beepPlayer *audio.Player, cyclesPerFrame int) *Game {
	if cyclesPerFrame <= 0 {
		cyclesPerFrame = 12 // Default value if not provided or invalid
	}
	if cpu == nil {
		cpu = cpuImpl.NewCPU(cpuImpl.XOChip, display) // Default to XOChip if not provided
	}
	src, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		log.Fatal(err)
	}
	Face := &text.GoTextFace{
		Source: src,
		Size:   20, // pixel size
	}
	m := Face.Metrics()
	lineHeight := m.HAscent + m.HDescent + m.HLineGap
	g := &Game{
		CPU:            cpu,
		EmuDisplay:     display,
		beepStream:     beepStream,
		beepPlayer:     beepPlayer,
		cyclesPerFrame: cyclesPerFrame,
		debugFace:      Face,
		lineHeight:     lineHeight,
	}

	g.debugToggleBtn = &Button{
		Rect:  image.Rect(8, 8, 100, 32),
		Label: "Debug",
		OnClick: func() {
			g.showDebug = !g.showDebug
		},
	}

	return g
}

func (g *Game) Update() error {

	sw, sh := ebiten.WindowSize()
	g.layoutDebugButton(sw, sh)

	g.debugToggleBtn.Update()

	if inpututil.IsKeyJustPressed(ebiten.KeyF9) {
		g.showDebug = !g.showDebug
	}

	for i := 0; i < g.cyclesPerFrame; i++ {
		if err := g.CPU.Cycle(); err != nil {
			return err
		}
	}
	g.CPU.TickTimers()

	if g.CPU.SoundTimer > 0 {
		g.beepStream.Activate()
	} else {
		g.beepStream.Deactivate()
	}

	if g.CPU.DrawFlag {
		g.EmuDisplay.UpdateFromMemory(g.CPU.Display[:], g.CPU.Display2[:], g.CPU.HighRes)
		g.CPU.DrawFlag = false
	}

	keys := keyboard.PollKeys()
	g.CPU.SetKeys(keys)
	return nil
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

func (g *Game) Draw(screen *ebiten.Image) {
	emuW, emuH := g.EmuDisplay.Size()
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()

	g.layoutDebugButton(sw, sh)

	scale := float64(sw) / float64(emuW)
	if s2 := float64(sh) / float64(emuH); s2 < scale {
		scale = s2
	}
	offsetX := (float64(sw) - float64(emuW)*scale) / 2
	offsetY := (float64(sh) - float64(emuH)*scale) / 2

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(offsetX, offsetY)
	op.Filter = ebiten.FilterNearest

	g.EmuDisplay.DrawScaled(screen, op)

	if g.showDebug {
		g.drawDebugOverlay(screen, sw, sh)
	}

	// ui elements part
	mx, my := ebiten.CursorPosition()
	hovered := g.debugToggleBtn.Contains(mx, my)
	g.debugToggleBtn.Draw(screen, g.debugFace, hovered)
}

func (g *Game) drawDebugOverlay(screen *ebiten.Image, w, h int) {
	padding := 8.0

	// top-left
	tl := fmt.Sprintf("FPS: %.1f\nPC: %04X\nI: %04X\nDT: %02X\nST: %02X",
		ebiten.ActualFPS(), g.CPU.PC, g.CPU.I, g.CPU.DelayTimer, g.CPU.SoundTimer)
	drawTextBox(screen, tl, g.debugFace, g.lineHeight, padding, padding, false)

	// top-right
	tr := fmt.Sprintf("V0: %02X \nV1: %02X \nV2: %02X \nV3: %02X \nV4: %02X \nV5: %02X \nV6: %02X \nV7: %02X \nV8: %02X \nV9: %02X \nVA: %02X \nVB: %02X \nVC: %02X \nVD: %02X \nVE: %02X \nVF: %02X",
		g.CPU.V[0x0], g.CPU.V[0x1], g.CPU.V[0x2], g.CPU.V[0x3],
		g.CPU.V[0x4], g.CPU.V[0x5], g.CPU.V[0x6], g.CPU.V[0x7],
		g.CPU.V[0x8], g.CPU.V[0x9], g.CPU.V[0xA], g.CPU.V[0xB],
		g.CPU.V[0xC], g.CPU.V[0xD], g.CPU.V[0xE], g.CPU.V[0xF])
	drawTextBox(screen, tr, g.debugFace, g.lineHeight, float64(w)-padding, padding, true)

	// bottom-left
	bl := fmt.Sprintf("SP: %02X", g.CPU.SP)
	_, blH := text.Measure(bl, g.debugFace, g.debugFace.Metrics().HLineGap)
	drawTextBox(screen, bl, g.debugFace, g.lineHeight, padding, float64(h)-blH-padding, false)

	// bottom-right
	br := fmt.Sprintf("Opcode: %04X", g.CPU.Opcode)
	_, brH := text.Measure(br, g.debugFace, g.debugFace.Metrics().HLineGap)
	drawTextBox(screen, br, g.debugFace, g.lineHeight, float64(w)-padding, float64(h)-brH-padding, true)
}

func (g *Game) layoutDebugButton(sw, sh int) {
	btnW, btnH := 92, 32
	margin := 16

	x := (sw - btnW) / 2
	y := sh - btnH - margin

	g.debugToggleBtn.Rect = image.Rect(x, y, x+btnW, y+btnH)
}

func drawTextBox(dst *ebiten.Image, s string, face text.Face, lineHeight, x, y float64, alignRight bool) {
	tw, th := text.Measure(s, face, lineHeight)
	bx := x
	if alignRight {
		bx = x - tw
	}

	padding := 4.0
	vector.FillRect(dst,
		float32(bx-padding), float32(y-padding),
		float32(tw+padding*2), float32(th+padding*2),
		color.RGBA{0, 0, 0, 160}, false)

	op := &text.DrawOptions{}
	op.LineSpacing = lineHeight
	op.GeoM.Translate(bx, y)
	text.Draw(dst, s, face, op)
}
