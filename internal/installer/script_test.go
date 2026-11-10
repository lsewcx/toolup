package installer

import (
	"context"
	"os/exec"
	"strings"
	"testing"
)

const testScriptURL = "https://astral.sh/uv/install.sh"

func TestScriptInstallerSkipsInstalledTool(t *testing.T) {
	runner := &fakeRunner{}
	inst := &ScriptInstaller{
		Runner:   runner,
		Tool:     "uv",
		URL:      testScriptURL,
		LookPath: func(string) (string, error) { return "/home/user/.local/bin/uv", nil },
	}

	if err := inst.Install(context.Background()); err != nil {
		t.Fatal(err)
	}

	if len(runner.calls) != 0 {
		t.Fatalf("calls = %#v, want no calls when tool is installed", runner.calls)
	}
}

func TestScriptInstallerRunsDownloadedScript(t *testing.T) {
	runner := &fakeRunner{}
	inst := &ScriptInstaller{
		Runner:   runner,
		Tool:     "uv",
		URL:      testScriptURL,
		LookPath: func(string) (string, error) { return "", exec.ErrNotFound },
	}

	if err := inst.Install(context.Background()); err != nil {
		t.Fatal(err)
	}

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

	script := got.args[1]
	for _, want := range []string{
		"curl -LsSf",
		testScriptURL,
		"| sh",
	} {
		if !strings.Contains(script, want) {
			t.Errorf("script does not contain %q:\n%s", want, script)
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
