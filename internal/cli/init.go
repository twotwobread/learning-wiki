package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/twotwobread/learning-wiki/internal/config"
	"github.com/twotwobread/learning-wiki/internal/vault"
)

func newInitCommand() *cobra.Command {
	var path string
	var create bool
	cmd := &cobra.Command{
		Use:   "init --vault <absolute-path> [--create]",
		Short: "Connect a local vault and prepare learning-note directories",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 || path == "" {
				return fmt.Errorf("%w: init requires --vault and accepts no positional arguments", ErrInvalidArguments)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			resolved, err := vault.Resolve(cmd.Context(), path, nil)
			if err != nil {
				return err
			}
			configPath, err := config.Path()
			if err != nil {
				return err
			}
			if err = config.Check(configPath, resolved); err != nil {
				return err
			}
			if err = vault.Prepare(resolved, create); err != nil {
				return fmt.Errorf("prepare vault: %w", err)
			}
			if err = config.Save(configPath, resolved); err != nil {
				return fmt.Errorf("vault directories prepared but local configuration not saved; safe to retry: %w", err)
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Vault ready: %s\nLocal configuration: %s\nOpen this folder in Windows Obsidian. Skills and plugins are not installed yet.\n", resolved, configPath)
			return err
		},
	}
	cmd.Flags().StringVar(&path, "vault", "", "Windows drive or WSL absolute path to the local vault")
	cmd.Flags().BoolVar(&create, "create", false, "Allow creation of a missing vault directory")
	return cmd
}
