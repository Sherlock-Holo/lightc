package rootfs

import (
	"syscall"

	"golang.org/x/xerrors"
)

func Unmount(rootFS *RootFS) error {
	if err := syscall.Unmount(rootFS.MergedDir, syscall.MNT_DETACH); err != nil {
		return xerrors.Errorf("unmount overlay failed:: %w", err)
	}

	return nil
}
