// Package diff reports drift between a profile and the live home directory.
package diff

import (
	"flag"
	"fmt"
	"io"

	"stowaway/internal/link"
	"stowaway/internal/profile"
)

// Options configures a run.
type Options struct {
	Profile string
}

// Run compares the profile against home and writes one line per drifted entry.
// It returns the number of findings; a clean tree yields zero.
func Run(home, repo string, opts Options, out io.Writer) (int, error) {
	p, err := profile.Load(repo, opts.Profile)
	if err != nil {
		return 0, err
	}
	entries := make([][2]string, 0, len(p.Entries))
	for _, e := range p.Entries {
		entries = append(entries, [2]string{e.Source, e.Target})
	}
	layout := link.Layout{Repo: repo, Home: home}
	rep := layout.Diff(entries)
	if rep.Clean() {
		return 0, nil
	}
	n, err := rep.Write(out)
	if err != nil {
		return n, err
	}
	fmt.Fprintf(out, "%d of %d entries drifted\n", n, rep.Checked)
	return n, nil
}

// ParseFlags binds the diff flags.
func ParseFlags(fs *flag.FlagSet) *Options {
	opts := &Options{}
	fs.StringVar(&opts.Profile, "profile", "default", "profile to compare")
	return opts
}
