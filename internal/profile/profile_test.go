package profile

import (
	"errors"
	"strings"
	"testing"
)

func TestParseEntries(t *testing.T) {
	body := []byte("# editor config\n.vimrc -> .vimrc\n\n  config/nvim -> .config/nvim  # nested dir\n")
	p, err := Parse("default", body)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got, want := len(p.Entries), 2; got != want {
		t.Fatalf("entries = %d, want %d", got, want)
	}
	if got, want := p.Entries[1].Source, "config/nvim"; got != want {
		t.Errorf("source = %q, want %q", got, want)
	}
	if got, want := p.Entries[1].Target, ".config/nvim"; got != want {
		t.Errorf("target = %q, want %q", got, want)
	}
}

func TestParseTable(t *testing.T) {
	cases := []struct {
		name string
		body string
		want string
	}{
		{"missing arrow", "a b", `expected "source -> target"`},
		{"empty source", " -> .x", "empty source"},
		{"empty target", ".x -> ", "empty target"},
		{"absolute source", "/etc/hosts -> hosts", "must be relative"},
		{"absolute target", "a -> /home/b", "must be relative"},
		{"parent escape", "../a -> a", "must not leave its root"},
		{"dot", ". -> .", `must not be "."`},
		{"double slash", "a//b -> c", "empty segment"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse("p", []byte(tc.body))
			if err == nil {
				t.Fatalf("Parse(%q) succeeded, want error", tc.body)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want it to mention %q", err, tc.want)
			}
		})
	}
}

func TestParseLineNumbers(t *testing.T) {
	_, err := Parse("default", []byte("# comment\nok -> ok\nbroken\n"))
	var le *LineError
	if !errors.As(err, &le) {
		t.Fatalf("error = %v, want a *LineError", err)
	}
	if le.Line != 3 {
		t.Errorf("line = %d, want 3", le.Line)
	}
	if le.File != "default" {
		t.Errorf("file = %q, want %q", le.File, "default")
	}
}

func TestParseEmptyProfile(t *testing.T) {
	if _, err := Parse("default", []byte("\n# only a comment\n")); err == nil {
		t.Fatal("Parse accepted a profile with no entries")
	}
}

func TestDuplicateTargets(t *testing.T) {
	p, err := Parse("p", []byte("a -> same\nb -> other\nc -> same\n"))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	dupes := p.DuplicateTargets()
	if len(dupes) != 1 || dupes[0] != "same" {
		t.Fatalf("duplicates = %v, want [same]", dupes)
	}
}

func TestValidName(t *testing.T) {
	for _, name := range []string{"default", "work-2021", "a.b"} {
		if err := ValidName(name); err != nil {
			t.Errorf("ValidName(%q) = %v, want nil", name, err)
		}
	}
	for _, name := range []string{"", ".", "..", "a/b", "two words"} {
		if err := ValidName(name); err == nil {
			t.Errorf("ValidName(%q) = nil, want an error", name)
		}
	}
}
