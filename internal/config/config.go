package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/pelletier/go-toml/v2"

	"github.com/sperains/distship/internal/i18n"
)

const CurrentVersion = 1

type Config struct {
	Version  int                `toml:"version"`
	Projects map[string]Project `toml:"projects"`
}

type Project struct {
	Name         string                 `toml:"name"`
	Environments map[string]Environment `toml:"environments"`
}

type Environment struct {
	Name      string    `toml:"name"`
	Directory string    `toml:"directory"`
	Build     []string  `toml:"build"`
	Artifact  string    `toml:"artifact"`
	Target    Target    `toml:"target"`
	Git       GitPolicy `toml:"git"`
}

type Target struct {
	Host      string `toml:"host"`
	Directory string `toml:"directory"`
}

type GitPolicy struct {
	AllowedBranches []string `toml:"allowed_branches,omitempty"`
	Dirty           string   `toml:"dirty,omitempty"`
}

func New() *Config {
	return &Config{Version: CurrentVersion, Projects: make(map[string]Project)}
}

func (c *Config) Clone() *Config {
	result := &Config{Version: c.Version, Projects: make(map[string]Project, len(c.Projects))}
	for projectID, project := range c.Projects {
		copyProject := Project{Name: project.Name, Environments: make(map[string]Environment, len(project.Environments))}
		for environmentID, environment := range project.Environments {
			environment.Build = append([]string(nil), environment.Build...)
			environment.Git.AllowedBranches = append([]string(nil), environment.Git.AllowedBranches...)
			copyProject.Environments[environmentID] = environment
		}
		result.Projects[projectID] = copyProject
	}
	return result
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, i18n.Wrap(i18n.ErrReadConfig, err, path)
	}
	cfg := New()
	decoder := toml.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(cfg); err != nil {
		return nil, i18n.Wrap(i18n.ErrParseConfig, err, path)
	}
	applyDefaults(cfg)
	return cfg, nil
}

func Save(path string, cfg *Config) error {
	data, err := toml.Marshal(cfg)
	if err != nil {
		return i18n.Wrap(i18n.ErrGenerateConfig, err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return i18n.Wrap(i18n.ErrCreateConfigDirectory, err)
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".projects-*.toml")
	if err != nil {
		return i18n.Wrap(i18n.ErrCreateTempConfig, err)
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if err := temp.Chmod(0o600); err != nil {
		temp.Close()
		return i18n.Wrap(i18n.ErrSetConfigPermissions, err)
	}
	if _, err := temp.Write(data); err != nil {
		temp.Close()
		return i18n.Wrap(i18n.ErrWriteTempConfig, err)
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return i18n.Wrap(i18n.ErrSyncTempConfig, err)
	}
	if err := temp.Close(); err != nil {
		return i18n.Wrap(i18n.ErrCloseTempConfig, err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return i18n.Wrap(i18n.ErrSaveConfig, err, path)
	}
	return nil
}

func (c *Config) EnvironmentCount() int {
	count := 0
	for _, project := range c.Projects {
		count += len(project.Environments)
	}
	return count
}

func applyDefaults(cfg *Config) {
	if cfg.Projects == nil {
		cfg.Projects = make(map[string]Project)
	}
	for projectID, project := range cfg.Projects {
		for environmentID, environment := range project.Environments {
			if environment.Git.Dirty == "" {
				environment.Git.Dirty = "warn"
			}
			project.Environments[environmentID] = environment
		}
		cfg.Projects[projectID] = project
	}
}

func (c *Config) Validate() error {
	var problems []validationIssue
	if c.Version != CurrentVersion {
		problems = append(problems, issue(i18n.ValidationVersion, CurrentVersion))
	}
	if len(c.Projects) == 0 {
		problems = append(problems, issue(i18n.ValidationProjectRequired))
	}
	projectIDs := sortedKeys(c.Projects)
	for _, projectID := range projectIDs {
		project := c.Projects[projectID]
		prefix := fmt.Sprintf("projects.%s", projectID)
		if !validID(projectID) {
			problems = append(problems, issue(i18n.ValidationProjectID, prefix))
		}
		if strings.TrimSpace(project.Name) == "" {
			problems = append(problems, issue(i18n.ValidationFieldRequired, prefix+".name"))
		}
		if len(project.Environments) == 0 {
			problems = append(problems, issue(i18n.ValidationEnvironmentRequired, prefix))
		}
		for _, environmentID := range sortedKeys(project.Environments) {
			environment := project.Environments[environmentID]
			envPrefix := prefix + ".environments." + environmentID
			if !validID(environmentID) {
				problems = append(problems, issue(i18n.ValidationEnvironmentID, envPrefix))
			}
			if strings.TrimSpace(environment.Name) == "" {
				problems = append(problems, issue(i18n.ValidationFieldRequired, envPrefix+".name"))
			}
			if !filepath.IsAbs(environment.Directory) {
				problems = append(problems, issue(i18n.ValidationAbsolutePath, envPrefix+".directory"))
			}
			if len(environment.Build) == 0 || strings.TrimSpace(environment.Build[0]) == "" {
				problems = append(problems, issue(i18n.ValidationBuildRequired, envPrefix+".build"))
			}
			if environment.Artifact == "" || filepath.Clean(environment.Artifact) == "." || filepath.IsAbs(environment.Artifact) || escapesProject(environment.Artifact) {
				problems = append(problems, issue(i18n.ValidationArtifactPath, envPrefix+".artifact"))
			}
			if !IsValidSSHTarget(environment.Target.Host) {
				problems = append(problems, issue(i18n.ValidationSSHTarget, envPrefix+".target.host"))
			}
			if !isSafeRemotePath(environment.Target.Directory) {
				problems = append(problems, issue(i18n.ValidationRemotePath, envPrefix+".target.directory"))
			}
			if environment.Git.Dirty != "allow" && environment.Git.Dirty != "warn" && environment.Git.Dirty != "deny" {
				problems = append(problems, issue(i18n.ValidationDirtyPolicy, envPrefix+".git.dirty"))
			}
		}
	}
	if len(problems) > 0 {
		return validationError(problems)
	}
	return nil
}

type validationIssue struct {
	key  i18n.Key
	args []any
}

func issue(key i18n.Key, args ...any) validationIssue {
	return validationIssue{key: key, args: args}
}

type validationError []validationIssue

func (e validationError) Error() string {
	return e.Localize(i18n.New(i18n.English))
}

func (e validationError) Localize(translator i18n.Translator) string {
	lines := make([]string, len(e))
	for index, problem := range e {
		lines[index] = fmt.Sprintf("  %d. %s", index+1, translator.T(problem.key, problem.args...))
	}
	return strings.Join(lines, "\n")
}

func validID(value string) bool {
	if value == "" {
		return false
	}
	for index, char := range value {
		valid := char >= 'a' && char <= 'z' || char >= '0' && char <= '9' || index > 0 && (char == '_' || char == '-')
		if !valid {
			return false
		}
	}
	return true
}

func escapesProject(path string) bool {
	cleaned := filepath.Clean(path)
	return cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator))
}

func isSafeRemotePath(path string) bool {
	if !strings.HasPrefix(path, "/") || strings.ContainsFunc(path, unicode.IsControl) {
		return false
	}
	cleaned := filepath.Clean(path)
	if cleaned == "/" || cleaned == "/www" || cleaned == "/root" || cleaned == "/home" || cleaned == "/Users" {
		return false
	}
	return len(strings.Split(strings.Trim(cleaned, "/"), "/")) >= 2
}

func sortedKeys[T any](values map[string]T) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
