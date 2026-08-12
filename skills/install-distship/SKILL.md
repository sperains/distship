---
name: install-distship
description: Install or upgrade DistShip from official GitHub Releases on supported macOS and Linux systems. Use when a user asks to install DistShip, update DistShip, download a specific DistShip version, verify a DistShip release archive, or diagnose whether the installed binary and PATH are ready for first use.
---

# Install DistShip

Install a verified DistShip release without building from source. Use the bundled installer to detect the platform, download the matching archive and checksum manifest, verify SHA-256, and install the binary into a user-selected directory.

## Workflow

1. Inspect the current state without changing it:

   ```bash
   uname -s
   uname -m
   command -v distship || true
   distship version 2>/dev/null || true
   ```

2. Select the version and destination:

   - Install the latest stable release unless the user names a version.
   - Default to `$HOME/.local/bin`; create it when needed.
   - Reuse an existing install directory only when the user is upgrading that installation and the directory is writable.
   - Ask before writing to a system-owned directory or using elevated privileges. Never introduce `sudo` without explicit approval.

3. Run the bundled installer from this skill directory:

   ```bash
   scripts/install.sh
   scripts/install.sh --version v0.1.0
   scripts/install.sh --install-dir /absolute/path
   ```

   Resolve `scripts/install.sh` relative to the directory containing this `SKILL.md`; do not assume the user's current working directory is the skill directory.

4. Verify and report:

   ```bash
   /absolute/install/path/distship version
   /absolute/install/path/distship --help
   ```

   Report the installed path and version. If the directory is not on `PATH`, provide the exact shell configuration line but do not edit shell startup files unless requested.

## Safety boundaries

- Support only the release platforms published by DistShip: macOS and Linux on x86-64 or ARM64.
- Download only from `github.com/sperains/distship` release URLs.
- Require a matching entry in the published checksum manifest before installation.
- Preserve DistShip configuration and deployment history during installs and upgrades.
- Do not remove an existing installation unless the user explicitly requests uninstallation.
- Stop and report unsupported platforms, missing verification tools, checksum failures, malformed version data, or binary version mismatches.

## First-use handoff

After installation succeeds, suggest only these next steps unless the user asks for deployment configuration:

```bash
distship version
distship init /absolute/path/to/project
```
