package libexec

import (
	"os"

	"github.com/Sherlock-Holo/lightc/info"
	"github.com/Sherlock-Holo/lightc/libexec/errors"
	"golang.org/x/xerrors"
)

func Stop(containerID string, kill bool) error {
	cInfo, err := info.GetInfo(containerID)

	containerNotExist := new(info.ContainerNotExist)
	switch {
	case xerrors.As(err, containerNotExist):
		return containerNotExist

	default:
		return xerrors.Errorf("get container info failed: %w", err)

	case err == nil:
	}

	if cInfo.Status == info.STOPPED {
		return errors.ContainerStopped{ID: containerID}
	}

	process, err := os.FindProcess(cInfo.Pid)
	if err != nil {
		return xerrors.Errorf("find container %s process failed: %w", containerID, err)
	}

	if !kill {
		if err := process.Kill(); err != nil {
			return xerrors.Errorf("kill container %s process failed: %w", containerID, err)
		}
	} else {
		if err := process.Signal(os.Interrupt); err != nil {
			return xerrors.Errorf("stop container %s process failed: %w", containerID, err)
		}
	}

	return nil
}
