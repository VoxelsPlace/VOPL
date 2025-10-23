package vopl

import (
	"fmt"
	"io"
)

// VPI18 encodes voxel updates as 18-bit entries: 12-bit Morton rank (0..4095) dentro do chunk 16^3, 6-bit color (0..63).
// A ordem de varredura do stream não importa, mas o índice é o rank em Z-order (Morton) e NÃO linear.
// O bitstream é contínuo, sem padding. Color==0 significa deletar/limpar o voxel naquele rank.

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
				// Morton rank dentro do 16^3
				idx := uint16(MortonRankFromXYZ(x, y, z))
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

// VPI18DecodeToEntries decodes a VPI18 bitstream into a slice of entries (Index, Color).
// Color==0 indicates deletion when applied.
func VPI18DecodeToEntries(data []byte) ([]VPI18Entry, error) {
	br := newBitReader(data)
	out := make([]VPI18Entry, 0, len(data))
	for {
		bits, err := br.readBits(18)
		if err != nil {
			if err == io.ErrUnexpectedEOF {
				return out, nil
			}
			return nil, err
		}
		idx := uint16(bits >> 6)
		col := uint8(bits & 0x3F)
		if idx >= 4096 {
			return nil, fmt.Errorf("VPI18 index out of range: %d", idx)
		}
		out = append(out, VPI18Entry{Index: idx, Color: col})
	}
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
		idx := uint16(bits >> 6) // Morton rank
		col := uint8(bits & 0x3F)
		if idx >= 4096 {
			return fmt.Errorf("VPI18 index out of range: %d", idx)
		}
		// Map Morton rank -> linear index -> (x,y,z)
		lin := mortonOrder[int(idx)]
		x := lin % Width
		y := (lin / Width) % Height
		z := lin / (Width * Height)
		if col == 0 {
			grid[y][x][z] = 0
		} else {
			grid[y][x][z] = col
		}
	}
}
