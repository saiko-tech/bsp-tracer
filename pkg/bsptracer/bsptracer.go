// Package bsptracer implements Ray-Tracing / Ray-Casting on top of github.com/Galaco/bsp.
// This is a port of https://github.com/ReactiioN1337/valve-bsp-parser.
package bsptracer

import (
	"github.com/galaco/bsp"
	"github.com/galaco/bsp/lumps"
	"github.com/galaco/bsp/primitives/brush"
	"github.com/galaco/bsp/primitives/brushside"
	"github.com/galaco/bsp/primitives/dispinfo"
	"github.com/galaco/bsp/primitives/disptris"
	"github.com/galaco/bsp/primitives/dispvert"
	"github.com/galaco/bsp/primitives/face"
	"github.com/galaco/bsp/primitives/leaf"
	"github.com/galaco/bsp/primitives/node"
	"github.com/galaco/bsp/primitives/plane"
	"github.com/galaco/studiomodel"
	"github.com/galaco/vpk2"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/pkg/errors"
)

const (
	distEpsilon      = float32(0.03125)
	maxSurfinfoVerts = 32
)

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
	game        *lumps.Game         // TODO: may be needed for props + leaves? or maybe not ...
	dispInfo    []dispinfo.DispInfo // TODO: trace against displacements
	dispVerts   []dispvert.DispVert
	dispTris    []disptris.DispTri

	// constructed by this package
	entities []map[string]string // TODO: not yet sure if we'll need this
	polygons []polygon
	props    []*studiomodel.StudioModel // TODO: place props in the world and trace against them
}

// LoadMap loads a map from a BSP file and VPKs.
// May return MissingModelsError if models can't be found - this is not fatal and the map can still be used.
func LoadMap(bspfile *bsp.Bsp, vpks ...*vpk.VPK) (Map, error) {
	entitiesStr := bspfile.Lump(bsp.LumpEntities).(*lumps.EntData).GetData()
	entities := parseEntities(entitiesStr)

	polygons := parsePolygons(bspfile)

	props, missingModelsErr := loadProps(bspfile, vpks)

	m := Map{
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
		game:        bspfile.Lump(bsp.LumpGame).(*lumps.Game).GetData(),
		dispInfo:    bspfile.Lump(bsp.LumpDispInfo).(*lumps.DispInfo).GetData(),
		dispVerts:   bspfile.Lump(bsp.LumpDispVerts).(*lumps.DispVert).GetData(),
		dispTris:    bspfile.Lump(bsp.LumpDispTris).(*lumps.DispTris).GetData(),
		entities:    entities,
		polygons:    polygons,
		props:       props,
	}

	if missingModelsErr != nil {
		return m, missingModelsErr
	}

	return m, nil
}

// LoadMapFromFileSystem loads a BSP map from the file system.
// vpkPaths is a list of paths to either single or multi VPKs to load models, in order of priority.
// for CS:GO, vpkPaths should be paths to ("SteamLibrary/steamapps/common/Counter-Strike Global Offensive/csgo/pak01", "SteamLibrary/steamapps/common/Counter-Strike Global Offensive/platform/platform_pak01")
// See also LoadMap()
func LoadMapFromFileSystem(mapPath string, vpkPaths ...string) (Map, error) {
	bspfile, err := bsp.ReadFromFile(mapPath)
	if err != nil {
		return Map{}, err
	}

	vpks := make([]*vpk.VPK, len(vpkPaths))

	for i, path := range vpkPaths {
		var err error

		vpks[i], err = vpk.Open(vpk.MultiVPK(path))
		if err != nil {
			vpks[i], err = vpk.Open(vpk.SingleVPK(path))
			if err != nil {
				return Map{}, errors.Wrapf(err, "failed to open vpk %q", path)
			}
		}
	}

	return LoadMap(bspfile, vpks...)
}

// IsVisible returns true if destination is visible from origin, as computed by
// a ray trace.
func (m Map) IsVisible(origin, destination mgl32.Vec3) bool {
	return m.TraceRay(origin, destination).Fraction >= 1
}

// Trace captures the result of a ray trace.
type Trace struct {
	AllSolid          bool
	StartSolid        bool
	Fraction          float32
	FractionLeftSolid float32
	EndPos            mgl32.Vec3
	Contents          int32
	Brush             *brush.Brush
	NumBrushSides     int32
}

// TraceRay traces a ray from origin to destination and returns the result.
func (m Map) TraceRay(origin, destination mgl32.Vec3) *Trace {
	out := &Trace{
		AllSolid:   true,
		StartSolid: true,
		Fraction:   1,
	}

	m.rayCastNode(0, 0, 1, origin, destination, out)

	if out.Fraction < 1 {
		for i := 0; i < 3; i++ {
			out.EndPos[i] = origin[i] + out.Fraction*(destination[i]-origin[i])
		}
	} else {
		out.EndPos = destination
	}

	return out
}

func (m Map) rayCastNode(nodeIndex int32, startFraction, endFraction float32,
	origin, destination mgl32.Vec3, out *Trace,
) {
	if out.Fraction <= startFraction {
		return
	}

	if nodeIndex < 0 {
		leaf := m.leaves[-nodeIndex-1]

		for i := uint16(0); i < leaf.NumLeafBrushes; i++ {
			brushIndex := m.leafBrushes[leaf.FirstLeafBrush+i]
			brush := &m.brushes[brushIndex]

			if brush.Contents&bsp.MASK_SHOT_HULL == 0 {
				continue
			}

			m.rayCastBrush(brush, origin, destination, out)

			if out.Fraction == 0 {
				return
			}

			out.Brush = brush
		}

		if out.StartSolid || out.Fraction < 1 {
			return
		}

		for i := uint16(0); i < leaf.NumLeafFaces; i++ {
			m.rayCastSurface(int(m.leafFaces[leaf.FirstLeafFace+i]),
				origin, destination, out)
		}

		// TODO: handle leaf props
		// TODO: handle leaf displacements

		return
	}

	node := m.nodes[nodeIndex]
	plane := m.planes[node.PlaneNum]

	var startDistance, endDistance float32

	if plane.AxisType < 3 {
		startDistance = origin[plane.AxisType] - plane.Distance
		endDistance = destination[plane.AxisType] - plane.Distance
	} else {
		startDistance = origin.Dot(plane.Normal) - plane.Distance
		endDistance = destination.Dot(plane.Normal) - plane.Distance
	}

	if startDistance >= 0 && endDistance >= 0 {
		m.rayCastNode(node.Children[0], startFraction, endFraction, origin, destination, out)
	} else if startDistance < 0 && endDistance < 0 {
		m.rayCastNode(node.Children[1], startFraction, endFraction, origin, destination, out)
	} else {
		var (
			sideID                        uint
			fractionFirst, fractionSecond float32
			middle                        mgl32.Vec3
		)

		if startDistance < endDistance {
			// back
			sideID = 1
			inversedDistance := 1 / (startDistance - endDistance)

			fractionFirst = (startDistance + mgl32.Epsilon) * inversedDistance
			fractionSecond = fractionFirst
		} else if endDistance < startDistance {
			// front
			sideID = 0
			inversedDistance := 1 / (startDistance - endDistance)

			fractionFirst = (startDistance + mgl32.Epsilon) * inversedDistance
			fractionSecond = (startDistance - mgl32.Epsilon) * inversedDistance
		} else {
			// front
			sideID = 0
			fractionFirst = 1
			fractionSecond = 0
		}
		if fractionFirst < 0 {
			fractionFirst = 0
		} else if fractionFirst > 1 {
			fractionFirst = 1
		}
		if fractionSecond < 0 {
			fractionSecond = 0
		} else if fractionSecond > 1 {
			fractionSecond = 1
		}

		fractionMiddle := startFraction + (endFraction-startFraction)*fractionFirst
		for i := 0; i < 3; i++ {
			middle[i] = origin[i] + fractionFirst*(destination[i]-origin[i])
		}

		m.rayCastNode(node.Children[sideID],
			startFraction, fractionMiddle, origin, middle, out)
		for i := 0; i < 3; i++ {
			middle[i] = origin[i] + fractionSecond*(destination[i]-origin[i])
		}

		m.rayCastNode(node.Children[(^sideID)&1],
			fractionMiddle, endFraction, middle, destination, out)
	}
}

func (m Map) rayCastBrush(brush *brush.Brush, origin, destination mgl32.Vec3, out *Trace) {
	if brush.NumSides != 0 {
		fractionToEnter := float32(-99)
		fractionToLeave := float32(1)
		startsOut := false
		endsOut := false

		for i := int32(0); i < brush.NumSides; i++ {
			brushSide := m.brushSides[brush.FirstSide+i]
			if brushSide.Bevel&0xff != 0 {
				continue
			}

			plane := m.planes[brushSide.PlaneNum]

			startDistance := origin.Dot(plane.Normal) - plane.Distance
			endDistance := destination.Dot(plane.Normal) - plane.Distance

			if startDistance > 0 {
				startsOut = true

				if endDistance > 0 {
					return
				}
			} else {
				if endDistance <= 0 {
					continue
				}
				endsOut = true
			}

			if startDistance > endDistance {
				fraction := startDistance - distEpsilon
				if fraction < 0 {
					fraction = 0
				}

				if fraction > fractionToEnter {
					fractionToEnter = fraction
				}
			} else {
				fraction := (startDistance + distEpsilon) / (startDistance - endDistance)
				if fraction < fractionToLeave {
					fractionToLeave = fraction
				}
			}
		}

		if startsOut && out.FractionLeftSolid-fractionToEnter > 0 {
			startsOut = false
		}

		out.NumBrushSides = brush.NumSides

		if !startsOut {
			out.StartSolid = true
			out.Contents = brush.Contents

			if !endsOut {
				out.AllSolid = true
				out.Fraction = 0
				out.FractionLeftSolid = 1
			} else if fractionToLeave != 1 && fractionToLeave > out.FractionLeftSolid {
				out.FractionLeftSolid = fractionToLeave

				if out.Fraction <= fractionToLeave {
					out.Fraction = 1
				}
			}

			return
		}

		if fractionToEnter < fractionToLeave {
			if fractionToEnter > -99 && fractionToEnter < out.Fraction {
				if fractionToEnter < 0 {
					fractionToEnter = 0
				}

				out.Fraction = fractionToEnter
				out.Brush = brush
				out.Contents = brush.Contents
			}
		}
	}
}

func (m Map) rayCastSurface(index int, origin, destination mgl32.Vec3, out *Trace) {
	if index >= len(m.polygons) {
		return
	}

	polygon := m.polygons[index]
	plane := polygon.plane
	dot1 := plane.dist(origin)
	dot2 := plane.dist(destination)

	if (dot1 > 0) != (dot2 > 0) {
		if dot1-dot2 < distEpsilon {
			return
		}

		t := dot1 / (dot1 - dot2)
		if t <= 0 {
			return
		}

		i := 0
		intersection := origin.Add(destination.Sub(origin).Mul(t))

		for ; i < polygon.numVerts; i++ {
			edgePlane := polygon.edgePlanes[i]
			if edgePlane.origin.Len() == 0 {
				edgePlane.origin = plane.origin.Sub(
					polygon.verts[i].Sub(polygon.verts[(i+1)%polygon.numVerts]))
				edgePlane.origin.Normalize()
				edgePlane.distance = edgePlane.origin.Dot(polygon.verts[i])
			}

			if edgePlane.dist(intersection) < 0 {
				break
			}
		}

		if i == polygon.numVerts {
			out.Fraction = 0.2
			out.EndPos = intersection
		}
	}
}
