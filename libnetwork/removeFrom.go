package libnetwork

import (
	"github.com/Sherlock-Holo/lightc/info"
	"github.com/Sherlock-Holo/lightc/libnetwork/internal/ipam"
	"golang.org/x/xerrors"
)

func RemoveContainerFromNetwork(cInfo *info.Info) error {
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
