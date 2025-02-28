// Package list prints the profiles a repository defines.
package list

import (
	"flag"
	"fmt"
	"io"

	"stowaway/internal/profile"
)

// Run writes one profile name per line to out and returns how many there are.
func Run(repo string, out io.Writer) (int, error) {
	names, err := profile.List(repo)
	if err != nil {
		return 0, err
	}
	for _, name := range names {
		fmt.Fprintln(out, name)
	}
	return len(names), nil
}

// ParseFlags binds the list flags. It takes none and exists so the command
// table stays uniform.
func ParseFlags(_ *flag.FlagSet) struct{} { return struct{}{} }
