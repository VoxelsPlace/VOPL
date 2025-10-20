package utils

import (
	"fmt"
	"os"

	"github.com/voxelsplace/vopl/go/vopl"
)

// RunUpdateVOPL applies a VPI18 bitstream update to an existing .vopl file and writes a regenerated .vopl.
func RunUpdateVOPL(vpi []byte, inputPath, outputPath string) error {
	grid, err := vopl.LoadVoplGrid(inputPath)
	if err != nil {
		return fmt.Errorf("failed to load input VOPL: %w", err)
	}
	if err := vopl.VPI18ApplyToGrid(grid, vpi); err != nil {
		return fmt.Errorf("failed to apply VPI18: %w", err)
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

// RunVPI18ToVOPL decodes a VPI18 stream and saves it as a .vopl chunk file.
func RunVPI18ToVOPL(vpi []byte, outPath string) error {
	grid, err := vopl.VPI18DecodeToGrid(vpi)
	if err != nil {
		return fmt.Errorf("failed to decode VPI18: %w", err)
	}
	if err := vopl.SaveVoplGrid(grid, outPath); err != nil {
		return fmt.Errorf("failed to save VOPL: %w", err)
	}
	if fi, err := os.Stat(outPath); err == nil {
		fmt.Printf(".vopl saved (%d bytes)\n", fi.Size())
	} else {
		fmt.Println(".vopl saved.")
	}
	return nil
}
