package libnetwork

import (
	"fmt"
	"net"

	"github.com/Sherlock-Holo/lightc/info"
	"github.com/Sherlock-Holo/lightc/libnetwork/endpoint"
	"github.com/Sherlock-Holo/lightc/libnetwork/internal/ipam"
	"github.com/Sherlock-Holo/lightc/libnetwork/internal/nat"
	"github.com/vishvananda/netlink"
	"golang.org/x/xerrors"
)

func AddContainerIntoNetwork(networkName string, cInfo *info.Info) error {
	nw, err := loadNetwork(networkName)
	if err != nil {
		return xerrors.Errorf("load network failed: %w", err)
	}

	_, _, err = net.ParseCIDR(nw.Subnet.String())
	if err != nil {
		return xerrors.Errorf("network subnet invalid: %w", err)
	}

	ip, err := ipam.IPAllAllocator.Allocate(nw.Subnet)
	if err != nil {
		return xerrors.Errorf("allocate ip failed: %w", err)
	}

	ep := &endpoint.Endpoint{
		ID:      fmt.Sprintf("%s-%s", networkName, cInfo.ID),
		IP:      ip,
		Network: nw,
		PortMap: cInfo.PortMap,
	}

	br, err := netlink.LinkByName(nw.Name)
	if err != nil {
		return xerrors.Errorf("get bridge failed: %w", err)
	}

	linkAttrs := netlink.NewLinkAttrs()
	linkAttrs.Name = ep.ID[:5]

	linkAttrs.MasterIndex = br.Attrs().Index

	ep.Device = &netlink.Veth{
		LinkAttrs: linkAttrs,
		PeerName:  "cif-" + ep.ID[:5],
	}

	if err := netlink.LinkAdd(ep.Device); err != nil {
		return xerrors.Errorf("add endpoint device failed: %w", err)
	}

	if err := netlink.LinkSetUp(ep.Device); err != nil {
		return xerrors.Errorf("set up endpoint device failed: %w", err)
	}

	if err := setEndpointIPAndRoute(ep, cInfo); err != nil {
		return xerrors.Errorf("set endpoint ip and route failed: %w", err)
	}

	nat.SetPortMap(ep)

	cInfo.Network = networkName
	cInfo.IPNet = nw.Subnet
	cInfo.IPNet.IP = ip

	return nil
}
