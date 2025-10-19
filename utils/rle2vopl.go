package utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"vopltool/vopl"
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
			return fmt.Errorf("erro ao converter RLE '%s': %w", p, err)
		}
		rle = append(rle, i)
	}

	grid, err := vopl.ExpandRLE(rle)
	if err != nil {
		return fmt.Errorf("erro ao expandir RLE: %w", err)
	}
	if err := vopl.SaveVoplGridV3(grid, outPath); err != nil {
		return fmt.Errorf("erro ao salvar VOPL: %w", err)
	}

	if fi, err := os.Stat(outPath); err == nil {
		fmt.Printf(".vopl salvo (%d bytes)\n", fi.Size())
	} else {
		fmt.Println(".vopl salvo.")
	}
	return nil
}
