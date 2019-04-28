package rootfs

import (
	"fmt"
	"path/filepath"
	"syscall"

	"golang.org/x/xerrors"
)

const mountArgs = "lowerdir=%s,upperdir=%s,workdir=%s"

func Mount(rootFs *RootFS) error {
	args := fmt.Sprintf(mountArgs, rootFs.LowerDir, rootFs.UpperDir, rootFs.WorkDir)

	if err := syscall.Mount("overlay", rootFs.MergedDir, "overlay", 0, args); err != nil {
		return xerrors.Errorf("mount overlay failed: %w", err)
	}

	if err := syscall.Mount(rootFs.dnsFile, filepath.Join(rootFs.MergedDir, "etc", "resolv.conf"), "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return xerrors.Errorf("bind resolv.conf file failed: %w", err)
	}

	if err := syscall.Mount(rootFs.hostnameFile, filepath.Join(rootFs.MergedDir, "etc", "hostname"), "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return xerrors.Errorf("bind hostname file failed: %w", err)
	}

	if err := syscall.Mount(rootFs.Hosts, filepath.Join(rootFs.MergedDir, "etc", "hosts"), "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return xerrors.Errorf("bind hosts file failed: %w", err)
	}

	return nil
}
