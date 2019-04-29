package libnetwork

import (
	"io/ioutil"
	"os"
	"syscall"

	"github.com/Sherlock-Holo/lightc/libnetwork/network"
	"github.com/Sherlock-Holo/lightc/paths"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

func ListNetwork() ([]*network.Network, error) {
	f, err := os.Open(paths.NetworkLock)
	if err != nil {
		return nil, xerrors.Errorf("open lock file failed: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return nil, xerrors.Errorf("lock network failed: %w", err)
	}
	defer func() {
		if err := syscall.Flock(int(f.Fd()), syscall.LOCK_UN); err != nil {
			logrus.Error(xerrors.Errorf("unlock network failed: %w", err))
		}
	}()

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
