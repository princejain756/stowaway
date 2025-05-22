package list

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, "profiles"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"work", "default"} {
		body := []byte(".vimrc -> .vimrc\n")
		if err := os.WriteFile(filepath.Join(repo, "profiles", name+".conf"), body, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var out bytes.Buffer
	n, err := Run(repo, &out)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if n != 2 {
		t.Fatalf("count = %d, want 2", n)
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 2 || lines[0] != "default" || lines[1] != "work" {
		t.Fatalf("output = %q, want default then work", out.String())
	}
}

func TestRunEmpty(t *testing.T) {
	var out bytes.Buffer
	n, err := Run(t.TempDir(), &out)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if n != 0 || out.String() != "" {
		t.Fatalf("count = %d output = %q, want nothing", n, out.String())
	}
}
