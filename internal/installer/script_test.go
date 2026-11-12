package installer

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
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
		LookPath: fakeLookPath("curl", "wget", "sh"),
	}

	if err := inst.Install(context.Background()); err != nil {
		t.Fatal(err)
	}

	want := "curl -LsSf " + testScriptURL + " | sh"
	assertScriptCommand(t, runner, "sh", want)
}

func TestScriptInstallerFallsBackToWget(t *testing.T) {
	runner := &fakeRunner{}
	inst := &ScriptInstaller{
		Runner:   runner,
		Tool:     "uv",
		URL:      testScriptURL,
		LookPath: fakeLookPath("wget", "sh"),
	}

	if err := inst.Install(context.Background()); err != nil {
		t.Fatal(err)
	}

	want := "wget -qO- " + testScriptURL + " | sh"
	assertScriptCommand(t, runner, "sh", want)
}

func TestScriptInstallerUsesConfiguredBashShell(t *testing.T) {
	runner := &fakeRunner{}
	inst := &ScriptInstaller{
		Runner:         runner,
		Tool:           "nvm",
		URL:            "https://raw.githubusercontent.com/nvm-sh/nvm/v0.41.0/install.sh",
		Shell:          "bash",
		InstalledPaths: []string{filepath.Join(t.TempDir(), "nvm.sh")},
		LookPath:       fakeLookPath("curl", "bash"),
	}

	if err := inst.Install(context.Background()); err != nil {
		t.Fatal(err)
	}

	want := "curl -LsSf https://raw.githubusercontent.com/nvm-sh/nvm/v0.41.0/install.sh | bash"
	assertScriptCommand(t, runner, "bash", want)
}

func TestScriptInstallerResolvesLatestVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/releases/latest" {
			http.Redirect(w, r, "/releases/tag/v9.9.9", http.StatusFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	runner := &fakeRunner{}
	inst := &ScriptInstaller{
		Runner:         runner,
		Tool:           "nvm",
		URL:            "https://example.com/nvm/{version}/install.sh",
		LatestURL:      server.URL + "/releases/latest",
		Shell:          "bash",
		InstalledPaths: []string{filepath.Join(t.TempDir(), "nvm.sh")},
		LookPath:       fakeLookPath("curl", "bash"),
	}

	if err := inst.Install(context.Background()); err != nil {
		t.Fatal(err)
	}

	want := "curl -LsSf https://example.com/nvm/v9.9.9/install.sh | bash"
	assertScriptCommand(t, runner, "bash", want)
}

func TestScriptInstallerFailsWhenLatestVersionUnknown(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	runner := &fakeRunner{}
	inst := &ScriptInstaller{
		Runner:         runner,
		Tool:           "nvm",
		URL:            "https://example.com/nvm/{version}/install.sh",
		LatestURL:      server.URL + "/releases/latest",
		Shell:          "bash",
		InstalledPaths: []string{filepath.Join(t.TempDir(), "nvm.sh")},
		LookPath:       fakeLookPath("curl", "bash"),
	}

	if err := inst.Install(context.Background()); err == nil {
		t.Fatal("expected an error when the latest version cannot be resolved")
	}
	if len(runner.calls) != 0 {
		t.Fatalf("calls = %#v, want no calls", runner.calls)
	}
}

func TestResolveLatestVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/releases/latest" {
			http.Redirect(w, r, "/releases/tag/v9.9.9", http.StatusFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	version, err := resolveLatestVersion(context.Background(), server.URL+"/releases/latest")
	if err != nil {
		t.Fatal(err)
	}
	if version != "v9.9.9" {
		t.Errorf("version = %q, want v9.9.9", version)
	}
}

func TestScriptInstallerSkipsWhenInstalledPathExists(t *testing.T) {
	nvmSh := filepath.Join(t.TempDir(), "nvm.sh")
	if err := os.WriteFile(nvmSh, []byte("# nvm\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	runner := &fakeRunner{}
	inst := &ScriptInstaller{
		Runner:         runner,
		Tool:           "nvm",
		URL:            "https://raw.githubusercontent.com/nvm-sh/nvm/v0.41.0/install.sh",
		Shell:          "bash",
		InstalledPaths: []string{nvmSh},
		LookPath:       fakeLookPath(),
	}

	if err := inst.Install(context.Background()); err != nil {
		t.Fatal(err)
	}

	if len(runner.calls) != 0 {
		t.Fatalf("calls = %#v, want no calls when installed path exists", runner.calls)
	}
}

func TestScriptInstallerExpandsEnvInInstalledPaths(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if err := os.MkdirAll(filepath.Join(home, ".nvm"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".nvm", "nvm.sh"), []byte("# nvm\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	runner := &fakeRunner{}
	inst := &ScriptInstaller{
		Runner:         runner,
		Tool:           "nvm",
		URL:            "https://raw.githubusercontent.com/nvm-sh/nvm/v0.41.0/install.sh",
		Shell:          "bash",
		InstalledPaths: []string{"$HOME/.nvm/nvm.sh"},
		LookPath:       fakeLookPath(),
	}

	if err := inst.Install(context.Background()); err != nil {
		t.Fatal(err)
	}

	if len(runner.calls) != 0 {
		t.Fatalf("calls = %#v, want no calls when installed path exists", runner.calls)
	}
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

func TestScriptInstallerWarnsWhenShellMissing(t *testing.T) {
	var logs bytes.Buffer
	original := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(original) })

	runner := &fakeRunner{}
	inst := &ScriptInstaller{
		Runner:         runner,
		Tool:           "nvm",
		URL:            "https://raw.githubusercontent.com/nvm-sh/nvm/v0.41.0/install.sh",
		Shell:          "bash",
		InstalledPaths: []string{filepath.Join(t.TempDir(), "nvm.sh")},
		LookPath:       fakeLookPath("curl", "wget"), // 没有 bash
	}

	err := inst.Install(context.Background())
	if err == nil {
		t.Fatal("expected an error when the configured shell is missing")
	}
	if len(runner.calls) != 0 {
		t.Fatalf("calls = %#v, want no calls", runner.calls)
	}

	warning := logs.String()
	for _, want := range []string{"WARN", "bash"} {
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

func assertScriptCommand(t *testing.T, runner *fakeRunner, shell, want string) {
	t.Helper()

	if len(runner.calls) != 1 {
		t.Fatalf("calls = %#v, want exactly one call", runner.calls)
	}

	got := runner.calls[0]
	if got.name != shell {
		t.Errorf("command = %q, want %q", got.name, shell)
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
