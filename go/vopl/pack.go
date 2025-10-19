package vopl

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
)

// ParseVOPLHeaderV3FromBytes parses a VOPL v3 header from the given full file bytes,
// returning the header and the payload slice.
func ParseVOPLHeaderV3FromBytes(data []byte) (VOPLHeaderV3, []byte, error) {
	var hdr VOPLHeaderV3
	if len(data) < 16 || string(data[:4]) != "VOPL" {
		return hdr, nil, fmt.Errorf("não é um VOPL válido")
	}
	r := bytes.NewReader(data[4:])
	var ver uint8
	if err := binary.Read(r, binary.LittleEndian, &ver); err != nil {
		return hdr, nil, err
	}
	if ver != 3 {
		return hdr, nil, fmt.Errorf("VOPL versão não suportada: %d", ver)
	}
	var enc uint8 // we read and discard here (per-file, not common)
	if err := binary.Read(r, binary.LittleEndian, &enc); err != nil {
		return hdr, nil, err
	}
	if err := binary.Read(r, binary.LittleEndian, &hdr.BPP); err != nil {
		return hdr, nil, err
	}
	if err := binary.Read(r, binary.LittleEndian, &hdr.W); err != nil {
		return hdr, nil, err
	}
	if err := binary.Read(r, binary.LittleEndian, &hdr.H); err != nil {
		return hdr, nil, err
	}
	if err := binary.Read(r, binary.LittleEndian, &hdr.D); err != nil {
		return hdr, nil, err
	}
	if err := binary.Read(r, binary.LittleEndian, &hdr.Pal); err != nil {
		return hdr, nil, err
	}
	if err := binary.Read(r, binary.LittleEndian, &hdr.PLen); err != nil {
		return hdr, nil, err
	}
	if uint32(len(data)-16) != hdr.PLen {
		return hdr, nil, fmt.Errorf("payload length inválido (esperado %d)", hdr.PLen)
	}
	hdr.Ver = 3
	payload := data[16:]
	return hdr, payload, nil
}

// BuildVOPLFromHeaderAndPayload reconstructs a full .vopl file from the given
// common header fields and the per-file encoding and payload.
func BuildVOPLFromHeaderAndPayload(h VOPLHeaderV3, enc uint8, payload []byte) []byte {
	var buf bytes.Buffer
	buf.WriteString("VOPL")
	_ = binary.Write(&buf, binary.LittleEndian, uint8(h.Ver))
	_ = binary.Write(&buf, binary.LittleEndian, enc)
	_ = binary.Write(&buf, binary.LittleEndian, h.BPP)
	_ = binary.Write(&buf, binary.LittleEndian, h.W)
	_ = binary.Write(&buf, binary.LittleEndian, h.H)
	_ = binary.Write(&buf, binary.LittleEndian, h.D)
	_ = binary.Write(&buf, binary.LittleEndian, h.Pal)
	_ = binary.Write(&buf, binary.LittleEndian, uint32(len(payload)))
	_, _ = buf.Write(payload)
	return buf.Bytes()
}

// PackCompression indicates the compression used for the pack content section.
type PackCompression uint8

const (
	PackCompNone PackCompression = 0
	PackCompZlib PackCompression = 1
)

const (
	packMagicStr = "VOPLPACK"
	packVersion  = 1
)

// PackEntry represents a single .vopl payload inside the pack.
type PackEntry struct {
	Name    string
	Enc     uint8
	Payload []byte
}

// Pack holds the common header information and entries.
type Pack struct {
	Header  VOPLHeaderV3 // common across all entries
	Entries []PackEntry
}

// Marshal encodes the pack into bytes, using the specified compression for the content section.
func (p *Pack) Marshal(comp PackCompression) ([]byte, error) {
	if p.Header.Ver != 3 {
		return nil, fmt.Errorf("apenas VOPL v3 é suportado no pack")
	}
	// Build uncompressed content: common fields + N entries [nameLen|name|enc|plen|payload]
	var content bytes.Buffer
	_ = binary.Write(&content, binary.LittleEndian, uint8(p.Header.Ver))
	_ = binary.Write(&content, binary.LittleEndian, p.Header.BPP)
	_ = binary.Write(&content, binary.LittleEndian, p.Header.W)
	_ = binary.Write(&content, binary.LittleEndian, p.Header.H)
	_ = binary.Write(&content, binary.LittleEndian, p.Header.D)
	_ = binary.Write(&content, binary.LittleEndian, p.Header.Pal)
	_ = binary.Write(&content, binary.LittleEndian, uint32(len(p.Entries)))
	for _, e := range p.Entries {
		nb := []byte(e.Name)
		if len(nb) > 0xFFFF {
			return nil, fmt.Errorf("nome muito longo: %s", e.Name)
		}
		_ = binary.Write(&content, binary.LittleEndian, uint16(len(nb)))
		_, _ = content.Write(nb)
		_ = binary.Write(&content, binary.LittleEndian, e.Enc)
		_ = binary.Write(&content, binary.LittleEndian, uint32(len(e.Payload)))
		_, _ = content.Write(e.Payload)
	}

	// Compress if requested
	var finalContent []byte
	var compType = comp
	switch comp {
	case PackCompNone:
		finalContent = content.Bytes()
	case PackCompZlib:
		var buf bytes.Buffer
		zw := zlib.NewWriter(&buf)
		if _, err := zw.Write(content.Bytes()); err != nil {
			return nil, err
		}
		if err := zw.Close(); err != nil {
			return nil, err
		}
		finalContent = buf.Bytes()
	default:
		return nil, fmt.Errorf("compressão não suportada: %d", comp)
	}

	// Build pack header
	var out bytes.Buffer
	out.WriteString(packMagicStr)
	_ = binary.Write(&out, binary.LittleEndian, uint8(packVersion))
	_ = binary.Write(&out, binary.LittleEndian, uint8(compType))
	_, _ = out.Write(finalContent)
	return out.Bytes(), nil
}

// UnmarshalPack parses a .voplpack from bytes and returns the pack structure and compression used.
func UnmarshalPack(data []byte) (*Pack, PackCompression, error) {
	if len(data) < 10 || string(data[:8]) != packMagicStr {
		return nil, 0, fmt.Errorf("não é um .voplpack válido")
	}
	if data[8] != packVersion {
		return nil, 0, fmt.Errorf("versão de pack não suportada: %d", data[8])
	}
	comp := PackCompression(data[9])
	contentBytes := data[10:]
	switch comp {
	case PackCompNone:
		// no-op
	case PackCompZlib:
		zr, err := zlib.NewReader(bytes.NewReader(contentBytes))
		if err != nil {
			return nil, 0, err
		}
		defer zr.Close()
		b, err := io.ReadAll(zr)
		if err != nil {
			return nil, 0, err
		}
		contentBytes = b
	default:
		return nil, 0, fmt.Errorf("tipo de compressão não suportado: %d", comp)
	}

	r := bytes.NewReader(contentBytes)
	var hdr VOPLHeaderV3
	if err := binary.Read(r, binary.LittleEndian, &hdr.Ver); err != nil {
		return nil, 0, err
	}
	if err := binary.Read(r, binary.LittleEndian, &hdr.BPP); err != nil {
		return nil, 0, err
	}
	if err := binary.Read(r, binary.LittleEndian, &hdr.W); err != nil {
		return nil, 0, err
	}
	if err := binary.Read(r, binary.LittleEndian, &hdr.H); err != nil {
		return nil, 0, err
	}
	if err := binary.Read(r, binary.LittleEndian, &hdr.D); err != nil {
		return nil, 0, err
	}
	if err := binary.Read(r, binary.LittleEndian, &hdr.Pal); err != nil {
		return nil, 0, err
	}
	var n uint32
	if err := binary.Read(r, binary.LittleEndian, &n); err != nil {
		return nil, 0, err
	}

	pack := &Pack{Header: hdr, Entries: make([]PackEntry, n)}
	for i := uint32(0); i < n; i++ {
		var nameLen uint16
		if err := binary.Read(r, binary.LittleEndian, &nameLen); err != nil {
			return nil, 0, err
		}
		nameBytes := make([]byte, nameLen)
		if _, err := io.ReadFull(r, nameBytes); err != nil {
			return nil, 0, err
		}
		var enc uint8
		if err := binary.Read(r, binary.LittleEndian, &enc); err != nil {
			return nil, 0, err
		}
		var plen uint32
		if err := binary.Read(r, binary.LittleEndian, &plen); err != nil {
			return nil, 0, err
		}
		payload := make([]byte, plen)
		if _, err := io.ReadFull(r, payload); err != nil {
			return nil, 0, err
		}
		pack.Entries[i] = PackEntry{Name: string(nameBytes), Enc: enc, Payload: payload}
	}
	return pack, comp, nil
}
