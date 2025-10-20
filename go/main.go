//go:build !(js && wasm)

package main

import (
	"fmt"
	"os"

	"github.com/voxelsplace/vopl/go/utils"
)

func usage() {
	fmt.Println("Usage: vopltool <command> [args]")
	fmt.Println("Commands:")
	fmt.Println("  vpi2vopl input.vpi output.vopl          (decode VPI18 stream to .vopl)")
	fmt.Println("  updatevopl input.vopl updates.vpi output.vopl  (apply VPI18 diff updates)")
	fmt.Println("  vopl2glb input.vopl output.glb         (convert .vopl -> .glb using greedy mesh)")
	fmt.Println("  voplpack2glb input.voplpack output.glb (convert .voplpack -> .glb, one node per entry)")
	fmt.Println("  vopl2voplpack output.voplpack input1.vopl [input2.vopl ...]   (pack multiple .vopl into a .voplpack)")
	fmt.Println("  voplpack2vopl input.voplpack output_dir  (unpack .voplpack into directory of .vopl files)")
	fmt.Println("  gennoise <percentage> <amount> <output_dir>                         (generate N random .vopl chunks with fixed fill %)")
	fmt.Println("  gennoise <percentageMin> <percentageMax> <amount> <output_dir>     (generate with per-file random fill in [min,max])")
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "vpi2vopl":
		if len(os.Args) != 4 {
			usage()
			os.Exit(1)
		}
		if err := utils.RunVPI18ToVOPLFile(os.Args[2], os.Args[3]); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	case "updatevopl":
		if len(os.Args) != 5 {
			usage()
			os.Exit(1)
		}
		updates, err := os.ReadFile(os.Args[3])
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		if err := utils.RunUpdateVOPL(updates, os.Args[2], os.Args[4]); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	case "vopl2glb":
		if len(os.Args) != 4 {
			usage()
			os.Exit(1)
		}
		if err := utils.RunVOPL2GLB(os.Args[2], os.Args[3]); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	case "voplpack2glb":
		if len(os.Args) != 4 {
			usage()
			os.Exit(1)
		}
		if err := utils.RunVOPLPACK2GLB(os.Args[2], os.Args[3]); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	case "vopl2voplpack":
		if len(os.Args) < 4 {
			usage()
			os.Exit(1)
		}
		output := os.Args[2]
		inputs := os.Args[3:]
		if err := utils.CreatePack(inputs, output); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	case "voplpack2vopl":
		if len(os.Args) != 4 {
			usage()
			os.Exit(1)
		}
		if err := utils.RunVOPLPACK2VOPL(os.Args[2], os.Args[3]); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	case "gennoise":
		// Two forms:
		// 1) gennoise <percentage> <amount> <output_dir>
		// 2) gennoise <percentageMin> <percentageMax> <amount> <output_dir>
		if len(os.Args) == 5 {
			var perc float64
			var amt int
			if _, err := fmt.Sscan(os.Args[2], &perc); err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}
			if _, err := fmt.Sscan(os.Args[3], &amt); err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}
			if err := utils.RunGenerateNoiseVOPL(perc, amt, os.Args[4]); err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}
		} else if len(os.Args) == 6 {
			var minP, maxP float64
			var amt int
			if _, err := fmt.Sscan(os.Args[2], &minP); err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}
			if _, err := fmt.Sscan(os.Args[3], &maxP); err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}
			if _, err := fmt.Sscan(os.Args[4], &amt); err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}
			if err := utils.RunGenerateNoiseVOPLRange(minP, maxP, amt, os.Args[5]); err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}
		} else {
			usage()
			os.Exit(1)
		}
	default:
		usage()
		os.Exit(1)
	}

	fmt.Println("Operation completed!")
}
