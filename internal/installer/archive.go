package installer

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/lsewcx/toolup/internal/execx"
)

// ArchiveInstaller 通过官方提供的二进制压缩包安装工具。
//
// 适用于只提供压缩包、没有安装脚本的工具（例如 ossutil 2.0）。
// 流程：下载压缩包 -> 校验 SHA256 -> 解压可执行文件 -> 安装到目标目录（必要时 sudo）。
type ArchiveInstaller struct {
	Runner execx.Runner
	// Tool 是工具名，同时也是安装后的可执行文件名，用于判断是否已经安装。
	Tool string
	// Version 是压缩包对应的工具版本，仅用于日志展示。
	Version string
	// URL 是官方压缩包下载地址。
	URL string
	// SHA256 是压缩包的期望校验和，为空时跳过校验。
	SHA256 string
	// Binary 是压缩包内需要安装的可执行文件名。
	Binary string
	// InstallDir 是二进制安装目录，例如 /usr/local/bin。
	InstallDir string
	// Symlink 是需要额外创建的软链接路径，例如 /usr/bin/ossutil，可为空。
	Symlink string

	// LookPath 用于检测命令是否存在，默认使用 exec.LookPath。
	// 抽成字段是为了方便测试时替换。
	LookPath func(string) (string, error)
}

func (a *ArchiveInstaller) Install(ctx context.Context) error {
	if a.Tool == "" {
		return fmt.Errorf("工具名不能为空")
	}
	if a.URL == "" {
		return fmt.Errorf("安装包地址不能为空")
	}
	if a.Binary == "" {
		return fmt.Errorf("可执行文件名不能为空")
	}
	if a.InstallDir == "" {
		return fmt.Errorf("安装目录不能为空")
	}

	lookPath := a.LookPath
	if lookPath == nil {
		lookPath = exec.LookPath
	}

	if path, err := lookPath(a.Tool); err == nil {
		slog.Info("工具已安装，跳过", "tool", a.Tool, "path", path)
		return nil
	}

	downloader, err := detectDownloader(lookPath)
	if err != nil {
		slog.Warn("缺少下载工具，无法安装 "+a.Tool,
			"hint", "请先安装 curl 或 wget，例如：apt install curl 或 dnf install wget")
		return err
	}

	dir, err := os.MkdirTemp("", "toolup-"+a.Tool+"-*")
	if err != nil {
		return fmt.Errorf("创建临时目录: %w", err)
	}
	defer os.RemoveAll(dir)

	archivePath := filepath.Join(dir, filepath.Base(a.URL))
	slog.Info("下载官方安装包", "tool", a.Tool, "version", a.Version, "url", a.URL)
	if err := a.download(ctx, downloader, archivePath); err != nil {
		return err
	}

	if a.SHA256 != "" {
		if err := verifySHA256(archivePath, a.SHA256); err != nil {
			return fmt.Errorf("校验 %s 安装包失败: %w", a.Tool, err)
		}
	}

	binaryPath, err := extractBinary(archivePath, dir, a.Binary)
	if err != nil {
		return fmt.Errorf("解压 %s 安装包: %w", a.Tool, err)
	}

	target := filepath.Join(a.InstallDir, a.Binary)
	slog.Info("安装可执行文件", "tool", a.Tool, "target", target)
	if err := a.Runner.RunElevated(ctx, "mv", binaryPath, target); err != nil {
		return fmt.Errorf("安装 %s 失败: %w", a.Tool, err)
	}

	if a.Symlink != "" {
		if err := a.Runner.RunElevated(ctx, "ln", "-sf", target, a.Symlink); err != nil {
			return fmt.Errorf("创建 %s 软链接失败: %w", a.Tool, err)
		}
	}

	slog.Info("安装完成", "tool", a.Tool, "version", a.Version, "path", target)
	return nil
}

func (a *ArchiveInstaller) download(ctx context.Context, downloader, path string) error {
	var args []string
	switch downloader {
	case "curl":
		args = []string{"-fsSL", a.URL, "-o", path}
	case "wget":
		args = []string{"-qO", path, a.URL}
	default:
		return fmt.Errorf("不支持的下载工具: %s", downloader)
	}

	if err := a.Runner.Run(ctx, downloader, args...); err != nil {
		return fmt.Errorf("下载 %s 失败: %w", a.URL, err)
	}
	return nil
}

// verifySHA256 校验文件的 SHA256 是否与期望值一致。
func verifySHA256(path, want string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}

	got := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(got, want) {
		return fmt.Errorf("SHA256 不匹配: 实际 %s，期望 %s", got, want)
	}
	return nil
}

// extractBinary 从压缩包中解压出指定的可执行文件，返回其临时路径。
func extractBinary(archivePath, dir, binary string) (string, error) {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", err
	}
	defer r.Close()

	for _, f := range r.File {
		if f.FileInfo().IsDir() || filepath.Base(f.Name) != binary {
			continue
		}

		src, err := f.Open()
		if err != nil {
			return "", err
		}

		dstPath := filepath.Join(dir, binary)
		dst, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o755)
		if err != nil {
			src.Close()
			return "", err
		}

		_, copyErr := io.Copy(dst, src)
		src.Close()
		closeErr := dst.Close()
		if copyErr != nil {
			return "", copyErr
		}
		if closeErr != nil {
			return "", closeErr
		}
		return dstPath, nil
	}

	return "", fmt.Errorf("压缩包中未找到 %s", binary)
}
