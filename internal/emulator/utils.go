package emulator

import (
	cpuImpl "chip8/internal/cpu"
	"log/slog"
)

func (g *Game) loadROM(data []byte) {
	// new cpu instance to reset the state with the last quirks and display reference
	newCPU := cpuImpl.NewCPU(g.CPU.Quirks, g.EmuDisplay)
	newCPU.LoadROM(data)

	g.CPU = newCPU
	g.currentROM = data
	g.isRomLoaded = len(data) > 0

	g.EmuDisplay.Clear()
	g.CPU.DrawFlag = false

	g.showDebug = false
	slog.Debug("ROM loaded", "size", len(data))
}
