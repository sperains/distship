# DistShip

Build locally. Review changes. Ship safely.

DistShip is a local-first CLI that previews Git changes, builds web projects, and deploys artifacts to SSH servers.

> Project status: early development. Configuration management and read-only deployment preflight checks are available. Build and upload are not implemented yet.

## Requirements

- macOS or Linux
- Git, SSH, and rsync for deployment features
- The build tool required by each configured project

## Build from source

```bash
go build ./...
go test ./...
```

## First use

```bash
distship init /absolute/path/to/project
# Or run inside a recognized project:
cd /absolute/path/to/project
distship init
distship list
distship check project:test
```

The project directory can be passed as an argument, detected from the current directory, or entered interactively. DistShip automatically uses the current directory only when it detects a buildable project or finds that directory in existing configuration, and always reports the decision. Use `distship init .` to select the current directory explicitly. The directory is checked before configuration questions begin.

DistShip then reports the detected project type, package manager, Git branch, build command, and artifact directory. Only reliable values become defaults. If the build command or artifact directory cannot be inferred, they remain required questions.

For an existing target, its display names and Git policies are preserved by default. DistShip shows field-level changes before asking for one final confirmation; when nothing changed, it exits without rewriting the file. Every saved configuration is reloaded and validated automatically.

When the selected directory already belongs to one configured target, DistShip reuses that target's project and environment IDs instead of deriving a new ID from the package name. If several targets share the directory, the current Git branch and then `test` are used to resolve an unambiguous default.

Use `distship init /path/to/project --advanced` to customize display names, artifact paths, allowed branches, and the dirty-working-tree policy.

The deployment target accepts an SSH alias, hostname, IP address, or `user@host`. Enter it together with the remote directory (`bt_250:/www/site`) or enter the SSH target first and the remote directory at the next prompt. An SSH alias is recommended but not required. Configure custom ports and other connection options in `~/.ssh/config`; this keeps future SSH and rsync behavior consistent.

## Check a target

```bash
distship check ipd:test
```

`check` validates the local project directory, build tool, Git branch and working-tree policy, SSH connection, and remote directory permissions. It is read-only: it does not build, create remote directories, upload files, or write deployment history. For Git repositories, it shows the exact current revision for traceability and up to three recent non-merge commits for deployment decisions.

## Remove a target

```bash
distship config remove ipd:test
```

Use the target ID shown by `distship list`. The command previews the selected target and asks for confirmation. It changes only the local configuration and never removes project or server files. Use `--yes` to skip confirmation. Empty projects are removed automatically; when the last target is removed, the configuration file is renamed to a timestamped backup instead of being deleted.

Use `--config <path>` to select an explicit configuration file. Otherwise DistShip follows its documented configuration lookup order and writes new configurations to the XDG configuration directory.

See [examples/projects.toml](examples/projects.toml) for the version 1 configuration format.

## Language

DistShip detects the terminal language from `LC_ALL`, `LC_MESSAGES`, and `LANG`, with English as the fallback. English and Simplified Chinese are bundled in the binary.

```bash
distship --lang en list
distship --lang zh-CN list
DISTSHIP_LANG=en distship list
```

The `--lang` option takes precedence over `DISTSHIP_LANG`, which takes precedence over the system locale. Commands, flags, and configuration keys always remain in English.

## Scope

Version 0.1 targets test and staging environments. Direct incremental upload is not atomic and does not provide automatic rollback.

See [README.zh-CN.md](README.zh-CN.md) for Chinese documentation.

## License

MIT
