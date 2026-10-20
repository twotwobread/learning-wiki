# Default learning-capture installation

## Goal

Running `learning-wiki init` should install the shared `learning-capture` skill alongside the Vault template and local configuration. This applies to a fresh installation and a repeat run on the same Vault. The existing npm `0.2.0` release remains unchanged; users of that release can add the skill now with `--agents codex`.

## Options considered

1. **Default installation with an explicit skip (chosen):** Install the shared skill by default and provide `--no-skill` for scripted use. The wizard preselects pi and Codex, and users may deselect both. This meets the expected default while retaining a clear way to decline the file creation.
2. **Always install:** Less CLI state, but removes the ability to decline a user-home file write.
3. **Keep opt-in and improve the message:** Avoids changed behavior but does not meet the requested default.

## Behavior and boundaries

- For `init --vault <path>`, omitted `--agents` means the shared skill is installed at `$HOME/.agents/skills/learning-capture/`. The template and config keep their existing locations and preservation rules. `--agents pi`, `--agents codex`, and `--agents pi,codex` remain accepted for compatibility; all select the same shared skill.
- `--no-skill` explicitly installs only the template and config. It cannot be combined with `--agents`, because those flags express conflicting requests. It requires `--vault` in noninteractive use, like other init flags.
- In the interactive wizard, pi and Codex start selected. Pressing Enter with the defaults installs the shared skill. Deselecting both explicitly requests template and config only. The final confirmation describes the selected outcome before any writes. Cancelling still creates no files.
- The optional Obsidian plugin remains off by default and requires its existing explicit consent. This change does not modify plugin, authentication, activation, `sessions/`, or generated `wiki/` files.
- Existing skills and templates remain byte-for-byte preserved on repeat init; conflicting user changes still stop installation before writes. A previous `0.2.0` template-only installation can gain the skill by rerunning init against the same Vault.
- The installed skill is portable Markdown plus Node script. No separate Codex plugin install or `.codex/skills` copy is created. Codex discovers `$HOME/.agents/skills`; restart Codex if the new skill is not shown.

## Verification

- CLI tests in isolated temporary HOME/Vault verify default install, repeat idempotence, explicit `--no-skill`, contradictory flags, refusal/cancel, and absence of unintended plugin files.
- Wizard tests verify default selections, explicit deselection, and final confirmation without relying on exact terminal rendering.
- Local package smoke test installs a built tarball in temporary HOME and runs `init` without `--agents`, then checks `SKILL.md`, `binding.json`, and `scripts/capture.mjs` before trying the capture helper.
- Actual Codex skill discovery and Obsidian app behavior remain separate manual checks.
