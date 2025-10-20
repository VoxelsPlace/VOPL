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
