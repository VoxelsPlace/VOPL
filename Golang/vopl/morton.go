package vopl

func expand3(v uint32) uint32 {
	v = (v | (v << 16)) & 0xFF0000FF
	v = (v | (v << 8)) & 0x0F00F00F
	v = (v | (v << 4)) & 0xC30C30C3
	v = (v | (v << 2)) & 0x49249249
	return v
}
func morton3D(x, y, z uint32) uint32 {
	return expand3(x) | (expand3(y) << 1) | (expand3(z) << 2)
}

var mortonOrder []int

func buildMortonOrder() []int {
	total := Width * Height * Depth
	type kv struct {
		key uint32
		i   int
	}
	idx := make([]kv, 0, total)
	i := 0
	for y := 0; y < Height; y++ {
		for z := 0; z < Depth; z++ {
			for x := 0; x < Width; x++ {
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

func flatten(grid *VoxelGrid) []uint8 {
	stream := make([]uint8, 0, Width*Height*Depth)
	lin := make([]uint8, Width*Height*Depth)
	p := 0
	for y := 0; y < Height; y++ {
		for z := 0; z < Depth; z++ {
			for x := 0; x < Width; x++ {
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
	for y := 0; y < Height; y++ {
		for z := 0; z < Depth; z++ {
			for x := 0; x < Width; x++ {
				grid[y][x][z] = back[p]
				p++
			}
		}
	}
}
