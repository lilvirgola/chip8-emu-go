package cpu

type Quirks struct {
	VFReset     bool // 8xy1/8xy2/8xy3 reset VF to 0 after the operation
	ShiftSetVX  bool // false = VX = VY before shift, true = VX is shifted directly
	IncrementI  bool // FX55/FX65 increments I after operation
	Wrapping    bool // sprites wrap around screen edges instead of clipping
	DisplayWait bool // Dxyn waits for vblank before drawing
	Jumping     bool // false = Bnnn uses V0, true = Bxnn uses Vx
	HalfScroll  bool // when true, 00Cn/00FB/00FC scroll distances are halved while in low-res (64x32) mode
}

// CosmacVIP is the original 1977 CHIP-8 interpreter behavior.
var CosmacVIP = Quirks{
	VFReset:     true,
	ShiftSetVX:  false,
	IncrementI:  true,
	Wrapping:    false,
	DisplayWait: true,
	Jumping:     false,
	HalfScroll:  true,
}

// Chip48 matches the HP-48 calculator CHIP-8 port.
var Chip48 = Quirks{
	VFReset:     false,
	ShiftSetVX:  true,
	IncrementI:  false,
	Wrapping:    false,
	DisplayWait: false,
	Jumping:     true,
	HalfScroll:  true,
}

// SuperchipLegacy is SUPER-CHIP 1.0.
var SuperchipLegacy = Quirks{
	VFReset:     false,
	ShiftSetVX:  true,
	IncrementI:  false,
	Wrapping:    false,
	DisplayWait: true,
	Jumping:     true,
	HalfScroll:  true,
}

// SuperchipModern is SUPER-CHIP 1.1 (most common "SCHIP" target).
var SuperchipModern = Quirks{
	VFReset:     false,
	ShiftSetVX:  true,
	IncrementI:  false,
	Wrapping:    false, // clips instead of wraps
	DisplayWait: true,
	Jumping:     true,
	HalfScroll:  false,
}

// XOChip is the modern extended CHIP-8 variant.
var XOChip = Quirks{
	VFReset:     false,
	ShiftSetVX:  false,
	IncrementI:  true,
	Wrapping:    true, // wraps instead of clipping
	DisplayWait: true,
	Jumping:     false,
	HalfScroll:  false,
}

var DefaultQuirks = CosmacVIP

func (q Quirks) IsXOChip() bool {
	return q == XOChip
}

func (q Quirks) IsSuperchip() bool {
	return q == SuperchipLegacy || q == SuperchipModern
}

func (q Quirks) IsSuperchipModern() bool {
	return q == SuperchipModern
}

func (q Quirks) IsSuperchipLegacy() bool {
	return q == SuperchipLegacy
}
func (q Quirks) IsChip48() bool {
	return q == Chip48
}

func (q Quirks) IsCosmacVIP() bool {
	return q == CosmacVIP
}

func GetAvailableQuirks() ([]Quirks, []string) {
	quirks := []Quirks{
		CosmacVIP,
		Chip48,
		SuperchipLegacy,
		SuperchipModern,
		XOChip,
	}
	names := []string{
		"cosmac-vip",
		"chip48",
		"superchip-legacy",
		"superchip-modern",
		"xochip",
	}
	return quirks, names
}
