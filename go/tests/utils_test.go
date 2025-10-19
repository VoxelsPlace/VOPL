package test

import (
	"bytes"
	"os"
	"testing"

	"github.com/voxelsplace/vopl/go/utils"
)

func TestRle2Vopl(t *testing.T) {
	rle := "256,1,34,7,12,0,4,7,12,0,4,7,12,0,4,7,12,0,4,7,12,0,4,7,12,0,4,7,12,0,4,7,12,0,4,7,12,0,4,7,12,0,4,7,12,0,4,7,12,0,34,7,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,2,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,2,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,2,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,2,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,2,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,2,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,2,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,2,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,2,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,2,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,2,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,2,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,2,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,2,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,2,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,1,0,1,19,3328,0"
	outPath := "output/test_output.vopl"
	expectedPath := "expected/fromrle.vopl"
	err := utils.RunRLE2VOPL(rle, outPath)
	if err != nil {
		t.Fatalf("RunRLE2VOPL failed: %v", err)
	}
	outData, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}
	expectedData, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("Failed to read expected file: %v", err)
	}
	if bytes.Equal(outData, expectedData) == false {
		t.Fatalf("Output does not match expected output")
	}
}

func TestVopl2Glb(t *testing.T) {
	inputPath := "expected/fromrle.vopl"
	outPath := "output/test_output.glb"
	expectedPath := "expected/fromvopl.glb"
	err := utils.RunVOPL2GLB(inputPath, outPath)
	if err != nil {
		t.Fatalf("RunVOPL2GLB failed: %v", err)
	}
	outData, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}
	expectedData, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("Failed to read expected file: %v", err)
	}
	if bytes.Equal(outData, expectedData) == false {
		t.Fatalf("Output does not match expected output")
	}
}

func TestRle2Vopl_InvalidRLE(t *testing.T) {
	rle := "34,7,abc,0"
	outPath := "test_output_invalid.vopl"
	err := utils.RunRLE2VOPL(rle, outPath)
	if err == nil {
		t.Fatalf("Expected error for invalid RLE input, got nil")
	}
}

func TestRle2Vopl_EmptyRLE(t *testing.T) {
	rle := ""
	outPath := "test_output_empty.vopl"
	err := utils.RunRLE2VOPL(rle, outPath)
	if err == nil {
		t.Fatalf("Expected error for empty RLE input, got nil")
	}
}
