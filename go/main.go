package main

import (
	"fmt"
	"os"

	"vopltool/utils"
)

func usage() {
	fmt.Println("Uso: vopltool <command> [args]")
	fmt.Println("Comandos:")
	fmt.Println("  rle2vopl \"10,0,4,1,...\" output.vopl    (gera .vopl v3 64 cores a partir de RLE)")
	fmt.Println("  vopl2glb input.vopl output.glb         (converte .vopl -> .glb com greedy mesh)")
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
			fmt.Println("Erro:", err)
			os.Exit(1)
		}
	case "vopl2glb":
		if len(os.Args) != 4 {
			usage()
			os.Exit(1)
		}
		if err := utils.RunVOPL2GLB(os.Args[2], os.Args[3]); err != nil {
			fmt.Println("Erro:", err)
			os.Exit(1)
		}
	default:
		usage()
		os.Exit(1)
	}

	fmt.Println("Operação completa!")
}
