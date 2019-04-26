package monitor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"

	"github.com/Sherlock-Holo/lightc/info"
	"github.com/Sherlock-Holo/lightc/libexec/internal/bufReadWriteCloser"
	"github.com/Sherlock-Holo/lightc/libexec/internal/nonOpWriteCloser"
	"github.com/Sherlock-Holo/lightc/libexec/internal/process"
	"github.com/Sherlock-Holo/lightc/libexec/resources"
	"github.com/Sherlock-Holo/lightc/libnetwork"
	"github.com/Sherlock-Holo/lightc/paths"
	"github.com/kr/pty"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

func Monitor() error {
	monitorPipeR := os.NewFile(3, "pipe")
	defer func() {
		_ = monitorPipeR.Close()
	}()

	cInfo := new(info.Info)
	if err := json.NewDecoder(monitorPipeR).Decode(cInfo); err != nil {
		return xerrors.Errorf("decode container info failed: %w", err)
	}

	parent, wPipe, err := process.NewParentProcess(
		cInfo.Envs,
		cInfo.RootFS,
	)
	if err != nil {
		return xerrors.Errorf("new parent process failed: %w", err)
	}

	if cInfo.TTY {
		ptm, err := pty.Start(parent)
		if err != nil {
			return xerrors.Errorf("start parent process with pty failed: %w", err)
		}

		stdout := bufReadWriteCloser.New()

		go func() {
			_, _ = io.Copy(stdout, ptm)
		}()

		cInfo.Stdin = ptm
		cInfo.Stdout = stdout
		cInfo.Stderr = ioutil.NopCloser(&bytes.Buffer{})
	} else {
		stdoutR, stdoutW, err := os.Pipe()
		if err != nil {
			return xerrors.Errorf("create stdout pipe failed: %w", err)
		}

		parent.Stdout = stdoutW

		stdout := bufReadWriteCloser.New()
		go func() {
			_, _ = io.Copy(stdout, stdoutR)
		}()

		cInfo.Stdout = stdout

		stderrR, stderrW, err := os.Pipe()
		if err != nil {
			return xerrors.Errorf("create stderr pipe failed: %w", err)
		}

		parent.Stderr = stderrW

		go func() {
			_, _ = io.Copy(stdout, stderrR)
		}()

		cInfo.Stderr = stdout

		cInfo.Stdin = nonOpWriteCloser.NonOpWriteCloser{}

		if err := parent.Start(); err != nil {
			return xerrors.Errorf("start parent process with pipe failed: %w", err)
		}
	}

	cInfo.Pid = parent.Process.Pid

	cInfo.UnixSocket = filepath.Join(paths.UnixSockDir, fmt.Sprintf(paths.SockFile, cInfo.ID))

	configFile, err := os.OpenFile(filepath.Join(paths.RootFSPath, cInfo.ID, paths.ConfigName), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		// kill parent because config file create failed
		_ = parent.Process.Kill()
		return xerrors.Errorf("create container %s config file failed: %w", cInfo.ID, err)
	}

	encoder := json.NewEncoder(configFile)
	encoder.SetIndent("", "    ")
	if err := encoder.Encode(cInfo); err != nil {
		// kill parent because config file save failed
		_ = parent.Process.Kill()
		return xerrors.Errorf("encode container %s config failed: %w", cInfo.ID, err)
	}

	// add container into network
	if cInfo.Network != "" {
		if err := libnetwork.AddContainerIntoNetwork(cInfo.Network, cInfo); err != nil {
			// kill parent because set network failed
			_ = parent.Process.Kill()
			return xerrors.Errorf("add container into network failed: %w", err)
		}
	}

	// send info to parent
	if err := json.NewEncoder(wPipe).Encode(cInfo); err != nil {
		// kill parent because send info failed
		_ = parent.Process.Kill()
		return xerrors.Errorf("send info to parent process failed: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// resources clean
	go resources.CleanResources(parent, cInfo, cInfo.RootFS, configFile, cInfo.RmAfterRun, cancel)

	listener, err := net.Listen("unix", cInfo.UnixSocket)
	if err != nil {
		// kill parent because listen unix socket failed
		_ = parent.Process.Kill()
		return xerrors.Errorf("listen container %s unix socket failed: %w", cInfo.ID, err)
	}
	defer func() {
		_ = listener.Close()
	}()

	connCh := make(chan net.Conn)
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				logrus.Error(xerrors.Errorf("accept unix socket connection failed: %w", err))
				continue
			}

			connCh <- conn
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil

		case conn := <-connCh:
			go func() {
				_, _ = io.Copy(cInfo.Stdin, conn)
			}()

			_, _ = io.Copy(conn, cInfo.Stdout)

			_ = conn.Close()
		}
	}
}
