package volume

import (
	"path/filepath"
	"strings"
)

type Volume struct {
	HostDir      string
	ContainerDir string
}

func Parse(vs []string, rootfsMergedDir string) []Volume {
	volumes := make([]Volume, 0, len(vs))

	for _, s := range vs {
		split := strings.Split(s, ":")

		hostDir := split[0]
		var containerDir string
		if len(split) > 1 {
			containerDir = filepath.Join(rootfsMergedDir, split[1])
		} else {
			containerDir = filepath.Join(rootfsMergedDir, hostDir)
		}

		volumes = append(volumes, Volume{
			HostDir:      hostDir,
			ContainerDir: containerDir,
		})
	}

	return volumes
}
