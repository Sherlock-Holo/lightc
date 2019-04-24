package libnetwork

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/Sherlock-Holo/lightc/info"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"golang.org/x/xerrors"
)

func enterNetns(iface netlink.Link, containerInfo *info.Info) (exitNetns func(), err error) {
	file, err := os.Open(filepath.Join("/proc", strconv.Itoa(containerInfo.Pid), "ns/net"))
	if err != nil {
		return nil, xerrors.Errorf("open /proc netns failed: %w", err)
	}

	fd := file.Fd()

	runtime.LockOSThread()

	if err := netlink.LinkSetNsFd(iface, int(fd)); err != nil {
		return nil, xerrors.Errorf("set iface netns failed: %w", err)
	}

	originNetns, err := netns.Get()

	if err := netns.Set(netns.NsHandle(fd)); err != nil {
		return nil, xerrors.Errorf("set current thread netns failed: %w", err)
	}

	return func() {
		if err := netns.Set(originNetns); err != nil {
			logrus.Error("recover origin netns failed: %w", err)
		}

		_ = originNetns.Close()
		runtime.UnlockOSThread()
		_ = file.Close()
	}, nil
}
