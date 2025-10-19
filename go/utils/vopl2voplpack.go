package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/voxelsplace/vopl/go/vopl"
)

// CreatePack reads .vopl files and writes a .voplpack to outputFile.
// It verifies common header fields across inputs and uses zlib compression.
func CreatePack(inputFiles []string, outputFile string) error {
	if len(inputFiles) == 0 {
		return fmt.Errorf("no .vopl files provided")
	}
	type item struct {
		name    string
		payload []byte
		enc     uint8
		hdr     vopl.VOPLHeader
		err     error
	}
	items := make([]item, len(inputFiles))

	var wg sync.WaitGroup
	for i := range inputFiles {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			path := inputFiles[i]
			b, err := os.ReadFile(path)
			if err != nil {
				items[i].err = err
				return
			}
			hdr, payload, err := vopl.ParseVOPLHeaderFromBytes(b)
			if err != nil {
				items[i].err = err
				return
			}
			// Read encoding (byte 5 after magic): at offset 5 in file
			enc := b[5]
			items[i] = item{
				name:    filepath.Base(path),
				payload: payload,
				enc:     enc,
				hdr:     hdr,
			}
		}(i)
	}
	wg.Wait()
	// check for errors and common fields
	common := items[0].hdr
	for i, it := range items {
		if it.err != nil {
			return it.err
		}
		if it.hdr.Ver != 3 {
			return fmt.Errorf("apenas VOPL é suportado (%s)", inputFiles[i])
		}
		if it.hdr.BPP != common.BPP || it.hdr.W != common.W || it.hdr.H != common.H || it.hdr.D != common.D || it.hdr.Pal != common.Pal {
			return fmt.Errorf("inconsistent parameters between files (%s)", inputFiles[i])
		}
	}

	pack := &vopl.Pack{Header: vopl.VOPLHeader{Ver: 3, BPP: common.BPP, W: common.W, H: common.H, D: common.D, Pal: common.Pal}}
	pack.Entries = make([]vopl.PackEntry, len(items))
	for i, it := range items {
		pack.Entries[i] = vopl.PackEntry{Name: it.name, Enc: it.enc, Payload: it.payload}
	}
	start := time.Now()
	data, err := pack.Marshal(vopl.PackCompZlib)
	if err != nil {
		return err
	}
	dur := time.Since(start)
	fmt.Printf("Compressão (vopl2voplpack) levou %d ms\n", dur.Milliseconds())
	return os.WriteFile(outputFile, data, 0o644)
}

// UnpackToDir writes .vopl files from a .voplpack into outputDir.
func UnpackToDir(packFile, outputDir string) error {
	data, err := os.ReadFile(packFile)
	if err != nil {
		return err
	}
	pack, _, err := vopl.UnmarshalPack(data)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}
	// Parallel write
	var wg sync.WaitGroup
	errCh := make(chan error, len(pack.Entries))
	for _, e := range pack.Entries {
		wg.Add(1)
		go func(e vopl.PackEntry) {
			defer wg.Done()
			voplBytes := vopl.BuildVOPLFromHeaderAndPayload(pack.Header, e.Enc, e.Payload)
			if err := os.WriteFile(filepath.Join(outputDir, e.Name), voplBytes, 0o644); err != nil {
				errCh <- err
			}
		}(e)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}

// UnpackToMemory returns names and raw .vopl bytes without writing to disk.
func UnpackToMemory(packFile string) ([]string, [][]byte, error) {
	data, err := os.ReadFile(packFile)
	if err != nil {
		return nil, nil, err
	}
	pack, _, err := vopl.UnmarshalPack(data)
	if err != nil {
		return nil, nil, err
	}
	names := make([]string, len(pack.Entries))
	blobs := make([][]byte, len(pack.Entries))
	for i, e := range pack.Entries {
		names[i] = e.Name
		blobs[i] = vopl.BuildVOPLFromHeaderAndPayload(pack.Header, e.Enc, e.Payload)
	}
	return names, blobs, nil
}
