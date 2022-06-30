package mollertrumbore

import "github.com/go-gl/mathgl/mgl32"

const mollerTrumboreEpsilon = float32(0.0000001)

type RayCastResult struct {
	Hit   bool
	Point mgl32.Vec3
}

// RayIntersectsTriangle determines if a ray intersects a triangle using https://en.wikipedia.org/wiki/M%C3%B6ller%E2%80%93Trumbore_intersection_algorithm
// taken from https://github.com/Galaco/kero/blob/dedc4e04e830cc2597308cbfe9e9bcbe30491fae/physics/collision/ray.go#L143
func RayIntersectsTriangle(rayOrigin mgl32.Vec3, rayVector mgl32.Vec3, inTriangle [3]mgl32.Vec3) (r RayCastResult) {
	vertex0 := inTriangle[0]
	vertex1 := inTriangle[1]
	vertex2 := inTriangle[2]

	var (
		edge1, edge2, h, s, q mgl32.Vec3
		a, f, u, v            float32
	)

	edge1 = vertex1.Sub(vertex0)
	edge2 = vertex2.Sub(vertex0)
	h = rayVector.Cross(edge2)
	a = edge1.Dot(h)

	if a > -mollerTrumboreEpsilon && a < mollerTrumboreEpsilon {
		return r // This ray is parallel to this triangle.
	}

	f = 1.0 / a
	s = rayOrigin.Sub(vertex0)
	u = f * s.Dot(h)

	if u < 0.0 || u > 1.0 {
		return r
	}

	q = s.Cross(edge1)
	v = f * rayVector.Dot(q)

	if v < 0.0 || u+v > 1.0 {
		return r
	}
	// At this stage we can compute t to find out where the intersection point is on the line.
	t := f * edge2.Dot(q)

	if t > mollerTrumboreEpsilon { // ray intersection
		r.Hit = true
		r.Point = rayOrigin.Add(rayVector.Mul(t))

		return r
	}

	// This means that there is a line intersection but not a ray intersection.
	return r
}
