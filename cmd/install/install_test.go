package install

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func repo(t *testing.T) (repoDir, homeDir string) {
	t.Helper()
	repoDir, homeDir = t.TempDir(), t.TempDir()
	if err := os.MkdirAll(filepath.Join(repoDir, "profiles"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, ".vimrc"), []byte("set number\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	body := ".vimrc -> .vimrc\nconfig/app/conf -> .config/app/conf\n"
	if err := os.WriteFile(filepath.Join(repoDir, "profiles", "default.conf"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(repoDir, "config", "app"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, "config", "app", "conf"), []byte("x=1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return repoDir, homeDir
}

func TestRun(t *testing.T) {
	repoDir, homeDir := repo(t)
	var out, errOut bytes.Buffer

	res, err := Run(homeDir, repoDir, Options{Profile: "default"}, &out, &errOut)
	if err != nil {
		t.Fatalf("Run: %v (stderr %q)", err, errOut.String())
	}
	if res.Linked != 2 {
		t.Fatalf("linked = %d, want 2", res.Linked)
	}
	for _, target := range []string{".vimrc", ".config/app/conf"} {
		info, err := os.Lstat(filepath.Join(homeDir, target))
		if err != nil {
			t.Fatalf("lstat %s: %v", target, err)
		}
		if info.Mode()&os.ModeSymlink == 0 {
			t.Fatalf("%s is not a symlink", target)
		}
	}
	if !strings.Contains(out.String(), "link") {
		t.Fatalf("output = %q, want a link line", out.String())
	}
}

func TestRunDryRun(t *testing.T) {
	repoDir, homeDir := repo(t)
	var out, errOut bytes.Buffer

	res, err := Run(homeDir, repoDir, Options{Profile: "default", DryRun: true}, &out, &errOut)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Linked != 0 {
		t.Fatalf("linked = %d, want 0 for a dry run", res.Linked)
	}
	if _, err := os.Lstat(filepath.Join(homeDir, ".vimrc")); !os.IsNotExist(err) {
		t.Fatalf("dry run created a link: %v", err)
	}
	if !strings.Contains(out.String(), "link ") {
		t.Fatalf("output = %q, want the planned link", out.String())
	}
}

func TestRunMissingProfile(t *testing.T) {
	repoDir, homeDir := repo(t)
	var out, errOut bytes.Buffer
	if _, err := Run(homeDir, repoDir, Options{Profile: "nope"}, &out, &errOut); err == nil {
		t.Fatal("Run accepted a missing profile")
	}
}

func TestRunConflict(t *testing.T) {
	repoDir, homeDir := repo(t)
	if err := os.WriteFile(filepath.Join(homeDir, ".vimrc"), []byte("mine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	if _, err := Run(homeDir, repoDir, Options{Profile: "default"}, &out, &errOut); err == nil {
		t.Fatal("Run overwrote a conflicted file")
	}
	if _, err := Run(homeDir, repoDir, Options{Profile: "default", Force: true}, &out, &errOut); err != nil {
		t.Fatalf("forced Run: %v", err)
	}
}
