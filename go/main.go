package main

import (
	"fmt"
	"os"

	"github.com/voxelsplace/vopl/go/utils"
)

func usage() {
	fmt.Println("Usage: vopltool <command> [args]")
	fmt.Println("Commands:")
	fmt.Println("  rle2vopl \"10,0,4,1,...\" output.vopl    (generate .vopl v3 with 64 colors from RLE)")
	fmt.Println("  vopl2glb input.vopl output.glb         (convert .vopl -> .glb using greedy mesh)")
	fmt.Println("  vopl2voplpack output.voplpack input1.vopl [input2.vopl ...]   (pack multiple .vopl into a .voplpack)")
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "rle2vopl":
		if len(os.Args) != 4 {
			usage()
			os.Exit(1)
		}
		if err := utils.RunRLE2VOPL(os.Args[2], os.Args[3]); err != nil {
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
	default:
		usage()
		os.Exit(1)
	}

	fmt.Println("Operation completed!")
}
