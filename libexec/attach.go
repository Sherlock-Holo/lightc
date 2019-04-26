package libexec

import (
	"encoding/json"
	"io"
	"net"
	"os"
	"path/filepath"

	"github.com/Sherlock-Holo/lightc/info"
	"github.com/Sherlock-Holo/lightc/libexec/errors"
	"github.com/Sherlock-Holo/lightc/paths"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/xerrors"
)

func Attach(containerID string) error {
	configFile, err := os.Open(filepath.Join(paths.RootFSPath, containerID, paths.ConfigName))
	switch {
	case os.IsNotExist(err):
		return errors.ContainerNotExist{ID: containerID}

	default:
		return xerrors.Errorf("open container %s config file failed: %w", containerID, err)

	case err == nil:
	}
	defer func() {
		_ = configFile.Close()
	}()

	cInfo := new(info.Info)
	if err := json.NewDecoder(configFile).Decode(cInfo); err != nil {
		return xerrors.Errorf("decode container %s config file failed: %w", containerID, err)
	}

	if cInfo.Status != info.RUNNING {
		return xerrors.Errorf("container %s is stopped", cInfo.ID)
	}

	conn, err := net.Dial("unix", cInfo.UnixSocket)
	if err != nil {
		return xerrors.Errorf("connect container %s monitor unix socket failed: %w", cInfo.ID, err)
	}
	defer func() {
		_ = conn.Close()
	}()

	if cInfo.TTY {
		oldState, err := terminal.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			return xerrors.Errorf("set stdin raw mode failed: %w", err)
		}
		defer func() {
			if err := terminal.Restore(int(os.Stdin.Fd()), oldState); err != nil {
				logrus.Error(xerrors.Errorf("restore stdin state from raw mode failed: %W", err))
			}
		}()
	}

	go func() {
		_, _ = io.Copy(conn, os.Stdin)
	}()

	_, _ = io.Copy(os.Stdout, conn)

	return nil
}
