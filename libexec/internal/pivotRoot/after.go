package pivotRoot

import (
	"os"
	"path/filepath"
	"syscall"

	"golang.org/x/xerrors"
)

func afterPivotRoot(supportCgroups []Cgroup) error {
	mountPoints := []mountPoint{
		{
			Src:    "proc",
			Dst:    "/proc",
			Mode:   0755,
			IsDir:  true,
			fsType: "proc",
			Flags:  defaultMountFlags,
		},

		/*{
			Src:    "/tmpfs",
			Dst:    "/dev",
			Mode:   0755,
			IsDir:  true,
			fsType: "tmpfs",
			Flags:  syscall.MS_NOSUID,
			Data:   "mode=755",
		},*/

		{
			Src:    "mqueue",
			Dst:    "/dev/mqueue",
			Mode:   0777,
			IsDir:  true,
			fsType: "mqueue",
			Flags:  defaultMountFlags,
			Data:   "mode=777",
		},

		{
			Src:    "sysfs",
			Dst:    "/sys",
			Mode:   0555,
			IsDir:  true,
			fsType: "sysfs",
			Flags:  defaultMountFlags,
			Data:   "mode=555",
		},

		{
			Src:    "shm",
			Dst:    "/dev/shm",
			Mode:   0777,
			IsDir:  true,
			fsType: "tmpfs",
			Flags:  defaultMountFlags,
			Data:   "mode=1777,size=65536k",
		},

		{
			Src:    "devpts",
			Dst:    "/dev/pts",
			Mode:   0620,
			IsDir:  true,
			fsType: "devpts",
			Flags:  syscall.MS_NOSUID | syscall.MS_NOEXEC,
			Data:   "newinstance,ptmxmode=0666,mode=620",
		},

		{
			Src:    "tmpfs",
			Dst:    "/sys/fs/cgroup",
			Mode:   0555,
			IsDir:  true,
			fsType: "tmpfs",
			Flags:  defaultMountFlags,
			Data:   "mode=755",
		},
	}

	for _, mp := range mountPoints {
		if _, err := os.Stat(mp.Dst); err != nil {
			if os.IsNotExist(err) {
				if mp.IsDir {
					if err := os.MkdirAll(mp.Dst, mp.Mode); err != nil {
						return xerrors.Errorf("create %s empty directory failed: %w", mp.Dst, err)
					}
				} else {
					if file, err := os.Create(mp.Dst); err != nil {
						return xerrors.Errorf("create %s empty file failed: %w", mp.Dst, err)
					} else {
						_ = file.Close()
					}
				}

			} else {
				return xerrors.Errorf("get %s stat failed: %w", mp.Dst, err)
			}
		}

		if err := syscall.Mount(mp.Src, mp.Dst, mp.fsType, mp.Flags, mp.Data); err != nil {
			return xerrors.Errorf("mount %s failed: %w", mp.Dst, err)
		}
	}

	/*if err := mknod(); err != nil {
		return xerrors.Errorf("mknod failed: %w", err)
	}*/

	links := [][2]string{
		{"/proc/self/fd", "/dev/fd"},
		{"/proc/self/fd/0", "/dev/stdin"},
		{"/proc/self/fd/1", "/dev/stdout"},
		{"/proc/self/fd/2", "/dev/stderr"},
	}
	// https://github.com/opencontainers/runc/blob/bbb17efcb4c0ab986407812a31ba333a7450064c/libcontainer/rootfs_linux.go#L476
	// kcore support can be toggled with CONFIG_PROC_KCORE; only create a symlink
	// in /dev if it exists in /proc.
	if _, err := os.Stat("/proc/kcore"); err == nil {
		links = append(links, [2]string{"/proc/kcore", "/dev/core"})
	}

	for _, l := range links {
		if err := syscall.Symlink(l[0], l[1]); err != nil {
			return xerrors.Errorf("symbolic link %s to %s failed: %w", l[0], l[1], err)
		}
	}

	if err := syscall.Chdir("/sys/fs/cgroup"); err != nil {
		return xerrors.Errorf("chdir to /sys/fs/cgroup failed: %w", err)
	}

	for _, cgroup := range supportCgroups {
		if cgroup.Symlink {
			if err := os.MkdirAll(filepath.Dir(cgroup.To), 0755); err != nil {
				return xerrors.Errorf("mkdir %s failed: %w", filepath.Dir(cgroup.To), err)
			}

			if err := os.Symlink(cgroup.From, cgroup.To); err != nil {
				return xerrors.Errorf("create cgroup symlink from %s to %s failed: %w", cgroup.From, cgroup.To, err)
			}
		} else {
			if err := os.MkdirAll(cgroup.Name, 0755); err != nil {
				return xerrors.Errorf("mkdir cgroup %s failed: %w", cgroup.Name, err)
			}

			if err := syscall.Mount("cgroup", cgroup.Name, "cgroup", defaultMountFlags, cgroup.Data); err != nil {
				return xerrors.Errorf("mount cgroup %s failed: %w", cgroup.Name, err)
			}
		}
	}

	if err := syscall.Chdir("/"); err != nil {
		return xerrors.Errorf("chdir to / failed: %w", err)
	}

	return nil
}
