package vopl

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

func SaveVoplGrid(grid *VoxelGrid, filename string) error {
	// Use a fixed BPP=6 (palette size 64) to guarantee consistent headers across chunks.
	data := SaveVoplGridToBytesWithBPP(grid, 6)
	return os.WriteFile(filename, data, 0644)
}

// SaveVoplGridToBytes returns the .vopl file as bytes instead of writing to disk.
func SaveVoplGridToBytes(grid *VoxelGrid) []byte {
	// Fixed BPP=6 ensures packable uniform headers.
	return SaveVoplGridToBytesWithBPP(grid, 6)
}

// SaveVoplGridToBytesWithBPP encodes a grid using the specified bits-per-pixel (1..8)
// and returns a complete .vopl file as bytes. Using a fixed BPP across chunks
// guarantees headers remain consistent and can be packed together.
func SaveVoplGridToBytesWithBPP(grid *VoxelGrid, bpp uint8) []byte {
	if bpp < 1 {
		bpp = 1
	}
	if bpp > 8 {
		bpp = 8
	}
	enc := bestEncoding(grid, bpp)
	hdr := VOPLHeader{Ver: 3, BPP: bpp, W: uint8(Width), H: uint8(Height), D: uint8(Depth), Pal: 64}
	return BuildVOPLFromHeaderAndPayload(hdr, uint8(enc.encoding), enc.payload)
}

func LoadVoplGrid(filename string) (*VoxelGrid, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return LoadVoplGridFromBytes(data)
}

// LoadVoplGridFromBytes parses a .vopl file from memory and returns the grid.
func LoadVoplGridFromBytes(data []byte) (*VoxelGrid, error) {
	if len(data) < 4 || string(data[:4]) != "VOPL" {
		return nil, fmt.Errorf("invalid format or not VOPL")
	}
	br := bytes.NewReader(data[4:])
	var ver uint8
	if err := binary.Read(br, binary.LittleEndian, &ver); err != nil {
		return nil, err
	}
	if ver != 3 {
		return nil, fmt.Errorf("only VOPL is supported (found %d)", ver)
	}
	return load(br)
}

func load(b *bytes.Reader) (*VoxelGrid, error) {
	var encByte, bpp, w, h, d uint8
	var palVer uint16
	var plen uint32
	_ = binary.Read(b, binary.LittleEndian, &encByte)
	_ = binary.Read(b, binary.LittleEndian, &bpp)
	_ = binary.Read(b, binary.LittleEndian, &w)
	_ = binary.Read(b, binary.LittleEndian, &h)
	_ = binary.Read(b, binary.LittleEndian, &d)
	_ = binary.Read(b, binary.LittleEndian, &palVer)
	_ = binary.Read(b, binary.LittleEndian, &plen)

	payload := make([]byte, plen)
	if _, err := io.ReadFull(b, payload); err != nil {
		return nil, err
	}
	if encByte&0x80 != 0 {
		var err error
		payload, err = zlibDecompress(payload)
		if err != nil {
			return nil, err
		}
	}
	enc := int(encByte & 0x7F)
	grid := new(VoxelGrid)
	switch enc {
	case encDense:
		br := newBitReader(payload)
		lin := make([]uint8, Width*Height*Depth)
		for i := 0; i < len(lin); i++ {
			v, err := br.readBits(bpp)
			if err != nil {
				return nil, err
			}
			lin[i] = uint8(v)
		}
		applyOrder(grid, lin)
	case encSparse:
		br := newBitReader(payload)
		lin := make([]uint8, Width*Height*Depth)
		cnt, err := br.readBits(16)
		if err != nil {
			return nil, err
		}
		for i := 0; i < int(cnt); i++ {
			idx, err := br.readBits(12)
			if err != nil {
				return nil, err
			}
			col, err := br.readBits(bpp)
			if err != nil {
				return nil, err
			}
			lin[int(idx)] = uint8(col)
		}
		applyOrder(grid, lin)
	case encSparse2:
		if len(payload) < 512 {
			return nil, fmt.Errorf("payload insuficiente para Sparse2")
		}
		bitmap := payload[:512]
		vals := payload[512:]
		br := newBitReader(vals)
		lin := make([]uint8, 0, Width*Height*Depth)
		total := Width * Height * Depth
		consumed := 0
		for i := 0; i < total; i++ {
			bit := (bitmap[i>>3] >> (uint(i) & 7)) & 1
			if bit == 0 {
				lin = append(lin, 0)
				continue
			}
			v, err := br.readBits(bpp)
			if err != nil {
				return nil, err
			}
			lin = append(lin, uint8(v))
			consumed++
		}
		applyOrder(grid, lin)
	default:
		return nil, fmt.Errorf("encoding desconhecido: %d", enc)
	}
	return grid, nil
}
