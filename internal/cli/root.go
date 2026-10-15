// Package cli defines the command-line interface without terminating the process.
package cli

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/twotwobread/learning-wiki/internal/vault"
)

// ErrInvalidArguments identifies invalid command-line input independently of wording.
var ErrInvalidArguments = errors.New("invalid command-line arguments")

// NewRootCommand returns an independent command tree for each invocation.
func NewRootCommand(version string) *cobra.Command {
	if version == "" {
		version = "dev"
	}
	cmd := &cobra.Command{
		Use:           "learning-wiki",
		Short:         "Collect learning from pi and Codex into a local Obsidian vault",
		Long:          "Learning Wiki connects WSL-based pi and Codex workflows to a local Obsidian vault. Use init to connect a local vault.",
		Version:       version,
		SilenceErrors: true,
		SilenceUsage:  true,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("%w: unexpected arguments %q", ErrInvalidArguments, args)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error { return cmd.Help() },
	}
	cmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		return fmt.Errorf("%w: %w", ErrInvalidArguments, err)
	})
	cmd.AddCommand(newInitCommand())
	cmd.CompletionOptions.DisableDefaultCmd = true
	return cmd
}

// Run routes output and returns 0 on success, 2 for invalid input, or 1 for other errors.
// Only the executable entry point may call os.Exit.
func Run(args []string, version string, stdout, stderr io.Writer) int {
	cmd := NewRootCommand(version)
	// Cobra treats nil as os.Args[1:]; use a non-nil empty slice for explicit no-argument runs.
	cmd.SetArgs(append([]string{}, args...))
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(stderr, err)
		if errors.Is(err, ErrInvalidArguments) || errors.Is(err, vault.ErrInvalidPath) {
			return 2
		}
		return 1
	}
	return 0
}
