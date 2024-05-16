package uninstall

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
	if err := os.Symlink(filepath.Join(repoDir, ".vimrc"), filepath.Join(homeDir, ".vimrc")); err != nil {
		t.Fatal(err)
	}
	return repoDir, homeDir
}

func TestRun(t *testing.T) {
	repoDir, homeDir := setup(t)
	var out, errOut bytes.Buffer
	if err := Run(homeDir, repoDir, Options{Profile: "default"}, &out, &errOut); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(homeDir, ".vimrc")); !os.IsNotExist(err) {
		t.Fatalf("link still present: %v", err)
	}
	if !strings.Contains(out.String(), "unlink") {
		t.Fatalf("output = %q, want an unlink line", out.String())
	}
}

func TestRunDryRun(t *testing.T) {
	repoDir, homeDir := setup(t)
	var out, errOut bytes.Buffer
	if err := Run(homeDir, repoDir, Options{Profile: "default", DryRun: true}, &out, &errOut); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(homeDir, ".vimrc")); err != nil {
		t.Fatalf("dry run removed the link: %v", err)
	}
	if !strings.Contains(out.String(), "would unlink") {
		t.Fatalf("output = %q, want a planned unlink", out.String())
	}
}

func TestRunPrunes(t *testing.T) {
	repoDir, homeDir := t.TempDir(), t.TempDir()
	if err := os.MkdirAll(filepath.Join(repoDir, "profiles"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(repoDir, "d"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, "d", "conf"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, "profiles", "default.conf"), []byte("d/conf -> .config/deep/conf\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(homeDir, ".config", "deep"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(repoDir, "d", "conf"), filepath.Join(homeDir, ".config", "deep", "conf")); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	if err := Run(homeDir, repoDir, Options{Profile: "default", Prune: true}, &out, &errOut); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if _, err := os.Stat(filepath.Join(homeDir, ".config")); !os.IsNotExist(err) {
		t.Fatalf("empty directories were not pruned: %v", err)
	}
}

func TestRunNoPrune(t *testing.T) {
	repoDir, homeDir := setup(t)
	if err := os.MkdirAll(filepath.Join(homeDir, "keep"), 0o755); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	if err := Run(homeDir, repoDir, Options{Profile: "default", Prune: false}, &out, &errOut); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if _, err := os.Stat(homeDir); err != nil {
		t.Fatalf("home was removed: %v", err)
	}
}
