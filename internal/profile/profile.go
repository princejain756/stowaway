package profile

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// ErrNotFound is returned by Load when the requested profile does not exist in
// the repository.
var ErrNotFound = errors.New("profile not found")

// Entry maps one file in the repository onto one path in the home directory.
// Source is relative to the repository root, Target is relative to $HOME. Both
// use forward slashes so that profile files are portable between machines.
type Entry struct {
	Source string
	Target string
}

// Profile is a named group of entries, loaded from profiles/<name>.conf.
type Profile struct {
	Name    string
	Entries []Entry
}

// LineError reports a malformed line inside a profile file.
type LineError struct {
	File string
	Line int
	Text string
	Err  error
}

func (e *LineError) Error() string {
	return fmt.Sprintf("%s:%d: %v: %q", e.File, e.Line, e.Err, e.Text)
}

func (e *LineError) Unwrap() error { return e.Err }

// Parse reads a profile body. An entry is written as "source -> target";
// anything after a '#' is a comment and blank lines are ignored.
func Parse(name string, body []byte) (*Profile, error) {
	p := &Profile{Name: name}
	for i, raw := range strings.Split(string(body), "\n") {
		line := strings.TrimSpace(stripComment(raw))
		if line == "" {
			continue
		}
		entry, err := parseEntry(line)
		if err != nil {
			return nil, &LineError{File: name, Line: i + 1, Text: line, Err: err}
		}
		p.Entries = append(p.Entries, entry)
	}
	if len(p.Entries) == 0 {
		return nil, fmt.Errorf("%s: profile is empty", name)
	}
	return p, nil
}

func stripComment(line string) string {
	if i := strings.IndexByte(line, '#'); i >= 0 {
		return line[:i]
	}
	return line
}

func parseEntry(line string) (Entry, error) {
	i := strings.Index(line, "->")
	if i < 0 {
		return Entry{}, errors.New("expected \"source -> target\"")
	}
	source := strings.TrimSpace(line[:i])
	target := strings.TrimSpace(line[i+2:])
	if err := checkRel("source", source); err != nil {
		return Entry{}, err
	}
	if err := checkRel("target", target); err != nil {
		return Entry{}, err
	}
	return Entry{Source: source, Target: target}, nil
}

func checkRel(kind, path string) error {
	switch {
	case path == "":
		return fmt.Errorf("empty %s", kind)
	case strings.HasPrefix(path, "/"):
		return fmt.Errorf("%s %q must be relative", kind, path)
	case path == ".":
		return fmt.Errorf("%s %q must not be \".\"", kind, path)
	}
	for _, part := range strings.Split(path, "/") {
		if part == ".." {
			return fmt.Errorf("%s %q must not leave its root", kind, path)
		}
	}
	if strings.Contains(path, "//") {
		return fmt.Errorf("%s %q has an empty segment", kind, path)
	}
	return nil
}

// DuplicateTargets returns the targets claimed by more than one entry, sorted.
func (p *Profile) DuplicateTargets() []string {
	seen := make(map[string]int, len(p.Entries))
	var dupes []string
	for _, e := range p.Entries {
		seen[e.Target]++
		if seen[e.Target] == 2 {
			dupes = append(dupes, e.Target)
		}
	}
	sort.Strings(dupes)
	return dupes
}
