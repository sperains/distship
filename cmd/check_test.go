package cmd

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sperains/distship/internal/config"
	gitinfo "github.com/sperains/distship/internal/git"
	"github.com/sperains/distship/internal/i18n"
	"github.com/sperains/distship/internal/preflight"
	"github.com/sperains/distship/internal/process"
)

type checkRunner struct {
	responses []process.Result
	commands  []process.Command
}

func (r *checkRunner) LookPath(name string) (string, error) {
	if name == "pnpm" || name == "git" || name == "ssh" {
		return "/usr/bin/" + name, nil
	}
	return "", errors.New("not found")
}

func (r *checkRunner) Run(_ context.Context, command process.Command) (process.Result, error) {
	r.commands = append(r.commands, command)
	if len(r.responses) == 0 {
		return process.Result{}, errors.New("unexpected command")
	}
	result := r.responses[0]
	r.responses = r.responses[1:]
	return result, nil
}

func TestCheckCommandRendersReadOnlyGitSummary(t *testing.T) {
	t.Setenv("DISTSHIP_LANG", "en")
	directory := t.TempDir()
	path := filepath.Join(t.TempDir(), "projects.toml")
	cfg := testConfig()
	project := cfg.Projects["site"]
	environment := project.Environments["staging"]
	environment.Directory = directory
	environment.Build = []string{"pnpm", "build-test"}
	environment.Git.AllowedBranches = []string{"test"}
	project.Environments["staging"] = environment
	cfg.Projects["site"] = project
	if err := config.Save(path, cfg); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	root, application := buildRootCommand(strings.NewReader(""), &output, &output, "en")
	runner := &checkRunner{responses: []process.Result{
		{Stdout: "true\n"},
		{Stdout: "test\n"},
		{Stdout: "mergehash\x00merge123\n"},
		{Stdout: "1234567890\x1f12345678\x1ffix: safe deploy\x1fSperains\x1f2026-08-12T10:30:00+08:00\x1e\n"},
		{},
		{Stdout: "existing\n"},
	}}
	application.runner = runner
	root.SetArgs([]string{"--config", path, "check", "site:staging"})
	if err := root.Execute(); err != nil {
		t.Fatalf("check error = %v\n%s", err, output.String())
	}
	for _, expected := range []string{
		"DistShip preflight · site:staging",
		"✓ READY TO DEPLOY",
		"Source  test @ merge123 · clean",
		"Target  web:/var/www/site",
		"Build   pnpm build-test → dist",
		"Recent commits · 1 shown · merges hidden",
		"fix: safe deploy",
		"12345678 · Sperains · 08-12 10:30",
		"Checks  local ✓ · build ✓ · Git ✓ · SSH ✓ · remote ✓",
		"No changes were made.",
	} {
		if !strings.Contains(output.String(), expected) {
			t.Errorf("output missing %q:\n%s", expected, output.String())
		}
	}
	for _, unexpected := range []string{"Local project directory is available", "Current revision:", "Remote directory exists and is writable"} {
		if strings.Contains(output.String(), unexpected) {
			t.Errorf("output contains verbose detail %q:\n%s", unexpected, output.String())
		}
	}
	if recentCommand := strings.Join(runner.commands[3].Args, " "); !strings.Contains(recentCommand, "--no-merges -3") {
		t.Fatalf("recent Git command does not exclude merges: %s", recentCommand)
	}
}

func TestCheckReportHandlesNoNonMergeCommits(t *testing.T) {
	var output bytes.Buffer
	application := &app{out: &output, translator: i18n.New(i18n.English)}
	environment := config.Environment{
		Build:    []string{"pnpm", "build"},
		Artifact: "dist",
		Target:   config.Target{Host: "example", Directory: "/www/site"},
	}
	report := preflight.Report{
		Repository:  gitinfo.Repository{IsRepository: true, Branch: "main", Revision: "a1b2c3d4"},
		RemoteState: preflight.RemoteExisting,
	}

	ref := config.TargetRef{ProjectID: "site", EnvironmentID: "test"}
	printCheckReport(application, ref, environment, report)
	if !strings.Contains(output.String(), "No non-merge commits found") {
		t.Fatalf("output = %s", output.String())
	}
}

func TestCheckReportElevatesWarnings(t *testing.T) {
	var output bytes.Buffer
	application := &app{out: &output, translator: i18n.New(i18n.English)}
	environment := config.Environment{
		Build:    []string{"pnpm", "build"},
		Artifact: "dist",
		Target:   config.Target{Host: "example", Directory: "/www/site"},
	}
	report := preflight.Report{
		Repository: gitinfo.Repository{
			IsRepository: true,
			Branch:       "main",
			Revision:     "a1b2c3d4",
			Changes:      gitinfo.Changes{Modified: 1},
		},
		RemoteState: preflight.RemoteExisting,
		Warnings:    []string{"dirty"},
	}

	ref := config.TargetRef{ProjectID: "site", EnvironmentID: "test"}
	printCheckReport(application, ref, environment, report)
	for _, expected := range []string{"! READY WITH WARNINGS", "main @ a1b2c3d4 · 1 changed entries", "Warnings", "Working tree has uncommitted changes"} {
		if !strings.Contains(output.String(), expected) {
			t.Errorf("output missing %q:\n%s", expected, output.String())
		}
	}
}
