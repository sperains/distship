# Releasing DistShip

This document is for project maintainers. DistShip releases are built by GoReleaser after a semantic-version tag is pushed.

## Before tagging

1. Start from a clean `main` branch and confirm that CI passes.
2. Review the `Unreleased` section in `CHANGELOG.md`.
3. Run the local release checks:

   ```bash
   go test ./...
   go vet ./...
   goreleaser check
   goreleaser release --snapshot --clean
   ```

4. Inspect the four archives and `checksums.txt` under `dist/`.
5. Run the host-compatible binary and confirm that `distship version` contains the expected snapshot version, commit, and build time.

## Create a release

Use a semantic version prefixed with `v`, for example `v0.1.0`:

```bash
git tag -a v0.1.0 -m "DistShip v0.1.0"
git push origin v0.1.0
```

The tag starts the release workflow. It reruns tests and static checks, then publishes:

- `distship_<version>_Darwin_arm64.tar.gz`
- `distship_<version>_Darwin_x86_64.tar.gz`
- `distship_<version>_Linux_arm64.tar.gz`
- `distship_<version>_Linux_x86_64.tar.gz`
- `checksums.txt`
- an updated `distship` Cask in `sperains/homebrew-tap`

Do not move or reuse a published tag. If a release is incorrect, fix the source and publish a new patch version.

## Verify the published release

1. Download one archive and `checksums.txt` from the GitHub Release.
2. Verify the archive:

   ```bash
   shasum -a 256 --ignore-missing --check checksums.txt
   ```

   On Linux, use `sha256sum --ignore-missing --check checksums.txt` instead. The command verifies the downloaded archive and ignores entries for the other platforms.

3. Extract the archive and run:

   ```bash
   ./distship version
   ./distship --help
   ```

4. Confirm that the GitHub Release notes describe the intended changes.

5. Verify the Homebrew release:

   ```bash
   brew update
   brew upgrade --cask distship
   distship version
   ```

## Homebrew publishing

GoReleaser publishes the Cask to `sperains/homebrew-tap` with the
repository-specific deploy key stored in the `HOMEBREW_TAP_DEPLOY_KEY` Actions
secret. The public key must retain write access to the Tap repository.
Prerelease tags do not update the stable Cask.
