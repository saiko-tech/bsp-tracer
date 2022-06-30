package bsptracer

import (
	"github.com/galaco/bsp"
	"github.com/galaco/bsp/lumps"
	"github.com/galaco/bsp/primitives/game"
	"github.com/galaco/studiomodel"
	"github.com/galaco/studiomodel/mdl"
	"github.com/galaco/studiomodel/phy"
	"github.com/go-gl/mathgl/mgl32"
)

type staticProp struct {
	prop      game.IStaticPropDataLump
	model     *studiomodel.StudioModel
	triangles [][3]mgl32.Vec3
}

func vectorITransform(in1 mgl32.Vec3, in2 mgl32.Mat3x4) (out mgl32.Vec3) {
	t := mgl32.Vec3{}
	t[0] = in1[0] - in2.Col(3)[0]
	t[1] = in1[1] - in2.Col(3)[1]
	t[2] = in1[2] - in2.Col(3)[2]

	out[0] = t[0]*in2.Col(0)[0] + t[1]*in2.Col(0)[1] + t[2]*in2.Col(0)[2]
	out[1] = t[0]*in2.Col(1)[0] + t[1]*in2.Col(1)[1] + t[2]*in2.Col(1)[2]
	out[2] = t[0]*in2.Col(2)[0] + t[1]*in2.Col(2)[1] + t[2]*in2.Col(2)[2]

	return out
}

func transformPhyVertex(bone *mdl.Bone, vertex mgl32.Vec3) (out mgl32.Vec3) {
	out[0] = 1 / 0.0254 * vertex[0]
	out[1] = 1 / 0.0254 * vertex[2]
	out[2] = 1 / 0.0254 * -vertex[1]

	if bone != nil {
		out = vectorITransform(out, bone.PoseToBone)
	} else {
		out[0] = 1 / 0.0254 * vertex[2]
		out[1] = 1 / 0.0254 * -vertex[0]
		out[2] = 1 / 0.0254 * -vertex[1]
	}
	return out
}

func triangles(prop game.IStaticPropDataLump, phy *phy.Phy) [][3]mgl32.Vec3 {
	if phy == nil {
		return nil
	}

	out := make([][3]mgl32.Vec3, len(phy.TriangleFaces))

	for i, t := range phy.TriangleFaces {
		a := prop.GetOrigin().Add(transformPhyVertex(nil, phy.Vertices[t.V1].Vec3()))
		b := prop.GetOrigin().Add(transformPhyVertex(nil, phy.Vertices[t.V2].Vec3()))
		c := prop.GetOrigin().Add(transformPhyVertex(nil, phy.Vertices[t.V3].Vec3()))

		out[i] = [3]mgl32.Vec3{a, b, c}
	}

	return out
}

func staticPropsByLeaf(bspfile *bsp.Bsp, models []*studiomodel.StudioModel) map[uint16][]staticProp {
	res := make(map[uint16][]staticProp)

	gameLump := bspfile.Lump(bsp.LumpGame).(*lumps.Game).GetData()
	spLump := gameLump.GetStaticPropLump()

	for _, p := range spLump.PropLumps {
		leafIndices := spLump.LeafLump.Leaf[p.GetFirstLeaf() : p.GetFirstLeaf()+p.GetLeafCount()]

		for _, i := range leafIndices {
			model := models[p.GetPropType()]

			var tris [][3]mgl32.Vec3

			// missing model
			if model != nil {
				tris = triangles(p, model.Phy)
			}

			res[i] = append(res[i], staticProp{
				prop:      p,
				model:     model,
				triangles: tris,
			})
		}
	}

	return res
}
