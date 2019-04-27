package pivotRoot

import (
	"syscall"

	"golang.org/x/sys/unix"
	"golang.org/x/xerrors"
)

type devNode struct {
	Dst   string
	Major uint32
	minor uint32
}

func mknod() error {
	nodes := []devNode{
		{
			Dst:   "/dev/random",
			Major: 1,
			minor: 8,
		},

		{
			Dst:   "/dev/urandom",
			Major: 1,
			minor: 9,
		},

		{
			Dst:   "/dev/zero",
			Major: 1,
			minor: 5,
		},

		{
			Dst:   "/dev/null",
			Major: 1,
			minor: 3,
		},

		{
			Dst:   "/dev/full",
			Major: 1,
			minor: 7,
		},
	}

	for _, node := range nodes {
		if err := syscall.Mknod(node.Dst, syscall.S_IFCHR, int(unix.Mkdev(node.Major, node.minor))); err != nil {
			return xerrors.Errorf("mknod %s major %d minor %d failed: %w", node.Dst, node.Major, node.minor, err)
		}
	}

	return nil
}
