//go:build js && wasm

package main

import (
	"syscall/js"

	"github.com/voxelsplace/vopl/go/api"
	"github.com/voxelsplace/vopl/go/vopl"
)

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

// decodeVopl takes a Uint8Array containing a .vopl file and returns a JS object:
//
//	{
//	  header: { ver, bpp, w, h, d, pal, payloadLength },
//	  grid: Uint8Array(Width*Height*Depth) with linear order (y-major: y,x,z)
//	}
func decodeVopl(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		return js.ValueOf("missing vopl bytes")
	}
	// Copy input bytes from JS to Go
	in := args[0]
	buf := make([]byte, in.Get("length").Int())
	js.CopyBytesToGo(buf, in)

	// Parse header for metadata
	hdr, _, err := vopl.ParseVOPLHeaderFromBytes(buf)
	if err != nil {
		return js.ValueOf(err.Error())
	}

	// Decode voxel grid
	grid, err := vopl.LoadVoplGridFromBytes(buf)
	if err != nil {
		return js.ValueOf(err.Error())
	}

	// Flatten to a linear Uint8Array with order (y, x, z)
	total := vopl.Width * vopl.Height * vopl.Depth
	flat := make([]byte, total)
	p := 0
	for y := 0; y < vopl.Height; y++ {
		for x := 0; x < vopl.Width; x++ {
			for z := 0; z < vopl.Depth; z++ {
				flat[p] = grid[y][x][z]
				p++
			}
		}
	}

	// Build JS return object
	result := js.Global().Get("Object").New()
	header := js.Global().Get("Object").New()
	header.Set("ver", int(hdr.Ver))
	header.Set("bpp", int(hdr.BPP))
	header.Set("w", int(hdr.W))
	header.Set("h", int(hdr.H))
	header.Set("d", int(hdr.D))
	header.Set("pal", int(hdr.Pal))
	header.Set("payloadLength", int(hdr.PLen))
	result.Set("header", header)

	arr := js.Global().Get("Uint8Array").New(len(flat))
	js.CopyBytesToJS(arr, flat)
	result.Set("grid", arr)

	return result
}

func main() {
	js.Global().Set("vopl2glb", js.FuncOf(vopl2glb))
	js.Global().Set("packVopls", js.FuncOf(packVopls))
	js.Global().Set("unpackVoplpack", js.FuncOf(unpackVoplpack))
	js.Global().Set("decodeVopl", js.FuncOf(decodeVopl))
	select {}
}
