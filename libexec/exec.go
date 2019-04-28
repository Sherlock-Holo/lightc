package libexec

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Sherlock-Holo/lightc/info"
	"github.com/Sherlock-Holo/lightc/libexec/errors"
	_ "github.com/Sherlock-Holo/lightc/libexec/internal/nsenter"
	"github.com/kr/pty"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/xerrors"
)

const (
	pidEnv = "lightc_pid"
	cmdEnv = "lightc_cmd"
)

func Exec(containerID string, cmdStr string, tty bool) error {
	cInfo, err := info.GetInfo(containerID)
	if err != nil {
		return xerrors.Errorf("get container %s info failed: %w", containerID, err)
	}

	if cInfo.Status == info.STOPPED {
		return errors.ContainerStopped{ID: containerID}
	}

	b, err := ioutil.ReadFile(filepath.Join("/proc", strconv.Itoa(cInfo.Pid), "environ"))
	if err != nil {
		return xerrors.Errorf("read container %s envs failed: %w", containerID, err)
	}

	cmd := exec.Command("/proc/self/exe", []string{"exec", containerID, cmdStr}...)

	cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%d", pidEnv, cInfo.Pid))
	cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", cmdEnv, cmdStr))
	cmd.Env = append(cmd.Env, strings.Split(string(b), "\u0000")...)

	if tty {
		oldState, err := terminal.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			return xerrors.Errorf("set stdin raw mode failed: %w", err)
		}
		defer func() {
			if err := terminal.Restore(int(os.Stdin.Fd()), oldState); err != nil {
				logrus.Error(xerrors.Errorf("restore stdin state from raw mode failed: %W", err))
			}
		}()

		ptm, err := pty.Start(cmd)
		if err != nil {
			return xerrors.Errorf("start exec process failed: %w", err)
		}

		go func() {
			_, _ = io.Copy(ptm, os.Stdin)

		}()

		_, _ = io.Copy(os.Stdout, ptm)

	} else {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return xerrors.Errorf("run exec process failed: %w", err)
		}
	}

	return nil
}
