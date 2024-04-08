// Package install links every entry of a profile into the home directory.
package install

import (
	"errors"
	"flag"
	"fmt"
	"io"

	"stowaway/internal/link"
	"stowaway/internal/profile"
)

// Options configures a run.
type Options struct {
	Profile string
	Force   bool
	DryRun  bool
}

// Result summarises what a run did.
type Result struct {
	Linked  int
	Skipped int
	Lines   []string
}

// Run installs the entries of opts.Profile from repo into home.
func Run(home, repo string, opts Options, out, errOut io.Writer) (Result, error) {
	var res Result
	p, err := profile.Load(repo, opts.Profile)
	if err != nil {
		return res, err
	}
	layout := link.Layout{Repo: repo, Home: home}
	failed := 0
	for _, e := range p.Entries {
		source, target := layout.Source(e.Source), layout.Target(e.Target)
		if opts.DryRun {
			state := layout.Inspect(source, target)
			line, err := describe(source, target, state, opts.Force)
			if err != nil {
				failed++
				fmt.Fprintf(errOut, "would fail %s: %v\n", e.Target, err)
				continue
			}
			res.Lines = append(res.Lines, line)
			continue
		}
		r, err := layout.Install(source, target, opts.Force)
		if err != nil {
			failed++
			fmt.Fprintf(errOut, "%s: %v\n", e.Target, err)
			continue
		}
		if r.Skipped {
			res.Skipped++
			continue
		}
		res.Linked++
		res.Lines = append(res.Lines, r.String())
	}
	for _, line := range res.Lines {
		fmt.Fprintln(out, line)
	}
	if failed > 0 {
		return res, fmt.Errorf("%d of %d entries failed", failed, len(p.Entries))
	}
	return res, nil
}

func describe(source, target string, state link.State, force bool) (string, error) {
	switch state {
	case link.Linked:
		return fmt.Sprintf("unchanged %s -> %s", source, target), nil
	case link.SourceMissing:
		return "", errors.New("source is missing")
	case link.Foreign, link.Elsewhere:
		if !force {
			return "", fmt.Errorf("target exists (%s)", state)
		}
		return fmt.Sprintf("replace %s -> %s", source, target), nil
	}
	return fmt.Sprintf("link %s -> %s", source, target), nil
}

// ParseFlags binds the install flags.
func ParseFlags(fs *flag.FlagSet) *Options {
	opts := &Options{}
	fs.StringVar(&opts.Profile, "profile", "default", "profile to install")
	fs.BoolVar(&opts.Force, "force", false, "replace files that are in the way")
	fs.BoolVar(&opts.DryRun, "dry-run", false, "print the plan without touching the filesystem")
	return opts
}
