package libnetwork

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/Sherlock-Holo/lightc/libnetwork/internal/nat"
	"github.com/coreos/go-iptables/iptables"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

func init() {
	// avoid init network in container
	if os.Args[0] == "/proc/self/exe" {
		return
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
		path := fmt.Sprintf("/proc/sys/net/ipv4/conf/%s/route_localnet", nw.Name)
		if err := ioutil.WriteFile(path, []byte(strconv.Itoa(1)), 0644); err != nil {
			logrus.Fatal(xerrors.Errorf("enable bridge local nat failed: %w", err))
		}
	}
}
