package vopl

import "math/bits"

func expand3(v uint32) uint32 {
	v = (v | (v << 16)) & 0x030000FF
	v = (v | (v << 8)) & 0x0300F00F
	v = (v | (v << 4)) & 0x030C30C3
	v = (v | (v << 2)) & 0x09249249
	return v
}

func morton3D(x, y, z uint32) uint32 {
	return expand3(x) | (expand3(y) << 1) | (expand3(z) << 2)
}

var mortonOrder []int

// linearToMortonRank maps linear index (x + y*Width + z*Width*Height) -> Morton rank (0..4095)
var linearToMortonRank [Width * Height * Depth]uint16

func buildMortonOrder() []int {
	total := Width * Height * Depth
	type kv struct {
		key uint32
		i   int
	}
	idx := make([]kv, 0, total)
	i := 0
	for y := range Height {
		for z := range Depth {
			for x := range Width {
				idx = append(idx, kv{morton3D(uint32(x), uint32(y), uint32(z)), i})
				i++
			}
		}
	}
	// insertion sort (estável e suficiente pro tamanho pequeno)
	for a := 1; a < len(idx); a++ {
		k := idx[a]
		b := a - 1
		for b >= 0 && idx[b].key > k.key {
			idx[b+1] = idx[b]
			b--
		}
		idx[b+1] = k
	}
	order := make([]int, total)
	for i := range idx {
		order[i] = idx[i].i
	}
	return order
}

func init() { mortonOrder = buildMortonOrder() }

// Build inverse mapping after mortonOrder is initialized
func init() {
	for rank, lin := range mortonOrder {
		linearToMortonRank[lin] = uint16(rank)
	}
}

func flatten(grid *VoxelGrid) []uint8 {
	stream := make([]uint8, 0, Width*Height*Depth)
	lin := make([]uint8, Width*Height*Depth)
	p := 0
	for y := range Height {
		for z := range Depth {
			for x := range Width {
				lin[p] = grid[y][x][z]
				p++
			}
		}
	}
	for _, i := range mortonOrder {
		stream = append(stream, lin[i])
	}
	return stream
}

func applyOrder(grid *VoxelGrid, lin []uint8) {
	back := make([]uint8, len(lin))
	for i, src := range mortonOrder {
		back[src] = lin[i]
	}
	p := 0
	for y := range Height {
		for z := range Depth {
			for x := range Width {
				grid[y][x][z] = back[p]
				p++
			}
		}
	}
}

// MortonRankFromXYZ returns the Morton rank (0..W*H*D-1) for the given (x,y,z).
func MortonRankFromXYZ(x, y, z int) uint16 {
	lin := x + y*Width + z*Width*Height
	return linearToMortonRank[lin]
}

// XYZFromMortonRank returns the (x,y,z) for a given Morton rank.
func XYZFromMortonRank(rank uint16) (x, y, z int) {
	lin := mortonOrder[int(rank)]
	x = lin % Width
	y = (lin / Width) % Height
	z = lin / (Width * Height)
	return
}

func Morton3D64(x, y, z uint32) uint64 {
	return part1By2(uint64(x)) |
		(part1By2(uint64(y)) << 1) |
		(part1By2(uint64(z)) << 2)
}

func MortonDecode3D64(index uint64) (x, y, z uint32) {
	x = uint32(compact1By2(index))
	y = uint32(compact1By2(index >> 1))
	z = uint32(compact1By2(index >> 2))
	return
}

func part1By2(x uint64) uint64 {
	x &= 0x1fffff
	x = (x | (x << 32)) & 0x1f00000000ffff
	x = (x | (x << 16)) & 0x1f0000ff0000ff
	x = (x | (x << 8)) & 0x100f00f00f00f00f
	x = (x | (x << 4)) & 0x10c30c30c30c30c3
	x = (x | (x << 2)) & 0x1249249249249249
	return x
}

func compact1By2(x uint64) uint64 {
	x &= 0x1249249249249249
	x = (x ^ (x >> 2)) & 0x10c30c30c30c30c3
	x = (x ^ (x >> 4)) & 0x100f00f00f00f00f
	x = (x ^ (x >> 8)) & 0x1f0000ff0000ff
	x = (x ^ (x >> 16)) & 0x1f00000000ffff
	x = (x ^ (x >> 32)) & 0x1fffff
	return x
}

func Morton3DMaxBits(w, d, h int) int {
	max := w - 1
	if d-1 > max {
		max = d - 1
	}
	if h-1 > max {
		max = h - 1
	}
	return bits.Len(uint(max)) * 3
}
