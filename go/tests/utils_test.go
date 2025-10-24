package test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/qmuntal/gltf"
	"github.com/voxelsplace/vopl/go/utils"
	"github.com/voxelsplace/vopl/go/vopl"
)

func makeSmallGrid() *vopl.VoxelGrid {
	var g vopl.VoxelGrid
	for y := range 2 {
		for z := range 2 {
			for x := range 3 {
				g[y][x][z] = uint8(1 + (x+y+z)%6)
			}
		}
	}
	return &g
}

// helper to build updates JSON for a single chunk id from a grid
func buildUpdatesJSONFromGrid(chunkID string, grid *vopl.VoxelGrid) []byte {
	// { chunkID: { idx: color } }
	indices := map[string]int{}
	for y := range vopl.Height {
		for z := range vopl.Depth {
			for x := range vopl.Width {
				c := int(grid[y][x][z])
				if c == 0 {
					continue
				}
				idx := x + y*vopl.Width + z*vopl.Width*vopl.Height
				indices[itoa(idx)] = c
			}
		}
	}
	outer := map[string]map[string]int{chunkID: indices}
	b, _ := json.Marshal(outer)
	return b
}

// fast itoa for small ints
func itoa(n int) string { return fmtInt(n) }

func fmtInt(n int) string {
	// simple base-10 conversion
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + (n % 10))
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func TestUtils_JSONToVOPLFile(t *testing.T) {
	grid := makeSmallGrid()
	updates := buildUpdatesJSONFromGrid("0", grid)
	inPath := "output/in.json"
	outPath := "output/from_json.vopl"
	_ = os.MkdirAll(filepath.Dir(outPath), 0o755)
	if err := os.WriteFile(inPath, updates, 0o644); err != nil {
		t.Fatalf("write json: %v", err)
	}
	if err := utils.RunJSONToVOPLFile(inPath, outPath); err != nil {
		t.Fatalf("RunJSONToVOPLFile: %v", err)
	}
	got, err := vopl.LoadVoplGrid(outPath)
	if err != nil {
		t.Fatalf("load vopl: %v", err)
	}
	if *got != *grid {
		t.Fatalf("grid mismatch after JSON->VOPL conversion")
	}
}

func TestUtils_UpdateVOPL_WithJSON_UtilsSuite(t *testing.T) {
	// Start from empty VOPL file and apply JSON updates
	var base vopl.VoxelGrid
	basePath := "output/base_utils.vopl"
	outPath := "output/updated_utils.vopl"
	_ = os.MkdirAll(filepath.Dir(basePath), 0o755)
	if err := vopl.SaveVoplGrid(&base, basePath); err != nil {
		t.Fatalf("save base vopl: %v", err)
	}
	updates := buildUpdatesJSONFromGrid("0", makeSmallGrid())
	if err := utils.RunUpdateVOPL(updates, basePath, outPath); err != nil {
		t.Fatalf("RunUpdateVOPL: %v", err)
	}
	got, err := vopl.LoadVoplGrid(outPath)
	if err != nil {
		t.Fatalf("load updated vopl: %v", err)
	}
	want := makeSmallGrid()
	if *got != *want {
		t.Fatalf("grid mismatch after update")
	}
}

func TestUtils_UpdateVOPL_ApplyGivenIndicesToEmpty(t *testing.T) {
	// Create an empty VOPL file and apply the provided update JSON
	var base vopl.VoxelGrid
	basePath := "output/base_specific.vopl"
	outPath := "output/updated_specific.vopl"
	_ = os.MkdirAll(filepath.Dir(basePath), 0o755)
	if err := vopl.SaveVoplGrid(&base, basePath); err != nil {
		t.Fatalf("save base vopl: %v", err)
	}

	updates := []byte(`{
		"26304528517632": {
			"0": 1,
			"15": 19,
			"3840": 13,
			"3855": 7
		}
	}`)

	if err := utils.RunUpdateVOPL(updates, basePath, outPath); err != nil {
		t.Fatalf("RunUpdateVOPL: %v", err)
	}

	got, err := vopl.LoadVoplGrid(outPath)
	if err != nil {
		t.Fatalf("load updated vopl: %v", err)
	}

	// Expected non-zero voxels derived from linear indices using idx = x + y*W + z*W*H
	// Indices: 0 -> (0,0,0)=1, 15 -> (15,0,0)=19, 3840 -> (0,0,15)=13, 3855 -> (15,0,15)=7
	checks := []struct {
		x, y, z int
		color   uint8
	}{
		{0, 0, 0, 1},
		{15, 0, 0, 19},
		{0, 0, 15, 13},
		{15, 0, 15, 7},
	}

	for _, c := range checks {
		if got[c.y][c.x][c.z] != c.color {
			t.Fatalf("voxel (%d,%d,%d) = %d, want %d", c.x, c.y, c.z, got[c.y][c.x][c.z], c.color)
		}
	}

	// Ensure all other voxels remain zero
	for y := 0; y < vopl.Height; y++ {
		for x := 0; x < vopl.Width; x++ {
			for z := 0; z < vopl.Depth; z++ {
				isChecked := (x == 0 && y == 0 && z == 0) ||
					(x == 15 && y == 0 && z == 0) ||
					(x == 0 && y == 0 && z == 15) ||
					(x == 15 && y == 0 && z == 15)
				if isChecked {
					continue
				}
				if got[y][x][z] != 0 {
					t.Fatalf("unexpected non-zero voxel at (%d,%d,%d): %d", x, y, z, got[y][x][z])
				}
			}
		}
	}
}

func TestUtils_UpdateVOPL_ToGLB_ApplyGivenIndices(t *testing.T) {
	// Start from empty VOPL, apply updates, convert to GLB, and validate result
	var base vopl.VoxelGrid
	basePath := "output/base_glb.vopl"
	updatedPath := "output/updated_glb.vopl"
	glbPath := "output/updated_glb.glb"
	_ = os.MkdirAll(filepath.Dir(basePath), 0o755)
	if err := vopl.SaveVoplGrid(&base, basePath); err != nil {
		t.Fatalf("save base vopl: %v", err)
	}

	updates := []byte(`{
		"26304528517632": {
			"0": 1,
			"15": 19,
			"3840": 13,
			"3855": 7
		}
	}`)

	if err := utils.RunUpdateVOPL(updates, basePath, updatedPath); err != nil {
		t.Fatalf("RunUpdateVOPL: %v", err)
	}
	if err := utils.RunVOPL2GLB(updatedPath, glbPath); err != nil {
		t.Fatalf("RunVOPL2GLB: %v", err)
	}
	// Ensure file exists and non-empty
	if fi, err := os.Stat(glbPath); err != nil || fi.Size() == 0 {
		t.Fatalf("glb not created or empty")
	}

	// Load GLB and validate mesh/accessors
	doc, err := gltf.Open(glbPath)
	if err != nil {
		t.Fatalf("open glb: %v", err)
	}
	if len(doc.Meshes) != 1 {
		t.Fatalf("expected 1 mesh, got %d", len(doc.Meshes))
	}
	if len(doc.Meshes[0].Primitives) != 1 {
		t.Fatalf("expected 1 primitive, got %d", len(doc.Meshes[0].Primitives))
	}
	prim := doc.Meshes[0].Primitives[0]
	posAccIdx, ok := prim.Attributes[gltf.POSITION]
	if !ok {
		t.Fatalf("POSITION accessor missing")
	}
	normAccIdx, ok := prim.Attributes[gltf.NORMAL]
	if !ok {
		t.Fatalf("NORMAL accessor missing")
	}
	colAccIdx, ok := prim.Attributes[gltf.COLOR_0]
	if !ok {
		t.Fatalf("COLOR_0 accessor missing")
	}
	if prim.Indices == nil {
		t.Fatalf("Indices accessor missing")
	}
	idxAccIdx := int(*prim.Indices)

	// Expect 4 isolated voxels -> 24 quads -> 96 verts and 144 indices (48 triangles)
	posAcc := doc.Accessors[posAccIdx]
	idxAcc := doc.Accessors[idxAccIdx]
	normAcc := doc.Accessors[normAccIdx]
	colorAcc := doc.Accessors[colAccIdx]

	if posAcc.Count != 96 {
		t.Fatalf("POSITION count = %d, want 96", posAcc.Count)
	}
	if normAcc.Count != posAcc.Count {
		t.Fatalf("NORMAL count = %d, want %d", normAcc.Count, posAcc.Count)
	}
	if colorAcc.Count != posAcc.Count {
		t.Fatalf("COLOR_0 count = %d, want %d", colorAcc.Count, posAcc.Count)
	}
	if idxAcc.Count != 144 {
		t.Fatalf("Indices count = %d, want 144", idxAcc.Count)
	}
}
