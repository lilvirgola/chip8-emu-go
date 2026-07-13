//go:build js

package emulator

import (
	"syscall/js"
)

func (g *Game) OnLoadROMClick() {
	document := js.Global().Get("document")
	input := document.Call("getElementById", "romFileInput")

	if input.IsUndefined() || input.IsNull() {
		input = document.Call("createElement", "input")
		input.Set("type", "file")
		input.Set("id", "romFileInput")
		input.Set("style", "display:none")
		document.Get("body").Call("appendChild", input)

		input.Set("onchange", js.FuncOf(func(this js.Value, args []js.Value) any {
			files := input.Get("files")
			if files.Get("length").Int() == 0 {
				return nil
			}
			file := files.Index(0)

			reader := js.Global().Get("FileReader").New()
			reader.Set("onload", js.FuncOf(func(this js.Value, args []js.Value) any {
				result := reader.Get("result") // ArrayBuffer
				data := make([]byte, result.Get("byteLength").Int())
				js.CopyBytesToGo(data, js.Global().Get("Uint8Array").New(result))
				g.loadROM(data)
				return nil
			}))
			reader.Call("readAsArrayBuffer", file)
			return nil
		}))
	}

	input.Call("click")
}
