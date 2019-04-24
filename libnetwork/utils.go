package libnetwork

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"syscall"

	"github.com/Sherlock-Holo/lightc/info"
	"github.com/Sherlock-Holo/lightc/libnetwork/endpoint"
	"github.com/Sherlock-Holo/lightc/libnetwork/internal/ipam"
	"github.com/Sherlock-Holo/lightc/libnetwork/internal/nat"
	"github.com/Sherlock-Holo/lightc/libnetwork/network"
	"github.com/Sherlock-Holo/lightc/paths"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"golang.org/x/xerrors"
)

func NewNetwork(name string, subnet net.IPNet) (*network.Network, error) {
	if _, err := os.Stat(filepath.Join(paths.BridgePath, name)); err == nil {
		return nil, xerrors.Errorf("network %s exists", name)
	}

	file, err := os.OpenFile(filepath.Join(paths.BridgePath, name), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, xerrors.Errorf("create network file failed: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		return nil, xerrors.Errorf("lock network file failed: %W", err)
	}
	defer func() {
		if err := syscall.Flock(int(file.Fd()), syscall.LOCK_UN); err != nil {
			logrus.Error(xerrors.Errorf("unlock network file failed: %W", err))
		}
	}()

	ip, exist, err := ipam.IPAllAllocator.AllocateSubnet(subnet)
	if err != nil {
		return nil, xerrors.Errorf("allocate subnet failed: %w", err)
	}

	if exist {
		_ = os.Remove(file.Name())
		return nil, xerrors.Errorf("subnet %s exists", subnet)
	}

	nw := &network.Network{
		Name:    name,
		Subnet:  subnet,
		Gateway: ip,
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	if err := encoder.Encode(nw); err != nil {
		return nil, xerrors.Errorf("encode network file failed: %w", err)
	}

	if err := initNetwork(nw); err != nil {
		return nil, xerrors.Errorf("init network failed: %w", err)
	}

	return nw, nil
}

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

	if err := os.Remove(filepath.Join(paths.BridgePath, nw.Name)); err != nil {
		return xerrors.Errorf("remove bridge %s config file failed: %w", nw.Name, err)
	}

	return nil
}

func loadNetwork(name string) (*network.Network, error) {
	loadPath := filepath.Join(paths.BridgePath, name)
	if _, err := os.Stat(loadPath); err != nil {
		if os.IsNotExist(err) {
			return nil, xerrors.Errorf("network %s doesn't exist", name)
		}
		return nil, xerrors.Errorf("get network stat failed: %w", err)
	}

	file, err := os.Open(loadPath)
	if err != nil {
		return nil, xerrors.Errorf("open network file failed: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		return nil, xerrors.Errorf("lock network file failed: %w", err)
	}
	defer func() {
		if err := syscall.Flock(int(file.Fd()), syscall.LOCK_UN); err != nil {
			logrus.Error(xerrors.Errorf("unlock net work file failed: %w", err))
		}
	}()

	nw := new(network.Network)
	if err := json.NewDecoder(file).Decode(nw); err != nil {
		return nil, xerrors.Errorf("decode network file failed: %w", err)
	}

	return nw, nil
}

func initNetwork(nw *network.Network) error {
	br, err := createBridge(nw.Name)
	if err != nil {
		return xerrors.Errorf("create bridge failed: %w", err)
	}

	gatewayIP := nw.Subnet
	gatewayIP.IP = nw.Gateway

	if err := setInterfaceIP(br, gatewayIP); err != nil {
		return xerrors.Errorf("set interface gateway IP failed: %w", err)
	}

	if err := setInterfaceUP(br); err != nil {
		return xerrors.Errorf("set bridge up failed: %w", err)
	}

	if err := nat.SetSNAT(nw.Name, nw.Subnet); err != nil {
		return xerrors.Errorf("setup iptables failed: %w", err)
	}

	return nil
}

func createBridge(bridgeName string) (br netlink.Link, err error) {
	if _, err := net.InterfaceByName(bridgeName); err == nil {
		return nil, nil
	}

	linkAttrs := netlink.NewLinkAttrs()
	linkAttrs.Name = bridgeName

	br = &netlink.Bridge{LinkAttrs: linkAttrs}

	if err := netlink.LinkAdd(br); err != nil {
		return nil, xerrors.Errorf("add br %s failed: %w", bridgeName, err)
	}

	return br, nil
}

func setInterfaceIP(iface netlink.Link, ipNet net.IPNet) error {
	if err := netlink.AddrAdd(iface, &netlink.Addr{
		IPNet: &ipNet,
	}); err != nil {
		return xerrors.Errorf("add addr to iface failed: %w", err)
	}
	return nil
}

func setInterfaceUP(iface netlink.Link) error {
	if err := netlink.LinkSetUp(iface); err != nil {
		return xerrors.Errorf("set bridge up failed: %w", err)
	}

	return nil
}

func setEndpointIPAndRoute(ep *endpoint.Endpoint, containerInfo *info.Info) error {
	peerLink, err := netlink.LinkByName(ep.Device.PeerName)
	if err != nil {
		return xerrors.Errorf("get peer name failed: %w", err)
	}

	exitNetns, err := enterNetns(peerLink, containerInfo)
	if err != nil {
		return xerrors.Errorf("enter netns failed: %w", err)
	}
	defer exitNetns()

	interfaceIP := ep.Network.Subnet
	interfaceIP.IP = ep.IP

	if err := setInterfaceIP(peerLink, interfaceIP); err != nil {
		return xerrors.Errorf("set container interface IP failed: %w", err)
	}

	if err := setInterfaceUP(peerLink); err != nil {
		return xerrors.Errorf("set up container interface failed: %w", err)
	}

	lo, err := netlink.LinkByName("lo")
	if err != nil {
		return xerrors.Errorf("get iface lo failed: %w", err)
	}

	if err := setInterfaceUP(lo); err != nil {
		return xerrors.Errorf("set up container lo failed: %w", err)
	}

	_, ipNet, _ := net.ParseCIDR("0.0.0.0/0")
	defaultRoute := &netlink.Route{
		LinkIndex: peerLink.Attrs().Index,
		Gw:        ep.Network.Gateway,
		Dst:       ipNet,
	}

	if err := netlink.RouteAdd(defaultRoute); err != nil {
		return xerrors.Errorf("add default route failed: %w", err)
	}

	return nil
}

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

func ListNetwork() ([]*network.Network, error) {
	infos, err := ioutil.ReadDir(paths.BridgePath)
	if err != nil {
		return nil, xerrors.Errorf("read bridge dir failed: %w", err)
	}

	nws := make([]*network.Network, 0, len(infos))

	for _, f := range infos {
		nw, err := loadNetwork(f.Name())
		if err != nil {
			return nil, xerrors.Errorf("load network %s failed: %w", f.Name(), err)
		}
		nws = append(nws, nw)
	}

	return nws, nil
}
