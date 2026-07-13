package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
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
		Use:   "chip8 [rom-file]",
		Short: "A CHIP-8 emulator",
		Args:  cobra.MaximumNArgs(1),
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

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func run(cmd *cobra.Command, args []string) error {
	programLevel := new(slog.LevelVar)
	if debugFlag {
		programLevel.Set(slog.LevelDebug)
	} else {
		programLevel.Set(slog.LevelInfo)
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: programLevel}))
	slog.SetDefault(logger)
	var romPath string
	if len(args) > 0 {
		romPath = args[0]
	}

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
	memSize := 65536 // maximum memory size for CHIP-8 and its variants
	slog.Debug("Using memsize", "memsize", memSize)
	var r *rom.ROM
	if romPath != "" {
		r, err = rom.Load(romPath, memSize)
		if err != nil {
			return err
		}
	} else {
		r = &rom.ROM{Data: nil}
		slog.Debug("No ROM path provided, starting without a ROM loaded")
	}

	slog.Debug("Using quirks", "quirks", quirks, "platform", platform, "octoConfigPath", octoConfigPath)
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

	if len(r.Data) > 0 {
		c.LoadROM(r.Data)
	}
	game := emulator.NewGame(c, displ, beep, player, cyclesPerFrm, r)
	ebiten.SetWindowSize(display.WindowSizeW, display.WindowSizeH)
	ebiten.SetWindowTitle("CHIP-8 Emulator")
	ebiten.SetTPS(ebitenTPS)
	return ebiten.RunGame(game)
}
