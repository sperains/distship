package build

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sperains/distship/internal/config"
	"github.com/sperains/distship/internal/process"
)

type Result struct {
	Artifact string
	Duration time.Duration
}

func Run(ctx context.Context, runner process.Runner, environment config.Environment, output io.Writer) (Result, error) {
	started := time.Now()
	command := process.Command{
		Name:   environment.Build[0],
		Args:   append([]string(nil), environment.Build[1:]...),
		Dir:    environment.Directory,
		Stdout: output,
		Stderr: output,
	}
	result, err := runner.Run(ctx, command)
	if err != nil {
		detail := strings.TrimSpace(result.Stderr)
		if detail == "" {
			detail = err.Error()
		}
		return Result{}, fmt.Errorf("build command failed: %s", detail)
	}
	artifact := filepath.Join(environment.Directory, environment.Artifact)
	info, err := os.Stat(artifact)
	if err != nil {
		return Result{}, fmt.Errorf("build artifact is unavailable: %w", err)
	}
	if !info.IsDir() {
		return Result{}, fmt.Errorf("build artifact is not a directory: %s", artifact)
	}
	projectRoot, err := filepath.EvalSymlinks(environment.Directory)
	if err != nil {
		return Result{}, fmt.Errorf("resolve project directory: %w", err)
	}
	resolvedArtifact, err := filepath.EvalSymlinks(artifact)
	if err != nil {
		return Result{}, fmt.Errorf("resolve build artifact: %w", err)
	}
	relative, err := filepath.Rel(projectRoot, resolvedArtifact)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return Result{}, fmt.Errorf("build artifact resolves outside the project: %s", artifact)
	}
	return Result{Artifact: resolvedArtifact, Duration: time.Since(started)}, nil
}
