package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sperains/distship/internal/config"
)

func TestInitValidateAndList(t *testing.T) {
	t.Setenv("DISTSHIP_LANG", "zh-CN")
	directory := t.TempDir()
	path := filepath.Join(t.TempDir(), "distship", "projects.toml")
	input := strings.Join([]string{
		"ipd",
		"",
		`pnpm "build test"`,
		"dist",
		"bt_250:/www/wwwgit/ipd-front",
		"y",
	}, "\n") + "\n"

	var initOutput bytes.Buffer
	initCommand := newRootCommand(strings.NewReader(input), &initOutput, &initOutput)
	initCommand.SetArgs([]string{"--config", path, "init", directory})
	if err := initCommand.Execute(); err != nil {
		t.Fatalf("init error = %v\n%s", err, initOutput.String())
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	environment := cfg.Projects["ipd"].Environments["test"]
	if got := strings.Join(environment.Build, "|"); got != "pnpm|build test" {
		t.Fatalf("build = %q, want quoted argument preserved", got)
	}
	if environment.Artifact != "dist" || environment.Git.Dirty != "warn" {
		t.Fatalf("defaults not applied: %+v", environment)
	}
	if !strings.Contains(initOutput.String(), "✓ 配置有效") {
		t.Fatalf("init did not report automatic validation:\n%s", initOutput.String())
	}

	var validateOutput bytes.Buffer
	validateCommand := newRootCommand(strings.NewReader(""), &validateOutput, &validateOutput)
	validateCommand.SetArgs([]string{"--config", path, "config", "validate"})
	if err := validateCommand.Execute(); err != nil {
		t.Fatalf("validate error = %v", err)
	}
	if !strings.Contains(validateOutput.String(), "✓ 配置有效") {
		t.Fatalf("unexpected validate output:\n%s", validateOutput.String())
	}

	var listOutput bytes.Buffer
	listCommand := newRootCommand(strings.NewReader(""), &listOutput, &listOutput)
	listCommand.SetArgs([]string{"--config", path, "--no-color", "list"})
	if err := listCommand.Execute(); err != nil {
		t.Fatalf("list error = %v", err)
	}
	for _, expected := range []string{"[1] ipd · test", "ipd:test", "bt_250:/www/wwwgit/ipd-front"} {
		if !strings.Contains(listOutput.String(), expected) {
			t.Errorf("list output missing %q:\n%s", expected, listOutput.String())
		}
	}
}

func TestAdvancedInitializationKeepsCustomNames(t *testing.T) {
	t.Setenv("DISTSHIP_LANG", "en")
	directory := t.TempDir()
	path := filepath.Join(t.TempDir(), "projects.toml")
	input := strings.Join([]string{
		"ipd",
		"IPD Website",
		"test",
		"Test Environment",
		"pnpm build-test",
		"dist",
		"bt_250",
		"/www/wwwgit/ipd-front",
		"test",
		"warn",
		"y",
	}, "\n") + "\n"
	var output bytes.Buffer
	command := newRootCommand(strings.NewReader(input), &output, &output)
	command.SetArgs([]string{"--config", path, "init", directory, "--advanced"})
	if err := command.Execute(); err != nil {
		t.Fatalf("advanced init error = %v\n%s", err, output.String())
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Projects["ipd"].Name != "IPD Website" || cfg.Projects["ipd"].Environments["test"].Name != "Test Environment" {
		t.Fatalf("custom names not preserved: %#v", cfg.Projects["ipd"])
	}
}

func TestQuickInitializationPreservesExistingDisplayNames(t *testing.T) {
	t.Setenv("DISTSHIP_LANG", "zh-CN")
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "package.json"), []byte(`{"name":"ipd","scripts":{"build-test":"vite build"}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "pnpm-lock.yaml"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "projects.toml")
	cfg := &config.Config{Version: 1, Projects: map[string]config.Project{
		"ipd": {Name: "IPD 网站", Environments: map[string]config.Environment{
			"test": {Name: "测试环境", Directory: directory, Build: []string{"pnpm", "build-test"}, Artifact: "dist", Target: config.Target{Host: "old", Directory: "/var/www/old"}, Git: config.GitPolicy{Dirty: "warn"}},
		}},
	}}
	if err := config.Save(path, cfg); err != nil {
		t.Fatal(err)
	}
	input := strings.Join([]string{"", "", "", "bt_250", "/www/wwwgit/ipd-front", "y"}, "\n") + "\n"
	var output bytes.Buffer
	command := newRootCommand(strings.NewReader(input), &output, &output)
	command.SetArgs([]string{"--config", path, "init", directory})
	if err := command.Execute(); err != nil {
		t.Fatalf("quick overwrite error = %v\n%s", err, output.String())
	}
	saved, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	project := saved.Projects["ipd"]
	if project.Name != "IPD 网站" || project.Environments["test"].Name != "测试环境" {
		t.Fatalf("display names changed: %#v", project)
	}
	if strings.Count(output.String(), "更新这个部署目标？") != 1 || strings.Contains(output.String(), "已存在，覆盖") {
		t.Fatalf("expected a single confirmation prompt:\n%s", output.String())
	}
	if !strings.Contains(output.String(), "配置变更") || !strings.Contains(output.String(), "当前：old:/var/www/old") || !strings.Contains(output.String(), "新值：bt_250:/www/wwwgit/ipd-front") {
		t.Fatalf("missing update diff:\n%s", output.String())
	}
}

func TestInitWithoutDirectoryArgumentRequiresInput(t *testing.T) {
	t.Setenv("DISTSHIP_LANG", "en")
	path := filepath.Join(t.TempDir(), "projects.toml")
	var output bytes.Buffer
	command := newRootCommand(strings.NewReader("\n"), &output, &output)
	command.SetArgs([]string{"--config", path, "init"})
	err := command.Execute()
	if err == nil || !strings.Contains(err.Error(), "Local directory is required") {
		t.Fatalf("error = %v, want required local directory", err)
	}
	if !strings.Contains(output.String(), "Local directory (path or . for the current directory):") {
		t.Fatalf("directory prompt missing:\n%s", output.String())
	}
}

func TestInitDetectsCurrentProjectDirectory(t *testing.T) {
	t.Setenv("DISTSHIP_LANG", "en")
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "package.json"), []byte(`{"name":"current-site","scripts":{"build-test":"vite build"}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "pnpm-lock.yaml"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(directory)
	path := filepath.Join(t.TempDir(), "projects.toml")
	input := strings.Join([]string{
		"",
		"",
		"",
		"root@119.36.78.123",
		"/www/site",
		"n",
	}, "\n") + "\n"
	var output bytes.Buffer
	command := newRootCommand(strings.NewReader(input), &output, &output)
	command.SetArgs([]string{"--config", path, "init"})
	if err := command.Execute(); err != nil {
		t.Fatalf("init error = %v\n%s", err, output.String())
	}
	if !strings.Contains(output.String(), "Detected current project directory") || strings.Contains(output.String(), "Local directory (path or . for the current directory):") {
		t.Fatalf("current project was not detected:\n%s", output.String())
	}
	if !strings.Contains(output.String(), "root@119.36.78.123:/www/site") {
		t.Fatalf("split deployment target missing from preview:\n%s", output.String())
	}
}

func TestInitRejectsMissingDirectory(t *testing.T) {
	t.Setenv("DISTSHIP_LANG", "en")
	path := filepath.Join(t.TempDir(), "projects.toml")
	missing := filepath.Join(t.TempDir(), "missing")
	var output bytes.Buffer
	command := newRootCommand(strings.NewReader(""), &output, &output)
	command.SetArgs([]string{"--config", path, "init", missing})
	err := command.Execute()
	if err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("error = %v, want missing directory error", err)
	}
}

func TestInitExistingTargetWithNoChangesDoesNotConfirm(t *testing.T) {
	t.Setenv("DISTSHIP_LANG", "en")
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "package.json"), []byte(`{"name":"jd_ipd","scripts":{"build-staging":"vite build"}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "projects.toml")
	cfg := testConfig()
	project := cfg.Projects["site"]
	environment := project.Environments["staging"]
	environment.Directory = directory
	environment.Build = []string{"npm", "run", "build-staging"}
	environment.Git.AllowedBranches = nil
	project.Environments["staging"] = environment
	cfg.Projects["site"] = project
	if err := config.Save(path, cfg); err != nil {
		t.Fatal(err)
	}
	input := strings.Join([]string{"", "", "", "", ""}, "\n") + "\n"
	var output bytes.Buffer
	command := newRootCommand(strings.NewReader(input), &output, &output)
	command.SetArgs([]string{"--config", path, "init", directory})
	if err := command.Execute(); err != nil {
		t.Fatalf("init error = %v\n%s", err, output.String())
	}
	if !strings.Contains(output.String(), "No configuration changes detected") || strings.Contains(output.String(), "Update this deployment target?") {
		t.Fatalf("unexpected no-change flow:\n%s", output.String())
	}
	if !strings.Contains(output.String(), "Project ID [site]:") || !strings.Contains(output.String(), "Environment ID [staging]:") {
		t.Fatalf("configured target was not reused as the default:\n%s", output.String())
	}
}

func TestEnglishOutput(t *testing.T) {
	t.Setenv("DISTSHIP_LANG", "en")
	path := filepath.Join(t.TempDir(), "projects.toml")
	if err := config.Save(path, testConfig()); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	command := newRootCommand(strings.NewReader(""), &output, &output)
	command.SetArgs([]string{"--config", path, "--lang", "en", "config", "validate"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"Configuration is valid", "Path:", "Projects: 1", "Targets: 1"} {
		if !strings.Contains(output.String(), expected) {
			t.Errorf("English output missing %q:\n%s", expected, output.String())
		}
	}
}

func TestSplitCommand(t *testing.T) {
	arguments, err := splitCommand(`pnpm run "build test" -- --mode=test`)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(arguments, "|"); got != "pnpm|run|build test|--|--mode=test" {
		t.Fatalf("splitCommand() = %q", got)
	}
	if _, err := splitCommand(`pnpm "build`); err == nil {
		t.Fatal("splitCommand() error = nil for unclosed quote")
	}
}

func TestParseDeployTarget(t *testing.T) {
	tests := []struct {
		input     string
		host      string
		directory string
		complete  bool
	}{
		{"bt_250:/www/wwwgit/ipd-front", "bt_250", "/www/wwwgit/ipd-front", true},
		{"root@119.36.78.123:/www/site", "root@119.36.78.123", "/www/site", true},
		{"119.36.78.123", "119.36.78.123", "", false},
		{"example.com", "example.com", "", false},
	}
	for _, test := range tests {
		host, directory, complete, err := parseDeployTarget(test.input)
		if err != nil {
			t.Fatalf("parseDeployTarget(%q) error = %v", test.input, err)
		}
		if host != test.host || directory != test.directory || complete != test.complete {
			t.Fatalf("parseDeployTarget(%q) = %q, %q, %v", test.input, host, directory, complete)
		}
	}
	for _, invalid := range []string{"host:22", "-oProxyCommand=x", "user name@host", "host/relative"} {
		if _, _, _, err := parseDeployTarget(invalid); err == nil {
			t.Fatalf("parseDeployTarget(%q) error = nil", invalid)
		}
	}
}

func TestMissingConfigSuggestsInit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.toml")
	var output bytes.Buffer
	command := newRootCommand(strings.NewReader(""), &output, &output)
	command.SetArgs([]string{"--config", path, "list"})
	err := command.Execute()
	if err == nil || !strings.Contains(err.Error(), "distship init") {
		t.Fatalf("error = %v, want init guidance", err)
	}
}

func testConfig() *config.Config {
	return &config.Config{
		Version: 1,
		Projects: map[string]config.Project{
			"site": {
				Name: "Website",
				Environments: map[string]config.Environment{
					"staging": {
						Name: "Staging", Directory: "/srv/site", Build: []string{"npm", "run", "build"}, Artifact: "dist",
						Target: config.Target{Host: "web", Directory: "/var/www/site"}, Git: config.GitPolicy{Dirty: "warn"},
					},
				},
			},
		},
	}
}
