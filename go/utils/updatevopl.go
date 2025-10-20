package utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/voxelsplace/vopl/go/vopl"
)

// RunUpdateVOPL applies an RLE patch to an existing .vopl file and writes a regenerated .vopl.
//
// RLE format: pairs of count,value separated by commas, e.g. "10,0,4,7,...".
// - value in [0..63] sets that palette index for 'count' voxels in raster order (y, then z, then x),
// - value == -1 means "skip" (advance 'count' voxels without changing them).
// The patch may cover the full 4096 voxels or only a prefix; any remaining voxels are left unchanged.
func RunUpdateVOPL(rleArg, inputPath, outputPath string) error {
	// Load existing grid from .vopl
	grid, err := vopl.LoadVoplGrid(inputPath)
	if err != nil {
		return fmt.Errorf("failed to load input VOPL: %w", err)
	}

	// Parse RLE string
	rleStr := strings.Trim(rleArg, "[] ")
	if rleStr == "" {
		return fmt.Errorf("empty RLE input")
	}
	parts := strings.Split(rleStr, ",")
	var rle []int
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		i, err := strconv.Atoi(p)
		if err != nil {
			return fmt.Errorf("failed to parse RLE '%s': %w", p, err)
		}
		rle = append(rle, i)
	}
	if len(rle)%2 != 0 {
		return fmt.Errorf("RLE must have count,value pairs")
	}

	total := vopl.Height * vopl.Width * vopl.Depth
	idx := 0
	for i := 0; i < len(rle) && idx < total; i += 2 {
		count := rle[i]
		val := rle[i+1]
		if count < 0 {
			return fmt.Errorf("invalid negative count: %d", count)
		}
		if val < -1 || val > 63 {
			return fmt.Errorf("invalid value: %d (allowed -1 or 0..63)", val)
		}
		if val == -1 {
			// skip without changes
			idx += count
			if idx > total {
				return fmt.Errorf("RLE exceeds chunk size")
			}
			continue
		}
		for j := 0; j < count; j++ {
			if idx >= total {
				return fmt.Errorf("RLE exceeds chunk size")
			}
			y := idx / (vopl.Width * vopl.Depth)
			z := (idx / vopl.Width) % vopl.Depth
			x := idx % vopl.Width
			grid[y][x][z] = uint8(val)
			idx++
		}
	}
	if idx > total {
		return fmt.Errorf("RLE exceeds chunk size")
	}
	// Save regenerated .vopl (writer auto-picks optimal encoding)
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
