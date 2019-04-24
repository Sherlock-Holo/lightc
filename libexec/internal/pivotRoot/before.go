package pivotRoot

import (
	"os"
	"path/filepath"
	"syscall"

	"golang.org/x/xerrors"
)

func beforePivotRoot(root string) error {
	if err := os.MkdirAll(filepath.Join(root, "/dev"), 0700); err != nil {
		return xerrors.Errorf("mkdir %s failed: %w", filepath.Join(root, "/dev"), err)
	}

	if err := syscall.Mount("tmpfs", filepath.Join(root, "/dev"), "tmpfs", syscall.MS_NOSUID, "mode=755"); err != nil {
		return xerrors.Errorf("mount %s failed: %w", filepath.Join(root, "/dev"), err)
	}

	binds := []string{
		"/dev/null",
		"/dev/full",
		"/dev/random",
		"/dev/urandom",
		"/dev/zero",
	}

	for _, b := range binds {
		file, err := os.Create(filepath.Join(root, b))
		if err != nil {
			return xerrors.Errorf("create %s empty file failed: %w", b, err)
		}
		_ = file.Close()

		if err := syscall.Mount(b, filepath.Join(root, b), "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
			return xerrors.Errorf("mount %s failed: %w", b, err)
		}
	}

	return nil
}
