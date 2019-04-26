package subsystems

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/xerrors"
)

type ResourceConfig struct {
	MemoryLimit string
	CpuShare    string
	CpuSet      string
}

type Subsystem interface {
	Name() string
	Set(cgroupPath string, res *ResourceConfig) error
	Apply(path string, pid int) error
	Remove(path string) error
}

var (
	Instances = []Subsystem{
		new(CPUSubSystem),
		new(CPUSetSubSystem),
		new(MemorySubSystem),
	}
)

func findCgroupMountPoint(subsystem string) string {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return ""
	}
	defer func() {
		_ = f.Close()
	}()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text := scanner.Text()
		fields := strings.Split(text, " ")
		for _, opt := range strings.Split(fields[len(fields)-1], ",") {
			if opt == subsystem {
				return fields[4]
			}
		}
	}

	// scanner may has error
	return ""
}

func getCgroupPath(subsystem, cgroupPath string, autoCreate bool) (string, error) {
	cgroupRoot := findCgroupMountPoint(subsystem)
	_, err := os.Stat(filepath.Join(cgroupRoot, cgroupPath))

	switch {
	case os.IsNotExist(err) && autoCreate:
		if err := os.Mkdir(filepath.Join(cgroupRoot, cgroupPath), 0755); err != nil {
			return "", xerrors.Errorf("create cgroup failed: %w", err)
		}

	default:
		return "", xerrors.Errorf("create cgroup failed: %w", err)

	case err == nil:
	}

	return filepath.Join(cgroupRoot, cgroupPath), nil
}
