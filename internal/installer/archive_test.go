package installer

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"os"
	"testing"
)

// archiveRunner 记录调用，并在下载命令执行时写出测试用的压缩包。
type archiveRunner struct {
	fakeRunner
	zipData []byte
}

func (r *archiveRunner) Run(ctx context.Context, name string, args ...string) error {
	if err := r.fakeRunner.Run(ctx, name, args...); err != nil {
		return err
	}

	if name != "curl" && name != "wget" {
		return nil
	}

	path := ""
	if name == "curl" {
		for i := range args {
			if args[i] == "-o" && i+1 < len(args) {
				path = args[i+1]
				break
			}
		}
	} else if len(args) >= 2 {
		path = args[1]
	}
	if path == "" {
		return nil
	}

	return os.WriteFile(path, r.zipData, 0o644)
}

func buildZip(t *testing.T, entryName string) []byte {
	t.Helper()

	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	entry, err := w.Create(entryName)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte("#!/bin/sh\necho fake ossutil\n")); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	return buf.Bytes()
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func newArchiveInstaller(runner *archiveRunner, lookPath func(string) (string, error), data []byte) *ArchiveInstaller {
	return &ArchiveInstaller{
		Runner:     runner,
		Tool:       "ossutil",
		Version:    "2.4.1",
		URL:        "https://example.com/ossutil-2.4.1-linux-amd64.zip",
		SHA256:     sha256Hex(data),
		Binary:     "ossutil",
		InstallDir: "/usr/local/bin",
		Symlink:    "/usr/bin/ossutil",
		LookPath:   lookPath,
	}
}

func TestArchiveInstallerInstallsBinary(t *testing.T) {
	data := buildZip(t, "ossutil-2.4.1-linux-amd64/ossutil")
	runner := &archiveRunner{zipData: data}
	inst := newArchiveInstaller(runner, fakeLookPath("curl"), data)

	if err := inst.Install(context.Background()); err != nil {
		t.Fatal(err)
	}

	if len(runner.calls) != 3 {
		t.Fatalf("calls = %#v, want download + mv + ln", runner.calls)
	}

	download := runner.calls[0]
	if download.name != "curl" || download.elevated {
		t.Errorf("download call = %#v, want non-elevated curl", download)
	}
	if len(download.args) != 4 || download.args[0] != "-fsSL" || download.args[2] != "-o" {
		t.Errorf("curl args = %#v, want [-fsSL URL -o path]", download.args)
	}

	mv := runner.calls[1]
	if mv.name != "mv" || !mv.elevated || len(mv.args) != 2 {
		t.Fatalf("mv call = %#v, want elevated mv src dst", mv)
	}
	if mv.args[1] != "/usr/local/bin/ossutil" {
		t.Errorf("mv target = %q, want /usr/local/bin/ossutil", mv.args[1])
	}

	ln := runner.calls[2]
	if ln.name != "ln" || !ln.elevated {
		t.Fatalf("ln call = %#v, want elevated ln", ln)
	}
	wantLn := []string{"-sf", "/usr/local/bin/ossutil", "/usr/bin/ossutil"}
	if len(ln.args) != len(wantLn) {
		t.Fatalf("ln args = %#v, want %#v", ln.args, wantLn)
	}
	for i := range wantLn {
		if ln.args[i] != wantLn[i] {
			t.Errorf("ln args = %#v, want %#v", ln.args, wantLn)
			break
		}
	}
}

func TestArchiveInstallerUsesWget(t *testing.T) {
	data := buildZip(t, "ossutil-2.4.1-linux-amd64/ossutil")
	runner := &archiveRunner{zipData: data}
	inst := newArchiveInstaller(runner, fakeLookPath("wget"), data)

	if err := inst.Install(context.Background()); err != nil {
		t.Fatal(err)
	}

	download := runner.calls[0]
	if download.name != "wget" || len(download.args) != 3 || download.args[0] != "-qO" {
		t.Fatalf("download call = %#v, want wget -qO path URL", download)
	}
	if download.args[2] != inst.URL {
		t.Errorf("wget URL = %q, want %q", download.args[2], inst.URL)
	}
}

func TestArchiveInstallerSkipsInstalledTool(t *testing.T) {
	runner := &archiveRunner{}
	inst := newArchiveInstaller(runner, fakeLookPath("ossutil", "curl"), nil)

	if err := inst.Install(context.Background()); err != nil {
		t.Fatal(err)
	}

	if len(runner.calls) != 0 {
		t.Fatalf("calls = %#v, want no calls when tool is installed", runner.calls)
	}
}

func TestArchiveInstallerRejectsChecksumMismatch(t *testing.T) {
	data := buildZip(t, "ossutil-2.4.1-linux-amd64/ossutil")
	runner := &archiveRunner{zipData: data}
	inst := newArchiveInstaller(runner, fakeLookPath("curl"), data)
	inst.SHA256 = sha256Hex([]byte("tampered"))

	err := inst.Install(context.Background())
	if err == nil {
		t.Fatal("expected an error for checksum mismatch")
	}

	if len(runner.calls) != 1 || runner.calls[0].elevated {
		t.Fatalf("calls = %#v, want only the download call", runner.calls)
	}
}

func TestArchiveInstallerWarnsWhenNoDownloader(t *testing.T) {
	var logs bytes.Buffer
	original := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(original) })

	runner := &archiveRunner{}
	inst := newArchiveInstaller(runner, fakeLookPath(), nil)

	err := inst.Install(context.Background())
	if err == nil {
		t.Fatal("expected an error when curl and wget are both missing")
	}
	if len(runner.calls) != 0 {
		t.Fatalf("calls = %#v, want no calls", runner.calls)
	}

	warning := logs.String()
	for _, want := range []string{"WARN", "curl", "wget"} {
		if !bytes.Contains([]byte(warning), []byte(want)) {
			t.Errorf("warning does not contain %q:\n%s", want, warning)
		}
	}
}

func TestArchiveInstallerMissingBinaryInZip(t *testing.T) {
	data := buildZip(t, "ossutil-2.4.1-linux-amd64/README.md")
	runner := &archiveRunner{zipData: data}
	inst := newArchiveInstaller(runner, fakeLookPath("curl"), data)

	if err := inst.Install(context.Background()); err == nil {
		t.Fatal("expected an error when the binary is missing from the archive")
	}
}
