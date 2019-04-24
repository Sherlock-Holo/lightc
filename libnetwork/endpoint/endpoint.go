package endpoint

import (
	"net"

	"github.com/Sherlock-Holo/lightc/libnetwork/network"
	"github.com/vishvananda/netlink"
)

type Endpoint struct {
	ID      string
	Device  *netlink.Veth
	IP      net.IP
	PortMap []string
	Network *network.Network
}
