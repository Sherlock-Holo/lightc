package volume

import (
	"os"
	"path/filepath"
	"syscall"

	"github.com/Sherlock-Holo/lightc/libstorage/errors"
	"github.com/Sherlock-Holo/lightc/paths"
	"golang.org/x/xerrors"
)

func Unmount(rootFSID string, volumes []Volume) error {
	rootfsPath := filepath.Join(paths.RootFSPath, rootFSID)

	if _, err := os.Stat(rootfsPath); err != nil {
		if os.IsNotExist(err) {
			return errors.RootFSNotExist{ID: rootFSID}
		}
		return xerrors.Errorf("get rootfs stat failed: %w", err)
	}

	containerRoot := filepath.Join(rootfsPath, paths.MergedFile)

	errs := errors.VolumeErr{Op: errors.VolumeUnmountOp}

	for _, v := range volumes {
		dir := filepath.Join(containerRoot, v.ContainerDir)
		if _, err := os.Stat(dir); err != nil {
			if os.IsNotExist(err) {
				continue
			}

			errs.Errs = append(errs.Errs, xerrors.Errorf("get volume dir stat failed: %w", err))
			continue
		}

		if err := syscall.Unmount(dir, syscall.MNT_DETACH); err != nil {
			errs.Errs = append(errs.Errs, xerrors.Errorf("unmount volume %s:%s failed: %w", v.HostDir, dir, err))
		}
	}

	if len(errs.Errs) > 0 {
		return errs
	}
	return nil
}
