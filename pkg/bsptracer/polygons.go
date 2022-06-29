package bsptracer

import (
	"github.com/galaco/bsp"
	"github.com/galaco/bsp/lumps"
	"github.com/go-gl/mathgl/mgl32"
)

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

func (v *vplane) dist(destination mgl32.Vec3) float32 {
	return v.origin.Dot(destination) - v.distance
}

func buildPolygons(bspfile *bsp.Bsp) []polygon {
	surfaces := bspfile.Lump(bsp.LumpFaces).(*lumps.Face).GetData()
	surfEdges := bspfile.Lump(bsp.LumpSurfEdges).(*lumps.Surfedge).GetData()
	vertices := bspfile.Lump(bsp.LumpVertexes).(*lumps.Vertex).GetData()
	edges := bspfile.Lump(bsp.LumpEdges).(*lumps.Edge).GetData()
	planes := bspfile.Lump(bsp.LumpPlanes).(*lumps.Planes).GetData()

	polygons := make([]polygon, len(surfaces), 2*len(surfaces))

	for _, surface := range surfaces {
		firstEdge := int(surface.FirstEdge)
		numEdges := int(surface.NumEdges)

		if numEdges < 3 || numEdges > maxSurfinfoVerts || surface.TexInfo <= 0 {
			continue
		}

		var (
			poly polygon
			edge mgl32.Vec3
		)

		for i := 0; i < numEdges; i++ {
			edgeIndex := surfEdges[firstEdge+i]
			if edgeIndex >= 0 {
				edge = vertices[edges[edgeIndex][0]]
			} else {
				edge = vertices[edges[-edgeIndex][1]]
			}

			poly.verts[i] = edge
		}

		poly.numVerts = numEdges
		poly.plane.origin = planes[surface.Planenum].Normal
		poly.plane.distance = planes[surface.Planenum].Distance
		polygons = append(polygons, poly)
	}

	return polygons
}
