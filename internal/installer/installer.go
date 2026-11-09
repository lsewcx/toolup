package installer

import (
	"context"
	"fmt"

	"github.com/lsewcx/toolup/internal/execx"
)

type Installer interface {
	Install(ctx context.Context, pkg string) error
}

func New(packageManager string, runner execx.Runner) (Installer, error) {
	switch packageManager {
	case "apt":
		return &APT{Runner: runner}, nil
	case "dnf":
		return &DNF{Runner: runner}, nil
	default:
		return nil, fmt.Errorf("不支持的包管理器: %s", packageManager)
	}
}
