package tar

import (
	"archive/tar"
	"bufio"
	"io"
	"os"
	"path/filepath"

	"github.com/Sherlock-Holo/lightc/paths"
	"golang.org/x/xerrors"
)

func Extract(src, imageName string) error {
	file, err := os.Open(src)
	if err != nil {
		return xerrors.Errorf("open file failed: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	bufReader := bufio.NewReader(file)
	tarReader := tar.NewReader(bufReader)

	for {
		header, err := tarReader.Next()

		switch {
		case xerrors.Is(err, io.EOF):
			return nil

		default:
			return xerrors.Errorf("get tar next header failed: %w", err)

		case header == nil:
			continue

		case err == nil:
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(filepath.Join(paths.ImagesPath, imageName, header.Name), os.FileMode(header.Mode)); err != nil {
				return xerrors.Errorf("create dir failed: %w", err)
			}

		case tar.TypeSymlink:
			if err := os.Symlink(header.Linkname, filepath.Join(paths.ImagesPath, imageName, header.Name)); err != nil {
				return xerrors.Errorf("create symbolic link failed: %w", err)
			}

		case tar.TypeReg:
			// f, err := os.Create(filepath.Join(paths.ImagesPath, imageName, header.Name))
			f, err := os.OpenFile(filepath.Join(paths.ImagesPath, imageName, header.Name), os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return xerrors.Errorf("create file failed: %w", err)
			}
			if _, err := io.Copy(f, tarReader); err != nil {
				_ = f.Close()
				return xerrors.Errorf("write file failed: %w", err)
			}
			_ = f.Close()
		}
	}
}
