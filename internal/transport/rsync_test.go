package transport

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/sperains/distship/internal/config"
	"github.com/sperains/distship/internal/process"
)

type fakeRunner struct {
	commands []process.Command
	result   process.Result
	err      error
}

func (r *fakeRunner) LookPath(name string) (string, error) {
	if name == "rsync" {
		return "/usr/bin/rsync", nil
	}
	return "", errors.New("not found")
}

func (r *fakeRunner) Run(_ context.Context, command process.Command) (process.Result, error) {
	r.commands = append(r.commands, command)
	return r.result, r.err
}

func TestEnsureRemoteUsesQuotedArgumentAndScript(t *testing.T) {
	runner := &fakeRunner{}
	target := config.Target{Host: "server", Directory: "/www/site's files"}
	if err := EnsureRemote(context.Background(), runner, target); err != nil {
		t.Fatal(err)
	}
	command := runner.commands[0]
	if command.Name != "ssh" || !strings.Contains(strings.Join(command.Args, " "), `'/www/site'\''s files'`) || command.Stdin != ensureRemoteScript {
		t.Fatalf("command = %#v", command)
	}
}

func TestUploadUsesIncrementalRsyncWithoutDelete(t *testing.T) {
	runner := &fakeRunner{}
	if err := Upload(context.Background(), runner, "/project/dist", config.Target{Host: "server", Directory: "/www/site"}, nil); err != nil {
		t.Fatal(err)
	}
	command := runner.commands[0]
	joined := strings.Join(command.Args, " ")
	if command.Name != "rsync" || !strings.Contains(joined, "/project/dist/ server:/www/site/") || strings.Contains(joined, "--delete") {
		t.Fatalf("command = %#v", command)
	}
}
