package emulator

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Button struct {
	Rect    image.Rectangle
	Label   string
	OnClick func()
}

func (b *Button) Contains(x, y int) bool {
	return image.Pt(x, y).In(b.Rect)
}

func (b *Button) Update() {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		println("mouse:", mx, my, "button rect:", b.Rect.Min.X, b.Rect.Min.Y, b.Rect.Max.X, b.Rect.Max.Y)

		if b.Contains(mx, my) {
			println("clicked:", b.Label)
			if b.OnClick != nil {
				b.OnClick()
			}
		}
	}
}

func (b *Button) Draw(dst *ebiten.Image, face text.Face, hovered bool) {
	clr := color.RGBA{60, 60, 60, 220}
	if hovered {
		clr = color.RGBA{90, 90, 90, 220}
	}
	vector.FillRect(dst,
		float32(b.Rect.Min.X), float32(b.Rect.Min.Y),
		float32(b.Rect.Dx()), float32(b.Rect.Dy()),
		clr, false)

	vector.StrokeRect(dst,
		float32(b.Rect.Min.X), float32(b.Rect.Min.Y),
		float32(b.Rect.Dx()), float32(b.Rect.Dy()),
		1, color.White, false)

	tw, th := text.Measure(b.Label, face, 0)
	op := &text.DrawOptions{}
	cx := float64(b.Rect.Min.X) + (float64(b.Rect.Dx())-tw)/2
	cy := float64(b.Rect.Min.Y) + (float64(b.Rect.Dy())-th)/2
	op.GeoM.Translate(cx, cy)
	text.Draw(dst, b.Label, face, op)
}
