package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/twotwobread/learning-wiki/internal/config"
	"github.com/twotwobread/learning-wiki/internal/resources"
	"github.com/twotwobread/learning-wiki/internal/vault"
)

func newInitCommand() *cobra.Command {
	var path string
	var create bool
	var agents []string
	cmd := &cobra.Command{
		Use:   "init [--vault <absolute-path>] [--create] [--agents pi,codex]",
		Short: "Connect a local vault and install learning resources",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("%w: init accepts no positional arguments", ErrInvalidArguments)
			}
			if err := resources.ValidateAgents(agents); err != nil {
				return fmt.Errorf("%w: %w", ErrInvalidArguments, err)
			}
			if path == "" && (cmd.Flags().Changed("vault") || cmd.Flags().Changed("create") || cmd.Flags().Changed("agents")) {
				return fmt.Errorf("%w: --vault requires a non-empty value; --create/--agents require --vault", ErrInvalidArguments)
			}
			if path == "" && (!isTerminal(cmd.InOrStdin()) || !isTerminal(cmd.OutOrStdout())) {
				return fmt.Errorf("%w: interactive setup requires a terminal; provide --vault <absolute-path> [--create] [--agents pi,codex]", ErrInvalidArguments)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, err := config.Path()
			if err != nil {
				return err
			}
			var resolved string
			if path == "" {
				plan, err := planWizard(cmd.Context(), huhPrompt{in: cmd.InOrStdin(), out: cmd.OutOrStdout()}, configPath)
				if err != nil {
					return err
				}
				resolved, create, agents = plan.Path, plan.Create, plan.Agents
			} else {
				resolved, err = vault.Resolve(cmd.Context(), path, nil)
				if err != nil {
					return err
				}
			}
			if err = config.Check(configPath, resolved); err != nil {
				return err
			}
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			files, err := resources.Plan(home, resolved, configPath, agents)
			if err != nil {
				return err
			}
			if err = resources.Check(files); err != nil {
				return err
			}
			if err = vault.Prepare(resolved, create); err != nil {
				return fmt.Errorf("prepare vault: %w", err)
			}
			if err = config.Save(configPath, resolved); err != nil {
				return fmt.Errorf("vault directories prepared but local configuration not saved; safe to retry: %w", err)
			}
			if err = resources.Install(files); err != nil {
				return fmt.Errorf("vault/configuration prepared but resources not fully installed; inspect conflict before retrying: %w", err)
			}
			if _, err = fmt.Fprintf(cmd.OutOrStdout(), "Vault ready: %s\nLocal configuration: %s\nLearning-note template installed.\n", resolved, configPath); err != nil {
				return err
			}
			if len(agents) > 0 {
				if _, err = fmt.Fprintln(cmd.OutOrStdout(), "Shared learning-capture skill installed in ~/.agents/skills/. Requires Node.js. pi: /reload then /skill:learning-capture. Codex: request learning-capture; restart if not discovered. Sandbox access to the vault may need approval."); err != nil {
					return err
				}
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Open this folder in Windows Obsidian. Obsidian plugin installation is not implemented yet.")
			return err
		},
	}
	cmd.Flags().StringVar(&path, "vault", "", "Windows drive or WSL absolute path to the local vault")
	cmd.Flags().BoolVar(&create, "create", false, "Allow creation of a missing vault directory")
	cmd.Flags().StringSliceVar(&agents, "agents", nil, "Install shared capture skill for pi,codex (visible to both); omit to install template only")
	return cmd
}
