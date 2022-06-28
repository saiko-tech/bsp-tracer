package bsptracer

import (
	"os"
	"testing"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/stretchr/testify/assert"
)

func TestBspTracer_NonExisting(t *testing.T) {
	t.Parallel()

	_, err := LoadMapFromFileSystem("../../testdata/does_not_exist.bsp")
	assert.NotNil(t, err)
}

func TestBspTracer_BadFile(t *testing.T) {
	t.Parallel()

	_, err := LoadMapFromFileSystem("./map_test.go")
	assert.NotNil(t, err)
}

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

func TestMap_TraceRay_de_cache(t *testing.T) {
	t.Parallel()

	m, err := LoadMapFromFileSystem("../../testdata/de_cache.bsp")
	assert.Error(t, err)
	assert.ErrorAs(t, err, new(MissingModelsError))

	type args struct {
		origin      mgl32.Vec3
		destination mgl32.Vec3
	}
	type out struct {
		visible bool
		trace   Trace
	}
	tests := []struct {
		name string
		args args
		want out
	}{
		{
			name: "A site -> A site, open",
			args: args{
				origin:      mgl32.Vec3{-12, 1444, 1751},
				destination: mgl32.Vec3{-233, 1343, 1751},
			},
			want: out{
				visible: true,
				trace:   Trace{AllSolid: true, StartSolid: true, Fraction: 1, EndPos: mgl32.Vec3{-233, 1343, 1751}},
			},
		},
		{
			name: "T spawn -> A site",
			args: args{
				mgl32.Vec3{3306, 431, 1723},
				mgl32.Vec3{-233, 1343, 1751},
			},
			want: out{
				visible: false,
				trace:   Trace{AllSolid: true, StartSolid: true, FractionLeftSolid: 1, EndPos: mgl32.Vec3{3306, 431, 1723}, Contents: 1, NumBrushSides: 7},
			},
		},
		{
			name: "T spawn -> T spawn",
			args: args{
				mgl32.Vec3{3306, 431, 1723},
				mgl32.Vec3{3303, 431, 1723},
			},
			want: out{
				visible: true,
				trace:   Trace{AllSolid: true, StartSolid: true, Fraction: 1, EndPos: mgl32.Vec3{3303, 431, 1723}},
			},
		},
		{
			name: "through door",
			args: args{
				mgl32.Vec3{207, 1948, 1751},
				mgl32.Vec3{259, 2251, 1752},
			},
			want: out{
				visible: true,
				trace:   Trace{AllSolid: true, StartSolid: true, Fraction: 1, EndPos: mgl32.Vec3{259, 2251, 1752}},
			},
		},
		{
			name: "through mid box",
			args: args{
				mgl32.Vec3{-94, 452, 1677},
				mgl32.Vec3{138, 396, 1677},
			},
			want: out{
				visible: true,
				trace:   Trace{AllSolid: true, StartSolid: true, Fraction: 1, EndPos: mgl32.Vec3{138, 396, 1677}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want.visible, m.IsVisible(tt.args.origin, tt.args.destination), "IsVisible(%v, %v)", tt.args.origin, tt.args.destination)

			actual := m.TraceRay(tt.args.origin, tt.args.destination)
			actual.Brush = nil // skip comparing this

			assert.Equalf(t, tt.want.trace, *actual, "TraceRay(%v, %v)", tt.args.origin, tt.args.destination)
		})
	}
}

// expects CS:GO to be installed in "$HOME/games/SteamLibrary/steamapps/common/Counter-Strike Global Offensive"
func TestLoadMap_de_cache_with_models(t *testing.T) {
	t.Parallel()

	userHome, err := os.UserHomeDir()
	assert.NoError(t, err)

	csgoDir := userHome + "/games/SteamLibrary/steamapps/common/Counter-Strike Global Offensive"

	_, err = LoadMapFromFileSystem("../../testdata/de_cache.bsp", csgoDir+"/csgo/pak01", csgoDir+"/platform/platform_pak01")
	assert.NoError(t, err)
}
