package utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/voxelsplace/vopl/go/vopl"
)

func RunRLE2VOPL(rleArg, outPath string) error {
	rleStr := strings.Trim(rleArg, "[] ")
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

	grid, err := vopl.ExpandRLE(rle)
	if err != nil {
		return fmt.Errorf("failed to expand RLE: %w", err)
	}
	if err := vopl.SaveVoplGridV3(grid, outPath); err != nil {
		return fmt.Errorf("failed to save VOPL: %w", err)
	}

	if fi, err := os.Stat(outPath); err == nil {
		fmt.Printf(".vopl saved (%d bytes)\n", fi.Size())
	} else {
		fmt.Println(".vopl saved.")
	}
	return nil
}
