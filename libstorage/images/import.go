package images

import (
	"os"
	"path/filepath"

	"github.com/Sherlock-Holo/lightc/libstorage/errors"
	"github.com/Sherlock-Holo/lightc/libstorage/images/internal/tar"
	"github.com/Sherlock-Holo/lightc/paths"
	"golang.org/x/xerrors"
)

func ImportImage(imagePath, imageName string) (err error) {
	if info, err := os.Stat(imagePath); err != nil {
		if os.IsNotExist(err) {
			// this error message is enough
			return err
		}
		return xerrors.Errorf("get import path failed: %w", err)
	} else if info.IsDir() {
		return xerrors.Errorf("import path is not file: %w", err)
	}

	if _, err := os.Stat(filepath.Join(paths.ImagesPath, imageName)); err == nil {
		return errors.ImageImportConflict{ConflictName: imageName}
	}

	if err := os.Mkdir(filepath.Join(paths.ImagesPath, imageName), 0700); err != nil {
		return xerrors.Errorf("mkdir failed: %w", err)
	}

	if err := tar.Extract(imagePath, imageName); err != nil {
		return xerrors.Errorf("extract failed: %w", err)
	}

	return nil
}
