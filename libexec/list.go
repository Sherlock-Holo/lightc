package libexec

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/Sherlock-Holo/lightc/info"
	"github.com/Sherlock-Holo/lightc/paths"
	"golang.org/x/xerrors"
)

func List(all bool) ([]info.Info, error) {
	stats, err := ioutil.ReadDir(paths.RootFSPath)
	if err != nil {
		return nil, xerrors.Errorf("read containers stat failed: %w", err)
	}

	cInfos := make([]info.Info, 0, len(stats))

	for _, stat := range stats {
		var cInfo info.Info
		b, err := ioutil.ReadFile(filepath.Join(paths.RootFSPath, stat.Name(), paths.ConfigName))
		if err != nil {
			return nil, xerrors.Errorf("read container %s config failed: %w", stat.Name(), err)
		}

		if err := json.Unmarshal(b, &cInfo); err != nil {
			return nil, xerrors.Errorf("decode container %s config failed: %w", stat.Name(), err)
		}

		if cInfo.Status == info.RUNNING || all {
			cInfos = append(cInfos, cInfo)
		}
	}

	return cInfos, nil
}
