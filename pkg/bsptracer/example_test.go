package bsptracer_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/go-gl/mathgl/mgl32"

	"github.com/saiko-tech/bsp-tracer/pkg/bsptracer"
)

func ExampleMap_IsVisible() {
	csgoDir := os.Getenv("CSGO_DIR") // should point to "SteamLibrary/steamapps/common/Counter-Strike Global Offensive"

	m, err := bsptracer.LoadMapFromFileSystem(csgoDir+"/csgo/maps/de_cache.bsp", csgoDir+"/csgo/pak01", csgoDir+"/platform/platform_pak01")
	if err != nil {
		panic(err)
	}

	fmt.Println("A site -> A site, open:", m.IsVisible(mgl32.Vec3{-12, 1444, 1751}, mgl32.Vec3{-233, 1343, 1751})) // true
	fmt.Println("T spawn -> A site:", m.IsVisible(mgl32.Vec3{3306, 431, 1723}, mgl32.Vec3{-233, 1343, 1751}))      // false
	fmt.Println("mid through box:", m.IsVisible(mgl32.Vec3{-94, 452, 1677}, mgl32.Vec3{138, 396, 1677}))           // false
	fmt.Println("T spawn -> T spawn:", m.IsVisible(mgl32.Vec3{3306, 431, 1723}, mgl32.Vec3{3300, 400, 1720}))      // true
}

func TestExample(t *testing.T) {
	csgoDir := csgoDir(t)

	t.Setenv("CSGO_DIR", csgoDir)

	ExampleMap_IsVisible()
}
