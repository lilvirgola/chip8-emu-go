package audio

import (
	"io"
	"log/slog"
	"math"
	"sync"
	"sync/atomic"
)

const (
	freq       = 440
	sampleRate = 44100
)

var _ io.Reader = (*Beep)(nil)

type Beep struct {
	active atomic.Bool

	acc        uint32 // accumulator for fallback square wave
	mu         sync.Mutex
	pattern    [16]byte // 128-bit XO-CHIP audio pattern buffer
	hasPattern bool
	pitch      byte    // XO-CHIP FX3A pitch value; 64 = 4000Hz base rate
	phase      float64 // fractional position into the 128-bit pattern, in bits
}

func NewBeep() *Beep {
	slog.Debug("Beep initialized")
	return &Beep{pitch: 64, acc: 0}
}

func (b *Beep) SetPattern(pattern [16]byte) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.pattern = pattern
	b.hasPattern = true
	slog.Debug("Beep pattern set", "pattern", pattern)
}

func (b *Beep) SetPitch(pitch byte) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.pitch = pitch
	slog.Debug("Beep pitch set", "pitch", pitch)
}

func (b *Beep) playbackRate() float64 {
	playbackRate := 4000.0 * math.Pow(2, (float64(b.pitch)-64.0)/48.0)
	slog.Debug("Beep playback rate calculated", "pitch", b.pitch, "rate", playbackRate)
	return playbackRate
}

func (b *Beep) Read(out []byte) (int, error) {
	// Ebiten expects 16-bit stereo PCM (4 bytes per frame)
	frames := len(out) / 4
	if frames == 0 {
		slog.Debug("Beep Read called with empty output buffer")
		return 0, nil
	}

	if !b.active.Load() {
		// Fill with silence (https://www.youtube.com/watch?v=u9Dg-g7t2l4)
		for i := range out {
			out[i] = 0
		}
		slog.Debug("Beep is inactive, outputting silence")
		return len(out), nil
	}

	b.mu.Lock()
	hasPattern := b.hasPattern
	pattern := b.pattern
	rate := b.playbackRate()
	phase := b.phase
	acc := b.acc
	b.mu.Unlock()

	if !hasPattern {
		const freq = 440.0
		stepf := freq * (1 << 16) / sampleRate
		step := uint32(stepf)

		// Loop over FRAMES, not bytes!
		for i := 0; i < frames; i++ {
			acc += step
			bit := (acc >> 15) & 1

			var sample int16
			if bit == 0 {
				sample = -32768
			} else {
				sample = 32767
			}
			out[i*4+0] = byte(sample)
			out[i*4+1] = byte(sample >> 8)
			out[i*4+2] = byte(sample)
			out[i*4+3] = byte(sample >> 8)
		}

		b.mu.Lock()
		b.acc = acc
		b.mu.Unlock()
		slog.Debug("Beep is active, outputting fallback square wave")
		return len(out), nil
	}

	// XO-CHIP pattern playback
	// bitsPerFrame is how many bits of the 128-bit pattern we consume per audio frame
	bitsPerFrame := rate / float64(sampleRate)

	for i := 0; i < frames; i++ {
		bitIndex := int(phase) % 128
		byteIndex := bitIndex / 8
		bitOffset := 7 - (bitIndex % 8)
		bit := (pattern[byteIndex] >> bitOffset) & 1

		var sample int16
		if bit == 0 {
			sample = -32768
		} else {
			sample = 32767
		}

		out[i*4+0] = byte(sample)
		out[i*4+1] = byte(sample >> 8)
		out[i*4+2] = byte(sample)
		out[i*4+3] = byte(sample >> 8)

		phase += bitsPerFrame
		if phase >= 128 {
			phase -= 128
		}
	}

	b.mu.Lock()
	b.phase = phase
	b.acc = acc // Save acc just in case :)
	b.mu.Unlock()
	slog.Debug("Beep is active, outputting XO-CHIP pattern", "pattern", pattern, "phase", phase)
	return len(out), nil
}

func (b *Beep) Err() error {
	slog.Debug("Beep Err called, returning nil")
	return nil
}

func (b *Beep) Activate() {
	slog.Debug("Beep activated")
	b.active.Store(true)
}

func (b *Beep) Deactivate() {
	slog.Debug("Beep deactivated")
	b.active.Store(false)
}
