package preflight

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/sperains/distship/internal/config"
	"github.com/sperains/distship/internal/process"
)

type response struct {
	result process.Result
	err    error
}

type fakeRunner struct {
	paths     map[string]bool
	responses []response
	commands  []process.Command
}

func (f *fakeRunner) LookPath(name string) (string, error) {
	if f.paths[name] {
		return "/usr/bin/" + name, nil
	}
	return "", errors.New("not found: " + name)
}

func (f *fakeRunner) Run(_ context.Context, command process.Command) (process.Result, error) {
	f.commands = append(f.commands, command)
	if len(f.responses) == 0 {
		return process.Result{}, errors.New("unexpected command")
	}
	result := f.responses[0]
	f.responses = f.responses[1:]
	return result.result, result.err
}

func TestCheckReturnsGitAndRemoteReport(t *testing.T) {
	runner := gitRunner("test", " M changed.go\x00", "existing\n")
	environment := checkedEnvironment(t.TempDir())
	report, err := Check(context.Background(), runner, environment)
	if err != nil {
		t.Fatal(err)
	}
	if !report.Repository.IsRepository || len(report.Repository.Recent) != 1 || report.Repository.Recent[0].Subject != "fix: safe deploy" || report.Repository.Changes.Total() != 1 {
		t.Fatalf("repository = %#v", report.Repository)
	}
	if report.RemoteState != RemoteExisting || !report.HasWarning(WarningDirty) {
		t.Fatalf("report = %#v", report)
	}
	recentCommand := strings.Join(runner.commands[3].Args, " ")
	if !strings.Contains(recentCommand, "--no-merges -3") {
		t.Fatalf("recent Git command does not exclude merges: %s", recentCommand)
	}
	sshCommand := runner.commands[len(runner.commands)-1]
	if sshCommand.Name != "ssh" || !strings.Contains(strings.Join(sshCommand.Args, " "), "bt_250") || sshCommand.Stdin != remoteCheckScript {
		t.Fatalf("ssh command = %#v", sshCommand)
	}
}

func TestCheckRejectsDisallowedBranch(t *testing.T) {
	runner := gitRunner("main", "", "existing\n")
	environment := checkedEnvironment(t.TempDir())
	_, err := Check(context.Background(), runner, environment)
	var checkError *CheckError
	if !errors.As(err, &checkError) || checkError.Kind != ErrorBranch {
		t.Fatalf("error = %v", err)
	}
	if len(runner.commands) != 5 {
		t.Fatalf("commands = %d, SSH should not run after branch rejection", len(runner.commands))
	}
}

func TestCheckRejectsDirtyWorkingTreeWhenDenied(t *testing.T) {
	runner := gitRunner("test", "?? new.go\x00", "existing\n")
	environment := checkedEnvironment(t.TempDir())
	environment.Git.Dirty = "deny"
	_, err := Check(context.Background(), runner, environment)
	var checkError *CheckError
	if !errors.As(err, &checkError) || checkError.Kind != ErrorDirty {
		t.Fatalf("error = %v", err)
	}
}

func TestCheckReportsCreatableRemoteDirectory(t *testing.T) {
	runner := gitRunner("test", "", "creatable\t/www/wwwgit\n")
	report, err := Check(context.Background(), runner, checkedEnvironment(t.TempDir()))
	if err != nil {
		t.Fatal(err)
	}
	if report.RemoteState != RemoteCreatable || report.RemoteParent != "/www/wwwgit" {
		t.Fatalf("report = %#v", report)
	}
}

func TestShellQuote(t *testing.T) {
	if got := shellQuote("/www/site's files"); got != `'/www/site'\''s files'` {
		t.Fatalf("shellQuote() = %q", got)
	}
}

func gitRunner(branch, status, remote string) *fakeRunner {
	return &fakeRunner{
		paths: map[string]bool{"pnpm": true, "git": true, "ssh": true},
		responses: []response{
			{result: process.Result{Stdout: "true\n"}},
			{result: process.Result{Stdout: branch + "\n"}},
			{result: process.Result{Stdout: "mergehash\x00merge123\n"}},
			{result: process.Result{Stdout: "1234567890\x1f12345678\x1ffix: safe deploy\x1fSperains\x1f2026-08-12T10:30:00+08:00\x1e\n"}},
			{result: process.Result{Stdout: status}},
			{result: process.Result{Stdout: remote}},
		},
	}
}

func checkedEnvironment(directory string) config.Environment {
	return config.Environment{
		Directory: directory,
		Build:     []string{"pnpm", "build-test"},
		Target:    config.Target{Host: "bt_250", Directory: "/www/wwwgit/site"},
		Git:       config.GitPolicy{AllowedBranches: []string{"test"}, Dirty: "warn"},
	}
}
