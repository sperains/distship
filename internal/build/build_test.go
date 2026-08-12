package build

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/sperains/distship/internal/config"
	"github.com/sperains/distship/internal/process"
)

type fakeRunner struct {
	command process.Command
	result  process.Result
	err     error
}

func (r *fakeRunner) LookPath(string) (string, error) { return "", nil }
func (r *fakeRunner) Run(_ context.Context, command process.Command) (process.Result, error) {
	r.command = command
	return r.result, r.err
}

func TestRunBuildsInProjectAndValidatesArtifact(t *testing.T) {
	directory := t.TempDir()
	if err := os.Mkdir(filepath.Join(directory, "dist"), 0o700); err != nil {
		t.Fatal(err)
	}
	runner := &fakeRunner{}
	result, err := Run(context.Background(), runner, config.Environment{Directory: directory, Build: []string{"pnpm", "build-test"}, Artifact: "dist"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	artifactInfo, statErr := os.Stat(result.Artifact)
	if runner.command.Name != "pnpm" || runner.command.Dir != directory || len(runner.command.Args) != 1 || runner.command.Args[0] != "build-test" || statErr != nil || !artifactInfo.IsDir() {
		t.Fatalf("command = %#v, result = %#v", runner.command, result)
	}
}

func TestRunStopsWhenBuildFailsOrArtifactIsMissing(t *testing.T) {
	directory := t.TempDir()
	failed := &fakeRunner{err: errors.New("exit 1"), result: process.Result{Stderr: "compile failed"}}
	if _, err := Run(context.Background(), failed, config.Environment{Directory: directory, Build: []string{"pnpm", "build"}, Artifact: "dist"}, nil); err == nil {
		t.Fatal("build failure error = nil")
	}
	missing := &fakeRunner{}
	if _, err := Run(context.Background(), missing, config.Environment{Directory: directory, Build: []string{"pnpm", "build"}, Artifact: "dist"}, nil); err == nil {
		t.Fatal("missing artifact error = nil")
	}
}

func TestRunRejectsArtifactSymlinkOutsideProject(t *testing.T) {
	directory := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(directory, "dist")); err != nil {
		t.Fatal(err)
	}
	if _, err := Run(context.Background(), &fakeRunner{}, config.Environment{Directory: directory, Build: []string{"pnpm", "build"}, Artifact: "dist"}, nil); err == nil {
		t.Fatal("outside artifact error = nil")
	}
}
