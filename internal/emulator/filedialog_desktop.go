//go:build !js

package emulator

import (
	"log/slog"
	"os"

	"github.com/sqweek/dialog"
)

func openROMDialog() (string, error) {
	return dialog.File().Filter("CHIP-8 ROM", "ch8", "c8", "bin").Load()
}

func (g *Game) OnLoadROMClick() {
	path, err := openROMDialog()
	if err != nil || path == "" {
		return // User canceled or error occurred
	}
	data, err := os.ReadFile(path)
	if err != nil {
		slog.Error("Failed to read ROM file", "path", path, "error", err)
		return
	}
	g.loadROM(data)
}
