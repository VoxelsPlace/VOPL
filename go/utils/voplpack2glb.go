package utils

import (
	"fmt"
	"math"
	"os"
	"path/filepath"

	"github.com/voxelsplace/vopl/go/vopl"

	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"
)

// RunVOPLPACK2GLB converts a .voplpack into a .glb.
// It creates one glTF Mesh/Node per entry in the pack and puts them all in a single scene.
func RunVOPLPACK2GLB(inPackPath, outGlbPath string) error {
	data, err := os.ReadFile(inPackPath)
	if err != nil {
		return err
	}
	pack, _, err := vopl.UnmarshalPack(data)
	if err != nil {
		return err
	}

	doc := gltf.NewDocument()
	doc.Asset.Generator = "VOPLPACK -> GLB"

	// Use a single default material; colors come from per-vertex COLOR_0 attribute.
	pbr := &gltf.PBRMetallicRoughness{BaseColorFactor: &[4]float32{1, 1, 1, 1}, MetallicFactor: gltf.Float(0), RoughnessFactor: gltf.Float(1)}
	material := &gltf.Material{PBRMetallicRoughness: pbr, AlphaMode: gltf.AlphaOpaque}
	doc.Materials = []*gltf.Material{material}

	// Arrange entries in a grid so they don't overlap
	n := len(pack.Entries)
	if n == 0 {
		return fmt.Errorf("vazio: nenhum entry no pack")
	}
	cols := int(math.Ceil(math.Sqrt(float64(n))))
	// No extra gap between models: place them exactly side-by-side
	stepX := float32(pack.Header.W)
	stepZ := float32(pack.Header.D)

	// For each entry: rebuild full .vopl bytes, parse grid, mesh it, write buffers.
	for i, e := range pack.Entries {
		voplBytes := vopl.BuildVOPLFromHeaderAndPayload(pack.Header, e.Enc, e.Payload)
		grid, err := vopl.LoadVoplGridFromBytes(voplBytes)
		if err != nil {
			return fmt.Errorf("entry %d (%s): %w", i, e.Name, err)
		}
		mesh := vopl.GenerateMesh(grid)

		positions := make([][3]float32, len(mesh.Vertices))
		colors := make([][4]float32, len(mesh.Vertices))
		for vi, v := range mesh.Vertices {
			positions[vi] = v.Position
			rgba, err := vopl.ParseHexColor(vopl.Palette[v.Color])
			if err != nil {
				return fmt.Errorf("entry %d (%s): %w", i, e.Name, err)
			}
			colors[vi] = rgba
		}
		indices := make([]uint32, len(mesh.Indices))
		copy(indices, mesh.Indices)

		posAccessor := modeler.WritePosition(doc, positions)
		// Note: Normals are omitted here for simplicity; viewers typically compute them.
		colorAccessor := modeler.WriteColor(doc, colors)
		indicesAccessor := modeler.WriteIndices(doc, indices)

		prim := &gltf.Primitive{
			Attributes: map[string]uint32{
				gltf.POSITION: uint32(posAccessor),
				gltf.COLOR_0:  uint32(colorAccessor),
			},
			Indices:  gltf.Index(uint32(indicesAccessor)),
			Material: gltf.Index(0),
		}

		m := &gltf.Mesh{Name: filepath.Base(e.Name), Primitives: []*gltf.Primitive{prim}}
		doc.Meshes = append(doc.Meshes, m)
		// Compute grid placement
		r := i / cols
		c := i % cols
		tx := float32(c) * stepX
		tz := float32(r) * stepZ
		node := &gltf.Node{Name: m.Name, Mesh: gltf.Index(uint32(len(doc.Meshes) - 1))}
		node.Translation = [3]float32{tx, 0, tz}
		doc.Nodes = append(doc.Nodes, node)
		doc.Scenes[0].Nodes = append(doc.Scenes[0].Nodes, uint32(len(doc.Nodes)-1))
	}

	return gltf.SaveBinary(doc, outGlbPath)
}
