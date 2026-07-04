package main

import (
	"chip8/internal/cpu"
	"chip8/internal/display"
	"chip8/internal/emulator"
	"chip8/internal/rom"
	"log"
	"os"

	myAudio "chip8/internal/audio"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: chip8 <rom-file>")
	}

	r, err := rom.Load(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	c := cpu.NewCPU()
	c.Debug.Enabled = true // Enable debug mode to print executed opcodes to the console
	c.Debug.Step = false   // Enable step mode to pause execution after each instruction
	c.LoadROM(r.Data)
	audioContext := audio.NewContext(44100)

	beep := myAudio.NewBeep()
	player, _ := audioContext.NewPlayer(beep)
	player.Play()

	game := emulator.NewGame(c, &display.Display{}, beep, player, 12)

	ebiten.SetWindowSize(640, 320)
	ebiten.SetWindowTitle("CHIP-8 Emulator")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
