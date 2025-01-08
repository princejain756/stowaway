package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func sandbox(t *testing.T) (repo, home string) {
	t.Helper()
	repo, home = t.TempDir(), t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, "profiles"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, ".vimrc"), []byte("set number\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "profiles", "default.conf"), []byte(".vimrc -> .vimrc\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	t.Setenv("STOWAWAY_REPO", repo)
	return repo, home
}

func TestRunInstallThenDiffThenUninstall(t *testing.T) {
	_, home := sandbox(t)
	var out, errOut bytes.Buffer

	if err := run([]string{"install"}, &out, &errOut); err != nil {
		t.Fatalf("install: %v (stderr %q)", err, errOut.String())
	}
	if _, err := os.Lstat(filepath.Join(home, ".vimrc")); err != nil {
		t.Fatalf("install created no link: %v", err)
	}

	out.Reset()
	if err := run([]string{"diff"}, &out, &errOut); err != nil {
		t.Fatalf("diff on a clean tree: %v", err)
	}

	out.Reset()
	if err := run([]string{"uninstall"}, &out, &errOut); err != nil {
		t.Fatalf("uninstall: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(home, ".vimrc")); !os.IsNotExist(err) {
		t.Fatalf("uninstall left the link: %v", err)
	}
}

func TestRunDiffExitCode(t *testing.T) {
	_, home := sandbox(t)
	var out, errOut bytes.Buffer
	if err := os.WriteFile(filepath.Join(home, ".vimrc"), []byte("mine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := run([]string{"diff"}, &out, &errOut)
	var exit exitError
	if !asExit(err, &exit) {
		t.Fatalf("diff error = %v, want an exitError", err)
	}
	if exit.code != 1 {
		t.Fatalf("exit code = %d, want 1", exit.code)
	}
}

func TestRunHelpAndVersion(t *testing.T) {
	sandbox(t)
	var out, errOut bytes.Buffer
	if err := run([]string{"help"}, &out, &errOut); err != nil {
		t.Fatalf("help: %v", err)
	}
	if !strings.Contains(out.String(), "usage:") {
		t.Fatalf("help output = %q", out.String())
	}
	out.Reset()
	if err := run([]string{"version"}, &out, &errOut); err != nil {
		t.Fatalf("version: %v", err)
	}
	if !strings.Contains(out.String(), version) {
		t.Fatalf("version output = %q", out.String())
	}
}

func TestRunUnknownCommand(t *testing.T) {
	sandbox(t)
	var out, errOut bytes.Buffer
	err := run([]string{"frobnicate"}, &out, &errOut)
	var exit exitError
	if !asExit(err, &exit) || exit.code != 2 {
		t.Fatalf("error = %v, want exit 2", err)
	}
}

func TestRunNoArgs(t *testing.T) {
	var out, errOut bytes.Buffer
	err := run(nil, &out, &errOut)
	var exit exitError
	if !asExit(err, &exit) || exit.code != 2 {
		t.Fatalf("error = %v, want exit 2", err)
	}
	if !strings.Contains(errOut.String(), "usage:") {
		t.Fatalf("stderr = %q, want usage", errOut.String())
	}
}

func TestRunList(t *testing.T) {
	sandbox(t)
	var out, errOut bytes.Buffer
	if err := run([]string{"list"}, &out, &errOut); err != nil {
		t.Fatalf("list: %v", err)
	}
	if strings.TrimSpace(out.String()) != "default" {
		t.Fatalf("list output = %q, want default", out.String())
	}
}

func TestRepoForOverride(t *testing.T) {
	repo, _ := sandbox(t)
	got, err := repoFor(t.TempDir())
	if err != nil {
		t.Fatalf("repoFor: %v", err)
	}
	if got != repo {
		t.Fatalf("repoFor = %q, want the override %q", got, repo)
	}
}

// asExit reports whether err carries an exit code and copies it to target.
func asExit(err error, target *exitError) bool {
	return errors.As(err, target)
}
