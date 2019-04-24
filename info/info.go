package info

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/Sherlock-Holo/lightc/libstorage/rootfs"
	"github.com/Sherlock-Holo/lightc/libstorage/volume"
	"github.com/Sherlock-Holo/lightc/paths"
	"golang.org/x/xerrors"
)

const (
	RUNNING = "running"
	STOPPED = "stopped"

	timeFormat = "2006-1-2 15:04:05 MST"
)

type Info struct {
	Pid           int             `json:"pid"`
	ID            string          `json:"id"`
	Cmd           []string        `json:"cmd"`
	CreateTime    CustomTime      `json:"create_time"`
	Status        string          `json:"status"`
	RootFS        *rootfs.RootFS  `json:"root_fs"`
	LogFile       string          `json:"log_file"`
	PortMap       []string        `json:"port_map"`
	Network       string          `json:"network"`
	IPNet         net.IPNet       `json:"ip_net"`
	NetworkDriver string          `json:"network_driver"`
	Volumes       []volume.Volume `json:"volumes"`
	UnixSocket    string          `json:"unix_socket"`
	TTY           bool            `json:"tty"`
	Detach        bool            `json:"detach"`
	RmAfterRun    bool            `json:"rm_after_run"`
	Envs          []string        `json:"envs"`

	Parent     *exec.Cmd       `json:"-"`
	StoppedCtx context.Context `json:"-"`
	Stdin      io.WriteCloser  `json:"-"`
	Stdout     io.ReadCloser   `json:"-"`
	Stderr     io.ReadCloser   `json:"-"`
}

type CustomTime time.Time

func (ct CustomTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(ct).Format(timeFormat))
}

func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}

	t, err := time.Parse(timeFormat, str)
	if err != nil {
		return err
	}
	*ct = CustomTime(t)

	return nil
}

func (ct CustomTime) String() string {
	return time.Time(ct).Format(timeFormat)
}

func GetInfo(containerID string) (*Info, error) {
	b, err := ioutil.ReadFile(filepath.Join(paths.LightcDir, containerID, paths.ConfigName))
	if err != nil {
		return nil, xerrors.Errorf("read config file %s failed: %w", containerID, err)
	}

	info := new(Info)

	if err := json.Unmarshal(b, info); err != nil {
		return nil, xerrors.Errorf("decode config failed: %w", err)
	}

	return info, nil
}
