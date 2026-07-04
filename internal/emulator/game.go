package emulator

import (
	"chip8/internal/cpu"
	"chip8/internal/display"
	"chip8/internal/keyboard"

	myAudio "chip8/internal/audio"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
)

type Game struct {
	CPU        *cpu.CPU
	Display    *display.Display
	beepPlayer *audio.Player
	beepStream *myAudio.Beep

	cyclesPerFrame int
}

func NewGame(cpu *cpu.CPU, display *display.Display, beepStream *myAudio.Beep, beepPlayer *audio.Player, cyclesPerFrame int) *Game {
	if cyclesPerFrame <= 0 {
		cyclesPerFrame = 12 // Default value if not provided or invalid
	}
	return &Game{
		CPU:            cpu,
		Display:        display,
		beepStream:     beepStream,
		beepPlayer:     beepPlayer,
		cyclesPerFrame: cyclesPerFrame,
	}
}

func (g *Game) Update() error {
	for i := 0; i < g.cyclesPerFrame; i++ {
		if err := g.CPU.Cycle(); err != nil {
			return err
		}
	}
	g.CPU.TickTimers()

	if g.CPU.SoundTimer > 0 {
		g.beepStream.Activate()
	} else {
		g.beepStream.Deactivate()
	}

	if g.CPU.DrawFlag {
		g.Display.UpdateFromMemory(g.CPU.Display[:], g.CPU.HighRes) // pass resolution
		g.CPU.DrawFlag = false
	}

	keys := keyboard.PollKeys()
	g.CPU.SetKeys(keys)
	return nil
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	if g.CPU.HighRes {
		return 128, 64
	}
	return 64, 32
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.Display.Draw(screen)
}
