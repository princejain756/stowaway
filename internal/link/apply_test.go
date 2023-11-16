package link

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstall(t *testing.T) {
	layout, source := fixture(t)
	target := layout.Target(".config/app/conf")

	res, err := layout.Install(source, target, false)
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	if res.Skipped {
		t.Fatal("Install skipped a fresh target")
	}
	if got := layout.Inspect(source, target); got != Linked {
		t.Fatalf("state after install = %v, want Linked", got)
	}

	again, err := layout.Install(source, target, false)
	if err != nil {
		t.Fatalf("second Install: %v", err)
	}
	if !again.Skipped || again.Action != "unchanged" {
		t.Fatalf("second Install = %+v, want an unchanged skip", again)
	}
}

func TestInstallRefusesForeign(t *testing.T) {
	layout, source := fixture(t)
	target := layout.Target(".vimrc")
	if err := os.WriteFile(target, []byte("mine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := layout.Install(source, target, false); err == nil {
		t.Fatal("Install overwrote a foreign file without force")
	}
	if _, err := layout.Install(source, target, true); err != nil {
		t.Fatalf("forced Install: %v", err)
	}
	if got := layout.Inspect(source, target); got != Linked {
		t.Fatalf("state after forced install = %v, want Linked", got)
	}
}

func TestInstallMissingSource(t *testing.T) {
	layout, source := fixture(t)
	if err := os.Remove(source); err != nil {
		t.Fatal(err)
	}
	if _, err := layout.Install(source, layout.Target(".vimrc"), false); err == nil {
		t.Fatal("Install accepted a missing source")
	}
}

func TestRemove(t *testing.T) {
	layout, source := fixture(t)
	target := layout.Target("nested/.vimrc")

	if _, err := layout.Remove(source, target, true); err != nil {
		t.Fatalf("Remove of an absent target: %v", err)
	}
	if _, err := layout.Install(source, target, false); err != nil {
		t.Fatal(err)
	}
	if _, err := layout.Remove(source, target, true); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if _, err := os.Lstat(target); !os.IsNotExist(err) {
		t.Fatalf("target still present after Remove: %v", err)
	}
	if _, err := os.Stat(filepath.Join(layout.Home, "nested")); !os.IsNotExist(err) {
		t.Fatalf("empty parent was not pruned: %v", err)
	}
}

func TestRemoveKeepsPopulatedParent(t *testing.T) {
	layout, source := fixture(t)
	target := layout.Target("nested/.vimrc")
	if _, err := layout.Install(source, target, false); err != nil {
		t.Fatal(err)
	}
	keep := filepath.Join(layout.Home, "nested", "keep")
	if err := os.WriteFile(keep, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := layout.Remove(source, target, true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(layout.Home, "nested")); err != nil {
		t.Fatalf("populated parent was pruned: %v", err)
	}
}

func TestRemoveForeign(t *testing.T) {
	layout, source := fixture(t)
	target := layout.Target(".vimrc")
	if err := os.WriteFile(target, []byte("mine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := layout.Remove(source, target, false); err == nil {
		t.Fatal("Remove deleted a file it does not manage")
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("foreign file was deleted: %v", err)
	}
}
