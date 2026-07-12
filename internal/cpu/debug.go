package cpu

import "fmt"

type Debug struct {
	Enabled bool
	Step    bool

	Cycle uint64
}

func hex(v uint16) string {
	return fmt.Sprintf("%X", v)
}
