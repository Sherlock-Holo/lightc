package rootfs

import (
	"io"
	"os"
	"path/filepath"

	"github.com/Sherlock-Holo/lightc/libstorage/errors"
	"github.com/Sherlock-Holo/lightc/paths"
	"golang.org/x/xerrors"
)

func Create(imageName string) (*RootFS, error) {
	id := generateInfoID()

	stat, err := os.Stat(filepath.Join(paths.ImagesPath, imageName))
	switch {
	case os.IsNotExist(err):
		return nil, errors.ImageNotFound{NotFoundName: imageName}

	default:
		return nil, xerrors.Errorf("get image stat failed: %w", err)

	case err == nil:
		if !stat.IsDir() {
			return nil, xerrors.Errorf("%s is an invalid file: %w", imageName, err)
		}
	}

	_, err = os.Stat(filepath.Join(paths.RootFSPath, id))
	switch {
	case err == nil:
		return nil, errors.RootFSCreateConflict{ConflictID: id}

	default:
		return nil, xerrors.Errorf("get rootfs stat failed: %w", err)

	case os.IsNotExist(err):
	}

	rootFS := &RootFS{
		ID:        id,
		ImageName: imageName,
		LowerDir:  filepath.Join(paths.ImagesPath, imageName),
		UpperDir:  filepath.Join(paths.RootFSPath, id, paths.UpperFile),
		WorkDir:   filepath.Join(paths.RootFSPath, id, paths.WorkFile),
		MergedDir: filepath.Join(paths.RootFSPath, id, paths.MergedFile),

		dnsFile:      filepath.Join(paths.RootFSPath, id, paths.DnsFile),
		hostnameFile: filepath.Join(paths.RootFSPath, id, paths.HostnameFile),
		Hosts:        filepath.Join(paths.RootFSPath, id, paths.HostsFile),
	}

	for _, dir := range []string{rootFS.UpperDir, rootFS.WorkDir, rootFS.MergedDir} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, xerrors.Errorf("mkdir overlay2 dir failed: %w", err)
		}
	}

	if f, err := os.OpenFile(rootFS.dnsFile, os.O_CREATE|os.O_RDWR, 0644); err != nil {
		return nil, xerrors.Errorf("create dns file failed: %w", err)
	} else {
		defer func() {
			_ = f.Close()
		}()

		originF, err := os.Open("/etc/resolv.conf")
		if err != nil {
			return nil, xerrors.Errorf("open host /etc/resolv.conf failed: %w", err)
		}
		defer func() {
			_ = originF.Close()
		}()

		if _, err := io.Copy(f, originF); err != nil {
			return nil, xerrors.Errorf("write dns file failed: %w", err)
		}
	}

	if f, err := os.OpenFile(rootFS.hostnameFile, os.O_CREATE|os.O_RDWR, 0644); err != nil {
		return nil, xerrors.Errorf("create hostname file failed: %w", err)
	} else {
		defer func() {
			_ = f.Close()
		}()

		if _, err := f.WriteString(rootFS.ID + "\n"); err != nil {
			return nil, xerrors.Errorf("write hostname file failed: %w", err)
		}
	}

	if f, err := os.OpenFile(rootFS.Hosts, os.O_CREATE|os.O_RDWR, 0644); err != nil {
		return nil, xerrors.Errorf("create hosts file failed: %w", err)
	} else {
		defer func() {
			_ = f.Close()
		}()

		ew := &errorWriter{w: f}

		_, _ = ew.WriteString("127.0.0.1	localhost\n")
		_, _ = ew.WriteString("::1	localhost ip6-localhost ip6-loopback\n")
		_, _ = ew.WriteString("fe00::0	ip6-localnet\n")
		_, _ = ew.WriteString("ff00::0	ip6-mcastprefix\n")
		_, _ = ew.WriteString("ff02::1	ip6-allnodes\n")
		if _, err := ew.WriteString("ff02::2	ip6-allrouters\n"); err != nil {
			return nil, xerrors.Errorf("write hosts file failed: %w", err)
		}
	}

	return rootFS, nil
}

type writable interface {
	io.Writer
	io.StringWriter
}

type errorWriter struct {
	err error
	w   writable
}

func (ew *errorWriter) WriteString(s string) (n int, err error) {
	if ew.err != nil {
		return 0, err
	}

	n, err = ew.w.WriteString(s)
	err = ew.err

	return n, err
}

func (ew *errorWriter) Write(p []byte) (n int, err error) {
	if ew.err != nil {
		return 0, err
	}

	n, err = ew.w.Write(p)
	err = ew.err

	return n, err
}
