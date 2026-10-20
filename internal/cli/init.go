package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/twotwobread/learning-wiki/internal/config"
	"github.com/twotwobread/learning-wiki/internal/hostpath"
	"github.com/twotwobread/learning-wiki/internal/obsidian"
	"github.com/twotwobread/learning-wiki/internal/resources"
	"github.com/twotwobread/learning-wiki/internal/vault"
)

func newInitCommandWithFetcher(fetch obsidian.Fetcher) *cobra.Command {
	return newInitCommandWithResolver(fetch, hostpath.ForCurrentHost())
}

func newInitCommandWithResolver(fetch obsidian.Fetcher, resolver hostpath.Resolver) *cobra.Command {
	var plugin bool
	var noSkill bool
	var path string
	var create bool
	var agents []string
	cmd := &cobra.Command{
		Use:   "init [--vault <absolute-path>] [--create] [--agents pi,codex | --no-skill] [--plugin]",
		Short: "Connect a local vault and install learning resources",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("%w: init accepts no positional arguments", ErrInvalidArguments)
			}
			if noSkill && cmd.Flags().Changed("agents") {
				return fmt.Errorf("%w: --no-skill cannot be combined with --agents", ErrInvalidArguments)
			}
			if err := resources.ValidateAgents(agents); err != nil {
				return fmt.Errorf("%w: %w", ErrInvalidArguments, err)
			}
			if path == "" && (cmd.Flags().Changed("vault") || cmd.Flags().Changed("create") || cmd.Flags().Changed("agents") || cmd.Flags().Changed("plugin") || cmd.Flags().Changed("no-skill")) {
				return fmt.Errorf("%w: --vault requires a non-empty value; --create/--agents/--no-skill/--plugin require --vault", ErrInvalidArguments)
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
				plan, err := planWizardWithResolver(cmd.Context(), huhPrompt{in: cmd.InOrStdin(), out: cmd.OutOrStdout()}, configPath, resolver)
				if err != nil {
					return err
				}
				resolved, create, agents, plugin = plan.Path, plan.Create, plan.Agents, plan.Plugin
			} else {
				resolved, err = resolver.Resolve(cmd.Context(), path)
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
			if noSkill {
				agents = nil
			}
			files, err := resources.Plan(home, resolved, configPath, agents)
			if err != nil {
				return err
			}
			if err = resources.Check(files); err != nil {
				return err
			}
			var pluginResult obsidian.Result
			if plugin {
				if _, err = fmt.Fprintf(cmd.OutOrStdout(), "Plugin request: %s %s (Apache-2.0), Obsidian >= %s required. Existing installations are preserved.\n", obsidian.Repository, obsidian.Version, obsidian.MinAppVersion); err != nil {
					return err
				}
				pluginResult, err = obsidian.Prepare(cmd.Context(), resolved, obsidian.PinnedRelease(), fetch)
				if err != nil {
					return err
				}
				files = append(files, pluginResult.Files...)
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
			if plugin {
				if pluginResult.Existing {
					_, err = fmt.Fprintf(cmd.OutOrStdout(), "Existing plugin %s preserved; no upgrade or provenance check performed.\n", pluginResult.Version)
				} else {
					_, err = fmt.Fprintf(cmd.OutOrStdout(), "Plugin %s files installed with SHA-256 verification.\n", pluginResult.Version)
				}
				if err != nil {
					return err
				}
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Open this vault in Obsidian. If installing plugin files, close/reopen Obsidian, review community-plugin trust, enable Karpathy LLM Wiki, then configure your provider there. No credentials or activation settings were changed.")
			return err
		},
	}
	cmd.Flags().StringVar(&path, "vault", "", "Absolute path to the local vault (WSL also accepts Windows drive paths)")
	cmd.Flags().BoolVar(&create, "create", false, "Allow creation of a missing vault directory")
	cmd.Flags().StringSliceVar(&agents, "agents", []string{"pi", "codex"}, "Install shared capture skill for pi,codex (visible to both; installed by default)")
	cmd.Flags().BoolVar(&noSkill, "no-skill", false, "Skip the default shared capture skill installation")
	cmd.Flags().BoolVar(&plugin, "plugin", false, "Consent to download Karpathy LLM Wiki 1.28.0 from GitHub (Apache-2.0); requires Obsidian >= 1.11.4; preserve existing installs")
	return cmd
}
