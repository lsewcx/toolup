package app

import (
	"context"
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
	// 清空 PATH，确保不会因为本机已安装 uv 而跳过安装。
	t.Setenv("PATH", t.TempDir())

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
	if !strings.Contains(got.args[1], "https://astral.sh/uv/install.sh") {
		t.Errorf("script does not install uv:\n%s", got.args[1])
	}
}

func TestInstallRejectsUnknownTool(t *testing.T) {
	if err := Install(context.Background(), "unknown-tool", &fakeRunner{}); err == nil {
		t.Fatal("expected an error for an unknown tool")
	}
}

var _ execx.Runner = (*fakeRunner)(nil)
