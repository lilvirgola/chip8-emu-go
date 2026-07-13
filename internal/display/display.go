package display

import "github.com/hajimehoshi/ebiten/v2"

const (
	loResW, loResH           = 64, 32
	hiResW, hiResH           = 128, 64
	WindowSizeW, WindowSizeH = 1280, 640
)

type Display struct {
	Plane0  [hiResW * hiResH]bool
	Plane1  [hiResW * hiResH]bool
	HighRes bool

	Colors [4][3]byte // 0=neither, 1=plane0, 2=plane1, 3=both

	buf  []byte
	img  *ebiten.Image
	imgW int
	imgH int
}

func NewDisplay() *Display {
	d := &Display{
		Colors: [4][3]byte{
			{0x00, 0x00, 0x00}, // background
			{0xFF, 0xFF, 0xFF}, // plane0 only
			{0x00, 0xFF, 0xFF}, // plane1 only
			{0xFF, 0xFF, 0x00}, // both
		},
	}
	d.setResolution(loResW, loResH)
	return d
}

func (d *Display) setResolution(w, h int) {
	if d.imgW == w && d.imgH == h {
		return
	}
	d.imgW, d.imgH = w, h
	d.buf = make([]byte, w*h*4)
	d.img = ebiten.NewImage(w, h)
}

func (d *Display) UpdateFromMemory(plane0, plane1 []bool, highRes bool) {
	d.HighRes = highRes

	if highRes {
		copy(d.Plane0[:], plane0)
		copy(d.Plane1[:], plane1)
		return
	}

	for y := 0; y < loResH; y++ {
		srcRow := y * hiResW
		dstRow := y * loResW
		for x := 0; x < loResW; x++ {
			d.Plane0[dstRow+x] = plane0[srcRow+x]
			d.Plane1[dstRow+x] = plane1[srcRow+x]
		}
	}
}

func (d *Display) Update() {
	w, h := loResW, loResH
	if d.HighRes {
		w, h = hiResW, hiResH
	}
	d.setResolution(w, h)

	for y := 0; y < h; y++ {
		row := y * w
		for x := 0; x < w; x++ {
			idx := row + x
			combo := 0
			if d.Plane0[idx] {
				combo |= 1
			}
			if d.Plane1[idx] {
				combo |= 2
			}
			rgb := d.Colors[combo]
			o := idx * 4
			d.buf[o] = rgb[0]
			d.buf[o+1] = rgb[1]
			d.buf[o+2] = rgb[2]
			d.buf[o+3] = 0xFF
		}
	}

	d.img.WritePixels(d.buf)
}

func (d *Display) Size() (int, int) {
	return d.imgW, d.imgH
}

func (d *Display) Image() *ebiten.Image {
	return d.img
}

func (d *Display) Draw(screen *ebiten.Image) {
	d.Update()
	w, h := d.Size()
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Scale(float64(sw)/float64(w), float64(sh)/float64(h))
	opts.Filter = ebiten.FilterNearest
	screen.DrawImage(d.img, opts)
}

func (d *Display) DrawScaled(dst *ebiten.Image, opts *ebiten.DrawImageOptions) {
	d.Update()
	if opts.Filter == 0 {
		opts.Filter = ebiten.FilterNearest
	}
	screen := dst
	screen.DrawImage(d.img, opts)
}

func (d *Display) Clear() {
	for i := range d.Plane0 {
		d.Plane0[i] = false
		d.Plane1[i] = false
	}
}
