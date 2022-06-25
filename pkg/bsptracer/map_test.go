package bsptracer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadMap(t *testing.T) {
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
}
