package vopl

import (
	"fmt"
	"io"
)

// VPI18 encodes voxel updates as 18-bit entries: 12-bit linear index (x + y*16 + z*256), 6-bit color (0..63)
// The bitstream is continuous, no padding. Color==0 means delete/clear the voxel at the given index.

// VPI18Entry represents a single VPI18 (index,color) pair. In diff streams, Color==0 means delete/clear.
type VPI18Entry struct {
	Index uint16 // 0..4095
	Color uint8  // 0..63 (0 = delete)
}

// VPI18EncodeGrid encodes the given grid into a VPI18 bitstream.
func VPI18EncodeGrid(grid *VoxelGrid) []byte {
	bw := newBitWriter()
	for y := 0; y < Height; y++ {
		for z := 0; z < Depth; z++ {
			for x := 0; x < Width; x++ {
				c := grid[y][x][z]
				if c == 0 {
					continue // encode only active updates; from empty this is a full build diff
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

// VPI18EncodeEntries encodes the provided entries exactly as given (including Color==0) into a VPI18 bitstream.
// This is useful for building sparse diffs that include deletions.
func VPI18EncodeEntries(entries []VPI18Entry) []byte {
	if len(entries) == 0 {
		return nil
	}
	bw := newBitWriter()
	for _, e := range entries {
		idx := uint16(e.Index) & 0x0FFF
		col := uint8(e.Color) & 0x3F
		entry := (uint64(idx) << 6) | uint64(col)
		bw.writeBits(entry, 18)
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

// VPI18ApplyToGrid applies a VPI18 stream of updates to an existing grid.
// Color==0 deletes/clears the voxel at Index; non-zero sets the color.
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
		x := int(idx % 16)
		y := int((idx / 16) % 16)
		z := int(idx / 256)
		if col == 0 {
			grid[y][x][z] = 0
		} else {
			grid[y][x][z] = col
		}
	}
}
