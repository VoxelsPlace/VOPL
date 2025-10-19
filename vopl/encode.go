package vopl

import (
	"bytes"
	"compress/zlib"
	"io"
)

const (
	encDense  = 0
	encSparse = 1
	encRLE    = 2
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

func encodeRLE(grid *VoxelGrid, bpp uint8) []byte {
	bw := newBitWriter()
	stream := flatten(grid)
	if len(stream) == 0 {
		return bw.bytes()
	}
	cur := stream[0]
	run := 1
	flush := func() {
		bw.writeBits(uint64(run-1), 8)
		bw.writeBits(uint64(cur), bpp)
	}
	for i := 1; i < len(stream); i++ {
		if stream[i] == cur && run < 256 {
			run++
			continue
		}
		flush()
		cur = stream[i]
		run = 1
	}
	flush()
	return bw.bytes()
}

func zlibCompress(b []byte) []byte {
	var buf bytes.Buffer
	zw := zlib.NewWriter(&buf)
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
	d := encodeDense(grid, bpp)
	s := encodeSparse(grid, bpp)
	r := encodeRLE(grid, bpp)
	best := encoded{encoding: encDense, payload: d}
	if len(s) < len(best.payload) {
		best = encoded{encoding: encSparse, payload: s}
	}
	if len(r) < len(best.payload) {
		best = encoded{encoding: encRLE, payload: r}
	}
	zb := zlibCompress(best.payload)
	if len(zb) < len(best.payload) {
		return encoded{encoding: best.encoding | 0x80, payload: zb}
	}
	return best
}
