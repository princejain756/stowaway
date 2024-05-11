// Package uninstall removes the links a profile installed.
package uninstall

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
	Prune   bool
	DryRun  bool
}

// Run unlinks the entries of opts.Profile from home.
func Run(home, repo string, opts Options, out, errOut io.Writer) error {
	p, err := profile.Load(repo, opts.Profile)
	if err != nil {
		return err
	}
	layout := link.Layout{Repo: repo, Home: home}
	failed := 0
	for _, e := range p.Entries {
		source, target := layout.Source(e.Source), layout.Target(e.Target)
		if opts.DryRun {
			state := layout.Inspect(source, target)
			if state != link.Linked && state != link.Dangling {
				fmt.Fprintf(errOut, "would skip %s (%s)\n", e.Target, state)
				failed++
				continue
			}
			fmt.Fprintf(out, "would unlink %s\n", e.Target)
			continue
		}
		r, err := layout.Remove(source, target, opts.Prune)
		if err != nil {
			failed++
			fmt.Fprintf(errOut, "%s: %v\n", e.Target, err)
			continue
		}
		if r.Skipped {
			continue
		}
		fmt.Fprintln(out, r.String())
	}
	if failed > 0 {
		return fmt.Errorf("%d of %d entries failed", failed, len(p.Entries))
	}
	return nil
}

// ParseFlags binds the uninstall flags.
func ParseFlags(fs *flag.FlagSet) *Options {
	opts := &Options{}
	fs.StringVar(&opts.Profile, "profile", "default", "profile to uninstall")
	fs.BoolVar(&opts.Prune, "prune", true, "remove directories left empty")
	fs.BoolVar(&opts.DryRun, "dry-run", false, "print the plan without touching the filesystem")
	return opts
}
