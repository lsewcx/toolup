package installer

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/lsewcx/toolup/internal/execx"
)

// ScriptInstaller 通过工具官方提供的安装脚本完成安装。
//
// 适合系统包管理器未收录的工具（例如 uv、nvm）。官方脚本通常不需要 root 权限，
// 会把可执行文件安装到用户目录（多为 ~/.local/bin 或 ~/.nvm），并写入 shell 配置。
type ScriptInstaller struct {
	Runner execx.Runner
	// Tool 是工具名，同时也是安装后的可执行文件名，用于判断是否已经安装。
	Tool string
	// URL 是官方安装脚本的地址，可包含 {version} 占位符。
	URL string
	// LatestURL 非空时先请求它解析最新版本号，
	// 并用解析结果替换 URL 中的 {version} 占位符（例如 nvm 的最新 release）。
	LatestURL string
	// Shell 是执行安装脚本使用的 shell，为空时默认 sh。
	// 官方脚本要求 bash 时需要显式指定（例如 nvm）。
	Shell string
	// InstalledPaths 是判断工具是否已安装的文件路径列表，支持环境变量。
	// 为空时使用 exec.LookPath 检测 Tool 可执行文件。
	InstalledPaths []string

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

	shell := s.Shell
	if shell == "" {
		shell = "sh"
	}

	if path, ok := s.installedPath(lookPath); ok {
		slog.Info("工具已安装，跳过", "tool", s.Tool, "path", path)
		return nil
	}

	downloader, err := detectDownloader(lookPath)
	if err != nil {
		slog.Warn("缺少下载工具，无法安装 "+s.Tool,
			"hint", "请先安装 curl 或 wget，例如：apt install curl 或 dnf install wget")
		return err
	}

	if _, err := lookPath(shell); err != nil {
		slog.Warn("缺少安装脚本所需的 shell，无法安装 "+s.Tool,
			"shell", shell,
			"hint", "请先安装 "+shell+"，例如：apt install "+shell+" 或 dnf install "+shell)
		return fmt.Errorf("未找到 %s，请先安装后重试", shell)
	}

	scriptURL, err := s.scriptURL(ctx)
	if err != nil {
		slog.Warn("解析最新版本失败，无法安装 "+s.Tool, "latest_url", s.LatestURL, "err", err)
		return err
	}

	// 有 curl 时优先使用 curl，否则回退到 wget 下载官方安装脚本。
	var command string
	if downloader == "curl" {
		command = fmt.Sprintf("curl -LsSf %s | %s", scriptURL, shell)
	} else {
		command = fmt.Sprintf("wget -qO- %s | %s", scriptURL, shell)
	}

	slog.Info("使用官方脚本安装", "tool", s.Tool, "script", scriptURL, "downloader", downloader, "shell", shell)
	if err := s.Runner.Run(ctx, shell, "-c", command); err != nil {
		return fmt.Errorf("安装 %s 失败: %w", s.Tool, err)
	}

	// 官方脚本一般会把可执行文件放进 ~/.local/bin（或 ~/.nvm）并写入 shell 配置，
	// 但当前进程的 PATH 不会随之更新，因此这里提示用户重新加载配置。
	slog.Info("安装完成，如果当前终端找不到该命令，请重新打开终端或重新加载 shell 配置", "tool", s.Tool)
	return nil
}

// scriptURL 返回实际使用的安装脚本地址。
// 配置了 LatestURL 时先解析最新版本号，再替换 URL 中的 {version} 占位符。
func (s *ScriptInstaller) scriptURL(ctx context.Context) (string, error) {
	if s.LatestURL == "" {
		return s.URL, nil
	}

	version, err := resolveLatestVersion(ctx, s.LatestURL)
	if err != nil {
		return "", fmt.Errorf("获取 %s 最新版本失败: %w", s.Tool, err)
	}

	slog.Info("解析到最新版本", "tool", s.Tool, "version", version)
	return strings.ReplaceAll(s.URL, "{version}", version), nil
}

// resolveLatestVersion 请求 latestURL 并跟随重定向，从最终 URL 的最后一段解析版本号。
// 例如 GitHub 的 releases/latest 会重定向到 releases/tag/v0.41.0。
func resolveLatestVersion(ctx context.Context, latestURL string) (string, error) {
	client := &http.Client{Timeout: 15 * time.Second}

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, latestURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("请求 %s 返回 HTTP %s", latestURL, resp.Status)
	}

	version := path.Base(resp.Request.URL.Path)
	if version == "" || version == "." || version == "/" || version == "latest" {
		return "", fmt.Errorf("无法从 %s 解析出版本号", resp.Request.URL)
	}
	return version, nil
}

// installedPath 判断工具是否已安装，返回已安装的路径。
// 配置了 InstalledPaths 时按文件是否存在判断，否则按可执行文件是否在 PATH 中判断。
func (s *ScriptInstaller) installedPath(lookPath func(string) (string, error)) (string, bool) {
	for _, candidate := range s.InstalledPaths {
		installed := os.ExpandEnv(candidate)
		if installed == "" {
			continue
		}
		if _, err := os.Stat(installed); err == nil {
			return installed, true
		}
	}

	if len(s.InstalledPaths) > 0 {
		return "", false
	}

	if executable, err := lookPath(s.Tool); err == nil {
		return executable, true
	}
	return "", false
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
