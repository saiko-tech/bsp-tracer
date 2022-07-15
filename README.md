# bsp-tracer

BSP (Source Engine Map) Ray Tracer / Ray Caster Library.

Allows to do static / out-of-engine visibility checks and ray casting on BSP map files.<br>
Can be used to get more accurate visibility info between players than there is available in CS:GO demos/replays (.dem files).

## Features

- [x] Faces (basic map geometry)
- [x] Brushes (walls / level shape)
- [x] Static Props (boxes, barrels, etc.)
  - [x] Orientation / Angle (currently all props are placed at 0 degrees)
- [ ] Displacements (terrain bumps and slopes)
- [ ] Entities ("dynamic" props - doors, vents, etc.)

## Example

```go
package main

import (
	"fmt"
	"os"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/saiko-tech/bsp-tracer/pkg/bsptracer"
)

func main() {
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
```

## Development

### Linting

Uses [golangci-lint](https://golangci-lint.run/)

### Project Layout

Follows https://github.com/golang-standards/project-layout

## Acknowledgements

- This library is based on the C++ [valve-bsp-parser](https://github.com/ReactiioN1337/valve-bsp-parser) by [@ReactiioN1337](https://github.com/ReactiioN1337)
- Thanks to [@jangler](https://github.com/jangler) for porting the C++ library to Go
- Thanks to [@Galaco](https://github.com/Galaco) for creating Go tooling for the BSP file format & source engine
