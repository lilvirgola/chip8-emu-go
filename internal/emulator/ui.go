package emulator

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func (g *Game) layoutButtons(sw, sh int) {
	btnW, btnH := 180, 32
	gap := 8
	margin := 16

	totalW := btnW*3 + gap*2
	startX := (sw - totalW) / 2
	y := sh - btnH - margin

	g.debugToggleBtn.Rect = image.Rect(
		startX,
		y,
		startX+btnW,
		y+btnH,
	)

	g.loadROMBtn.Rect = image.Rect(
		startX+btnW+gap,
		y,
		startX+btnW*2+gap,
		y+btnH,
	)

	g.cpuTypeSelector.Rect = image.Rect(
		startX+btnW*2+gap*2,
		y,
		startX+btnW*3+gap*2,
		y+btnH,
	)
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

func drawCenteredTextBox(dst *ebiten.Image, s string, face text.Face, lineHeight, screenW, screenH float64) {
	tw, th := text.Measure(s, face, lineHeight)

	padding := 8.0

	x := (screenW - tw) / 2
	y := (screenH - th) / 2

	vector.FillRect(
		dst,
		float32(x-padding),
		float32(y-padding),
		float32(tw+padding*2),
		float32(th+padding*2),
		color.RGBA{0, 0, 0, 180},
		false,
	)

	op := &text.DrawOptions{}
	op.LineSpacing = lineHeight
	op.GeoM.Translate(x, y)
	text.Draw(dst, s, face, op)
}
