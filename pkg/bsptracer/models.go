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
	open(string) (io.ReadCloser, error)
}

func loadModelPart[T any](fs virtualFileSystem, filePath string, reader func(io.Reader) (T, error)) (T, error) {
	var def T

	f, err := fs.open(filePath)
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

func loadModel(fs virtualFileSystem, filePath string) (*studiomodel.StudioModel, error) {
	prop := strings.Split(filePath, ".mdl")[0]

	mdlData, err := loadModelPart(fs, prop+".mdl", mdl.ReadFromStream)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read mdl")
	}

	vvdData, err := loadModelPart(fs, prop+".vvd", vvd.ReadFromStream)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read vvd")
	}

	vtxData, err := loadModelPart(fs, prop+".dx90.vtx", vtx.ReadFromStream)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read vtx")
	}

	phyData, err := loadModelPart(fs, prop+".phy", phy.ReadFromStream)
	if err != nil && !errors.Is(err, errFileNotFound) { // .phy is ok to be missing, it's optional
		return nil, errors.Wrap(err, "failed to read phy")
	}

	return &studiomodel.StudioModel{
		Filename: prop,
		Mdl:      mdlData,
		Vvd:      vvdData,
		Vtx:      vtxData,
		Phy:      phyData,
	}, nil
}

type MissingModelsError struct {
	missingModels []string
}

func (m MissingModelsError) Error() string {
	return fmt.Sprintf(`missing models: ("%s")`, strings.Join(m.missingModels, `", "`))
}

func loadModels(bspfile *bsp.Bsp, vpks []*vpk.VPK) ([]*studiomodel.StudioModel, error) {
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
		prop, err := loadModel(fs, model)
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
