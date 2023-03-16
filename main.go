// Command stowaway keeps dotfiles in a repository and links them into $HOME.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
)

const version = "0.4.0"

const usage = `stowaway keeps a repository of dotfiles and links them into your home directory.

usage: stowaway <command> [flags]

commands:
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
	command := args[0]
	switch command {
	case "help", "-h", "--help":
		fmt.Fprint(out, usage)
		return nil
	case "version":
		fmt.Fprintf(out, "stowaway %s\n", version)
		return nil
	}
	fmt.Fprintf(errOut, "stowaway: unknown command %q\n\n%s", command, usage)
	return exitError{2}
}
