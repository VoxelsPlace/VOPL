## VOPL File Format (v1, v2, v3) and VOPLPACK

This document specifies the binary layout and behavior of `.vopl` and `.voplpack` exactly as implemented here. It covers field sizes, endianness, bit-level packing, voxel ordering (3D Morton/Z-order), encodings, compression, palette mapping, container, and runnable examples in Go and JavaScript.

### Scope
- All multi-byte integers are little-endian.
- Voxel value = palette index. 0 = empty/transparent; nonzero = solid.
- The implementation uses fixed chunk size `Width=16`, `Height=16`, `Depth=16`. v2/v3 headers carry w/h/d but the decoder assumes 16³ regardless.
- Bitstreams are LSB-first within each byte.


## .vopl v3 (grid format)

### Header (16 bytes total, including magic)
- magic: 4 bytes ASCII `VOPL`
- ver: uint8 = 3
- enc: uint8
  - bit7 (0x80): 1 if payload is zlib-compressed; 0 otherwise
  - bits[6:0]: encoding id: 0=dense, 1=sparse, 2=rle
- bpp: uint8 (here: 6 for 64-color palette 0..63)
- w: uint8 (encoded, decoder assumes 16)
- h: uint8 (encoded, decoder assumes 16)
- d: uint8 (encoded, decoder assumes 16)
- pal: uint16 (here: 64)
- plen: uint32 (payload length in bytes)

Immediately after the 16-byte header, `plen` bytes of payload follow.

### Payload encodings (N=16*16*16=4096)
Values are emitted in 3D Morton/Z-order (see Ordering).

- Encoding 0: dense
  - Sequence of N values, each `bpp` bits, tightly bit-packed.
- Encoding 1: sparse
  - count: 16 bits (# of nonzero entries)
  - Repeated `count` times:
    - idx: 8 bits (Morton position index, 0..255)
    - value: `bpp` bits
  - Note: idx is 8-bit. Only first 256 Morton positions are addressable in sparse mode as implemented.
- Encoding 2: rle
  - Repeat until N values reconstructed:
    - run_minus_1: 8 bits (actual run length = run_minus_1+1, range 1..256)
    - value: `bpp` bits

If `enc & 0x80 != 0`, the raw stream above is zlib-compressed; decompress before decoding.

### Ordering (3D Morton/Z-order)
Grid index access is `grid[y][x][z]` with 0-based `x∈[0,16)`, `y∈[0,16)`, `z∈[0,16)`.
The linear stream order is ascending by key `morton3D(x,y,z) = expand3(x) | (expand3(y)<<1) | (expand3(z)<<2)` where `expand3` spreads the lower 8 bits of a value into every third bit position.

### Bit packing (LSB-first)
- Writer: OR low `bits` of value into an accumulator shifted by current count; emit bytes from the low 8 bits as soon as available; flush residual low bits as a final byte.
- Reader: Refill by OR-ing next byte shifted by current count; take low `bits`; shift right by `bits`.

### Validation
- Magic must be `VOPL`, ver must be 3.
- `plen` must equal remaining file length after header.
- If compressed, payload must be valid zlib.
- Decoder reconstructs a 16×16×16 grid, ignoring w/h/d.


## .vopl v2 (legacy grid)

### Header (15 bytes total, including magic)
- magic: 4 bytes ASCII `VOPL`
- ver: uint8 = 2
- enc: uint8 (bit7=zlib, bits[6:0]=encoding id 0/1/2)
- w: uint8 (ignored by decoder; assumes 16)
- h: uint8 (ignored)
- d: uint8 (ignored)
- pal: uint16
- plen: uint32

### Payload encodings
- bpp is implicitly 5 (palette indices 0..31)
- Dense: N values, 5 bits each
- Sparse: count: 8 bits; then `count`× (idx: 8 bits, value: 5 bits)
- RLE: repeated (run_minus_1: 8 bits, value: 5 bits) until N values
- Optional zlib if `enc & 0x80 != 0`


## .vopl v1 (legacy mesh)

### Layout
- magic: 4 bytes ASCII `VOPL`
- ver: uint8 = 1
- numVerts: uint16
- numIdx: uint16
- vertices: `numVerts` × (position: 3×float32, color: uint8)
- indices: `numIdx` × uint16

This stores a greedy-meshed surface, not a grid.


## Palette (64 entries)

Index 0 is transparent RGBA. 1..63 are opaque RGB unless noted.

- 0: #00000000
- 1: #000000
- 2: #3C3C3C
- 3: #787878
- 4: #D2D2D2
- 5: #FFFFFF
- 6: #600018
- 7: #ED1C24
- 8: #FF7F27
- 9: #F6AA09
- 10: #F9DD3B
- 11: #FFFABC
- 12: #0EB968
- 13: #13E67B
- 14: #87FF5E
- 15: #0C816E
- 16: #10AEA6
- 17: #13E1BE
- 18: #28509E
- 19: #4093E4
- 20: #60F7F2
- 21: #6B50F6
- 22: #99B1FB
- 23: #780C99
- 24: #AA38B9
- 25: #E09FF9
- 26: #CB007A
- 27: #EC1F80
- 28: #F38DA9
- 29: #684634
- 30: #95682A
- 31: #F8B277
- 32: #AAAAAA
- 33: #A50E1E
- 34: #FA8072
- 35: #E45C1A
- 36: #D6B594
- 37: #9C8431
- 38: #C5AD31
- 39: #E8D45F
- 40: #4A6B3A
- 41: #5A944A
- 42: #84C573
- 43: #0F799F
- 44: #BBFAF2
- 45: #7DC7FF
- 46: #4D31B8
- 47: #4A4284
- 48: #7A71C4
- 49: #B5AEF1
- 50: #DBA463
- 51: #D18051
- 52: #FFC5A5
- 53: #9B5249
- 54: #D18078
- 55: #FAB6A4
- 56: #7B6352
- 57: #9C846B
- 58: #333941
- 59: #6D758D
- 60: #B3B9D1
- 61: #6D643F
- 62: #948C6B
- 63: #CDC59E


## VOPLPACK (.voplpack)

Bundles multiple `.vopl` v3 payloads with shared common header.

### File header
- magic: 8 bytes ASCII `VOPLPACK`
- packVersion: uint8 = 1
- compression: uint8 (0=none, 1=zlib). If 1, the entire content section below is zlib-compressed.

### Content section (uncompressed view)
- common header subset:
  - ver: uint8 (must be 3)
  - bpp: uint8
  - w: uint8
  - h: uint8
  - d: uint8
  - pal: uint16
- n: uint32 (# entries)
- entries (repeat n times):
  - nameLen: uint16
  - name: `nameLen` bytes
  - enc: uint8 (same semantics as `.vopl` enc)
  - plen: uint32
  - payload: `plen` bytes (raw stream; may itself be zlib if enc bit7 set)

No footer, no checksums.


## Go examples

Dense v3 writer for a 16³ grid of palette indices (already in Morton order):

```go
package main
import (
	"bytes"
	"encoding/binary"
)
func writeVOPLv3Dense(grid []uint8) []byte {
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

Dense v3 writer (Morton order indices computed on the fly; no zlib):

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
function buildVOPLv3Dense(grid){
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
- bpp must be ≤8; used values: 5 (v2), 6 (v3).
- Decoder expects exactly `plen` bytes of payload after header.
- Decoder ignores w/h/d from the header and reconstructs 16×16×16.


## Reimplementing from this spec
- Implement LSB-first bit I/O.
- Implement 3D Morton order over 16×16×16.
- Implement dense/sparse/rle encoders/decoders with exact field widths.
- Optionally apply zlib to the raw encoding stream if it reduces size; set `enc|=0x80`.
- Write/read v3 headers exactly as specified; follow v2/v1 layouts above.
