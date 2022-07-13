package bsptracer_test

import (
	"log"
	"os"
	"os/exec"
	"testing"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/stretchr/testify/assert"

	"github.com/saiko-tech/bsp-tracer/pkg/bsptracer"
)

func TestMain(m *testing.M) {
	log.Println("downloading testdata ...")

	err := exec.Command("../../testdata/download.sh").Run()
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(m.Run())
}

// expects CS:GO to be installed in "$HOME/games/SteamLibrary/steamapps/common/Counter-Strike Global Offensive"
func csgoDir(tb testing.TB) string {
	tb.Helper()

	userHome, err := os.UserHomeDir()
	assert.NoError(tb, err)

	return userHome + "/games/SteamLibrary/steamapps/common/Counter-Strike Global Offensive"
}

func TestBspTracer_NonExisting(t *testing.T) {
	t.Parallel()

	_, err := bsptracer.LoadMapFromFileSystem("../../testdata/does_not_exist.bsp")
	assert.NotNil(t, err)
}

func TestBspTracer_BadFile(t *testing.T) {
	t.Parallel()

	_, err := bsptracer.LoadMapFromFileSystem("./map_test.go")
	assert.NotNil(t, err)
}

func TestMap_TraceRay_de_cache(t *testing.T) {
	t.Parallel()

	csgoDir := csgoDir(t)

	m, err := bsptracer.LoadMapFromFileSystem("../../testdata/de_cache.bsp", csgoDir+"/csgo/pak01", csgoDir+"/platform/platform_pak01")
	assert.NoError(t, err)

	type args struct {
		origin      mgl32.Vec3
		destination mgl32.Vec3
	}
	type out struct {
		visible bool
		trace   bsptracer.Trace
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
				trace:   bsptracer.Trace{AllSolid: true, StartSolid: true, Fraction: 1, EndPos: mgl32.Vec3{-233, 1343, 1751}},
			},
		},
		{
			name: "T spawn -> A site",
			args: args{
				origin:      mgl32.Vec3{3306, 431, 1723},
				destination: mgl32.Vec3{-233, 1343, 1751},
			},
			want: out{
				visible: false,
				trace:   bsptracer.Trace{AllSolid: true, StartSolid: true, FractionLeftSolid: 1, EndPos: mgl32.Vec3{3306, 431, 1723}, Contents: 1, NumBrushSides: 7},
			},
		},
		{
			name: "T spawn -> T spawn",
			args: args{
				origin:      mgl32.Vec3{3306, 431, 1723},
				destination: mgl32.Vec3{3303, 431, 1723},
			},
			want: out{
				visible: true,
				trace:   bsptracer.Trace{AllSolid: true, StartSolid: true, Fraction: 1, EndPos: mgl32.Vec3{3303, 431, 1723}},
			},
		},
		{
			name: "through door",
			args: args{
				origin:      mgl32.Vec3{207, 1948, 1751},
				destination: mgl32.Vec3{259, 2251, 1752},
			},
			want: out{
				visible: false,
				trace:   bsptracer.Trace{AllSolid: true, StartSolid: true, Fraction: 1, EndPos: mgl32.Vec3{259, 2251, 1752}},
			},
		},
		{
			name: "through mid box",
			args: args{
				origin:      mgl32.Vec3{-94, 452, 1677},
				destination: mgl32.Vec3{138, 396, 1677},
			},
			want: out{
				visible: false,
				trace:   bsptracer.Trace{AllSolid: true, StartSolid: true, Fraction: 1, EndPos: mgl32.Vec3{138, 396, 1677}},
			},
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want.visible, m.IsVisible(tt.args.origin, tt.args.destination), "IsVisible(%v, %v)", tt.args.origin, tt.args.destination)

			actual := m.TraceRay(tt.args.origin, tt.args.destination)
			actual.Brush = nil // skip comparing this

			assert.Equalf(t, tt.want.trace, *actual, "TraceRay(%v, %v)", tt.args.origin, tt.args.destination)
		})
	}
}

func TestLoadMap_de_cache_with_models(t *testing.T) {
	t.Parallel()

	csgoDir := csgoDir(t)

	_, err := bsptracer.LoadMapFromFileSystem("../../testdata/de_cache.bsp", csgoDir+"/csgo/pak01", csgoDir+"/platform/platform_pak01")
	assert.NoError(t, err)
}

func BenchmarkTraceBox(b *testing.B) {
	csgoDir := csgoDir(b)

	m, err := bsptracer.LoadMapFromFileSystem("../../testdata/de_cache.bsp", csgoDir+"/csgo/pak01", csgoDir+"/platform/platform_pak01")
	assert.NoError(b, err)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		assert.False(b, m.IsVisible(mgl32.Vec3{-94, 452, 1677}, mgl32.Vec3{138, 396, 1677}))
	}
}
