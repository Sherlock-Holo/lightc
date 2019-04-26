package images

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/Sherlock-Holo/lightc/paths"
	"golang.org/x/xerrors"
)

func List() (images []Image, err error) {
	infos, err := ioutil.ReadDir(paths.ImagesPath)
	if err != nil {
		return nil, xerrors.Errorf("read images dir failed: %w", err)
	}

	for _, info := range infos {
		if !info.IsDir() {
			continue
		}

		var totalSize int64

		if err := filepath.Walk(filepath.Join(paths.ImagesPath, info.Name()), func(path string, i os.FileInfo, err error) error {
			if filepath.Join(paths.ImagesPath, i.Name()) == path {
				return nil
			}

			if i.IsDir() {
				return nil
			}

			totalSize += i.Size()
			return nil
		}); err != nil {
			return nil, xerrors.Errorf("get image %s size failed: %w", err)
		}

		images = append(images, Image{
			Name: info.Name(),
			Size: sizeInt(totalSize).String(),
		})
	}

	return images, nil
}

type sizeInt int64

func (si sizeInt) String() string {
	switch {
	case si < 1024:
		return strconv.Itoa(int(si)) + "B"

	case si < 1024*1024:
		return fmt.Sprintf("%.2fKB", float64(si)/1024)

	case si < 1024*1024*1024:
		return fmt.Sprintf("%.2fMB", float64(si)/(1024*1024))

	default:
		return fmt.Sprintf("%.2fGB", float64(si)/(1024*1024*1024))
	}
}
