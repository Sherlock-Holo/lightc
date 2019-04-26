package libnetwork

import (
	"os"
	"path/filepath"

	"github.com/Sherlock-Holo/lightc/libnetwork/internal/ipam"
	"github.com/Sherlock-Holo/lightc/libnetwork/internal/nat"
	"github.com/Sherlock-Holo/lightc/paths"
	"github.com/vishvananda/netlink"
	"golang.org/x/xerrors"
)

func RemoveNetwork(name string) error {
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
