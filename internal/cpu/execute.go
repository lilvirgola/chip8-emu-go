package cpu

import (
	"fmt"
	"math/rand"
)

func (c *CPU) DecodeAndExecute(opcode uint16) error {
	x := (opcode & 0x0F00) >> 8 // x register index
	y := (opcode & 0x00F0) >> 4 // y register index
	n := opcode & 0x000F        // last nibble (n) value
	nn := byte(opcode & 0x00FF) // last byte (nn) value
	nnn := opcode & 0x0FFF      // last 12 bits (nnn) value

	switch opcode & 0xF000 {
	case 0x0000:
		if opcode&0xFFF0 == 0x00C0 { // 00CN: scroll display down N pixels
			c.scrollDown(int(opcode & 0x000F))
			break
		}
		if opcode&0xFFF0 == 0x00D0 { // 00CN: scroll display up N pixels
			c.scrollUp(int(opcode & 0x000F))
			break
		}
		switch opcode {
		case 0x00E0: // Clear the display
			for i := range c.Display {
				c.Display[i] = false
			}
			c.DrawFlag = true

		case 0x00EE: // Return from subroutine
			if c.SP == 0 {
				return fmt.Errorf("stack underflow")
			}
			c.SP--
			c.PC = c.Stack[c.SP]

		case 0x00FB: // scroll right 4 (or 2 in lores w/ HalfScroll)
			c.scrollRight()

		case 0x00FC: // scroll left 4 (or 2 in lores w/ HalfScroll)
			c.scrollLeft()

		case 0x00FE: // low-res mode
			c.HighRes = false

		case 0x00FF: // high-res mode
			c.HighRes = true

		}

	case 0x1000: // Jump to address NNN
		c.PC = nnn

	case 0x2000: // Call subroutine at NNN
		if c.SP >= 16 {
			return fmt.Errorf("stack overflow")
		}
		c.Stack[c.SP] = c.PC
		c.SP++

		c.PC = nnn

	case 0x3000: // Skip next instruction if Vx == NN
		if c.V[x] == nn {
			c.PC += 2
		}

	case 0x4000: // Skip next instruction if Vx != NN
		if c.V[x] != nn {
			c.PC += 2
		}

	case 0x5000: // Skip next instruction if Vx == Vy
		if n == 0 {
			if c.V[x] == c.V[y] {
				c.PC += 2
			}
		}

	case 0x6000: // Load NN into Vx
		c.V[x] = nn

	case 0x7000: // Add NN to Vx (NOTE: Since V[x] is a byte, overflow wraps around automatically in Go)
		c.V[x] += nn

	case 0x8000:
		switch n { // switch on the last nibble to determine the operation
		case 0: // Vx = Vy
			c.V[x] = c.V[y]

		case 1: // Vx = Vx OR Vy
			c.V[x] |= c.V[y]
			if c.Quirks.VFReset {
				c.V[0xF] = 0
			}

		case 2: // Vx = Vx AND Vy
			c.V[x] &= c.V[y]
			c.V[x] |= c.V[y]
			if c.Quirks.VFReset {
				c.V[0xF] = 0
			}

		case 3: // Vx = Vx XOR Vy
			c.V[x] ^= c.V[y]
			c.V[x] |= c.V[y]
			if c.Quirks.VFReset {
				c.V[0xF] = 0
			}

		case 4: // Vx = Vx + Vy, set VF = carry (eg. addition with carry)
			sum := uint16(c.V[x]) + uint16(c.V[y])
			if sum > 0xFF {
				c.V[0xF] = 1 // Set carry flag
			} else {
				c.V[0xF] = 0
			}
			c.V[x] = byte(sum)

		case 5: // Vx = Vx - Vy, set VF = NOT borrow (eg. subtraction with borrow)
			if c.V[x] >= c.V[y] {
				c.V[0xF] = 1
			} else {
				c.V[0xF] = 0
			}

			c.V[x] -= c.V[y]

		case 6: // Vx = Vx SHR 1, set VF = least significant bit of Vx before the shift
			tmp := c.V[x]
			if c.Quirks.ShiftSetVX {
				tmp = c.V[x]
			} else {
				tmp = c.V[y]
			}
			c.V[0xF] = tmp & 0x1
			c.V[x] = tmp >> 1

		case 7: // Vx = Vy - Vx, set VF = NOT borrow
			if c.V[y] >= c.V[x] {
				c.V[0xF] = 1
			} else {
				c.V[0xF] = 0
			}

			c.V[x] = c.V[y] - c.V[x]

		case 0xE:
			tmp := c.V[x]

			if c.Quirks.ShiftSetVX {
				tmp = c.V[x]
			} else {
				tmp = c.V[y]
			}

			c.V[0xF] = tmp >> 7
			c.V[x] = tmp << 1
		}

	case 0x9000: // Skip next instruction if Vx != Vy
		if n == 0 {
			if c.V[x] != c.V[y] {
				c.PC += 2
			}
		}

	case 0xA000: // Set I = NNN (eg. load index register with address NNN)
		c.I = nnn

	case 0xB000: // jump with offset: PC = NNN + V0
		if c.Quirks.Jumping {
			c.PC = nnn + uint16(c.V[x]) // Bxnn: use Vx (x = high nibble of nnn)
		} else {
			c.PC = nnn + uint16(c.V[0]) // Bnnn: use V0
		}

	case 0xC000: // Set Vx = random byte AND NN
		c.V[x] = byte(rand.Uint32()&0xFF) & nn

	case 0xD000: // draw sprite at (Vx, Vy) with width 8 pixels and height N pixels. Set VF = collision.
		if c.Quirks.DisplayWait && c.WaitingForFrame {
			c.PC -= 2
			return nil
		}

		w, h := c.screenWidth(), c.screenHeight()
		xPos := int(c.V[x]) % w
		yPos := int(c.V[y]) % h
		c.V[0xF] = 0

		height := int(n)
		spriteWidth := 8
		if n == 0 && c.HighRes { // SuperChip big-sprite mode: 16x16
			height = 16
			spriteWidth = 16
		}

		for row := 0; row < height; row++ {
			if !c.Quirks.Wrapping && yPos+row >= h {
				break
			}
			var spriteRow uint16
			if spriteWidth == 16 {
				spriteRow = uint16(c.Memory[c.I+uint16(row*2)])<<8 | uint16(c.Memory[c.I+uint16(row*2+1)])
			} else {
				spriteRow = uint16(c.Memory[c.I+uint16(row)])
			}
			for col := 0; col < spriteWidth; col++ {
				if spriteRow&(1<<(spriteWidth-1-col)) == 0 {
					continue
				}
				if !c.Quirks.Wrapping && xPos+col >= w {
					continue
				}
				xCoord := (xPos + col) % w
				yCoord := (yPos + row) % h
				index := yCoord*128 + xCoord
				if c.Display[index] {
					c.V[0xF] = 1
				}
				c.Display[index] = !c.Display[index]
			}
		}
		c.DrawFlag = true
		if c.Quirks.DisplayWait {
			c.WaitingForFrame = true
		}

	case 0xE000:
		switch nn {
		case 0x9E: // Skip next instruction if key with the value of Vx is pressed
			if c.Keys[c.V[x]] {
				c.PC += 2
			}
		case 0xA1: // Skip next instruction if key with the value of Vx is not pressed
			if !c.Keys[c.V[x]] {
				c.PC += 2
			}
		}

	case 0xF000:
		switch nn {
		case 0x07: // Set Vx = delay timer value
			c.V[x] = c.DelayTimer
		case 0x0A: // Fx0A: wait for a key press, store the key in Vx
			pressed := false
			for i := byte(0); i < 16; i++ {
				if c.Keys[i] {
					c.V[x] = i
					pressed = true
					break
				}
			}
			if !pressed {
				c.PC -= 2 // re-execute this instruction until a key is pressed
			}
		case 0x15: // Set delay timer = Vx
			c.DelayTimer = c.V[x]
		case 0x18: // Set sound timer = Vx
			c.SoundTimer = c.V[x]
		case 0x1E: // Set I = I + Vx
			c.I += uint16(c.V[x])
		case 0x29: // Set I = location of sprite for digit Vx
			c.I = uint16(c.V[x]) * 5 // Each digit sprite is 5 bytes long
		case 0x30: // Fx30: Set I = location of big sprite for digit Vx (SuperChip/XO-CHIP)
			c.I = 0xA0 + uint16(c.V[x])*10 // each big-font digit is 10 bytes long
		case 0x33: // Store BCD representation of Vx in memory locations I, I+1, and I+2
			c.Memory[c.I] = c.V[x] / 100
			c.Memory[c.I+1] = (c.V[x] / 10) % 10
			c.Memory[c.I+2] = c.V[x] % 10
		case 0x55: // Store registers V0 through Vx in memory starting at location I
			for i := uint16(0); i <= x; i++ {
				c.Memory[c.I+i] = c.V[i]
			}

			if c.Quirks.IncrementI {
				c.I += x + 1
			}
		case 0x65:
			for i := uint16(0); i <= x; i++ {
				c.V[i] = c.Memory[c.I+i]
			}

			if c.Quirks.IncrementI {
				c.I += x + 1
			}
		}
	default:
		return fmt.Errorf("unknown opcode %04X", opcode)
	}

	return nil
}
