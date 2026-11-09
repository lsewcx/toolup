package installer

import (
	"context"
	"fmt"

	"github.com/lsewcx/toolup/internal/execx"
)

type DNF struct {
	Runner execx.Runner
}

func (d *DNF) Install(ctx context.Context, pkg string) error {
	if pkg == "" {
		return fmt.Errorf("dnf 包名不能为空")
	}

	if err := d.Runner.RunElevated(ctx, "dnf", "install", "-y", pkg); err != nil {
		return fmt.Errorf("dnf 安装 %s 失败: %w", pkg, err)
	}
	return nil
}
