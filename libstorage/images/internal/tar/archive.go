package tar

import (
	"archive/tar"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sherlock-Holo/lightc/paths"
	"golang.org/x/xerrors"
)

func realArchive(parent, src string, tw *tar.Writer) error {
	relativePath := filepath.Join(parent, src)

	info, err := os.Lstat(relativePath)
	if err != nil {
		return xerrors.Errorf("get src %s stat failed: %w", relativePath, err)
	}

	if !info.IsDir() {
		var link string
		if info.Mode()&os.ModeSymlink != 0 {
			link, err = os.Readlink(relativePath)
			if err != nil {
				return xerrors.Errorf("read symbolic link failed: %w", err)
			}
		}

		header, err := tar.FileInfoHeader(info, link)
		if err != nil {
			return xerrors.Errorf("create file info header failed: %w", err)
		}

		header.Name = relativePath

		if err := tw.WriteHeader(header); err != nil {
			return xerrors.Errorf("write header failed: %w", err)
		}

		// https://stackoverflow.com/a/40003617
		if !info.Mode().IsRegular() {
			return nil
		}

		file, err := os.Open(relativePath)
		if err != nil {
			return xerrors.Errorf("open file %s failed: %w", relativePath, err)
		}
		defer func() {
			_ = file.Close()
		}()

		if _, err := io.Copy(tw, file); err != nil {
			return xerrors.Errorf("write file %s into tar failed: %w", relativePath, err)
		}

		return nil
	}

	infos, err := ioutil.ReadDir(relativePath)
	for _, info := range infos {
		if err := realArchive(relativePath, info.Name(), tw); err != nil {
			return xerrors.Errorf("recursive write file %s failed: %w", relativePath, err)
		}
	}

	return nil
}

func Archive(src, dst string) error {
	tarFile, err := os.OpenFile(dst, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0600)
	if err != nil {
		return xerrors.Errorf("create tar file failed: %w", err)
	}

	if err := os.Chdir(paths.LightcDir); err != nil {
		return xerrors.Errorf("change dir failed: %w", err)
	}

	src = strings.TrimPrefix(src, paths.LightcDir+string(filepath.Separator))

	tw := tar.NewWriter(tarFile)

	if err := realArchive("", src, tw); err != nil {
		return xerrors.Errorf("archive failed: %w", err)
	}

	if err := tw.Close(); err != nil {
		return xerrors.Errorf("close tar writer failed: %w", err)
	}
	return nil
}
