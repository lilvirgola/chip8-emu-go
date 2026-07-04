package cpu

func (c *CPU) Fetch() uint16 {
	opcode := uint16(c.Memory[c.PC])<<8 |
		uint16(c.Memory[c.PC+1])

	c.PC += 2

	return opcode
}
