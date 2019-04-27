package ipam

import (
	"encoding/json"
	"net"
	"os"
	"strings"
	"syscall"

	"github.com/Sherlock-Holo/lightc/paths"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

type IPAM struct {
	SubnetAllocatorPath string
	SubnetMap           map[string]*subnetConfig `json:"subnet_map"`
}

type subnetConfig struct {
	IPBitmap string `json:"ip_bitmap"`
}

var IPAllAllocator = &IPAM{SubnetAllocatorPath: paths.IPAllocatorPath, SubnetMap: make(map[string]*subnetConfig)}

func (ipam *IPAM) load(f *os.File) error {
	ipam.SubnetMap = make(map[string]*subnetConfig)

	if err := json.NewDecoder(f).Decode(&ipam.SubnetMap); err != nil {
		return xerrors.Errorf("decode subnet allocator config failed: %w", err)
	}

	return nil
}

func (ipam *IPAM) dump(f *os.File) error {
	if err := f.Truncate(0); err != nil {
		return xerrors.Errorf("truncate ipam config file failed: %w", err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		return xerrors.Errorf("seek ipam config file to head failed: %w", err)
	}

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "    ")

	if err := encoder.Encode(ipam.SubnetMap); err != nil {
		return xerrors.Errorf("encode subnet allocator config failed: %w", err)
	}

	return nil
}

func (ipam *IPAM) Allocate(subnet net.IPNet) (ip net.IP, err error) {
	// open ipam file
	loadFile, err := os.OpenFile(ipam.SubnetAllocatorPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, xerrors.Errorf("open ipam file failed: %w", err)
	}
	defer func() {
		_ = loadFile.Close()
	}()

	// lock ipam file
	if err := syscall.Flock(int(loadFile.Fd()), syscall.LOCK_EX); err != nil {
		return nil, xerrors.Errorf("lock ipam file failed: %w", err)
	}
	defer func() {
		if err := syscall.Flock(int(loadFile.Fd()), syscall.LOCK_UN); err != nil {
			logrus.Error(xerrors.Errorf("unlock ipam file failed: %w", err))
		}
	}()

	if err := ipam.load(loadFile); err != nil {
		return nil, xerrors.Errorf("load ipam file failed: %w", err)
	}

	if _, ok := ipam.SubnetMap[subnet.String()]; !ok {
		return nil, nil
	}

	// allocate
	ip = ipam.allocate(subnet)

	// dump ipam file
	if err := ipam.dump(loadFile); err != nil {
		return nil, xerrors.Errorf("dump ipam file failed: %w", err)
	}

	return ip, nil
}

func (ipam *IPAM) Release(subnet net.IPNet, ip net.IP) (err error) {
	// open ipam file
	loadFile, err := os.OpenFile(ipam.SubnetAllocatorPath, os.O_RDWR, 0600)
	if err != nil {
		return xerrors.Errorf("open ipam file failed: %w", err)
	}
	defer func() {
		_ = loadFile.Close()
	}()

	// lock ipam file
	if err := syscall.Flock(int(loadFile.Fd()), syscall.LOCK_EX); err != nil {
		return xerrors.Errorf("lock ipam file failed: %w", err)
	}
	defer func() {
		if err := syscall.Flock(int(loadFile.Fd()), syscall.LOCK_UN); err != nil {
			logrus.Error(xerrors.Errorf("unlock ipam file failed: %w", err))
		}
	}()

	// load ipam file
	if err := ipam.load(loadFile); err != nil {
		return xerrors.Errorf("load ipam file failed: %w", err)
	}

	if _, ok := ipam.SubnetMap[subnet.String()]; !ok {
		return xerrors.Errorf("subnet %s doesn't exist", subnet.String())
	}

	// release
	releaseIP := net.ParseIP(ip.String())
	releaseIP = releaseIP.To4()
	releaseIP[3]--

	var index int
	for j := 4; j > 0; j-- {
		index += int((releaseIP[j-1] - subnet.IP[j-1]) << uint((4-j)*8))
	}

	subnetStr := subnet.String()

	alloc := []rune(ipam.SubnetMap[subnetStr].IPBitmap)
	alloc[index] = '0'
	ipam.SubnetMap[subnetStr].IPBitmap = string(alloc)

	// dump ipam file
	if err := ipam.dump(loadFile); err != nil {
		return xerrors.Errorf("dump ipam file failed: %w", err)
	}

	return
}

func (ipam *IPAM) DeleteSubnet(subnet net.IPNet) error {
	// open ipam file
	loadFile, err := os.OpenFile(ipam.SubnetAllocatorPath, os.O_RDWR, 0600)
	if err != nil {
		return xerrors.Errorf("open ipam file failed: %w", err)
	}
	defer func() {
		_ = loadFile.Close()
	}()

	// lock ipam file
	if err := syscall.Flock(int(loadFile.Fd()), syscall.LOCK_EX); err != nil {
		return xerrors.Errorf("lock ipam file failed: %w", err)
	}
	defer func() {
		if err := syscall.Flock(int(loadFile.Fd()), syscall.LOCK_UN); err != nil {
			logrus.Error(xerrors.Errorf("unlock ipam file failed: %w", err))
		}
	}()

	// load ipam file
	if err := ipam.load(loadFile); err != nil {
		return xerrors.Errorf("load ipam file failed: %w", err)
	}

	// delete subnet
	var (
		alloc     *subnetConfig
		ok        bool
		subnetStr = subnet.String()
	)

	if alloc, ok = ipam.SubnetMap[subnetStr]; !ok {
		return xerrors.Errorf("subnet %s doesn't exist", subnetStr)
	}

	if strings.Count(alloc.IPBitmap, "1") > 1 {
		return xerrors.Errorf("subnet %s is used", subnetStr)
	}

	delete(ipam.SubnetMap, subnetStr)

	// dump ipam file
	if err := ipam.dump(loadFile); err != nil {
		return xerrors.Errorf("dump ipam file failed: %w", err)
	}

	return nil
}

func (ipam *IPAM) AllocateSubnet(subnet net.IPNet) (ip net.IP, exist bool, err error) {
	// open ipam file
	loadFile, err := os.OpenFile(ipam.SubnetAllocatorPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, false, xerrors.Errorf("open ipam file failed: %w", err)
	}
	defer func() {
		_ = loadFile.Close()
	}()

	// lock ipam file
	if err := syscall.Flock(int(loadFile.Fd()), syscall.LOCK_EX); err != nil {
		return nil, false, xerrors.Errorf("lock ipam file failed: %w", err)
	}
	defer func() {
		if err := syscall.Flock(int(loadFile.Fd()), syscall.LOCK_UN); err != nil {
			logrus.Error(xerrors.Errorf("unlock ipam file failed: %w", err))
		}
	}()

	stat, err := loadFile.Stat()
	if err != nil {
		return nil, false, xerrors.Errorf("get ipam stat failed: %w", err)
	}
	if stat.Size() > 0 {
		// load ipam
		if err := ipam.load(loadFile); err != nil {
			return nil, false, xerrors.Errorf("load ipam file failed: %w", err)
		}
	}

	if _, ok := ipam.SubnetMap[subnet.String()]; ok {
		return nil, true, nil
	}

	ip = ipam.allocate(subnet)

	// dump ipam file
	if err := ipam.dump(loadFile); err != nil {
		logrus.Error("dump ipam file failed: %w", err)
	}

	return ip, false, nil
}

func (ipam *IPAM) allocate(subnet net.IPNet) (ip net.IP) {
	subnet.IP = subnet.IP.To4()

	subnetStr := subnet.String()

	one, size := subnet.Mask.Size()

	if _, ok := ipam.SubnetMap[subnetStr]; !ok {
		ipam.SubnetMap[subnetStr] = &subnetConfig{
			IPBitmap: strings.Repeat("0", 1<<uint(size-one)),
		}
	}

	for i := range ipam.SubnetMap[subnetStr].IPBitmap {
		if ipam.SubnetMap[subnetStr].IPBitmap[i] == '0' {
			alloc := []rune(ipam.SubnetMap[subnetStr].IPBitmap)
			alloc[i] = '1'
			ipam.SubnetMap[subnetStr].IPBitmap = string(alloc)

			ip = make(net.IP, net.IPv4len)
			copy(ip, subnet.IP)

			for j := 4; j > 0; j-- {
				ip[4-j] += uint8(i >> uint((j-1)*8))
			}

			ip[3]++
			break
		}
	}

	return ip
}
