## VOPL File Format and VOPLPACK

This document specifies the binary layout and behavior of `.vopl` and `.voplpack` exactly as implemented here. It covers field sizes, endianness, bit-level packing, voxel ordering (3D Morton/Z-order), encodings, compression, palette mapping, container, and runnable examples in Go and JavaScript.

### Scope


## CLI and packages

This module is now library-first and exposes a thin CLI under `cmd/vopltool`.

- Run without installing:

  - `go run ./cmd/vopltool --help`
  - `go run ./cmd/vopltool updatevopl input.vopl updates.json output.vopl`

- Install the CLI:

  - `go install github.com/voxelsplace/vopl/go/cmd/vopltool@latest`
  - Then invoke `vopltool` from your `$GOBIN` (`~/go/bin` by default).

- Importable packages:

  - `github.com/voxelsplace/vopl/go/vopl` — core format code (encode/decode, mesh, pack, etc.)
  - `github.com/voxelsplace/vopl/go/api` — byte-oriented helpers for web/wasm and programmatic use
  - `github.com/voxelsplace/vopl/go/utils` — filesystem-oriented helpers used by the CLI



## .vopl (grid format)

### Header (16 bytes total, including magic)
  - bit7 (0x80): 1 if payload is zlib-compressed; 0 otherwise
  - bits[6:0]: encoding id: 0=dense, 1=sparse, 2=rle

Immediately after the 16-byte header, `plen` bytes of payload follow.

### Payload encodings (N=16*16*16=4096)
Values are emitted in 3D Morton/Z-order (see Ordering).

  - Sequence of N values, each `bpp` bits, tightly bit-packed.
  - count: 16 bits (# of nonzero entries)
  - Repeated `count` times:
    - idx: 8 bits (Morton position index, 0..255)
    - value: `bpp` bits
  - Note: idx is 8-bit. Only first 256 Morton positions are addressable in this legacy sparse mode.
  - Repeat until N values reconstructed:
    - run_minus_1: 8 bits (actual run length = run_minus_1+1, range 1..256)
    - value: `bpp` bits

  - 4096-bit occupancy bitmap (512 bytes), LSB-first within each byte, Morton order.
  - Followed by bit-packed stream of all nonzero values, each `bpp` bits, in Morton order.

  - Sequence of blocks until N values reconstructed:
    - flag: 1 byte (0 = literal block, 1 = zero-run block)
    - len: unsigned varint (LEB128-style, 7-bit groups)
    - if literal: then `len` values follow, bit-packed at `bpp` bits each
    - if zero-run: no payload for the block; it expands to `len` zeros

If `enc & 0x80 != 0`, the raw stream above is zlib-compressed (BestCompression); decompress before decoding.

### Ordering (3D Morton/Z-order)
Grid index access is `grid[y][x][z]` with 0-based `x∈[0,16)`, `y∈[0,16)`, `z∈[0,16)`.
The linear stream order is ascending by key `morton3D(x,y,z) = expand3(x) | (expand3(y)<<1) | (expand3(z)<<2)` where `expand3` spreads the lower 8 bits of a value into every third bit position.

### Bit packing (LSB-first)

### Validation



## Palette (64 entries)

Index 0 is transparent RGBA. 1..63 are opaque RGB unless noted.



## VOPLPACK (.voplpack)

Bundles multiple `.vopl` payloads with shared common header.

### File header

### Content section (uncompressed view)
  - ver: uint8 (must be 3)
  - bpp: uint8
  - w: uint8
  - h: uint8
  - d: uint8
  - pal: uint16
  - nameLen: uint16
  - name: `nameLen` bytes
  - enc: uint8 (same semantics as `.vopl` enc)
  - plen: uint32
  - payload: `plen` bytes (raw stream; may itself be zlib if enc bit7 set)

No footer, no checksums.


## Go examples

Dense writer for a 16³ grid of palette indices (already in Morton order):

```go
package main
import (
	"bytes"
	"encoding/binary"
)
func writeVOPDense(grid []uint8) []byte {
	bpp := uint8(6)
	w,h,d := uint8(16),uint8(16),uint8(16)
	pal := uint16(64)
	enc := uint8(0)
	acc := uint64(0)
	nbits := uint8(0)
	payload := make([]byte,0,3072)
	for _,v := range grid {
		acc |= (uint64(v)&((1<<bpp)-1))<<nbits
		nbits += bpp
		for nbits >= 8 {
			payload = append(payload, byte(acc))
			acc >>= 8
			nbits -= 8
		}
	}
	if nbits>0 { payload = append(payload, byte(acc)) }
	var buf bytes.Buffer
	buf.WriteString("VOPL")
	binary.Write(&buf, binary.LittleEndian, uint8(3))
	binary.Write(&buf, binary.LittleEndian, enc)
	binary.Write(&buf, binary.LittleEndian, bpp)
	binary.Write(&buf, binary.LittleEndian, w)
	binary.Write(&buf, binary.LittleEndian, h)
	binary.Write(&buf, binary.LittleEndian, d)
	binary.Write(&buf, binary.LittleEndian, pal)
	binary.Write(&buf, binary.LittleEndian, uint32(len(payload)))
	buf.Write(payload)
	return buf.Bytes()
}
```

RLE encoder (enc=2, no zlib):

```go
func encodeRLEV3(stream []uint8, bpp uint8) []byte {
	acc := uint64(0)
	nbits := uint8(0)
	out := make([]byte,0)
	emit := func(v uint64, bits uint8) {
		acc |= (v & ((1<<bits)-1)) << nbits
		nbits += bits
		for nbits >= 8 {
			out = append(out, byte(acc))
			acc >>= 8
			nbits -= 8
		}
	}
	if len(stream)==0 { return out }
	cur := stream[0]
	run := 1
	flush := func() { emit(uint64(run-1),8); emit(uint64(cur),bpp) }
	for i:=1;i<len(stream);i++ {
		if stream[i]==cur && run<256 { run++; continue }
		flush(); cur=stream[i]; run=1
	}
	flush()
	if nbits>0 { out = append(out, byte(acc)) }
	return out
}
```


## JavaScript examples
Dense writer (Morton order indices computed on the fly; no zlib):

```js
function writeBitsLSB(acc, v, bits, out) {
  acc.v |= (v & ((1n<<BigInt(bits))-1n)) << acc.n;
  acc.n += BigInt(bits);
  while (acc.n >= 8n) {
    out.push(Number(acc.v & 0xFFn));
    acc.v >>= 8n;
    acc.n -= 8n;
  }
}
function flushLSB(acc, out) { if (acc.n>0n) { out.push(Number(acc.v & 0xFFn)); acc.v=0n; acc.n=0n; } }
function mortonExpand3(v){
  v = (v | (v<<16n)) & 0xFF0000FFn;
  v = (v | (v<<8n)) & 0x0F00F00Fn;
  v = (v | (v<<4n)) & 0xC30C30C3n;
  v = (v | (v<<2n)) & 0x49249249n;
  return v;
}
function morton3D(x,y,z){
  return mortonExpand3(BigInt(x)) | (mortonExpand3(BigInt(y))<<1n) | (mortonExpand3(BigInt(z))<<2n);
}
function mortonOrder16() {
  const idx=[]; let i=0;
  for(let y=0;y<16;y++) for(let z=0;z<16;z++) for(let x=0;x<16;x++) idx.push({k:Number(morton3D(x,y,z)),i:i++});
  idx.sort((a,b)=>a.k-b.k);
  return idx.map(e=>e.i);
}
function encodeDenseV3(grid){
  const order=mortonOrder16();
  const out=[]; const acc={v:0n,n:0n};
  for(let j=0;j<order.length;j++) writeBitsLSB(acc, BigInt(grid[order[j]]), 6, out);
  flushLSB(acc,out); return Uint8Array.from(out);
}
function buildVOPLDense(grid){
  const payload=encodeDenseV3(grid);
  const bytes=new Uint8Array(16+payload.length);
  bytes.set([0x56,0x4F,0x50,0x4C, 3, 0, 6, 16, 16, 16, 64, 0, 0, 0, 0, 0], 0);
  const dv=new DataView(bytes.buffer);
  dv.setUint32(12,payload.length,true);
  bytes.set(payload,16);
  return bytes;
}
```


## Edge cases and constraints
- Sparse idx is 8-bit; only first 256 Morton positions are addressable in sparse mode here.
- RLE max run is 256; split longer runs.
- bpp must be ≤8; used value: 6.
- Decoder expects exactly `plen` bytes of payload after header.
- Decoder ignores w/h/d from the header and reconstructs 16×16×16.

- Optionally apply zlib to the raw encoding stream if it reduces size; set `enc|=0x80`.
- Write/read headers exactly as specified.
