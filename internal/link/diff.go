package link

import (
	"fmt"
	"io"
	"sort"
	"text/tabwriter"
)

// Finding is one entry whose target drifted from what the profile describes.
type Finding struct {
	Source string
	Target string
	State  State
	// Note carries extra context: the current link destination when the target
	// points elsewhere, or the difference in file contents when it is foreign.
	Note string
}

// Report is the result of comparing a set of entries against the filesystem.
type Report struct {
	Findings []Finding
	Checked  int
}

// Clean reports whether the installation matched the profile exactly.
func (r Report) Clean() bool { return len(r.Findings) == 0 }

// Sort orders findings by target path for stable output.
func (r Report) Sort() {
	sort.Slice(r.Findings, func(i, j int) bool {
		return r.Findings[i].Target < r.Findings[j].Target
	})
}

// Write renders the report, one finding per line. It returns the number of
// findings written.
func (r Report) Write(w io.Writer) (int, error) {
	r.Sort()
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for _, f := range r.Findings {
		line := fmt.Sprintf("%s\t%s", f.State, f.Target)
		if f.Note != "" {
			line += "\t" + f.Note
		}
		if _, err := fmt.Fprintln(tw, line); err != nil {
			return 0, err
		}
	}
	if err := tw.Flush(); err != nil {
		return 0, err
	}
	return len(r.Findings), nil
}

// Diff compares every entry against the filesystem. A target that is already
// the expected link is not reported.
func (l Layout) Diff(entries [][2]string) Report {
	var rep Report
	for _, e := range entries {
		source, target := l.Source(e[0]), l.Target(e[1])
		state := l.Inspect(source, target)
		rep.Checked++
		if state == Linked {
			continue
		}
		rep.Findings = append(rep.Findings, Finding{
			Source: source,
			Target: target,
			State:  state,
			Note:   l.note(source, target, state),
		})
	}
	return rep
}

func (l Layout) note(source, target string, state State) string {
	switch state {
	case Elsewhere:
		if dest, err := readlinkPath(target); err == nil {
			return "points at " + dest
		}
	case Foreign:
		if same, err := sameContents(source, target); err == nil && same {
			return "identical contents"
		}
		return "regular file"
	case Dangling:
		return "source deleted"
	}
	return ""
}
