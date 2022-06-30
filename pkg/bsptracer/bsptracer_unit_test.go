package bsptracer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadMap_de_cache(t *testing.T) {
	t.Parallel()

	m, err := LoadMapFromFileSystem("../../testdata/de_cache.bsp")
	assert.Error(t, err)
	assert.ErrorAs(t, err, new(MissingModelsError))

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
}
