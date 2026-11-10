package app

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/lsewcx/toolup/internal/execx"
)

type call struct {
	name string
	args []string
}

type fakeRunner struct {
	calls []call
}

func (f *fakeRunner) Run(_ context.Context, name string, args ...string) error {
	f.calls = append(f.calls, call{name: name, args: args})
	return nil
}

func (f *fakeRunner) RunElevated(_ context.Context, name string, args ...string) error {
	f.calls = append(f.calls, call{name: name, args: args})
	return nil
}

func TestInstallFallsBackToScript(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("官方安装脚本只支持 Linux/macOS")
	}

	// 构造一个只有 curl 的 PATH：uv 不存在，下载工具存在。
	path := t.TempDir()
	writeExecutable(t, filepath.Join(path, "curl"))
	t.Setenv("PATH", path)

	runner := &fakeRunner{}
	if err := Install(context.Background(), "uv", runner); err != nil {
		t.Fatal(err)
	}

	if len(runner.calls) != 1 {
		t.Fatalf("calls = %#v, want exactly one call", runner.calls)
	}

	got := runner.calls[0]
	if got.name != "sh" || len(got.args) != 2 || got.args[0] != "-c" {
		t.Fatalf("call = %#v, want sh -c <script>", got)
	}
	for _, want := range []string{"curl -LsSf", "https://astral.sh/uv/install.sh", "| sh"} {
		if !strings.Contains(got.args[1], want) {
			t.Errorf("script does not contain %q:\n%s", want, got.args[1])
		}
	}
}

func TestInstallFailsWithoutDownloader(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("官方安装脚本只支持 Linux/macOS")
	}

	// PATH 里既没有 uv，也没有 curl/wget。
	t.Setenv("PATH", t.TempDir())

	err := Install(context.Background(), "uv", &fakeRunner{})
	if err == nil {
		t.Fatal("expected an error when curl and wget are both missing")
	}
	if !strings.Contains(err.Error(), "curl") || !strings.Contains(err.Error(), "wget") {
		t.Errorf("error should mention curl and wget: %v", err)
	}
}

func TestInstallRejectsUnknownTool(t *testing.T) {
	if err := Install(context.Background(), "unknown-tool", &fakeRunner{}); err == nil {
		t.Fatal("expected an error for an unknown tool")
	}
}

func writeExecutable(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
}

var _ execx.Runner = (*fakeRunner)(nil)
