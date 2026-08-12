package preflight

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sperains/distship/internal/config"
	gitinfo "github.com/sperains/distship/internal/git"
	"github.com/sperains/distship/internal/process"
)

type ErrorKind string

const (
	ErrorLocalDirectory ErrorKind = "local_directory"
	ErrorBuildTool      ErrorKind = "build_tool"
	ErrorGit            ErrorKind = "git"
	ErrorBranch         ErrorKind = "branch"
	ErrorDirty          ErrorKind = "dirty"
	ErrorSSH            ErrorKind = "ssh"
	ErrorRemote         ErrorKind = "remote"
)

type CheckError struct {
	Kind   ErrorKind
	Detail string
	Cause  error
}

func (e *CheckError) Error() string { return e.Detail }
func (e *CheckError) Unwrap() error { return e.Cause }

type RemoteState string

const (
	RemoteExisting  RemoteState = "existing"
	RemoteCreatable RemoteState = "creatable"
)

type Report struct {
	GitAvailable bool
	Repository   gitinfo.Repository
	RemoteState  RemoteState
	RemoteParent string
	Warnings     []string
}

func Check(ctx context.Context, runner process.Runner, environment config.Environment) (Report, error) {
	var report Report
	if err := checkDirectory(environment.Directory); err != nil {
		return report, &CheckError{Kind: ErrorLocalDirectory, Detail: err.Error(), Cause: err}
	}
	if len(environment.Build) == 0 {
		return report, &CheckError{Kind: ErrorBuildTool, Detail: "build command is empty"}
	}
	if err := checkExecutable(runner, environment.Directory, environment.Build[0]); err != nil {
		return report, &CheckError{Kind: ErrorBuildTool, Detail: err.Error(), Cause: err}
	}

	if _, err := runner.LookPath("git"); err == nil {
		report.GitAvailable = true
		repository, inspectErr := gitinfo.Inspect(ctx, runner, environment.Directory)
		if inspectErr != nil {
			if ctx.Err() != nil {
				return report, &CheckError{Kind: ErrorGit, Detail: ctx.Err().Error(), Cause: ctx.Err()}
			}
			return report, &CheckError{Kind: ErrorGit, Detail: inspectErr.Error(), Cause: inspectErr}
		}
		report.Repository = repository
		if repository.IsRepository {
			branch := repository.Branch
			if repository.Detached {
				branch = "HEAD"
				report.Warnings = append(report.Warnings, "detached")
			}
			if len(environment.Git.AllowedBranches) > 0 && !contains(environment.Git.AllowedBranches, branch) {
				return report, &CheckError{Kind: ErrorBranch, Detail: fmt.Sprintf("%s (allowed: %s)", branch, strings.Join(environment.Git.AllowedBranches, ", "))}
			}
			if repository.Changes.Total() > 0 {
				switch environment.Git.Dirty {
				case "deny":
					return report, &CheckError{Kind: ErrorDirty, Detail: formatChanges(repository.Changes)}
				case "warn":
					report.Warnings = append(report.Warnings, "dirty")
				}
			}
		} else {
			report.Warnings = append(report.Warnings, "not_git")
		}
	} else {
		report.Warnings = append(report.Warnings, "git_unavailable")
	}

	if _, err := runner.LookPath("ssh"); err != nil {
		return report, &CheckError{Kind: ErrorSSH, Detail: err.Error(), Cause: err}
	}
	remote, err := checkRemote(ctx, runner, environment.Target)
	if err != nil {
		if ctx.Err() != nil {
			return Report{}, &CheckError{Kind: ErrorSSH, Detail: ctx.Err().Error(), Cause: ctx.Err()}
		}
		return report, err
	}
	report.RemoteState = remote.RemoteState
	report.RemoteParent = remote.RemoteParent
	return report, nil
}

func checkDirectory(directory string) error {
	info, err := os.Stat(directory)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", directory)
	}
	return nil
}

func checkExecutable(runner process.Runner, directory, executable string) error {
	if !strings.ContainsRune(executable, filepath.Separator) {
		_, err := runner.LookPath(executable)
		return err
	}
	path := executable
	if !filepath.IsAbs(path) {
		path = filepath.Join(directory, path)
	}
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() || info.Mode().Perm()&0o111 == 0 {
		return fmt.Errorf("not executable: %s", path)
	}
	return nil
}

func checkRemote(ctx context.Context, runner process.Runner, target config.Target) (Report, error) {
	command := "sh -s -- " + shellQuote(target.Directory)
	result, err := runner.Run(ctx, process.Command{
		Name:  "ssh",
		Args:  []string{"-o", "BatchMode=yes", "-o", "ConnectTimeout=10", target.Host, command},
		Stdin: remoteCheckScript,
	})
	if err != nil {
		detail := strings.TrimSpace(result.Stderr)
		if detail == "" {
			detail = err.Error()
		}
		return Report{}, &CheckError{Kind: ErrorSSH, Detail: detail, Cause: err}
	}
	fields := strings.Split(strings.TrimSpace(result.Stdout), "\t")
	switch fields[0] {
	case string(RemoteExisting):
		return Report{RemoteState: RemoteExisting}, nil
	case string(RemoteCreatable):
		parent := ""
		if len(fields) > 1 {
			parent = fields[1]
		}
		return Report{RemoteState: RemoteCreatable, RemoteParent: parent}, nil
	default:
		return Report{}, &CheckError{Kind: ErrorRemote, Detail: target.Directory}
	}
}

func contains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func formatChanges(changes gitinfo.Changes) string {
	return fmt.Sprintf("modified=%d added=%d deleted=%d untracked=%d", changes.Modified, changes.Added, changes.Deleted, changes.Untracked)
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}

const remoteCheckScript = `target=$1
if [ -d "$target" ]; then
  if [ -w "$target" ]; then
    printf 'existing\n'
    exit 0
  fi
  printf 'not-writable\n'
  exit 0
fi

parent=$target
while [ ! -e "$parent" ]; do
  next=$(dirname "$parent")
  if [ "$next" = "$parent" ]; then
    break
  fi
  parent=$next
done

if [ -d "$parent" ] && [ -w "$parent" ]; then
  printf 'creatable\t%s\n' "$parent"
  exit 0
fi

printf 'not-writable\n'
`
