package cpu

type Quirks struct {
	ShiftSetVX bool // false = VX = VY before shift, true = VX is shifted directly
	IncrementI bool // FX55/FX65 increments I after operation
	Wrapping   bool
}

var DefaultQuirks = Quirks{
	ShiftSetVX: true,  // Default behavior for shift instructions
	IncrementI: false, // Default behavior for FX55/FX65 instructions
	Wrapping:   false, // Default behavior for memory wrapping
}
