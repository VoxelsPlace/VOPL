package utils

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/voxelsplace/vopl/go/vopl"
)

// generateNoiseGrid creates a 16x16x16 grid with a given percentage of voxels filled
// with random palette indices in the range [1..63]. Remaining voxels are 0 (empty).
func generateNoiseGrid(percentage float64, r *rand.Rand) *vopl.VoxelGrid {
	if percentage < 0 {
		percentage = 0
	}
	if percentage > 100 {
		percentage = 100
	}
	total := vopl.Width * vopl.Height * vopl.Depth // 4096
	want := int(float64(total)*(percentage/100.0) + 0.5)
	if want < 0 {
		want = 0
	}
	if want > total {
		want = total
	}

	// Build list of all positions and shuffle partially
	idx := make([]int, total)
	for i := range idx {
		idx[i] = i
	}
	// Fisher-Yates shuffle only first 'want' items for efficiency
	for i := 0; i < want; i++ {
		j := i + r.Intn(total-i)
		idx[i], idx[j] = idx[j], idx[i]
	}

	var grid vopl.VoxelGrid
	for k := 0; k < want; k++ {
		i := idx[k]
		y := i / (vopl.Width * vopl.Depth)
		rem := i % (vopl.Width * vopl.Depth)
		x := rem / vopl.Depth
		z := rem % vopl.Depth
		// random color 1..63 (0 is empty)
		color := uint8(1 + r.Intn(63))
		grid[y][x][z] = color
	}
	return &grid
}

// RunGenerateNoiseVOPL creates 'amount' .vopl files named 0.vopl..(amount-1).vopl
// in outDir, each containing random noise with the specified percentage fill.
func RunGenerateNoiseVOPL(percentage float64, amount int, outDir string) error {
	// Backward-compatible wrapper using a fixed percentage for all files.
	return RunGenerateNoiseVOPLRange(percentage, percentage, amount, outDir)
}

// RunGenerateNoiseVOPLRange generates amount .vopl files with a random fill percentage
// uniformly sampled in [percentageMin, percentageMax] for each file.
func RunGenerateNoiseVOPLRange(percentageMin, percentageMax float64, amount int, outDir string) error {
	if amount < 0 {
		amount = 0
	}
	if outDir == "" {
		outDir = "."
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	// clamp and normalize range
	if percentageMin < 0 {
		percentageMin = 0
	}
	if percentageMax > 100 {
		percentageMax = 100
	}
	if percentageMax < percentageMin {
		percentageMin, percentageMax = percentageMax, percentageMin
	}

	// Seed base once and derive per-file seeds deterministically
	baseSeed := uint64(time.Now().UnixNano())
	for i := 0; i < amount; i++ {
		// derive a seed per file using a Weyl-like progression (unsigned math)
		const weyl = uint64(0x9e3779b97f4a7c15)
		seed := baseSeed ^ (uint64(i)+1)*weyl
		r := rand.New(rand.NewSource(int64(seed & 0x7fffffffffffffff)))

		// random percentage per file within [min,max]
		perc := percentageMin
		if percentageMax > percentageMin {
			perc = percentageMin + r.Float64()*(percentageMax-percentageMin)
		}

		grid := generateNoiseGrid(perc, r)
		path := filepath.Join(outDir, fmt.Sprintf("%d.vopl", i))
		if err := vopl.SaveVoplGrid(grid, path); err != nil {
			return fmt.Errorf("falha ao salvar %s: %w", path, err)
		}
	}
	return nil
}
