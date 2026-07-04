package display

import "github.com/hajimehoshi/ebiten/v2"

type Display struct {
	Pixels  [128 * 64]bool
	HighRes bool
	buf     []byte
}

func (d *Display) UpdateFromMemory(mem []bool, highRes bool) {
	copy(d.Pixels[:], mem)
	d.HighRes = highRes
}

func (d *Display) Draw(screen *ebiten.Image) {
	bounds := screen.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	need := w * h * 4
	if len(d.buf) != need {
		d.buf = make([]byte, need)
	}

	const stride = 128 // CPU.Display is always stored at 128-wide stride

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			on := d.Pixels[y*stride+x]
			var v byte
			if on {
				v = 0xFF
			}
			o := (y*w + x) * 4
			d.buf[o], d.buf[o+1], d.buf[o+2], d.buf[o+3] = v, v, v, 0xFF
		}
	}
	screen.WritePixels(d.buf)
}
