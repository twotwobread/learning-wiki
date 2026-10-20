# Tagged npm Release Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Publish future versions from a `vX.Y.Z` tag only after native package tests and a protected GitHub Environment approval.

**Architecture:** A small Node release tool validates package metadata and the exact tarballs from the existing pack script. One build job uploads those files, Linux and macOS jobs test the same artifact, and a final OIDC job publishes platform packages before the main package. Registry integrity checks make retries safe after partial publication.

**Tech Stack:** Node.js 22.23.3, Go 1.27.2, npm CLI 11.21.0 for publishing, GitHub Actions, npm Trusted Publishing.

## Global Constraints

- Publish exactly `@suyoung9137/learning-wiki-linux-x64`, `@suyoung9137/learning-wiki-darwin-arm64`, and `@suyoung9137/learning-wiki` at one stable `X.Y.Z` version.
- Require the tag `vX.Y.Z` on `main`; keep `v0.2.0` as the already completed manual release.
- Use a single immutable artifact containing `artifacts.json` and the three corresponding tarballs for both native tests and publication.
- The publish job uses environment `npm-release`, `contents: read`, and `id-token: write`; no npm token is stored.
- Create npm Trusted Publisher connections for workflow `release.yml` and environment `npm-release` shortly before the first automated release because unvalidated connections expire after two days.
- Tests use temporary HOME/Vault paths. Never install the optional Obsidian plugin in CI.

## File responsibilities

- `scripts/release.mjs`: release CLI commands `preflight`, `verify`, and `publish`; metadata checks, tarball integrity, registry comparison, and ordered publication.
- `tests/release.test.mjs`: black-box contract tests for metadata, artifact tampering, existing versions, mismatch, and partial publish retry.
- `.github/workflows/release.yml`: tag trigger; preflight, one build, two native smoke jobs, protected OIDC publish job.
- `docs/npm.md`: operator setup, release order, environment and npm trust configuration, retry behavior.

---

### Task 1: Preflight contract

**Files:**
- Create: `scripts/release.mjs`
- Create: `tests/release.test.mjs`

**Interfaces:**
- `validateRelease(tag, main, targets, platforms)` returns the version or throws for malformed tag, target mismatch, version mismatch, or optional dependency mismatch.
- CLI `node scripts/release.mjs preflight vX.Y.Z` additionally checks a clean checkout and `HEAD` ancestry in `origin/main`.

- [x] **Step 1: Write failing tests** for valid `v0.3.0`, malformed `v0.3`, mismatched platform version, missing optional dependency, and unsupported target metadata. Use literal fixture package objects and `assert.throws`.
- [x] **Step 2: Run** `mise exec -- node --test tests/release.test.mjs`; expect the import or export to fail because `release.mjs` does not exist.
- [x] **Step 3: Implement** `validateRelease` with `^v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$`; derive platform package expectations from every entry in `targets.json`, and require package names, `os`/`cpu` arrays, optional dependencies, and versions to match. In CLI preflight run `git diff --quiet`, `git diff --cached --quiet`, `git status --porcelain --untracked-files=normal`, and `git merge-base --is-ancestor HEAD origin/main` with argument arrays, never shell interpolation.
- [x] **Step 4: Run** `mise exec -- node --test tests/release.test.mjs` and `mise exec -- npm test`; expect all tests to pass.
- [x] **Step 5: Commit** `scripts/release.mjs` and `tests/release.test.mjs` with message `feat: validate tagged npm release metadata`.

### Task 2: Exact artifact verification

**Files:**
- Modify: `scripts/release.mjs`
- Modify: `tests/release.test.mjs`

**Interfaces:**
- `verifyArtifacts(dir, version)` returns the three manifest entries in platform-first/main-last order. Each entry contains `name`, `version`, `file`, and `integrity`.
- CLI `node scripts/release.mjs verify vX.Y.Z dist` exits nonzero for any invalid or extra artifact.

- [x] **Step 1: Write failing tests** using `t.TempDir` equivalent `mkdtemp` and tiny `.tgz` fixtures built with `tar`; assert exact expected names and filenames, correct SHA-512, tampering failure, missing or extra tarball failure, and package.json name/version mismatch inside the tarball.
- [x] **Step 2: Run** `mise exec -- node --test tests/release.test.mjs`; expect missing `verifyArtifacts` behavior.
- [x] **Step 3: Implement** manifest shape and exact file set checks; compute `sha512-` plus base64 from each tarball; read `package/package.json` with `tar -xOzf` via `execFileSync`; compare name/version; reject path escapes in manifest filenames.
- [x] **Step 4: Run** the focused tests and full `mise exec -- npm test`; expect all tests to pass.
- [x] **Step 5: Commit** with message `feat: verify exact npm release artifacts`.

### Task 3: Safe ordered publication

**Files:**
- Modify: `scripts/release.mjs`
- Modify: `tests/release.test.mjs`

**Interfaces:**
- `publishArtifacts(entries, dir, { lookup, publish, wait })` checks registry integrity before and after each publish, skips only a matching existing version, and publishes platform entries before main.
- CLI `node scripts/release.mjs publish vX.Y.Z dist` uses public npm registry reads and `npm publish ./<tarball> --access public` through `execFileSync`.

- [x] **Step 1: Write failing tests** with injected registry/publish functions: absent versions publish in platform-first order; identical existing versions skip; mismatched existing integrity stops; a successful publish waits for registry visibility before continuing; a partial retry publishes only remaining entries.
- [x] **Step 2: Run** `mise exec -- node --test tests/release.test.mjs`; expect missing `publishArtifacts` behavior.
- [x] **Step 3: Implement** GET `https://registry.npmjs.org/<encoded-package>/<version>` with 404 meaning absent and other non-200 responses failing. Poll for up to five minutes after publish; accept only the exact manifest integrity. Keep stdout free of credentials.
- [x] **Step 4: Run** focused and full Node tests; expect all to pass.
- [x] **Step 5: Commit** with message `feat: make npm publication retry safe`.

### Task 4: Protected GitHub release workflow and operator docs

**Files:**
- Create: `.github/workflows/release.yml`
- Modify: `docs/npm.md`

**Interfaces:**
- Workflow `push.tags: [v*]`; jobs `preflight`, `build`, `smoke` (Linux x64 and macOS arm64 matrix), `publish`.
- The build uploads `dist/artifacts.json` and exactly three tarballs. Smoke and publish jobs download the same named artifact into `dist`.

- [x] **Step 1: Review the workflow contract** against the approved design: tag trigger, environment approval, OIDC permission, native runners, artifact handoff, and job dependencies. Use `actionlint` for YAML, expression, action, and shell validation; avoid a static test that mirrors workflow text.
- [x] **Step 3: Add workflow**: checkout with full history and fetch `origin/main`; preflight before building; install mise-pinned tools; run Go tests, race tests, vet, Node tests, pack and verify; upload artifact; smoke on `ubuntu-24.04` and `macos-15` with `npm run test:package`; protected publish job installs npm `11.21.0`, verifies artifacts, and invokes release publish script. Use `concurrency` to prevent overlapping releases.
- [x] **Step 4: Update docs** with the 0.2.0 release status, the exact environment protection and npm trusted publisher settings, the two-day first-publish window, and the documented partial retry procedure.
- [x] **Step 5: Run** `actionlint .github/workflows/release.yml` if available, `mise exec -- npm test`, `mise exec -- go test ./...`, `mise exec -- go test -race ./...`, `mise exec -- npm run pack:local`, `mise exec -- node scripts/release.mjs verify v0.2.0 dist`, and `mise exec -- npm run test:package`; expect all checks to pass.
- [x] **Step 6: Commit** with message `ci: add protected OIDC npm release workflow`.

## Final review

- [ ] Check the workflow against the approved design and official npm/GitHub documentation.
- [ ] Run `git diff --check` and inspect every changed file.
- [ ] Push the branch and open a PR into `main`; leave GitHub Environment and npm Trusted Publisher connections unconfigured until just before the first automated release.
