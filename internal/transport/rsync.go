package transport

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/sperains/distship/internal/config"
	"github.com/sperains/distship/internal/process"
)

func CheckAvailable(runner process.Runner) error {
	if _, err := runner.LookPath("rsync"); err != nil {
		return fmt.Errorf("rsync is unavailable: %w", err)
	}
	return nil
}

func EnsureRemote(ctx context.Context, runner process.Runner, target config.Target) error {
	command := "sh -s -- " + shellQuote(target.Directory)
	result, err := runner.Run(ctx, process.Command{
		Name:  "ssh",
		Args:  []string{"-o", "BatchMode=yes", "-o", "ConnectTimeout=10", target.Host, command},
		Stdin: ensureRemoteScript,
	})
	if err != nil {
		return commandError("create remote directory", result, err)
	}
	return nil
}

func Upload(ctx context.Context, runner process.Runner, artifact string, target config.Target, output io.Writer) error {
	source := filepath.Clean(artifact) + string(filepath.Separator)
	destination := target.Host + ":" + strings.TrimSuffix(target.Directory, "/") + "/"
	result, err := runner.Run(ctx, process.Command{
		Name:   "rsync",
		Args:   []string{"-az", "--itemize-changes", "--human-readable", "--", source, destination},
		Stdout: output,
		Stderr: output,
	})
	if err != nil {
		return commandError("upload artifacts", result, err)
	}
	return nil
}

func commandError(action string, result process.Result, err error) error {
	detail := strings.TrimSpace(result.Stderr)
	if detail == "" {
		detail = err.Error()
	}
	return fmt.Errorf("%s: %s", action, detail)
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}

const ensureRemoteScript = `target=$1
mkdir -p -- "$target"
test -d "$target"
test -w "$target"
`
