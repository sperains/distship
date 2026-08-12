# Changelog

All notable changes to this project will be documented in this file.

The format is based on Keep a Changelog, and this project follows Semantic
Versioning.

## Unreleased

- Added Homebrew Cask installation and automatic Tap updates for stable releases.
- Updated GitHub Actions dependencies to current releases and pinned their
  immutable commits.
- Added a companion Codex skill for verified DistShip installation and upgrades.
- Redesigned the English and Chinese READMEs around project positioning,
  verified installation, quick start, deployment flow, scope, a preflight
  preview, and a collapsible initialization example.
- Replaced separate project and environment ID prompts with one deployment
  target concept, while keeping project detection automatic in the default
  initialization flow.
- Replaced environment-specific SSH prompt examples with neutral alias and
  user-qualified host examples.

## [0.1.0] - 2026-08-12

- Added reproducible GoReleaser builds for macOS and Linux on ARM64 and x86-64.
- Added a tag-triggered GitHub Release workflow with SHA-256 checksums and
  embedded version metadata.
- Added Go build information fallback for locally built and source-installed
  version output.
- Documented binary installation, upgrades, uninstallation, release
  verification, and the maintainer release process.
- Added English and Chinese SSH setup, fingerprint verification, and connection
  troubleshooting guides.
- Clarified that version 0.1 supports static frontend artifacts only and
  excludes backend or runtime service management.
- Added confirmed local builds, artifact validation, remote directory creation,
  incremental rsync upload, per-target deployment locks, and local success
  history.
- Added a network-free `deploy --dry-run` preview and deployment change ranges
  based on matching local history.
- Present deployment readiness first, condense successful checks, and show the
  exact Git revision with the three most recent non-merge commits.
- Initial project skeleton.
- Interactive configuration initialization.
- Configuration validation and grouped project listing.
- Built-in English and Simplified Chinese terminal interfaces with automatic
  locale detection.
- Streamlined initialization with detected defaults, an advanced mode, and
  automatic post-write validation.
- Accepted project directories as command arguments or required interactive
  input, with immediate validation and explicit detection results.
- Added explicit current-project detection and flexible SSH targets using
  aliases, hostnames, IP addresses, or user-qualified hosts.
- Accepted deployment targets as either a combined SSH target and remote path
  or two guided inputs.
- Replaced duplicate overwrite prompts with field-level diffs and a single
  confirmation for existing targets.
- Safe local target removal with confirmation, validation, and timestamped
  backup of the last configuration.
- Standardized target selection on the copyable `project:environment` ID.
- Reused configured target IDs when initialization is run again for the same
  local directory.
- Added read-only deployment preflight checks for local tools, Git policy, SSH
  connectivity, and remote directory permissions.
