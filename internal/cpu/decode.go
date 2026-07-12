package cpu

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
)

type Opcode struct {
	mask        uint16 // Bits to check (1 = check, 0 = ignore/wildcard)
	value       uint16 // Expected value for the checked bits
	function    func(c *CPU, opcode uint16) error
	description string
	asmFormat   string
}

// Helper functions for some opcode fields
func getX(opcode uint16) uint16   { return (opcode & 0x0F00) >> 8 }
func getY(opcode uint16) uint16   { return (opcode & 0x00F0) >> 4 }
func getN(opcode uint16) uint16   { return opcode & 0x000F }
func getNN(opcode uint16) byte    { return byte(opcode & 0x00FF) }
func getNNN(opcode uint16) uint16 { return opcode & 0x0FFF }

// reges for parsing the assembly format strings
var asmRegex = regexp.MustCompile(`%([0-9]*)(\[([0-9]+)\])([A-Za-z])`)

func (op Opcode) FormatASM(opcode uint16) string {
	nnn := opcode & 0x0FFF
	x := (opcode & 0x0F00) >> 8
	y := (opcode & 0x00F0) >> 4
	n := opcode & 0x000F
	nn := opcode & 0x00FF

	args := []uint16{nnn, x, y, n, nn}

	// replace all format verbs manually to avoid fmt.Sprintf's %!(EXTRA ...) error
	// this is **SO** overkill, but it allows us to have a custom format string with indexed arguments
	// and currently i don't see a better way to do it
	return asmRegex.ReplaceAllStringFunc(op.asmFormat, func(match string) string {
		submatches := asmRegex.FindStringSubmatch(match)
		widthStr := submatches[1] // e.g., "03" from %03[1]X
		idxStr := submatches[3]   // e.g., "1" from %[1]X
		verb := submatches[4]     // e.g., "X"
		idx, _ := strconv.Atoi(idxStr)
		val := args[idx-1]
		formatStr := "%" + widthStr + verb
		return fmt.Sprintf(formatStr, val)
	})
}

// opcodes is the dispatch table of all the currently supported opcodes.
// Order matters, though in this set they are mutually exclusive... so not really :)
var opcodes = []Opcode{
	// 0x0000 Family
	{0xFFF0, 0x00C0, op00CN, "Scroll display down N pixels", "SCRD %[4]X"},
	{0xFFF0, 0x00D0, op00DN, "Scroll display up N pixels", "SCRU %[4]X"},
	{0xFFFF, 0x00E0, op00E0, "Clear screen", "CLS"},
	{0xFFFF, 0x00EE, op00EE, "Return from subroutine", "RET"},
	{0xFFFF, 0x00FB, op00FB, "Scroll right", "SCRR"},
	{0xFFFF, 0x00FC, op00FC, "Scroll left", "SCRL"},
	{0xFFFF, 0x00FE, op00FE, "Set low-res mode", "LOW"},
	{0xFFFF, 0x00FF, op00FF, "Set high-res mode", "HIGH"},

	// 0x1000 to 0x4000
	{0xF000, 0x1000, op1NNN, "Jump to address NNN", "JP %03[1]X"},
	{0xF000, 0x2000, op2NNN, "Call subroutine at NNN", "CALL %03[1]X"},
	{0xF000, 0x3000, op3XNN, "Skip next if Vx == NN", "SE V%[2]X, %02[5]X"},
	{0xF000, 0x4000, op4XNN, "Skip next if Vx != NN", "SNE V%[2]X, %02[5]X"},

	// 0x5000 Family
	{0xF00F, 0x5000, op5XY0, "Skip next if Vx == Vy", "SE V%[2]X, V%[3]X"},
	{0xF00F, 0x5002, op5XY2, "Save Vx..Vy to memory at I", "SAVE V%[2]X..V%[3]X"},
	{0xF00F, 0x5003, op5XY3, "Load Vx..Vy from memory at I", "LOAD V%[2]X..V%[3]X"},
	{0xF00F, 0x5004, op5XY4, "Load palette", "PALETTE %[2]X, %[3]X"},

	// 0x6000 to 0x7000
	{0xF000, 0x6000, op6XNN, "Set Vx = NN", "LD V%[2]X, %02[5]X"},
	{0xF000, 0x7000, op7XNN, "Add NN to Vx", "ADD V%[2]X, %02[5]X"},

	// 0x8000 Family
	{0xF00F, 0x8000, op8XY0, "Set Vx = Vy", "LD V%[2]X, V%[3]X"},
	{0xF00F, 0x8001, op8XY1, "Vx = Vx OR Vy", "OR V%[2]X, V%[3]X"},
	{0xF00F, 0x8002, op8XY2, "Vx = Vx AND Vy", "AND V%[2]X, V%[3]X"},
	{0xF00F, 0x8003, op8XY3, "Vx = Vx XOR Vy", "XOR V%[2]X, V%[3]X"},
	{0xF00F, 0x8004, op8XY4, "Vx = Vx + Vy, VF = carry", "ADD V%[2]X, V%[3]X"},
	{0xF00F, 0x8005, op8XY5, "Vx = Vx - Vy, VF = NOT borrow", "SUB V%[2]X, V%[3]X"},
	{0xF00F, 0x8006, op8XY6, "Vx = Vx SHR 1", "SHR V%[2]X"},
	{0xF00F, 0x8007, op8XY7, "Vx = Vy - Vx, VF = NOT borrow", "SUBN V%[2]X, V%[3]X"},
	{0xF00F, 0x800E, op8XYE, "Vx = Vx SHL 1", "SHL V%[2]X"},

	// 0x9000 to 0xC000
	{0xF00F, 0x9000, op9XY0, "Skip next if Vx != Vy", "SNE V%[2]X, V%[3]X"},
	{0xF000, 0xA000, opANNN, "Set I = NNN", "LD I, %03[1]X"},
	{0xF000, 0xB000, opBNNN, "Jump to NNN + V0", "JP %03[1]X"},
	{0xF000, 0xC000, opCXNN, "Vx = random AND NN", "RND V%[2]X, %02[5]X"},

	// 0xD000
	{0xF000, 0xD000, opDXYN, "Draw sprite", "DRW V%[2]X, V%[3]X, %[4]X"},

	// 0xE000 Family
	{0xF0FF, 0xE09E, opEX9E, "Skip next if key Vx pressed", "SKP V%[2]X"},
	{0xF0FF, 0xE0A1, opEXA1, "Skip next if key Vx not pressed", "SKNP V%[2]X"},

	// 0xF000 Family
	{0xFFFF, 0xF000, opF000, "Set I = long NNNN", "LD I, LONG"},
	{0xF0FF, 0xF001, opFX01, "Select drawing plane", "PLANE %[2]X"},
	{0xF0FF, 0xF002, opFX02, "Set audio pattern", "AUDIO"},
	{0xF0FF, 0xF007, opFX07, "Vx = delay timer", "LD V%[2]X, DT"},
	{0xF0FF, 0xF00A, opFX0A, "Wait for key press", "LD V%[2]X, K"},
	{0xF0FF, 0xF015, opFX15, "Set delay timer = Vx", "LD DT, V%[2]X"},
	{0xF0FF, 0xF018, opFX18, "Set sound timer = Vx", "LD ST, V%[2]X"},
	{0xF0FF, 0xF01E, opFX1E, "I = I + Vx", "ADD I, V%[2]X"},
	{0xF0FF, 0xF029, opFX29, "I = sprite location for digit Vx", "LD F, V%[2]X"},
	{0xF0FF, 0xF030, opFX30, "I = big sprite location for digit Vx", "LD HF, V%[2]X"},
	{0xF0FF, 0xF033, opFX33, "Store BCD of Vx at I", "LD B, V%[2]X"},
	{0xF0FF, 0xF03A, opFX3A, "Set audio pitch", "PITCH V%[2]X"},
	{0xF0FF, 0xF055, opFX55, "Store V0..Vx in memory at I", "LD [I], V%[2]X"},
	{0xF0FF, 0xF065, opFX65, "Load V0..Vx from memory at I", "LD V%[2]X, [I]"},
	{0xF0FF, 0xF075, opFX75, "Save V0..Vx to flags", "LD R, V%[2]X"},
	{0xF0FF, 0xF085, opFX85, "Load V0..Vx from flags", "LD V%[2]X, R"},
}

// Execution Functions eg the actual implementation of the opcodes

func op00CN(c *CPU, opcode uint16) error {
	c.scrollDown(int(getN(opcode)))
	return nil
}

func op00DN(c *CPU, opcode uint16) error {
	c.scrollUp(int(getN(opcode)))
	return nil
}

func op00E0(c *CPU, opcode uint16) error {
	if c.SelectedPlanes&1 != 0 {
		for i := range c.Display {
			c.Display[i] = false
		}
	}
	if c.SelectedPlanes&2 != 0 {
		for i := range c.Display2 {
			c.Display2[i] = false
		}
	}
	c.DrawFlag = true
	return nil
}

func op00EE(c *CPU, opcode uint16) error {
	if c.SP == 0 {
		return fmt.Errorf("stack underflow")
	}
	c.SP--
	c.PC = c.Stack[c.SP]
	return nil
}

func op00FB(c *CPU, opcode uint16) error {
	c.scrollRight()
	return nil
}

func op00FC(c *CPU, opcode uint16) error {
	c.scrollLeft()
	return nil
}

func op00FE(c *CPU, opcode uint16) error {
	c.HighRes = false
	return nil
}

func op00FF(c *CPU, opcode uint16) error {
	c.HighRes = true
	return nil
}

func op1NNN(c *CPU, opcode uint16) error {
	c.PC = getNNN(opcode)
	return nil
}

func op2NNN(c *CPU, opcode uint16) error {
	if c.SP >= 255 {
		return fmt.Errorf("stack overflow")
	}
	c.Stack[c.SP] = c.PC
	c.SP++
	c.PC = getNNN(opcode)
	return nil
}

func op3XNN(c *CPU, opcode uint16) error {
	if c.V[getX(opcode)] == getNN(opcode) {
		c.PC += 2
	}
	return nil
}

func op4XNN(c *CPU, opcode uint16) error {
	if c.V[getX(opcode)] != getNN(opcode) {
		c.PC += 2
	}
	return nil
}

func op5XY0(c *CPU, opcode uint16) error {
	if c.V[getX(opcode)] == c.V[getY(opcode)] {
		c.PC += 2
	}
	return nil
}

func op5XY2(c *CPU, opcode uint16) error {
	x, y := getX(opcode), getY(opcode)
	if x <= y {
		for i := x; i <= y; i++ {
			c.Memory[c.I+(i-x)] = c.V[i]
		}
	} else {
		for i := x; ; i-- {
			c.Memory[c.I+(x-i)] = c.V[i]
			if i == y {
				break
			}
		}
	}
	return nil
}

func op5XY3(c *CPU, opcode uint16) error {
	x, y := getX(opcode), getY(opcode)
	if x <= y {
		for i := x; i <= y; i++ {
			c.V[i] = c.Memory[c.I+(i-x)]
		}
	} else {
		for i := x; ; i-- {
			c.V[i] = c.Memory[c.I+(x-i)]
			if i == y {
				break
			}
		}
	}
	return nil
}

func op5XY4(c *CPU, opcode uint16) error { // apparently this opcode is only a suggestion made by Timendus to load a palette of colors, not actually used, but we implement it for completeness
	x := getX(opcode) // Start plane bitmask
	y := getY(opcode) // End plane bitmask

	// Load colors from memory at I
	for combo := x; combo <= y; combo++ {
		rgb332 := c.Memory[c.I+uint16(combo-x)]

		// Convert RGB332 to RGB888
		r := (rgb332 >> 5) & 0x07 // Top 3 bits
		g := (rgb332 >> 2) & 0x07 // Middle 3 bits
		b := rgb332 & 0x03        // Bottom 2 bits

		// Scale to 8-bit values
		r8 := byte((uint16(r) * 255) / 7)
		g8 := byte((uint16(g) * 255) / 7)
		b8 := byte((uint16(b) * 255) / 3)

		// Update the display color for this combo
		if c.DisplayRef != nil {
			c.DisplayRef.Colors[combo] = [3]byte{r8, g8, b8}
		}
	}

	return nil
}

func op6XNN(c *CPU, opcode uint16) error {
	c.V[getX(opcode)] = getNN(opcode)
	return nil
}

func op7XNN(c *CPU, opcode uint16) error {
	c.V[getX(opcode)] += getNN(opcode)
	return nil
}

func op8XY0(c *CPU, opcode uint16) error {
	c.V[getX(opcode)] = c.V[getY(opcode)]
	return nil
}

func op8XY1(c *CPU, opcode uint16) error {
	c.V[getX(opcode)] |= c.V[getY(opcode)]
	if c.Quirks.VFReset {
		c.V[0xF] = 0
	}
	return nil
}

func op8XY2(c *CPU, opcode uint16) error {
	c.V[getX(opcode)] &= c.V[getY(opcode)]
	if c.Quirks.VFReset {
		c.V[0xF] = 0
	}
	return nil
}

func op8XY3(c *CPU, opcode uint16) error {
	c.V[getX(opcode)] ^= c.V[getY(opcode)]
	if c.Quirks.VFReset {
		c.V[0xF] = 0
	}
	return nil
}

func op8XY4(c *CPU, opcode uint16) error {
	x, y := getX(opcode), getY(opcode)
	sum := uint16(c.V[x]) + uint16(c.V[y])
	if sum > 0xFF {
		c.V[0xF] = 1
	} else {
		c.V[0xF] = 0
	}
	c.V[x] = byte(sum)
	return nil
}

func op8XY5(c *CPU, opcode uint16) error {
	x, y := getX(opcode), getY(opcode)
	if c.V[x] >= c.V[y] {
		c.V[0xF] = 1
	} else {
		c.V[0xF] = 0
	}
	c.V[x] -= c.V[y]
	return nil
}

func op8XY6(c *CPU, opcode uint16) error {
	x, y := getX(opcode), getY(opcode)
	var tmp byte
	if c.Quirks.ShiftSetVX {
		tmp = c.V[x]
	} else {
		tmp = c.V[y]
	}
	c.V[0xF] = tmp & 0x1
	c.V[x] = tmp >> 1
	return nil
}

func op8XY7(c *CPU, opcode uint16) error {
	x, y := getX(opcode), getY(opcode)
	if c.V[y] >= c.V[x] {
		c.V[0xF] = 1
	} else {
		c.V[0xF] = 0
	}
	c.V[x] = c.V[y] - c.V[x]
	return nil
}

func op8XYE(c *CPU, opcode uint16) error {
	x, y := getX(opcode), getY(opcode)
	var tmp byte
	if c.Quirks.ShiftSetVX {
		tmp = c.V[x]
	} else {
		tmp = c.V[y]
	}
	c.V[0xF] = tmp >> 7
	c.V[x] = tmp << 1
	return nil
}

func op9XY0(c *CPU, opcode uint16) error {
	if c.V[getX(opcode)] != c.V[getY(opcode)] {
		c.PC += 2
	}
	return nil
}

func opANNN(c *CPU, opcode uint16) error {
	c.I = getNNN(opcode)
	return nil
}

func opBNNN(c *CPU, opcode uint16) error {
	if c.Quirks.Jumping {
		c.PC = getNNN(opcode) + uint16(c.V[getX(opcode)])
	} else {
		c.PC = getNNN(opcode) + uint16(c.V[0])
	}
	return nil
}

func opCXNN(c *CPU, opcode uint16) error {
	c.V[getX(opcode)] = byte(rand.Uint32()&0xFF) & getNN(opcode)
	return nil
}

// i am hoping i am doing this correctly,
// this is my best effort to implement the drawing opcode.
func opDXYN(c *CPU, opcode uint16) error {
	if c.Quirks.DisplayWait && c.WaitingForFrame {
		c.PC -= 2
		return nil
	}

	x, y := getX(opcode), getY(opcode)
	n := getN(opcode)

	w, h := c.screenWidth(), c.screenHeight()
	xPos := int(c.V[x]) % w
	yPos := int(c.V[y]) % h
	c.V[0xF] = 0

	height := int(n)
	spriteWidth := 8
	is16x16 := false
	isLowRes16x16 := false

	if n == 0 {
		if c.HighRes {
			// High-res mode: 16x16 sprite at actual size
			height = 16
			spriteWidth = 16
			is16x16 = true
		} else {
			// Low-res mode: 16x16 sprite data, but drawn as 8x16 with pixel doubling
			height = 16
			spriteWidth = 16
			is16x16 = true
			isLowRes16x16 = true
		}
	}

	memPtr := c.I
	planes := []struct {
		active bool
		buf    *[128 * 64]bool
	}{
		{c.SelectedPlanes&1 != 0, &c.Display},
		{c.SelectedPlanes&2 != 0, &c.Display2},
	}

	for _, p := range planes {
		if !p.active {
			continue
		}
		for row := 0; row < height; row++ {
			var spriteRow uint16
			if is16x16 {
				spriteRow = uint16(c.Memory[memPtr+uint16(row*2)])<<8 | uint16(c.Memory[memPtr+uint16(row*2+1)])
			} else {
				spriteRow = uint16(c.Memory[memPtr+uint16(row)])
			}

			if isLowRes16x16 {
				// Low-res 16x16: draw as 8x16 with pixel doubling
				// Each 2x2 block of bits becomes 1 pixel, then doubled
				for col := 0; col < 8; col++ {
					// Sample 2x2 block from the 16-bit sprite row
					bit00 := (spriteRow >> (15 - col*2)) & 1
					bit01 := (spriteRow >> (14 - col*2)) & 1
					bit10 := (spriteRow >> (15 - col*2)) & 1
					bit11 := (spriteRow >> (14 - col*2)) & 1

					// Average the 2x2 block (if any bit is set, draw the pixel)
					pixel := bit00 | bit01 | bit10 | bit11
					if pixel == 0 {
						continue
					}

					// Draw doubled pixel (2x2 on screen)
					for dy := 0; dy < 2; dy++ {
						for dx := 0; dx < 2; dx++ {
							screenX := xPos + col*2 + dx
							screenY := yPos + row*2 + dy

							if !c.Quirks.Wrapping {
								if screenX >= w || screenY >= h {
									continue
								}
							}

							screenX = screenX % w
							screenY = screenY % h
							index := screenY*128 + screenX

							if p.buf[index] {
								c.V[0xF] = 1
							}
							p.buf[index] = !p.buf[index]
						}
					}
				}
			} else {
				// Normal drawing (8xN or 16x16 in high-res)
				if !c.Quirks.Wrapping && yPos+row >= h {
					break
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
					if p.buf[index] {
						c.V[0xF] = 1
					}
					p.buf[index] = !p.buf[index]
				}
			}
		}
	}
	c.DrawFlag = true
	if c.Quirks.DisplayWait {
		c.WaitingForFrame = true
	}
	return nil
}

func opEX9E(c *CPU, opcode uint16) error {
	if c.Keys[c.V[getX(opcode)]] {
		c.PC += 2
	}
	return nil
}

func opEXA1(c *CPU, opcode uint16) error {
	if !c.Keys[c.V[getX(opcode)]] {
		c.PC += 2
	}
	return nil
}

func opF000(c *CPU, opcode uint16) error {
	c.I = uint16(c.Memory[c.PC])<<8 | uint16(c.Memory[c.PC+1])
	c.PC += 2
	return nil
}

func opFX01(c *CPU, opcode uint16) error {
	c.SelectedPlanes = byte(getX(opcode)) & 0x3
	return nil
}

func opFX02(c *CPU, opcode uint16) error {
	copy(c.AudioPattern[:], c.Memory[c.I:c.I+16])
	if c.Beep != nil {
		c.Beep.SetPattern(c.AudioPattern)
	}
	return nil
}

func opFX07(c *CPU, opcode uint16) error {
	c.V[getX(opcode)] = c.DelayTimer
	return nil
}

func opFX0A(c *CPU, opcode uint16) error {
	pressed := false
	for i := byte(0); i < 16; i++ {
		if c.Keys[i] {
			c.V[getX(opcode)] = i
			pressed = true
			break
		}
	}
	if !pressed {
		c.PC -= 2
	}
	return nil
}

func opFX15(c *CPU, opcode uint16) error {
	c.DelayTimer = c.V[getX(opcode)]
	return nil
}

func opFX18(c *CPU, opcode uint16) error {
	c.SoundTimer = c.V[getX(opcode)]
	return nil
}

func opFX1E(c *CPU, opcode uint16) error {
	c.I += uint16(c.V[getX(opcode)])
	return nil
}

func opFX29(c *CPU, opcode uint16) error {
	c.I = uint16(c.V[getX(opcode)]) * 5
	return nil
}

func opFX30(c *CPU, opcode uint16) error {
	c.I = 0xA0 + uint16(c.V[getX(opcode)])*10
	return nil
}

func opFX33(c *CPU, opcode uint16) error {
	x := getX(opcode)
	c.Memory[c.I] = c.V[x] / 100
	c.Memory[c.I+1] = (c.V[x] / 10) % 10
	c.Memory[c.I+2] = c.V[x] % 10
	return nil
}

func opFX3A(c *CPU, opcode uint16) error {
	c.Pitch = c.V[getX(opcode)]
	if c.Beep != nil {
		c.Beep.SetPitch(c.V[getX(opcode)])
	}
	return nil
}

func opFX55(c *CPU, opcode uint16) error {
	x := getX(opcode)
	for i := uint16(0); i <= x; i++ {
		c.Memory[c.I+i] = c.V[i]
	}
	if c.Quirks.IncrementI {
		c.I += x + 1
	}
	return nil
}

func opFX65(c *CPU, opcode uint16) error {
	x := getX(opcode)
	for i := uint16(0); i <= x; i++ {
		c.V[i] = c.Memory[c.I+i]
	}
	if c.Quirks.IncrementI {
		c.I += x + 1
	}
	return nil
}

func opFX75(c *CPU, opcode uint16) error {
	x := getX(opcode)
	for i := uint16(0); i <= x; i++ {
		c.Flags[i] = c.V[i]
	}
	return nil
}

func opFX85(c *CPU, opcode uint16) error {
	x := getX(opcode)
	for i := uint16(0); i <= x; i++ {
		c.V[i] = c.Flags[i]
	}
	return nil
}
