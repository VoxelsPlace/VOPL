package vopl

import (
    "bytes"
    "encoding/binary"
    "fmt"
    "io"
    "os"
)

func SaveVoplGridV3(grid *VoxelGrid, filename string) error {
	const bpp = 6 // 64 cores (0..63)
	enc := bestEncoding(grid, bpp)
	buf := new(bytes.Buffer)
	buf.WriteString("VOPL")
	_ = binary.Write(buf, binary.LittleEndian, uint8(3))
	_ = binary.Write(buf, binary.LittleEndian, uint8(enc.encoding))
	_ = binary.Write(buf, binary.LittleEndian, uint8(bpp))
	_ = binary.Write(buf, binary.LittleEndian, uint8(Width))
	_ = binary.Write(buf, binary.LittleEndian, uint8(Height))
	_ = binary.Write(buf, binary.LittleEndian, uint8(Depth))
	_ = binary.Write(buf, binary.LittleEndian, uint16(64)) // palVer
	_ = binary.Write(buf, binary.LittleEndian, uint32(len(enc.payload)))
	buf.Write(enc.payload)
	return os.WriteFile(filename, buf.Bytes(), 0644)
}

func LoadVoplGrid(filename string) (*VoxelGrid, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    if len(data) < 4 || string(data[:4]) != "VOPL" {
        return nil, fmt.Errorf("formato inválido ou não é VOPL")
    }
    br := bytes.NewReader(data[4:])
    var ver uint8
    if err := binary.Read(br, binary.LittleEndian, &ver); err != nil {
        return nil, err
    }
    if ver != 3 {
        return nil, fmt.Errorf("apenas VOPL v3 é suportado (encontrado %d)", ver)
    }
    return loadV3(br)
}

// v1 and v2 removed: only v3 is supported

func loadV3(b *bytes.Reader) (*VoxelGrid, error) {
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
			idx, err := br.readBits(8)
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
	case encRLE:
		br := newBitReader(payload)
		lin := make([]uint8, 0, Width*Height*Depth)
		for len(lin) < Width*Height*Depth {
			ln, err := br.readBits(8)
			if err != nil {
				return nil, err
			}
			col, err := br.readBits(bpp)
			if err != nil {
				return nil, err
			}
			count := int(ln) + 1
			for j := 0; j < count; j++ {
				lin = append(lin, uint8(col))
			}
		}
		if len(lin) != Width*Height*Depth {
			return nil, fmt.Errorf("RLE inválido (v3)")
		}
		applyOrder(grid, lin)
	default:
		return nil, fmt.Errorf("encoding desconhecido (v3): %d", enc)
	}
	return grid, nil
}

// v1 mesh loader and meshToGrid removed

func ExpandRLE(rle []int) (*VoxelGrid, error) {
	if len(rle)%2 != 0 {
		return nil, fmt.Errorf("RLE deve ter pares count-value")
	}
	var grid VoxelGrid
	total := Height * Width * Depth
	idx := 0
	for i := 0; i < len(rle); i += 2 {
		count := rle[i]
		value := rle[i+1]
		if value < 0 || value > 63 {
			return nil, fmt.Errorf("Valor inválido: %d (0-63)", value)
		}
		for j := 0; j < count; j++ {
			if idx >= total {
				return nil, fmt.Errorf("RLE excede o tamanho do chunk")
			}
			y := idx / (Width * Depth)
			z := (idx / Width) % Depth
			x := idx % Width
			grid[y][x][z] = uint8(value)
			idx++
		}
	}
	if idx != total {
		return nil, fmt.Errorf("RLE não preenche o chunk inteiro (%d/%d)", idx, total)
	}
	return &grid, nil
}
