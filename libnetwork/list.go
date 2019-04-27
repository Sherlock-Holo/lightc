package libnetwork

import (
	"io/ioutil"

	"github.com/Sherlock-Holo/lightc/libnetwork/network"
	"github.com/Sherlock-Holo/lightc/paths"
	"golang.org/x/xerrors"
)

func ListNetwork() ([]*network.Network, error) {
	stats, err := ioutil.ReadDir(paths.BridgePath)
	if err != nil {
		return nil, xerrors.Errorf("read bridge dir failed: %w", err)
	}

	nws := make([]*network.Network, 0, len(stats))

	for _, stat := range stats {
		nw, err := loadNetwork(stat.Name())
		if err != nil {
			return nil, xerrors.Errorf("load network %s failed: %w", stat.Name(), err)
		}
		nws = append(nws, nw)
	}

	return nws, nil
}
