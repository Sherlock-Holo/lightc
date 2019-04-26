package libexec

import (
	"io"
	"net"
	"os"

	"github.com/Sherlock-Holo/lightc/info"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/xerrors"
)

func Attach(containerID string) error {
	cInfo, err := info.GetInfo(containerID)
	notExist := new(info.ContainerNotExist)
	switch {
	case xerrors.As(err, notExist):
		return notExist

	default:
		return xerrors.Errorf("get container info failed: %w", err)

	case err == nil:
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
