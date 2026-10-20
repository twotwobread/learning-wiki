# Default learning-capture installation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Install the shared `learning-capture` skill when `init` runs with its default options, while allowing an explicit skip.

**Architecture:** Keep the existing resource planner and shared skill directory. Change the CLI default selection and add a mutually exclusive `--no-skill` flag. Preselect both agents in the interactive form, whose empty selection still means template only.

**Tech Stack:** Go, Cobra, Huh, Node.js npm package smoke test.

## Global Constraints

- Preserve existing user files; a conflict must stop before Vault or config writes.
- Keep the Obsidian plugin behind its current explicit consent.
- Use `t.TempDir()` for HOME and Vault tests. Use the versions in `mise.toml`.
- Existing `--agents pi`, `--agents codex`, and `--agents pi,codex` continue to select the one shared skill.

---

### Task 1: CLI default and explicit skip

**Files:**
- Modify: `internal/cli/init.go`
- Modify: `internal/cli/resources_test.go`
- Modify: `internal/cli/init_test.go`

**Interfaces:** `newInitCommandWithResolver` continues to pass `[]string` to `resources.Plan`; `--no-skill` converts the default to nil and conflicts with explicitly provided `--agents`.

- [x] **Step 1: Write failing tests.** In `resources_test.go`, remove `--agents pi,codex` from the existing repeat-init test and assert `SKILL.md`, `binding.json`, and `capture.mjs` exist and are preserved. Add a test running `init --vault <temp> --create --no-skill` and assert the template/config exist while the shared skill and plugin files do not. In `init_test.go`, add a test that `--no-skill --agents codex` returns exit 2 without creating Vault/config and that bare `--no-skill` needs `--vault`.
- [x] **Step 2: Verify red.** Run `mise exec -- go test ./internal/cli -run 'TestInit(Default|NoSkill|SelectedAgents)' -count=1`; expect default skill absence or unknown flag failure.
- [x] **Step 3: Implement.** Set `--agents` default to `[]string{"pi", "codex"}`. Add `--no-skill`, reject it when `--agents` was explicitly changed, include it in the noninteractive `--vault` validation, and pass nil agents to `resources.Plan` when selected. Update flag/help text to describe the default.
- [x] **Step 4: Verify green.** Run the same focused command and `mise exec -- go test ./internal/cli`.
- [x] **Step 5: Commit.** Commit only the Go CLI/tests when passing.

### Task 2: Interactive default and package/docs integration

**Files:**
- Modify: `internal/cli/wizard.go`
- Modify: `internal/cli/wizard_test.go`
- Modify: `scripts/test-package.mjs`
- Modify: `docs/capture.md`, `docs/development.md`, `docs/npm.md`

**Interfaces:** `huhPrompt.Agents` returns the selected agent IDs; nil remains the explicit no-skill result.

- [x] **Step 1: Write failing test.** Add a focused test for the initial Huh selection through the form's behavior, or a small pure default-selection helper if terminal input cannot be driven reliably. Keep a scripted prompt test showing explicit empty selection remains empty and confirmation cancellation writes nothing. Change the package smoke to omit `--agents` and check all three skill files.
- [x] **Step 2: Verify red.** Run `mise exec -- go test ./internal/cli -run 'TestWizard(Default|Deselection)' -count=1` and `mise exec -- npm run test:package` against a freshly built package; expect the default-selection or package assertion to fail.
- [x] **Step 3: Implement.** Initialize Huh's selected values to both agent IDs. Clarify the selection and final confirmation text. Document default install, `--no-skill`, Codex discovery/restart, and 0.2.0's explicit `--agents` requirement.
- [x] **Step 4: Verify green.** Run `mise exec -- go test ./...`, `mise exec -- go test -race ./...`, `mise exec -- npm test`, `mise exec -- npm run pack:local`, and `mise exec -- npm run test:package`; inspect `git diff --check`.
- [x] **Step 5: Commit.** Commit the wizard, package smoke, and docs after verification.
