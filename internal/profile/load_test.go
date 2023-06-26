package profile

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoad(t *testing.T) {
	repo := t.TempDir()
	write(t, filepath.Join(repo, Dir, "work"+Ext), "a -> a\n")

	p, err := Load(repo, "work")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if p.Name != "work" || len(p.Entries) != 1 {
		t.Fatalf("profile = %+v, want name work with one entry", p)
	}
}

func TestLoadMissing(t *testing.T) {
	if _, err := Load(t.TempDir(), "nope"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestLoadRejectsDuplicateTargets(t *testing.T) {
	repo := t.TempDir()
	write(t, filepath.Join(repo, Dir, "dup"+Ext), "a -> same\nb -> same\n")
	if _, err := Load(repo, "dup"); err == nil {
		t.Fatal("Load accepted a profile with duplicate targets")
	}
}

func TestList(t *testing.T) {
	repo := t.TempDir()
	write(t, filepath.Join(repo, Dir, "beta"+Ext), "a -> a\n")
	write(t, filepath.Join(repo, Dir, "alpha"+Ext), "a -> a\n")
	write(t, filepath.Join(repo, Dir, "notes.txt"), "ignored")

	names, err := List(repo)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	want := []string{"alpha", "beta"}
	if len(names) != len(want) {
		t.Fatalf("names = %v, want %v", names, want)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("names = %v, want %v", names, want)
		}
	}
}

func TestListMissingDir(t *testing.T) {
	names, err := List(t.TempDir())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(names) != 0 {
		t.Fatalf("names = %v, want none", names)
	}
}

func TestRoot(t *testing.T) {
	repo := t.TempDir()
	write(t, filepath.Join(repo, Dir, "default"+Ext), "a -> a\n")
	nested := filepath.Join(repo, "deep", "deeper")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := Root(nested)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	if got != repo {
		t.Fatalf("Root = %q, want %q", got, repo)
	}
}

func TestRootNotFound(t *testing.T) {
	dir := t.TempDir()
	if _, err := Root(dir); !errors.Is(err, ErrNoRepo) {
		t.Fatalf("error = %v, want ErrNoRepo", err)
	}
}
