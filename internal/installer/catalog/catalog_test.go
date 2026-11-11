package catalog

import (
	"strings"
	"testing"
)

func TestLookupGit(t *testing.T) {
	tool, ok := Lookup("git")
	if !ok {
		t.Fatal("git should be registered")
	}

	for _, manager := range []string{"apt", "dnf"} {
		if tool.Packages[manager] != "git" {
			t.Errorf("git package for %s = %q, want git", manager, tool.Packages[manager])
		}
	}
}

func TestLookupUV(t *testing.T) {
	tool, ok := Lookup("uv")
	if !ok {
		t.Fatal("uv should be registered")
	}

	if tool.InstallScript != "https://astral.sh/uv/install.sh" {
		t.Errorf("uv install script = %q, want official installer", tool.InstallScript)
	}
}

func TestLookupOssutil(t *testing.T) {
	tool, ok := Lookup("ossutil")
	if !ok {
		t.Fatal("ossutil should be registered")
	}

	archive := tool.InstallArchive
	if archive == nil {
		t.Fatal("ossutil should use the archive installation")
	}

	if archive.Version != "2.4.1" {
		t.Errorf("ossutil version = %q, want 2.4.1", archive.Version)
	}
	if archive.Binary != "ossutil" {
		t.Errorf("ossutil binary = %q, want ossutil", archive.Binary)
	}
	if archive.InstallDir != "/usr/local/bin" {
		t.Errorf("ossutil install dir = %q, want /usr/local/bin", archive.InstallDir)
	}
	if archive.Symlink != "/usr/bin/ossutil" {
		t.Errorf("ossutil symlink = %q, want /usr/bin/ossutil", archive.Symlink)
	}

	// 文档中提供的 6 个平台安装包。
	platforms := []string{
		"linux/386",
		"linux/amd64",
		"linux/arm",
		"linux/arm64",
		"darwin/amd64",
		"darwin/arm64",
	}
	for _, platform := range platforms {
		artifact, ok := archive.Artifacts[platform]
		if !ok {
			t.Errorf("missing ossutil artifact for %s", platform)
			continue
		}
		if !strings.HasPrefix(artifact.URL, "https://gosspublic.alicdn.com/ossutil/v2/2.4.1/") {
			t.Errorf("artifact URL for %s = %q, want official download URL", platform, artifact.URL)
		}
		if len(artifact.SHA256) != 64 {
			t.Errorf("artifact SHA256 for %s = %q, want 64 hex characters", platform, artifact.SHA256)
		}
	}
}
