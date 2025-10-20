package vopl

import (
	"bytes"
	"compress/zlib"
	"io"
)

const (
	encDense  = 0
	encSparse = 1
	// encRLE = 2, removed
	encSparse2 = 3 // occupancy bitmap + nonzero values
	// encRLE0 = 4, removed
)

type encoded struct {
	encoding int
	payload  []byte
}

func encodeDense(grid *VoxelGrid, bpp uint8) []byte {
	bw := newBitWriter()
	for _, c := range flatten(grid) {
		bw.writeBits(uint64(c), bpp)
	}
	return bw.bytes()
}

func encodeSparse(grid *VoxelGrid, bpp uint8) []byte {
	bw := newBitWriter()
	stream := flatten(grid)
	count := 0
	for _, c := range stream {
		if c != 0 {
			count++
		}
	}
	bw.writeBits(uint64(count), 16)
	if count == 0 {
		return bw.bytes()
	}
	for i, c := range stream {
		if c == 0 {
			continue
		}
		bw.writeBits(uint64(i), 8)
		bw.writeBits(uint64(c), bpp)
	}
	return bw.bytes()
}

// encodeRLE removed

func encodeSparse2(grid *VoxelGrid, bpp uint8) []byte {
	stream := flatten(grid)
	// 4096-bit occupancy bitmap -> 512 bytes
	bitmap := make([]byte, 512)
	nonzeros := make([]uint8, 0, len(stream))
	for i, v := range stream {
		if v != 0 {
			bitmap[i>>3] |= 1 << (uint(i) & 7)
			nonzeros = append(nonzeros, v)
		}
	}
	if len(nonzeros) == 0 {
		// only bitmap, no values
		return append([]byte{}, bitmap...)
	}
	bw := newBitWriter()
	for _, c := range nonzeros {
		bw.writeBits(uint64(c), bpp)
	}
	values := bw.bytes()
	out := make([]byte, 0, 512+len(values))
	out = append(out, bitmap...)
	out = append(out, values...)
	return out
}

// encodeRLE0 removed

func zlibCompress(b []byte) []byte {
	var buf bytes.Buffer
	zw, _ := zlib.NewWriterLevel(&buf, zlib.BestCompression)
	_, _ = zw.Write(b)
	_ = zw.Close()
	return buf.Bytes()
}

func zlibDecompress(b []byte) ([]byte, error) {
	zr, err := zlib.NewReader(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	return io.ReadAll(zr)
}

func bestEncoding(grid *VoxelGrid, bpp uint8) encoded {
	candidates := []encoded{
		{encoding: encDense, payload: encodeDense(grid, bpp)},
		{encoding: encSparse, payload: encodeSparse(grid, bpp)},
		{encoding: encSparse2, payload: encodeSparse2(grid, bpp)},
	}
	best := encoded{encoding: candidates[0].encoding, payload: candidates[0].payload}
	for _, c := range candidates[1:] {
		if len(c.payload) < len(best.payload) {
			best = c
		}
	}
	// also compare compressed versions of each
	for _, c := range candidates {
		zb := zlibCompress(c.payload)
		if len(zb) < len(best.payload) {
			best = encoded{encoding: c.encoding | 0x80, payload: zb}
		}
	}
	return best
}
