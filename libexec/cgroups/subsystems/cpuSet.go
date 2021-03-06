package subsystems

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"golang.org/x/xerrors"
)

type CPUSetSubSystem struct{}

func (cs *CPUSetSubSystem) Name() string {
	return "cpuset"
}

func (cs *CPUSetSubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	subSysCgroupPath, err := getCgroupPath(cs.Name(), cgroupPath, true)
	if err != nil {
		return err
	}

	if res.CpuShare != "" {
		if err := ioutil.WriteFile(filepath.Join(subSysCgroupPath, "cpuset.cpus"), []byte(res.CpuShare), 0644); err != nil {
			return xerrors.Errorf("set cgroup cpu share failed: %w", err)
		}
	}

	return nil
}

func (cs *CPUSetSubSystem) Apply(path string, pid int) error {
	subSysCgroupPath, err := getCgroupPath(cs.Name(), path, false)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(subSysCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
		return xerrors.Errorf("set cgroup proc failed: %w", err)
	}

	return nil
}

func (cs *CPUSetSubSystem) Remove(path string) error {
	subSysCgroupPath, err := getCgroupPath(cs.Name(), path, false)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(subSysCgroupPath); err != nil {
		return xerrors.Errorf("remove subsystem failed: %w", err)
	}
	return nil
}
