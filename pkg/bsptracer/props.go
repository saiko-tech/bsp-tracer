package bsptracer

import (
	"fmt"
	"io"
	"strings"

	"github.com/galaco/bsp"
	"github.com/galaco/bsp/lumps"
	"github.com/galaco/studiomodel"
	"github.com/galaco/studiomodel/mdl"
	"github.com/galaco/studiomodel/phy"
	"github.com/galaco/studiomodel/vtx"
	"github.com/galaco/studiomodel/vvd"
	vpk "github.com/galaco/vpk2"
	"github.com/pkg/errors"
)

type virtualFileSystem interface {
	GetFile(string) (io.ReadCloser, error)
}

func loadPropPart[T any](fs virtualFileSystem, filePath string, reader func(io.Reader) (T, error)) (T, error) {
	var def T

	f, err := fs.GetFile(filePath)
	if err != nil {
		return def, errors.Wrapf(err, "failed to open prop part file %q", filePath)
	}

	defer f.Close()

	part, err := reader(f)
	if err != nil {
		return def, errors.Wrapf(err, "failed to read prop part from %q", filePath)
	}

	return part, nil
}

func loadProp(fs virtualFileSystem, filePath string) (*studiomodel.StudioModel, error) {
	mdlData, err := loadPropPart(fs, filePath+".mdl", mdl.ReadFromStream)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read mdl")
	}

	vvdData, err := loadPropPart(fs, filePath+".vvd", vvd.ReadFromStream)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read vvd")
	}

	vtxData, err := loadPropPart(fs, filePath+".dx90.vtx", vtx.ReadFromStream)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read vtx")
	}

	phyData, err := loadPropPart(fs, filePath+".phy", phy.ReadFromStream)
	if err != nil && !errors.Is(err, errFileNotFound) {
		return nil, errors.Wrap(err, "failed to read phy")
	}

	return &studiomodel.StudioModel{
		Filename: filePath,
		Mdl:      mdlData,
		Vvd:      vvdData,
		Vtx:      vtxData,
		Phy:      phyData,
	}, nil
}

// LoadProp loads a single prop/model of known filepath
func LoadProp(fs virtualFileSystem, path string) (*studiomodel.StudioModel, error) {
	prop, err := loadProp(fs, strings.Split(path, ".mdl")[0])
	if err != nil {
		return nil, errors.Wrap(err, "failed to read studiomodel")
	}

	return prop, nil
}

type MissingModelsError struct {
	missingModels []string
}

func (m MissingModelsError) Error() string {
	return fmt.Sprintf(`missing models: ("%s")`, strings.Join(m.missingModels, `", "`))
}

func loadProps(bspfile *bsp.Bsp, vpks []*vpk.VPK) ([]*studiomodel.StudioModel, error) {
	fs := vfs{
		pakfile: bspfile.Lump(bsp.LumpPakfile).(*lumps.Pakfile).GetData(),
		vpks:    vpks,
	}

	var (
		props         []*studiomodel.StudioModel
		missingModels []string
	)

	gameLump := bspfile.Lump(bsp.LumpGame).(*lumps.Game).GetData()

	for _, model := range gameLump.GetStaticPropLump().DictLump.Name {
		prop, err := LoadProp(fs, model)
		if err != nil {
			missingModels = append(missingModels, model)

			props = append(props, nil)

			continue
		}

		props = append(props, prop)
	}

	if len(missingModels) > 0 {
		return props, MissingModelsError{
			missingModels: missingModels,
		}
	}

	return props, nil
}
