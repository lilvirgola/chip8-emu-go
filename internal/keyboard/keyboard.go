package keyboard

import "github.com/hajimehoshi/ebiten/v2"

func PollKeys() [16]bool {
	var k [16]bool

	k[0x1] = ebiten.IsKeyPressed(ebiten.Key1)
	k[0x2] = ebiten.IsKeyPressed(ebiten.Key2)
	k[0x3] = ebiten.IsKeyPressed(ebiten.Key3)
	k[0xC] = ebiten.IsKeyPressed(ebiten.KeyC)

	k[0x4] = ebiten.IsKeyPressed(ebiten.Key4)
	k[0x5] = ebiten.IsKeyPressed(ebiten.Key5)
	k[0x6] = ebiten.IsKeyPressed(ebiten.Key6)
	k[0xD] = ebiten.IsKeyPressed(ebiten.KeyD)

	k[0x7] = ebiten.IsKeyPressed(ebiten.Key7)
	k[0x8] = ebiten.IsKeyPressed(ebiten.Key8)
	k[0x9] = ebiten.IsKeyPressed(ebiten.Key9)
	k[0xE] = ebiten.IsKeyPressed(ebiten.KeyE)

	k[0xA] = ebiten.IsKeyPressed(ebiten.KeyA)
	k[0x0] = ebiten.IsKeyPressed(ebiten.Key0)
	k[0xB] = ebiten.IsKeyPressed(ebiten.KeyB)
	k[0xF] = ebiten.IsKeyPressed(ebiten.KeyF)

	return k
}
