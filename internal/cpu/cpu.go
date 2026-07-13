package cpu

import (
	myAudio "chip8/internal/audio"
	"chip8/internal/display"
	"fmt"
	"log/slog"
)

type CPU struct {
	Memory          [65536]byte    // 64KB memory
	V               [16]byte       // 16 general-purpose registers (V0 to VF)
	I               uint16         // Index register
	PC              uint16         // Program counter
	SP              byte           // Stack pointer
	Stack           [256]uint16    // Stack
	DelayTimer      byte           // Delay timer
	SoundTimer      byte           // Sound timer
	Display         [128 * 64]bool // plane 0
	Display2        [128 * 64]bool // plane 1
	SelectedPlanes  byte           // bitmask: bit0=plane0, bit1=plane1
	AudioPattern    [16]byte
	Pitch           byte
	HighRes         bool          // true if in high-resolution mode (128x64), false for low-resolution (64x32)
	Keys            [16]bool      // Key states
	DrawFlag        bool          // Flag to indicate if the display needs to be redrawn
	Quirks          Quirks        // Quirks for specific CHIP-8 implementations
	Debug           Debug         // Debugging information
	WaitingForFrame bool          // set after a draw when DisplayWait quirk is on
	Beep            *myAudio.Beep // Audio interface for beep sound
	Flags           [16]byte      // RPL user flags (SCHIP/XO-CHIP extension)
	DisplayRef      *display.Display
	Opcode          uint16 // Last executed opcode (for debugging)
}

func NewCPU(quirks Quirks, display *display.Display) *CPU {
	cpu := &CPU{
		PC:             0x200,  // Programs start at memory location 0x200
		SelectedPlanes: 0x1,    // Default to plane 0
		Quirks:         quirks, // Use provided quirks
		DisplayRef:     display,
	}
	loadFont(cpu) // Load the font set into memory
	slog.Debug("CPU initialized", "quirks", quirks)
	return cpu
}

func (c *CPU) Cycle() error { // emulates a single cycle of the CPU (fetch, decode, execute)
	opcode := c.Fetch()
	c.Opcode = opcode // store the last executed opcode for debugging

	op, err := c.DecodeAndExecute(opcode)

	c.Debug.Cycle++

	slog.Debug("CPU cycle executed", "cycle", c.Debug.Cycle, "pc", c.PC-2, "opcode", opcode, "instruction", op.FormatASM(opcode), "i", c.I, "v0", c.V[0], "v1", c.V[1], "v2", c.V[2], "description", op.description)

	if c.Debug.Step { // TODO: add a better handle for step mode
		slog.Debug("STEP MODE: press Enter to continue")
		fmt.Scanln()
	}

	return err
}

func (c *CPU) TickTimers() { // emulates the decrementing of the delay and sound timers (NOTE timers must decrement at ~60 Hz, not per instruction)
	if c.DelayTimer > 0 {
		c.DelayTimer--
	}

	if c.SoundTimer > 0 {
		c.SoundTimer--
	}
	c.WaitingForFrame = false // new frame has started, drawing allowed again
	slog.Debug("Timers ticked", "delayTimer", c.DelayTimer, "soundTimer", c.SoundTimer)
}

func (c *CPU) LoadROM(data []byte) { // loads a ROM into memory starting at 0x200
	copy(c.Memory[0x200:], data)
	c.PC = 0x200
	slog.Debug("ROM loaded", "size", len(data))
}

func (c *CPU) SetKeys(keys [16]bool) { // updates the state of the keys (NOTE this should be called from the main loop to update the key states using the ebiten engine key state)
	c.Keys = keys
	slog.Debug("Keys updated", "keys", keys)
}
