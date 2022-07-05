package collision

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

const mollerTrumboreEpsilon = float32(0.0000001)

type RayCastResult struct {
	T     float64
	Hit   bool
	Point mgl32.Vec3
}

// RayIntersectsAxisAlignedBoundingBox determines whether ray intersects an axis-aligned bounding box.
// taken from https://github.com/Galaco/kero/blob/dedc4e04e830cc2597308cbfe9e9bcbe30491fae/physics/collision/ray.go#L73
func RayIntersectsAxisAlignedBoundingBox(origin, direction, min, max mgl32.Vec3) (r RayCastResult) {
	// Any component of direction could be 0!
	// Address this by using a small number, close to
	// 0 in case any of directions components are 0
	dir := direction
	if dir[0] == 0 {
		dir[0] = 0.00001
	}
	if dir[1] == 0 {
		dir[1] = 0.00001
	}
	if dir[2] == 0 {
		dir[2] = 0.00001
	}

	t1 := float64((min[0] - origin[0]) / dir[0])
	t2 := float64((max[0] - origin[0]) / dir[0])
	t3 := float64((min[1] - origin[1]) / dir[1])
	t4 := float64((max[1] - origin[1]) / dir[1])
	t5 := float64((min[2] - origin[2]) / dir[2])
	t6 := float64((max[2] - origin[2]) / dir[2])

	tmin := math.Max(math.Max(math.Min(t1, t2), math.Min(t3, t4)), math.Min(t5, t6))
	tmax := math.Min(math.Min(math.Max(t1, t2), math.Max(t3, t4)), math.Max(t5, t6))

	// if tmax < 0, ray is intersecting AABB
	// but entire AABB is behing it's origin
	if tmax < 0 {
		return r
	}

	// if tmin > tmax, ray doesn't intersect AABB
	if tmin > tmax {
		return r
	}

	t_result := tmin

	// If tmin is < 0, tmax is closer
	if tmin < 0.0 {
		t_result = tmax
	}

	r.Hit = true
	r.T = t_result
	r.Point = origin.Add(direction).Mul(float32(t_result))

	return r
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
