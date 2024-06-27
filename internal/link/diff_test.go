package link

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiffCleanAndDrifted(t *testing.T) {
	layout, source := fixture(t)
	entries := [][2]string{{".vimrc", ".vimrc"}}

	rep := layout.Diff(entries)
	if rep.Clean() {
		t.Fatal("diff of an empty home is clean, want the absent entry reported")
	}
	if got, want := rep.Checked, 1; got != want {
		t.Fatalf("checked = %d, want %d", got, want)
	}
	if got := rep.Findings[0].State; got != Absent {
		t.Fatalf("state = %v, want Absent", got)
	}

	if _, err := layout.Install(source, layout.Target(".vimrc"), false); err != nil {
		t.Fatal(err)
	}
	if rep := layout.Diff(entries); !rep.Clean() {
		t.Fatalf("diff of an installed tree = %+v, want clean", rep)
	}

	if err := os.Remove(layout.Target(".vimrc")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(layout.Target(".vimrc"), []byte("edited\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	rep = layout.Diff(entries)
	if len(rep.Findings) != 1 {
		t.Fatalf("findings = %d, want 1", len(rep.Findings))
	}
	if got := rep.Findings[0].State; got != Foreign {
		t.Fatalf("state = %v, want Foreign", got)
	}
	if got := rep.Findings[0].Note; got == "" {
		t.Error("foreign finding has no note")
	}
}

func TestDiffNotes(t *testing.T) {
	layout, source := fixture(t)
	target := layout.Target(".vimrc")

	if err := os.Symlink(filepath.Join(layout.Repo, "gone"), target); err != nil {
		t.Fatal(err)
	}
	note := layout.note(source, target, Elsewhere)
	if !strings.Contains(note, "points at") {
		t.Fatalf("elsewhere note = %q", note)
	}

	if err := os.Remove(target); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("set number\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if note := layout.note(source, target, Foreign); note != "identical contents" {
		t.Fatalf("identical-file note = %q", note)
	}

	if err := os.WriteFile(target, []byte("set nope\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if note := layout.note(source, target, Foreign); note != "regular file" {
		t.Fatalf("different-file note = %q", note)
	}
	if note := layout.note(source, target, Dangling); note != "source deleted" {
		t.Fatalf("dangling note = %q", note)
	}
}

func TestDiffWrite(t *testing.T) {
	layout, source := fixture(t)
	target := layout.Target("b/.vimrc")
	if _, err := layout.Install(source, layout.Target("a/.vimrc"), false); err != nil {
		t.Fatal(err)
	}
	rep := layout.Diff([][2]string{{".vimrc", "b/.vimrc"}, {".vimrc", "a/.vimrc"}})
	var buf bytes.Buffer
	n, err := rep.Write(&buf)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if n != 1 {
		t.Fatalf("written = %d, want 1", n)
	}
	if !strings.Contains(buf.String(), target) {
		t.Fatalf("output = %q, want it to mention %q", buf.String(), target)
	}
	if strings.Contains(buf.String(), "a/.vimrc") {
		t.Fatalf("output mentions the healthy entry: %q", buf.String())
	}
}
