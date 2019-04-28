package pivotRoot

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"

	"golang.org/x/xerrors"
)

func PivotRoot(root string) error {
	stats, err := ioutil.ReadDir("/sys/fs/cgroup")
	if err != nil {
		return xerrors.Errorf("get support cgroup failed: %w", err)
	}

	cgroups := make([]Cgroup, 0, len(stats))
	for _, stat := range stats {
		var cgroup Cgroup

		if stat.Mode()&os.ModeSymlink != 0 {
			from, err := os.Readlink(fmt.Sprintf("/sys/fs/cgroup/%s", stat.Name()))
			if err != nil {
				return xerrors.Errorf("get cgroup %s real from failed: %w", stat.Name(), err)
			}

			cgroup = Cgroup{
				Symlink: true,
				From:    from,
				To:      stat.Name(),
			}
		} else {
			cgroup = Cgroup{
				Name: stat.Name(),
				Data: stat.Name(),
			}

			if cgroup.Name == "systemd" {
				cgroup.Data = "xattr,name=systemd"
			}
		}

		cgroups = append(cgroups, cgroup)
	}

	// fix syscall.PivotRoot report invalid argument
	if err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		return xerrors.Errorf("make parent mount private error: %v", err)
	}

	if err := beforePivotRoot(root); err != nil {
		return xerrors.Errorf("mount before pivot root failed: %w", err)
	}

	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return xerrors.Errorf("mount rootfs to itself failed: %w", err)
	}

	pivotDir := filepath.Join(root, ".pivot_root")
	if err := os.Mkdir(pivotDir, 0777); err != nil {
		if !os.IsExist(err) {
			return err
		}
	}

	// pivot root
	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return xerrors.Errorf("pivot_root failed: %w", err)
	}

	if err := syscall.Chdir("/"); err != nil {
		return xerrors.Errorf("chdir / failed: %w", err)
	}

	if err := afterPivotRoot(cgroups); err != nil {
		return xerrors.Errorf("mount after pivot root failed: %w", err)
	}

	pivotDir = filepath.Join("/", ".pivot_root")
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return xerrors.Errorf("unmount pivot_root dir failed: %w", err)
	}

	_ = os.Remove(pivotDir)

	return nil
}
