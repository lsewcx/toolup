package installer

import (
	"bytes"
	"context"
	"log/slog"
	"os/exec"
	"strings"
	"testing"
)

const testScriptURL = "https://astral.sh/uv/install.sh"

// fakeLookPath 模拟 PATH 中已存在的命令集合。
func fakeLookPath(commands ...string) func(string) (string, error) {
	return func(name string) (string, error) {
		for _, command := range commands {
			if name == command {
				return "/usr/bin/" + name, nil
			}
		}
		return "", exec.ErrNotFound
	}
}

func TestScriptInstallerSkipsInstalledTool(t *testing.T) {
	runner := &fakeRunner{}
	inst := &ScriptInstaller{
		Runner:   runner,
		Tool:     "uv",
		URL:      testScriptURL,
		LookPath: fakeLookPath("uv", "curl"),
	}

	if err := inst.Install(context.Background()); err != nil {
		t.Fatal(err)
	}

	if len(runner.calls) != 0 {
		t.Fatalf("calls = %#v, want no calls when tool is installed", runner.calls)
	}
}

func TestScriptInstallerUsesCurlWhenAvailable(t *testing.T) {
	runner := &fakeRunner{}
	inst := &ScriptInstaller{
		Runner:   runner,
		Tool:     "uv",
		URL:      testScriptURL,
		LookPath: fakeLookPath("curl", "wget"),
	}

	if err := inst.Install(context.Background()); err != nil {
		t.Fatal(err)
	}

	want := "curl -LsSf " + testScriptURL + " | sh"
	assertScriptCommand(t, runner, want)
}

func TestScriptInstallerFallsBackToWget(t *testing.T) {
	runner := &fakeRunner{}
	inst := &ScriptInstaller{
		Runner:   runner,
		Tool:     "uv",
		URL:      testScriptURL,
		LookPath: fakeLookPath("wget"),
	}

	if err := inst.Install(context.Background()); err != nil {
		t.Fatal(err)
	}

	want := "wget -qO- " + testScriptURL + " | sh"
	assertScriptCommand(t, runner, want)
}

func TestScriptInstallerWarnsWhenNoDownloader(t *testing.T) {
	var logs bytes.Buffer
	original := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(original) })

	runner := &fakeRunner{}
	inst := &ScriptInstaller{
		Runner:   runner,
		Tool:     "uv",
		URL:      testScriptURL,
		LookPath: fakeLookPath(), // curl 和 wget 都不存在
	}

	err := inst.Install(context.Background())
	if err == nil {
		t.Fatal("expected an error when curl and wget are both missing")
	}
	if len(runner.calls) != 0 {
		t.Fatalf("calls = %#v, want no calls", runner.calls)
	}

	warning := logs.String()
	for _, want := range []string{"WARN", "curl", "wget"} {
		if !strings.Contains(warning, want) {
			t.Errorf("warning does not contain %q:\n%s", want, warning)
		}
	}
}

func TestScriptInstallerValidatesInput(t *testing.T) {
	tests := []struct {
		name string
		inst *ScriptInstaller
	}{
		{name: "empty tool", inst: &ScriptInstaller{Runner: &fakeRunner{}, URL: testScriptURL}},
		{name: "empty url", inst: &ScriptInstaller{Runner: &fakeRunner{}, Tool: "uv"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.inst.Install(context.Background()); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func assertScriptCommand(t *testing.T, runner *fakeRunner, want string) {
	t.Helper()

	if len(runner.calls) != 1 {
		t.Fatalf("calls = %#v, want exactly one call", runner.calls)
	}

	got := runner.calls[0]
	if got.name != "sh" {
		t.Errorf("command = %q, want sh", got.name)
	}
	if len(got.args) != 2 || got.args[0] != "-c" {
		t.Fatalf("args = %#v, want [\"-c\" script]", got.args)
	}
	if got.elevated {
		t.Error("official installer should run as the current user, not elevated")
	}
	if got.args[1] != want {
		t.Errorf("script = %q, want %q", got.args[1], want)
	}
}
