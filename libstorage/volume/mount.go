package volume

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/Sherlock-Holo/lightc/libstorage/errors"
	"github.com/Sherlock-Holo/lightc/paths"
	"golang.org/x/xerrors"
)

func Mount(rootFSID string, volumeStr ...string) (volumes []Volume, err error) {
	volumes = make([]Volume, 0, len(volumeStr))

	for _, str := range volumeStr {
		split := strings.Split(str, ":")
		if len(split) < 2 {
			continue
		}
		volumes = append(volumes, Volume{
			HostDir:      split[0],
			ContainerDir: split[1],
		})
	}

	rootfsPath := filepath.Join(paths.RootFSPath, rootFSID)

	if _, err := os.Stat(rootfsPath); err != nil {
		if os.IsNotExist(err) {
			return nil, errors.RootFSNotExist{ID: rootFSID}
		}
		return nil, xerrors.Errorf("get rootfs stat failed: %w", err)
	}

	containerRoot := filepath.Join(rootfsPath, paths.MergedFile)

	errs := errors.VolumeErr{Op: errors.VolumeMountOp}

	for _, v := range volumes {
		dir := filepath.Join(containerRoot, v.ContainerDir)
		if _, err := os.Stat(dir); err != nil {
			if os.IsNotExist(err) {
				if err := os.MkdirAll(dir, 0777); err != nil {
					errs.Errs = append(errs.Errs, xerrors.Errorf("mkdir container volume dir failed: %w", err))
					continue
				}
			} else {
				errs.Errs = append(errs.Errs, xerrors.Errorf("get volume dir stat failed: %w", err))
			}
		}

		if err := syscall.Mount(v.HostDir, dir, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
			errs.Errs = append(errs.Errs, xerrors.Errorf("bind volume dir %s:%s failed: %w", v.HostDir, dir, err))
		}
	}

	if len(errs.Errs) > 0 {
		return nil, errs
	}
	return volumes, nil
}
