package images

import (
	"os"
	"path/filepath"

	"github.com/Sherlock-Holo/lightc/libstorage/errors"
	"github.com/Sherlock-Holo/lightc/paths"
	"golang.org/x/xerrors"
)

func Delete(imageName string) error {
	path := filepath.Join(paths.ImagesPath, imageName)

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return errors.ImageNotFound{NotFoundName: imageName}
		}
		return xerrors.Errorf("get image %s dir stat failed: %w", imageName, err)
	}

	if err := os.RemoveAll(path); err != nil {
		return xerrors.Errorf("remove image %s dir failed: %w", imageName, err)
	}

	return nil
}
