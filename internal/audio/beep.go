package audio

import (
	"io"
	"sync/atomic"
)

const (
	freq       = 440
	sampleRate = 44100
)

var _ io.Reader = (*Beep)(nil)

type Beep struct {
	acc    uint32
	active atomic.Bool
}

func NewBeep() *Beep {
	return &Beep{acc: 0, active: atomic.Bool{}}
}

func (b *Beep) Read(out []byte) (int, error) {

	if !b.active.Load() {
		for i := range out {
			out[i] = 0
		}
		return len(out), nil
	}

	// square wave ~440Hz
	freq := 440.0
	step := uint32(freq * (1 << 16) / sampleRate)

	for i := 0; i < len(out); i++ {
		b.acc += step
		bit := (b.acc >> 16) & 1
		if bit == 0 {
			out[i] = 127
		} else {
			out[i] = 128
		}
	}

	return len(out), nil
}

func (b *Beep) Err() error {
	return nil
}

func (b *Beep) Activate() {
	b.active.Store(true)
}

func (b *Beep) Deactivate() {
	b.active.Store(false)
}
