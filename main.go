// Command gh-octomerge is a GitHub CLI extension that guides you through
// installing the octomerge GitHub App on one of your organizations.
package main

import (
	"os"

	"github.com/octomerge/gh-octomerge/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
