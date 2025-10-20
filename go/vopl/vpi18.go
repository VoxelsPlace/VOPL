package vopl

import (
	"fmt"
	"io"
)

// VPI18 encodes non-zero voxels as 18-bit entries: 12-bit linear index (x + y*16 + z*256), 6-bit color (1..63)
// The bitstream is continuous, no padding. Voxels with color 0 are omitted.

// VPI18EncodeGrid encodes the given grid into a VPI18 bitstream.
func VPI18EncodeGrid(grid *VoxelGrid) []byte {
	bw := newBitWriter()
	for y := 0; y < Height; y++ {
		for z := 0; z < Depth; z++ {
			for x := 0; x < Width; x++ {
				c := grid[y][x][z]
				if c == 0 {
					continue
				}
				// linear index inside 16^3 chunk: x + y*16 + z*256
				idx := uint16(x + y*16 + z*256)
				// pack: upper 12 bits index, lower 6 bits color
				entry := (uint64(idx) << 6) | (uint64(c) & 0x3F)
				bw.writeBits(entry, 18)
			}
		}
	}
	return bw.bytes()
}

// VPI18DecodeToGrid decodes a full VPI18 stream into a new grid (starting from all zeros).
func VPI18DecodeToGrid(data []byte) (*VoxelGrid, error) {
	var grid VoxelGrid
	if err := VPI18ApplyToGrid(&grid, data); err != nil {
		return nil, err
	}
	return &grid, nil
}

// VPI18ApplyToGrid applies a VPI18 stream as updates to an existing grid.
// Entries with color==0 are ignored by definition and thus never appear in the stream.
func VPI18ApplyToGrid(grid *VoxelGrid, data []byte) error {
	br := newBitReader(data)
	for {
		bits, err := br.readBits(18)
		if err != nil {
			if err == io.ErrUnexpectedEOF {
				// partial trailing bits are treated as end of stream
				return nil
			}
			return err
		}
		idx := uint16(bits >> 6)
		col := uint8(bits & 0x3F)
		if idx >= 4096 {
			return fmt.Errorf("VPI18 index out of range: %d", idx)
		}
		if col == 0 {
			// spec: color 0 is not transmitted; ignore if encountered
			continue
		}
		x := int(idx % 16)
		y := int((idx / 16) % 16)
		z := int(idx / 256)
		grid[y][x][z] = col
	}
}
