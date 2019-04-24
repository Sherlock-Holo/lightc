package pivotRoot

import (
	"os"
	"syscall"

	"golang.org/x/xerrors"
)

func afterPivotRoot() error {
	mountPoints := []mountPoint{
		{
			Src:    "proc",
			Dst:    "/proc",
			IsDir:  true,
			fsType: "proc",
			Flags:  defaultMountFlags,
		},

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
			Data:   "mode=777",
		},

		{
			Src:    "devpts",
			Dst:    "/dev/pts",
			Mode:   0620,
			IsDir:  true,
			fsType: "devpts",
			Flags:  syscall.MS_NOSUID | syscall.MS_NOEXEC,
			Data:   "ptmxmode=000,mode=620",
		},
	}

	for _, mp := range mountPoints {
		if _, err := os.Stat(mp.Dst); err != nil {
			if os.IsNotExist(err) {
				if mp.IsDir {
					if err := os.Mkdir(mp.Dst, mp.Mode); err != nil {
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

	return nil
}
