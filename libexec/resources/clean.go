package resources

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"

	"github.com/Sherlock-Holo/lightc/info"
	"github.com/Sherlock-Holo/lightc/libnetwork"
	"github.com/Sherlock-Holo/lightc/libstorage/rootfs"
	"github.com/Sherlock-Holo/lightc/libstorage/volume"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

func CleanResources(parent *exec.Cmd, cInfo *info.Info, rootFS *rootfs.RootFS, configFile *os.File, rmAfterRun bool, done context.CancelFunc) {
	_ = parent.Wait()

	_ = cInfo.Stdout.Close()
	_ = cInfo.Stderr.Close()

	cInfo.Status = info.STOPPED

	if err := volume.Unmount(rootFS.ID, cInfo.Volumes); err != nil {
		logrus.Error(xerrors.Errorf("unmount volume failed: %w", err))
	}

	if err := rootfs.Unmount(rootFS); err != nil {
		logrus.Error(xerrors.Errorf("unmount rootfs failed: %w", err))
	}

	if err := libnetwork.RemoveContainerFromNetwork(cInfo); err != nil {
		logrus.Error(xerrors.Errorf("remove container %s from network %s failed: %w", cInfo.ID, cInfo.Network, err))
	}

	if configFile != nil {
		if err := configFile.Truncate(0); err != nil {
			logrus.Error(xerrors.Errorf("truncate container %s config file failed: %w", cInfo.ID, err))
		} else {
			encoder := json.NewEncoder(configFile)
			encoder.SetIndent("", "    ")
			if err := encoder.Encode(cInfo); err != nil {
				logrus.Error(xerrors.Errorf("encode container %s config file failed: %w", cInfo.ID, err))
			}
		}

		_ = configFile.Close()
	}

	if rmAfterRun {
		if err := rootfs.Delete(rootFS.ID); err != nil {
			logrus.Error(xerrors.Errorf("remove rootfs %s failed: %w", rootFS.ID, err))
		}
	}

	if done != nil {
		done()
	}
}
