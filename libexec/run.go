package libexec

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/Sherlock-Holo/lightc/info"
	"github.com/Sherlock-Holo/lightc/libexec/internal/process"
	"github.com/Sherlock-Holo/lightc/libexec/resources"
	"github.com/Sherlock-Holo/lightc/libnetwork"
	"github.com/Sherlock-Holo/lightc/libstorage/rootfs"
	"github.com/Sherlock-Holo/lightc/libstorage/volume"
	"github.com/Sherlock-Holo/lightc/paths"
	"github.com/kr/pty"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/xerrors"
)

func Run(
	cmdStr string,
	rootFS *rootfs.RootFS,
	envs []string,
	networkName string,
	volumes []volume.Volume,
	tty bool,
	detach bool,
	rmAfterRun bool,
) (*info.Info, error) {
	cInfo := &info.Info{
		ID:         rootFS.ID,
		RootFS:     rootFS,
		Cmd:        strings.Split(cmdStr, " "),
		CreateTime: info.CustomTime(time.Now()),
		// Parent:     parent,
		Status:     info.RUNNING,
		Volumes:    volumes,
		TTY:        tty,
		Detach:     detach,
		RmAfterRun: rmAfterRun,
		Envs:       envs,
		Network:    networkName,
		ImageName:  rootFS.ImageName,
		CmdStr:     cmdStr,
	}

	if !detach {
		parent, wPipe, err := process.NewParentProcess(
			envs,
			rootFS,
		)
		if err != nil {
			return nil, xerrors.Errorf("new parent process failed: %w", err)
		}
		defer func() {
			_ = wPipe.Close()
		}()

		cInfo.Parent = parent

		if tty {
			ptm, err := pty.Start(parent)
			if err != nil {
				return nil, xerrors.Errorf("start parent process with pty failed: %w", err)
			}

			cInfo.Stdin = ptm
			cInfo.Stdout = ptm
			cInfo.Stderr = ptm
		} else {
			stdoutR, stdoutW, err := os.Pipe()
			if err != nil {
				return nil, xerrors.Errorf("create stdout pipe failed: %w", err)
			}

			parent.Stdout = stdoutW
			cInfo.Stdout = stdoutR

			stderrR, stderrW, err := os.Pipe()
			if err != nil {
				return nil, xerrors.Errorf("create stderr pipe failed: %w", err)
			}

			parent.Stderr = stderrW
			cInfo.Stderr = stderrR

			if err := parent.Start(); err != nil {
				return nil, xerrors.Errorf("start parent process with pipe failed: %w", err)
			}
		}

		cInfo.Pid = parent.Process.Pid

		configFile, err := os.OpenFile(filepath.Join(paths.RootFSPath, cInfo.ID, paths.ConfigName), os.O_CREATE|os.O_RDWR, 0600)
		if err != nil {
			// kill parent because config file create failed
			_ = parent.Process.Kill()
			return nil, xerrors.Errorf("create container %s config file failed: %w", cInfo.ID, err)
		}

		encoder := json.NewEncoder(configFile)
		encoder.SetIndent("", "    ")
		if err := encoder.Encode(cInfo); err != nil {
			// kill parent because config file save failed
			_ = parent.Process.Kill()
			return nil, xerrors.Errorf("encode container %s config failed: %w", cInfo.ID, err)
		}

		// add container into network
		if networkName != "" {
			if err := libnetwork.AddContainerIntoNetwork(networkName, cInfo); err != nil {
				// kill parent because set network failed
				_ = parent.Process.Kill()
				return nil, xerrors.Errorf("add container into network failed: %w", err)
			}
		}

		// send info to parent
		if err := json.NewEncoder(wPipe).Encode(cInfo); err != nil {
			// kill parent because send info failed
			_ = parent.Process.Kill()
			return nil, xerrors.Errorf("send info to parent process failed: %w", err)
		}

		// resources clean
		defer resources.CleanResources(parent, cInfo, rootFS, configFile, rmAfterRun, nil)

		if tty {
			oldState, err := terminal.MakeRaw(int(os.Stdin.Fd()))
			if err != nil {
				return nil, xerrors.Errorf("set stdin raw mode failed: %w", err)
			}
			defer func() {
				if err := terminal.Restore(int(os.Stdin.Fd()), oldState); err != nil {
					logrus.Error(xerrors.Errorf("restore stdin state from raw mode failed: %W", err))
				}
			}()

			signalCh := make(chan os.Signal, 1)
			signal.Notify(signalCh, syscall.SIGWINCH)

			if err := pty.InheritSize(os.Stdin, cInfo.Stdin.(*os.File)); err != nil {
				logrus.Error(xerrors.Errorf("set ptm size failed: %w", err))
			}

			go func() {
				for range signalCh {
					if err := pty.InheritSize(os.Stdin, cInfo.Stdin.(*os.File)); err != nil {
						logrus.Error(xerrors.Errorf("set ptm size failed: %w", err))
					}
				}
			}()
		}

		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			_, _ = io.Copy(cInfo.Stdin, os.Stdin)
			cancel()
		}()
		go func() {
			_, _ = io.Copy(os.Stdout, cInfo.Stdout)
			cancel()
		}()

		<-ctx.Done()

		return cInfo, nil
	}

	monitorPipeR, monitorPipeW, err := os.Pipe()
	if err != nil {
		return nil, xerrors.Errorf("create monitor pipe failed: %w", err)
	}
	defer func() {
		_ = monitorPipeW.Close()
	}()

	monitor := exec.Command("/proc/self/exe", "monitor")
	monitor.ExtraFiles = append(monitor.ExtraFiles, monitorPipeR)

	if err := monitor.Start(); err != nil {
		return nil, xerrors.Errorf("start monitor failed: %w", err)
	}

	if err := json.NewEncoder(monitorPipeW).Encode(cInfo); err != nil {
		return nil, xerrors.Errorf("write container info to monitor failed: %w", err)
	}

	return cInfo, nil
}
