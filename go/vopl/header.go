package vopl

// VOPLHeaderV3 represents the fixed fields in a VOPL v3 header.
// Kept in its own file for clarity and reuse across pack/unpack helpers.
// Note: The per-file 'encoding' byte is not part of this common header struct
// because it varies per entry and is stored alongside each payload when packing.
// Ver must be 3 for VOPL v3.

type VOPLHeaderV3 struct {
	Ver  uint8
	BPP  uint8
	W    uint8
	H    uint8
	D    uint8
	Pal  uint16
	PLen uint32 // payload length when parsing full .vopl files
}
