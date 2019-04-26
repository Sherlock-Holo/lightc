package subsystems

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"golang.org/x/xerrors"
)

type MemorySubSystem struct{}

func (ms *MemorySubSystem) Name() string {
	return "memory"
}

func (ms *MemorySubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	subSysCgroupPath, err := getCgroupPath(ms.Name(), cgroupPath, true)
	if err != nil {
		return err
	}

	if res.MemoryLimit != "" {
		if err := ioutil.WriteFile(filepath.Join(subSysCgroupPath, "memory.limit_in_bytes"), []byte(res.MemoryLimit), 0644); err != nil {
			return xerrors.Errorf("set cgroup memory limit failed: %w", err)
		}
	}

	return nil
}

func (ms *MemorySubSystem) Apply(path string, pid int) error {
	subSysCgroupPath, err := getCgroupPath(ms.Name(), path, false)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(subSysCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
		return xerrors.Errorf("set cgroup proc failed: %w", err)
	}
	return nil
}

func (ms *MemorySubSystem) Remove(path string) error {
	subSysCgroupPath, err := getCgroupPath(ms.Name(), path, false)
	if err != nil {
		return err
	}

	if err := os.RemoveAll(subSysCgroupPath); err != nil {
		return xerrors.Errorf("remove subsystem failed: %w", err)
	}
	return nil
}
