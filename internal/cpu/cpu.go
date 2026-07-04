package cpu

import "fmt"

type CPU struct {
	Memory          [65536]byte    // 64KB memory
	V               [16]byte       // 16 general-purpose registers (V0 to VF)
	I               uint16         // Index register
	PC              uint16         // Program counter
	SP              byte           // Stack pointer
	Stack           [16]uint16     // Stack
	DelayTimer      byte           // Delay timer
	SoundTimer      byte           // Sound timer
	Display         [128 * 64]bool // 128x64 monochrome display
	HighRes         bool           // true if in high-resolution mode (128x64), false for low-resolution (64x32)
	Keys            [16]bool       // Key states
	DrawFlag        bool           // Flag to indicate if the display needs to be redrawn
	Quirks          Quirks         // Quirks for specific CHIP-8 implementations
	Debug           Debug          // Debugging information
	WaitingForFrame bool           // set after a draw when DisplayWait quirk is on
}

func NewCPU(quirks Quirks) *CPU {
	cpu := &CPU{
		PC:     0x200,  // Programs start at memory location 0x200
		I:      0,      // Initialize index register
		SP:     0,      // Initialize stack pointer
		Quirks: quirks, // Use provided quirks
	}
	loadFont(cpu) // Load the font set into memory at 0x50
	return cpu
}

func (c *CPU) Cycle() error { // emulates a single cycle of the CPU (fetch, decode, execute)
	opcode := c.Fetch()

	err := c.DecodeAndExecute(opcode)

	c.Debug.Cycle++

	if c.Debug.Enabled {
		fmt.Printf(
			"[CYCLE %d] PC=%03X OPCODE=%04X %-12s I=%03X V0=%02X V1=%02X V2=%02X\n",
			c.Debug.Cycle,
			c.PC-2,
			opcode,
			Disassemble(opcode),
			c.I,
			c.V[0], c.V[1], c.V[2],
		)
	}

	if c.Debug.Step {
		fmt.Println("STEP MODE: press Enter to continue")
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
}

func (c *CPU) LoadROM(data []byte) { // loads a ROM into memory starting at 0x200
	copy(c.Memory[0x200:], data)
	c.PC = 0x200
}

func (c *CPU) SetKeys(keys [16]bool) { // updates the state of the keys (NOTE this should be called from the main loop to update the key states using the ebiten engine key state)
	c.Keys = keys
}
