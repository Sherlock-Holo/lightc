package nat

import (
	"github.com/coreos/go-iptables/iptables"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

func init() {
	var err error
	Iptables, err = iptables.New()
	if err != nil {
		logrus.Fatal(xerrors.Errorf("new iptables failed: %w", err))
	}

	setCustomChain()
}
