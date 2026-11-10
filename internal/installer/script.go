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

	// LookPath 用于检测命令是否存在，默认使用 exec.LookPath。
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

	downloader, err := detectDownloader(lookPath)
	if err != nil {
		slog.Warn("缺少下载工具，无法安装 "+s.Tool,
			"hint", "请先安装 curl 或 wget，例如：apt install curl 或 dnf install wget")
		return err
	}

	// 有 curl 时优先使用 curl，否则回退到 wget 下载官方安装脚本。
	var command string
	if downloader == "curl" {
		command = fmt.Sprintf("curl -LsSf %s | sh", s.URL)
	} else {
		command = fmt.Sprintf("wget -qO- %s | sh", s.URL)
	}

	slog.Info("使用官方脚本安装", "tool", s.Tool, "script", s.URL, "downloader", downloader)
	if err := s.Runner.Run(ctx, "sh", "-c", command); err != nil {
		return fmt.Errorf("安装 %s 失败: %w", s.Tool, err)
	}

	// 官方脚本一般会把可执行文件放进 ~/.local/bin 并写入 shell 配置，
	// 但当前进程的 PATH 不会随之更新，因此这里提示用户重新加载配置。
	slog.Info("安装完成，如果当前终端找不到该命令，请重新打开终端或重新加载 shell 配置", "tool", s.Tool)
	return nil
}

// detectDownloader 返回当前系统可用的下载工具，优先 curl，其次 wget。
// 两者都不存在时返回提示安装的错误。
func detectDownloader(lookPath func(string) (string, error)) (string, error) {
	for _, name := range []string{"curl", "wget"} {
		if _, err := lookPath(name); err == nil {
			return name, nil
		}
	}
	return "", fmt.Errorf("未找到 curl 或 wget，请先安装其中之一后重试（例如：apt install curl 或 dnf install wget）")
}
