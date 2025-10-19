package vopl

import (
    "bytes"
    "compress/zlib"
    "io"
)

const (
    encDense   = 0
    encSparse  = 1
    encRLE     = 2
    encSparse2 = 3 // occupancy bitmap + nonzero values
    encRLE0    = 4 // zero-run and literal blocks with varint lengths
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

func encodeRLE0(grid *VoxelGrid, bpp uint8) []byte {
    // Block format:
    //   flag:1 (1=zero-run, 0=literal block)
    //   if zero-run: len:varint
    //   if literal: len:varint, then len values (bpp bits each)
    stream := flatten(grid)
    if len(stream) == 0 {
        return nil
    }
    // header assembled in bytes; values for literal blocks are bit-packed
    header := make([]byte, 0, 256)
    bw := newBitWriter()
    i := 0
    for i < len(stream) {
        if stream[i] == 0 {
            // zero run
            j := i + 1
            for j < len(stream) && stream[j] == 0 {
                j++
            }
            run := uint32(j - i)
            header = append(header, 0x80) // flag=1 in MSB form for clarity (we’ll store as 1 byte flag)
            header = writeUVarint(header, run)
            i = j
            continue
        }
        // literal block
        j := i + 1
        for j < len(stream) && stream[j] != 0 {
            j++
        }
        ln := uint32(j - i)
        header = append(header, 0x00)
        header = writeUVarint(header, ln)
        for k := i; k < j; k++ {
            bw.writeBits(uint64(stream[k]), bpp)
        }
        i = j
    }
    values := bw.bytes()
    out := make([]byte, 0, len(header)+len(values))
    out = append(out, header...)
    out = append(out, values...)
    return out
}

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
        {encoding: encRLE, payload: encodeRLE(grid, bpp)},
        {encoding: encSparse2, payload: encodeSparse2(grid, bpp)},
        {encoding: encRLE0, payload: encodeRLE0(grid, bpp)},
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
