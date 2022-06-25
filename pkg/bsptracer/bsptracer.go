// Package bsptracer implements Ray-Tracing / Ray-Casting on top of github.com/Galaco/bsp.
// This is a port of https://github.com/ReactiioN1337/valve-bsp-parser.
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

const maxSurfinfoVerts = 32

// Map is a loaded BSP map.
type Map struct {
	// loaded by bsp package
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

	// constructed by this package
	polygons []polygon
}

// LoadMap loads a BSP map from a file.
func LoadMap(directory, mapName string) (*Map, error) {
	bspfile, err := bsp.ReadFromFile(filepath.Join(directory, mapName))
	if err != nil {
		return nil, err
	}
	m := &Map{
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
	}
	m.parsePolygons()
	return m, nil
}

func (m *Map) parsePolygons() {
	m.polygons = make([]polygon, len(m.surfaces), 2*len(m.surfaces))

	for _, surface := range m.surfaces {
		firstEdge := int(surface.FirstEdge)
		numEdges := int(surface.NumEdges)

		if numEdges < 3 || numEdges > maxSurfinfoVerts {
			continue
		}
		if surface.TexInfo <= 0 {
			continue
		}

		var polygon polygon
		var edge mgl32.Vec3

		for i := 0; i < numEdges; i++ {
			edgeIndex := m.surfEdges[firstEdge+i]
			if edgeIndex >= 0 {
				edge = m.vertices[m.edges[edgeIndex][0]]
			} else {
				edge = m.vertices[m.edges[-edgeIndex][1]]
			}
			polygon.verts[i] = edge
		}

		polygon.numVerts = numEdges
		polygon.plane.origin = m.planes[surface.Planenum].Normal
		polygon.plane.distance = m.planes[surface.Planenum].Distance
		m.polygons = append(m.polygons, polygon)
	}
}

type polygon struct {
	verts      [maxSurfinfoVerts]mgl32.Vec3
	numVerts   int
	plane      vplane
	edgePlanes []vplane
	vec2d      [maxSurfinfoVerts]mgl32.Vec3
	skip       int
}

type vplane struct {
	origin   mgl32.Vec3
	distance float32
}
