package vopl

const (
	Height = 16
	Width  = 16
	Depth  = 16
)

// VoxelGrid[y][x][z]
type VoxelGrid [Height][Width][Depth]uint8
