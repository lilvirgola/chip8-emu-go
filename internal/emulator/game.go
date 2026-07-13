package emulator

import (
	"bytes"
	"chip8/internal/cpu"
	"chip8/internal/display"
	"chip8/internal/keyboard"
	"fmt"
	"image"
	"io/fs"
	"log"
	"log/slog"
	"sync"

	myAudio "chip8/internal/audio"

	cpuImpl "chip8/internal/cpu"
	"chip8/internal/rom"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/goregular"
)

type Game struct {
	CPU          *cpu.CPU
	EmuDisplay   *display.Display
	DebugDisplay *display.Display
	beepPlayer   *audio.Player
	beepStream   *myAudio.Beep

	cyclesPerFrame int
	debugFace      text.Face
	lineHeight     float64
	showDebug      bool

	// rom loading
	isRomLoaded bool
	currentROM  []byte
	romMu       sync.Mutex
	pendingROM  []byte
	loadingROM  bool
	romLoadErr  error

	// UI elements
	debugToggleBtn  *Button
	loadROMBtn      *Button
	cpuTypeSelector *CPUTypeSelector
}

func NewGame(cpu *cpu.CPU, display *display.Display, beepStream *myAudio.Beep, beepPlayer *audio.Player, cyclesPerFrame int, rom *rom.ROM) *Game {
	if cyclesPerFrame <= 0 {
		cyclesPerFrame = 12 // Default value if not provided or invalid
	}
	if cpu == nil {
		cpu = cpuImpl.NewCPU(cpuImpl.XOChip, display) // Default to XOChip if not provided
	}
	src, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		log.Fatal(err)
	}
	Face := &text.GoTextFace{
		Source: src,
		Size:   20, // pixel size
	}
	m := Face.Metrics()
	lineHeight := m.HAscent + m.HDescent + m.HLineGap
	g := &Game{
		CPU:            cpu,
		EmuDisplay:     display,
		beepStream:     beepStream,
		beepPlayer:     beepPlayer,
		cyclesPerFrame: cyclesPerFrame,
		debugFace:      Face,
		lineHeight:     lineHeight,
		currentROM:     rom.Data,
		isRomLoaded:    len(rom.Data) > 0,
	}

	g.debugToggleBtn = &Button{
		Rect:  image.Rect(0, 0, 120, 32),
		Label: "Debug",
		OnClick: func() {
			g.showDebug = !g.showDebug
		},
	}

	g.loadROMBtn = &Button{
		Rect:  image.Rect(0, 0, 120, 32),
		Label: "Load ROM",
		OnClick: func() {
			g.OnLoadROMClick()
		},
	}
	quirks, names := cpuImpl.GetAvailableQuirks()
	g.cpuTypeSelector = NewCPUTypeSelector(
		quirks,
		names,
		func(t cpuImpl.Quirks) {
			g.switchCPUType(t)
		},
	)
	slog.Debug("Game initialized", "cyclesPerFrame", cyclesPerFrame, "quirks", cpu.Quirks, "romSize", len(rom.Data), "currentROM", len(g.currentROM))
	return g
}

func (g *Game) Update() error {

	sw, sh := display.WindowSizeW, display.WindowSizeH
	g.layoutButtons(sw, sh)

	g.debugToggleBtn.Update()
	g.loadROMBtn.Update()
	g.cpuTypeSelector.Update()

	if inpututil.IsKeyJustPressed(ebiten.KeyF9) {
		g.showDebug = !g.showDebug
	}

	if files := ebiten.DroppedFiles(); files != nil {
		g.romMu.Lock()
		if !g.loadingROM {
			g.loadingROM = true
			go g.loadROMFromFS(files)
		}
		g.romMu.Unlock()
	}

	g.romMu.Lock()
	rom := g.pendingROM
	g.pendingROM = nil
	err := g.romLoadErr
	g.romLoadErr = nil
	g.romMu.Unlock()

	if err != nil {
		slog.Error("ROM load error", "error", err)
	}

	if rom != nil {
		g.currentROM = rom
		g.loadROM(rom)
	}
	if g.isRomLoaded {
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
			g.EmuDisplay.UpdateFromMemory(g.CPU.Display[:], g.CPU.Display2[:], g.CPU.HighRes)
			g.CPU.DrawFlag = false
		}
	}
	keys := keyboard.PollKeys()
	g.CPU.SetKeys(keys)
	return nil
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return display.WindowSizeW, display.WindowSizeH
}

func (g *Game) Draw(screen *ebiten.Image) {
	emuW, emuH := g.EmuDisplay.Size()
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()

	g.layoutButtons(sw, sh)

	scale := float64(sw) / float64(emuW)
	if s2 := float64(sh) / float64(emuH); s2 < scale {
		scale = s2
	}
	offsetX := (float64(sw) - float64(emuW)*scale) / 2
	offsetY := (float64(sh) - float64(emuH)*scale) / 2

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(offsetX, offsetY)
	op.Filter = ebiten.FilterNearest

	g.EmuDisplay.DrawScaled(screen, op)

	if !g.isRomLoaded {
		msg := `Please load a ROM (.ch8/.c8/.bin) or use the "Load ROM" button`

		drawCenteredTextBox(
			screen,
			msg,
			g.debugFace,
			g.lineHeight,
			float64(screen.Bounds().Dx()),
			float64(screen.Bounds().Dy()),
		)
	}

	if g.showDebug {
		g.drawDebugOverlay(screen, sw, sh)
	}

	// ui elements part
	mx, my := ebiten.CursorPosition()
	hovered := g.debugToggleBtn.Contains(mx, my)
	g.debugToggleBtn.Draw(screen, g.debugFace, hovered)
	hovered = g.loadROMBtn.Contains(mx, my)
	g.loadROMBtn.Draw(screen, g.debugFace, hovered)
	hovered = g.cpuTypeSelector.Contains(mx, my)
	g.cpuTypeSelector.Draw(screen, g.debugFace, hovered)
}

func (g *Game) loadROMFromFS(files fs.FS) {
	slog.Debug("Loading ROM from dropped files", "fs_type", fmt.Sprintf("%T", files))

	err := fs.WalkDir(files, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			slog.Error("walk error", "error", err)
			return err
		}

		slog.Debug("ENTRY", "path", path, "isDir", d.IsDir())

		return nil
	})

	slog.Debug("walk finished", "error", err)
	defer func() {
		g.romMu.Lock()
		g.loadingROM = false
		g.romMu.Unlock()
	}()
}
