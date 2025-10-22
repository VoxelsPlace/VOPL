package test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/voxelsplace/vopl/go/api"
	"github.com/voxelsplace/vopl/go/utils"
	"github.com/voxelsplace/vopl/go/vopl"
)

func makeTestGrid() *vopl.VoxelGrid {
	var g vopl.VoxelGrid
	for y := 0; y < 2; y++ {
		for z := 0; z < 2; z++ {
			for x := 0; x < 4; x++ {
				c := uint8(1 + ((x + z + y) % 6)) // 1..6
				g[y][x][z] = c
			}
		}
	}
	return &g
}

func TestVPI18_Roundtrip(t *testing.T) {
	grid := makeTestGrid()
	vpi := vopl.VPI18EncodeGrid(grid)
	dec, err := vopl.VPI18DecodeToGrid(vpi)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	vpi2 := vopl.VPI18EncodeGrid(dec)
	if !bytes.Equal(vpi, vpi2) {
		t.Fatalf("VPI18 encode->decode->encode not stable")
	}
}

func TestAPI_VPI18ToVOPL_ThenBack(t *testing.T) {
	grid := makeTestGrid()
	vpi := vopl.VPI18EncodeGrid(grid)

	voplBytes, err := api.VPI18ToVOPLBytes(vpi)
	if err != nil {
		t.Fatalf("VPI18ToVOPLBytes failed: %v", err)
	}
	gotGrid, err := vopl.LoadVoplGridFromBytes(voplBytes)
	if err != nil {
		t.Fatalf("LoadVoplGridFromBytes failed: %v", err)
	}
	vpiBack, err := api.VOPLToVPI18(voplBytes)
	if err != nil {
		t.Fatalf("VOPLToVPI18 failed: %v", err)
	}
	got2, err := vopl.VPI18DecodeToGrid(vpiBack)
	if err != nil {
		t.Fatalf("VPI18DecodeToGrid: %v", err)
	}
	if *got2 != *gotGrid {
		t.Fatalf("grids differ after roundtrip")
	}
}

func TestUtils_UpdateVOPL_WithVPI18(t *testing.T) {
	// start from empty, apply updates using VPI18
	var base vopl.VoxelGrid
	basePath := "output/base.vopl"
	_ = os.MkdirAll(filepath.Dir(basePath), 0o755)
	if err := vopl.SaveVoplGrid(&base, basePath); err != nil {
		t.Fatalf("save base: %v", err)
	}
	vpi := vopl.VPI18EncodeGrid(makeTestGrid())
	outPath := "output/updated.vopl"
	if err := utils.RunUpdateVOPL(vpi, basePath, outPath); err != nil {
		t.Fatalf("update failed: %v", err)
	}
	got, err := vopl.LoadVoplGrid(outPath)
	if err != nil {
		t.Fatalf("load updated: %v", err)
	}
	want := makeTestGrid()
	if *got != *want {
		t.Fatalf("grid mismatch after update")
	}
}

func TestVPI18_DiffDeletionsAndAdds(t *testing.T) {
	// Start with a small grid
	grid := makeTestGrid()
	// Choose one existing voxel to delete and one empty to add
	// existing: (x=0,y=0,z=0) -> index 0
	delIdx := uint16(0)
	// add: (x=5,y=0,z=0) -> index 5
	addIdx := uint16(5)
	// Sanity: ensure initial values
	if grid[0][0][0] == 0 {
		t.Fatalf("expected initial voxel at (0,0,0) to be non-zero")
	}
	// ensure target add position is empty prior to diff
	grid[0][5][0] = 0
	// Build diff entries: delete at 0, add at 5 with color 9
	entries := []vopl.VPI18Entry{
		{Index: delIdx, Color: 0}, // delete
		{Index: addIdx, Color: 9}, // add
	}
	diff := vopl.VPI18EncodeEntries(entries)
	if err := vopl.VPI18ApplyToGrid(grid, diff); err != nil {
		t.Fatalf("apply diff failed: %v", err)
	}
	if got := grid[0][0][0]; got != 0 {
		t.Fatalf("expected deletion at (0,0,0), got %d", got)
	}
	x := int(addIdx % 16)
	y := int((addIdx / 16) % 16)
	z := int(addIdx / 256)
	if got := grid[y][x][z]; got != 9 {
		t.Fatalf("expected add at (%d,%d,%d)=9, got %d", x, y, z, got)
	}
}
