package utils

// RunVOPLPACK2VOPL extracts all .vopl files from a .voplpack into the given output directory.
// It preserves the original entry names inside the pack (typically ending with .vopl).
func RunVOPLPACK2VOPL(inPackPath, outDir string) error {
	return UnpackToDir(inPackPath, outDir)
}
