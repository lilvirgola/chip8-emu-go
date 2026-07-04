package display

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

type Display struct {
	Pixels [64 * 32]bool
}

func (d *Display) UpdateFromMemory(mem []bool) {
	copy(d.Pixels[:], mem)
}

func (d *Display) Draw(screen *ebiten.Image) {
	for y := 0; y < 32; y++ {
		for x := 0; x < 64; x++ {
			if d.Pixels[y*64+x] {
				screen.Set(x, y, color.White)
			} else {
				screen.Set(x, y, color.Black)
			}
		}
	}
}
