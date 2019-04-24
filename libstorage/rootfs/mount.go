package rootfs

import (
	"fmt"
	"syscall"

	"golang.org/x/xerrors"
)

const mountArgs = "lowerdir=%s,upperdir=%s,workdir=%s"

func Mount(rootFs *RootFS) error {
	args := fmt.Sprintf(mountArgs, rootFs.LowerDir, rootFs.UpperDir, rootFs.WorkDir)

	if err := syscall.Mount("overlay", rootFs.MergedDir, "overlay", 0, args); err != nil {
		return xerrors.Errorf("mount overlay failed: %w", err)
	}

	return nil
}
