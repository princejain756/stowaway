package profile

import (
	"errors"
	"os"
	"path/filepath"
)

// ErrNoRepo is returned by Root when no directory above the start path holds a
// profiles directory.
var ErrNoRepo = errors.New("no stowaway repository found (looked for a \"profiles\" directory)")

// Root walks up from start looking for a repository root, which is any
// directory containing a profiles directory.
func Root(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		info, err := os.Stat(filepath.Join(dir, Dir))
		if err == nil && info.IsDir() {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", ErrNoRepo
		}
		dir = parent
	}
}
