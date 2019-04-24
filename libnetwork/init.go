package libnetwork

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/Sherlock-Holo/lightc/libnetwork/internal/nat"
	"github.com/Sherlock-Holo/lightc/paths"
	"github.com/coreos/go-iptables/iptables"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

func init() {
	if err := os.MkdirAll(paths.BridgePath, 0700); err != nil {
		logrus.Fatal(xerrors.Errorf("create bridge dir failed: %w", err))
	}

	if err := os.MkdirAll(filepath.Dir(paths.IPAllocatorPath), 0700); err != nil {
		logrus.Fatal(xerrors.Errorf("create ipam dir failed: %w", err))
	}

	// allow forward
	if err := ioutil.WriteFile("/proc/sys/net/ipv4/ip_forward", []byte(strconv.Itoa(1)), 0644); err != nil {
		logrus.Fatal(xerrors.Errorf("set ip forward failed: %w", err))
	}

	var err error
	if nat.Iptables, err = iptables.New(); err != nil {
		logrus.Fatal(xerrors.Errorf("new iptables setter failed: %w", err))
	}

	nws, err := ListNetwork()
	if err != nil {
		logrus.Error(xerrors.Errorf("get all network failed: %w", err))
	}

	// allow nat local network
	for _, nw := range nws {
		path := filepath.Join("/proc/sys/net/ipv4/conf", nw.Name, "route_localnet")
		if err := ioutil.WriteFile(path, []byte(strconv.Itoa(1)), 0644); err != nil {
			logrus.Fatal(xerrors.Errorf("enable bridge local nat failed: %w", err))
		}
	}
}
