package process

import (
	"encoding/json"
	"os"
	"os/exec"
	"syscall"

	"github.com/Sherlock-Holo/lightc/info"
	"github.com/Sherlock-Holo/lightc/libexec/internal/pivotRoot"
	"github.com/Sherlock-Holo/lightc/libstorage/rootfs"
	"golang.org/x/sys/unix"
	"golang.org/x/xerrors"
)

func NewParentProcess(envs []string, rootFS *rootfs.RootFS) (cmd *exec.Cmd, wPipe *os.File, err error) {
	cmd = exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUSER |
			syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET |
			syscall.CLONE_NEWIPC |
			unix.CLONE_NEWCGROUP,

		UidMappings: []syscall.SysProcIDMap{
			{
				HostID:      os.Geteuid(),
				ContainerID: 0,
				Size:        4294967295,
			},
		},

		GidMappings: []syscall.SysProcIDMap{
			{
				HostID:      os.Getegid(),
				ContainerID: 0,
				Size:        4294967295,
			},
		},
	}

	/*cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET |
			syscall.CLONE_NEWIPC |
			unix.CLONE_NEWCGROUP,
	}*/

	cmd.Dir = rootFS.MergedDir
	cmd.Env = append(cmd.Env, envs...)

	rPipe, wPipe, err := os.Pipe()
	if err != nil {
		return nil, nil, xerrors.Errorf("create pipe failed: %w", err)
	}

	cmd.ExtraFiles = append(cmd.ExtraFiles, rPipe)

	return cmd, wPipe, nil
}

func InitProcess() error {
	rPipe := os.NewFile(3, "pipe")

	cInfo := new(info.Info)

	if err := json.NewDecoder(rPipe).Decode(cInfo); err != nil {
		return xerrors.Errorf("decode container info failed: %w", err)
	}
	_ = rPipe.Close()

	if err := pivotRoot.PivotRoot(cInfo.RootFS.MergedDir); err != nil {
		return xerrors.Errorf("pivot root failed: %w", err)
	}

	if err := syscall.Sethostname([]byte(cInfo.ID)); err != nil {
		return xerrors.Errorf("set hostname failed: %w", err)
	}

	execPath, err := exec.LookPath(cInfo.Cmd[0])
	if err != nil {
		return xerrors.Errorf("look exec path failed: %w", err)
	}

	if err := syscall.Exec(execPath, cInfo.Cmd, os.Environ()); err != nil {
		return xerrors.Errorf("exec failed: %w", err)
	}

	return nil
}
