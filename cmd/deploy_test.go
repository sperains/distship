package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sperains/distship/internal/config"
	"github.com/sperains/distship/internal/history"
	"github.com/sperains/distship/internal/process"
)

func TestDeployBuildsUploadsAndWritesHistoryAfterConfirmation(t *testing.T) {
	t.Setenv("DISTSHIP_LANG", "en")
	stateHome := t.TempDir()
	t.Setenv("XDG_STATE_HOME", stateHome)
	directory := t.TempDir()
	if err := os.Mkdir(filepath.Join(directory, "dist"), 0o700); err != nil {
		t.Fatal(err)
	}
	path := deployTestConfig(t, directory)
	runner := successfulDeployRunner("existing\n")

	var output bytes.Buffer
	root, application := buildRootCommand(strings.NewReader(""), &output, &output, "en")
	application.runner = runner
	root.SetArgs([]string{"--config", path, "deploy", "site:staging", "--yes"})
	if err := root.Execute(); err != nil {
		t.Fatalf("deploy error = %v\n%s", err, output.String())
	}
	for _, expected := range []string{"DistShip deployment · site:staging", "READY TO DEPLOY", "Recent commits · 1 shown", "Building locally", "Uploading artifacts", "Deployment completed", "Local deployment history updated"} {
		if !strings.Contains(output.String(), expected) {
			t.Errorf("output missing %q:\n%s", expected, output.String())
		}
	}
	if len(runner.commands) != 8 || runner.commands[6].Name != "pnpm" || runner.commands[7].Name != "rsync" {
		t.Fatalf("commands = %#v", runner.commands)
	}
	if args := strings.Join(runner.commands[7].Args, " "); strings.Contains(args, "--delete") {
		t.Fatalf("rsync unexpectedly deletes remote files: %s", args)
	}
	historyPath := filepath.Join(stateHome, "distship", "history.jsonl")
	record, found, err := history.Last(historyPath, config.TargetRef{ProjectID: "site", EnvironmentID: "staging"}, config.Target{Host: "web", Directory: "/var/www/site"})
	if err != nil || !found || record.Commit != "mergehash" || record.Branch != "test" || record.Artifact != "dist" {
		t.Fatalf("history = %#v, found = %v, error = %v", record, found, err)
	}
}

func TestDeployDryRunDoesNotConnectBuildUploadOrWriteHistory(t *testing.T) {
	t.Setenv("DISTSHIP_LANG", "en")
	stateHome := t.TempDir()
	t.Setenv("XDG_STATE_HOME", stateHome)
	directory := t.TempDir()
	path := deployTestConfig(t, directory)
	runner := successfulDeployRunner("existing\n")

	var output bytes.Buffer
	root, application := buildRootCommand(strings.NewReader(""), &output, &output, "en")
	application.runner = runner
	root.SetArgs([]string{"--config", path, "deploy", "site:staging", "--dry-run"})
	if err := root.Execute(); err != nil {
		t.Fatalf("dry run error = %v\n%s", err, output.String())
	}
	if len(runner.commands) != 5 {
		t.Fatalf("dry run commands = %#v", runner.commands)
	}
	for _, command := range runner.commands {
		if command.Name == "ssh" || command.Name == "pnpm" || command.Name == "rsync" {
			t.Fatalf("dry run executed %s", command.Name)
		}
	}
	if !strings.Contains(output.String(), "DRY RUN · REMOTE NOT CHECKED") || strings.Count(output.String(), "Dry run complete") != 1 {
		t.Fatalf("dry run output = %s", output.String())
	}
	if _, err := os.Stat(filepath.Join(stateHome, "distship", "history.jsonl")); !os.IsNotExist(err) {
		t.Fatalf("dry run wrote history: %v", err)
	}
}

func TestDeployCancellationStopsBeforeBuild(t *testing.T) {
	t.Setenv("DISTSHIP_LANG", "en")
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	directory := t.TempDir()
	path := deployTestConfig(t, directory)
	runner := successfulDeployRunner("existing\n")

	var output bytes.Buffer
	root, application := buildRootCommand(strings.NewReader("n\n"), &output, &output, "en")
	application.runner = runner
	root.SetArgs([]string{"--config", path, "deploy", "site:staging"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if len(runner.commands) != 6 || !strings.Contains(output.String(), "Cancelled. Nothing was built or uploaded.") {
		t.Fatalf("commands = %#v\n%s", runner.commands, output.String())
	}
}

func TestDeployBuildFailureDoesNotUploadOrWriteHistory(t *testing.T) {
	t.Setenv("DISTSHIP_LANG", "en")
	stateHome := t.TempDir()
	t.Setenv("XDG_STATE_HOME", stateHome)
	directory := t.TempDir()
	path := deployTestConfig(t, directory)
	runner := successfulDeployRunner("existing\n")
	runner.failAt = 6
	runner.failure = errors.New("exit 1")
	runner.failResult = process.Result{Stderr: "compile failed"}

	var output bytes.Buffer
	root, application := buildRootCommand(strings.NewReader(""), &output, &output, "en")
	application.runner = runner
	root.SetArgs([]string{"--config", path, "deploy", "site:staging", "--yes"})
	err := root.Execute()
	if err == nil || !strings.Contains(err.Error(), "compile failed") {
		t.Fatalf("error = %v", err)
	}
	if len(runner.commands) != 7 {
		t.Fatalf("commands after build failure = %#v", runner.commands)
	}
	if _, err := os.Stat(filepath.Join(stateHome, "distship", "history.jsonl")); !os.IsNotExist(err) {
		t.Fatalf("failed deploy wrote history: %v", err)
	}
}

func TestDeployShowsChangesSinceMatchingLocalHistory(t *testing.T) {
	t.Setenv("DISTSHIP_LANG", "en")
	stateHome := t.TempDir()
	t.Setenv("XDG_STATE_HOME", stateHome)
	directory := t.TempDir()
	path := deployTestConfig(t, directory)
	historyPath := filepath.Join(stateHome, "distship", "history.jsonl")
	if err := history.Append(historyPath, history.Record{
		Project:     "site",
		Environment: "staging",
		Host:        "web",
		Directory:   "/var/www/site",
		Commit:      "previous",
	}); err != nil {
		t.Fatal(err)
	}
	runner := &checkRunner{responses: []process.Result{
		{Stdout: "true\n"},
		{Stdout: "test\n"},
		{Stdout: "mergehash\x00merge123\n"},
		{Stdout: "1234567890\x1f12345678\x1ffix: recent\x1fSperains\x1f2026-08-12T10:30:00+08:00\x1e\n"},
		{},
		{Stdout: "existing\n"},
		{},
		{},
		{Stdout: "7\n"},
		{Stdout: "abcdef123\x1fabcdef12\x1ffeat: deployed change\x1fAlice\x1f2026-08-12T09:30:00+08:00\x1e\n"},
	}}

	var output bytes.Buffer
	root, application := buildRootCommand(strings.NewReader("n\n"), &output, &output, "en")
	application.runner = runner
	root.SetArgs([]string{"--config", path, "deploy", "site:staging"})
	if err := root.Execute(); err != nil {
		t.Fatalf("deploy preview error = %v\n%s", err, output.String())
	}
	for _, expected := range []string{"Changes since local last deployment · 7 commits", "feat: deployed change", "abcdef12 · Alice"} {
		if !strings.Contains(output.String(), expected) {
			t.Errorf("output missing %q:\n%s", expected, output.String())
		}
	}
	if strings.Contains(output.String(), "fix: recent") {
		t.Fatalf("output used recent commits instead of deployment range:\n%s", output.String())
	}
}

func TestDeployCreatesRemoteDirectoryOnlyAfterSuccessfulBuild(t *testing.T) {
	t.Setenv("DISTSHIP_LANG", "en")
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	directory := t.TempDir()
	if err := os.Mkdir(filepath.Join(directory, "dist"), 0o700); err != nil {
		t.Fatal(err)
	}
	path := deployTestConfig(t, directory)
	runner := successfulDeployRunner("creatable\t/var/www\n")
	runner.responses = append(runner.responses, process.Result{})

	var output bytes.Buffer
	root, application := buildRootCommand(strings.NewReader(""), &output, &output, "en")
	application.runner = runner
	root.SetArgs([]string{"--config", path, "deploy", "site:staging", "--yes"})
	if err := root.Execute(); err != nil {
		t.Fatalf("deploy error = %v\n%s", err, output.String())
	}
	if len(runner.commands) != 9 || runner.commands[6].Name != "pnpm" || runner.commands[7].Name != "ssh" || runner.commands[8].Name != "rsync" {
		t.Fatalf("commands = %#v", runner.commands)
	}
	if !strings.Contains(runner.commands[7].Stdin, "mkdir -p") {
		t.Fatalf("remote creation command = %#v", runner.commands[7])
	}
}

func TestDeployUploadFailureDoesNotWriteHistory(t *testing.T) {
	t.Setenv("DISTSHIP_LANG", "en")
	stateHome := t.TempDir()
	t.Setenv("XDG_STATE_HOME", stateHome)
	directory := t.TempDir()
	if err := os.Mkdir(filepath.Join(directory, "dist"), 0o700); err != nil {
		t.Fatal(err)
	}
	path := deployTestConfig(t, directory)
	runner := successfulDeployRunner("existing\n")
	runner.failAt = 7
	runner.failure = errors.New("exit 23")
	runner.failResult = process.Result{Stderr: "transfer failed"}

	var output bytes.Buffer
	root, application := buildRootCommand(strings.NewReader(""), &output, &output, "en")
	application.runner = runner
	root.SetArgs([]string{"--config", path, "deploy", "site:staging", "--yes"})
	err := root.Execute()
	if err == nil || !strings.Contains(err.Error(), "transfer failed") {
		t.Fatalf("error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(stateHome, "distship", "history.jsonl")); !os.IsNotExist(err) {
		t.Fatalf("failed upload wrote history: %v", err)
	}
}

func deployTestConfig(t *testing.T, directory string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "projects.toml")
	cfg := testConfig()
	project := cfg.Projects["site"]
	environment := project.Environments["staging"]
	environment.Directory = directory
	environment.Build = []string{"pnpm", "build-test"}
	environment.Artifact = "dist"
	environment.Git.AllowedBranches = []string{"test"}
	project.Environments["staging"] = environment
	cfg.Projects["site"] = project
	if err := config.Save(path, cfg); err != nil {
		t.Fatal(err)
	}
	return path
}

func successfulDeployRunner(remote string) *checkRunner {
	return &checkRunner{responses: []process.Result{
		{Stdout: "true\n"},
		{Stdout: "test\n"},
		{Stdout: "mergehash\x00merge123\n"},
		{Stdout: "1234567890\x1f12345678\x1ffix: safe deploy\x1fSperains\x1f2026-08-12T10:30:00+08:00\x1e\n"},
		{},
		{Stdout: remote},
		{},
		{},
	}}
}
