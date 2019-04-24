package paths

import "path/filepath"

const (
	LightcDir   = "/var/lib/lightc"
	UnixSockDir = "/var/run/lightc"

	ConfigName = "config.json"
	UpperFile  = "diff"
	WorkFile   = "work"
	MergedFile = "merged"
	LogFile    = "container.log"
	SockFile   = "%s.sock"
)

var (
	NetworkPath     = filepath.Join(LightcDir, "network")
	BridgePath      = filepath.Join(NetworkPath, "bridge")
	IPAllocatorPath = filepath.Join(NetworkPath, "ipam/subnet.json")
	ImagesPath      = filepath.Join(LightcDir, "images")
	RootFSPath      = filepath.Join(LightcDir, "rootfs")
)
