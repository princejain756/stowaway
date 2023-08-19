package link

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Result is the outcome of installing or removing one entry.
type Result struct {
	Source  string
	Target  string
	Action  string
	Skipped bool
}

func (r Result) String() string {
	if r.Skipped {
		return fmt.Sprintf("%s %s -> %s (skipped)", r.Action, r.Source, r.Target)
	}
	return fmt.Sprintf("%s %s -> %s", r.Action, r.Source, r.Target)
}

// Install creates the symlink for one entry. An entry that is already linked
// reports the action "unchanged" and is not touched. A foreign file at the
// target is an error unless force is set, in which case it is replaced.
func (l Layout) Install(source, target string, force bool) (Result, error) {
	res := Result{Source: source, Target: target, Action: "link"}
	state := l.Inspect(source, target)
	switch state {
	case Linked:
		res.Action = "unchanged"
		res.Skipped = true
		return res, nil
	case SourceMissing:
		return res, errors.New("source is missing")
	case Foreign, Elsewhere:
		if !force {
			return res, fmt.Errorf("target exists (%s)", state)
		}
		if err := os.Remove(target); err != nil {
			return res, err
		}
		res.Action = "replace"
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return res, err
	}
	if err := os.Symlink(source, target); err != nil {
		return res, err
	}
	return res, nil
}

// Remove deletes the symlink for one entry. Only links that point back at the
// given source are removed. prune asks for now-empty parent directories up to
// the home directory to be cleaned up as well.
func (l Layout) Remove(source, target string, prune bool) (Result, error) {
	res := Result{Source: source, Target: target, Action: "unlink"}
	switch l.Inspect(source, target) {
	case Absent:
		if _, err := os.Lstat(target); os.IsNotExist(err) {
			res.Action = "missing"
			res.Skipped = true
			return res, nil
		}
	case Linked, Dangling:
		if err := os.Remove(target); err != nil {
			return res, err
		}
		if prune {
			pruneDirs(filepath.Dir(target), l.Home)
		}
		return res, nil
	}
	res.Skipped = true
	return res, fmt.Errorf("target is not managed (%s)", l.Inspect(source, target))
}

func pruneDirs(dir, stop string) {
	for {
		if filepath.Clean(dir) == filepath.Clean(stop) || len(dir) <= len(stop) {
			return
		}
		if err := os.Remove(dir); err != nil {
			return
		}
		dir = filepath.Dir(dir)
	}
}
