package app

import (
	"context"
	"fmt"

	"github.com/lsewcx/toolup/internal/execx"
	"github.com/lsewcx/toolup/internal/installer"
	"github.com/lsewcx/toolup/internal/installer/catalog"
	"github.com/lsewcx/toolup/internal/utils"
)

func Install(ctx context.Context, toolName string, runner execx.Runner) error {
	tool, ok := catalog.Lookup(toolName)
	if !ok {
		return fmt.Errorf("不支持的工具: %s", toolName)
	}

	info, err := utils.GetOSInfo()
	if err != nil {
		return fmt.Errorf("检测系统信息: %w", err)
	}

	// 优先使用系统包管理器；包管理器未收录该工具时回退到官方安装脚本。
	if pkg, ok := tool.Packages[info.PackageManager]; ok {
		inst, err := installer.New(info.PackageManager, runner)
		if err != nil {
			return err
		}
		return inst.Install(ctx, pkg)
	}

	if tool.InstallScript != "" {
		script := &installer.ScriptInstaller{
			Runner: runner,
			Tool:   tool.Name,
			URL:    tool.InstallScript,
		}
		return script.Install(ctx)
	}

	return fmt.Errorf("工具 %s 不支持包管理器 %s", toolName, info.PackageManager)
}
