package process

import (
	"bytes"
	"context"
	"io"
	"os/exec"
	"strings"
)

type Command struct {
	Name   string
	Args   []string
	Dir    string
	Stdin  string
	Stdout io.Writer
	Stderr io.Writer
}

type Result struct {
	Stdout string
	Stderr string
}

type Runner interface {
	LookPath(name string) (string, error)
	Run(ctx context.Context, command Command) (Result, error)
}

type OSRunner struct{}

func (OSRunner) LookPath(name string) (string, error) {
	return exec.LookPath(name)
}

func (OSRunner) Run(ctx context.Context, command Command) (Result, error) {
	process := exec.CommandContext(ctx, command.Name, command.Args...)
	process.Dir = command.Dir
	if command.Stdin != "" {
		process.Stdin = strings.NewReader(command.Stdin)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	process.Stdout = writer(&stdout, command.Stdout)
	process.Stderr = writer(&stderr, command.Stderr)
	err := process.Run()
	return Result{Stdout: stdout.String(), Stderr: stderr.String()}, err
}

func writer(capture *bytes.Buffer, output io.Writer) io.Writer {
	if output == nil {
		return capture
	}
	return io.MultiWriter(capture, output)
}
