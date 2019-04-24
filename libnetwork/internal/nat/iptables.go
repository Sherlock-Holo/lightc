package nat

import (
	"fmt"
	"net"
	"strings"

	"github.com/Sherlock-Holo/lightc/info"
	"github.com/Sherlock-Holo/lightc/libnetwork/endpoint"
	"github.com/Sherlock-Holo/lightc/libnetwork/internal/ipam"
	"github.com/coreos/go-iptables/iptables"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

const (
	table       = "nat"
	customChain = "lightc"
	snatArgs    = "-s %s ! -o %s -j MASQUERADE"
	portMapArgs = "! -i %s -p tcp -m tcp --dport %s -j DNAT --to-destination %s:%s"
)

var Iptables *iptables.IPTables

func SetPortMap(ep *endpoint.Endpoint) {
	for _, pm := range ep.PortMap {
		portMap := strings.Split(pm, ":")
		if len(portMap) != 2 {
			logrus.Error(xerrors.Errorf("port map format error %s", pm))
			continue
		}

		args := fmt.Sprintf(portMapArgs, ep.Network.Name, portMap[0], ep.IP, portMap[1])
		if err := Iptables.AppendUnique(table, customChain, strings.Split(args, " ")...); err != nil {
			logrus.Error(xerrors.Errorf("iptables set DNAT failed: %w", err))
		}
	}
}

func UnsetPortMap(cInfo *info.Info) {
	for _, pm := range cInfo.PortMap {
		portMap := strings.Split(pm, ":")
		if len(portMap) != 2 {
			logrus.Error(xerrors.Errorf("port map format error %s", pm))
			continue
		}

		args := fmt.Sprintf(portMapArgs, cInfo.Network, portMap[0], cInfo.IPNet.IP, portMap[1])
		if err := Iptables.Delete(table, customChain, strings.Split(args, " ")...); err != nil {
			logrus.Error(xerrors.Errorf("iptables delete DNAT failed: %w", err))
		}
	}
}

// set custom chain lightc into nat table
func setCustomChain() {
	_ = Iptables.NewChain(table, customChain)
	if err := Iptables.AppendUnique(table, "OUTPUT", "-m", "addrtype", "--dst-type", "LOCAL", "-j", customChain); err != nil {
		logrus.Fatal(xerrors.Errorf("set output chain target failed: %w", err))
	}

	if err := Iptables.AppendUnique(table, "PREROUTING", "-m", "addrtype", "--dst-type", "LOCAL", "-j", customChain); err != nil {
		logrus.Fatal(xerrors.Errorf("set output chain target failed: %w", err))
	}
}

func SetSNAT(bridgeName string, subnet net.IPNet) error {
	if err := Iptables.AppendUnique(table, "POSTROUTING", "-o", bridgeName, "-m", "addrtype", "--src-type", "LOCAL", "-j", "MASQUERADE"); err != nil {
		return xerrors.Errorf("set src-type LOCAL failed: %w", err)
	}

	args := fmt.Sprintf(snatArgs, subnet.String(), bridgeName)
	if err := Iptables.AppendUnique(table, "POSTROUTING", strings.Split(args, " ")...); err != nil {
		return xerrors.Errorf("set SNAT failed: %w", err)
	}

	return nil
}

func UnsetSNAT(bridgeName string, subnet *net.IPNet) error {
	if err := Iptables.Delete(table, "POSTROUTING", "-o", bridgeName, "-m", "addrtype", "--src-type", "LOCAL", "-j", "MASQUERADE"); err != nil {
		return xerrors.Errorf("unset src-type LOCAL failed: %w", err)
	}

	args := fmt.Sprintf(snatArgs, subnet.String(), bridgeName)
	if err := Iptables.Delete(table, "POSTROUTING", strings.Split(args, " ")...); err != nil {
		return xerrors.Errorf("unset SNAT failed: %w", err)
	}

	delete(ipam.IPAllAllocator.SubnetMap, subnet.String())

	return nil
}
