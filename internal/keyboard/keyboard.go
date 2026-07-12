package keyboard

import "github.com/hajimehoshi/ebiten/v2"

func PollKeys() [16]bool {
	var k [16]bool

	k[0x1] = ebiten.IsKeyPressed(ebiten.Key1)
	k[0x2] = ebiten.IsKeyPressed(ebiten.Key2)
	k[0x3] = ebiten.IsKeyPressed(ebiten.Key3)
	k[0xC] = ebiten.IsKeyPressed(ebiten.Key4)

	k[0x4] = ebiten.IsKeyPressed(ebiten.KeyQ)
	k[0x5] = ebiten.IsKeyPressed(ebiten.KeyW)
	k[0x6] = ebiten.IsKeyPressed(ebiten.KeyE) || ebiten.IsKeyPressed(ebiten.KeySpace)
	k[0xD] = ebiten.IsKeyPressed(ebiten.KeyR)

	k[0x7] = ebiten.IsKeyPressed(ebiten.KeyA)
	k[0x8] = ebiten.IsKeyPressed(ebiten.KeyS)
	k[0x9] = ebiten.IsKeyPressed(ebiten.KeyD)
	k[0xE] = ebiten.IsKeyPressed(ebiten.KeyF)

	k[0xA] = ebiten.IsKeyPressed(ebiten.KeyZ)
	k[0x0] = ebiten.IsKeyPressed(ebiten.KeyX)
	k[0xB] = ebiten.IsKeyPressed(ebiten.KeyC)
	k[0xF] = ebiten.IsKeyPressed(ebiten.KeyV)

	return k
}
