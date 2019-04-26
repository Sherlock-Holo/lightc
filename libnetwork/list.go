package libnetwork

import (
	"io/ioutil"

	"github.com/Sherlock-Holo/lightc/libnetwork/network"
	"github.com/Sherlock-Holo/lightc/paths"
	"golang.org/x/xerrors"
)

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
