package octoconfig

import "chip8/internal/cpu"

func (cfg Config) ToQuirks() cpu.Quirks {
	return cpu.Quirks{
		ShiftSetVX:  cfg.QuirkShift,
		IncrementI:  !cfg.QuirkLoadStore,
		Jumping:     cfg.QuirkJump0,
		VFReset:     !cfg.QuirkLogic,
		Wrapping:    !cfg.QuirkClip,
		DisplayWait: cfg.QuirkVBlank,
		HalfScroll:  false,
	}
}
