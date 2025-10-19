package utils

import (
	"math"

	"vopltool/vopl"

	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"
)

func RunVOPL2GLB(inPath, outPath string) error {
	grid, err := vopl.LoadVoplGrid(inPath)
	if err != nil {
		return err
	}

	mesh := vopl.GenerateMesh(grid)

	positions := make([][3]float32, len(mesh.Vertices))
	colors := make([][4]float32, len(mesh.Vertices))
	hasAlpha := false

	for i, v := range mesh.Vertices {
		positions[i] = v.Position
		rgba, err := vopl.ParseHexColor(vopl.Palette[v.Color])
		if err != nil {
			return err
		}
		colors[i] = rgba
		if rgba[3] < 1.0 {
			hasAlpha = true
		}
	}

	indices := make([]uint32, len(mesh.Indices))
	copy(indices, mesh.Indices)

	// flat normals per face
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
	doc.Asset.Generator = "VOPL v3 -> GLB"

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

	pbr := &gltf.PBRMetallicRoughness{
		BaseColorFactor: &[4]float32{1, 1, 1, 1},
		MetallicFactor:  gltf.Float(0),
		RoughnessFactor: gltf.Float(1),
	}
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

	return gltf.SaveBinary(doc, outPath)
}
