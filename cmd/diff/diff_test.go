package diff

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setup(t *testing.T) (repoDir, homeDir string) {
	t.Helper()
	repoDir, homeDir = t.TempDir(), t.TempDir()
	if err := os.MkdirAll(filepath.Join(repoDir, "profiles"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, ".vimrc"), []byte("set number\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, "profiles", "default.conf"), []byte(".vimrc -> .vimrc\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return repoDir, homeDir
}

func TestRunClean(t *testing.T) {
	repoDir, homeDir := setup(t)
	if err := os.Symlink(filepath.Join(repoDir, ".vimrc"), filepath.Join(homeDir, ".vimrc")); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	n, err := Run(homeDir, repoDir, Options{Profile: "default"}, &out)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if n != 0 {
		t.Fatalf("findings = %d, want 0", n)
	}
	if out.String() != "" {
		t.Fatalf("output = %q, want empty", out.String())
	}
}

func TestRunDrift(t *testing.T) {
	repoDir, homeDir := setup(t)
	if err := os.WriteFile(filepath.Join(homeDir, ".vimrc"), []byte("edited\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	n, err := Run(homeDir, repoDir, Options{Profile: "default"}, &out)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if n != 1 {
		t.Fatalf("findings = %d, want 1", n)
	}
	got := out.String()
	if !strings.Contains(got, "foreign") {
		t.Fatalf("output = %q, want a foreign state", got)
	}
	if !strings.Contains(got, "1 of 1 entries drifted") {
		t.Fatalf("output = %q, want a summary line", got)
	}
}

func TestRunAbsent(t *testing.T) {
	repoDir, homeDir := setup(t)
	var out bytes.Buffer
	n, err := Run(homeDir, repoDir, Options{Profile: "default"}, &out)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if n != 1 {
		t.Fatalf("findings = %d, want 1", n)
	}
	if !strings.Contains(out.String(), "absent") {
		t.Fatalf("output = %q, want an absent state", out.String())
	}
}
