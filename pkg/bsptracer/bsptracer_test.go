package bsptracer

import (
	"testing"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/stretchr/testify/assert"
)

type testCast struct {
	from, to mgl32.Vec3
	visible  bool
	trace    Trace
}

var testCasts = []testCast{
	// A site -> A site, open
	{mgl32.Vec3{-12, 1444, 1751}, mgl32.Vec3{-233, 1343, 1751}, true,
		Trace{true, true, 1, 0, mgl32.Vec3{-233, 1343, 1751}, 0, nil, 0}},
	// T spawn -> A site
	{mgl32.Vec3{3306, 431, 1723}, mgl32.Vec3{-233, 1343, 1751}, false,
		Trace{true, true, 0, 1, mgl32.Vec3{3306, 431, 1723}, 1, nil, 7}},
	// T spawn -> T spawn
	{mgl32.Vec3{3306, 431, 1723}, mgl32.Vec3{3303, 431, 1723}, true,
		Trace{true, true, 1, 0, mgl32.Vec3{3303, 431, 1723}, 0, nil, 0}},
	// through door
	{mgl32.Vec3{207, 1948, 1751}, mgl32.Vec3{259, 2251, 1752}, true,
		Trace{true, true, 1, 0, mgl32.Vec3{259, 2251, 1752}, 0, nil, 0}},
	// through mid box
	{mgl32.Vec3{-94, 452, 1677}, mgl32.Vec3{138, 396, 1677}, true,
		Trace{true, true, 1, 0, mgl32.Vec3{138, 396, 1677}, 0, nil, 0}},
}

func TestBspTracer(t *testing.T) {
	// test loading nonexistent file
	m, err := LoadMap("../../testdata", "does_not_exist.bsp")
	assert.Nil(t, m)
	assert.NotNil(t, err)

	// test loading this file (it shouldn't work)
	m, err = LoadMap(".", "map_test.go")
	assert.Nil(t, m)
	assert.NotNil(t, err)

	// test loading good file
	m, err = LoadMap("../../testdata", "de_cache.bsp")
	assert.Nil(t, err)
	if !assert.NotNil(t, m) {
		return
	}
	assert.Equal(t, 5560, len(m.brushes))
	assert.Equal(t, 39815, len(m.brushSides))
	assert.Equal(t, 129415, len(m.edges))
	assert.Equal(t, 24072, len(m.leafBrushes))
	assert.Equal(t, 18843, len(m.leafFaces))
	assert.Equal(t, 8906, len(m.leaves))
	assert.Equal(t, 8648, len(m.nodes))
	assert.Equal(t, 30626, len(m.planes))
	assert.Equal(t, 23221, len(m.surfaces))
	assert.Equal(t, 185200, len(m.surfEdges))
	assert.Equal(t, 48496, len(m.vertices))
	assert.Equal(t, 46442, len(m.polygons))

	// test ray tracing
	for _, cast := range testCasts {
		assert.Equal(t, cast.visible, m.IsVisible(cast.from, cast.to))
		trace := m.TraceRay(cast.from, cast.to)
		trace.Brush = nil // skip comparing this
		assert.Equal(t, cast.trace, *trace)
	}
}
