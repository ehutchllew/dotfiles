// Command issue-tracker reads and writes feature specs and tickets.
package main

import (
	"fmt"
	"os"

	"github.com/ehutchllew/dotfiles/tools/issue-tracker/internal/cli"
)

func main() {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "issue-tracker: can't read the working directory:", err)
		os.Exit(1)
	}
	os.Exit(cli.Run(os.Args[1:], cli.Env{Dir: dir, Stdin: os.Stdin, Stdout: os.Stdout, Stderr: os.Stderr}))
}
