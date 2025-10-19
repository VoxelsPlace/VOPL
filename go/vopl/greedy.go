package vopl

type dirSpec struct {
	normal [3]float32
	u, v   int
	du, dv [3]int
}

var directions = []dirSpec{
	{[3]float32{1, 0, 0}, 1, 2, [3]int{0, 1, 0}, [3]int{0, 0, 1}},
	{[3]float32{-1, 0, 0}, 1, 2, [3]int{0, 1, 0}, [3]int{0, 0, 1}},
	{[3]float32{0, 1, 0}, 0, 2, [3]int{1, 0, 0}, [3]int{0, 0, 1}},
	{[3]float32{0, -1, 0}, 0, 2, [3]int{1, 0, 0}, [3]int{0, 0, 1}},
	{[3]float32{0, 0, 1}, 0, 1, [3]int{1, 0, 0}, [3]int{0, 1, 0}},
	{[3]float32{0, 0, -1}, 0, 1, [3]int{1, 0, 0}, [3]int{0, 1, 0}},
}

func getVoxel(grid *VoxelGrid, x, y, z int) uint8 {
	if x < 0 || x >= Width || y < 0 || y >= Height || z < 0 || z >= Depth {
		return 0
	}
	return grid[y][x][z]
}

func addQuad(mesh *Mesh, dir dirSpec, start [3]int, w, h int, color uint8, perp int) {
	base := [3]float32{}
	base[perp] = float32(start[0])
	if dir.normal[perp] > 0 {
		base[perp] += 1
	}
	base[dir.u] = float32(start[1])
	base[dir.v] = float32(start[2])

	verts := [4]Vertex{
		{Position: base, Color: color},
		{Position: [3]float32{base[0] + float32(dir.du[0]*h), base[1] + float32(dir.du[1]*h), base[2] + float32(dir.du[2]*h)}, Color: color},
		{Position: [3]float32{base[0] + float32(dir.du[0]*h) + float32(dir.dv[0]*w), base[1] + float32(dir.du[1]*h) + float32(dir.dv[1]*w), base[2] + float32(dir.du[2]*h) + float32(dir.dv[2]*w)}, Color: color},
		{Position: [3]float32{base[0] + float32(dir.dv[0]*w), base[1] + float32(dir.dv[1]*w), base[2] + float32(dir.dv[2]*w)}, Color: color},
	}

	swap := (dir.normal[perp] < 0) != (perp == 1)
	if swap {
		verts[1], verts[3] = verts[3], verts[1]
	}

	baseIdx := uint32(len(mesh.Vertices))
	mesh.Vertices = append(mesh.Vertices, verts[:]...)
	mesh.Indices = append(mesh.Indices, baseIdx, baseIdx+1, baseIdx+2, baseIdx, baseIdx+2, baseIdx+3)
}

func GenerateMesh(grid *VoxelGrid) *Mesh {
	mesh := &Mesh{}
	dims := [3]int{Width, Height, Depth}

	for _, dir := range directions {
		perp := 3 - dir.u - dir.v

		for p := 0; p < dims[perp]; p++ {
			mask := make([][]uint8, dims[dir.u])
			visited := make([][]bool, dims[dir.u])
			for i := range mask {
				mask[i] = make([]uint8, dims[dir.v])
				visited[i] = make([]bool, dims[dir.v])
			}

			for u := 0; u < dims[dir.u]; u++ {
				for v := 0; v < dims[dir.v]; v++ {
					pos := [3]int{}
					pos[dir.u] = u
					pos[dir.v] = v
					pos[perp] = p

					voxel := getVoxel(grid, pos[0], pos[1], pos[2])
					if voxel == 0 {
						continue
					}

					adj := pos
					if dir.normal[perp] < 0 {
						adj[perp] = p - 1
					} else {
						adj[perp] = p + 1
					}

					if adj[perp] < 0 || adj[perp] >= dims[perp] || getVoxel(grid, adj[0], adj[1], adj[2]) == 0 {
						mask[u][v] = voxel
					}
				}
			}

			for u := 0; u < dims[dir.u]; u++ {
				for v := 0; v < dims[dir.v]; {
					if mask[u][v] == 0 || visited[u][v] {
						v++
						continue
					}
					color := mask[u][v]
					width := 1
					for w := v + 1; w < dims[dir.v] && mask[u][w] == color && !visited[u][w]; w++ {
						width++
					}
					height := 1
					stop := false
					for h := u + 1; h < dims[dir.u] && !stop; h++ {
						for w := v; w < v+width; w++ {
							if mask[h][w] != color || visited[h][w] {
								stop = true
								break
							}
						}
						if !stop {
							height++
						}
					}
					for hu := u; hu < u+height; hu++ {
						for hv := v; hv < v+width; hv++ {
							visited[hu][hv] = true
						}
					}
					addQuad(mesh, dir, [3]int{p, u, v}, width, height, color, perp)
					v += width
				}
			}
		}
	}
	return mesh
}
