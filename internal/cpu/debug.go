package cpu

import "fmt"

type Debug struct {
	Enabled bool
	Step    bool

	Cycle uint64
}

func Disassemble(opcode uint16) string {
	return instructionToString(opcode)
}

func hex(v uint16) string {
	return fmt.Sprintf("%X", v)
}
