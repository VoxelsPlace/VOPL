package test

import (
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

func TestUtils_VPI18ToVOPLFile(t *testing.T) {
	// Build a VPI18 file from a small grid
	grid := makeSmallGrid()
	vpi := vopl.VPI18EncodeGrid(grid)
	inPath := "output/in.vpi"
	outPath := "output/from_vpi.vopl"
	_ = os.MkdirAll(filepath.Dir(outPath), 0o755)
	if err := os.WriteFile(inPath, vpi, 0o644); err != nil {
		t.Fatalf("write vpi: %v", err)
	}
	if err := utils.RunVPI18ToVOPLFile(inPath, outPath); err != nil {
		t.Fatalf("RunVPI18ToVOPLFile: %v", err)
	}
	got, err := vopl.LoadVoplGrid(outPath)
	if err != nil {
		t.Fatalf("load vopl: %v", err)
	}
	if *got != *grid {
		t.Fatalf("grid mismatch after VPI18->VOPL conversion")
	}
}

func TestUtils_UpdateVOPL_WithVPI18_UtilsSuite(t *testing.T) {
	// Start from empty VOPL file and apply VPI18 updates
	var base vopl.VoxelGrid
	basePath := "output/base_utils.vopl"
	outPath := "output/updated_utils.vopl"
	_ = os.MkdirAll(filepath.Dir(basePath), 0o755)
	if err := vopl.SaveVoplGrid(&base, basePath); err != nil {
		t.Fatalf("save base vopl: %v", err)
	}
	updates := vopl.VPI18EncodeGrid(makeSmallGrid())
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
