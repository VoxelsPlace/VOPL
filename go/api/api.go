package api

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"
	"github.com/voxelsplace/vopl/go/vopl"
)

// RLEToVOPLBytes converts an RLE string (e.g., "10,0,4,1,...") to a .vopl file as bytes.
func RLEToVOPLBytes(rleArg string) ([]byte, error) {
	rleStr := strings.Trim(rleArg, "[] ")
	parts := strings.Split(rleStr, ",")
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty RLE input")
	}
	var rle []int
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		i, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("failed to parse RLE '%s': %w", p, err)
		}
		rle = append(rle, i)
	}

	grid, err := vopl.ExpandRLE(rle)
	if err != nil {
		return nil, fmt.Errorf("failed to expand RLE: %w", err)
	}
	// Return in-memory .vopl bytes
	return vopl.SaveVoplGridToBytes(grid), nil
}

// VOPLToGLB takes a .vopl file bytes and returns a .glb bytes using greedy mesh
func VOPLToGLB(voplBytes []byte) ([]byte, error) {
	grid, err := vopl.LoadVoplGridFromBytes(voplBytes)
	if err != nil {
		return nil, err
	}
	mesh := vopl.GenerateMesh(grid)

	positions := make([][3]float32, len(mesh.Vertices))
	colors := make([][4]float32, len(mesh.Vertices))
	hasAlpha := false
	for i, v := range mesh.Vertices {
		positions[i] = v.Position
		rgba, err := vopl.ParseHexColor(vopl.Palette[v.Color])
		if err != nil {
			return nil, err
		}
		colors[i] = rgba
		if rgba[3] < 1.0 {
			hasAlpha = true
		}
	}
	indices := make([]uint32, len(mesh.Indices))
	copy(indices, mesh.Indices)
	// compute flat normals per face (same as utils)
	normals := make([][3]float32, len(positions))
	for i := 0; i < len(indices); i += 3 {
		v0, v1, v2 := indices[i], indices[i+1], indices[i+2]
		p0, p1, p2 := positions[v0], positions[v1], positions[v2]
		vec1 := [3]float32{p1[0] - p0[0], p1[1] - p0[1], p1[2] - p0[2]}
		vec2 := [3]float32{p2[0] - p0[0], p2[1] - p0[1], p2[2] - p0[2]}
		cross := [3]float32{
			vec1[1]*vec2[2] - vec1[2]*vec2[1],
			vec1[2]*vec2[0] - vec1[0]*vec2[2],
			vec1[0]*vec2[1] - vec1[1]*vec2[0],
		}
		length := float32(math.Sqrt(float64(cross[0]*cross[0] + cross[1]*cross[1] + cross[2]*cross[2])))
		if length > 0 {
			cross[0] /= length
			cross[1] /= length
			cross[2] /= length
		}
		normals[v0] = cross
		normals[v1] = cross
		normals[v2] = cross
	}

	doc := gltf.NewDocument()
	doc.Asset.Generator = "VOPL -> GLB"
	posAccessor := modeler.WritePosition(doc, positions)
	normalAccessor := modeler.WriteNormal(doc, normals)
	colorAccessor := modeler.WriteColor(doc, colors)
	indicesAccessor := modeler.WriteIndices(doc, indices)
	prim := &gltf.Primitive{
		Attributes: map[string]uint32{
			gltf.POSITION: uint32(posAccessor),
			gltf.NORMAL:   uint32(normalAccessor),
			gltf.COLOR_0:  uint32(colorAccessor),
		},
		Indices: gltf.Index(uint32(indicesAccessor)),
	}
	pbr := &gltf.PBRMetallicRoughness{BaseColorFactor: &[4]float32{1, 1, 1, 1}, MetallicFactor: gltf.Float(0), RoughnessFactor: gltf.Float(1)}
	material := &gltf.Material{PBRMetallicRoughness: pbr}
	if hasAlpha {
		material.AlphaMode = gltf.AlphaBlend
	} else {
		material.AlphaMode = gltf.AlphaOpaque
	}
	doc.Materials = []*gltf.Material{material}
	prim.Material = gltf.Index(0)
	meshGltf := &gltf.Mesh{Name: "ChunkMesh", Primitives: []*gltf.Primitive{prim}}
	doc.Meshes = []*gltf.Mesh{meshGltf}
	node := &gltf.Node{Mesh: gltf.Index(0)}
	doc.Nodes = []*gltf.Node{node}
	doc.Scenes[0].Nodes = append(doc.Scenes[0].Nodes, uint32(0))

	var out bytes.Buffer
	enc := gltf.NewEncoder(&out)
	enc.AsBinary = true
	if err := enc.Encode(doc); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// PackVOPLs builds a .voplpack from provided file blobs and names. All headers must match.
func PackVOPLs(files map[string][]byte) ([]byte, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("no files")
	}
	// parse and validate
	type item struct {
		name    string
		enc     uint8
		payload []byte
		hdr     vopl.VOPLHeader
	}
	items := make([]item, 0, len(files))
	var common vopl.VOPLHeader
	first := true
	for name, data := range files {
		hdr, payload, err := vopl.ParseVOPLHeaderFromBytes(data)
		if err != nil {
			return nil, err
		}
		if hdr.Ver != 3 {
			return nil, fmt.Errorf("apenas VOPL é suportado (%s)", name)
		}
		if first {
			common = hdr
			first = false
		} else if hdr.BPP != common.BPP || hdr.W != common.W || hdr.H != common.H || hdr.D != common.D || hdr.Pal != common.Pal {
			return nil, fmt.Errorf("inconsistent parameters (%s)", name)
		}
		enc := data[5]
		items = append(items, item{name, enc, payload, hdr})
	}
	pack := &vopl.Pack{Header: vopl.VOPLHeader{Ver: 3, BPP: common.BPP, W: common.W, H: common.H, D: common.D, Pal: common.Pal}}
	pack.Entries = make([]vopl.PackEntry, len(items))
	for i, it := range items {
		pack.Entries[i] = vopl.PackEntry{Name: it.name, Enc: it.enc, Payload: it.payload}
	}
	return pack.Marshal(vopl.PackCompZlib)
}

// UnpackVOPLPACKToMemory returns a map of file name -> .vopl bytes from a .voplpack blob.
func UnpackVOPLPACKToMemory(packBytes []byte) (map[string][]byte, error) {
	pack, _, err := vopl.UnmarshalPack(packBytes)
	if err != nil {
		return nil, err
	}
	out := make(map[string][]byte, len(pack.Entries))
	for _, e := range pack.Entries {
		out[e.Name] = vopl.BuildVOPLFromHeaderAndPayload(pack.Header, e.Enc, e.Payload)
	}
	return out, nil
}

// end of file
