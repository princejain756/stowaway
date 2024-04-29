// Command stowaway keeps dotfiles in a repository and links them into $HOME.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"stowaway/cmd/install"
	"stowaway/cmd/uninstall"
	"stowaway/internal/profile"
)

const version = "0.4.0"

const usage = `stowaway keeps a repository of dotfiles and links them into your home directory.

usage: stowaway <command> [flags]

commands:
  install     link the entries of a profile into your home directory
  uninstall   remove the links a profile installed
  version     print the version
  help        print this text

Run "stowaway <command> -h" for the flags of a command.
`

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		var exit exitError
		if errors.As(err, &exit) {
			os.Exit(exit.code)
		}
		fmt.Fprintln(os.Stderr, "stowaway:", err)
		os.Exit(1)
	}
}

// exitError carries a status code for runs that ended without an internal
// failure, such as a diff that found drift.
type exitError struct{ code int }

func (e exitError) Error() string { return fmt.Sprintf("exit %d", e.code) }

func run(args []string, out, errOut io.Writer) error {
	if len(args) == 0 {
		fmt.Fprint(errOut, usage)
		return exitError{2}
	}
	command, rest := args[0], args[1:]
	switch command {
	case "help", "-h", "--help":
		fmt.Fprint(out, usage)
		return nil
	case "version":
		fmt.Fprintf(out, "stowaway %s\n", version)
		return nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	fs := flag.NewFlagSet(command, flag.ContinueOnError)
	fs.SetOutput(errOut)
	fs.Usage = func() { fmt.Fprintf(errOut, "usage: stowaway %s [flags]\n", command) }

	switch command {
	case "install":
		opts := install.ParseFlags(fs)
		if err := parse(fs, rest); err != nil {
			return err
		}
		repo, err := repoFor(cwd)
		if err != nil {
			return err
		}
		_, err = install.Run(home, repo, *opts, out, errOut)
		return err
	case "uninstall":
		opts := uninstall.ParseFlags(fs)
		if err := parse(fs, rest); err != nil {
			return err
		}
		repo, err := repoFor(cwd)
		if err != nil {
			return err
		}
		return uninstall.Run(home, repo, *opts, out, errOut)
	}
	fmt.Fprintf(errOut, "stowaway: unknown command %q\n\n%s", command, usage)
	return exitError{2}
}

func parse(fs *flag.FlagSet, args []string) error {
	if err := fs.Parse(args); err != nil {
		return exitError{2}
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}
	return nil
}

// repoFor finds the repository holding the working directory, preferring the
// STOWAWAY_REPO override when it is set.
func repoFor(cwd string) (string, error) {
	if override := os.Getenv("STOWAWAY_REPO"); override != "" {
		abs, err := filepath.Abs(override)
		if err != nil {
			return "", err
		}
		return abs, nil
	}
	return profile.Root(cwd)
}
