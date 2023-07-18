// Package link creates and removes the symlinks that make up an installation,
// and inspects them to report drift.
package link

import (
	"os"
	"path/filepath"
)

// State describes how a target path stands relative to the entry that should
// own it.
type State int

const (
	// Linked means the target is a symlink pointing at the expected source.
	Linked State = iota
	// Absent means the target does not exist.
	Absent
	// Foreign means the target exists but is not a symlink and is not ours.
	Foreign
	// Elsewhere means the target is a symlink pointing somewhere else.
	Elsewhere
	// Dangling means the target is our symlink but its source is gone.
	Dangling
	// SourceMissing means the repository file is missing while the target is
	// not a link at all.
	SourceMissing
)

func (s State) String() string {
	switch s {
	case Linked:
		return "linked"
	case Absent:
		return "absent"
	case Foreign:
		return "foreign"
	case Elsewhere:
		return "elsewhere"
	case Dangling:
		return "dangling"
	case SourceMissing:
		return "source-missing"
	}
	return "unknown"
}

// Layout resolves relative repository and home paths to absolute ones.
type Layout struct {
	Repo string
	Home string
}

// Source is the absolute path of a repository file.
func (l Layout) Source(rel string) string { return filepath.Join(l.Repo, rel) }

// Target is the absolute path of a home path.
func (l Layout) Target(rel string) string { return filepath.Join(l.Home, rel) }

// Inspect reports the state of target with respect to the file it should link
// to.
func (l Layout) Inspect(source, target string) State {
	info, err := os.Lstat(target)
	if os.IsNotExist(err) {
		if _, err := os.Stat(source); os.IsNotExist(err) {
			return SourceMissing
		}
		return Absent
	}
	if err != nil {
		return Foreign
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return Foreign
	}
	dest, err := os.Readlink(target)
	if err != nil {
		return Foreign
	}
	if !filepath.IsAbs(dest) {
		dest = filepath.Join(filepath.Dir(target), dest)
	}
	if !sameFile(dest, source) {
		return Elsewhere
	}
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return Dangling
	}
	return Linked
}

func sameFile(a, b string) bool {
	if filepath.Clean(a) == filepath.Clean(b) {
		return true
	}
	ai, err := os.Stat(a)
	if err != nil {
		return false
	}
	bi, err := os.Stat(b)
	if err != nil {
		return false
	}
	return os.SameFile(ai, bi)
}
