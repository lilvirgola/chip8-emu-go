package rom

import (
	"fmt"
	"os"
)

type ROM struct {
	Data []byte
}

func Load(path string, memSize int) (*ROM, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read rom: %w", err)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("empty rom")
	}
	if len(data) > memSize-0x200 {
		return nil, fmt.Errorf("rom too large: %d bytes (max %d)", len(data), memSize-0x200)
	}
	return &ROM{Data: data}, nil
}
