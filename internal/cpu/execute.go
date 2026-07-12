package cpu

import (
	"fmt"
)

func (c *CPU) DecodeAndExecute(opcode uint16) (Opcode, error) {
	for _, op := range opcodes {
		// searching for the opcode that matches the mask and value
		if (opcode & op.mask) == op.value {
			return op, op.function(c, opcode)
		}
	}
	return Opcode{}, fmt.Errorf("unknown opcode %04X", opcode)
}
