package cpu

func instructionToString(opcode uint16) string {
	// x := (opcode & 0x0F00) >> 8 // x register index
	// y := (opcode & 0x00F0) >> 4 // y register index
	n := opcode & 0x000F        // last nibble (n) value
	nn := byte(opcode & 0x00FF) // last byte (nn) value
	// nnn := opcode & 0x0FFF      // last 12 bits (nnn) value
	switch opcode & 0xF000 {
	case 0x0000:
		if opcode&0xFFF0 == 0x00C0 { // 00CN: scroll display down N pixels
			return "SCROLL DOWN " + hex(opcode&0x000F)
		}
		if opcode&0xFFF0 == 0x00D0 { // 00CN: scroll display down N pixels
			return "SCROLL UP " + hex(opcode&0x000F)
		}
		switch opcode {
		case 0x00E0: // Clear the display
			return "CLS"

		case 0x00EE: // Return from subroutine
			return "RET"

		case 0x00FB: // scroll right 4 (or 2 in lores w/ HalfScroll)
			return "SCROLL RIGHT"

		case 0x00FC: // scroll left 4 (or 2 in lores w/ HalfScroll)
			return "SCROLL LEFT"

		case 0x00FE: // low-res mode
			return "LORES"

		case 0x00FF: // high-res mode
			return "HIRES"
		}

	case 0x1000: // Jump to address NNN
		return "JP " + hex(opcode&0x0FFF)

	case 0x2000: // Call subroutine at NNN
		return "CALL " + hex(opcode&0x0FFF)

	case 0x3000: // Skip next instruction if Vx == NN
		return "SE V" + hex((opcode>>8)&0xF) + ", " + hex(opcode&0xFF)

	case 0x4000: // Skip next instruction if Vx != NN
		return "SNE V" + hex((opcode>>8)&0xF) + ", " + hex(opcode&0xFF)

	case 0x5000: // Skip next instruction if Vx == Vy
		return "SE V" + hex((opcode>>8)&0xF) + ", V" + hex((opcode>>4)&0xF)

	case 0x6000: // Load NN into Vx
		return "LD V" + hex((opcode>>8)&0xF) + ", " + hex(opcode&0xFF)

	case 0x7000: // Add NN to Vx
		return "ADD V" + hex((opcode>>8)&0xF) + ", " + hex(opcode&0xFF)

	case 0x8000:
		switch n { // switch on the last nibble to determine the operation
		case 0: // Vx = Vy
			return "LD V" + hex((opcode>>8)&0xF) + ", V" + hex((opcode>>4)&0xF)

		case 1: // Vx = Vx OR Vy
			return "OR V" + hex((opcode>>8)&0xF) + ", V" + hex((opcode>>4)&0xF)

		case 2: // Vx = Vx AND Vy
			return "AND V" + hex((opcode>>8)&0xF) + ", V" + hex((opcode>>4)&0xF)

		case 3: // Vx = Vx XOR Vy
			return "XOR V" + hex((opcode>>8)&0xF) + ", V" + hex((opcode>>4)&0xF)

		case 4: // Vx = Vx + Vy, set VF = carry (eg. addition with carry)
			return "ADD V" + hex((opcode>>8)&0xF) + ", V" + hex((opcode>>4)&0xF)

		case 5: // Vx = Vx - Vy, set VF = NOT borrow (eg. subtraction with borrow)
			return "SUB V" + hex((opcode>>8)&0xF) + ", V" + hex((opcode>>4)&0xF)

		case 6: // Vx = Vx SHR 1, set VF = least significant bit of Vx before the shift
			return "SHR V" + hex((opcode>>8)&0xF) + ", V" + hex((opcode>>4)&0xF)

		case 7: // Vx = Vy - Vx, set VF = NOT borrow
			return "SUBN V" + hex((opcode>>8)&0xF) + ", V" + hex((opcode>>4)&0xF)

		case 0xE: // Vx = Vx SHL 1, set VF = most significant bit of Vx before the shift
			return "SHL V" + hex((opcode>>8)&0xF) + ", V" + hex((opcode>>4)&0xF)
		}

	case 0x9000: // Skip next instruction if Vx != Vy
		return "SNE V" + hex((opcode>>8)&0xF) + ", V" + hex((opcode>>4)&0xF)

	case 0xA000: // Set I = NNN (eg. load index register with address NNN)
		return "LD I, " + hex(opcode&0x0FFF)

	case 0xB000: // jump with offset: PC = NNN + V0
		return "JP " + hex(opcode&0x0FFF) + ", V0"

	case 0xC000: // Set Vx = random byte AND NN
		return "RND V" + hex((opcode>>8)&0xF) + ", " + hex(opcode&0xFF)

	case 0xD000: // draw sprite at (Vx, Vy) with width 8 pixels and height N pixels. Set VF = collision.
		return "DRW V" + hex((opcode>>8)&0xF) + ", V" + hex((opcode>>4)&0xF) + ", " + hex(opcode&0xF)

	case 0xE000:
		switch nn {
		case 0x9E: // Skip next instruction if key with the value of Vx is pressed
			return "SKP V" + hex((opcode>>8)&0xF)
		case 0xA1: // Skip next instruction if key with the value of Vx is not pressed
			return "SKNP V" + hex((opcode>>8)&0xF)
		}

	case 0xF000:
		switch nn {
		case 0x07: // Set Vx = delay timer value
			return "LD V" + hex((opcode>>8)&0xF) + ", DT"
		case 0x0A: // Wait for a key press, store the value of the key in Vx
			return "LD V" + hex((opcode>>8)&0xF) + ", K"
		case 0x15: // Set delay timer = Vx
			return "LD DT, V" + hex((opcode>>8)&0xF)
		case 0x18: // Set sound timer = Vx
			return "LD ST, V" + hex((opcode>>8)&0xF)
		case 0x1E: // Set I = I + Vx
			return "ADD I, V" + hex((opcode>>8)&0xF)
		case 0x29: // Set I = location of sprite for digit Vx
			return "LD F, V" + hex((opcode>>8)&0xF)
		case 0x33: // Store BCD representation of Vx in memory locations I, I+1, and I+2
			return "LD B, V" + hex((opcode>>8)&0xF)
		case 0x55: // Store registers V0 through Vx in memory starting at location I
			return "LD [I], V" + hex((opcode>>8)&0xF)
		case 0x65: // Read registers V0 through Vx from memory starting at location I
			return "LD V" + hex((opcode>>8)&0xF) + ", [I]"
		}
	default:
		return "UNKNOWN " + hex(opcode)
	}
	return "UNKNOWN " + hex(opcode)
}
