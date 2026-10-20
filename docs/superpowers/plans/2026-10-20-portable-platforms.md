# Portable Platform Support Implementation Plan

> **For agentic workers:** Implement the tasks in order with red/green tests. Keep the existing WSL contract and preserve existing files.

**Goal:** Run the published CLI on WSL Linux x64 and native macOS arm64, with a small path boundary and a declarative binary target list that can grow to other operating systems and architectures.

**Architecture:** `internal/hostpath` resolves user input into a native absolute path; Vault, config, resources, and Obsidian retain their own file lifecycle rules. One JSON target catalog drives package selection and build targets. Each binary has a platform-restricted npm package; the main package selects it automatically.

**Tech Stack:** Go 1.27.2, Node 22.23.3, npm platform packages, GitHub Actions.

## Global Constraints

- Preserve existing documents, settings, credentials, and plugin activation files.
- Plugin downloads require the existing explicit user opt-in.
- Preserve WSL Windows-drive conversion and local absolute-path handling.
- Use `mise.toml` versions and isolated temporary directories for tests.
- Do not present cross-compilation as proof of native execution.

## Task 1: Path Boundary and macOS Filesystem Compatibility

**Files:** Create `internal/hostpath/path.go`, `internal/hostpath/wsl.go`, `internal/hostpath/path_test.go`; modify `internal/vault/vault.go`, `internal/cli/init.go`, `internal/cli/wizard.go`, `internal/cli/root.go`, `internal/resources/install.go`, `internal/obsidian/plugin.go` and focused tests.

**Interfaces:** `hostpath.Resolver` resolves user input. `hostpath.ForCurrentHost()` selects native paths and WSL conversion. `resources.File` carries the root whose descendants are managed, so ancestor checks stop there.

- [ ] Add tests for native absolute paths, rejected Windows paths outside WSL, WSL conversion, and cause preservation. Run `mise exec -- go test ./internal/hostpath ./internal/cli` and confirm the intended failure.
- [ ] Implement `hostpath.Resolver` with `Resolve(context.Context, string) (string, error)` and injected Windows conversion. Preserve shell-free `wslpath` execution.
- [ ] Route both `init --vault` and the wizard through the same resolver, then move path tests out of `vault`; run focused tests to green.
- [ ] Add an integration test with a Vault under a symlinked system ancestor but a real managed root, while keeping rejection of symlinks *inside* the managed root. Confirm the red result on macOS.
- [ ] Bound resource and plugin ancestor validation to their declared root. Run `mise exec -- go test ./...` and confirm the previously failing macOS cases pass.

## Task 2: OS/Architecture npm Target Catalog

**Files:** Create `npm/learning-wiki/bin/targets.json`, `npm/learning-wiki-darwin-arm64/package.json`; modify `npm/learning-wiki/package.json`, `npm/learning-wiki/bin/learning-wiki.cjs`, `scripts/package.mjs`, `tests/launcher.test.mjs`, `scripts/test-package.mjs`.

**Interfaces:** Target catalog maps Node `platform`/`arch` to npm package, Go `GOOS`/`GOARCH`, and executable basename. Package manifests and catalog must agree exactly. The main package retains `bin/learning-wiki.cjs`.

- [ ] Write launcher tests for Linux x64, Darwin arm64, unsupported platforms, missing platform package, argument forwarding, and exit/signal forwarding; run `mise exec -- npm test` to observe target-selection failures.
- [ ] Implement catalog lookup in the launcher, preserving error codes and process behavior. Run `mise exec -- npm test` to green.
- [ ] Add a package contract test for catalog/manifests/version alignment and native package selection. Confirm it fails before changing the packer.
- [ ] Make `scripts/package.mjs` iterate the catalog, build `CGO_ENABLED=0` binaries with matching `GOOS`/`GOARCH`, include notices, pack all packages, and record artifacts. Validate manifests before packaging.
- [ ] Make `scripts/test-package.mjs` install the main tarball plus only the current host's platform tarball. Run `mise exec -- npm run pack:local` and `mise exec -- npm run test:package` on macOS.

## Task 3: Native Verification and Documentation

**Files:** Modify `.github/workflows/ci.yml`, `README.md`, `docs/npm.md`, `docs/development.md`, `docs/obsidian.md`, `docs/capture.md`, `internal/resources/assets/SKILL.md`, and environment-specific CLI help.

- [ ] Run the Go and Node tests plus local package smoke test on macOS arm64. Record what was actually exercised.
- [ ] Add Linux x64 and macOS arm64 CI jobs that each execute the locally packed native binary and verify repeatable initialization in a temporary Vault.
- [ ] Update supported-platform, path-input, Obsidian, and capture instructions without claiming unverified plugin or cloud integration.
- [ ] Run `mise exec -- go test ./...`, `mise exec -- go test -race ./...`, `mise exec -- npm test`, `mise exec -- npm run pack:local`, `mise exec -- npm run test:package`, and `git diff --check` before committing.
- [ ] Commit and push the branch, then create a PR targeting `main` with the verification evidence and any environment limits.
