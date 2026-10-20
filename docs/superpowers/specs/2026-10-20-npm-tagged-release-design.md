# Tagged npm Release Design

## Goal and release boundary

Publish the first `0.2.0` release manually from the three tarballs produced and tested by the successful `main` CI run for commit `be0997c13482eca8a313e653d34c6ee9165161be`. This establishes all three npm package names. Later releases start from a `vX.Y.Z` tag on `main`, require a GitHub Environment approval, and publish with npm Trusted Publishing (OIDC).

The public packages are `@suyoung9137/learning-wiki-linux-x64`, `@suyoung9137/learning-wiki-darwin-arm64`, and `@suyoung9137/learning-wiki`. The two platform packages must be available before the main package is published because the main package declares them as optional dependencies at the exact same version.

## First release: manual `0.2.0`

The operator signs in to npm in their own terminal. No npm credential or OTP is stored in this repository, GitHub Variables, or chat. Confirm the npm identity is `suyoung9137` and that `0.2.0` is absent for each package. Download the tarballs from the successful `main` CI run, compare the Linux and macOS job copies, verify each file against `artifacts.json`, and inspect the package contents. Publish the exact verified tarballs with `--access public`, platform packages first and main package last. After each publish, compare the registry's `dist.integrity` with the verified tarball integrity. Install `@suyoung9137/learning-wiki@0.2.0` from the registry on macOS arm64 and run the CLI in an isolated temporary directory. Perform the equivalent Linux x64 registry-install smoke test on a GitHub-hosted runner or WSL; report it as unverified until it runs. Create the `v0.2.0` tag at the released `main` commit after publication succeeds.

If a publish stops partway through, do not republish an existing name/version or silently continue. Verify the already published file's registry integrity against the release artifact, then resume with the remaining packages. A different integrity at the same name/version is an irrecoverable release mismatch and requires human review and a new version.

## Later releases: workflow architecture

Add `.github/workflows/release.yml` with a `v*` tag trigger. The workflow checks out the tagged commit with enough history to verify that it belongs to `main`. A preflight verifies that the tag has the exact form `vX.Y.Z`, the main and both platform package versions equal `X.Y.Z`, the target catalog matches the package manifests, and the repository is clean before building. This preflight does not write to npm.

A single build job uses the Go and Node versions pinned in `mise.toml`, runs the project tests, runs the existing `npm run pack:local`, verifies each tarball's name, version, and integrity, and uploads exactly those three tarballs plus `artifacts.json` as an immutable workflow artifact. Native Linux x64 and macOS arm64 jobs download that same artifact and execute `npm run test:package` from the downloaded files. The publish job depends on both native jobs, downloads the same artifact, verifies its manifest and integrity again, and uses GitHub Environment `npm-release` with required reviewer approval and a `main` tag restriction.

The publish job grants only `contents: read` and `id-token: write`, installs npm CLI `11.21.0` with the pinned Node `22.23.3`, and publishes through npm Trusted Publishing. Each of the three npm packages must trust the repository, exact workflow filename `release.yml`, and environment `npm-release`, with direct `npm publish` permission. No long-lived npm token is used. For each package in platform-first order, the job queries whether the exact version already exists. It skips only when the registry integrity equals the verified artifact; it fails on a mismatch. The main package is published last. A subsequent read-only registry check confirms all three versions and integrities.

OIDC setup occurs only after the manual `0.2.0` publication has established the package names. A newly created npm Trusted Publisher must complete its first successful publish within two days, so create the three trust relationships shortly before the first automated release rather than immediately after `0.2.0`. The repo is public and the workflow uses GitHub-hosted runners, allowing npm to attach provenance when its Trusted Publishing prerequisites are met. Changes to npm package trust settings and the GitHub Environment are operational setup steps documented separately from the source workflow.

## Failure and verification contracts

- A tag not on `main`, malformed tag, version mismatch, missing platform package, or changed artifact integrity fails before approval or publish.
- A native smoke failure prevents the publish job from starting.
- An unapproved Environment job cannot publish.
- A retry after a partial publish skips only byte-identical existing package versions and publishes remaining versions in order.
- Tests use temporary HOME/Vault paths and do not install the optional Obsidian plugin or alter existing user configuration.
- The release verification reports CLI installation and file behavior separately from Obsidian app integration, cloud synchronization, and learning quality.

## Scope

The workflow automates packaging and npm release for the existing Linux x64 and macOS arm64 targets. It does not add platforms, update plugin generation, or alter the `sessions/` source-content contract. The first `0.2.0` publication remains a manual operation; later releases use the tag workflow.
