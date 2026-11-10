package catalog

import "testing"

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
