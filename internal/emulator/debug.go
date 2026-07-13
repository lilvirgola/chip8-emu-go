package emulator

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

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
