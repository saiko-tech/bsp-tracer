package bsptracer

import (
	"archive/zip"
	"io"
	"strings"

	"github.com/galaco/vpk2"
	"github.com/pkg/errors"
)

type vfs struct {
	pakfile *zip.Reader
	vpks    []*vpk.VPK

	pakfileIndex map[string]*zip.File
}

var errFileNotFound = errors.New("file not found")

func (v vfs) open(path string) (io.ReadCloser, error) {
	f, err := v.pakfile.Open(path)
	if err == nil {
		stat, err := f.Stat()
		if err == nil && stat.Size() > 0 {
			return f, nil
		}
	}

	// try case-insensitive
	if v.pakfileIndex == nil {
		v.pakfileIndex = make(map[string]*zip.File)

		for _, f := range v.pakfile.File {
			v.pakfileIndex[strings.ToLower(f.Name)] = f
		}
	}

	pakF, ok := v.pakfileIndex[strings.ToLower(path)]
	if ok {
		f, err := pakF.Open()
		if err == nil {
			return f, nil
		}
	}

	// try vpk
	for _, vpkF := range v.vpks {
		f, err = vpkF.Open(path)
		if err == nil {
			stat, err := f.Stat()
			if err == nil && stat.Size() > 0 {
				return f, nil
			}
		}
	}

	return nil, errors.Wrapf(errFileNotFound, "%s not found", path)
}
