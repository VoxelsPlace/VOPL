//go:build js && wasm

package main

import (
	"syscall/js"

	"github.com/voxelsplace/vopl/go/api"
)

func rle2vopl(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		return js.ValueOf("missing rle string")
	}
	rle := args[0].String()
	out, err := api.RLEToVOPLBytes(rle)
	if err != nil {
		return js.ValueOf(err.Error())
	}
	uint8arr := js.Global().Get("Uint8Array").New(len(out))
	js.CopyBytesToJS(uint8arr, out)
	return uint8arr
}

func vopl2glb(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		return js.ValueOf("missing vopl bytes")
	}
	buf := make([]byte, args[0].Get("length").Int())
	js.CopyBytesToGo(buf, args[0])
	out, err := api.VOPLToGLB(buf)
	if err != nil {
		return js.ValueOf(err.Error())
	}
	uint8arr := js.Global().Get("Uint8Array").New(len(out))
	js.CopyBytesToJS(uint8arr, out)
	return uint8arr
}

func packVopls(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		return js.ValueOf("missing files object")
	}
	filesObj := args[0]
	files := map[string][]byte{}
	keys := js.Global().Get("Object").Call("keys", filesObj)
	for i := 0; i < keys.Length(); i++ {
		k := keys.Index(i).String()
		v := filesObj.Get(k)
		b := make([]byte, v.Get("length").Int())
		js.CopyBytesToGo(b, v)
		files[k] = b
	}
	out, err := api.PackVOPLs(files)
	if err != nil {
		return js.ValueOf(err.Error())
	}
	uint8arr := js.Global().Get("Uint8Array").New(len(out))
	js.CopyBytesToJS(uint8arr, out)
	return uint8arr
}

func unpackVoplpack(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		return js.ValueOf("missing pack bytes")
	}
	buf := make([]byte, args[0].Get("length").Int())
	js.CopyBytesToGo(buf, args[0])
	files, err := api.UnpackVOPLPACKToMemory(buf)
	if err != nil {
		return js.ValueOf(err.Error())
	}
	// return an object mapping names->Uint8Array
	result := js.Global().Get("Object").New()
	for name, b := range files {
		arr := js.Global().Get("Uint8Array").New(len(b))
		js.CopyBytesToJS(arr, b)
		result.Set(name, arr)
	}
	return result
}

func main() {
	js.Global().Set("rle2vopl", js.FuncOf(rle2vopl))
	js.Global().Set("vopl2glb", js.FuncOf(vopl2glb))
	js.Global().Set("packVopls", js.FuncOf(packVopls))
	js.Global().Set("unpackVoplpack", js.FuncOf(unpackVoplpack))
	select {}
}
