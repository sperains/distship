# Installing DistShip

DistShip publishes checksum-verified archives for macOS and Linux. Use the
Homebrew Cask, the Codex skill for a guided installation, or a release archive.

## Supported platforms

| Platform | Architectures        | Release archive       |
| -------- | -------------------- | --------------------- |
| macOS    | Apple Silicon, Intel | `distship_*_Darwin_*` |
| Linux    | ARM64, x86-64        | `distship_*_Linux_*`  |

Windows is not currently supported.

## Install with Homebrew

```bash
brew install --cask sperains/tap/distship
```

Homebrew installs the matching macOS or Linux binary and verifies the archive
checksum declared by the Cask.

The current macOS binaries are checksum-verified but are not yet signed or
notarized with an Apple Developer ID. As a temporary compatibility measure, the
Cask removes the quarantine attribute from the installed `distship` binary
only. It does not change the system-wide Gatekeeper configuration.

## Install with Codex

Ask Codex to install the skill from this repository:

```text
Install the skill from https://github.com/sperains/distship/tree/main/skills/install-distship,
then use $install-distship to install the latest stable release.
```

The skill detects the operating system and architecture, downloads the matching
release, verifies its SHA-256 checksum, and installs it to `$HOME/.local/bin` by
default. It can also install a specific version or use another destination.

## Install manually

### 1. Download

Download the archive matching your platform and `checksums.txt` from the
[latest GitHub Release](https://github.com/sperains/distship/releases/latest).
Keep both files in the same directory.

### 2. Verify the archive

Run the command for your operating system from the download directory:

```bash
# macOS
shasum -a 256 --ignore-missing --check checksums.txt

# Linux
sha256sum --ignore-missing --check checksums.txt
```

Continue only when the downloaded archive reports `OK`.

### 3. Install the binary

Replace `<archive>` with the downloaded archive name:

```bash
tar -xzf <archive>
mkdir -p "$HOME/.local/bin"
install -m 0755 distship "$HOME/.local/bin/distship"
"$HOME/.local/bin/distship" version
```

If `distship` is not found by name, add this line to your shell profile and open
a new terminal:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

## Upgrade

If DistShip was installed with Homebrew, upgrade it with:

```bash
brew upgrade --cask distship
```

Otherwise, run the Codex skill again or download and verify the newer release
before replacing the existing binary. Upgrades do not change DistShip
configuration or local deployment history.

After upgrading, confirm the active version:

```bash
distship version
```

## Uninstall

For a Homebrew installation, run:

```bash
brew uninstall --cask distship
```

Run `command -v distship` to locate the active binary, then remove that exact
file. Configuration and deployment history are intentionally preserved.

Only remove these directories when their local records are no longer needed:

```text
${XDG_CONFIG_HOME:-$HOME/.config}/distship
${XDG_STATE_HOME:-$HOME/.local/state}/distship
```

For the next step, return to the [quick start](../README.md#quick-start).
