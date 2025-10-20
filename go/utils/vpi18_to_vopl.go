package utils

import (
	"fmt"
	"os"

	"github.com/voxelsplace/vopl/go/vopl"
)

// RunVPI18ToVOPLFile reads a VPI18 stream from a file and writes a .vopl file.
func RunVPI18ToVOPLFile(inPath, outPath string) error {
	vpi, err := os.ReadFile(inPath)
	if err != nil {
		return fmt.Errorf("read VPI18: %w", err)
	}
	grid, err := vopl.VPI18DecodeToGrid(vpi)
	if err != nil {
		return fmt.Errorf("decode VPI18: %w", err)
	}
	if err := vopl.SaveVoplGrid(grid, outPath); err != nil {
		return fmt.Errorf("save VOPL: %w", err)
	}
	return nil
}
