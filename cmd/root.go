// Package cmd wires up the gh-octomerge command line. It follows a
// command-first design: this layer only parses flags and delegates to the
// install domain package, which holds all of the actual behavior.
package cmd

import (
	"context"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"

	"github.com/octomerge/gh-octomerge/install"
)

// version is overridden at release time via -ldflags, e.g.
// -X github.com/octomerge/gh-octomerge/cmd.version=v1.2.3
var version = "dev"

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "octomerge",
		Short: "Install the octomerge GitHub App on your organization",
		Long: "octomerge is a gh extension that walks you through installing the\n" +
			"octomerge GitHub App (\"Your GitHub merging assistant\") on one of your\n" +
			"GitHub organizations.",
		SilenceUsage:  true, // fang renders errors; don't dump usage on a RunE failure
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			org, _ := cmd.Flags().GetString("org")
			yes, _ := cmd.Flags().GetBool("yes")
			return install.Run(cmd.Context(), install.Options{Org: org, AutoConfirm: yes})
		},
	}
	root.Flags().StringP("org", "o", "", "target GitHub organization (skips the picker)")
	root.Flags().BoolP("yes", "y", false, "skip the confirmation prompt")
	return root
}

// Execute builds the root command and runs it through fang, which adds styled
// help, a --version flag, error rendering, and signal-aware cancellation.
func Execute() error {
	return fang.Execute(context.Background(), newRootCmd(), fang.WithVersion(version))
}
