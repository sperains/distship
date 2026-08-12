package git

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/sperains/distship/internal/process"
)

type rangeResponse struct {
	result process.Result
	err    error
}

type rangeRunner struct {
	responses []rangeResponse
	commands  []process.Command
}

func (r *rangeRunner) LookPath(string) (string, error) { return "", nil }

func (r *rangeRunner) Run(_ context.Context, command process.Command) (process.Result, error) {
	r.commands = append(r.commands, command)
	response := r.responses[0]
	r.responses = r.responses[1:]
	return response.result, response.err
}

func TestParseStatus(t *testing.T) {
	changes := parseStatus(" M changed.go\x00A  added.go\x00 D deleted.go\x00?? new.go\x00R  renamed.go\x00old.go\x00")
	if changes.Modified != 2 || changes.Added != 1 || changes.Deleted != 1 || changes.Untracked != 1 || changes.Total() != 5 {
		t.Fatalf("changes = %#v", changes)
	}
}

func TestParseCommits(t *testing.T) {
	output := "aaa\x1fa1\x1ffeat: first\x1fAlice\x1f2026-08-12T10:30:00+08:00\x1e\n" +
		"bbb\x1fb2\x1ffix: second\x1fBob\x1f2026-08-11T09:20:00+08:00\x1e\n"
	commits, err := parseCommits(output)
	if err != nil {
		t.Fatal(err)
	}
	if len(commits) != 2 || commits[0].Subject != "feat: first" || commits[1].Author != "Bob" {
		t.Fatalf("commits = %#v", commits)
	}
	if commits[0].Time.Format("2006-01-02 15:04") != "2026-08-12 10:30" {
		t.Fatalf("commit time = %s", commits[0].Time)
	}
}

func TestParseCommitsAllowsNoNonMergeHistory(t *testing.T) {
	commits, err := parseCommits("")
	if err != nil || len(commits) != 0 {
		t.Fatalf("commits = %#v, error = %v", commits, err)
	}
}

func TestChangesSinceReturnsNonMergeRange(t *testing.T) {
	runner := &rangeRunner{responses: []rangeResponse{
		{},
		{},
		{result: process.Result{Stdout: "7\n"}},
		{result: process.Result{Stdout: "aaa\x1fa1\x1ffix: safe deploy\x1fAlice\x1f2026-08-12T10:30:00+08:00\x1e\n"}},
	}}
	result, err := ChangesSince(context.Background(), runner, "/project", "base", "head", 5)
	if err != nil || result.Relation != RangeAvailable || result.Total != 7 || len(result.Commits) != 1 {
		t.Fatalf("result = %#v, error = %v", result, err)
	}
	if command := strings.Join(runner.commands[3].Args, " "); !strings.Contains(command, "--no-merges -5") || !strings.Contains(command, "base..head") {
		t.Fatalf("log command = %s", command)
	}
}

func TestChangesSinceReportsMissingAndDivergedHistory(t *testing.T) {
	missingRunner := &rangeRunner{responses: []rangeResponse{{result: process.Result{Stderr: "fatal: Not a valid object name"}, err: errors.New("missing")}}}
	missing, err := ChangesSince(context.Background(), missingRunner, "/project", "base", "head", 5)
	if err != nil || missing.Relation != RangeMissing {
		t.Fatalf("missing = %#v, error = %v", missing, err)
	}

	divergedRunner := &rangeRunner{responses: []rangeResponse{{}, {err: errors.New("not ancestor")}}}
	diverged, err := ChangesSince(context.Background(), divergedRunner, "/project", "base", "head", 5)
	if err != nil || diverged.Relation != RangeDiverged {
		t.Fatalf("diverged = %#v, error = %v", diverged, err)
	}
}
