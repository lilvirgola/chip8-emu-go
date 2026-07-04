package cpu

func (c *CPU) screenWidth() int {
	if c.HighRes {
		return 128
	}
	return 64
}

func (c *CPU) screenHeight() int {
	if c.HighRes {
		return 64
	}
	return 32
}

func (c *CPU) scrollAmount(base int) int {
	if !c.HighRes && c.Quirks.HalfScroll {
		return base / 2
	}
	return base
}

func (c *CPU) scrollUp(n int) {
	n = c.scrollAmount(n)
	w, h := c.screenWidth(), c.screenHeight()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := y*128 + x
			srcY := y + n
			if c.Quirks.Wrapping {
				srcY = srcY % h
				c.Display[idx] = c.Display[srcY*128+x]
			} else if srcY < h {
				c.Display[idx] = c.Display[srcY*128+x]
			} else {
				c.Display[idx] = false
			}
		}
	}
	c.DrawFlag = true
}

func (c *CPU) scrollDown(n int) {
	n = c.scrollAmount(n)
	w, h := c.screenWidth(), c.screenHeight()
	for y := h - 1; y >= 0; y-- {
		for x := 0; x < w; x++ {
			idx := y*128 + x
			srcY := y - n
			if c.Quirks.Wrapping {
				srcY = ((srcY % h) + h) % h
				c.Display[idx] = c.Display[srcY*128+x]
			} else if srcY >= 0 {
				c.Display[idx] = c.Display[srcY*128+x]
			} else {
				c.Display[idx] = false
			}
		}
	}
	c.DrawFlag = true
}

func (c *CPU) scrollLeft() {
	n := c.scrollAmount(4)
	w, h := c.screenWidth(), c.screenHeight()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := y*128 + x
			srcX := x + n
			if c.Quirks.Wrapping {
				srcX = srcX % w
				c.Display[idx] = c.Display[y*128+srcX]
			} else if srcX < w {
				c.Display[idx] = c.Display[y*128+srcX]
			} else {
				c.Display[idx] = false
			}
		}
	}
	c.DrawFlag = true
}

func (c *CPU) scrollRight() {
	n := c.scrollAmount(4)
	w, h := c.screenWidth(), c.screenHeight()
	for y := 0; y < h; y++ {
		for x := w - 1; x >= 0; x-- {
			idx := y*128 + x
			srcX := x - n
			if c.Quirks.Wrapping {
				srcX = ((srcX % w) + w) % w
				c.Display[idx] = c.Display[y*128+srcX]
			} else if srcX >= 0 {
				c.Display[idx] = c.Display[y*128+srcX]
			} else {
				c.Display[idx] = false
			}
		}
	}
	c.DrawFlag = true
}
