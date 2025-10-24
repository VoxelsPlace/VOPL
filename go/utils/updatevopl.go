package utils

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/voxelsplace/vopl/go/vopl"
)

// updatesJSON is { "<chunkId>": { "<index>": <color>, ... }, ... }
type updatesJSON map[string]map[string]int

// indexToXYZ converts linear index (x + y*W + z*W*H) into coordinates.
func indexToXYZ(index int) (x, y, z int) {
	w, h := vopl.Width, vopl.Height
	wh := w * h
	z = index / wh
	rem := index - z*wh
	y = rem / w
	x = rem - y*w
	return
}

// RunUpdateVOPL applies a JSON updates blob to an existing .vopl file and writes a regenerated .vopl.
// The JSON format is: { "<chunkId>": { "<index>": <color>, ... } }
// Only a single chunk is supported here; extra chunks are ignored.
func RunUpdateVOPL(jsonUpdates []byte, inputPath, outputPath string) error {
	grid, err := vopl.LoadVoplGrid(inputPath)
	if err != nil {
		return fmt.Errorf("failed to load input VOPL: %w", err)
	}
	if err := applyJSONToGrid(grid, jsonUpdates); err != nil {
		return err
	}
	if err := vopl.SaveVoplGrid(grid, outputPath); err != nil {
		return fmt.Errorf("failed to save VOPL: %w", err)
	}
	if fi, err := os.Stat(outputPath); err == nil {
		fmt.Printf(".vopl updated (%d bytes)\n", fi.Size())
	} else {
		fmt.Println(".vopl updated.")
	}
	return nil
}

// RunJSONToVOPL applies JSON updates to an empty grid and saves it as a .vopl chunk file.
func RunJSONToVOPL(jsonUpdates []byte, outPath string) error {
	var grid vopl.VoxelGrid
	if err := applyJSONToGrid(&grid, jsonUpdates); err != nil {
		return err
	}
	if err := vopl.SaveVoplGrid(&grid, outPath); err != nil {
		return fmt.Errorf("failed to save VOPL: %w", err)
	}
	if fi, err := os.Stat(outPath); err == nil {
		fmt.Printf(".vopl saved (%d bytes)\n", fi.Size())
	} else {
		fmt.Println(".vopl saved.")
	}
	return nil
}

// RunJSONToVOPLFile reads a JSON file and writes a .vopl built from it (applied over empty grid).
func RunJSONToVOPLFile(inJSONPath, outPath string) error {
	data, err := os.ReadFile(inJSONPath)
	if err != nil {
		return err
	}
	return RunJSONToVOPL(data, outPath)
}

func applyJSONToGrid(grid *vopl.VoxelGrid, jsonBlob []byte) error {
	var up updatesJSON
	if err := json.Unmarshal(jsonBlob, &up); err != nil {
		return fmt.Errorf("invalid updates JSON: %w", err)
	}
	// Apply all entries for the first chunk we see. Additional chunks are ignored for this single-grid tool.
	for _, indices := range up {
		for idxStr, col := range indices {
			// parse idx
			var idx int
			if _, err := fmt.Sscan(idxStr, &idx); err != nil {
				return fmt.Errorf("invalid voxel index '%s': %w", idxStr, err)
			}
			if idx < 0 || idx >= vopl.Width*vopl.Height*vopl.Depth {
				// ignore out of bounds
				continue
			}
			if col < 0 {
				col = 0
			}
			if col > 255 {
				col = 255
			}
			x, y, z := indexToXYZ(idx)
			grid[y][x][z] = uint8(col)
		}
		break // only one chunk supported here
	}
	return nil
}
