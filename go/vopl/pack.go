package vopl

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"math"

	xxhash "github.com/cespare/xxhash/v2"
	"github.com/klauspost/compress/zstd"
)

// ParseVOPLHeaderFromBytes parses a VOPL header from the given full file bytes,
// returning the header and the payload slice.
func ParseVOPLHeaderFromBytes(data []byte) (VOPLHeader, []byte, error) {
	var hdr VOPLHeader
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
func BuildVOPLFromHeaderAndPayload(h VOPLHeader, enc uint8, payload []byte) []byte {
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
	PackCompZstd PackCompression = 2
)

const (
	packMagicStr = "VOPLPACK"
	packVersion1 = 1
	packVersion2 = 2
)

// PackLayout specifies how the content section encodes entries.
type PackLayout uint8

const (
	// LayoutRaw stores entries as independent payload blobs (v1 behavior).
	LayoutRaw PackLayout = 0
	// LayoutCDC stores a content-defined chunk dictionary and entries as sequences of chunk refs.
	LayoutCDC PackLayout = 1
)

// PackEntry represents a single .vopl payload inside the pack.
type PackEntry struct {
	Name    string
	Enc     uint8
	Payload []byte
}

// Pack holds the common header information and entries.
type Pack struct {
	Header  VOPLHeader // common across all entries
	Entries []PackEntry
}

// Marshal encodes using v1 raw layout with optional zlib compression (backward compatible).
// Prefer using MarshalEx for advanced layouts and codecs.
func (p *Pack) Marshal(comp PackCompression) ([]byte, error) {
	return p.MarshalEx(LayoutRaw, comp)
}

// MarshalEx encodes the pack into bytes with the specified layout and compression codec.
// LayoutRaw mirrors v1 semantics; LayoutCDC (v2) builds a chunk dictionary for deduplication across entries.
func (p *Pack) MarshalEx(layout PackLayout, comp PackCompression) ([]byte, error) {
	if p.Header.Ver != 3 {
		return nil, fmt.Errorf("apenas VOPL é suportado no pack")
	}
	// Decide version early: v1 for raw+none/zlib to stay backward-compatible; otherwise v2.
	version := uint8(packVersion2)
	if layout == LayoutRaw && (comp == PackCompNone || comp == PackCompZlib) {
		version = uint8(packVersion1)
	}
	var content bytes.Buffer
	_ = binary.Write(&content, binary.LittleEndian, uint8(p.Header.Ver))
	_ = binary.Write(&content, binary.LittleEndian, p.Header.BPP)
	_ = binary.Write(&content, binary.LittleEndian, p.Header.W)
	_ = binary.Write(&content, binary.LittleEndian, p.Header.H)
	_ = binary.Write(&content, binary.LittleEndian, p.Header.D)
	_ = binary.Write(&content, binary.LittleEndian, p.Header.Pal)

	switch layout {
	case LayoutRaw:
		// v1 style content (no layout byte). If using v2 container, we include layout byte.
		if version >= packVersion2 {
			_ = binary.Write(&content, binary.LittleEndian, uint8(LayoutRaw))
		}
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
	case LayoutCDC:
		// content-defined chunking parameters
		target := uint32(4096)
		minSz := uint32(2048)
		maxSz := uint32(16384)
		_ = binary.Write(&content, binary.LittleEndian, uint8(LayoutCDC))
		_ = binary.Write(&content, binary.LittleEndian, target)
		_ = binary.Write(&content, binary.LittleEndian, minSz)
		_ = binary.Write(&content, binary.LittleEndian, maxSz)

		// Build chunk dictionary across all entry payloads
		dict, sequences := buildCDCIndex(p.Entries, int(target), int(minSz), int(maxSz))
		// Write blocks
		_ = binary.Write(&content, binary.LittleEndian, uint32(len(dict)))
		for _, blk := range dict {
			_ = binary.Write(&content, binary.LittleEndian, uint32(len(blk)))
			_, _ = content.Write(blk)
		}
		// Write entries
		_ = binary.Write(&content, binary.LittleEndian, uint32(len(p.Entries)))
		for i, e := range p.Entries {
			nb := []byte(e.Name)
			if len(nb) > 0xFFFF {
				return nil, fmt.Errorf("nome muito longo: %s", e.Name)
			}
			_ = binary.Write(&content, binary.LittleEndian, uint16(len(nb)))
			_, _ = content.Write(nb)
			_ = binary.Write(&content, binary.LittleEndian, e.Enc)
			_ = binary.Write(&content, binary.LittleEndian, uint32(len(e.Payload))) // original raw length
			seq := sequences[i]
			_ = binary.Write(&content, binary.LittleEndian, uint32(len(seq)))
			for _, idx := range seq {
				_ = binary.Write(&content, binary.LittleEndian, uint32(idx))
			}
		}
	default:
		return nil, fmt.Errorf("layout não suportado: %d", layout)
	}

	// Compress if requested
	var finalContent []byte
	var compType = comp
	switch comp {
	case PackCompNone:
		finalContent = content.Bytes()
	case PackCompZlib:
		var buf bytes.Buffer
		zw, _ := zlib.NewWriterLevel(&buf, zlib.BestCompression)
		if _, err := zw.Write(content.Bytes()); err != nil {
			return nil, err
		}
		if err := zw.Close(); err != nil {
			return nil, err
		}
		finalContent = buf.Bytes()
	case PackCompZstd:
		enc, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedDefault))
		if err != nil {
			return nil, err
		}
		finalContent = enc.EncodeAll(content.Bytes(), nil)
	default:
		return nil, fmt.Errorf("compressão não suportada: %d", comp)
	}

	// Build pack header
	var out bytes.Buffer
	out.WriteString(packMagicStr)
	_ = binary.Write(&out, binary.LittleEndian, version)
	_ = binary.Write(&out, binary.LittleEndian, uint8(compType))
	_, _ = out.Write(finalContent)
	return out.Bytes(), nil
}

// UnmarshalPack parses a .voplpack from bytes and returns the pack structure and compression used.
func UnmarshalPack(data []byte) (*Pack, PackCompression, error) {
	if len(data) < 10 || string(data[:8]) != packMagicStr {
		return nil, 0, fmt.Errorf("não é um .voplpack válido")
	}
	version := data[8]
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
	case PackCompZstd:
		dec, err := zstd.NewReader(nil)
		if err != nil {
			return nil, 0, err
		}
		defer dec.Close()
		b, err := dec.DecodeAll(contentBytes, nil)
		if err != nil {
			return nil, 0, err
		}
		contentBytes = b
	default:
		return nil, 0, fmt.Errorf("tipo de compressão não suportado: %d", comp)
	}

	r := bytes.NewReader(contentBytes)
	var hdr VOPLHeader
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

	// v1 has no layout byte; v2 includes layout after common header
	var layout PackLayout = LayoutRaw
	if version >= packVersion2 {
		var lb uint8
		if err := binary.Read(r, binary.LittleEndian, &lb); err != nil {
			return nil, 0, err
		}
		layout = PackLayout(lb)
	} else if version != packVersion1 {
		return nil, 0, fmt.Errorf("versão de pack não suportada: %d", version)
	}

	switch layout {
	case LayoutRaw:
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
	case LayoutCDC:
		// read CDC params
		var target, minSz, maxSz uint32
		if err := binary.Read(r, binary.LittleEndian, &target); err != nil {
			return nil, 0, err
		}
		if err := binary.Read(r, binary.LittleEndian, &minSz); err != nil {
			return nil, 0, err
		}
		if err := binary.Read(r, binary.LittleEndian, &maxSz); err != nil {
			return nil, 0, err
		}
		var nBlocks uint32
		if err := binary.Read(r, binary.LittleEndian, &nBlocks); err != nil {
			return nil, 0, err
		}
		blocks := make([][]byte, nBlocks)
		for i := uint32(0); i < nBlocks; i++ {
			var blen uint32
			if err := binary.Read(r, binary.LittleEndian, &blen); err != nil {
				return nil, 0, err
			}
			b := make([]byte, blen)
			if _, err := io.ReadFull(r, b); err != nil {
				return nil, 0, err
			}
			blocks[i] = b
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
			var rawLen uint32
			if err := binary.Read(r, binary.LittleEndian, &rawLen); err != nil {
				return nil, 0, err
			}
			var seqLen uint32
			if err := binary.Read(r, binary.LittleEndian, &seqLen); err != nil {
				return nil, 0, err
			}
			// reconstruct payload by concatenating referenced blocks
			var total uint64
			idxs := make([]uint32, seqLen)
			for j := uint32(0); j < seqLen; j++ {
				var idx uint32
				if err := binary.Read(r, binary.LittleEndian, &idx); err != nil {
					return nil, 0, err
				}
				if idx >= nBlocks {
					return nil, 0, fmt.Errorf("índice de bloco inválido: %d", idx)
				}
				idxs[j] = idx
				total += uint64(len(blocks[idx]))
				if total > uint64(rawLen)+uint64(maxSz) { // sanity
					return nil, 0, fmt.Errorf("tamanho inconsistente em sequência CDC")
				}
			}
			payload := make([]byte, 0, int(total))
			for _, idx := range idxs {
				payload = append(payload, blocks[idx]...)
			}
			if uint32(len(payload)) != rawLen {
				// Trim if last block overshot (shouldn't happen with fixed concatenation, but guard)
				if uint32(len(payload)) > rawLen {
					payload = payload[:rawLen]
				}
			}
			pack.Entries[i] = PackEntry{Name: string(nameBytes), Enc: enc, Payload: payload}
		}
		_ = target
		_ = minSz
		_ = maxSz // future use/validation
		return pack, comp, nil
	default:
		return nil, 0, fmt.Errorf("layout desconhecido: %d", layout)
	}
}

// buildCDCIndex performs a content-defined chunking (CDC) over all entry payloads,
// building a dictionary of unique chunks and returning for each entry the sequence of chunk indices.
func buildCDCIndex(entries []PackEntry, target, minSz, maxSz int) ([][]byte, [][]int) {
	// Build gear table deterministically using xxhash over index seeding.
	gear := make([]uint64, 256)
	seed := xxhash.Sum64([]byte("vopl-cdc-gear-seed"))
	for i := 0; i < 256; i++ {
		var b [16]byte
		// mix seed with index to avoid zeros
		binary.LittleEndian.PutUint64(b[:8], seed+uint64(i)*0x9E3779B185EBCA87)
		binary.LittleEndian.PutUint64(b[8:], ^(seed + uint64(i)*0xC2B2AE3D27D4EB4F))
		v := xxhash.Sum64(b[:])
		if v == 0 {
			v = 0x9E3779B185EBCA87
		}
		gear[i] = v
	}

	// chunk map and storage
	blocks := make([][]byte, 0, 256)
	index := make(map[uint64]int, 1024)
	seqs := make([][]int, len(entries))

	// boundary mask
	// Choose mask so that average size ~= target (power of two assumption)
	pow := 1 << int(math.Round(math.Log2(float64(target))))
	if pow <= 0 {
		pow = 4096
	}
	mask := uint64(pow - 1)

	addBlock := func(b []byte) int {
		h := xxhash.Sum64(b)
		if idx, ok := index[h]; ok {
			// very rare collision risk; verify to be safe
			if bytes.Equal(blocks[idx], b) {
				return idx
			}
		}
		idx := len(blocks)
		blocks = append(blocks, append([]byte(nil), b...))
		index[h] = idx
		return idx
	}

	for i, e := range entries {
		data := e.Payload
		if len(data) == 0 {
			seqs[i] = nil
			continue
		}
		var seq []int
		start := 0
		var h uint64 = 0
		// Slide and cut
		for pos := 0; pos < len(data); pos++ {
			h = (h<<1 + gear[int(data[pos])])
			// enforce min size
			if pos-start+1 < minSz {
				continue
			}
			// cut if boundary or max size reached (except at end)
			if ((h & mask) == 0) || (pos-start+1 >= maxSz) {
				blk := data[start : pos+1]
				idx := addBlock(blk)
				seq = append(seq, idx)
				start = pos + 1
				h = 0
			}
		}
		// tail
		if start < len(data) {
			blk := data[start:]
			idx := addBlock(blk)
			seq = append(seq, idx)
		}
		seqs[i] = seq
	}
	return blocks, seqs
}
