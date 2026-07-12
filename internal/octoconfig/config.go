package octoconfig

import (
	"bufio"
	"fmt"
	"image/color"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	// ui.*
	Windowed       bool
	SoftwareRender bool
	WinScale       int
	WinWidth       int
	WinHeight      int
	Volume         int

	// core.*
	TickRate  int
	MaxROM    int
	Rotation  int
	Font      string
	TouchMode string

	// color.*
	Plane0     color.RGBA
	Plane1     color.RGBA
	Plane2     color.RGBA
	Plane3     color.RGBA
	Background color.RGBA
	Sound      color.RGBA

	// quirks.*
	QuirkShift     bool
	QuirkLoadStore bool
	QuirkJump0     bool
	QuirkLogic     bool
	QuirkClip      bool
	QuirkVBlank    bool
}

func Default() Config {
	return Config{
		Windowed:  true,
		WinScale:  2,
		WinWidth:  480,
		WinHeight: 272,
		Volume:    4,
		TickRate:  2000,
		MaxROM:    65024,
		Rotation:  0,
		Font:      "octo",
		TouchMode: "none",

		Plane0:     hexColor("000000"),
		Plane1:     hexColor("FFFFFF"),
		Plane2:     hexColor("FF6600"),
		Plane3:     hexColor("662200"),
		Background: hexColor("000000"),
		Sound:      hexColor("000000"),
	}
}

func Load(path string) (Config, error) {
	cfg := Default()

	f, err := os.Open(path)
	if err != nil {
		return cfg, fmt.Errorf("failed to open octo config: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return cfg, fmt.Errorf("octo config line %d: malformed entry %q", lineNum, line)
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		if err := cfg.apply(key, value); err != nil {
			return cfg, fmt.Errorf("octo config line %d: %w", lineNum, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return cfg, fmt.Errorf("failed reading octo config: %w", err)
	}

	return cfg, nil
}

func (cfg *Config) apply(key, value string) error {
	switch key {
	case "ui.windowed":
		return setBool(&cfg.Windowed, value)
	case "ui.software_render":
		return setBool(&cfg.SoftwareRender, value)
	case "ui.win_scale":
		return setInt(&cfg.WinScale, value)
	case "ui.win_width":
		return setInt(&cfg.WinWidth, value)
	case "ui.win_height":
		return setInt(&cfg.WinHeight, value)
	case "ui.volume":
		return setInt(&cfg.Volume, value)

	case "core.tickrate":
		return setInt(&cfg.TickRate, value)
	case "core.max_rom":
		return setInt(&cfg.MaxROM, value)
	case "core.rotation":
		return setInt(&cfg.Rotation, value)
	case "core.font":
		cfg.Font = value
	case "core.touch_mode":
		cfg.TouchMode = value

	case "color.plane0":
		cfg.Plane0 = hexColor(value)
	case "color.plane1":
		cfg.Plane1 = hexColor(value)
	case "color.plane2":
		cfg.Plane2 = hexColor(value)
	case "color.plane3":
		cfg.Plane3 = hexColor(value)
	case "color.background":
		cfg.Background = hexColor(value)
	case "color.sound":
		cfg.Sound = hexColor(value)

	case "quirks.shift":
		return setBool(&cfg.QuirkShift, value)
	case "quirks.loadstore":
		return setBool(&cfg.QuirkLoadStore, value)
	case "quirks.jump0":
		return setBool(&cfg.QuirkJump0, value)
	case "quirks.logic":
		return setBool(&cfg.QuirkLogic, value)
	case "quirks.clip":
		return setBool(&cfg.QuirkClip, value)
	case "quirks.vblank":
		return setBool(&cfg.QuirkVBlank, value)

	default:
		// unknown key, ignore it
	}
	return nil
}

func setBool(dst *bool, raw string) error {
	n, err := strconv.Atoi(raw)
	if err != nil {
		return fmt.Errorf("expected 0 or 1, got %q: %w", raw, err)
	}
	*dst = n != 0
	return nil
}

func setInt(dst *int, raw string) error {
	n, err := strconv.Atoi(raw)
	if err != nil {
		return fmt.Errorf("expected integer, got %q: %w", raw, err)
	}
	*dst = n
	return nil
}

func hexColor(hex string) color.RGBA {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return color.RGBA{A: 0xFF}
	}
	r, _ := strconv.ParseUint(hex[0:2], 16, 8)
	g, _ := strconv.ParseUint(hex[2:4], 16, 8)
	b, _ := strconv.ParseUint(hex[4:6], 16, 8)
	return color.RGBA{R: byte(r), G: byte(g), B: byte(b), A: 0xFF}
}
