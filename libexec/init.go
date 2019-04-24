package libexec

import (
	"github.com/Sherlock-Holo/lightc/libexec/internal/process"
	"golang.org/x/xerrors"
)

func InitPorcess() error {
	if err := process.InitProcess(); err != nil {
		return xerrors.Errorf("init container process failed: %w", err)
	}

	return nil
}
