package libnetwork

import (
	"os"
	"syscall"

	"github.com/Sherlock-Holo/lightc/info"
	"github.com/Sherlock-Holo/lightc/libnetwork/internal/ipam"
	"github.com/Sherlock-Holo/lightc/paths"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

func RemoveContainerFromNetwork(cInfo *info.Info) error {
	f, err := os.Open(paths.NetworkLock)
	if err != nil {
		return xerrors.Errorf("open lock file failed: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return xerrors.Errorf("lock network failed: %w", err)
	}
	defer func() {
		if err := syscall.Flock(int(f.Fd()), syscall.LOCK_UN); err != nil {
			logrus.Error(xerrors.Errorf("unlock network failed: %w", err))
		}
	}()

	if cInfo.IPNet.IP == nil {
		return nil
	}

	nw, err := loadNetwork(cInfo.Network)
	if err != nil {
		return xerrors.Errorf("load network failed: %w", err)
	}

	if err := ipam.IPAllAllocator.Release(nw.Subnet, cInfo.IPNet.IP); err != nil {
		return xerrors.Errorf("release ip failed: %w", err)
	}
	return nil
}
