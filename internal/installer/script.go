package installer

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"runtime"

	"github.com/lsewcx/toolup/internal/execx"
)

// ScriptInstaller 通过工具官方提供的安装脚本完成安装。
//
// 适合系统包管理器未收录的工具（例如 uv）。官方脚本通常不需要 root 权限，
// 会把可执行文件安装到用户目录（多为 ~/.local/bin），并写入 shell 配置。
type ScriptInstaller struct {
	Runner execx.Runner
	// Tool 是工具名，同时也是安装后的可执行文件名，用于判断是否已经安装。
	Tool string
	// URL 是官方安装脚本的地址。
	URL string

	// LookPath 用于检测工具是否已在 PATH 中，默认使用 exec.LookPath。
	// 抽成字段是为了方便测试时替换。
	LookPath func(string) (string, error)
}

func (s *ScriptInstaller) Install(ctx context.Context) error {
	if s.Tool == "" {
		return fmt.Errorf("工具名不能为空")
	}
	if s.URL == "" {
		return fmt.Errorf("安装脚本地址不能为空")
	}
	if runtime.GOOS == "windows" {
		return fmt.Errorf("暂不支持在 Windows 上通过脚本安装 %s", s.Tool)
	}

	lookPath := s.LookPath
	if lookPath == nil {
		lookPath = exec.LookPath
	}

	if path, err := lookPath(s.Tool); err == nil {
		slog.Info("工具已安装，跳过", "tool", s.Tool, "path", path)
		return nil
	}

	// 使用官方安装命令：curl -LsSf <script> | sh
	command := fmt.Sprintf("curl -LsSf %s | sh", s.URL)
	slog.Info("使用官方脚本安装", "tool", s.Tool, "script", s.URL)
	if err := s.Runner.Run(ctx, "sh", "-c", command); err != nil {
		return fmt.Errorf("安装 %s 失败: %w", s.Tool, err)
	}

	// 官方脚本一般会把可执行文件放进 ~/.local/bin 并写入 shell 配置，
	// 但当前进程的 PATH 不会随之更新，因此这里提示用户重新加载配置。
	slog.Info("安装完成，如果当前终端找不到该命令，请重新打开终端或重新加载 shell 配置", "tool", s.Tool)
	return nil
}
