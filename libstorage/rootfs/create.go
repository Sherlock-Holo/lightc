package rootfs

import (
	"os"
	"path/filepath"

	"github.com/Sherlock-Holo/lightc/libstorage/errors"
	"github.com/Sherlock-Holo/lightc/paths"
	"golang.org/x/xerrors"
)

func Create(imageName string) (*RootFS, error) {
	id := generateInfoID()

	stat, err := os.Stat(filepath.Join(paths.ImagesPath, imageName))
	switch {
	case os.IsNotExist(err):
		return nil, errors.ImageNotFound{NotFoundName: imageName}

	default:
		return nil, xerrors.Errorf("get image stat failed: %w", err)

	case err == nil:
		if !stat.IsDir() {
			return nil, xerrors.Errorf("%s is an invalid file: %w", imageName, err)
		}
	}

	_, err = os.Stat(filepath.Join(paths.RootFSPath, id))
	switch {
	case err == nil:
		return nil, errors.RootFSCreateConflict{ConflictID: id}

	default:
		return nil, xerrors.Errorf("get rootfs stat failed: %w", err)

	case os.IsNotExist(err):
	}

	rootFS := &RootFS{
		ID:        id,
		ImageName: imageName,
		LowerDir:  filepath.Join(paths.ImagesPath, imageName),
		UpperDir:  filepath.Join(paths.RootFSPath, id, paths.UpperFile),
		WorkDir:   filepath.Join(paths.RootFSPath, id, paths.WorkFile),
		MergedDir: filepath.Join(paths.RootFSPath, id, paths.MergedFile),
	}

	for _, dir := range []string{rootFS.UpperDir, rootFS.WorkDir, rootFS.MergedDir} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, xerrors.Errorf("mkdir overlay2 dir failed: %w", err)
		}
	}

	return rootFS, nil
}
