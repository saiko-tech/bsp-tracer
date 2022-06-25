package bsptracer

import (
	"path/filepath"

	"github.com/galaco/bsp"
	"github.com/galaco/bsp/lumps"
	"github.com/galaco/bsp/primitives/brush"
	"github.com/galaco/bsp/primitives/brushside"
	"github.com/galaco/bsp/primitives/face"
	"github.com/galaco/bsp/primitives/leaf"
	"github.com/galaco/bsp/primitives/node"
	"github.com/galaco/bsp/primitives/plane"
	"github.com/go-gl/mathgl/mgl32"
)

// Map is a loaded BSP map.
type Map struct {
	brushes     []brush.Brush
	brushSides  []brushside.BrushSide
	edges       [][2]uint16
	leafBrushes []uint16
	leafFaces   []uint16
	leaves      []leaf.Leaf
	nodes       []node.Node
	planes      []plane.Plane
	surfaces    []face.Face
	surfEdges   []int32
	vertices    []mgl32.Vec3
}

// LoadMap loads a BSP map from a file.
func LoadMap(directory, mapName string) (*Map, error) {
	bspfile, err := bsp.ReadFromFile(filepath.Join(directory, mapName))
	if err != nil {
		return nil, err
	}
	return &Map{
		brushes:     bspfile.Lump(bsp.LumpBrushes).(*lumps.Brush).GetData(),
		brushSides:  bspfile.Lump(bsp.LumpBrushSides).(*lumps.BrushSide).GetData(),
		edges:       bspfile.Lump(bsp.LumpEdges).(*lumps.Edge).GetData(),
		leafBrushes: bspfile.Lump(bsp.LumpLeafBrushes).(*lumps.LeafBrush).GetData(),
		leafFaces:   bspfile.Lump(bsp.LumpLeafFaces).(*lumps.LeafFace).GetData(),
		leaves:      bspfile.Lump(bsp.LumpLeafs).(*lumps.Leaf).GetData(),
		nodes:       bspfile.Lump(bsp.LumpNodes).(*lumps.Node).GetData(),
		planes:      bspfile.Lump(bsp.LumpPlanes).(*lumps.Planes).GetData(),
		surfaces:    bspfile.Lump(bsp.LumpFaces).(*lumps.Face).GetData(),
		surfEdges:   bspfile.Lump(bsp.LumpSurfEdges).(*lumps.Surfedge).GetData(),
		vertices:    bspfile.Lump(bsp.LumpVertexes).(*lumps.Vertex).GetData(),
	}, nil
}
