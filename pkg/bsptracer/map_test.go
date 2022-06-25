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
	// TODO test for same object totals as C++ library on same file
}
