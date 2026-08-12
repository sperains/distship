# DistShip

**Build locally. Review changes. Ship safely.**

[![CI](https://github.com/sperains/distship/actions/workflows/ci.yml/badge.svg)](https://github.com/sperains/distship/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/sperains/distship)](https://github.com/sperains/distship/releases/latest)
[![License](https://img.shields.io/github/license/sperains/distship)](LICENSE)

[English](README.md) · [简体中文](README.zh-CN.md)

DistShip is a local-first CLI for deploying static frontend artifacts to SSH
servers. It shows the Git changes you are about to ship, checks the target,
builds on your machine, and uploads only after confirmation.

![DistShip deployment preflight showing Git changes and target readiness](docs/assets/distship-preflight.png)

## Why DistShip?

- **Know what will ship.** Review the exact Git revision and recent non-merge
  commits before deployment.
- **Catch problems early.** Validate the project, build tool, branch policy, SSH
  access, and remote permissions without changing anything.
- **Keep builds local.** Use each frontend project's own build command and
  upload the resulting static artifacts over SSH.
- **Stay in control.** Preview the plan, confirm the target, and keep a local
  record of successful deployments.
- **Reuse standard SSH.** Keep keys, ports, jump hosts, and aliases in
  `~/.ssh/config` instead of application configuration.

## Quick start

### 1. Install

#### With Codex

Ask Codex to install the bundled [`install-distship`](skills/install-distship/SKILL.md)
skill, then use it to install the latest verified release:

```text
Install the skill from https://github.com/sperains/distship/tree/main/skills/install-distship,
then use $install-distship to install the latest stable release.
```

#### Without Codex

Download the archive for your platform and `checksums.txt` from the
[latest release](https://github.com/sperains/distship/releases/latest), then
follow the [installation guide](docs/INSTALLATION.md) to verify and install it.

### 2. Configure SSH

Verify that the deployment target works with your system SSH client:

```bash
ssh staging-web
```

An SSH alias is recommended but not required. DistShip also accepts a hostname,
IP address, or `user@host`. See the
[SSH configuration guide](docs/SSH_CONFIGURATION.md) for keys, custom ports,
jump hosts, fingerprint verification, and troubleshooting.

### 3. Initialize a project

Run inside a frontend project, or pass its path explicitly:

```bash
cd /path/to/frontend-project
distship init

# Equivalent explicit form:
distship init /path/to/frontend-project
```

DistShip detects supported frontend project metadata, proposes reliable
defaults, previews the configuration, and validates it after saving. Use
`--advanced` to customize the target ID, names, artifact paths, allowed
branches, and dirty-working-tree policy.

<!-- markdownlint-disable MD013 MD033 -->
<details>
<summary>View the complete initialization flow</summary>

```text
$ distship init
DistShip initialization

The current directory is used when it looks like a deployable project; otherwise enter a project directory. Use --advanced to configure every field.

✓ Detected current project directory
  /Users/example/projects/web_test

Project analysis

  Local directory: /Users/example/projects/web_test
  Project: web_test
  Project type: Node.js
  Package manager: npm
  Git branch: not detected
  Build command: npm run build
  Artifact: not detected

Deployment target

  Project: web-test
  Deployment environment [test]:

  Deployment target ID: web-test:test
  The ID appears in list and is used by check, deploy, and remove commands.

Build command [npm run build]:
Artifact directory (for example, dist): dist
SSH server (for example, staging-web or deploy@example.com): staging-web
Remote deployment directory (absolute path): /var/www/web-test

Configuration preview

  Deployment target ID: web-test:test
  Project: web_test
  Environment: test
  Local directory: /Users/example/projects/web_test
  Build command: npm run build
  Artifact: dist
  Target: staging-web:/var/www/web-test
  Allowed branches: any branch (warning shown before deployment)
  Working tree: warn
  Configuration: /Users/example/.config/distship/projects.toml

Only the local configuration will change. No server connection will be made.
Save this deployment target? [N]: y

✓ Configuration written
✓ Configuration is valid

  Path: /Users/example/.config/distship/projects.toml
  Deployment target ID: web-test:test

Next step: distship check web-test:test
```

</details>
<!-- markdownlint-enable MD013 MD033 -->

### 4. List, check, and deploy

```bash
distship list
```

```text
[1] storefront · test
    ID: storefront:test
    Local: /Users/example/projects/storefront
    Remote: staging-web:/var/www/storefront
    Branches: test

[2] operations · staging
    ID: operations:staging
    Local: /Users/example/projects/operations
    Remote: staging-web:/var/www/operations
    Branches: main

[3] docs-site · test
    ID: docs-site:test
    Local: /Users/example/projects/docs-site
    Remote: staging-web:/var/www/docs-site
    Branches: any branch (warning shown before deployment)
```

Copy a target ID from the list and use it for the remaining commands:

```bash
distship check storefront:test
distship deploy storefront:test --dry-run
distship deploy storefront:test
```

Target IDs follow the `project:environment` format.

## Deployment flow

```text
init configuration
       ↓
read-only preflight
       ↓
review target + Git changes + build plan
       ↓
confirm → local build → artifact validation → rsync upload
       ↓
record successful deployment locally
```

`distship check` is read-only: it does not build, create remote directories,
upload files, or write deployment history. `distship deploy --dry-run` previews
the local plan without building or connecting to the server.

During a real deployment, DistShip:

1. Runs the same preflight checks.
2. Shows the exact source revision and recent non-merge commits.
3. Builds locally and validates the configured artifact directory.
4. Creates the remote directory only when needed.
5. Uploads incrementally with `rsync` without deleting unrelated remote files.
6. Records the successful deployment under the XDG state directory.

Use `--yes` only when skipping the ordinary confirmation is intentional. Safety
failures remain blocking.

## Requirements and scope

| Area                | Current support                |
| ------------------- | ------------------------------ |
| Operating systems   | macOS and Linux                |
| Projects            | Static frontend artifacts      |
| Environments        | Test and staging               |
| Transport           | System SSH and `rsync`         |
| Interface languages | English and Simplified Chinese |

Git, SSH, `rsync`, and the configured project build tool must be available
locally. The remote SSH user must be able to write to the target directory.

Version 0.1 does **not** manage backend services, containers, databases, process
managers, server-side builds, service restarts, or runtime configuration.
Uploads are incremental rather than atomic, and automatic rollback is not
provided.

## Common commands

```bash
distship init [project-directory]
distship list
distship check <project:environment>
distship deploy <project:environment>
distship config validate
distship config remove <project:environment>
distship version
```

Add `--advanced` to `init`; add `--dry-run` or `--yes` to `deploy`. Use
`--config <path>` to select a configuration file and `--lang en` or
`--lang zh-CN` to override terminal language detection.

## Documentation

- [Installation and upgrades](docs/INSTALLATION.md)
- [SSH configuration](docs/SSH_CONFIGURATION.md)
- [Configuration example](examples/projects.toml)
- [Release process](docs/RELEASING.md)
- [Contributing](CONTRIBUTING.md)
- [Security policy](SECURITY.md)
- [Changelog](CHANGELOG.md)

## Build from source

```bash
go build ./...
go test ./...
```

## License

DistShip is available under the [MIT License](LICENSE).
