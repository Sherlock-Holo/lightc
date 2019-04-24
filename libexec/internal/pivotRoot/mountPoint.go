package pivotRoot

import (
	"os"
	"syscall"
)

const defaultMountFlags = syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV

type mountPoint struct {
	Src    string
	Dst    string
	Mode   os.FileMode
	IsDir  bool
	fsType string
	Flags  uintptr
	Data   string
}
