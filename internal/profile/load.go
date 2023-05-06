package profile

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Dir is the directory inside a repository that holds profile files.
const Dir = "profiles"

// Ext is the file extension of a profile file.
const Ext = ".conf"

// Path returns the file that would hold the named profile.
func Path(repo, name string) string {
	return filepath.Join(repo, Dir, name+Ext)
}

// Load reads the named profile from the repository rooted at repo.
func Load(repo, name string) (*Profile, error) {
	if err := ValidName(name); err != nil {
		return nil, err
	}
	body, err := os.ReadFile(Path(repo, name))
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	p, err := Parse(name, body)
	if err != nil {
		return nil, err
	}
	if dupes := p.DuplicateTargets(); len(dupes) > 0 {
		return nil, errors.New(strings.Join(dupes, ", ") + ": target claimed more than once")
	}
	return p, nil
}

// List returns the names of every profile in the repository, sorted.
func List(repo string) ([]string, error) {
	dir := filepath.Join(repo, Dir)
	names, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []string
	for _, n := range names {
		if n.IsDir() || !strings.HasSuffix(n.Name(), Ext) {
			continue
		}
		name := strings.TrimSuffix(n.Name(), Ext)
		if ValidName(name) == nil {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out, nil
}

// ValidName reports whether name is usable as a profile file name.
func ValidName(name string) error {
	switch {
	case name == "":
		return errors.New("empty profile name")
	case name == "." || name == "..":
		return errors.New("invalid profile name " + name)
	}
	if strings.ContainsAny(name, "/\\ \t") {
		return errors.New("invalid profile name " + name)
	}
	return nil
}
