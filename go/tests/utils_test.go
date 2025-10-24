package test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

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
