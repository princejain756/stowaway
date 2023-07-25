package link

import (
	"os"
	"path/filepath"
	"testing"
)

// fixture builds a repository and a home directory with one source file and
// returns a layout over them.
func fixture(t *testing.T) (Layout, string) {
	t.Helper()
	repo, home := t.TempDir(), t.TempDir()
	source := filepath.Join(repo, ".vimrc")
	if err := os.WriteFile(source, []byte("set number\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return Layout{Repo: repo, Home: home}, source
}

func TestInspectStates(t *testing.T) {
	layout, source := fixture(t)
	target := layout.Target(".vimrc")

	if got := layout.Inspect(source, target); got != Absent {
		t.Fatalf("fresh target = %v, want Absent", got)
	}

	if err := os.Symlink(source, target); err != nil {
		t.Fatal(err)
	}
	if got := layout.Inspect(source, target); got != Linked {
		t.Fatalf("linked target = %v, want Linked", got)
	}

	other := filepath.Join(layout.Repo, "other")
	if err := os.WriteFile(other, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(target); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(other, target); err != nil {
		t.Fatal(err)
	}
	if got := layout.Inspect(source, target); got != Elsewhere {
		t.Fatalf("misaimed target = %v, want Elsewhere", got)
	}

	if err := os.Remove(target); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("mine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := layout.Inspect(source, target); got != Foreign {
		t.Fatalf("regular file target = %v, want Foreign", got)
	}

	if err := os.Remove(target); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(source); err != nil {
		t.Fatal(err)
	}
	if got := layout.Inspect(source, target); got != SourceMissing {
		t.Fatalf("missing source = %v, want SourceMissing", got)
	}
}

func TestInspectDangling(t *testing.T) {
	layout, source := fixture(t)
	target := layout.Target(".vimrc")
	if err := os.Symlink(source, target); err != nil {
		t.Fatal(err)
	}
	// Remove the file but keep the directory entry, so the link still resolves
	// to the recorded path while the source itself is gone.
	if err := os.Remove(source); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(source, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(source); err != nil {
		t.Fatal(err)
	}
	if got := layout.Inspect(source, target); got != Dangling {
		t.Fatalf("dangling target = %v, want Dangling", got)
	}
}

func TestInspectRelativeLink(t *testing.T) {
	layout, source := fixture(t)
	target := layout.Target("sub/.vimrc")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	rel, err := filepath.Rel(filepath.Dir(target), source)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(rel, target); err != nil {
		t.Fatal(err)
	}
	if got := layout.Inspect(source, target); got != Linked {
		t.Fatalf("relative link = %v, want Linked", got)
	}
}

func TestStateStrings(t *testing.T) {
	want := map[State]string{
		Linked: "linked", Absent: "absent", Foreign: "foreign",
		Elsewhere: "elsewhere", Dangling: "dangling", SourceMissing: "source-missing",
	}
	for state, text := range want {
		if got := state.String(); got != text {
			t.Errorf("State(%d) = %q, want %q", state, got, text)
		}
	}
}
