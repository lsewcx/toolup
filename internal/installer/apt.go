package installer

import (
	"context"
	"fmt"

	"github.com/lsewcx/toolup/internal/execx"
)

type APT struct {
	Runner execx.Runner
}

func (a *APT) Install(ctx context.Context, pkg string) error {
	if pkg == "" {
		return fmt.Errorf("apt 包名不能为空")
	}

	if err := a.Runner.RunElevated(ctx, "apt-get", "install", "-y", "--no-install-recommends", pkg); err != nil {
		return fmt.Errorf("apt 安装 %s 失败: %w", pkg, err)
	}

	return nil
}
