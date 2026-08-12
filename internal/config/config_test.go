package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sperains/distship/internal/i18n"
)

func TestValidConfig(t *testing.T) {
	cfg := validConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestValidateReportsAllRelevantProblems(t *testing.T) {
	cfg := New()
	cfg.Version = 2
	cfg.Projects["IPD!"] = Project{
		Environments: map[string]Environment{
			"Test!": {
				Directory: "relative",
				Artifact:  "../dist",
				Target:    Target{Directory: "/"},
				Git:       GitPolicy{Dirty: "unknown"},
			},
		},
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() error = nil, want validation error")
	}
	message := i18n.New(i18n.SimplifiedChinese).Error(err)
	for _, expected := range []string{"version 必须为 1", "项目标识", ".name 不能为空", ".directory 必须是绝对路径", ".build", ".artifact", ".target.host", ".target.directory", ".git.dirty"} {
		if !strings.Contains(message, expected) {
			t.Errorf("Validate() error missing %q:\n%s", expected, message)
		}
	}
}

func TestValidateRejectsProjectRootAsArtifact(t *testing.T) {
	cfg := validConfig()
	project := cfg.Projects["ipd"]
	environment := project.Environments["test"]
	environment.Artifact = "."
	project.Environments["test"] = environment
	cfg.Projects["ipd"] = project

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), ".artifact") {
		t.Fatalf("Validate() error = %v, want artifact error", err)
	}
}

func TestValidateRejectsRemotePathWithControlCharacters(t *testing.T) {
	cfg := validConfig()
	project := cfg.Projects["ipd"]
	environment := project.Environments["test"]
	environment.Target.Directory = "/www/site\nother"
	project.Environments["test"] = environment
	cfg.Projects["ipd"] = project
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), ".target.directory") {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestSaveLoadAndDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "projects.toml")
	cfg := validConfig()
	environment := cfg.Projects["ipd"].Environments["test"]
	environment.Git.Dirty = ""
	project := cfg.Projects["ipd"]
	project.Environments["test"] = environment
	cfg.Projects["ipd"] = project

	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got := loaded.Projects["ipd"].Environments["test"].Git.Dirty; got != "warn" {
		t.Fatalf("dirty default = %q, want warn", got)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("config permissions = %o, want 600", info.Mode().Perm())
	}
}

func TestCloneDoesNotShareMutableValues(t *testing.T) {
	source := validConfig()
	clone := source.Clone()
	project := clone.Projects["ipd"]
	environment := project.Environments["test"]
	environment.Build[0] = "npm"
	environment.Git.AllowedBranches[0] = "main"
	project.Environments["test"] = environment
	clone.Projects["ipd"] = project

	original := source.Projects["ipd"].Environments["test"]
	if original.Build[0] != "pnpm" || original.Git.AllowedBranches[0] != "test" {
		t.Fatalf("Clone() mutated source: %#v", original)
	}
}

func TestLoadMissingFileWrapsNotExist(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "missing.toml"))
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Load() error = %v, want os.ErrNotExist", err)
	}
}

func validConfig() *Config {
	return &Config{
		Version: 1,
		Projects: map[string]Project{
			"ipd": {
				Name: "IPD",
				Environments: map[string]Environment{
					"test": {
						Name:      "测试环境",
						Directory: "/Users/example/ipd",
						Build:     []string{"pnpm", "build-test"},
						Artifact:  "dist",
						Target:    Target{Host: "bt_250", Directory: "/www/wwwgit/ipd-front"},
						Git:       GitPolicy{AllowedBranches: []string{"test"}, Dirty: "warn"},
					},
				},
			},
		},
	}
}
