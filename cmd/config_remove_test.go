package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sperains/distship/internal/config"
)

func TestConfigRemoveKeepsRemainingTargetsAndValidates(t *testing.T) {
	t.Setenv("DISTSHIP_LANG", "en")
	path := filepath.Join(t.TempDir(), "projects.toml")
	cfg := testConfig()
	project := cfg.Projects["site"]
	project.Environments["production"] = config.Environment{
		Name: "Production", Directory: "/srv/site", Build: []string{"npm", "run", "build"}, Artifact: "dist",
		Target: config.Target{Host: "web", Directory: "/var/www/site-production"}, Git: config.GitPolicy{Dirty: "deny"},
	}
	cfg.Projects["site"] = project
	if err := config.Save(path, cfg); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	command := newRootCommand(strings.NewReader(""), &output, &output)
	command.SetArgs([]string{"--config", path, "config", "remove", "site:staging", "--yes"})
	if err := command.Execute(); err != nil {
		t.Fatalf("remove error = %v\n%s", err, output.String())
	}
	saved, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, exists := saved.Projects["site"].Environments["staging"]; exists {
		t.Fatal("removed target still exists")
	}
	if _, exists := saved.Projects["site"].Environments["production"]; !exists {
		t.Fatal("remaining target was removed")
	}
	for _, expected := range []string{"Deployment target removed", "Configuration is valid", "No local project or server files will be removed"} {
		if !strings.Contains(output.String(), expected) {
			t.Errorf("output missing %q:\n%s", expected, output.String())
		}
	}
}

func TestConfigRemoveArchivesLastTarget(t *testing.T) {
	t.Setenv("DISTSHIP_LANG", "en")
	directory := t.TempDir()
	path := filepath.Join(directory, "projects.toml")
	if err := config.Save(path, testConfig()); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	command := newRootCommand(strings.NewReader(""), &output, &output)
	command.SetArgs([]string{"--config", path, "config", "remove", "site:staging", "--yes"})
	if err := command.Execute(); err != nil {
		t.Fatalf("remove error = %v\n%s", err, output.String())
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("active configuration still exists or stat failed: %v", err)
	}
	backups, err := filepath.Glob(path + ".bak-*")
	if err != nil || len(backups) != 1 {
		t.Fatalf("backups = %#v, error = %v", backups, err)
	}
	archived, err := config.Load(backups[0])
	if err != nil {
		t.Fatal(err)
	}
	if err := archived.Validate(); err != nil {
		t.Fatalf("archived configuration invalid: %v", err)
	}
	if !strings.Contains(output.String(), "configuration was archived") || !strings.Contains(output.String(), backups[0]) {
		t.Fatalf("unexpected output:\n%s", output.String())
	}
}

func TestConfigRemoveCanBeCancelled(t *testing.T) {
	t.Setenv("DISTSHIP_LANG", "en")
	path := filepath.Join(t.TempDir(), "projects.toml")
	if err := config.Save(path, testConfig()); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	command := newRootCommand(strings.NewReader("n\n"), &output, &output)
	command.SetArgs([]string{"--config", path, "config", "remove", "site:staging"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	saved, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if saved.EnvironmentCount() != 1 || !strings.Contains(output.String(), "Configuration unchanged") {
		t.Fatalf("configuration changed after cancellation:\n%s", output.String())
	}
}

func TestConfigRemoveMissingTargetListsAvailableTargets(t *testing.T) {
	t.Setenv("DISTSHIP_LANG", "en")
	path := filepath.Join(t.TempDir(), "projects.toml")
	if err := config.Save(path, testConfig()); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	command := newRootCommand(strings.NewReader(""), &output, &output)
	command.SetArgs([]string{"--config", path, "config", "remove", "missing:test", "--yes"})
	err := command.Execute()
	if err == nil || !strings.Contains(err.Error(), "site:staging") {
		t.Fatalf("error = %v, want available target", err)
	}
}

func TestConfigRemoveRejectsInvalidTargetID(t *testing.T) {
	t.Setenv("DISTSHIP_LANG", "en")
	path := filepath.Join(t.TempDir(), "projects.toml")
	if err := config.Save(path, testConfig()); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	command := newRootCommand(strings.NewReader(""), &output, &output)
	command.SetArgs([]string{"--config", path, "config", "remove", "site", "--yes"})
	err := command.Execute()
	if err == nil || !strings.Contains(err.Error(), "project:environment") {
		t.Fatalf("error = %v, want target ID guidance", err)
	}
}

func TestConfigRemoveRejectsLegacyArguments(t *testing.T) {
	t.Setenv("DISTSHIP_LANG", "en")
	var output bytes.Buffer
	command := newRootCommand(strings.NewReader(""), &output, &output)
	command.SetArgs([]string{"config", "remove", "site", "staging"})
	if err := command.Execute(); err == nil {
		t.Fatal("legacy two-argument target selector was accepted")
	}
}
