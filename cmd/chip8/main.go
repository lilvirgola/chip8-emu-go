package main

import (
	"fmt"
	"log"
	"time"

	myAudio "chip8/internal/audio"
	"chip8/internal/cpu"
	"chip8/internal/display"
	"chip8/internal/emulator"
	"chip8/internal/octoconfig"
	"chip8/internal/rom"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/spf13/cobra"
)

const ebitenTPS = 60

var (
	platform       string
	octoConfigPath string
	debugFlag      bool
	stepFlag       bool
	cyclesPerFrm   int

	// individual quirk overrides
	vfReset     bool
	shiftSetVX  bool
	incrementI  bool
	wrapping    bool
	displayWait bool
	jumping     bool

	// test flags for development/debugging
	testColors bool
)

func platformQuirks(name string) (cpu.Quirks, error) {
	switch name {
	case "vip", "cosmac", "cosmac-vip":
		return cpu.CosmacVIP, nil
	case "chip48":
		return cpu.Chip48, nil
	case "schip-legacy", "superchip-legacy":
		return cpu.SuperchipLegacy, nil
	case "schip", "schip-modern", "superchip-modern":
		return cpu.SuperchipModern, nil
	case "xochip", "xo-chip":
		return cpu.XOChip, nil
	default:
		return cpu.Quirks{}, fmt.Errorf("unknown platform %q (valid: vip, chip48, schip-legacy, schip-modern, xochip)", name)
	}
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "chip8 <rom-file>",
		Short: "A CHIP-8 emulator",
		Args:  cobra.ExactArgs(1),
		RunE:  run,
	}

	rootCmd.Flags().StringVar(&platform, "platform", "vip",
		"quirk preset: vip, chip48, schip-legacy, schip-modern, xochip")
	rootCmd.Flags().StringVar(&octoConfigPath, "octo-config", "", "path to an Octo config file (overrides --platform quirks/colors)")
	rootCmd.Flags().BoolVar(&debugFlag, "debug", false, "enable debug logging of executed opcodes")
	rootCmd.Flags().BoolVar(&stepFlag, "step", false, "pause after each instruction (requires --debug)")
	rootCmd.Flags().IntVar(&cyclesPerFrm, "cycles", 12, "CPU cycles executed per frame")

	// Individual quirk overrides — unset means "use the platform preset's value"
	rootCmd.Flags().BoolVar(&vfReset, "vf-reset", false, "override: reset VF after 8xy1/8xy2/8xy3")
	rootCmd.Flags().BoolVar(&shiftSetVX, "shift-vx", false, "override: shift VX directly instead of VY->VX")
	rootCmd.Flags().BoolVar(&incrementI, "increment-i", false, "override: increment I after FX55/FX65")
	rootCmd.Flags().BoolVar(&wrapping, "wrapping", false, "override: sprites wrap instead of clip")
	rootCmd.Flags().BoolVar(&displayWait, "display-wait", false, "override: Dxyn waits for vblank")
	rootCmd.Flags().BoolVar(&jumping, "jumping", false, "override: Bxnn uses Vx instead of Bnnn using V0")

	// Test flags for development/debugging (TODO:fix this test)
	rootCmd.Flags().BoolVar(&testColors, "test-colors", false, "test colors")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func run(cmd *cobra.Command, args []string) error {

	if testColors {
		displ := display.NewDisplay()
		displ.HighRes = true
		displ.Colors = [4][3]byte{
			{0x00, 0x00, 0x00}, // black
			{0xFF, 0x00, 0x00}, // red
			{0x00, 0xFF, 0x00}, // green
			{0x00, 0x00, 0xFF}, // blue
		}

		// Draw test patterns (these coordinates now match the 128x64 buffer)
		// Red square (plane0 only) at top-left
		for y := 10; y < 30; y++ {
			for x := 10; x < 30; x++ {
				displ.Plane0[y*128+x] = true
			}
		}

		// Green square (plane1 only) at top-right
		for y := 10; y < 30; y++ {
			for x := 100; x < 120; x++ {
				displ.Plane1[y*128+x] = true
			}
		}

		// Blue square (both planes) at bottom
		for y := 40; y < 60; y++ {
			for x := 54; x < 74; x++ {
				displ.Plane0[y*128+x] = true
				displ.Plane1[y*128+x] = true
			}
		}

		game := emulator.NewGame(nil, displ, nil, nil, cyclesPerFrm)
		ebiten.SetWindowSize(1280, 640)
		ebiten.SetWindowTitle("CHIP-8 Emulator - Color Test")
		ebiten.SetTPS(ebitenTPS)
		return ebiten.RunGame(game)
	}
	romPath := args[0]

	quirks, err := platformQuirks(platform)
	if err != nil {
		return err
	}
	var displayColors *octoconfig.Config
	if octoConfigPath != "" {
		cfg, err := octoconfig.Load(octoConfigPath)
		if err != nil {
			return err
		}
		quirks = cfg.ToQuirks()
		displayColors = &cfg
		cyclesPerFrm = cfg.TickRate / ebitenTPS // convert instructions/sec to instructions/frame
	}

	// Apply explicit overrides only if the flag was actually passed on the CLI
	if cmd.Flags().Changed("vf-reset") {
		quirks.VFReset = vfReset
	}
	if cmd.Flags().Changed("shift-vx") {
		quirks.ShiftSetVX = shiftSetVX
	}
	if cmd.Flags().Changed("increment-i") {
		quirks.IncrementI = incrementI
	}
	if cmd.Flags().Changed("wrapping") {
		quirks.Wrapping = wrapping
	}
	if cmd.Flags().Changed("display-wait") {
		quirks.DisplayWait = displayWait
	}
	if cmd.Flags().Changed("jumping") {
		quirks.Jumping = jumping
	}
	memSize := 4096
	if platform == "xochip" || platform == "xo-chip" {
		memSize = 65536
	}
	r, err := rom.Load(romPath, memSize)
	if err != nil {
		return err
	}
	fmt.Printf("DEBUG: quirks = %+v\n", quirks)
	beep := myAudio.NewBeep()

	audioContext := audio.NewContext(44100)

	player, _ := audioContext.NewPlayer(beep)
	player.SetBufferSize(50 * time.Millisecond)
	player.SetVolume(0.5)
	player.Play()
	displ := display.NewDisplay()
	if displayColors != nil {
		displ.Colors = [4][3]byte{
			{displayColors.Background.R, displayColors.Background.G, displayColors.Background.B},
			{displayColors.Plane0.R, displayColors.Plane0.G, displayColors.Plane0.B},
			{displayColors.Plane1.R, displayColors.Plane1.G, displayColors.Plane1.B},
			{displayColors.Plane2.R, displayColors.Plane2.G, displayColors.Plane2.B},
		}
	}

	c := cpu.NewCPU(quirks, displ)
	c.Debug.Enabled = debugFlag
	c.Debug.Step = stepFlag && debugFlag // step only makes sense with debug on
	c.LoadROM(r.Data)
	game := emulator.NewGame(c, displ, beep, player, cyclesPerFrm)
	ebiten.SetWindowSize(1280, 640)
	ebiten.SetWindowTitle("CHIP-8 Emulator")
	ebiten.SetTPS(ebitenTPS)
	return ebiten.RunGame(game)
}
