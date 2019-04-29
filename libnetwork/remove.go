package libnetwork

import (
	"os"
	"path/filepath"
	"syscall"

	"github.com/Sherlock-Holo/lightc/libnetwork/internal/ipam"
	"github.com/Sherlock-Holo/lightc/libnetwork/internal/nat"
	"github.com/Sherlock-Holo/lightc/paths"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"golang.org/x/xerrors"
)

func RemoveNetwork(name string) error {
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

	nw, err := loadNetwork(name)
	if err != nil {
		return xerrors.Errorf("load network failed: %w", err)
	}

	if err := ipam.IPAllAllocator.DeleteSubnet(nw.Subnet); err != nil {
		return xerrors.Errorf("delete subnet failed: %w", err)
	}

	link, err := netlink.LinkByName(name)
	if err != nil {
		return xerrors.Errorf("get network bridge failed: %w", err)
	}

	if err := netlink.LinkDel(link); err != nil {
		return xerrors.Errorf("delete bridge failed: %w", err)
	}

	if err := nat.UnsetSNAT(nw.Name, nw.Subnet); err != nil {
		return xerrors.Errorf("unset SNAT failed: %w", err)
	}

	if err := os.Remove(filepath.Join(paths.BridgePath, nw.Name)); err != nil {
		return xerrors.Errorf("remove bridge %s config file failed: %w", nw.Name, err)
	}

	return nil
}
