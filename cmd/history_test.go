package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sperains/distship/internal/history"
)

func TestHistoryCommandFiltersWithoutConfiguration(t *testing.T) {
	stateHome := t.TempDir()
	t.Setenv("XDG_STATE_HOME", stateHome)
	t.Setenv("DISTSHIP_LANG", "en")
	path := filepath.Join(stateHome, "distship", "history.jsonl")
	for _, record := range []history.Record{
		{Project: "site", Environment: "test", Commit: "first", Host: "old", Directory: "/www/old", Artifact: "dist"},
		{Project: "other", Environment: "test", Commit: "other", Host: "web", Directory: "/www/other", Artifact: "dist"},
		{Project: "site", Environment: "test", Commit: "latest", Host: "new", Directory: "/www/new", Artifact: "build"},
	} {
		if err := history.Append(path, record); err != nil {
			t.Fatal(err)
		}
	}

	var output bytes.Buffer
	command := newRootCommand(strings.NewReader(""), &output, &output)
	command.SetArgs([]string{"--no-color", "history", "site:test", "--limit", "1"})
	if err := command.Execute(); err != nil {
		t.Fatalf("history error = %v\n%s", err, output.String())
	}
	for _, expected := range []string{"1 shown · site:test", "Target: new:/www/new", "Artifact: build"} {
		if !strings.Contains(output.String(), expected) {
			t.Errorf("history output missing %q:\n%s", expected, output.String())
		}
	}
	for _, unexpected := range []string{"old:/www/old", "other:test"} {
		if strings.Contains(output.String(), unexpected) {
			t.Errorf("history output contains %q:\n%s", unexpected, output.String())
		}
	}
}

func TestHistoryCommandAllowsMissingFile(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	t.Setenv("DISTSHIP_LANG", "en")
	var output bytes.Buffer
	command := newRootCommand(strings.NewReader(""), &output, &output)
	command.SetArgs([]string{"history"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "No local deployment history found.") {
		t.Fatalf("history output = %s", output.String())
	}
}

func TestHistoryCommandRejectsInvalidInputs(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	for _, args := range [][]string{{"history", "site"}, {"history", "--limit", "0"}, {"history", "site:test", "extra"}} {
		var output bytes.Buffer
		command := newRootCommand(strings.NewReader(""), &output, &output)
		command.SetArgs(args)
		if err := command.Execute(); err == nil {
			t.Fatalf("%v error = nil", args)
		}
	}
}

func TestHistoryCommandReportsMalformedLine(t *testing.T) {
	stateHome := t.TempDir()
	t.Setenv("XDG_STATE_HOME", stateHome)
	path := filepath.Join(stateHome, "distship", "history.jsonl")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("not-json\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	command := newRootCommand(strings.NewReader(""), &output, &output)
	command.SetArgs([]string{"history"})
	err := command.Execute()
	if err == nil || !strings.Contains(err.Error(), path) || !strings.Contains(err.Error(), "line 1") {
		t.Fatalf("history error = %v", err)
	}
}
