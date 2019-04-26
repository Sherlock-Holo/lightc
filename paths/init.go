package paths

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

func init() {
	for _, p := range []string{LightcDir, UnixSockDir, NetworkPath, BridgePath, ImagesPath, RootFSPath, filepath.Dir(IPAllocatorPath)} {
		if err := os.MkdirAll(p, 0700); err != nil {
			logrus.Fatal(xerrors.Errorf("mkdir -p %s failed: %w", p, err))
		}
	}
}
