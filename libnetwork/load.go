package libnetwork

import (
	"encoding/json"
	"os"
	"path/filepath"
	"syscall"

	"github.com/Sherlock-Holo/lightc/libnetwork/network"
	"github.com/Sherlock-Holo/lightc/paths"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

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
